package binchunk

import (
	"fmt"
	"strconv"

	"nskbz.cn/lua/instruction"
	"nskbz.cn/lua/tool"
)

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
	LineStart       uint32                    //启始行号
	LineEnd         uint32                    //结束行号
	NumParams       byte                      //固定参数个数
	IsVararg        byte                      //是否有可变参数
	MaxRegisterSize byte                      //寄存器个数
	Codes           []instruction.Instruction //指令表
	Constants       []interface{}             //常量表
	Upvalues        []Upvalue
	Protos          []*Prototype //子函数列表
	LineInfo        []uint32     //行号表 记录每条指令对应源代码中的行号
	LocVars         []LocVar
	UpvalueNames    []string //与Upvalues一一对应
}

type Upvalue struct {
	Instack byte //捕获的变量是否在栈上：1(在，表示当前函数的局部变量) ; 0(不在，表示更外围函数的局部变量)
	Idx     byte //instack==1表示栈上索引，instack==0表示外围函数的upvalue表索引
}

type LocVar struct {
	VarName   string
	StartLine uint32
	EndLine   uint32
}

// to do
func (p *Prototype) ToBytes() []byte {
	return nil
}

type prototypeInfo struct {
	Source          string
	LineStart       uint32   //启始行号
	LineEnd         uint32   //结束行号
	NumParams       byte     //固定参数个数
	IsVararg        byte     //是否有可变参数
	MaxRegisterSize byte     //寄存器个数
	Codes           []string //指令表
	Constants       []string //常量表

	Protos   []*prototypeInfo //子函数列表
	LocVars  []LocVar
	UpValues []upvalueInfo
}

type upvalueInfo struct {
	Upvalue
	Name string
}

func ProtoToProtoInfo(proto *Prototype) *prototypeInfo {
	protoInfo := &prototypeInfo{
		Source:          proto.Source,
		LineStart:       proto.LineStart,
		LineEnd:         proto.LineEnd,
		NumParams:       proto.NumParams,
		IsVararg:        proto.IsVararg,
		MaxRegisterSize: proto.MaxRegisterSize,
		Codes:           _getProtoInfoCodes(proto),
		Constants:       _getProtoInfoConstants(proto),
		LocVars:         proto.LocVars,
		UpValues:        _getProtoInfoUpvalues(proto),
	}
	subProtoInfos := make([]*prototypeInfo, len(proto.Protos))
	for i, v := range proto.Protos {
		subProtoInfos[i] = ProtoToProtoInfo(v)
	}
	protoInfo.Protos = subProtoInfos
	return protoInfo
}

func _getProtoInfoCodes(proto *Prototype) []string {
	n := len(proto.Codes)
	ops := make([]string, n)
	for i := 0; i < n; i++ {
		line := -1
		if i < len(proto.LineInfo) {
			line = int(proto.LineInfo[i])
		}
		ops[i] = tool.ReplaceTabToSpace(strconv.Itoa(i)+"\t"+strconv.Itoa(line)+"\t"+proto.Codes[i].Info(), 8)
	}
	return ops
}

func _getProtoInfoConstants(proto *Prototype) []string {
	n := len(proto.Constants)
	cs := make([]string, n)
	for i := 0; i < n; i++ {
		cs[i] = strconv.Itoa(i) + "  " + fmt.Sprintf("%v", proto.Constants[i])
	}
	return cs
}

func _getProtoInfoUpvalues(proto *Prototype) []upvalueInfo {
	n := len(proto.Upvalues)
	ups := make([]upvalueInfo, n)
	for i := 0; i < n; i++ {
		ups[i] = upvalueInfo{
			proto.Upvalues[i],
			proto.UpvalueNames[i],
		}
	}
	return ups
}
