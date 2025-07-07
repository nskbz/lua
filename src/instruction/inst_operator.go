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

// EQ: if ((RK(B) == RK(C)) ~= A) then pc++
//
// forexample:
// local cond = (1 == 2)
//
//		if cond then
//	    print("true")
//		else
//	    print("false")
//
// end
// ==============================
// EQ    0 1 2      ; 比较 1 和 2
// LOADBOOL 0 0 1      ; R0 = false, 并跳过下一条指令
// JMP    2         ; 跳转到 else 部分
// ...              ; true 分支的代码
// JMP    3         ; 跳过 else 部分
// ...              ; false 分支的代码
func eq(i Instruction, vm api.LuaVM) { doCompare(i, vm, api.CompareOp_EQ) }

// LT: if ((RK(B) <  RK(C)) ~= A) then pc++
func lt(i Instruction, vm api.LuaVM) { doCompare(i, vm, api.CompareOp_LT) }

// LE: if ((RK(B) <= RK(C)) ~= A) then pc++
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

// TEST 更轻量：只需检查单个值的布尔状态
// EQ 更复杂：需要比较两个值的类型和内容
// 对于代码 if a then print("true") end：
// instructions:
// TEST    0 0 1    ; 测试寄存器0是否为真
// JMP     1        ; 如果假则跳过
// ...
// 对于代码 if a == b then print("equal") end：
// instructions:
// EQ      0 1 2    ; 比较寄存器1和2
// JMP     1        ; 如果不相等则跳过
//
// if not (R(A) == C) then pc++
// 只判断R(A)的布尔值（只有nil和false为假）
func test(i Instruction, vm api.LuaVM) {
	a, _, c := i.ABC()
	a += 1
	if vm.ToBoolean(a) != (c != 0) {
		vm.AddPC(1)
	}
}
