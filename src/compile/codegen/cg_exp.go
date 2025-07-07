package codegen

import (
	"fmt"

	"nskbz.cn/lua/compile/ast"
	"nskbz.cn/lua/compile/lexer"
	"nskbz.cn/lua/instruction"
)

var anonymousFuncIdx int = 0

// 生成表达式对应的指令
// a:寄存器起始索引
// n:LOADNIL的长度,FUNCCALL的nReturn(n==0没返回值,n<0接受所有返回值)
// 返回该表达式所在的行数,用于debug信息记录
//
// 所有cgXXXExp方法进入时都自带一个寄存器空间,其索引为a由cgExp方法的调用者申请
func cgExp(fi *funcInfo, exp ast.Exp, a, n int) int {
	switch e := exp.(type) {
	case *ast.NilExp:
		fi.LOADNIL(a, n)
		fi.recordInsLine(e.Line)
		return e.Line
	case *ast.FalseExp:
		fi.LOADBOOL(a, 0, 0)
		fi.recordInsLine(e.Line)
		return e.Line
	case *ast.TrueExp:
		fi.LOADBOOL(a, 1, 0)
		fi.recordInsLine(e.Line)
		return e.Line
	case *ast.IntExp:
		fi.LOADK(a, e.Val)
		fi.recordInsLine(e.Line)
		return e.Line
	case *ast.FloatExp:
		fi.LOADK(a, e.Val)
		fi.recordInsLine(e.Line)
		return e.Line
	case *ast.StringExp:
		fi.LOADK(a, e.Str)
		fi.recordInsLine(e.Line)
		return e.Line
	case *ast.ParensExp:
		return cgExp(fi, e.Exp, a, n)
	case *ast.VarargExp:
		cgVarargExp(fi, e, a, n)
		return e.Line
	case *ast.FuncDefExp:
		cgFuncDefExp(fi, fmt.Sprintf("$%d", anonymousFuncIdx), e, a) //匿名函数
		anonymousFuncIdx++
		return e.DefLine
	case *ast.TableConstructExp:
		cgTableConstructExp(fi, e, a)
		return e.Line
	case *ast.ConcatExp:
		cgConcatExp(fi, e, a)
		return e.Line
	case *ast.NameExp:
		cgNameExp(fi, e, a)
		return e.Line
	case *ast.TableAccessExp:
		cgTableAccessExp(fi, e, a)
		return e.LastLine
	case *ast.FuncCallExp:
		cgFuncCallExp(fi, e, a, n)
		return e.Line
	case *ast.UnitaryOpExp:
		cgUnitaryOpExp(fi, e, a)
		return e.Line
	case *ast.DualOpExp:
		cgDualOpExp(fi, e, a)
		return e.Line
	}
	return -1
}

func cgVarargExp(fi *funcInfo, exp *ast.VarargExp, a, n int) {
	if !fi.isVararg {
		panic("vararg paramter must be inside of vararg function")
	}
	fi.Vararg(a, n)
	fi.recordInsLine(exp.Line)
}

// R(A) := function(args) body end
func cgFuncDefExp(fi *funcInfo, funcName string, exp *ast.FuncDefExp, a int) {
	if fi.funcName != "" { //""表示最高层函数,不需要添加':'
		funcName = fmt.Sprintf("%s:%s", fi.funcName, funcName)
	}
	subFunc := newFuncInfo(fi, funcName, exp)
	fi.subFuncs = append(fi.subFuncs, subFunc)

	//将子函数的参数添加入"它自身"的局部变量中
	for _, v := range exp.ArgList {
		subFunc.newLocalVar(v, exp.DefLine, exp.LastLine)
	}
	//生成子函数的指令
	subFunc.enterScope(false) //进入函数作用域,即作用域0
	cgBlock(subFunc, exp.Block)
	subFunc.exitScope()  //退出函数作用域,最高层次的作用域
	subFunc.RETURN(0, 0) //lua默认每个方法最后都会加上RETURN指令
	subFunc.recordInsLine(exp.LastLine)

	//生成CLOSURE
	bx := len(fi.subFuncs) - 1 //该子函数于父函数中的索引
	fi.CLOSURE(a, bx)
	fi.recordInsLine(exp.DefLine)
}

func cgTableConstructExp(fi *funcInfo, exp *ast.TableConstructExp, a int) {
	//NEWTABLE,创建表,计算其表创建所需的数组长度和字典长度
	var nArrs int64 = 0 //记住table的数组部分索引是从1开始的
	nKeys := len(exp.Keys)
	for _, key := range exp.Keys {
		if i, ok := key.(*ast.IntExp); ok && i.Val > nArrs {
			nArrs = i.Val
		}
	}
	fi.NEWTABLE(a, int(nArrs), nKeys-int(nArrs)) //由于外部allocReg从而传入的a所以这里无需allocReg
	fi.recordInsLine(exp.Line)
	//添加元素,包括数组部分和字典部分,利用SETTABLE指令
	for i := 0; i < nKeys; i++ {
		b := fi.allocReg()
		cgExp(fi, exp.Keys[i], b, 1)
		c := fi.allocReg()
		cgExp(fi, exp.Vals[i], c, 1)
		fi.SETTABLE(a, b, c)
		fi.recordInsLine(exp.Line)
		fi.freeRegs(2)
	}

}

func cgConcatExp(fi *funcInfo, exp *ast.ConcatExp, a int) {
	oldUsed := fi.usedRegs
	for _, exp := range exp.Exps {
		idx := fi.allocReg()
		cgExp(fi, exp, idx, 1)
	}
	c := fi.usedRegs - 1
	b := c + 1 - len(exp.Exps) //c-b+1==len(exps)
	fi.CONCAT(a, b, c)
	fi.recordInsLine(exp.Line)
	fi.usedRegs = oldUsed
}

func cgNameExp(fi *funcInfo, exp *ast.NameExp, a int) {
	varName := exp.Name
	// loacal var
	if idx := fi.indexOfLocalVar(varName); idx >= 0 {
		fi.MOVE(a, idx)
		fi.recordInsLine(exp.Line)
		return
	}
	// upvalue
	if idx := fi.indexOfUpvalue(varName); idx >= 0 {
		fi.GETUPVAL(a, idx)
		fi.recordInsLine(exp.Line)
		return
	}
	// global var => _ENV['x']
	cgTableAccessExp(fi, &ast.TableAccessExp{
		LastLine:   exp.Line,
		PrefixExp:  &ast.NameExp{exp.Line, "_ENV"},
		CurrentExp: &ast.StringExp{exp.Line, varName}, //最后的TableAccessExp的CurrentExp类型必须为StringExp，不能使用NameExp，会导致重复cgName方法调用
	}, a)
}

func cgTableAccessExp(fi *funcInfo, exp *ast.TableAccessExp, a int) {
	b := fi.allocReg() //R(B)
	cgExp(fi, exp.PrefixExp, b, 1)
	c := fi.allocReg() //RK(C)
	cgExp(fi, exp.CurrentExp, c, 1)
	//生成获取table键值对指令
	fi.GETTABLE(a, b, c)
	fi.recordInsLine(exp.LastLine)
	fi.freeRegs(2)
}

// n>0则返回n个返回值,n==0则不返回,n<=-1则返回所有返回值
func cgFuncCallExp(fi *funcInfo, exp *ast.FuncCallExp, a, n int) {
	//1.将函数调用所需的方法及其参数的装载指令生成
	lastArgIsVarargOrFuncCall := false
	nArgs := 0 //记录方法一共传递的参数
	//生成装载函数指令
	//OOP类型的函数需要特殊处理
	if ta, ok := exp.Method.(*ast.TableAccessExp); ok {
		cgExp(fi, ta.PrefixExp, a, 1) //PrefixExp应为NameExp，所以这里处理后a处应为对象实例
		keyExp := ta.CurrentExp.(*ast.StringExp)
		//将其方法加入该函数的常量池中
		fi.allocReg()                                                  //SELF会用到R(A+1),所以需要申请一个寄存器空间
		c := instruction.ConstantBase + fi.indexOfConstant(keyExp.Str) //这里为什么加0x100 参考vm.GetRK方法
		fi.SELF(a, a, c)
		fi.recordInsLine(keyExp.Line)
		nArgs++ //自身需要作为参数传入
	} else {
		//普通类型的方法直接生成，将方法定义move到a中
		cgExp(fi, exp.Method, a, 1)
	}
	//生成装载参数指令,按顺序压入栈
	//这里可能需要多个寄存器空间,所以记录开始usedRegs好方便后续释放
	oldUsed := fi.usedRegs
	for i, arg := range exp.Exps {
		a := fi.allocReg()
		if i == len(exp.Exps)-1 && _isVarargOrFuncCall(arg) { //最后参数为vararg或funcCall
			lastArgIsVarargOrFuncCall = true
			cgExp(fi, arg, a, -1) //-1，这里接受方法调用的所有返回值
		} else {
			cgExp(fi, arg, a, 1)
		}
		nArgs++
	}
	fi.usedRegs = oldUsed

	if lastArgIsVarargOrFuncCall {
		nArgs = -1
	}

	//2.生成函数调用指令
	fi.CALL(a, nArgs, n)
	fi.recordInsLine(exp.LastLine)
}

func cgUnitaryOpExp(fi *funcInfo, exp *ast.UnitaryOpExp, a int) {
	b := fi.allocReg()
	cgExp(fi, exp.A, b, 1)
	fi.UnitaryOP(exp.Line, exp.Op, a, b)
	fi.freeReg()
}

func cgDualOpExp(fi *funcInfo, exp *ast.DualOpExp, a int) {
	switch exp.Op {
	case lexer.TOKEN_OP_OR, lexer.TOKEN_OP_AND:
		b := fi.allocReg()
		cgExp(fi, exp.A, b, 1)
		fi.freeReg()
		if exp.Op == lexer.TOKEN_OP_OR {
			fi.TESTSET(a, b, 1) //or:第一个表达式为真就为真
		} else {
			fi.TESTSET(a, b, 0) //and:第一个表达式为假就为假
		}
		fi.recordInsLine(exp.Line)

		//or和and具有短路效应
		jmpToEnd := fi.JMP(0, 0)
		fi.recordInsLine(exp.Line)
		//经过第一个表达式的TESTSET判断到这里
		//则说明影响最后真假的只需看第二个表达式的值
		//即 false or exp2 & true and exp2
		c := fi.allocReg()
		cgExp(fi, exp.B, c, 1)
		fi.freeReg()
		fi.MOVE(a, c) //直接将表达式2的真假传给R(A)
		fi.recordInsLine(exp.Line)

		fi.fixSbx(jmpToEnd, fi.pc()-jmpToEnd)
	default:
		b := fi.allocReg()
		cgExp(fi, exp.A, b, 1)
		c := fi.allocReg()
		cgExp(fi, exp.B, c, 1)
		fi.DualOP(exp.Line, exp.Op, a, b, c)
		fi.freeRegs(2)
	}
}
