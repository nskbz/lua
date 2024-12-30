package vm

//       		  |			                             32bit										  |
//   iABC: |B 			9bit|C 		9bit|A	 	 8bit|Opcode 	 6bit|
//   iABx: |Bx						  18bit|A 		8bit|Opcode 	6bit|
//iAsBx: |sBx     				  18bit|A 		8bit|Opcode 	6bit|
//    iAx: |Ax		             	 					26bit|Opcode	 6bit|

type Instruction uint32

func (code Instruction) opcode() opcode {
	opcode := int(code & 0x3F)
	if opcode >= len(instructions) || opcode < 0 {
		panic("error code")
	}
	return instructions[opcode]
}

func (code Instruction) OpMode() byte {
	return code.opcode().opMode
}

func (code Instruction) Create() interface{} {
	switch code.OpMode() {
	case IABC:
		return newABC(code)
	case IABx:
		return newAbx(code)
	case IAsBx:
		return newAsBx(code)
	case IAx:
		return newAx(code)
	}
	panic("create instruction error")
}

// 指令类型
const (
	IABC = iota
	IABx
	IAsBx //sBx就是sign Bx有符号的Bx
	IAx
)

// 操作数类型
const (
	ArgN = iota //没有使用该位置的参数
	ArgU        //使用了该位置的参数
	ArgR        //参数为寄存器地址,IAsBx类型指令下是跳转地址
	ArgK        //参数为常量表地址或寄存器地址
	ArgJ        //jump地址
)

type opcode struct {
	testFlag byte
	argAMode byte
	argBMode byte
	argCMode byte
	opMode   byte
	name     string //打印方便
}

func (op opcode) TestFlag() bool {
	return op.testFlag != 0
}
func (op opcode) InstructionName() string {
	return op.name
}

// ABC类型指令
type ABC struct {
	opcode
	a int
	b int
	c int
}

func (abc *ABC) A() int         { return abc.a }
func (abc *ABC) B() int         { return abc.b }
func (abc *ABC) C() int         { return abc.c }
func (abc *ABC) UsedArgB() bool { return abc.argBMode != ArgN }
func (abc *ABC) UsedArgC() bool { return abc.argCMode != ArgN }
func newABC(i Instruction) ABC {
	return ABC{
		opcode: i.opcode(),
		a:      int(i >> 6 & 0xFF),
		c:      int(i >> 14 & 0x01FF),
		b:      int(i >> 23 & 0x01FF),
	}
}

const MAX_BX = 2 ^ 18 - 1
const MAX_SBX = MAX_BX >> 1

// ABx类型指令
type ABx struct {
	opcode
	a  int
	bx int
}

func (abx *ABx) A() int         { return abx.a }
func (abx *ABx) Bx() int        { return abx.bx }
func (abx *ABx) UsedArgB() bool { return abx.argBMode != ArgN }
func (abx *ABx) BisArgK() bool  { return abx.argBMode == ArgK }
func newAbx(i Instruction) ABx {
	return ABx{
		opcode: i.opcode(),
		a:      int(i >> 6 & 0xFF),
		bx:     int(i >> 14),
	}
}

// AsBx类型指令
type AsBx struct {
	opcode
	a   int
	sbx int
}

func (asbx *AsBx) A() int   { return asbx.a }
func (asbx *AsBx) Sbx() int { return asbx.sbx }
func newAsBx(i Instruction) AsBx {
	return AsBx{
		opcode: i.opcode(),
		a:      int(i >> 6 & 0xFF),
		sbx:    int(i>>14) - MAX_SBX,
	}
}

// Ax类型指令
type Ax struct {
	opcode
	ax int
}

func (ax *Ax) Ax() int { return ax.ax }
func newAx(i Instruction) Ax {
	return Ax{
		opcode: i.opcode(),
		ax:     int(i >> 6),
	}
}
