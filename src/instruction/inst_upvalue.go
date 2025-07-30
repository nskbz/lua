package instruction

import "nskbz.cn/lua/api"

// R(A) := UpValue[B]
func getUpval(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1 //Upval索引也是1开始
	uvIdx := vm.UpvalueIndex(b)
	vm.Copy(uvIdx, a)
}

// UpValue[B] := R(A)
func setUpval(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1
	vm.PushValue(a)
	uvIdx := vm.UpvalueIndex(b)
	vm.Replace(uvIdx)
}

// R(A) := UpValue[B][RK(C)]
func getTabUp(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1

	vm.GetRK(c)
	vm.GetTable(vm.UpvalueIndex(b))
	vm.Replace(a)
}

// UpValue[A][RK(B)] := RK(C)
func setTabUp(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a += 1
	vm.GetRK(b)
	vm.GetRK(c)
	vm.SetTable(vm.UpvalueIndex(a))
}
