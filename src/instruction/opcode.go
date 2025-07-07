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
	{0, ArgR, ArgN, ArgN, IABx, "LOADKX  ", loadKX},     // R(A) := Kst(extra arg) ，同EXTRAARG搭配使用
	{0, ArgR, ArgU, ArgU, IABC, "LOADBOOL", loadBool},   // R(A) := (bool)B; if (C) pc++
	{0, ArgR, ArgU, ArgN, IABC, "LOADNIL ", loadNil},    // R(A), R(A+1), ..., R(A+B) := nil
	{0, ArgR, ArgU, ArgN, IABC, "GETUPVAL", getUpval},   // R(A) := UpValue[B]
	{0, ArgR, ArgU, ArgRK, IABC, "GETTABUP", getTabUp},  //R(A) := UpValue[B][RK(C)]
	{0, ArgR, ArgR, ArgRK, IABC, "GETTABLE", getTable},  //R(A) := R(B)[RK(C)]
	{0, ArgU, ArgK, ArgK, IABC, "SETTABUP", setTabUp},   // UpValue[A][RK(B)] := RK(C)
	{0, ArgR, ArgU, ArgN, IABC, "SETUPVAL", setUpval},   // UpValue[B] := R(A)
	{0, ArgU, ArgRK, ArgRK, IABC, "SETTABLE", setTable}, //R(A)[RK(B)] := RK(C)
	{0, ArgR, ArgU, ArgU, IABC, "NEWTABLE", newTable},   //R(A) := {} (size = B,C) ,B为数组的长度,C为字典的大小
	{0, ArgR, ArgR, ArgK, IABC, "SELF    ", self},       // R(A+1) := R(B); R(A) := R(B)[RK(C)]
	{0, ArgR, ArgRK, ArgRK, IABC, "ADD     ", add},      // R(A) := RK(B) + RK(C)
	{0, ArgR, ArgRK, ArgRK, IABC, "SUB     ", sub},      // R(A) := RK(B) - RK(C)
	{0, ArgR, ArgRK, ArgRK, IABC, "MUL     ", mul},      // R(A) := RK(B) * RK(C)
	{0, ArgR, ArgRK, ArgRK, IABC, "MOD     ", mod},      // R(A) := RK(B) % RK(C)
	{0, ArgR, ArgRK, ArgRK, IABC, "POW     ", pow},      // R(A) := RK(B) ^ RK(C)
	{0, ArgR, ArgRK, ArgRK, IABC, "DIV     ", div},      // R(A) := RK(B) / RK(C)
	{0, ArgR, ArgRK, ArgRK, IABC, "IDIV    ", idiv},     // R(A) := RK(B) // RK(C)
	{0, ArgR, ArgRK, ArgRK, IABC, "BAND    ", and},      // R(A) := RK(B) & RK(C)
	{0, ArgR, ArgRK, ArgRK, IABC, "BOR     ", or},       // R(A) := RK(B) | RK(C)
	{0, ArgR, ArgRK, ArgRK, IABC, "BXOR    ", xor},      // R(A) := RK(B) ~ RK(C)
	{0, ArgR, ArgRK, ArgRK, IABC, "SHL     ", shl},      // R(A) := RK(B) << RK(C)
	{0, ArgR, ArgRK, ArgRK, IABC, "SHR     ", shr},      // R(A) := RK(B) >> RK(C)
	{0, ArgR, ArgR, ArgN, IABC, "UNM     ", opposite},   // R(A) := -R(B)
	{0, ArgR, ArgR, ArgN, IABC, "BNOT    ", bnot},       // R(A) := ~R(B)
	{0, ArgR, ArgR, ArgN, IABC, "NOT     ", not},        // R(A) := not R(B)
	{0, ArgR, ArgR, ArgN, IABC, "LEN     ", valLen},     // R(A) := length of R(B)
	{0, ArgR, ArgR, ArgR, IABC, "CONCAT  ", concat},     // R(A) := R(B).. ... ..R(C)
	{0, ArgU, ArgU, ArgN, IAsBx, "JMP     ", jump},      // pc+=sBx; if (A) close all upvalues >= R(A - 1)
	{1, ArgU, ArgK, ArgK, IABC, "EQ      ", eq},         // if ((RK(B) == RK(C)) ~= A) then pc++
	{1, ArgU, ArgK, ArgK, IABC, "LT      ", lt},         // if ((RK(B) <  RK(C)) ~= A) then pc++
	{1, ArgU, ArgK, ArgK, IABC, "LE      ", le},         // if ((RK(B) <= RK(C)) ~= A) then pc++
	{1, ArgU, ArgN, ArgU, IABC, "TEST    ", test},       // if not (R(A) == C) then pc++
	{1, ArgR, ArgR, ArgU, IABC, "TESTSET ", testset},    // if (R(B) == C) then R(A) := R(B) else pc++
	{0, ArgR, ArgU, ArgU, IABC, "CALL    ", call},       // R(A), ... ,R(A+C-2) := R(A)(R(A+1), ... ,R(A+B-1))
	{0, ArgR, ArgU, ArgN, IABC, "TAILCALL", tailcall},   // return R(A)(R(A+1), ... ,R(A+B-1))
	{0, ArgU, ArgU, ArgN, IABC, "RETURN  ", luaReturn},  // return R(A),...,R(A+B-2)
	{0, ArgR, ArgU, ArgN, IAsBx, "FORLOOP ", forLoop},   // R(A)+=R(A+2); if R(A) <?= R(A+1) then { pc+=sBx; R(A+3)=R(A) }
	{0, ArgR, ArgU, ArgN, IAsBx, "FORPREP ", forPrep},   // R(A)-=R(A+2); pc+=sBx
	{0, ArgR, ArgN, ArgU, IABC, "TFORCALL", tForCall},   // R(A+3), ... ,R(A+2+C) := R(A)(R(A+1), R(A+2))
	{0, ArgR, ArgU, ArgN, IAsBx, "TFORLOOP", tForLoop},  // if R(A+1) ~= nil then { R(A)=R(A+1); pc += sBx }
	{0, ArgU, ArgU, ArgU, IABC, "SETLIST ", setList},    // R(A)[(C-1)*FPF+i] := R(A+i), 1 <= i <= B
	{0, ArgR, ArgU, ArgN, IABx, "CLOSURE ", closure},    //R(A) := closure(KPROTO[Bx])
	{0, ArgR, ArgU, ArgN, IABC, "VARARG  ", vararg},     // R(A), R(A+1), ..., R(A+B-2) = vararg
	// 扩展其他指令的参数空间，支持更大的常量表索引
	// forexample:
	//
	//	local t = {
	//	    [1] = "very large constant table entry",
	//	    -- ... 数百个条目 ...
	//	    [1000] = "needs extraarg"
	//	}
	//
	// 编译器可能生成类似：
	// LOADK    R1  K1     ; 尝试加载常规常量
	// EXTRAARG 0x12345678 ; 提供额外的高位索引
	//
	// EXTRAARG指令与其他指令搭配使用的且是第二顺位的指令，
	// 所以可以在第一顺位的指令处理中就处理了，因此不需要额外的处理函数
	{0, ArgU, ArgU, ArgU, IAx, "EXTRAARG", nil},
}
