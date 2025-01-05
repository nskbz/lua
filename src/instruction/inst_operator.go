package instruction

import "nskbz.cn/lua/api"

func doDualArith(i Instruction, vm api.LuaVM, op api.ArithOp) {
	a, b, c := i.ABC()
	vm.GetRK(b)
	vm.GetRK(c)
	vm.Arith(op)
	vm.Replace(a + 1)
}

func doUnaryArith(i Instruction, vm api.LuaVM, op api.ArithOp) {
	a, b, _ := i.ABC()
	vm.PushValue(b + 1)
	vm.Arith(op)
	vm.Replace(a + 1)
}

func add(i Instruction, vm api.LuaVM)      { doDualArith(i, vm, api.ArithOp_ADD) }
func sub(i Instruction, vm api.LuaVM)      { doDualArith(i, vm, api.ArithOp_SUB) }
func mul(i Instruction, vm api.LuaVM)      { doDualArith(i, vm, api.ArithOp_MUL) }
func mod(i Instruction, vm api.LuaVM)      { doDualArith(i, vm, api.ArithOp_MOD) }
func pow(i Instruction, vm api.LuaVM)      { doDualArith(i, vm, api.ArithOp_POW) }
func div(i Instruction, vm api.LuaVM)      { doDualArith(i, vm, api.ArithOp_DIV) }
func idiv(i Instruction, vm api.LuaVM)     { doDualArith(i, vm, api.ArithOp_IDIV) }
func and(i Instruction, vm api.LuaVM)      { doDualArith(i, vm, api.ArithOp_AND) }
func or(i Instruction, vm api.LuaVM)       { doDualArith(i, vm, api.ArithOp_OR) }
func xor(i Instruction, vm api.LuaVM)      { doDualArith(i, vm, api.ArithOp_XOR) }
func shl(i Instruction, vm api.LuaVM)      { doDualArith(i, vm, api.ArithOp_SHL) }
func shr(i Instruction, vm api.LuaVM)      { doDualArith(i, vm, api.ArithOp_SHR) }
func opposite(i Instruction, vm api.LuaVM) { doUnaryArith(i, vm, api.ArithOp_OPPOSITE) }
func bnot(i Instruction, vm api.LuaVM)     { doUnaryArith(i, vm, api.ArithOp_NOT) }

func doCompare(i Instruction, vm api.LuaVM, op api.CompareOp) {
	a, b, c := i.ABC()
	vm.GetRK(b)
	vm.GetRK(c)
	bc := vm.Compare(-1, 0, op)
	if bc != (a != 0) {
		vm.AddPC(1)
	}
	vm.Pop(2)
}

func eq(i Instruction, vm api.LuaVM) { doCompare(i, vm, api.CompareOp_EQ) }
func lt(i Instruction, vm api.LuaVM) { doCompare(i, vm, api.CompareOp_LT) }
func le(i Instruction, vm api.LuaVM) { doCompare(i, vm, api.CompareOp_LE) }

func not(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	vm.PushBoolean(!vm.ToBoolean(b + 1))
	vm.Replace(a + 1)
}

func testset(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	b += 1
	a += 1
	if vm.ToBoolean(b) == (c != 0) {
		vm.Copy(b, a)
	} else {
		vm.AddPC(1)
	}
}

func test(i Instruction, vm api.LuaVM) {
	a, _, c := i.ABC()
	a += 1
	if vm.ToBoolean(a) != (c != 0) {
		vm.AddPC(1)
	}
}
