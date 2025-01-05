package instruction

import "nskbz.cn/lua/api"

func loadNil(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	vm.PushNil()
	for i := a + 1; i <= a+b; i++ {
		vm.Copy(-1, i)
	}
	vm.Pop(1)
}

func loadBool(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	vm.PushBoolean(b != 0)
	vm.Replace(a + 1)
	if c != 0 {
		vm.AddPC(1)
	}
}

func loadK(i Instruction, vm api.LuaVM) {
	a, bx := i.ABx()
	vm.GetConst(bx)
	vm.Replace(a + 1)
}

func loadKX(i Instruction, vm api.LuaVM) {
	a, _ := i.ABx()
	ex := Instruction(vm.Fetch())
	ax := ex.Ax()
	vm.GetConst(ax)
	vm.Replace(a + 1)
}
