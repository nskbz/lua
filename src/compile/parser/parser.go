package parser

import (
	"nskbz.cn/lua/compile/ast"
	"nskbz.cn/lua/compile/lexer"
)

func Parse(chunk []byte, chunkName string) *ast.Block {
	l := lexer.NewLexer(chunk, chunkName) //词法分析
	block := parseBlock(l)                //语法分析
	l.AssertAndSkipToken(lexer.TOKEN_EOF)
	return block
}
