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
