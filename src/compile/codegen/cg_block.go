package codegen

import (
	"nskbz.cn/lua/compile/ast"
)

func cgBlock(fi *funcInfo, block *ast.Block) {
	for _, stat := range block.Stats {
		cgStat(fi, stat, block.LastLine)
	}
	if block.RetExps != nil {
		cgRetStat(fi, block.RetExps, block.LastLine)
	}
}

// 生成方法返回指令
func cgRetStat(fi *funcInfo, exps []ast.Exp, blockLastLine int) {
	//没有返回值，直接return
	if exps == nil {
		fi.RETURN(0, 0)
		fi.recordInsLine(blockLastLine)
		return
	}
	retNum := len(exps)
	multRet := _isVarargOrFuncCall(exps[retNum-1]) //返回表达式中是否有vararg和函数调用

	for i, exp := range exps {
		idx := fi.allocReg() //预留返回值的位置，模拟但不会真的存值，就是一个占位的，模拟真实的函数调用栈中的寄存器行为
		if i == retNum-1 && multRet {
			cgExp(fi, exp, idx, 0)
		}
		cgExp(fi, exp, idx, 1)
	}
	fi.freeRegs(retNum) //释放占位的寄存器
	a := fi.usedRegs    //返回值寄存器的起始索引
	if multRet {        //如果包含vararg或函数调用的返回语句，需要特殊处理
		fi.RETURN(a, -1) //-1即返回所有返回值
		fi.recordInsLine(blockLastLine)
		return
	}
	fi.RETURN(a, retNum)
	fi.recordInsLine(blockLastLine)
}

func _isVarargOrFuncCall(exp ast.Exp) bool {
	switch exp.(type) {
	case *ast.VarargExp, *ast.FuncCallExp:
		return true
	}
	return false
}
