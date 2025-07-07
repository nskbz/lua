package codegen

import (
	"fmt"
	"strings"

	"nskbz.cn/lua/compile/ast"
	"nskbz.cn/lua/instruction"
)

// blockLastLine用于debug生成变量作用域信息
func cgStat(fi *funcInfo, stat ast.Stat, blockLastLine int) {
	switch stat.(type) {
	case *ast.BreakStat:
		cgBreakStat(fi, stat)
	case *ast.GotoStat:
		cgGotoStat(fi, stat)
	case *ast.LabelStat:
		cgLabelStat(fi, stat)
	case *ast.FuncCallStat:
		cgFuncCallStat(fi, stat)
	case *ast.DoStat:
		cgDoStat(fi, stat)
	case *ast.WhileStat:
		cgWhileStat(fi, stat)
	case *ast.RepeatStat:
		cgRepeatStat(fi, stat)
	case *ast.IfStat:
		cgIfStat(fi, stat)
	case *ast.ForInStat:
		cgForInStat(fi, stat)
	case *ast.ForNumStat:
		cgForNumStat(fi, stat)
	case *ast.LocalVarStat:
		cgLocalVarStat(fi, stat, blockLastLine)
	case *ast.AssignStat:
		cgAssignStat(fi, stat)
	case *ast.LocalFuncDefStat:
		cgLocalFuncDefStat(fi, stat)
	case *ast.OopFuncDefStat:
		cgOopFuncDefStat(fi, stat)
	default:
		panic("not support stat")
	}
}

func cgBreakStat(fi *funcInfo, stat ast.Stat) {
	breakStat := stat.(*ast.BreakStat)
	//break语句需要先使用空JMP指令占位，等后续对JMP指令进行填充
	pc := fi.JMP(0, 0)
	fi.recordInsLine(breakStat.Line) //debug
	//记录对应break语句的pc信息
	fi.breakMap.add(pc)
}

func cgGotoStat(fi *funcInfo, stat ast.Stat) {
	gotoStat := stat.(*ast.GotoStat)
	//先用JMP占位，后续填充
	pc := fi.JMP(0, 0)
	fi.recordInsLine(gotoStat.Line)
	fi.gotoMap.add(gotoStat.Name, pc)
}

func cgLabelStat(fi *funcInfo, stat ast.Stat) {
	labelStat := stat.(*ast.LabelStat)
	pc := fi.pc() + 1 //记录label指向的PC，即当前指令的下一位置
	fi.labelMap.add(labelStat.Name, pc)
}

// 语句只会以调用形式单独出现,不需要返回值
func cgFuncCallStat(fi *funcInfo, stat ast.Stat) {
	idx := fi.allocReg()
	cgFuncCallExp(fi, stat.(*ast.FuncCallExp), idx, 0) //n==0 不返回
	fi.freeReg()
}

func cgDoStat(fi *funcInfo, stat ast.Stat) {
	doStat := stat.(*ast.DoStat)
	fi.enterScope(false)
	cgBlock(fi, doStat.Block)
	fi.closeOpenUpvals() //退出作用域需要解绑Upvalue
	fi.exitScope()
}

/*
		_______________
	  / false? jmp		|
	 /           		|

while exp do block end<-|

	|_______/
	   jmp
*/
func cgWhileStat(fi *funcInfo, stat ast.Stat) {
	loopStart := fi.pc() + 1 //条件表达式的pc
	whileStat := stat.(*ast.WhileStat)
	//解析条件表达式
	idx := fi.allocReg()
	cgExp(fi, whileStat.Exp, idx, 1)
	fi.freeReg()
	//通过判断表达式的真假部署分支
	fi.TEST(idx, 0)
	fi.recordInsLine(whileStat.ExpLine)
	//false 则跳转到loopEnd,这里提前不知道,类似break需先占位后填充
	jmpToEnd := fi.JMP(0, 0)
	fi.recordInsLine(whileStat.ExpLine)
	//true 则进行循环体
	fi.enterScope(true)
	cgBlock(fi, whileStat.Block)
	//fi.closeOpenUpvals() 这里应该可以通过设置下面JMP指令的a参数达到一样的效果，待验证
	fi.JMP(fi.getJmpArgA(), loopStart-(fi.pc()+1)-1) //跳回loopStart,这里由于还未插入JMP指令所以PC指向的是上一条,而JMP指令的正确PC应为fi.pc()+1
	fi.recordInsLine(whileStat.ExpLine)
	fi.exitScope()
	//填充上面的JMP指令
	loopEnd := fi.pc() + 1
	fi.fixSbx(jmpToEnd, loopEnd-jmpToEnd-1) //-1是因为JMP指令执行时pc已经指向下一条指令了
}

/*
	_______________________________
	       		| false jmp to block end
	            |

repeat block until exp

	|		        |
	|___true________|
*/
func cgRepeatStat(fi *funcInfo, stat ast.Stat) {
	loopBefore := fi.pc()
	repeatStat := stat.(*ast.RepeatStat)
	//repeat先执行循环体，所以先填充循环体
	fi.enterScope(true)
	cgBlock(fi, repeatStat.Block)
	//解析条件表达式
	idx := fi.allocReg()
	cgExp(fi, repeatStat.Exp, idx, 1)
	fi.freeReg()
	//判断条件
	fi.TEST(idx, 0) //这里待验证是1还是0
	fi.recordInsLine(repeatStat.ExpLine)
	//R(A)==true,跳转开头执行循环体
	fi.JMP(fi.getJmpArgA(), loopBefore-(fi.pc()+1))
	fi.recordInsLine(repeatStat.ExpLine)
	//R(A)==false则PC++,直接就是下一个语句的指令了所以只需释放资源,退出作用域
	fi.closeOpenUpvals()
	fi.exitScope()
}

/*
	 	  _________________         _______________		   ____________
		/ false? jmp      	|     / false? jmp   	|    / false? jmp  /
	  /                  	V   /                  	V   /             /

if exp1 then block1 elseif exp2 then block2 elseif true then block3 end

	\           \            |           \      	|  	  	 	|
	\___________\			 |___________\			|___________|
	 	true               		true                    true
*/
func cgIfStat(fi *funcInfo, stat ast.Stat) {
	ifStat := stat.(*ast.IfStat)
	nExps := len(ifStat.Exps) //>=1
	jmpToNextExp := make([]int, nExps)
	jmpToEnd := make([]int, nExps)
	expsPc := make([]int, nExps)
	for i := 0; i < nExps; i++ {
		//记录条件表达式开始时的PC
		expsPc[i] = fi.pc() + 1
		//解析条件表达式
		idx := fi.allocReg()
		insLine := cgExp(fi, ifStat.Exps[i], idx, 1)
		fi.freeReg()
		//判断表达式
		fi.TEST(idx, 0)
		fi.recordInsLine(insLine)
		//R(A)==false
		//else块没有下一个跳转点，按理最后一个条件表达式不需要填充JMP，但是为了配合TEST指令必须要有一个占位不然PC执行顺序就乱了
		//特殊考虑只有if块的语句,需要跳转至结束
		jmpToNextExp[i] = fi.JMP(fi.getJmpArgA(), 0)
		fi.recordInsLine(insLine)
		//R(A)==true  ,else块结束后也就到达IF STAT结尾,也不需要填充JMP跳转结尾
		fi.enterScope(false)
		cgBlock(fi, ifStat.Blocks[i])
		if i < nExps-1 {
			jmpToEnd[i] = fi.JMP(fi.getJmpArgA(), 0)    //最后一个块虽然已经是语句结束点，但仍需要关闭Upvalues，所以最后一个块的JMP无需填充偏移量
			fi.recordInsLine(ifStat.Blocks[i].LastLine) //if语句块中结束跳转指令的行数应当记录为该块end的行数
		}
		fi.closeOpenUpvals()
		fi.exitScope()
	}
	statEnd := fi.pc() //记录结束PC
	//填充jmpToNextExp
	for i := 0; i < nExps-1; i++ {
		fi.fixSbx(jmpToNextExp[i], expsPc[i+1]-jmpToNextExp[i]-1)
	}
	fi.fixSbx(jmpToNextExp[nExps-1], statEnd-jmpToNextExp[nExps-1])
	//填充jmpToEnd,最后一个无需填充
	for i := 0; i < nExps-1; i++ {
		fi.fixSbx(jmpToEnd[i], statEnd-jmpToEnd[i])
	}
}

/*
for name = init, limit, step do
-- 循环体
end
通过FORPREP和FORLOOP实现
*/
func cgForNumStat(fi *funcInfo, stat ast.Stat) {
	forNumStat := stat.(*ast.ForNumStat)
	fi.enterScope(true)
	//分配start,end,step和R(A+3)所需的寄存器
	cgLocalVarStat(fi, &ast.LocalVarStat{
		LastLine:     forNumStat.LineOfDo,
		LocalVarList: []string{"(for init)", "(for limit)", "(for step)"}, //这里就对应前面funcInfo.getJmpArgA()方法
		ExpList:      []ast.Exp{forNumStat.Init, forNumStat.Limit, forNumStat.Step},
	}, forNumStat.LineOfFor)
	fi.newLocalVar(forNumStat.Name, forNumStat.LineOfDo, forNumStat.LineOfEnd) //将循环变量添加为局部变量，其使用的是R(A+3)而非R(A)
	//获取R(A)的索引
	idx := fi.usedRegs - 4
	//设置循环初始化指令
	forPrep := fi.FORPREP(idx, 0)
	fi.recordInsLine(forNumStat.LineOfFor)
	//解析循环体
	cgBlock(fi, forNumStat.Block)
	fi.closeOpenUpvals()
	//设置循环指令
	forLoop := fi.FORLOOP(idx, 0)
	fi.recordInsLine(forNumStat.LineOfEnd)
	//填充FORPREP和FORLOOP指令的sBx参数
	fi.fixSbx(forPrep, forLoop-forPrep-1)
	fi.fixSbx(forLoop, forPrep-forLoop)
	fi.exitScope() //exitScope会自动释放局部变量的寄存器
}

// 不同于上面通过跳转重复执行，这里"试试"重复填充循环次数的循环体内指令
func cgForNumStat2(fi *funcInfo, stat ast.Stat) {
	forNumStat := stat.(*ast.ForNumStat)

	switch forNumStat.Step.(type) {
	case *ast.IntExp:
		init := forNumStat.Init.(*ast.IntExp).Val
		limit := forNumStat.Limit.(*ast.IntExp).Val
		step := forNumStat.Step.(*ast.IntExp).Val
		for i := init; i < limit; i += step {
			fi.enterScope(true)
			cgBlock(fi, forNumStat.Block)
			fi.closeOpenUpvals()
			fi.exitScope()
		}
	case *ast.FloatExp:
		init := forNumStat.Init.(*ast.FloatExp).Val
		limit := forNumStat.Limit.(*ast.FloatExp).Val
		step := forNumStat.Step.(*ast.FloatExp).Val
		for i := init; i < limit; i += step {
			fi.enterScope(true)
			cgBlock(fi, forNumStat.Block)
			fi.closeOpenUpvals()
			fi.exitScope()
		}
	default:
		panic("not support type")
	}

}

/*
for namelist in iterator, state, controlVar do block end
通过TFORCALL和TFORLOOP实现
*/
func cgForInStat(fi *funcInfo, stat ast.Stat) {
	forInStat := stat.(*ast.ForInStat)
	fi.enterScope(true)
	//创建for循环相关局部变量
	cgLocalVarStat(fi, &ast.LocalVarStat{
		LastLine:     forInStat.LineOfFor,
		LocalVarList: []string{"(for_iterator)", "(for_state)", "(for_controlVar)"},
		ExpList:      forInStat.ExpList,
	}, forInStat.LineOfDo)
	//添加用户局部变量
	for _, v := range forInStat.NameList {
		fi.newLocalVar(v, forInStat.LineOfDo, forInStat.LineOfEnd)
	}

	jmpToTFORCALL := fi.JMP(0, 0)
	fi.recordInsLine(forInStat.LineOfFor)
	//解析循环体
	loopBegin := fi.pc()
	cgBlock(fi, forInStat.Block)
	fi.closeOpenUpvals()
	//填充JMP
	fi.fixSbx(jmpToTFORCALL, fi.pc()-jmpToTFORCALL)

	//设置TFORCALL: R(A+3), ... ,R(A+2+C) := R(A)(R(A+1), R(A+2))
	//作者写法：fi.indexOfLocalVar("(for_iterator)")
	idx := fi.usedRegs - len(forInStat.NameList) - 3 //todo 待验证
	if fi.indexOfLocalVar("(for_iterator)") != idx {
		panic("error idx")
	}
	//len(forInStat.NameList)=((A+2+C)-(A+3))+1 , 即变量个数=R(A+3)到R(A+2+C)的个数，注意计算个数要加1
	c := len(forInStat.NameList)
	fi.TFORCALL(idx, c)
	fi.recordInsLine(forInStat.LineOfFor)
	//设置TFORLOOP: if R(A+1) ~= nil then { R(A)=R(A+1); pc += sBx } , 这里的R(A)是controlVar
	fi.TFORLOOP(idx+2, loopBegin-fi.pc()-1)
	fi.recordInsLine(forInStat.LineOfFor)

	fi.exitScope()
}

// 这里对于局部变量需要申请寄存器allocReg()进行存储，当退出作用域的时候exitScope()时会自动释放其寄存器空间
// 局部变量用于方法的执行所以只能在离开作用域的时候freeReg()
func cgLocalVarStat(fi *funcInfo, stat ast.Stat, scopeLastLine int) {
	localValStat := stat.(*ast.LocalVarStat)
	nVars := len(localValStat.LocalVarList)
	nExps := len(localValStat.ExpList)
	oldUsed := fi.usedRegs
	//当nVars==nExps时，一个Var对应一个Exp
	if nVars == nExps {
		for _, exp := range localValStat.ExpList {
			idx := fi.allocReg()
			cgExp(fi, exp, idx, 1)
		}
	} else if nVars < nExps { //当nVars<nExps时，仍需要生成所有表达式的指令，虽然后面的不会被使用
		for i, exp := range localValStat.ExpList {
			idx := fi.allocReg()
			if i == nExps-1 && _isVarargOrFuncCall(exp) { //最后一个表达式可能是vararg或funcCall
				cgExp(fi, exp, idx, 0) //n==0，将所有返回值压入栈
			} else {
				cgExp(fi, exp, idx, 1)
			}
		}
	} else if nVars > nExps { //当nVars>nExps时，则可能存在vararg或funcCall有多个返回值，不足nVars的变量需填充nil
		multRet := false
		for i, exp := range localValStat.ExpList {
			idx := fi.allocReg()
			//最后一个表达式为vararg或funcCall时
			if i == nExps-1 && _isVarargOrFuncCall(exp) {
				multRet = true
				n := nVars - nExps + 1 //因为最后一个表达式会被使用,则所缺的值又会多一个,所以需要的返回值数量要+1
				cgExp(fi, exp, idx, n)
			} else {
				cgExp(fi, exp, idx, 1)
			}
		}
		//不足nVars的变量且没有vararg或funcCall这种多个返回值的，就需填充nil
		if !multRet {
			n := nVars - nExps
			idx := fi.allocRegs(n)
			fi.LOADNIL(idx, n)
			fi.recordInsLine(localValStat.LastLine)
		}
	}

	//依次绑定变量，使得一个Var对应一个Exp，上面只是开辟了寄存器空间并没有绑定
	//如果最后申请的寄存器空间个数小于变量数，后续多出来的变量则赋值为nil
	fi.usedRegs = oldUsed
	for _, v := range localValStat.LocalVarList {
		fi.newLocalVar(v, localValStat.LastLine, scopeLastLine)
	}

}

// explist ::=exp { ',' exp}
// for example:
// table[k],a=(1+3%2),"jack"
func cgAssignStat(fi *funcInfo, stat ast.Stat) {
	assignStat := stat.(*ast.AssignStat)
	nVars := len(assignStat.VarList)
	nExps := len(assignStat.ExpList)
	oldUsed := fi.usedRegs
	//有可能有表相关的变量，需要使用GETTABLE来装载;R(A) := R(B)[RK(C)]
	tRegs := make([]int, nVars)
	kRegs := make([]int, nVars)
	//为处理VarList中可能存在的表访问,先将操作表所需的寄存器索引生成
	for i, v := range assignStat.VarList {
		if ta, ok := v.(*ast.TableAccessExp); ok {
			tRegs[i] = fi.allocReg()
			cgExp(fi, ta.PrefixExp, tRegs[i], 1)
			kRegs[i] = fi.allocReg()
			cgExp(fi, ta.CurrentExp, kRegs[i], 1)
		}
	}
	//处理ExpList
	vRegs := make([]int, nVars) //为每个变量的新值申请临时空间，记录idx
	if nVars <= nExps {         //nVars<=nExps,则一个表达式依次对应一个变量,多余的舍弃
		for i, exp := range assignStat.ExpList {
			a := fi.allocReg()
			if i == nExps-1 && _isVarargOrFuncCall(exp) {
				cgExp(fi, exp, a, 0) //不返回
			} else {
				cgExp(fi, exp, a, 1)
			}
			if i < nVars { //只需要nVars个表达式值
				vRegs[i] = a
			}
		}
	} else if nVars > nExps { //nVars > nExps,则要考虑最后存在vararg或funcCall的情况
		multiRet := false
		for i, exp := range assignStat.ExpList {
			a := fi.allocReg()
			if i == nExps-1 && _isVarargOrFuncCall(exp) {
				multiRet = true
				n := nVars - nExps //待验证
				fi.allocRegs(n)
				cgExp(fi, exp, a, n)
			} else {
				cgExp(fi, exp, a, 1)
			}
			if i < nVars { //只需要nVars个表达式值
				vRegs[i] = a
			}
		}
		if !multiRet {
			n := nVars - nExps
			a := fi.allocRegs(n)
			fi.LOADNIL(a, n)
			fi.recordInsLine(assignStat.LastLine)
		}
	}

	//将临时存储的变量装载到真正的变量中
	for i := 0; i < nVars; i++ {
		v := assignStat.VarList[i]
		if _, ok := v.(*ast.TableAccessExp); ok { //说明是表内元素
			fi.SETTABLE(tRegs[i], kRegs[i], vRegs[i])
			fi.recordInsLine(assignStat.LastLine)
		} else if nameExp, ok := v.(*ast.NameExp); ok {
			//局部变量
			if idx := fi.indexOfLocalVar(nameExp.Name); idx >= 0 {
				fi.MOVE(idx, vRegs[i])
				fi.recordInsLine(assignStat.LastLine)
				continue
			}
			//Upvalue变量
			if idx := fi.indexOfUpvalue(nameExp.Name); idx >= 0 {
				fi.SETUPVAL(vRegs[i], idx)
				fi.recordInsLine(assignStat.LastLine)
				continue
			}
			//global var:全局变量,是指无需声明即可直接使用、且在任何作用域中均可访问的变量
			a := fi.indexOfUpvalue("_ENV")
			b := 0x100 + fi.indexOfConstant(nameExp.Name)
			fi.SETTABLE(a, b, vRegs[i])
			fi.recordInsLine(assignStat.LastLine)
		}
	}

	//释放寄存器空间
	fi.usedRegs = oldUsed
}

// local function myFunction(params)
//
//	-- 函数体
//	return result
//
// end
func cgLocalFuncDefStat(fi *funcInfo, stat ast.Stat) {
	localFuncDefStat := stat.(*ast.LocalFuncDefStat)
	a := fi.newLocalVar(localFuncDefStat.Name, localFuncDefStat.DefLine, localFuncDefStat.Body.LastLine)
	cgFuncDefExp(fi, localFuncDefStat.Name, localFuncDefStat.Body, a)
}

func cgOopFuncDefStat(fi *funcInfo, stat ast.Stat) {
	oopFuncDefStat := stat.(*ast.OopFuncDefStat)
	funcPath := _cgOopFuncName(oopFuncDefStat.Name)
	funcIdx := fi.newLocalVar(funcPath, oopFuncDefStat.DefLine, oopFuncDefStat.Body.Block.LastLine) //获取存储方法的索引
	cgFuncDefExp(fi, funcPath, oopFuncDefStat.Body, funcIdx)                                        //生成clousure

	//将对象方法与对象绑定
	dotIdx := strings.LastIndex(funcPath, ".")
	varName := funcPath[:dotIdx]    //获取对象名,本质是table
	funcName := funcPath[dotIdx+1:] //获取方法名,即表中元素的key
	varIdx := -1
	if idx := fi.indexOfLocalVar(varName); idx > -1 {
		varIdx = idx
	} else if idx := fi.indexOfUpvalue(varName); idx > -1 {
		varIdx = idx
	}
	if varIdx == -1 {
		panic(fmt.Sprintf("bound OopFunc error,there is no var called %s", varName))
	}
	funcNameIdx := -1
	if idx := fi.indexOfConstant(funcName); idx > -1 {
		funcNameIdx = instruction.ConstantBase + idx
	}
	if funcNameIdx == -1 {
		panic(fmt.Sprintf("bound OopFunc error,there is no constant called %s", funcName))
	}

	fi.SETTABLE(varIdx, funcNameIdx, funcIdx) //绑定指令
	fi.recordInsLine(oopFuncDefStat.DefLine)
}

func _cgOopFuncName(name ast.Exp) string {
	switch exp := name.(type) {
	case *ast.NameExp:
		return exp.Name
	case *ast.StringExp:
		return exp.Str
	case *ast.TableAccessExp:
		pre := _cgOopFuncName(exp.PrefixExp)
		current := _cgOopFuncName(exp.CurrentExp)
		return pre + "." + current
	}
	panic("unknown type")
}
