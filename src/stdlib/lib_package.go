package stdlib

import (
	"fmt"
	"os"
	"strings"

	"nskbz.cn/lua/api"
	"nskbz.cn/lua/tool"
)

const ( //package.config默认值
	PACKAGE_CONF_DIR_SEP   = "/" //目录分隔符
	PACKAGE_CONF_PATH_SEP  = ";" //路径分隔符
	PACKAGE_CONF_PATH_MARK = "?" //路径占位符,会被模块名替换
	PACKAGE_CONF_EXEC_DIR  = "!" //可选的目录分隔替换标记（表示模块名中的_转换为目录分隔符）
	PACKAGE_CONF_IGMARK    = "-" //???
)

var packageFuncs map[string]api.GoFunc = map[string]api.GoFunc{
	"searchpath": packageSearchPath,
}

var searchers []api.GoFunc = []api.GoFunc{
	preloadSearcher,
	luaSearcher,
}

func OpenPackageLib(vm api.LuaVM) int {
	vm.NewTable() //package_table
	for k, v := range packageFuncs {
		vm.PushValue(0) //packge_table_copy
		vm.PushGoFunction(v, 1)
		vm.SetField(-1, k)
	}

	//注册全局require函数
	vm.PushValue(0) //packge_table_copy
	vm.PushGoFunction(packageRequire, 1)
	vm.SetGlobal("require")

	//创建searchers_table
	vm.NewTable()
	for i, v := range searchers {
		vm.PushValue(-1) //package_table
		vm.PushGoFunction(v, 1)
		vm.SetI(-1, int64(i+1))
	}
	vm.SetField(-1, "searchers") //设置searchers

	//设置config
	vm.PushString(PACKAGE_CONF_DIR_SEP + "\n" + PACKAGE_CONF_PATH_SEP + "\n" +
		PACKAGE_CONF_PATH_MARK + "\n" + PACKAGE_CONF_EXEC_DIR + "\n" +
		PACKAGE_CONF_IGMARK + "\n")
	vm.SetField(-1, "config")

	//设置path
	vm.PushString("./?.lua;./?/init.lua") //目前只从当前目录下找
	vm.SetField(-1, "path")

	//设置loaded
	vm.GetSubTable(api.LUA_REGISTRY_INDEX, api.LUA_LOADED_TABLE)
	vm.SetField(-1, "loaded")

	//设置preloaded ，设置'_LOADED'时的全局表还在栈顶所以不需要PushGlobalTable
	vm.GetSubTable(api.LUA_REGISTRY_INDEX, api.LUA_PRELOAD_TABLE)
	vm.SetField(-1, "preload")

	return 1
}

// require(modname)
// 返回对应模块(table)
func packageRequire(vm api.LuaVM) int {
	vm.CheckString(1)
	modname := vm.ToString(1)
	//检查是否已经加载过了
	if vm.GetSubTable(api.LUA_REGISTRY_INDEX, api.LUA_LOADED_TABLE) && vm.GetField(0, modname) == api.LUAVALUE_TABLE {
		return 1 //已加载的模块直接返回
	}
	vm.Pop(1) //pop GetField
	//没有加载过，则依次使用注册的searcher进行搜索
	find := false
	if !vm.GetSubTable(0, "package") { //获取package_table
		panic("can't get package from _LOADED!")
	}
	vm.GetSubTable(0, "searchers") //if package exist,searchers must be involved
	for i := 1; i <= vm.RawLen(0); i++ {
		if vm.GetI(0, int64(i)) != api.LUAVALUE_FUNCTION {
			tool.Fatal(vm, fmt.Sprintf("searcher must be a function, but %s", vm.TypeName2(0)))
		}
		vm.PushString(modname)
		vm.Call(1, 1) //调用searcher尝试获取Loader方法
		if vm.Type(0) == api.LUAVALUE_FUNCTION {
			find = true
			break //此时栈顶为Loader
		}
		vm.Pop(1) //弹出Call的非table返回值
	}

	if !find {
		return vm.Error2("module [%s] can't find!", modname)
	}
	vm.PushString(modname)  //modename会作为每个loader函数的唯一入参
	vm.Call(1, 1)           //获取loader执行后的table返回值
	vm.PushValue(0)         //module_table_copy
	vm.SetField(2, modname) //添加进'_LOADED'中
	return 1
}

// package.searchpath(name, path [, path_mark [, exec_dir]])
// 用于在指定的路径模板中搜索模块或文件的实际路径
// 参数：
// name			string	要搜索的模块名
// path			string	路径模板字符串（多个路径用默认用';'分隔）
// path_mark	string	可选，替换标记（默认为 '?' 相当于通配符,会被模块名替换）
// exec_dir		string	可选，目录分隔符替换标记（默认为 '!' 将_替换为目录分隔符）
// 返回值：
// 如果找到文件，返回第一个匹配的完整路径
// 如果未找到，返回 nil 加上错误信息（描述所有尝试过的路径）
func packageSearchPath(vm api.LuaVM) int {
	vm.CheckString(1)
	vm.CheckString(2)
	modename := vm.ToString(1)
	path := vm.ToString(2)
	vm.PushValue(vm.UpvalueIndex(1)) //push package_table
	vm.GetField(0, "config")
	configs := strings.Split(vm.ToString(0), "\n")
	dir_sep := configs[0]
	path_sep := configs[1]
	path_mark := configs[2]
	exec_dir := configs[3]
	if vm.GetTop() >= 3 && vm.IsString(3) { //解析第三个参数
		path_mark = vm.ToString(3)
	}
	if vm.GetTop() >= 4 && vm.IsString(4) { //解析第四个参数
		exec_dir = vm.ToString(4)
	}

	ok, str := _searchPath(modename, path, dir_sep, path_sep, path_mark, exec_dir)
	if !ok {
		vm.PushString(fmt.Sprintf("package.searcher error : %s", str)) //push error msg
		return 1
	}
	vm.PushString(str)
	return 1
}

func _searchPath(modname, path, dir_sep, path_sep, path_mark, exec_dir string) (bool, string) {
	errmsg := ""
	for _, v := range strings.Split(path, path_sep) {
		v = strings.ReplaceAll(v, path_mark, modname)
		if exec_dir == "!" {
			v = strings.ReplaceAll(v, "_", dir_sep)
		}
		if _, err := os.Stat(v); err != nil {
			tool.Error("searchPath[%s] error: %s", v, err.Error())
			errmsg += fmt.Sprintf("\n\tno file '%s'", v)
			continue
		}
		return true, v
	}
	return false, errmsg
}

/*
*	Searcher 搜索器,用于找到模块所对应的加载方法
 */

// preload搜索器
func preloadSearcher(vm api.LuaVM) int {
	vm.CheckString(1)
	modename := vm.ToString(1)
	vm.PushValue(vm.UpvalueIndex(1))                           //push package_table
	vm.GetField(api.LUA_REGISTRY_INDEX, api.LUA_PRELOAD_TABLE) //push preload_table
	if vm.GetField(0, modename) != api.LUAVALUE_FUNCTION {
		errmsg := fmt.Sprintf("loader must be a function, but %s ; nil meanings maybe can't find %s from package.preload.", vm.TypeName2(0), modename)
		tool.Error("From preloadSearcher : %s", errmsg)
		vm.PushString(errmsg)
	}
	return 1
}

// lua搜索器
func luaSearcher(vm api.LuaVM) int {
	vm.CheckString(1)
	modename := vm.ToString(1)
	vm.PushValue(vm.UpvalueIndex(1)) //push package_table
	vm.GetField(0, "path")
	path := vm.ToString(0)
	vm.GetField(-1, "config")
	configs := strings.Split(vm.ToString(0), "\n") //[DIR_SEP][PATH_SEP][PATH_MARK][EXEC_DIR][IGMARK]

	ok, str := _searchPath(modename, path, configs[0], configs[1], configs[2], configs[3])
	vm.Pop(4) //弹出modename,package_table,path,config,此时栈上为空

	if !ok {
		vm.PushString(fmt.Sprintf("From luaSearcher : can't find loader %s", str))
		return 1
	}
	//加载找到的lua文件
	vm.PushString(str)
	if baseLoadFile(vm) != 1 { //包装一下错误信息
		errmsg := vm.ToString(0)
		vm.Pop(1)
		vm.PushString(fmt.Sprintf("From luaSearcher : %s", errmsg))
	}
	return 1
}
