package instruction

import (
	"nskbz.cn/lua/api"
)

// R(A), R(A+1), ..., R(A+B) := nil
func loadNil(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	vm.PushNil()
	for i := a; i <= a+b; i++ {
		vm.Copy(0, i)
	}
	vm.Pop(1)
}

// R(A) := (bool)B; if (C) pc++
func loadBool(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	vm.PushBoolean(b != 0)
	vm.Replace(a + 1)
	if c != 0 {
		vm.AddPC(1)
	}
}

// R(A) := Kst(Bx)
func loadK(i Instruction, vm api.LuaVM) {
	a, bx := i.ABx()
	vm.GetConst(bx)
	vm.Replace(a + 1)
}

// R(A) := Kst(extra arg)
func loadKX(i Instruction, vm api.LuaVM) {
	a, _ := i.ABx()
	//获取EXTRAARG指令
	ex := Instruction(vm.Fetch())
	ax := ex.Ax()
	vm.GetConst(ax)
	vm.Replace(a + 1)
}
