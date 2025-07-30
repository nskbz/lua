package parser

import (
	"nskbz.cn/lua/compile/ast"
	"nskbz.cn/lua/compile/lexer"
	"nskbz.cn/lua/number"
)

// 解析名称(标识符)列表
// namelist ::=Name { ',' Name}
// Name为普通的变量名
func parseIdentifierList(l *lexer.Lexer) []string {
	names, hasVararg := parseParamList(l)
	if hasVararg {
		panic("here are vararg!error!")
	}
	return names
}

// 解析变量名列表
// varlist ::=var { ',' var}
// var ::=Name | prefixexp '[' exp ']' | prefixexp '.' Name
// 可以看到var的规则更加复杂
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

// 解析参数列表
// function (parlist) funcbody
// parlist ::= namelist [‘,’ ‘...’] | ‘...’
// 不同于namelist，parlist最后可以有可变参数
func parseParamList(l *lexer.Lexer) (names []string, hasVararg bool) {
	if l.CheckToken(lexer.TOKEN_SEP_RPAREN) { //0个参数
		return
	} else if l.CheckToken(lexer.TOKEN_VARARG) { //参数只有一个且为可变参数
		hasVararg = true
		l.NextToken() //skip '...'
		return
	}
	//解析一般情况即若干个name+可选的一个vararg
	t := l.AssertIdentifier()
	names = append(names, t.Val())
	l.NextToken() //skip current param
	for l.CheckToken(lexer.TOKEN_SEP_COMMA) {
		l.NextToken() //skip ','
		if l.CheckToken(lexer.TOKEN_VARARG) {
			hasVararg = true
			l.NextToken() //skip '...'
			break
		}
		t = l.AssertIdentifier()
		names = append(names, t.Val())
		l.NextToken()
	}
	return
}

// 这个方法只解析方法体即function和方法名后面的部分
// (parList) block end
//
// for example:
// -- 将匿名函数（没有函数名只有定义）赋值给变量，右侧就是表达式，即函数定义
// local myFunc = function(a, b)
//
//	return a + b
//
// end
// print(myFunc(3, 5))  -- 输出: 8
func parseFuncDefExp(l *lexer.Lexer, lineForFunc int) ast.Exp {
	l.AssertAndSkipToken(lexer.TOKEN_SEP_LPAREN) //skip '('
	pars, hasVararg := parseParamList(l)
	l.AssertAndSkipToken(lexer.TOKEN_SEP_RPAREN) //skip ')'

	block := parseBlock(l)
	lastLine := l.AssertAndSkipToken(lexer.TOKEN_KW_END).Line()
	return &ast.FuncDefExp{
		DefLine:  lineForFunc,
		LastLine: lastLine,
		ArgList:  pars,
		IsVararg: hasVararg,
		Block:    block,
	}
}

func parseExpList(l *lexer.Lexer) []ast.Exp {
	exps := []ast.Exp{}
	if exp := parseExp(l); exp != nil {
		exps = append(exps, exp)
	}
	for l.CheckToken(lexer.TOKEN_SEP_COMMA) {
		l.NextToken() //skip ','
		if exp := parseExp(l); exp != nil {
			exps = append(exps, exp)
		}
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
		exp2 := _parseExp9(l)
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
// 这些都是右结合的运算符
func _parseExp2(l *lexer.Lexer) ast.Exp {
	exp := _parseExp1(l)
	//如果exp不为空则说明应当为减法而非负数
	if exp != nil {
		return exp
	}

	// if exp!=nil &&l.CheckToken(lexer.TOKEN_OP_SUB){
	// 	t := l.NextToken()//skip '-'
	// 	exp =&ast.DualOpExp{
	// 		Line: t.Line(),
	// 		Op: t.Kind(),
	// 		A: exp,
	// 		B: _parseExp2(l),
	// 	}
	// }
	if l.CheckToken(lexer.TOKEN_OP_NOT) || l.CheckToken(lexer.TOKEN_OP_LEN) ||
		l.CheckToken(lexer.TOKEN_OP_UNM) || l.CheckToken(lexer.TOKEN_OP_BNOT) {
		t := l.NextToken()
		exp = &ast.UnitaryOpExp{
			Line: t.Line(),
			Op:   t.Kind(),
			A:    _parseExp2(l),
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
	case lexer.TOKEN_VARARG:
		t := l.NextToken() //skip '...'
		return &ast.VarargExp{
			Line: t.Line(),
		}
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
		lineOfFunc := l.AssertAndSkipToken(lexer.TOKEN_KW_FUNCTION).Line() //记录函数定义开始的行号
		return parseFuncDefExp(l, lineOfFunc)
	default:
		return parsePrefixExp(l)
	}
}

func parseNumberExp(l *lexer.Lexer) ast.Exp {
	t := l.NextToken()
	if i, ok := number.ParseInteger(t.Val()); ok {
		return &ast.IntExp{
			Line: t.Line(),
			Val:  i,
		}
	}
	if f, ok := number.ParseFloat(t.Val()); ok {
		return &ast.FloatExp{
			Line: t.Line(),
			Val:  f,
		}
	}
	panic("parse number error:" + t.Val())
}

// 数组式表
//
//	local array = {10, 20, 30, 40, 50}
//
// -- 访问元素
// print(array[1])  -- 输出: 10
// print(array[3])  -- 输出: 30
//
// 字典表（键值对）:
//
//	local dict = {
//	    name = "Lua",
//	    version = "5.4",
//	    type = "scripting language"
//	}
//
// -- 访问元素
// print(dict["name"])  -- 输出: Lua
// print(dict.version)  -- 输出: 5.4
//
// 混合式表:
//
//	local mixed = {
//	    "apple",          -- 数组索引1
//	    color = "yellow", -- 字典部分
//	    count = 5         -- 字典部分
//	    "banana",         -- 数组索引2
//	}
//
// print(mixed[1])       -- 输出: apple
// print(mixed.color)    -- 输出: yellow
//
// 表构造中使用表达式:
//
//	local x,y=14,16
//	local tableWithExpr = {
//	    sum = x + y,
//	    product = x * y,
//	    [x + y] = "thirty"
//	}
//
// print(tableWithExpr.sum)      -- 输出: 30
// print(tableWithExpr[30])      -- 输出: thirty
func parseTableConstructorExp(l *lexer.Lexer) ast.Exp {
	line := l.Line() //start line
	l.NextToken()    //skip '{'

	keys := []ast.Exp{}
	values := []ast.Exp{}

	if !l.CheckToken(lexer.TOKEN_SEP_RCURLY) {
		i := 1
		key, value := _parseTableField(l)
		if key == nil {
			key = &ast.IntExp{Line: l.Line(), Val: int64(i)}
			i++
		}
		keys = append(keys, key)
		values = append(values, value)
		for _checkTableFieldEnd(l) {
			l.NextToken() //skip ','|';'
			key, value = _parseTableField(l)
			if key == nil && value == nil { //都为空则直接跳出
				break
			}
			if key == nil {
				key = &ast.IntExp{Line: l.Line(), Val: int64(i)}
				i++
			}
			keys = append(keys, key)
			values = append(values, value)
		}
	}

	l.AssertAndSkipToken(lexer.TOKEN_SEP_RCURLY)
	return &ast.TableConstructExp{
		Line:     line,
		LastLine: l.Line(),
		Keys:     keys,
		Vals:     values,
	}
}

func _checkTableFieldEnd(l *lexer.Lexer) bool {
	return l.CheckToken(lexer.TOKEN_SEP_COMMA) || l.CheckToken(lexer.TOKEN_SEP_SEMI)
}

// field ::= '[' exp ']' '=' exp |	Name '=' exp | exp
//
// '[' exp ']' '=' exp |	Name '=' exp ==> k,v
// exp ==> nil,v
// 结尾 ==> nil,nil
func _parseTableField(l *lexer.Lexer) (k, v ast.Exp) {
	//考虑特殊情况
	//即表构造表达式以;}结尾，如下：
	// local person = {
	//     name = "Alice";
	//     age = 25;
	// }
	if l.CheckToken(lexer.TOKEN_SEP_RCURLY) {
		return
	}

	// {["x"]=y}
	if l.CheckToken(lexer.TOKEN_SEP_LBRACK) {
		l.NextToken() //skip '['
		k = parseExp(l)
		l.AssertAndSkipToken(lexer.TOKEN_SEP_RBRACK) //skip ']'
		l.AssertAndSkipToken(lexer.TOKEN_OP_ASSIGN)  // skip '='
		v = parseExp(l)
		return
	}

	// {x=y}
	// 将语法糖转换成{["x"]=y}
	if l.CheckToken(lexer.TOKEN_IDENTIFIER) {
		identifier := l.NextToken()
		if l.CheckToken(lexer.TOKEN_OP_ASSIGN) {
			k = &ast.StringExp{Line: identifier.Line(), Str: identifier.Val()}
			l.NextToken() //skip '='
			v = parseExp(l)
		}
		// t := l.NextToken()
		// v = &ast.StringExp{Line: t.Line(), Str: t.Val()}
		return
	}

	// {x}
	v = parseExp(l)
	return
}
