package parser

import (
	"nskbz.cn/lua/compile/ast"
	"nskbz.cn/lua/compile/lexer"
)

func parseBlock(l *lexer.Lexer) *ast.Block {
	stats := parseStats(l)
	retExps := parseRetExps(l)
	return &ast.Block{
		Stats:    stats,
		RetExps:  retExps,
		LastLine: l.Line(),
	}
}

func parseStats(l *lexer.Lexer) []ast.Stat {
	stats := []ast.Stat{}
	nt := l.LookToken()
	for !_isReturnOrBlockEnd(nt.Kind()) {
		stat := parseStat(l)
		if _, ok := stat.(*ast.EmptyStat); !ok {
			stats = append(stats, stat)
		}
		nt = l.LookToken()
	}
	return stats
}

func _isReturnOrBlockEnd(tokenKind int) bool {
	switch tokenKind {
	case lexer.TOKEN_EOF, lexer.TOKEN_KW_RETURN, lexer.TOKEN_KW_END,
		lexer.TOKEN_KW_UNTIL, lexer.TOKEN_KW_ELSE, lexer.TOKEN_KW_ELSEIF:
		return true
	}
	return false
}

// retstat ::= return [explist] [‘;’]
func parseRetExps(l *lexer.Lexer) []ast.Exp {
	nt := l.LookToken()
	if nt.Kind() != lexer.TOKEN_KW_RETURN {
		return nil
	}

	l.NextToken() //skip 'return'
	nt = l.LookToken()
	switch nt.Kind() {
	case lexer.TOKEN_EOF, lexer.TOKEN_KW_END,
		lexer.TOKEN_KW_UNTIL, lexer.TOKEN_KW_ELSE, lexer.TOKEN_KW_ELSEIF:
		return []ast.Exp{}
	case lexer.TOKEN_SEP_SEMI:
		l.NextToken()
		return []ast.Exp{}
	}
	exps := parseExpList(l)
	if nt := l.LookToken(); nt.Kind() == lexer.TOKEN_SEP_SEMI {
		l.NextToken()
	}
	return exps
}
