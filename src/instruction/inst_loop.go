package instruction

import "nskbz.cn/lua/api"

func forPrep(i Instruction, vm api.LuaVM) {
	a, sbx := i.AsBx()
	a += 1
	vm.PushValue(a)
	vm.PushValue(a + 2)
	vm.Arith(api.ArithOp_SUB)
	vm.Replace(a)
	vm.AddPC(sbx)
}

func forLoop(i Instruction, vm api.LuaVM) {
	a, sbx := i.AsBx()
	a += 1
	vm.PushValue(a + 2)
	vm.PushValue(a)
	vm.Arith(api.ArithOp_ADD)
	vm.Replace(a)
	idx1, idx2 := 0, 0
	if vm.ToFloat(a+2) < 0 {
		//step<0
		idx1 = a + 1
		idx2 = a
	} else {
		//step>=0
		idx1 = a
		idx2 = a + 1
	}
	if !vm.Compare(idx1, idx2, api.CompareOp_LE) {
		return
	}
	vm.Copy(a, a+3)
	vm.AddPC(sbx)
}
