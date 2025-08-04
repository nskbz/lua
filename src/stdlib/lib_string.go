package stdlib

import (
	"strings"

	"nskbz.cn/lua/api"
)

var stringFuncs map[string]api.GoFunc = map[string]api.GoFunc{
	"len":   stringLen,
	"upper": stringUpper,
	"lower": stringLower,
}

func OpenStringLib(vm api.LuaVM) int {
	vm.NewTable()
	for k, v := range stringFuncs {
		vm.PushGoFunction(v, 0)
		vm.SetField(-1, k)
	}
	//设置元表以支持 'string':len() 语法
	vm.NewTable()    //创建元表
	vm.PushValue(-1) //stringFuncs table
	vm.SetField(-1, "__index")
	vm.PushString("jack") //推入一个string类型用于下面绑定string类型的元表
	vm.PushValue(-1)      //meta_table_copy
	vm.SetMetaTable(-1)
	vm.Pop(2) //pop "jack" meta_table
	return 1
}

// string.len(s)
// 用于获取字符串的字节长度（byte length）
func stringLen(vm api.LuaVM) int {
	vm.CheckString(1)
	str := vm.ToString(1)
	vm.PushInteger(int64(len(str)))
	return 1
}

func stringUpper(vm api.LuaVM) int {
	vm.CheckString(1)
	str := vm.ToString(1)
	vm.PushString(strings.ToUpper(str))
	return 1
}

func stringLower(vm api.LuaVM) int {
	vm.CheckString(1)
	str := vm.ToString(1)
	vm.PushString(strings.ToLower(str))
	return 1
}
