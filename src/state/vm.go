package state

import (
	"math"

	"nskbz.cn/lua/api"
	"nskbz.cn/lua/binchunk"
)

type luaVM struct {
	*luaState
	proto *binchunk.Prototype
	pc    int
}

func NewVM(stackSize int, proto *binchunk.Prototype) api.LuaVM {
	return &luaVM{
		luaState: NewState(stackSize + 8),
		proto:    proto,
		pc:       0,
	}
}

func (vm *luaVM) PC() int {
	return vm.pc
}

func (vm *luaVM) AddPC(n int) {
	add := vm.pc + n
	if add < 0 || add > math.MaxInt {
		panic("pc overflow")
	}
	vm.pc = add
}

func (vm *luaVM) Fetch() uint32 {
	code := vm.proto.Codes[vm.pc]
	vm.AddPC(1)
	return code
}

func (vm *luaVM) GetConst(idx int) {
	if idx < 0 || idx >= len(vm.proto.Constants) {
		panic("constant's index out of range")
	}
	c := vm.proto.Constants[idx]
	vm.luaState.stack.push(c)
}

func (vm *luaVM) GetRK(arg int) {
	if arg > 0xFF {
		//最高位不为0表示常量表索引
		vm.GetConst(arg & 0xFF)
	} else {
		//最高位为0表示寄存器索引
		vm.PushValue(arg + 1)
	}
}
