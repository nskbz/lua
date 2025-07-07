package codegen

import (
	"nskbz.cn/lua/binchunk"
	"nskbz.cn/lua/compile/ast"
)

// 代码生成
func GenProto(chunck *ast.Block, chunckName string) *binchunk.Prototype {
	fd := &ast.FuncDefExp{
		DefLine:  0,
		LastLine: chunck.LastLine,
		ArgList:  []string{},
		IsVararg: true,
		Block:    chunck,
	} //主函数
	fi := newFuncInfo(nil, "", fd)
	fi.newLocalVar("_ENV", 0, chunck.LastLine) //全局变量实质为最上层的局部变量
	idx := fi.allocReg()
	cgFuncDefExp(fi, chunckName, fd, idx) //指令生成过程中已经记录了Regs的最大使用数量
	fi.freeReg()
	proto := toProto(fi.subFuncs[0])
	return proto
}
