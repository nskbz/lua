package instruction

import (
	"fmt"
	"strconv"
	"strings"

	"nskbz.cn/lua/api"
	"nskbz.cn/lua/tool"
)

//		|                 			32bit							  |
//iABC: |B 			9bit|C 			9bit|A	 	8bit|Opcode 	  6bit|
//iABx: |Bx						   18bit|A 		8bit|Opcode 	  6bit|
//iAsBx:|sBx     				   18bit|A 		8bit|Opcode 	  6bit|
//iAx:  |Ax		             	 			   26bit|Opcode	      6bit|

type Instruction uint32

// 指令类型
const (
	IABC = iota
	IABx
	IAsBx //sBx就是sign Bx有符号的Bx
	IAx
)

// 操作数类型
const (
	ArgU byte = 1      //使用了该位置的参数
	ArgN byte = 1 << 1 //没有使用该位置的参数
	//R和K默认+1即标志其也属于ArgU
	ArgR  byte = 1<<2 + 1 //参数为寄存器地址,IAsBx类型指令下是跳转地址
	ArgK  byte = 1<<3 + 1 //参数为常量表地址
	ArgRK byte = ArgR | ArgK
)

const ConstantBase = 0x100

const MAX_BX = 1<<18 - 1
const MAX_SBX = MAX_BX >> 1

func (code Instruction) opcode() opcode {
	op := int(code & 0x3F)
	if op >= len(instructions) || op < 0 {
		panic("error code")
	}
	return instructions[op]
}

// 获取uint32 code对应的指令结构类型
func (code Instruction) OpMode() int {
	return code.opcode().opMode
}

// 执行该指令
func (code Instruction) Execute(vm api.LuaVM) {
	h := code.opcode().handler
	if h == nil {
		panic(fmt.Sprintf("instruction[%s] mapping action is nil", code.Name()))
	}
	h(code, vm)
}

func printStack(s api.LuaVM) {
	for i := 1; i <= s.RegisterCount(); i++ {
		tp := s.Type(i)
		switch tp {
		case api.LUAVALUE_BOOLEAN:
			fmt.Printf("[%t]", s.ToBoolean(i))
		case api.LUAVALUE_NUMBER:
			fmt.Printf("[%g]", s.ToFloat(i))
		case api.LUAVALUE_STRING:
			fmt.Printf("[%q]", s.ToString(i))
		default:
			fmt.Printf("[%s]", s.TypeName(tp))
		}
	}
	fmt.Println()
}

func (code Instruction) Name() string {
	return code.opcode().InstructionName()
}

func (code Instruction) Info() string {
	sb := strings.Builder{}
	sb.WriteString(code.opcode().InstructionName() + "\t")
	// ABC指令
	switch i := code.OpMode(); i {
	case IABC:
		a, b, c := code.ABC()
		sb.WriteString(strconv.Itoa(a) + "\t")
		if code.ModArgB(ArgU) {
			sb.WriteString(strconv.Itoa(b))
		}
		sb.WriteByte('\t')
		if code.ModArgC(ArgU) {
			sb.WriteString(strconv.Itoa(c))
		}
		sb.WriteByte('\t')
	case IABx: //IABx中的LOADK可以有b不使用的情况需要处理
		a, bx := code.ABx()
		sb.WriteString(strconv.Itoa(a) + "\t")
		if code.ModArgB(ArgU) {
			sb.WriteString(strconv.Itoa(bx))
		}
		sb.WriteString("\t\t")
	case IAsBx:
		a, sbx := code.AsBx()
		sb.WriteString(strconv.Itoa(a) + "\t" + strconv.Itoa(sbx) + "\t\t")
	case IAx:
		ax := code.Ax()
		sb.WriteString(strconv.Itoa(ax) + "\t\t\t")
	default:
		panic(fmt.Sprintf("not support instruction of type[%d]", i))
	}

	return tool.ReplaceTabToSpace(sb.String(), 4)
}

func (code Instruction) ABC() (a, b, c int) {
	a = int(code >> 6 & 0xFF)
	c = int(code >> 14 & 0x01FF)
	b = int(code >> 23 & 0x01FF)
	return
}

func (code Instruction) ABx() (a, bx int) {
	a = int(code >> 6 & 0xFF)
	bx = int(code >> 14)
	return
}

func (code Instruction) AsBx() (a, sbx int) {
	a = int(code >> 6 & 0xFF)
	sbx = int(code>>14) - MAX_SBX //todo 这样为什么能达到还原sbx的效果 涉及补码知识?
	return
}

func (code Instruction) Ax() int {
	return int(code >> 6)
}

func (code Instruction) ModArgA(mod byte) bool {
	return code.opcode().CheckArgAMod(mod)
}

func (code Instruction) ModArgB(mod byte) bool {
	return code.opcode().CheckArgBMod(mod)
}

func (code Instruction) ModArgC(mod byte) bool {
	return code.opcode().CheckArgCMod(mod)
}

type action func(Instruction, api.LuaVM)

type opcode struct {
	testFlag byte
	argAMode byte
	argBMode byte
	argCMode byte
	opMode   int
	name     string //打印方便

	handler action
}

func (op opcode) TestFlag() bool {
	return op.testFlag != 0
}

func (op opcode) CheckArgAMod(mod byte) bool {
	return op.argAMode&mod != 0
}

func (op opcode) CheckArgBMod(mod byte) bool {
	return op.argBMode&mod != 0
}

func (op opcode) CheckArgCMod(mod byte) bool {
	return op.argCMode&mod != 0
}

func (op opcode) InstructionName() string {
	return op.name
}
