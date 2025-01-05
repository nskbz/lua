package instruction

import "nskbz.cn/lua/api"

func move(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	vm.Copy(a+1, b+1)
}

func jump(i Instruction, vm api.LuaVM) {
	a, bx := i.AsBx()
	vm.AddPC(bx)
	if a != 0 {
		panic("todo")
	}
}

func valLen(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	vm.Len(b + 1)
	vm.Replace(a + 1)
}

func concat(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1
	c += 1
	n := c - b + 1
	vm.CheckStack(n)
	for i := b; i <= c; i++ {
		vm.PushValue(i)
	}
	vm.Concat(n)
	vm.Replace(a)
}
