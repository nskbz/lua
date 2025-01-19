package instruction

import "nskbz.cn/lua/api"

// R(A) := UpValue[B][RK(C)]
func getTabUp(i Instruction, vm api.LuaVM) {
	a, _, c := i.ABC()
	a += 1

	vm.PushGlobalTable()
	vm.GetRK(c)
	vm.GetTable(-1)
	vm.Replace(a)
	vm.Pop(1) //弹出global
}
