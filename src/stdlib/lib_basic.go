package stdlib

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"nskbz.cn/lua/api"
	"nskbz.cn/lua/tool"
)

var baseFuncs = map[string]api.GoFunc{
	"print":        basePrint,
	"assert":       baseAssert,
	"error":        baseError,
	"select":       baseSelect,
	"ipairs":       baseIPairs,
	"pairs":        basePairs,
	"next":         baseNext,
	"load":         baseLoad,
	"loadfile":     baseLoadFile,
	"dofile":       baseDoFile,
	"pcall":        basePCall,
	"xpcall":       baseXPCall,
	"getmetatable": baseGetMetaTable,
	"setmetatable": baseSetMetaTable,
	"rawequal":     baseRawEqual,
	"rawlen":       baseRawLen,
	"rawget":       baseRawGet,
	"rawset":       baseRawSet,
	"type":         baseType,
	"tostring":     baseToString,
	"tonumber":     baseToNumber,
}

func OpenBaseLib(vm api.LuaVM) int {
	//基础库可以直接访问不需要使用modname.funcname的原因就是对于基础库的每一个func都进行了SetGlobal
	for k, v := range baseFuncs {
		vm.PushGoFunction(v, 0)
		vm.SetGlobal(k)
	}
	//将全局表命名为'_G'放入注册中心
	vm.PushGlobalTable()
	vm.PushValue(0) //copy
	vm.SetField(-1, "_G")
	//注册全局版本
	vm.PushString("LUA 5.3")
	vm.SetField(-1, "_VERSION")
	return 1 //只返回全局表
}

func basePrint(vm api.LuaVM) int {
	nArgs := vm.GetTop()
	for i := 1; i <= nArgs; i++ {
		if vm.IsBoolean(i) {
			fmt.Printf("%t", vm.ToBoolean(i))
		} else if vm.IsString(i) {
			fmt.Print(vm.ToString(i))
		} else {
			fmt.Print(vm.TypeName(vm.Type(i)))
		}
		if i < nArgs {
			fmt.Print("\t")
		}
	}
	fmt.Println()
	return 0
}

// assert (v [, message])
//
// Calls error if the value of its argument v is false (i.e., nil or false);
// otherwise, returns all its arguments. In case of error, message is the error object;
// when absent, it defaults to "assertion failed!"
func baseAssert(vm api.LuaVM) int {
	nArgs := vm.GetTop()
	b := vm.ToBoolean(-(nArgs - 1)) //判断v的真假
	if b {
		return nArgs
	}
	errMsg := "assertion failed!"
	if s, ok := vm.ToStringX(-(nArgs - 1) + 1); ok {
		errMsg = s
	}
	return vm.Error2(errMsg)
}

// error (message [, level])
//
// Terminates the last protected function called and returns message as the error object. Function error never returns.
func baseError(vm api.LuaVM) int {
	level := vm.OptInteger(2, 1)
	if vm.Type(1) == api.LUAVALUE_STRING && level > 0 {
		//to do
	}
	return vm.Error()
}

// select (index, ···)
//
// 当index是数字时，返回从index位置开始的所有参数
// 当index是字符串"#"时，返回可变参数的总个数
func baseSelect(vm api.LuaVM) int {
	nArgs := vm.GetTop() - 1
	if idx, ok := vm.ToIntegerX(1); ok {
		if idx < 0 {
			idx = int64(nArgs) + idx
		} else if idx > int64(nArgs) {
			idx = int64(nArgs)
		}
		return nArgs - int(idx) + 1
	}
	if str, ok := vm.ToStringX(1); ok && str == "#" {
		vm.PushInteger(int64(nArgs)) //压入可变参数总个数
		return 1
	}
	return vm.Error2("the type of index parameter not support!!")
}

// 用于数组遍历
func baseIPairs(vm api.LuaVM) int {
	vm.PushGoFunction(func(vm api.LuaVM) int {
		i := vm.ToInteger(2) + 1 //index先+1,所以下标从1开始
		vm.PushInteger(i)
		vm.Replace(2)         //保存index
		vm.PushInteger(i)     //压入index
		val := vm.GetTable(1) //压入stack[index]
		if val != api.LUAVALUE_NIL {
			return 2
		}
		return 1
	}, 0)
	vm.PushValue(1)
	vm.PushInteger(0) //index初始化为0
	return 3
}

// 用于map遍历
func basePairs(vm api.LuaVM) int {
	vm.PushGoFunction(baseNext, 0) //推入迭代函数
	vm.PushValue(1)                //推入需遍历的目标
	vm.PushNil()                   //推入迭代器第二个参数，初始值
	return 3
}

// 通用for循环的迭代器函数，迭代器返回2个参数，分别为key,value。如果key==nil则表示没有键值对了
func baseNext(vm api.LuaVM) int {
	vm.SetTop(2)
	if vm.Next(1) {
		return 2
	} else {
		vm.PushNil()
		return 1
	}
}

// load (chunk [, chunkname [, mode [, env]]])
// chunk: 要加载的代码块，可以是字符串或函数，当chunk是一个函数时，它会重复调用这个函数来获取代码片段("字符串")，直到函数返回nil或空字符串为止。
// chunkname: 用于错误消息和调试信息(可选)
// mode: 控制代码类型(可选，"b"二进制，"t"文本，或"bt"两者)
// env: 设置代码块的环境表(可选，Lua 5.2+),如果不设置则默认会使用"_ENV"作为加载代码块的环境。
//
// 用于动态加载和编译Lua代码。它可以将字符串或函数形式的代码编译为可执行的Lua块（chunk）
//
// PS：既然可以直接定义函数然后调用，为什么要使用load函数加载函数作为chunk这种方式？
// 这种设计确实有它的特定用途和优势
// for example:
// -- 当代码需要动态构建/获取时
// local dynamic_code_loader = function()
//
//	-- 代码可能来自文件/网络/用户输入等
//	return get_next_code_chunk()
//
// end
// local func = load(dynamic_code_loader)
// if func then func() end
func baseLoad(vm api.LuaVM) int {
	chunkType := vm.Type(1)
	mode := vm.OptString(3, "bt")
	var chunk string
	var chunkName string
	var env interface{}
	switch chunkType {
	case api.LUAVALUE_STRING: //chunk为字符串
		chunk = vm.ToString(1)
	case api.LUAVALUE_FUNCTION: //chunk为函数，会重复调用这个函数来获取代码片段("字符串")，直到函数返回nil或空字符串为止。
		sb := strings.Builder{}
		vm.PushValue(1) //压入需要load的方法
		vm.Call(0, 1)
		s := vm.ToString(0) //如果Call返回的是nil也会被解析为空字符串
		//获取完整的动态代码
		for s != "" {
			sb.WriteString(s)
			vm.Pop(1)
			vm.PushValue(1)
			vm.Call(0, 1)
			s = vm.ToString(0)
		}
		chunk = sb.String()
	default:
		tool.Fatal(vm, fmt.Sprintf("chunk prarmeter error: not support type[%s]!", chunkType.String()))
	}
	chunkName = vm.OptString(2, chunk)

	if !vm.IsNone(4) { //有env参数,即存在Upvalue捕获变量
		env = vm.ToPointer(4) //获取环境(luatable)
	}

	if _doLoad(vm, []byte(chunk), chunkName, mode, env) == api.LUA_OK {
		return 1
	}
	vm.PushNil()
	vm.PushValue(2) //chunkname用作错误信息
	return 2        //return nil,errmsg
}

// 返回 -1 表示load失败
// 返回 0(LUA_OK) 表示load成功
func _doLoad(vm api.LuaVM, chunk []byte, chunkName string, mode string, env interface{}) int {
	status := -1
	if env == nil {
		status = vm.Load([]byte(chunk), chunkName, mode) //没有环境,默认加载全局环境
	} else {
		status = vm.LoadWithEnv([]byte(chunk), chunkName, mode, env)
	}
	return status
}

// loadfile ([filename [, mode [, env]]])
//
// Similar to load, but gets the chunk from file filename or from the standard input, if no file name is given.
// filename: 要加载的 Lua 文件路径
// mode (可选): 控制文件内容的类型，"b": 只允许二进制代码;"t": 只允许文本代码;"bt": 允许二进制或文本代码（默认值）
// env (可选): 指定代码块的环境表
//
// 用于加载 Lua 代码文件但不立即执行它
func baseLoadFile(vm api.LuaVM) int {
	var filename string
	mode := vm.OptString(2, "bt")
	var f *os.File
	var chunk []byte
	var env interface{}
	if vm.IsNone(1) {
		filename = "stdIn"
		f = os.Stdin
	} else {
		if vm.Type(1) != api.LUAVALUE_STRING {
			tool.Fatal(vm, fmt.Sprintf("filename prarmeter error: not support type[%s]!", vm.Type(1).String()))
		}
		filename = vm.ToString(1)
		file, err := os.Open(filename)
		if err != nil {
			tool.Fatal(vm, fmt.Sprintf("open file[%s] error: ", filename)+err.Error())
		}
		f = file
	}

	if data, err := io.ReadAll(f); err != nil {
		tool.Fatal(vm, fmt.Sprintf("read file[%s] error: ", filename)+err.Error())
	} else {
		chunk = data
	}

	if !vm.IsNone(3) {
		env = vm.ToPointer(3) //获取环境
	}

	if _doLoad(vm, chunk, filename, mode, env) == api.LUA_OK {
		return 1
	}
	vm.PushNil()
	vm.PushString(fmt.Sprintf("loadfile[%s] error", filename)) //chunkname用作错误信息
	return 2                                                   //return nil,errmsg
}

// dofile ([filename])
//
// filename: 要加载并执行的 Lua 文件路径（字符串类型）
//
// 用于加载并"执行"(不同于loadfile) Lua 代码文件的函数，它是一个简单直接的文件执行方式
func baseDoFile(vm api.LuaVM) int {
	if baseLoadFile(vm) == 1 { //baseLoadFile返回1说明执行load成功
		start := vm.GetTop()
		vm.Call(0, api.LUA_MULTRET) //文件调用后的返回值全部返回
		n := 0
		if vm.GetTop() != 0 {
			n = vm.GetTop() - start + 1
		}
		return n
	}
	return 2
}

// pcall (f [, arg1, ···])
//
// Calls function f with the given arguments in protected mode. This means that any error inside f is not propagated; instead, pcall catches the error and returns a status code.
func basePCall(vm api.LuaVM) int {
	nArgs := vm.GetTop() - 1
	status := vm.PCall(nArgs, api.LUA_MULTRET, false)
	vm.PushBoolean(status == api.LUA_OK)
	vm.Insert(-1)
	return vm.GetTop()
}

// xpcall (func, errhandler [, arg1, ···])
//
// status, result1, result2, ... = xpcall(func, errhandler, arg1, arg2, ...)
//
// func: 要保护的函数调用（需要执行的函数）
// errhandler: 错误处理函数，当func出错时会被调用
// arg1, arg2, ...: 传递给func的参数
// 如果func执行成功，后续返回值为func的返回值；如果 func 执行失败，后续返回值为msgh的返回值
//
// xpcall 是 Lua 中一个增强的错误处理函数，它比基础的 pcall 提供了更多的错误处理能力，允许你指定一个自定义的错误处理函数。
func baseXPCall(vm api.LuaVM) int {
	nArgs := vm.GetTop() - 2
	//首先交换func和errhandler在栈中的位置,方便后续方法调用
	vm.PushValue(1) //copy func
	vm.PushValue(2) //copy errhandler
	vm.Replace(1)
	vm.Replace(2)
	//进行pcall调用
	status := vm.PCall(nArgs, api.LUA_MULTRET, true)
	vm.PushBoolean(status == api.LUA_OK)
	vm.Replace(1)
	return vm.GetTop()
}

func baseGetMetaTable(vm api.LuaVM) int {
	if !vm.GetMetaTable(1) {
		vm.PushNil()
	}
	return 1
}

func baseSetMetaTable(vm api.LuaVM) int {
	vm.SetMetaTable(1)
	return 1
}

// rawequal (v1, v2)
// Checks whether v1 is equal to v2, without invoking the __eq metamethod. Returns a boolean.
func baseRawEqual(vm api.LuaVM) int {
	result := vm.RawEqual(1, 2)
	vm.PushBoolean(result)
	return 1
}

// rawlen (v)
// Returns the length of the object v, which must be a table or a string, without invoking the __len metamethod. Returns an integer.
func baseRawLen(vm api.LuaVM) int {
	if len := vm.RawLen(1); len >= 0 {
		vm.PushInteger(int64(len))
		return 1
	}
	return vm.Error2("expected table or string!!")
}

// rawget (table, index)
// Gets the real value of table[index], without invoking the __index metamethod. table must be a table; index may be any value.
func baseRawGet(vm api.LuaVM) int {
	if vm.Type(1) != api.LUAVALUE_TABLE {
		return vm.Error2(fmt.Sprintf("the first parameter was expected to be a table,but got a %s", vm.Type(1).String()))
	}
	vm.CheckAny(2) //确保index参数存在
	vm.RawGet(1)
	return 1
}

// rawset (table, index, value)
// Sets the real value of table[index] to value, without invoking the __newindex metamethod. table must be a table, index any value different from nil and NaN, and value any Lua value.
// This function returns table.
func baseRawSet(vm api.LuaVM) int {
	if vm.Type(1) != api.LUAVALUE_TABLE {
		return vm.Error2(fmt.Sprintf("the first parameter was expected to be a table,but got a %s", vm.Type(1).String()))
	}
	switch vm.Type(2) {
	case api.LUAVALUE_NIL, api.LUAVALUE_NONE:
		return vm.Error2(fmt.Sprintf("the second parameter was %s which not support", vm.Type(2).String()))
	}
	vm.CheckAny(3)
	vm.RawSet(1)
	vm.PushValue(1) //栈顶压入table方便返回
	return 1
}

// type (v)
// Returns the type of its only argument, coded as a string. The possible results of this function are "nil" (a string, not the value nil), "number", "string", "boolean", "table", "function", "thread", and "userdata".
func baseType(vm api.LuaVM) int {
	vm.CheckAny(1)
	typeName := vm.TypeName(vm.Type(1))
	vm.PushString(typeName)
	return 1
}

// tostring (v)
// Receives a value of any type and converts it to a string in a human-readable format. (For complete control of how numbers are converted, use string.format.)
// If the metatable of v has a __tostring field, then tostring calls the corresponding value with v as argument, and uses the result of the call as its result.
func baseToString(vm api.LuaVM) int {
	vm.CheckAny(1)
	str := vm.ToString2(1)
	vm.PushString(str)
	return 1
}

// tonumber (e [, base])
//
// value: 需要转换的值（字符串或数字）
// base (可选): 进制基数（2-36），默认为10进制
// 返回值：转换成功返回对应的数字；转换失败则返回 nil
//
// tonumber 是Lua中用于将值转换为数字的核心函数
func baseToNumber(vm api.LuaVM) int {
	vm.CheckAny(1)
	var base int64 = 10
	if !vm.IsNoneOrNil(2) { //如果没有base参数,默认转化为10进制
		base = vm.ToInteger(2)
	}

	switch vm.Type(1) {
	case api.LUAVALUE_NUMBER:
		vm.PushValue(1)
	case api.LUAVALUE_STRING:
		str := vm.ToString(1)
		if str == "" {
			vm.PushNil()
			break
		}
		_toNumber(vm, str, int(base)) //默认10进制
	default:
		vm.PushNil()
	}
	return 1
}

func _toNumber(vm api.LuaVM, str string, base int) {
	str = strings.Trim(str, " ")
	if len(strings.Split(str, ".")) == 2 || strings.Contains(str, "e") { //判断是否为小数形式或指数形式
		f, err := strconv.ParseFloat(str, 64)
		if err != nil {
			vm.PushNil()
			tool.Debug(fmt.Sprintf("parse %s to float error : ", str) + err.Error())
		} else {
			vm.PushFloat(f)
		}
		return
	}
	//执行整数转换流程
	i, err := strconv.ParseInt(str, base, 64)
	if err != nil {
		vm.PushNil()
		tool.Debug(fmt.Sprintf("parse %s to int error : ", str) + err.Error())
	} else {
		vm.PushInteger(i)
	}
}
