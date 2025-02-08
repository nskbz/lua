package main

import (
	"fmt"
	"io"
	"os"

	"nskbz.cn/lua/api"
	"nskbz.cn/lua/state"
)

func main() {
	f, err := os.Open("../luac.out")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	datas, err := io.ReadAll(f)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	vm := state.New()
	vm.Register("print", print)
	vm.Register("getmetatable", getMetaTable)
	vm.Register("setmetatable", setMetaTable)
	vm.Register("next", next)
	vm.Register("pairs", pairs)
	vm.Register("ipairs", iPairs)
	vm.Load(datas, "test.lua", "b")
	vm.Call(0, 0)
}

func print(vm api.LuaVM) int {
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

func getMetaTable(vm api.LuaVM) int {
	if !vm.GetMetaTable(1) {
		vm.PushNil()
	}
	return 1
}

func setMetaTable(vm api.LuaVM) int {
	vm.SetMetaTable(1)
	return 1
}

func next(vm api.LuaVM) int {
	vm.SetTop(2)
	if vm.Next(1) {
		return 2
	} else {
		vm.PushNil()
		return 1
	}
}

// 用于map遍历
func pairs(vm api.LuaVM) int {
	vm.PushGoFunction(next, 0)
	vm.PushValue(1)
	vm.PushNil()
	return 3
}

// 用于数组遍历
func iPairs(vm api.LuaVM) int {
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
