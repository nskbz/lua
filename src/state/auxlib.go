package state

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"nskbz.cn/lua/api"
	"nskbz.cn/lua/stdlib"
	"nskbz.cn/lua/tool"
)

/*
	Error-report functions
*/

func (s *luaState) Error2(format string, a ...interface{}) int {
	s.PushString(fmt.Sprintf(format, a...))
	return s.Error()
}

func (s *luaState) ArgError(arg int, extraMsg string) int {
	// bad argument #arg to 'funcname' (extramsg)
	return s.Error2("bad argument %d to (%s)", arg, extraMsg)
}

/*
	Argument check functions
*/

func (s *luaState) ArgCheck(cond bool, arg int, extraMsg string) {
	if !cond {
		s.ArgError(arg, extraMsg)
	}
}

func (s *luaState) CheckAny(arg int) {
	if api.LUAVALUE_NONE == s.Type(arg) {
		s.ArgError(arg, fmt.Sprintf("arg[%d] is not a LuaValueType!", arg))
	}
}

func (s *luaState) CheckType(arg int, t api.LuaValueType) {
	at := s.Type(arg)
	if t != at {
		s.ArgError(arg, fmt.Sprintf("arg[%d](%s) is not LuaValue (%s) !", arg, s.TypeName(at), s.TypeName(t)))
	}
}

func (s *luaState) CheckNumber(idx int) bool {
	_, ok := s.ToIntegerX(idx)
	if ok {
		return true
	}
	_, ok = s.ToFloatX(idx)
	if ok {
		return true
	}
	return false
}

func (s *luaState) CheckInteger(arg int) int64 {
	i, ok := s.ToIntegerX(arg)
	if !ok {
		s.CheckType(arg, api.LUAVALUE_NUMBER)
	}
	return i
}

func (s *luaState) CheckFloat(arg int) float64 {
	i, ok := s.ToFloatX(arg)
	if !ok {
		s.CheckType(arg, api.LUAVALUE_NUMBER)
	}
	return i
}

func (s *luaState) CheckString(arg int) string {
	i, ok := s.ToStringX(arg)
	if !ok {
		s.CheckType(arg, api.LUAVALUE_STRING)
	}
	return i
}

func (s *luaState) OptInteger(arg int, d int64) int64 {
	if s.IsNoneOrNil(arg) {
		return d
	}
	return s.CheckInteger(arg)
}

func (s *luaState) OptFloat(arg int, d float64) float64 {
	if s.IsNoneOrNil(arg) {
		return d
	}
	return s.CheckFloat(arg)
}

func (s *luaState) OptString(arg int, d string) string {
	if s.IsNoneOrNil(arg) {
		return d
	}
	return s.CheckString(arg)
}

/*
	Load functions
*/

// (luaL_loadfile（L, filename) || lua_pcall(L, 0, LUA_MULTRET, 0)）
func (s *luaState) DoFile(filename string) bool {
	return s.LoadFile(filename) != api.LUA_OK || s.PCall(0, api.LUA_MULTRET, false) != api.LUA_OK
}

func (s *luaState) LoadFile(filename string) int {
	return s.LoadFileX(filename, "bt") //默认是二进制或文本模式
}

// http://www.lua.org/manual/5.3/manual.html#luaL_loadfilex
func (s *luaState) LoadFileX(filename, mode string) int {
	var data []byte
	var err error
	if len(filename) == 0 {
		data, err = io.ReadAll(os.Stdin)
	} else {
		if f, e := os.Open(filename); e != nil {
			panic(e.Error())
		} else {
			data, err = io.ReadAll(f)
		}
	}
	if err != nil {
		panic(err.Error())
	}

	//如果文件中的第一行以#开头，则忽略它
	if data[0] == '#' {
		return api.LUA_ERR_FILE
	}
	return s.Load(data, "@"+filename, mode)
}

// (luaL_loadstring(L, str) || lua_pcall(L, 0, LUA_MULTRET, 0))
func (s *luaState) DoString(str string) bool {
	return s.LoadString(str) != api.LUA_OK || s.PCall(0, api.LUA_MULTRET, false) != api.LUA_OK
}

var loadStringIdx int = 0

func (s *luaState) LoadString(str string) int {
	result := s.Load([]byte(str), "@string"+strconv.Itoa(loadStringIdx), "bt")
	loadStringIdx++
	return result
}

/*
	Other functions
*/

func (s *luaState) TypeName2(idx int) string {
	return s.TypeName(s.Type(idx))
}

func (s *luaState) Len2(idx int) int64 {
	s.Len(idx) //压入长度
	i, ok := s.ToIntegerX(0)
	if !ok {
		s.Error2("object length is not an integer")
	}
	s.Pop(1) //弹出长度
	return i
}

// 将给定索引的LuaValue转换字符串型。结果字符串压入堆栈，并由函数返回。如果该LuaValue存在元方法"__tostring",则应调用元方法
func (s *luaState) ToString2(idx int) string {
	idx = s.AbsIndex(idx)
	if s.CallMeta(idx, META_TOSTRING) && s.Type(idx) == api.LUAVALUE_TABLE { //tostring方法只对table生效
		if str, ok := s.ToStringX(0); !ok {
			s.Error2(META_TOSTRING + " must return a string")
		} else {
			return str
		}
	}
	//这里不能用ToStringX方法尝试转化,ToStringX规定只能转换Number
	//而该方法要求能转换所有LuaValue类型，所以需要单独处理
	tp := s.Type(idx)
	if tp == api.LUAVALUE_STRING {
		s.PushValue(idx)
		return s.ToString(0)
	}

	switch tp {
	case api.LUAVALUE_NIL:
		s.PushString("nil")
		return "nil"
	case api.LUAVALUE_BOOLEAN:
		if s.ToBoolean(idx) {
			s.PushString("true")
			break
		}
		s.PushString("false")
	case api.LUAVALUE_NUMBER:
		if i, ok := s.ToIntegerX(idx); ok {
			s.PushString(fmt.Sprintf("%d", i))
		} else if f, ok := s.ToFloatX(idx); ok {
			s.PushString(fmt.Sprintf("%g", f))
		}
	case api.LUAVALUE_STRING:
		s.PushValue(idx)
	default:
		s.PushString(fmt.Sprintf("%s: %p", s.Type(idx).String(), s.ToPointer(idx)))

		// if !s.GetMetaTable(idx) {

		// 	break
		// }
		// mf := s.GetMetafield(0, META_TOSTRING) //'__tostring'为自定义返回函数返回一个字符串

		// //mf不为NIL才说明有'__tostring',即GetMetafield成功并压入了'__tostring'的值,所以需要移除该值
		// //为NIL则没有压入'__tostring'的值,所以无需移除
		// if mf != api.LUAVALUE_NIL {
		// 	s.Replace(-1)
		// }
	}
	return s.CheckString(0)
}

// 确保获取表中的表元素。table=R(idx) and type(table[fname])==table。如果fname对应的键值是table则返回true并将其压入栈，反之类型不是table则返回false并创建table
func (s *luaState) GetSubTable(idx int, fname string) bool {
	idx = s.AbsIndex(idx)
	if s.Type(idx) != api.LUAVALUE_TABLE {
		tool.Fatal(s, fmt.Sprintf("GetSubTable error: expected table,but %s", s.TypeName2(idx)))
	}
	val := s.GetField(idx, fname) //压入value
	if val == api.LUAVALUE_TABLE {
		return true
	}
	//如果不是table类型则创建一个table
	s.Pop(1) //弹出上面GetField的值
	s.NewTable()
	s.PushValue(0)
	s.SetField(idx, fname)
	return false
}

// 将索引为obj的对象的元表中的字段e压入堆栈，并返回压入值的类型。如果对象没有元表，或者元表没有此字段，则不推送任何内容并返回LUA_NIL。
func (s *luaState) GetMetafield(obj int, e string) api.LuaValueType {
	hasMetaTable := s.GetMetaTable(obj)
	if hasMetaTable {
		metaField := s.GetField(0, e) //压入key=e的luavalue

		if metaField == api.LUAVALUE_NIL {
			s.Pop(2)
		} else {
			s.Remove(-1) //移除metatable
		}
		return metaField
	}
	return api.LUAVALUE_NIL
}

// 调用元方法;如果索引obj处的对象有元表，并且这个元表有一个字段e，则此函数调用该字段，并将该对象作为其唯一参数，函数返回true并将调用返回的值压入堆栈。如果没有元表或元方法，则此函数返回false（不向堆栈上压入任何值）。
func (s *luaState) CallMeta(obj int, e string) bool {
	obj = s.AbsIndex(obj)
	if mf := s.GetMetafield(obj, e); mf != api.LUAVALUE_NIL {
		s.PushValue(obj)           //压入对象自身作为唯一参数
		s.Call(1, api.LUA_MULTRET) //LUA_MULTRET待验证
		return true
	}
	return false
}

func (s *luaState) OpenLibs() {
	libs := map[string]api.GoFunc{
		"_G":        stdlib.OpenBaseLib,
		"math":      stdlib.OpenMathLib,
		"table":     stdlib.OpenTableLib,
		"string":    stdlib.OpenStringLib,
		"os":        stdlib.OpenOsLib,
		"package":   stdlib.OpenPackageLib,
		"coroutine": stdlib.OpenCoroutineLib,
	}
	for lib, funcs := range libs {
		s.RequireF(lib, funcs, true) //golbal==true,即所有库都会加入全局表'_G'中
		s.Pop(1)
	}
}

// 确保modname模块加载
// 如果modname不存在于包中package.loaded,则以字符串modname作为参数调用函数openf，并在包中package.loaded装载模块
// 如果glb为true，还将模块存储到全局modname中
// 如果glb为false，则只有当前加载该模块的文件才能使用该模块
// !!!!!该方法需要在堆栈上留下模块的副本。
func (s *luaState) RequireF(modname string, openf api.GoFunc, glb bool) {
	s.GetSubTable(api.LUA_REGISTRY_INDEX, api.LUA_LOADED_TABLE) //获取所有已加载的库
	s.GetField(api.LUA_REGISTRY_INDEX, modname)                 //尝试获取modename的库
	//如果GetField成功,即存在modname对应的键值;反之没有get成功栈顶则为nil,则需要加载该模块
	if s.IsNil(0) {
		s.Pop(1) //弹出nil
		s.PushGoFunction(openf, 0)
		s.PushString(modname)   //作为唯一参数
		s.Call(1, 1)            //调用openf库加载方法并要获取唯一返回值于栈顶(luatable)
		s.PushValue(0)          //复制一份openf的返回值
		s.SetField(-2, modname) //将[modname,openf的返回值]注册进"_LOADED",执行后栈上由于上面复制了一份openf的返回值所以还剩下["_LOADED",openf返回值]
	}
	s.Remove(-1) //移除"_LOADED"table，此时栈上只剩openf的返回值

	//是否为全局模块
	if glb {
		s.PushValue(0)       //copy一份用于SetGlobal
		s.SetGlobal(modname) // _G[modname] = module
	}
}

// 通过funcs创建一个在栈顶的模块(table)
func (s *luaState) NewLib(funcs map[string]api.GoFunc) {
	s.NewTable()
	for k, v := range funcs {
		s.PushGoFunction(v, 0)
		s.SetField(-1, k)
	}
}
