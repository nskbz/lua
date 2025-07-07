package compile

import (
	"nskbz.cn/lua/binchunk"
	"nskbz.cn/lua/compile/codegen"
	"nskbz.cn/lua/compile/parser"
)

func Compile(chunck []byte, chunckname string) *binchunk.Prototype {
	block := parser.Parse(chunck, chunckname) //词法分析+语法分析=>AST(抽象语法树)
	proto := codegen.GenProto(block, chunckname)
	return proto
}
