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
	vm.Load(datas, "test.lua", "b")
	vm.Call(0, 0)
}

func print(ls api.LuaVM) int {
	nArgs := ls.GetTop()
	for i := 1; i <= nArgs; i++ {
		if ls.IsBoolean(i) {
			fmt.Printf("%t", ls.ToBoolean(i))
		} else if ls.IsString(i) {
			fmt.Print(ls.ToString(i))
		} else {
			fmt.Print(ls.TypeName(ls.Type(i)))
		}
		if i < nArgs {
			fmt.Print("\t")
		}
	}
	fmt.Println()
	return 0
}
