package instruction

import (
	"fmt"

	"nskbz.cn/lua/api"
)

//       		  |			                             32bit										  |
//   iABC: |B 			9bit|C 		9bit|A	 	 8bit|Opcode 	 6bit|
//   iABx: |Bx						  18bit|A 		8bit|Opcode 	6bit|
//iAsBx: |sBx     				  18bit|A 		8bit|Opcode 	6bit|
//    iAx: |Ax		             	 					26bit|Opcode	 6bit|

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
func (code Instruction) OpMode() byte {
	return code.opcode().opMode
}

// 执行该指令
func (code Instruction) Execute(vm api.LuaVM) {
	h := code.opcode().handler
	if h == nil {
		panic(fmt.Sprintf("instruction[%s] mapping action is nil", code.InstructionName()))
	}
	h(code, vm)
}

func (code Instruction) InstructionName() string {
	return code.opcode().InstructionName()
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
	sbx = int(code>>14) - MAX_SBX
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
	opMode   byte
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
