package vm

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
	//t| A	| B   |	 C	|指令类型| 指令名
	{0, ArgR, ArgR, ArgN, IABC, "MOVE    "}, //将B中数据移动至A
	{0, ArgR, ArgK, ArgN, IABx, "LOADK   "},
	{0, ArgR, ArgN, ArgN, IABx, "LOADKX  "},
	{0, ArgR, ArgU, ArgU, IABC, "LOADBOOL"},
	{0, ArgR, ArgU, ArgN, IABC, "LOADNIL "},
	{0, ArgR, ArgU, ArgN, IABC, "GETUPVAL"},
	{0, ArgR, ArgU, ArgK, IABC, "GETTABUP"},
	{0, ArgR, ArgR, ArgK, IABC, "GETTABLE"},
	{0, ArgU, ArgK, ArgK, IABC, "SETTABUP"},
	{0, ArgU, ArgU, ArgN, IABC, "SETUPVAL"},
	{0, ArgU, ArgK, ArgK, IABC, "SETTABLE"},
	{0, ArgR, ArgU, ArgU, IABC, "NEWTABLE"},
	{0, ArgR, ArgR, ArgK, IABC, "SELF    "},
	{0, ArgR, ArgK, ArgK, IABC, "ADD     "},
	{0, ArgR, ArgK, ArgK, IABC, "SUB     "},
	{0, ArgR, ArgR, ArgK, IABC, "MUL     "},
	{0, ArgR, ArgK, ArgK, IABC, "MOD     "},
	{0, ArgR, ArgK, ArgK, IABC, "POW     "},
	{0, ArgR, ArgK, ArgK, IABC, "DIV     "},
	{0, ArgR, ArgK, ArgK, IABC, "IDIV    "},
	{0, ArgR, ArgK, ArgK, IABC, "BAND    "},
	{0, ArgR, ArgK, ArgK, IABC, "BOR     "},
	{0, ArgR, ArgK, ArgK, IABC, "BXOR    "},
	{0, ArgR, ArgK, ArgK, IABC, "SHL     "},
	{0, ArgR, ArgK, ArgK, IABC, "SHR     "},
	{0, ArgR, ArgR, ArgN, IABC, "UNM     "},
	{0, ArgR, ArgR, ArgN, IABC, "BNOT    "},
	{0, ArgR, ArgR, ArgN, IABC, "NOT     "},
	{0, ArgR, ArgR, ArgN, IABC, "LEN     "},
	{0, ArgR, ArgR, ArgR, IABC, "CONCAT  "},
	{0, ArgU, ArgJ, ArgN, IAsBx, "JMP     "},
	{1, ArgU, ArgK, ArgK, IABC, "EQ      "},
	{1, ArgU, ArgK, ArgK, IABC, "LT      "},
	{1, ArgU, ArgK, ArgK, IABC, "LE      "},
	{1, ArgU, ArgN, ArgU, IABC, "TEST    "},
	{1, ArgR, ArgR, ArgU, IABC, "TESTSET "},
	{0, ArgR, ArgU, ArgU, IABC, "CALL    "},
	{0, ArgR, ArgU, ArgU, IABC, "TAILCALL"},
	{0, ArgU, ArgU, ArgN, IABC, "RETURN  "},
	{0, ArgR, ArgJ, ArgN, IAsBx, "FORLOOP "},
	{0, ArgR, ArgJ, ArgN, IAsBx, "FORPREP "},
	{0, ArgU, ArgN, ArgU, IABC, "TFORCALL"},
	{0, ArgR, ArgJ, ArgN, IAsBx, "TFORLOOP"},
	{0, ArgU, ArgU, ArgU, IABC, "SETLIST "},
	{0, ArgR, ArgU, ArgN, IABx, "CLOSURE "},
	{0, ArgR, ArgU, ArgN, IABC, "VARARG  "},
	{0, ArgU, ArgU, ArgU, IAx, "EXTRAARG"},
}
