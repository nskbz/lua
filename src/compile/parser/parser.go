package parser

import (
	"nskbz.cn/lua/compile/ast"
	"nskbz.cn/lua/compile/lexer"
)

func Parse(chunk []byte, chunkName string) *ast.Block {
	l := lexer.NewLexer(chunk, chunkName)
	block := parseBlock(l)
	l.AssertAndSkipToken(lexer.TOKEN_EOF)
	return block
}
