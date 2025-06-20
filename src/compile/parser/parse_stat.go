package parser

import (
	"nskbz.cn/lua/compile/ast"
	"nskbz.cn/lua/compile/lexer"
)

func parseStat(l *lexer.Lexer) ast.Stat {
	var stat ast.Stat
	nt := l.LookToken()
	switch nt.Kind() {
	case lexer.TOKEN_SEP_SEMI:
		stat = parseEmptyStat(l)
	case lexer.TOKEN_KW_BREAK:
		stat = parseBreakStat(l)
	case lexer.TOKEN_SEP_LABEL:
		stat = parseLabelStat(l)
	case lexer.TOKEN_KW_GOTO:
		stat = parseGotoStat(l)
	case lexer.TOKEN_KW_DO:
		stat = parseDoStat(l)
	case lexer.TOKEN_KW_WHILE:
		stat = parseWhileStat(l)
	case lexer.TOKEN_KW_REPEAT:
		stat = parseRepeatStat(l)
	case lexer.TOKEN_KW_IF:
		stat = parseIfStat(l)
	case lexer.TOKEN_KW_FOR:
		stat = parseForStat(l)
	case lexer.TOKEN_KW_LOCAL:
		stat = parseLocalValOrFuncStat(l)
	case lexer.TOKEN_KW_FUNCTION:
		stat = parseFuncDefStat(l)
	default:
		stat = parseAssignOrFuncCallStat(l)
	}
	return stat
}

func parseEmptyStat(l *lexer.Lexer) ast.Stat {
	l.NextToken()
	return &ast.EmptyStat{}
}

func parseBreakStat(l *lexer.Lexer) ast.Stat {
	l.NextToken()
	nt := l.LookToken()
	return &ast.BreakStat{
		Line: nt.Line(),
	}
}

func parseLabelStat(l *lexer.Lexer) ast.Stat {
	l.NextToken() //skip '::'
	identifier := l.AssertAndSkipToken(lexer.TOKEN_IDENTIFIER)
	l.AssertAndSkipToken(lexer.TOKEN_SEP_LABEL)
	return &ast.LabelStat{Name: identifier.Val()}
}

func parseGotoStat(l *lexer.Lexer) ast.Stat {
	l.NextToken()
	identifier := l.AssertAndSkipToken(lexer.TOKEN_IDENTIFIER)
	return &ast.GotoStat{
		Name: identifier.Val(),
	}
}

func parseDoStat(l *lexer.Lexer) ast.Stat {
	l.AssertAndSkipToken(lexer.TOKEN_KW_DO) //skip 'do'
	block := parseBlock(l)
	l.AssertAndSkipToken(lexer.TOKEN_KW_END)
	return &ast.DoStat{
		Block: block,
	}
}

func parseWhileStat(l *lexer.Lexer) ast.Stat {
	l.NextToken() //skip 'while'
	exp := parseExp(l)
	doStat := parseDoStat(l).(*ast.DoStat)
	return &ast.WhileStat{
		Exp:   exp,
		Block: doStat.Block,
	}
}

func parseRepeatStat(l *lexer.Lexer) ast.Stat {
	l.NextToken() //skip 'repeat'
	block := parseBlock(l)
	l.AssertAndSkipToken(lexer.TOKEN_KW_UNTIL)
	//l.AssertAndSkipToken(lexer.TOKEN_SEP_LPAREN)
	exp := parseExp(l)
	//l.AssertAndSkipToken(lexer.TOKEN_SEP_RPAREN)
	return &ast.RepeatStat{
		Exp:   exp,
		Block: block,
	}
}

func parseIfStat(l *lexer.Lexer) ast.Stat {
	exps := []ast.Exp{}
	blocks := []*ast.Block{}

	//处理if
	l.NextToken() //skip 'if'
	exps = append(exps, parseExp(l))
	l.AssertAndSkipToken(lexer.TOKEN_KW_THEN)
	blocks = append(blocks, parseBlock(l))

	//处理elseif
	for l.CheckToken(lexer.TOKEN_KW_ELSEIF) {
		if l.CheckToken(lexer.TOKEN_KW_ELSE) {
			break
		}
		l.NextToken() //skip 'elseif'
		exps = append(exps, parseExp(l))
		l.AssertAndSkipToken(lexer.TOKEN_KW_THEN) //skip 'then'
		blocks = append(blocks, parseBlock(l))
	}

	//处理else
	if l.CheckToken(lexer.TOKEN_KW_ELSE) {
		l.NextToken() //skip 'else'
		exps = append(exps, ast.TrueExp{Line: l.Line()})
		blocks = append(blocks, parseBlock(l))
	}

	l.AssertAndSkipToken(lexer.TOKEN_KW_END)
	return &ast.IfStat{
		Exps:   exps,
		Blocks: blocks,
	}
}

func parseForStat(l *lexer.Lexer) ast.Stat {
	f := l.NextToken() //skip 'for'

	names := []string{}
	exps := []ast.Exp{}

	name := l.AssertAndSkipToken(lexer.TOKEN_IDENTIFIER)
	names = append(names, name.Val())
	if l.CheckToken(lexer.TOKEN_OP_ASSIGN) {
		return _parseForNumStat(l, f.Line(), names, exps)
	} else if l.CheckToken(lexer.TOKEN_SEP_COMMA) {
		return _parseForInStat(l, names, exps)
	}
	panic("error parse ForStat")
}

func _parseForNumStat(l *lexer.Lexer, lineOfFor int, names []string, exps []ast.Exp) ast.Stat {
	l.AssertAndSkipToken(lexer.TOKEN_OP_ASSIGN) //skip '='
	exps = append(exps, parseExp(l))            //添加初始化表达式

	l.AssertAndSkipToken(lexer.TOKEN_SEP_COMMA)
	exps = append(exps, parseExp(l)) //添加限制表达式

	if l.CheckToken(lexer.TOKEN_SEP_COMMA) {
		l.NextToken() //skip ','
		exps = append(exps, parseExp(l))
	} else { //没有步长，添加默认步长表达式
		exps = append(exps, &ast.IntExp{
			Line: l.Line(),
			Val:  1,
		})
	}

	lineForDo := l.AssertAndSkipToken(lexer.TOKEN_KW_DO).Line()
	block := parseBlock(l)
	l.AssertAndSkipToken(lexer.TOKEN_KW_END)
	return &ast.ForNumStat{
		LineOfFor: lineOfFor,
		LineOfDo:  lineForDo,
		Name:      names[0],
		Init:      exps[0],
		Limit:     exps[1],
		Step:      exps[2],
		Block:     block,
	}
}

func _parseForInStat(l *lexer.Lexer, names []string, exps []ast.Exp) ast.Stat {
	l.AssertAndSkipToken(lexer.TOKEN_SEP_COMMA)
	names = append(names, parseIdentifierList(l)...)

	l.AssertAndSkipToken(lexer.TOKEN_KW_IN)

	exps = append(exps, parseExpList(l))
	lineForDo := l.AssertAndSkipToken(lexer.TOKEN_KW_DO).Line()
	block := parseBlock(l)
	l.AssertAndSkipToken(lexer.TOKEN_KW_END)
	return &ast.ForInStat{
		LineOfDo: lineForDo,
		NameList: names,
		ExpList:  exps,
		Block:    block,
	}
}

func parseLocalValOrFuncStat(l *lexer.Lexer) ast.Stat {
	l.NextToken() //skip 'local'
	if l.CheckToken(lexer.TOKEN_KW_FUNCTION) {
		return _parseLocalFunc(l)
	}
	return _parseLocalVal(l)
}

func _parseLocalVal(l *lexer.Lexer) ast.Stat {
	names := parseIdentifierList(l) //因为是local开头所以这里都是定义变量，不存在有前缀表达式的变量，所以就直接解析为标识符
	var exps []ast.Exp
	if l.CheckToken(lexer.TOKEN_OP_ASSIGN) {
		l.NextToken()
		exps = parseExpList(l)
	}
	return &ast.LocalVarStat{
		LastLine:     l.Line(),
		LocalVarList: names,
		ExpList:      exps,
	}
}

func _parseLocalFunc(l *lexer.Lexer) ast.Stat {
	l.AssertAndSkipToken(lexer.TOKEN_KW_FUNCTION)
	funcName := l.AssertIdentifier()
	l.NextToken() //skip funcName
	funcDef := parseFuncDefExp(l).(*ast.FuncDefExp)
	return &ast.LocalFuncDefStat{
		Name: funcName.Val(),
		Body: funcDef,
	}
}

func parseAssignOrFuncCallStat(l *lexer.Lexer) ast.Stat {
	context := *l //保护现场
	prefixExp := parsePrefixExp(l)
	if fc, ok := prefixExp.(*ast.FuncCallExp); ok { //to do 将冒号语法糖还原
		return fc
	}
	*l = context //返回现场
	return _parseAssignStat(l, prefixExp)
}

func _parseAssignStat(l *lexer.Lexer, prefixExp ast.Exp) *ast.AssignStat {
	//解析为varlist，区别_parseLocalVal中解析为identifier list，因为赋值语句都是使用已经定义好的变量所以可能会有前缀变量
	vars := parseVarList(l)
	l.AssertAndSkipToken(lexer.TOKEN_OP_ASSIGN) //skip '='
	exps := parseExpList(l)
	return &ast.AssignStat{
		LastLine: l.Line(),
		VarList:  vars,
		ExpList:  exps,
	}
}

func parseFuncDefStat(l *lexer.Lexer) *ast.OopFuncDefStat {
	l.AssertAndSkipToken(lexer.TOKEN_KW_FUNCTION) //skip 'function'
	funcNameExp, hasColon := _parseFuncName(l)
	funcDef := parseFuncDefExp(l).(*ast.FuncDefExp)
	if hasColon { //处理冒号语法糖
		// v:name(args) => v.name(self, args)
		funcDef.ParList = append([]string{"self"}, funcDef.ParList...)
	}
	return &ast.OopFuncDefStat{
		Name: funcNameExp,
		Body: funcDef,
	}
}

// 因为方法名中可能存在':'语法糖，所以需要注意
// funcname ::= Name {'.' Name} [':' Name]
// for example:
// local M = {}
// function M.MyFunc(self, a)
// -- body
// end
// -- 下面的函数调用等价
// M.MyFunc(M,'test')
// M:MyFunc('test') -- 冒号语法糖会默认将M作为第一个参数
func _parseFuncName(l *lexer.Lexer) (exp ast.Exp, hasColon bool) {
	l.AssertIdentifier()
	t := l.NextToken()
	exp = &ast.NameExp{
		Line: t.Line(),
		Name: t.Val(),
	}
	for l.CheckToken(lexer.TOKEN_SEP_DOT) {
		l.NextToken() //skip '.'
		t := l.AssertAndSkipToken(lexer.TOKEN_IDENTIFIER)
		exp = &ast.TableAccessExp{
			LastLine:  t.Line(),
			PrefixExp: exp,
			CurrentExp: &ast.StringExp{
				Line: t.Line(),
				Str:  t.Val(),
			},
		}
	}

	//判断是否有冒号语法糖
	if l.CheckToken(lexer.TOKEN_SEP_COLON) {
		l.NextToken() //skip ':'
		t := l.AssertAndSkipToken(lexer.TOKEN_IDENTIFIER)
		exp = &ast.TableAccessExp{
			LastLine:  t.Line(),
			PrefixExp: exp,
			CurrentExp: &ast.StringExp{
				Line: t.Line(),
				Str:  t.Val(),
			},
		}
		hasColon = true
	}

	return
}
