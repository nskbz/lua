package main

import (
	"fmt"
	"io"
	"os"

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
	vm := state.New(20)
	vm.Load(datas, "test.lua", "b")
	vm.Call(0, 0)
}
