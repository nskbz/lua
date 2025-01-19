package binchunk

const (
	LUA_SIGNATURE    = "\x1bLua"
	LUAC_VERSION     = 0x53
	LUAC_FORMAT      = 0
	LUAC_DATA        = "\x19\x93\r\n\x1a\n"
	CINT_SIZE        = 4
	CSIZET_SIZE      = 8
	INSTRUCTION_SIZE = 4
	LUA_INTEGER_SIZE = 8
	LUA_FLOAT_SIZE   = 8
	LUAC_INT         = 0x5678
	LUAC_FLOAT       = 370.5
)

type binChunk struct {
	header
	sizeUpvalues byte
	mainFunc     interface{}
}

type header struct {
	signature       [4]byte //不是数据类型所以不管小端还是大端都应该一样顺序所以采取单个byte的4B
	version         byte
	format          byte
	luacData        [6]byte
	cintSize        byte
	csizetSize      byte
	instructionSize byte
	luaIntSize      byte
	luaFloatSize    byte
	luacInt         int64   //用于检测大小端方式
	luacFloat       float64 //用于检测浮点数格式
}

const (
	TAG_NIL       = 0x00
	TAG_BOOLEAN   = 0x01
	TAG_FLOAT     = 0x03
	TAG_INTEGER   = 0x13
	TAG_SHORT_STR = 0x04
	TAG_LONG_STR  = 0x14
)

type Prototype struct {
	Source          string
	LineStart       uint32 //启始行号
	LineEnd         uint32 //结束行号
	NumParams       byte   //固定参数个数
	IsVararg        byte
	MaxRegisterSize byte          //寄存器个数
	Codes           []uint32      //指令表
	Constants       []interface{} //常量表
	Upvalues        []Upvalue
	Protos          []*Prototype //子函数列表
	LineInfo        []uint32     //行号表 记录每条指令对应源代码中的行号
	LocVars         []LocVar
	UpvalueNames    []string //与Upvalues一一对应
}

type Upvalue struct {
	Instack byte
	Idx     byte
}

type LocVar struct {
	VarName string
	StartPC uint32
	EndPC   uint32
}
