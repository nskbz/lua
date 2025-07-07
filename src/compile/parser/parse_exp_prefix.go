package parser

import (
	"nskbz.cn/lua/compile/ast"
	"nskbz.cn/lua/compile/lexer"
)

// prefixexp ::= var | functioncall | ‘(’ exp ‘)’
// var ::=  Name | prefixexp ‘[’ exp ‘]’ | prefixexp ‘.’ Name
// functioncall ::=  prefixexp args | prefixexp ‘:’ Name args
func parsePrefixExp(l *lexer.Lexer) ast.Exp {
	var first ast.Exp
	if l.CheckToken(lexer.TOKEN_IDENTIFIER) {
		t := l.NextToken()
		first = &ast.NameExp{
			Line: t.Line(),
			Name: t.Val(),
		}
	} else if l.CheckToken(lexer.TOKEN_SEP_LPAREN) {
		first = parseParenExp(l)
	}
	//else {
	// 	panic("no support this prefix class")
	// }

	return _finishPrefixExp(l, first)
}

func _finishPrefixExp(l *lexer.Lexer, exp ast.Exp) ast.Exp {
	//exp进来的类型为NameExp | ParenExp
	for {
		if l.CheckToken(lexer.TOKEN_SEP_DOT) {
			l.NextToken() //skip '.'
			identifier := l.AssertAndSkipToken(lexer.TOKEN_IDENTIFIER)
			exp = &ast.TableAccessExp{
				LastLine:  identifier.Line(),
				PrefixExp: exp,
				CurrentExp: &ast.StringExp{
					Line: identifier.Line(),
					Str:  identifier.Val(),
				},
			}
		} else if l.CheckToken(lexer.TOKEN_SEP_LBRACK) {
			l.NextToken() //skip '['
			keyExp := parseExp(l)
			l.AssertAndSkipToken(lexer.TOKEN_SEP_RBRACK) //skip ']'
			exp = &ast.TableAccessExp{
				LastLine:   l.Line(),
				PrefixExp:  exp,
				CurrentExp: keyExp,
			}
		} else if l.CheckToken(lexer.TOKEN_SEP_COLON) { // class:myfunc ...
			l.NextToken() //skip ':'
			key := l.AssertAndSkipToken(lexer.TOKEN_IDENTIFIER)
			exp = &ast.TableAccessExp{
				LastLine:  l.Line(),
				PrefixExp: exp,
				CurrentExp: &ast.StringExp{
					Line: l.Line(),
					Str:  key.Val(),
				},
			}
		} else if l.CheckToken(lexer.TOKEN_SEP_LPAREN) { // myfunc ("hello")
			exp = _parseStandardFuncCallExp(l, exp)
		} else if l.CheckToken(lexer.TOKEN_SEP_LCURLY) { //	myfunc {"hello"}
			exp = _parseTableFuncCallExp(l, exp)
		} else if l.CheckToken(lexer.TOKEN_STRING) { //	myfunc "hello"
			exp = _parseStringFuncCallExp(l, exp)
		} else {
			return exp
		}
	}
}

func parseParenExp(l *lexer.Lexer) ast.Exp {
	l.AssertAndSkipToken(lexer.TOKEN_SEP_LPAREN) //skip '('
	exp := parseExp(l)
	l.AssertAndSkipToken(lexer.TOKEN_SEP_RPAREN) //skip ')'
	switch exp.(type) {
	case *ast.VarargExp, *ast.FuncCallExp, *ast.NameExp, *ast.TableAccessExp:
		return &ast.ParensExp{
			Exp: exp,
		}
	}
	return exp
}

func _parseStandardFuncCallExp(l *lexer.Lexer, method ast.Exp) ast.Exp {
	line := l.AssertAndSkipToken(lexer.TOKEN_SEP_LPAREN).Line() //skip '('
	exps := parseExpList(l)
	l.AssertAndSkipToken(lexer.TOKEN_SEP_RPAREN) //skip ')'
	return &ast.FuncCallExp{
		Line:     line,
		LastLine: l.Line(),
		Method:   method,
		Exps:     exps,
	}
}

func _parseStringFuncCallExp(l *lexer.Lexer, method ast.Exp) ast.Exp {
	t := l.AssertAndSkipToken(lexer.TOKEN_STRING)
	exps := []ast.Exp{&ast.StringExp{
		Line: t.Line(),
		Str:  t.Val(),
	}}
	return &ast.FuncCallExp{
		Line:     t.Line(),
		LastLine: t.Line(),
		Method:   method,
		Exps:     exps,
	}
}

func _parseTableFuncCallExp(l *lexer.Lexer, method ast.Exp) ast.Exp {
	t := l.AssertAndSkipToken(lexer.TOKEN_SEP_LCURLY) //skip '{'
	table := parseTableConstructorExp(l)
	exps := []ast.Exp{table}
	l.AssertAndSkipToken(lexer.TOKEN_SEP_RCURLY) //skip '}'
	return &ast.FuncCallExp{
		Line:     t.Line(),
		LastLine: t.Line(),
		Method:   method,
		Exps:     exps,
	}
}
