package instruction

const (
	OP_MOVE = iota
	OP_LOADK
	OP_LOADKX
	OP_LOADBOOL
	OP_LOADNIL
	OP_GETUPVAL
	OP_GETTABUP
	OP_GETTABLE
	OP_SETTABUP
	OP_SETUPVAL
	OP_SETTABLE
	OP_NEWTABLE
	OP_SELF
	OP_ADD
	OP_SUB
	OP_MUL
	OP_MOD
	OP_POW
	OP_DIV
	OP_IDIV
	OP_BAND
	OP_BOR
	OP_BXOR
	OP_SHL
	OP_SHR
	OP_UNM
	OP_BNOT
	OP_NOT
	OP_LEN
	OP_CONCAT
	OP_JMP
	OP_EQ
	OP_LT
	OP_LE
	OP_TEST
	OP_TESTSET
	OP_CALL
	OP_TAILCALL
	OP_RETURN
	OP_FORLOOP
	OP_FORPREP
	OP_TFORCALL
	OP_TFORLOOP
	OP_SETLIST
	OP_CLOSURE
	OP_VARARG
	OP_EXTRAARG
)

// 数组索引与上面一一对应
var instructions = []opcode{
	//	是否为逻辑判断指令|	 A	| 	B   |	 C	|指令类型| 指令名
	//ABC模式下每个参数9bit最高位表示该参数是否为常量表索引
	{0, ArgR, ArgR, ArgN, IABC, "MOVE    ", move},       // R(A) := R(B)
	{0, ArgR, ArgK, ArgN, IABx, "LOADK   ", loadK},      // R(A) := Kst(Bx)
	{0, ArgR, ArgN, ArgN, IABx, "LOADKX  ", loadKX},     // R(A) := Kst(extra arg)
	{0, ArgR, ArgU, ArgU, IABC, "LOADBOOL", loadBool},   // R(A) := (bool)B; if (C) pc++
	{0, ArgR, ArgU, ArgN, IABC, "LOADNIL ", loadNil},    // R(A), R(A+1), ..., R(A+B) := nil
	{0, ArgR, ArgU, ArgN, IABC, "GETUPVAL", getUpval},   // R(A) := UpValue[B]
	{0, ArgR, ArgU, ArgRK, IABC, "GETTABUP", getTabUp},  //R(A) := UpValue[B][RK(C)]
	{0, ArgR, ArgR, ArgRK, IABC, "GETTABLE", getTable},  //R(A) := R(B)[RK(C)]
	{0, ArgU, ArgK, ArgK, IABC, "SETTABUP", setTabUp},   // UpValue[A][RK(B)] := RK(C)
	{0, ArgR, ArgU, ArgN, IABC, "SETUPVAL", setUpval},   // UpValue[B] := R(A)
	{0, ArgU, ArgRK, ArgRK, IABC, "SETTABLE", setTable}, //R(A)[RK(B)] := RK(C)
	{0, ArgR, ArgU, ArgU, IABC, "NEWTABLE", newTable},   //R(A) := {} (size = B,C)
	{0, ArgR, ArgR, ArgK, IABC, "SELF    ", self},       // R(A+1) := R(B); R(A) := R(B)[RK(C)]
	{0, ArgR, ArgRK, ArgRK, IABC, "ADD     ", add},      // R(A) := RK(B) + RK(C)
	{0, ArgR, ArgRK, ArgRK, IABC, "SUB     ", sub},
	{0, ArgR, ArgRK, ArgRK, IABC, "MUL     ", mul},
	{0, ArgR, ArgRK, ArgRK, IABC, "MOD     ", mod},
	{0, ArgR, ArgRK, ArgRK, IABC, "POW     ", pow},
	{0, ArgR, ArgRK, ArgRK, IABC, "DIV     ", div},
	{0, ArgR, ArgRK, ArgRK, IABC, "IDIV    ", idiv},
	{0, ArgR, ArgRK, ArgRK, IABC, "BAND    ", and},
	{0, ArgR, ArgRK, ArgRK, IABC, "BOR     ", or},
	{0, ArgR, ArgRK, ArgRK, IABC, "BXOR    ", xor},
	{0, ArgR, ArgRK, ArgRK, IABC, "SHL     ", shl},
	{0, ArgR, ArgRK, ArgRK, IABC, "SHR     ", shr},
	{0, ArgR, ArgR, ArgN, IABC, "UNM     ", opposite}, // R(A) := -R(B)
	{0, ArgR, ArgR, ArgN, IABC, "BNOT    ", bnot},     // R(A) := ~R(B)
	{0, ArgR, ArgR, ArgN, IABC, "NOT     ", not},      // R(A) := not R(B)
	{0, ArgR, ArgR, ArgN, IABC, "LEN     ", valLen},
	{0, ArgR, ArgR, ArgR, IABC, "CONCAT  ", concat},
	{0, ArgU, ArgU, ArgN, IAsBx, "JMP     ", jump}, // pc+=sBx; if (A) close all upvalues >= R(A - 1)
	{1, ArgU, ArgK, ArgK, IABC, "EQ      ", eq},    // if ((RK(B) == RK(C)) ~= A) then pc++
	{1, ArgU, ArgK, ArgK, IABC, "LT      ", lt},    // if ((RK(B) <  RK(C)) ~= A) then pc++
	{1, ArgU, ArgK, ArgK, IABC, "LE      ", le},    // if ((RK(B) <= RK(C)) ~= A) then pc++
	{1, ArgU, ArgN, ArgU, IABC, "TEST    ", test},
	{1, ArgR, ArgR, ArgU, IABC, "TESTSET ", testset},
	{0, ArgR, ArgU, ArgU, IABC, "CALL    ", call},      // R(A), ... ,R(A+C-2) := R(A)(R(A+1), ... ,R(A+B-1))
	{0, ArgR, ArgU, ArgU, IABC, "TAILCALL", tailcall},  // return R(A)(R(A+1), ... ,R(A+B-1))
	{0, ArgU, ArgU, ArgN, IABC, "RETURN  ", luaReturn}, // return R(A),...,R(A+B-2)
	{0, ArgR, ArgU, ArgN, IAsBx, "FORLOOP ", forLoop},
	{0, ArgR, ArgU, ArgN, IAsBx, "FORPREP ", forPrep},
	{0, ArgU, ArgN, ArgU, IABC, "TFORCALL", nil},
	{0, ArgR, ArgU, ArgN, IAsBx, "TFORLOOP", nil},
	{0, ArgU, ArgU, ArgU, IABC, "SETLIST ", setList}, // R(A)[(C-1)*FPF+i] := R(A+i), 1 <= i <= B
	{0, ArgR, ArgU, ArgN, IABx, "CLOSURE ", closure}, //R(A) := closure(KPROTO[Bx])
	{0, ArgR, ArgU, ArgN, IABC, "VARARG  ", vararg},  // R(A), R(A+1), ..., R(A+B-2) = vararg
	{0, ArgU, ArgU, ArgU, IAx, "EXTRAARG", nil},
}
