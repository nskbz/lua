package codegen

import (
	"fmt"

	"nskbz.cn/lua/compile/ast"
	. "nskbz.cn/lua/compile/lexer"
	. "nskbz.cn/lua/instruction"
)

var arithAndBitwiseMapping = map[int]int{
	TOKEN_OP_ADD:  OP_ADD,
	TOKEN_OP_SUB:  OP_SUB,
	TOKEN_OP_MUL:  OP_MUL,
	TOKEN_OP_MOD:  OP_MOD,
	TOKEN_OP_POW:  OP_POW,
	TOKEN_OP_DIV:  OP_DIV,
	TOKEN_OP_IDIV: OP_IDIV,
	TOKEN_OP_BAND: OP_BAND,
	TOKEN_OP_BOR:  OP_BOR,
	TOKEN_OP_BXOR: OP_BXOR,
	TOKEN_OP_SHL:  OP_SHL,
	TOKEN_OP_SHR:  OP_SHR,
}

type funcInfo struct {
	parent *funcInfo //上层函数的指针,用于捕获upvalue

	funcName string          //用于debug
	exp      *ast.FuncDefExp //用于debug

	constants map[interface{}]int //常量表
	usedRegs  int                 //已经使用寄存器数
	maxRegs   int                 //用于记录该函数所需的最大寄存器数量,便于后便创建Proto

	localVars []*localVarInfo          //方法域(最外层)中的所有局部变量列表,用于debug
	upvalVars map[string]*upvalInfo    //方法域中的所有捕获变量
	scope     int                      //当前作用域层级,>=0
	scopeVars map[string]*localVarInfo //当前作用域范围内的所有局部变量

	breakMap breakMap //记录break的退出位置
	labelMap labelMap //记录label的位置
	gotoMap  gotoMap  //记录goto位置

	instructions []Instruction //方法的所有指令
	lineOfIns    []uint32      //指令对应的行号,用于debug

	subFuncs  []*funcInfo //嵌套的子函数
	numParams int         //该方法的参数个数
	isVararg  bool        //是否有可变参数
}

func newFuncInfo(parent *funcInfo, funcName string, funcDef *ast.FuncDefExp) *funcInfo {
	fi := &funcInfo{
		funcName:     funcName,
		exp:          funcDef,
		parent:       parent,
		constants:    make(map[interface{}]int),
		usedRegs:     0,
		maxRegs:      0,
		localVars:    make([]*localVarInfo, 0, 8),
		upvalVars:    make(map[string]*upvalInfo),
		scope:        -1,
		scopeVars:    make(map[string]*localVarInfo),
		breakMap:     breakMap{},
		labelMap:     labelMap{},
		gotoMap:      gotoMap{},
		instructions: make([]Instruction, 0, 8),
		lineOfIns:    make([]uint32, 0),
		subFuncs:     make([]*funcInfo, 0),
		numParams:    len(funcDef.ArgList),
		isVararg:     funcDef.IsVararg,
	}
	fi.breakMap.breaks = [][]int{}
	fi.breakMap.scope = &fi.scope
	fi.labelMap.labels = []map[string]int{}
	fi.labelMap.scope = &fi.scope
	fi.gotoMap.gotos = []map[string][]int{}
	fi.gotoMap.scope = &fi.scope
	return fi
}

// 获取常量索引，如果该常量不在表中，则加入表并返回其索引
// 索引从0开始
func (fi *funcInfo) indexOfConstant(constant interface{}) int {
	if idx, ok := fi.constants[constant]; ok {
		return idx
	}
	i := len(fi.constants)
	fi.constants[constant] = i
	return i
}

// 分配一个寄存器并返回其索引
// 索引范围[0,254]
func (fi *funcInfo) allocReg() int {
	fi.usedRegs++
	if fi.usedRegs >= 255 {
		panic("regs out of number")
	}
	if fi.maxRegs < fi.usedRegs {
		fi.maxRegs = fi.usedRegs
	}
	return fi.usedRegs - 1
}

// 释放最近的寄存器
func (fi *funcInfo) freeReg() {
	fi.usedRegs--
}

// 申请n个寄存器并返回第一个寄存器的索引
func (fi *funcInfo) allocRegs(n int) int {
	for i := 0; i < n; i++ {
		fi.allocReg()
	}
	return fi.usedRegs - n
}

// 释放n个寄存器
func (fi *funcInfo) freeRegs(n int) {
	for i := 0; i < n; i++ {
		fi.freeReg()
	}
}

// 进入新的作用域
func (fi *funcInfo) enterScope(isLoopScope bool) {
	fi.scope++
	fi.breakMap.create(isLoopScope)
	fi.labelMap.create()
	fi.gotoMap.create()
}

// 退出当前作用域，需要释放不用的信息
func (fi *funcInfo) exitScope() {
	// 退出作用域就意味着Break语句的跳转位置知道了，就需要填充Break语句的JMP占位指令的参数
	// 然后释放掉使用过的break信息
	pendingBreaks := fi.breakMap.pop()
	a := fi.getJmpArgA() //获取JMP指令需要释放的起始寄存器空间
	for _, breakPc := range pendingBreaks {
		sBx := fi.pc() - breakPc
		sBx += MAX_SBX //????why????
		i := sBx<<14 | a<<6 | OP_JMP
		fi.instructions[breakPc] = Instruction(i) //填充break的跳转位置
	}

	//处理goto
	labels := fi.labelMap.pop()
	for label, labelPc := range labels {
		gotoPCs := fi.gotoMap.pop(label)
		for _, v := range gotoPCs {
			fi.fixSbx(v, labelPc-v-1) //填充goto语句的JMP跳转位置
		}

	}

	//释放局部变量，即释放其分配的寄存器
	fi.scope--
	for _, v := range fi.scopeVars {
		if v.scope > fi.scope {
			fi.freeLocalVar(v)
		}
	}
}

// 获取需要释放的寄存器空间的起始索引
func (fi *funcInfo) getJmpArgA() int {
	hasCapturedVar := false
	regIndex := fi.maxRegs
	//从当前作用域中
	for _, localVar := range fi.scopeVars {
		for v := localVar; v != nil && v.scope == fi.scope; v = v.prev { //作用域下的所有同名局部变量都需要释放寄存器空间
			if localVar.captured {
				hasCapturedVar = true
			}
			// v.name[0] != '(' why???? ===> 对于一些语言层面的变量我们约定用"(...)"为名字，区别用户变量名
			// for example：对于生成forNum语句，我们需要R(A)~R(A+2)的变量来辅助语言的实现，而这三个局部变量名分别为{"(for init)","(for limit)","(for step)"}
			if localVar.slot < regIndex && v.name[0] != '(' {
				regIndex = localVar.slot
			}
		}
	}

	if hasCapturedVar {
		return regIndex + 1 //加一因为JMP：if (A) close all upvalues >= R(A - 1)
	}
	return 0
}

// 释放当前作用域下的局部变量
func (fi *funcInfo) freeLocalVar(v *localVarInfo) {
	fi.freeReg()       //释放一个寄存器位置,,,,这里释放最上面的寄存器是否存在问题？
	if v.prev == nil { //该局部变量上层没有同名的则删除该变量名
		delete(fi.scopeVars, v.name)
	} else if v.prev.scope == v.scope { //同一作用域下同名的局部变量都要删除
		fi.freeLocalVar(v.prev)
	} else {
		fi.scopeVars[v.name] = v.prev //切换为上层的局部变量
	}
}

// 添加局部变量到当前作用域并返回其对应的寄存器索引
func (fi *funcInfo) newLocalVar(name string, startLine, endLine int) int {
	lv := &localVarInfo{
		prev:      nil,
		name:      name,
		scope:     fi.scope,
		slot:      fi.allocReg(),
		captured:  false,
		startLine: startLine,
		endLine:   endLine,
	}
	//当前作用域下有同名的变量
	if v, ok := fi.scopeVars[name]; ok {
		lv.prev = v
	}
	fi.scopeVars[name] = lv //添加到当前作用域
	fi.localVars = append(fi.localVars, lv)
	return lv.slot
}

// 获取当前作用域下局部变量名为name的寄存器索引
func (fi *funcInfo) indexOfLocalVar(name string) int {
	if v, ok := fi.scopeVars[name]; ok {
		return v.slot
	}
	return -1
}

// 获取name的Upvalue索引，如果是才遇见的Upvalue则尝试进行捕获，捕获失败返回-1
func (fi *funcInfo) indexOfUpvalue(name string) int {
	//如果是已经绑定了的upvalue直接返回其索引
	if v, ok := fi.upvalVars[name]; ok {
		return v.index
	}
	//未被绑定的upvalue则去上层函数进行寻找
	if fi.parent != nil {
		//在上层函数中找到了对应的需要捕获的局部变量，则进行绑定
		if v, ok := fi.parent.scopeVars[name]; ok {
			idx := len(fi.upvalVars)
			fi.upvalVars[name] = &upvalInfo{
				localVarSlot: v.slot,
				upvalIndex:   -1,
				index:        idx,
			}
			v.captured = true //局部变量标记为已捕获
			return idx
		}
		//上层函数中没有捕获到变量，则说明需捕获的变量在更上层
		//这里的向上递归有点层层捕获的意思，即最下层如果要捕获第一层的变量x，则必须先由第二层捕获x，再由第三层捕获x，依次到最下层
		if uvIdx := fi.parent.indexOfUpvalue(name); uvIdx > -1 {
			idx := len(fi.upvalVars)
			fi.upvalVars[name] = &upvalInfo{
				localVarSlot: -1,
				upvalIndex:   uvIdx,
				index:        idx,
			}
			return idx
		}
	}
	return -1 //捕获外部变量失败
}

func (fi *funcInfo) closeOpenUpvals() {
	a := fi.getJmpArgA()
	if a > 0 {
		fi.JMP(a, 0) //不跳转只执行关闭Upvalues
		fi.recordInsLine(-1)
	}
}

func (fi *funcInfo) recordInsLine(line int) {
	fi.lineOfIns = append(fi.lineOfIns, uint32(line))
}

/*
	装载各种虚拟机指令的方法,只有这些方法才会影响PC
*/

// iABC: |B 			9bit|C 			9bit|A	 	8bit|Opcode 	  6bit|
func (fi *funcInfo) addInstructionOfABC(opcode, a, b, c int) {
	i := b<<23 | c<<14 | a<<6 | opcode
	fi.instructions = append(fi.instructions, Instruction(i))
}

// iABx: |Bx						   18bit|A 		8bit|Opcode 	  6bit|
func (fi *funcInfo) addInstructionOfBx(opcode, a, bx int) {
	i := bx<<14 | a<<6 | opcode
	fi.instructions = append(fi.instructions, Instruction(i))
	//fmt.Printf("i[%d] = %032b\n", len(fi.instructions)-1, i)
}

// iAsBx:|sBx     				   18bit|A 		8bit|Opcode 	  6bit|
func (fi *funcInfo) addInstructionOfsBx(opcode, a, sbx int) {
	if sbx > MAX_SBX || sbx < -int(MAX_SBX) {
		panic(fmt.Sprintf("sbx %d out of [-%d,%d)", sbx, MAX_SBX, MAX_SBX))
	}
	i := (sbx+MAX_SBX)<<14 | a<<6 | opcode //todo=>为什么加MAX_SBX???
	// fmt.Printf("i+\t\t = %032b\n", uint32(i))
	fi.instructions = append(fi.instructions, Instruction(i))
}

// iAx:  |Ax		             	 			   26bit|Opcode	      6bit|
func (fi *funcInfo) addInstructionOfAx(opcode, ax int) {
	i := ax<<6 | opcode
	fi.instructions = append(fi.instructions, Instruction(i))
}

// 返回最后一条指令的索引
func (fi *funcInfo) pc() int { return len(fi.instructions) - 1 }

// 更新sbx数值
func (fi *funcInfo) fixSbx(pc, sbx int) {
	i := uint32(fi.instructions[pc])
	i = i << 18 >> 18               //先左移18bit把sBx的原数据移除再右移18bit复位
	i = i | uint32(sbx+MAX_SBX)<<14 //sbx + MAX_SBX ???不懂 涉及补码知识？
	fi.instructions[pc] = Instruction(i)
}

// R(A) := R(B)
func (fi *funcInfo) MOVE(a, b int) {
	fi.addInstructionOfABC(OP_MOVE, a, b, 0)
}

// R(A) := (bool)B; if (C) pc++
func (fi *funcInfo) LOADBOOL(a, b, c int) {
	fi.addInstructionOfABC(OP_LOADBOOL, a, b, c)
}

// 装载n个nil到R(A)....R(A+n-1)里
// R(A), R(A+1), ..., R(A+B) := nil
func (fi *funcInfo) LOADNIL(a, n int) {
	fi.addInstructionOfABC(OP_LOADNIL, a, n-1, 0)
}

// 将常量表中索引为Bx的值装载到A寄存器中
// R(A) := Kst(Bx)
func (fi *funcInfo) LOADK(a int, constant interface{}) {
	kIdx := fi.indexOfConstant(constant) //所有字面量都记录为常量存储
	if kIdx <= MAX_BX {
		fi.addInstructionOfBx(OP_LOADK, a, kIdx)
		return
	}
	//超出iBx类型指令索引范围就需要使用LOADKX和EXTRAARG指令来扩展索引范围
	fi.addInstructionOfBx(OP_LOADKX, a, 0)
	fi.addInstructionOfAx(OP_EXTRAARG, kIdx)
}

// R(A) := UpValue[B]
func (fi *funcInfo) GETUPVAL(a, b int) {
	fi.addInstructionOfABC(OP_GETUPVAL, a, b, 0)
}

// UpValue[B] := R(A)
func (fi *funcInfo) SETUPVAL(a, b int) {
	fi.addInstructionOfABC(OP_SETUPVAL, a, b, 0)
}

// R(A) := UpValue[B][RK(C)]
func (fi *funcInfo) GETTABUP(a, b, c int) {
	fi.addInstructionOfABC(OP_GETTABUP, a, b, c)
}

// UpValue[A][RK(B)] := RK(C)
func (fi *funcInfo) SETTABUP(a, b, c int) {
	fi.addInstructionOfABC(OP_SETTABUP, a, b, c)
}

// R(A) := R(B)[RK(C)]
func (fi *funcInfo) GETTABLE(a, b, c int) {
	fi.addInstructionOfABC(OP_GETTABLE, a, b, c)
}

// R(A)[RK(B)] := RK(C)
func (fi *funcInfo) SETTABLE(a, b, c int) {
	fi.addInstructionOfABC(OP_SETTABLE, a, b, c)
}

// R(A) := {} (size = B,C)
func (fi *funcInfo) NEWTABLE(a, b, c int) {
	fi.addInstructionOfABC(OP_NEWTABLE, a, b, c)
}

// R(A+1) := R(B); R(A) := R(B)[RK(C)]
func (fi *funcInfo) SELF(a, b, c int) {
	fi.addInstructionOfABC(OP_SELF, a, b, c)
}

// R(A), R(A+1), ..., R(A+B-2) = vararg
func (fi *funcInfo) Vararg(a, nReturn int) {
	//(A+B-2)-(A)+1=nReturn-->nReturn=B-1
	fi.addInstructionOfABC(OP_VARARG, a, nReturn+1, 0)
}

// R(A)[(C-1)*FPF+i] := R(A+i), 1 <= i <= B
func (fi *funcInfo) SETLIST(a, b, c int) {
	fi.addInstructionOfABC(OP_SETLIST, a, b, c)
}

// R(A) := closure(KPROTO[Bx])
func (fi *funcInfo) CLOSURE(a, bx int) {
	fi.addInstructionOfBx(OP_CLOSURE, a, bx)
}

// R(A) := R(B).. ... ..R(C)
func (fi *funcInfo) CONCAT(a, b, c int) {
	fi.addInstructionOfABC(OP_CONCAT, a, b, c)
}

// pc+=sBx; if (A) close all upvalues >= R(A - 1)
func (fi *funcInfo) JMP(a, sbx int) int {
	fi.addInstructionOfsBx(OP_JMP, a, sbx)
	return len(fi.instructions) - 1 //返回该JUMP指令的pc
}

// if not (R(A) == C) then pc++
func (fi *funcInfo) TEST(a, c int) {
	fi.addInstructionOfABC(OP_TEST, a, 0, c)
}

// if (R(B) == C) then R(A) := R(B) else pc++
func (fi *funcInfo) TESTSET(a, b, c int) {
	fi.addInstructionOfABC(OP_TESTSET, a, b, c)
}

// nArg==B nReturn==C,nReturn==0则不返回,nReturn<=-1则返回所有返回值
// R(A), ... ,R(A+C-2) := R(A)(R(A+1), ... ,R(A+B-1))
func (fi *funcInfo) CALL(a, nArg, nReturn int) {
	//(A+B-1)-(A+1)+1=B-1=nArg-->B=nArg+1
	//(A+C-2)-A+1=C-1=nReturn-->C=nReturn+1
	fi.addInstructionOfABC(OP_CALL, a, nArg+1, nReturn+1)
}

// nArg==B
// return R(A)(R(A+1), ... ,R(A+B-1))
func (fi *funcInfo) TAILCALL(a, nArg int) {
	fi.addInstructionOfABC(OP_TAILCALL, a, nArg+1, 0)
}

// nReturn==B , nReturn==-1表示返回全部表值
// return R(A),...,R(A+B-2)
func (fi *funcInfo) RETURN(a, nReturn int) {
	fi.addInstructionOfABC(OP_RETURN, a, nReturn+1, 0)
}

// R(A)+=R(A+2); if R(A) <?= R(A+1) then { pc+=sBx; R(A+3)=R(A) }
func (fi *funcInfo) FORLOOP(a, sbx int) int {
	fi.addInstructionOfsBx(OP_FORLOOP, a, sbx)
	return len(fi.instructions) - 1
}

// R(A)-=R(A+2); pc+=sBx
func (fi *funcInfo) FORPREP(a, sbx int) int {
	fi.addInstructionOfsBx(OP_FORPREP, a, sbx)
	return len(fi.instructions) - 1
}

// R(A+3), ... ,R(A+2+C) := R(A)(R(A+1), R(A+2))
func (fi *funcInfo) TFORCALL(a, c int) {
	fi.addInstructionOfABC(OP_TFORCALL, a, 0, c)
}

// if R(A+1) ~= nil then { R(A)=R(A+1); pc += sBx }
func (fi *funcInfo) TFORLOOP(a, sbx int) {
	fi.addInstructionOfsBx(OP_TFORLOOP, a, sbx)
}

// r[a] = op r[b]
func (fi *funcInfo) UnitaryOP(line, opcode, a, b int) {
	//这里的opcode是通过词法分析的，所以应当使用lexer包下定义的常量
	switch opcode {
	case TOKEN_OP_UNM: // R(A) := -R(B)
		fi.addInstructionOfABC(OP_UNM, a, b, 0)
	case TOKEN_OP_BNOT: // R(A) := ~R(B)
		fi.addInstructionOfABC(OP_BNOT, a, b, 0)
	case TOKEN_OP_NOT: // R(A) := not R(B)
		fi.addInstructionOfABC(OP_NOT, a, b, 0)
	case TOKEN_OP_LEN: // R(A) := length of R(B)
		fi.addInstructionOfABC(OP_LEN, a, b, 0)
	}
	fi.recordInsLine(line)
}

// r[a] = rk[b] op rk[c]
func (fi *funcInfo) DualOP(line, opcode, a, b, c int) {
	// arith & bitwise
	if instruction, ok := arithAndBitwiseMapping[opcode]; ok {
		fi.addInstructionOfABC(instruction, a, b, c)
		fi.recordInsLine(line)
		return
	}
	// relation
	switch opcode {
	case TOKEN_OP_EQ: // if ((RK(B) == RK(C)) ~= A) then pc++
		fi.addInstructionOfABC(OP_EQ, 1, b, c)
	case TOKEN_OP_NE:
		fi.addInstructionOfABC(OP_EQ, 0, b, c)
	case TOKEN_OP_LT: // if ((RK(B) <  RK(C)) ~= A) then pc++
		fi.addInstructionOfABC(OP_LT, 1, b, c)
	case TOKEN_OP_GT:
		fi.addInstructionOfABC(OP_LT, 1, c, b)
	case TOKEN_OP_LE: // if ((RK(B) <= RK(C)) ~= A) then pc++
		fi.addInstructionOfABC(OP_LE, 1, b, c)
	case TOKEN_OP_GE:
		fi.addInstructionOfABC(OP_LE, 1, c, b)
	default:
		panic(fmt.Sprintf("not support tokenOpcode[%d]", opcode))
	}
	fi.recordInsLine(line)
	// relationOP运算后会将布尔结果存在a索引处
	fi.JMP(0, 1)
	fi.recordInsLine(line)
	fi.LOADBOOL(a, 0, 1)
	fi.recordInsLine(line)
	fi.LOADBOOL(a, 1, 0)
	fi.recordInsLine(line)
}

// 局部变量
// for example:
//
// function f()
//
//	local a,b=1,2;
//	do
//		print(a,b)
//		local a,b=3,4
//		print(a,b)
//	end
//	print(a,b)
//
// end
//
// 对于上面的这种多个作用域的情况，同一个局部变量名可以对应不同的存储位置
// 我们也可以发现作用域都是层层递进的，对于某个特定名字的局部变量，访问它获取的值应是最近的定义位置
// 局部变量的生命也随着出对应作用域而终结，返回上一层应切换成上一层作用域的存储位置
// 所以这种单一递进的方式可以采用单向链表的结构
type localVarInfo struct {
	prev     *localVarInfo //上层作用域的同名局部变量
	name     string        //变量名
	scope    int           //作用域
	slot     int           //寄存器的索引
	captured bool          //是否被捕获

	//localvar scope
	startLine int
	endLine   int
}

// 由于break语句用于打断循环，基于跳转JMP指令的实现
// JMP指令：pc+=sBx; if (A) close all upvalues >= R(A - 1)
//
// 有两个难处理的点
// 1.break语句可能处于更深层次的作用域中
// for example:
// -- 双重循环中的break
// for i = 1, 3 do
//
//	print("外层循环 i =", i)
//	for j = 1, 3 do
//	    if j == 2 then
//	        break  -- 只跳出内层j循环
//	    end
//	    print("  内层循环 j =", j)
//	end
//
// end
// 这里break处于scope=3,需要跳转到scope=2的作用域内
//
// 2.通常break语句在作用域结束前出现，由于顺序处理的缘故所以处理break的时候并不知晓最后跳转的具体位置，所以先使用一条JMP指令占着位置等待后续填充参数
type breakMap struct {
	scope  *int    //指向funcInfo.scope
	breaks [][]int //breaks[scope][pc]，同一作用域下可以有多个break，记录每条break语句(JMP指令)的pc位置方便后续填充
}

// 创建一条表项用于记录该新作用域的break信息
func (b *breakMap) create(isLoop bool) {
	if *b.scope >= len(b.breaks) {
		for i := len(b.breaks); i <= *b.scope; i++ {
			b.breaks = append(b.breaks, nil)
		}
		//b.breaks = append(b.breaks, nil)
	}
	if isLoop {
		b.breaks[*b.scope] = []int{}
	} else {
		b.breaks[*b.scope] = nil
	}
}

// 添加当前作用域下break的pc到[对应退出后的作用域]下
// 一开始并不知道break的结束地方在哪，需要后面的信息才能填充，所以先用JMP占位并记录PC用于后面退出作用域时填充
func (b *breakMap) add(pc int) {
	//break跳出循环作用域,当前作用域可能不是,所以需要找最近的循环作用域
	for i := *b.scope; i >= 0; i-- {
		if v := b.breaks[i]; v != nil {
			b.breaks[i] = append(b.breaks[i], pc)
			return
		}
	}
	panic(fmt.Sprintf("there is no loop to break from bottom scope[%d] to top scope[0]", *b.scope))
}

// 弹出最深作用域的break信息
func (b *breakMap) pop() []int {
	breaks := b.breaks[*b.scope]
	if breaks != nil {
		b.breaks[*b.scope] = nil
	}
	return breaks
}

// label和goto与break类似
// goto到label的时候，label可以在goto语句出现之前也能是之后，所以goto的时候先生成JMP占位后续填充
// 但不同于break，当退出当前作用域时，goto的目的label或许还未出现，所以我们每次退出作用域的时候都需要进行goto的填充
type labelMap struct {
	scope  *int
	labels []map[string]int //labels[scope]map[label]pc
}

func (l *labelMap) create() {
	if *l.scope >= len(l.labels) {
		for i := len(l.labels); i <= *l.scope; i++ {
			l.labels = append(l.labels, nil)
		}
	}
	l.labels[*l.scope] = make(map[string]int)
}

func (l *labelMap) add(label string, pc int) {
	if len(l.labels) >= *l.scope && l.labels[*l.scope] != nil {
		// if l.labels[*l.scope] == nil {
		// 	l.labels[*l.scope] = make(map[string]int)
		// }
		lm := l.labels[*l.scope]
		if _, ok := lm[label]; ok {
			panic("duplicate label in the same scope") //同一作用域不能有重复的label
		}
		lm[label] = pc
		return
	}
	panic("add label error")
}

// 弹出当前作用域内的所有label标签
func (l *labelMap) pop() map[string]int {
	ls := l.labels[*l.scope]
	l.labels[*l.scope] = nil
	return ls
}

type gotoMap struct {
	scope *int
	gotos []map[string][]int
}

func (g *gotoMap) create() {
	if *g.scope >= len(g.gotos) {
		for i := len(g.gotos); i <= *g.scope; i++ {
			g.gotos = append(g.gotos, nil)
		}
	}
	g.gotos[*g.scope] = make(map[string][]int, 0)
}

func (g *gotoMap) add(label string, pc int) {
	if *g.scope <= len(g.gotos) && g.gotos[*g.scope] != nil {
		gm := g.gotos[*g.scope]
		if _, ok := gm[label]; !ok {
			gm[label] = make([]int, 0)
		}
		gm[label] = append(gm[label], pc)
		return
	}
	panic("add goto error")
}

// 弹出到目前作用域的所有跳转标签为label的goto
func (g *gotoMap) pop(label string) []int {
	result := []int{}
	nGotos := len(g.gotos)

	for i := nGotos - 1; i >= *g.scope; i-- {
		gm := g.gotos[i]
		if pc, ok := gm[label]; ok {
			result = append(result, pc...)
			delete(gm, label) //删除查找到的gotoPC
		}
	}

	return result
}

// Upvalue表
// upvalue是共享的：内层函数和外层函数访问的是同一个变量
// 闭包保持状态：即使外层函数已返回，被捕获的变量依然存在
// for example:
// local a = 1
// function test()
//
//	a = a + 1
//	local b = 100
//	return function()
//	    a = a - 1
//	    b = b + 1
//	    return a, b
//	end
//
// end
//
// print(test()()) -- 输出 1 101
// -----------------------------------------------------------------------
// function <test.lua:5,9> (10 instructions at 0x5daaa14612d0)
// 0 params, 2 slots, 2 upvalues, 0 locals, 1 constant, 0 functions
//
//	1       [6]     GETUPVAL        0 0     ; a
//	2       [6]     SUB             0 0 -1  ; - 1
//	3       [6]     SETUPVAL        0 0     ; a
//	4       [7]     GETUPVAL        0 1     ; b
//	5       [7]     ADD             0 0 -1  ; - 1
//	6       [7]     SETUPVAL        0 1     ; b
//	7       [8]     GETUPVAL        0 0     ; a
//	8       [8]     GETUPVAL        1 1     ; b
//	9       [8]     RETURN          0 3
//	10      [9]     RETURN          0 1
//
// 为什么需要记录这些信息，应该是需要访问的变量为Upvalue捕获变量时需要对应的索引信息，翻译成GETUPVAL和SETUPVAL
type upvalInfo struct {
	localVarSlot int //捕获外部局部变量的寄存器索引,该值不为-1说明捕获的就是上层函数的局部变量,为-1则说明至少在上上层函数
	upvalIndex   int //外围函数中Upvalue表中的索引,该值不为-1说明捕获的不是上层函数的局部变量,可能是上上层函数甚至上上上层函数
	index        int //Upalue在本函数中出现的顺序,即索引
}
