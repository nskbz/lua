package parser

import (
	"nskbz.cn/lua/compile/ast"
	"nskbz.cn/lua/compile/lexer"
)

func parseIdentifierList(l *lexer.Lexer) []string {
	names := []string{}
	t := l.AssertIdentifier()
	names = append(names, t.Val())
	l.NextToken()
	for l.CheckToken(lexer.TOKEN_SEP_COMMA) {
		l.NextToken()
		nt := l.AssertIdentifier()
		names = append(names, nt.Val())
		l.NextToken()
	}
	return names
}

func parseVarList(l *lexer.Lexer) []ast.Exp {
	exps := []ast.Exp{}
	v0 := parsePrefixExp(l) //解析第一个var
	exps = append(exps, _checkVar(v0))
	for l.CheckToken(lexer.TOKEN_SEP_COMMA) {
		l.NextToken() //skip ','
		v := parsePrefixExp(l)
		exps = append(exps, _checkVar(v))
	}
	return exps
}

// var ::=Name | prefixexp '[' exp ']' | prefixexp '.' Name
func _checkVar(exp ast.Exp) ast.Exp {
	switch exp.(type) {
	case *ast.NameExp, *ast.TableAccessExp:
		return exp
	}
	// l.AssertAndSkipToken(-1) ???????
	panic("unreachable")
}

// 这个方法只解析方法体即function和方法名后面的部分
//
// (parList) block end
func parseFuncDefExp(l *lexer.Lexer) ast.Exp {
	l.AssertAndSkipToken(lexer.TOKEN_SEP_LPAREN) //skip '('
	pars := []string{}                           //to do
	l.AssertAndSkipToken(lexer.TOKEN_SEP_RPAREN) //skip ')'

	block := parseBlock(l)
	lastLine := l.AssertAndSkipToken(lexer.TOKEN_KW_END).Line()
	return &ast.FuncDefExp{
		Line:     0,
		LastLine: lastLine,
		ParList:  pars,
		IsVararg: false,
		Block:    block,
	}
}

func parsePrefixExp(l *lexer.Lexer) ast.Exp {
	return &ast.PrefixExp{}
}

func parseExpList(l *lexer.Lexer) []ast.Exp {
	exps := []ast.Exp{}
	exps = append(exps, parseExp(l))
	for l.CheckToken(lexer.TOKEN_SEP_COMMA) {
		l.NextToken() //skip ','
		exps = append(exps, parseExp(l))
		l.NextToken() //skip exp
	}
	return exps
}

/*
exp ::=  nil | false | true | Numeral | LiteralString | ‘...’ | functiondef |
	 prefixexp | tableconstructor | exp binop exp | unop exp
*/
/*
exp   ::= exp12
exp12 ::= exp11 {or exp11}
exp11 ::= exp10 {and exp10}
exp10 ::= exp9 {(‘<’ | ‘>’ | ‘<=’ | ‘>=’ | ‘~=’ | ‘==’) exp9}
exp9  ::= exp8 {‘|’ exp8}
exp8  ::= exp7 {‘~’ exp7}
exp7  ::= exp6 {‘&’ exp6}
exp6  ::= exp5 {(‘<<’ | ‘>>’) exp5}
exp5  ::= exp4 {‘..’ exp4}
exp4  ::= exp3 {(‘+’ | ‘-’) exp3}
exp3  ::= exp2 {(‘*’ | ‘/’ | ‘//’ | ‘%’) exp2}
exp2  ::= {(‘not’ | ‘#’ | ‘-’ | ‘~’)} exp1
exp1  ::= exp0 {‘^’ exp2}
exp0  ::= nil | false | true | Numeral | LiteralString| ‘...’ | functiondef | prefixexp | tableconstructor
*/
func parseExp(l *lexer.Lexer) ast.Exp {
	return _parseExp12(l)
}

// exp12 ::= exp11 {or exp11}
func _parseExp12(l *lexer.Lexer) ast.Exp {
	exp := _parseExp11(l)
	for l.CheckToken(lexer.TOKEN_OP_OR) {
		t := l.NextToken() //skip 'or'
		//or运算符是左结合的
		//例如，表达式 a or b or c 会被解释为 (a or b) or c，先计算 a or b，
		//如果 a 为真，则整个表达式的结果就是 a，不会计算 b，
		//如果 a 为假，则计算 b，如果 b 也为假，再计算 c，以此类推。
		//错误做法：
		// exp2 := parseExp(l)  =>这种递归的形式是右结合，这种是将双目运算符的B表达式作为一个整体，递归解析，是为右结合
		// 正确做法:
		// 通过循环，依次将两个表达式合并为下一个双目运算符的A表达式，即左结合
		exp2 := _parseExp11(l)
		exp = &ast.DualOpExp{
			Line: t.Line(),
			Op:   t.Kind(),
			A:    exp,
			B:    exp2,
		}
	}
	return exp
}

// exp11 ::= exp10 {and exp10}
func _parseExp11(l *lexer.Lexer) ast.Exp {
	exp := _parseExp10(l)
	for l.CheckToken(lexer.TOKEN_OP_AND) {
		t := l.NextToken() //skip 'and'
		//and运算符是左结合的
		exp2 := _parseExp10(l)
		exp = &ast.DualOpExp{
			Line: t.Line(),
			Op:   t.Kind(),
			A:    exp,
			B:    exp2,
		}
	}
	return exp
}

// exp10 ::= exp9 {(‘<’ | ‘>’ | ‘<=’ | ‘>=’ | ‘~=’ | ‘==’) exp9}
func _parseExp10(l *lexer.Lexer) ast.Exp {
	exp := _parseExp9(l)
	for l.CheckToken(lexer.TOKEN_OP_GT) || l.CheckToken(lexer.TOKEN_OP_LT) || l.CheckToken(lexer.TOKEN_OP_LE) ||
		l.CheckToken(lexer.TOKEN_OP_GE) || l.CheckToken(lexer.TOKEN_OP_EQ) || l.CheckToken(lexer.TOKEN_OP_NE) {
		t := l.NextToken()
		exp2 := _parseExp9
		exp = &ast.DualOpExp{
			Line: t.Line(),
			Op:   t.Kind(),
			A:    exp,
			B:    exp2,
		}
	}
	return exp
}

// exp9  ::= exp8 {‘|’ exp8}
func _parseExp9(l *lexer.Lexer) ast.Exp {
	exp := _parseExp8(l)
	for l.CheckToken(lexer.TOKEN_OP_BOR) {
		t := l.NextToken()
		exp2 := _parseExp8(l)
		exp = &ast.DualOpExp{
			Line: t.Line(),
			Op:   t.Kind(),
			A:    exp,
			B:    exp2,
		}
	}
	return exp
}

// exp8  ::= exp7 {‘~’ exp7}
func _parseExp8(l *lexer.Lexer) ast.Exp {
	exp := _parseExp7(l)
	for l.CheckToken(lexer.TOKEN_OP_WAVE) {
		t := l.NextToken()
		exp2 := _parseExp7(l)
		exp = &ast.DualOpExp{
			Line: t.Line(),
			Op:   t.Kind(),
			A:    exp,
			B:    exp2,
		}
	}
	return exp
}

// exp7  ::= exp6 {‘&’ exp6}
func _parseExp7(l *lexer.Lexer) ast.Exp {
	exp := _parseExp6(l)
	for l.CheckToken(lexer.TOKEN_OP_BAND) {
		t := l.NextToken()
		exp2 := _parseExp6(l)
		exp = &ast.DualOpExp{
			Line: t.Line(),
			Op:   t.Kind(),
			A:    exp,
			B:    exp2,
		}
	}
	return exp
}

// exp6  ::= exp5 {(‘<<’ | ‘>>’) exp5}
func _parseExp6(l *lexer.Lexer) ast.Exp {
	exp := _parseExp5(l)
	for l.CheckToken(lexer.TOKEN_OP_SHR) || l.CheckToken(lexer.TOKEN_OP_SHL) {
		t := l.NextToken()
		exp2 := _parseExp5(l)
		exp = &ast.DualOpExp{
			Line: t.Line(),
			Op:   t.Kind(),
			A:    exp,
			B:    exp2,
		}
	}
	return exp
}

// exp5  ::= exp4 {‘..’ exp4}
func _parseExp5(l *lexer.Lexer) ast.Exp {
	exp := _parseExp4(l)
	exps := []ast.Exp{exp}
	var t lexer.Token
	for l.CheckToken(lexer.TOKEN_OP_CONCAT) {
		t = l.NextToken()
		exps = append(exps, _parseExp4(l))
	}
	if len(exps) > 1 {
		exp = &ast.ConcatExp{
			Line: t.Line(),
			Exps: exps,
		}
	}
	return exp
}

// exp4  ::= exp3 {(‘+’ | ‘-’) exp3}
func _parseExp4(l *lexer.Lexer) ast.Exp {
	exp := _parseExp3(l)
	for l.CheckToken(lexer.TOKEN_OP_ADD) || l.CheckToken(lexer.TOKEN_OP_SUB) {
		t := l.NextToken()
		exp2 := _parseExp3(l)
		exp = &ast.DualOpExp{
			Line: t.Line(),
			Op:   t.Kind(),
			A:    exp,
			B:    exp2,
		}
	}
	return exp
}

// exp3  ::= exp2 {(‘*’ | ‘/’ | ‘//’ | ‘%’) exp2}
func _parseExp3(l *lexer.Lexer) ast.Exp {
	exp := _parseExp2(l)
	for l.CheckToken(lexer.TOKEN_OP_MUL) || l.CheckToken(lexer.TOKEN_OP_DIV) ||
		l.CheckToken(lexer.TOKEN_OP_IDIV) || l.CheckToken(lexer.TOKEN_OP_MOD) {
		t := l.NextToken()
		exp2 := _parseExp2(l)
		exp = &ast.DualOpExp{
			Line: t.Line(),
			Op:   t.Kind(),
			A:    exp,
			B:    exp2,
		}
	}
	return exp
}

// exp2  ::= {(‘not’ | ‘#’ | ‘-’ | ‘~’)} exp1
func _parseExp2(l *lexer.Lexer) ast.Exp {
	exp := _parseExp1(l)
	if l.CheckToken(lexer.TOKEN_OP_NOT) || l.CheckToken(lexer.TOKEN_OP_LEN) ||
		l.CheckToken(lexer.TOKEN_OP_UNM) || l.CheckToken(lexer.TOKEN_OP_BNOT) {
		t := l.NextToken()
		exp = &ast.UnitaryOpExp{
			Line: t.Line(),
			Op:   t.Kind(),
			A:    _parseExp1(l),
		}
	}
	return exp
}

// exp1  ::= exp0 {‘^’ exp2}
// 特别的'^'乘方运算符是右结合
func _parseExp1(l *lexer.Lexer) ast.Exp {
	exp := _parseExp0(l)
	if l.CheckToken(lexer.TOKEN_OP_POW) {
		t := l.NextToken()
		exp = &ast.DualOpExp{
			Line: t.Line(),
			Op:   t.Kind(),
			A:    exp,
			B:    _parseExp2(l), //右结合则把'^'后的B表达式看作一个整体解析
		}
	}
	return exp
}

// exp0  ::= nil | false | true | Numeral | LiteralString| ‘...’ | functiondef | prefixexp | tableconstructor
func _parseExp0(l *lexer.Lexer) ast.Exp {
	nt := l.LookToken()
	switch nt.Kind() {
	case lexer.TOKEN_KW_NIL:
		t := l.NextToken()
		return &ast.NilExp{Line: t.Line()}
	case lexer.TOKEN_KW_TRUE:
		t := l.NextToken()
		return &ast.TrueExp{Line: t.Line()}
	case lexer.TOKEN_KW_FALSE:
		t := l.NextToken()
		return &ast.FalseExp{Line: t.Line()}
	case lexer.TOKEN_NUMBER:
		return parseNumberExp(l)
	case lexer.TOKEN_STRING:
		t := l.NextToken()
		return &ast.StringExp{Line: t.Line(), Str: t.Val()}
	case lexer.TOKEN_SEP_LCURLY:
		return parseTableConstructorExp(l)
	case lexer.TOKEN_KW_FUNCTION:
		l.NextToken() //skip 'function'
		return parseFuncDefExp(l)
	default:
		return parsePrefixExp(l)
	}
}

func parseNumberExp(l *lexer.Lexer) ast.Exp {
	return nil
}

func parseTableConstructorExp(l *lexer.Lexer) ast.Exp {
	return nil
}
