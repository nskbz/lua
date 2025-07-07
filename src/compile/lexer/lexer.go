package lexer

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var reNewLine = regexp.MustCompile("\r\n|\n\r|\r|\n")

var reDecEscapeSeq = regexp.MustCompile(`^\\[0-9]{1,3}`)          //十进制ASCII码
var reHexEscapeSeq = regexp.MustCompile(`^\\x[0-9a-fA-F]{2}`)     //十六进制ASCII码
var reUnicodeEscapeSeq = regexp.MustCompile(`^\\u{[0-9a-fA-F]+}`) //unicode码
var reNumber = regexp.MustCompile(`^0[xX][0-9a-fA-F]*(\.[0-9a-fA-F]*)?([pP][+\-]?[0-9]+)?|^[0-9]*(\.[0-9]*)?([eE][+\-]?[0-9]+)?`)
var reIdentifier = regexp.MustCompile(`^[_\d\w]+`)
var reShortStr = regexp.MustCompile(`(?s)(^'(\\\\|\\'|\\\n|\\z\s*|[^'\n])*')|(^"(\\\\|\\"|\\\n|\\z\s*|[^"\n])*")`)

type Token struct {
	kind  int //TOKEN类型
	value string

	line int //TOKEN所在行号
	i    int //TOKEN索引
}

func (t *Token) Kind() int   { return t.kind }
func (t *Token) Line() int   { return t.line }
func (t *Token) Val() string { return t.value }

type Lexer struct {
	sourceName string //源文件名
	data       []byte //源代码
	i          int
	line       int //当前行号

	//下一token缓存
	cache *Token
}

func NewLexer(chunk []byte, name string) *Lexer {
	return &Lexer{
		sourceName: name,
		data:       chunk,
		i:          0,
		line:       1,
		cache:      nil,
	}
}

func (l *Lexer) Line() int { return l.line }

func (l *Lexer) error(format string, err ...interface{}) {
	info := fmt.Sprintf(format, err...)
	errInfo := fmt.Sprintf("file:[%s] line:[%d] err[%s]", l.sourceName, l.line, info)
	panic(errInfo)
}

func (l *Lexer) char() byte {
	return l.data[l.i]
}

func (l *Lexer) empty() bool {
	return l.i >= len(l.data)
}

func (l *Lexer) string() string {
	return string(l.data[l.i:])
}

func (l *Lexer) next(n int) {
	l.i += n
}

// 判断data是否以s开头
func (l *Lexer) test(s string) bool {
	length := len(s)
	return string(l.data[l.i:l.i+length]) == s
}

// 判断是否是空白字符
func (l *Lexer) isWhiteSpace() bool {
	return checkWhiteSpace(l.char())
}

func (l *Lexer) isNewLine() bool {
	c := l.char()
	return c == '\r' || c == '\n'
}

func (l *Lexer) ScanLongString() string {
	return l.scanLongString()
}

func (l *Lexer) scanLongString() string {
	format := false //记录长字符串格式是否正确
	sb := strings.Builder{}
	if !l.test("[") {
		l.error("prefix error")
	}
	l.next(1)
	sb.WriteByte(']')

	//匹配若干'='
	for !l.empty() && l.test("=") {
		l.next(1)
		sb.WriteByte('=')
	}
	if l.test("[") {
		l.next(1)
		sb.WriteByte(']')
		separator := sb.String() //生成结束标记

		sb.Reset()
		for !l.empty() {
			if l.test(separator) { //如果后续存在结束标记，则说明长字符串格式正确
				format = true
				l.next(len(separator))
				break
			}
			sb.WriteByte(l.char())
			l.next(1)
		}
	}

	if !format {
		l.error("suffix error")
	}

	//处理换行序列
	str := reNewLine.ReplaceAllString(sb.String(), "\n")
	l.line += strings.Count(str, "\n")
	if len(str) > 0 && str[0] == '\n' {
		str = str[1:]
	}
	return str
}

// 短字符串可以包括转义序列
func (l *Lexer) scanShortString() string {
	if found := reShortStr.FindString(l.string()); found != "" {
		l.next(len(found))
		found = found[1 : len(found)-1]
		if strings.Contains(found, "\\") {
			l.line += len(reNewLine.FindAllString(found, -1))
			found = l.escape(found)
		}
		return found
	}
	l.error("scan short string error")
	return ""
}

// 转义处理
func (l *Lexer) escape(str string) string {
	buf := bytes.Buffer{}

	for len(str) > 0 {
		//非转义
		if str[0] != '\\' {
			buf.WriteByte(str[0])
			str = str[1:]
			continue //非转义直接跳过
		}
		//转义
		if len(str) <= 1 {
			l.error("error format string")
		}
		switch str[1] { //str[0] == '\'
		case 'a':
			buf.WriteByte('\a')
			str = str[2:]
		case 'b':
			buf.WriteByte('\b')
			str = str[2:]
		case 't':
			buf.WriteByte('\t')
			str = str[2:]
		case 'n':
			buf.WriteByte('\n')
			str = str[2:]
		case 'v':
			buf.WriteByte('\v')
			str = str[2:]
		case 'f':
			buf.WriteByte('\f')
			str = str[2:]
		case 'r':
			buf.WriteByte('\r')
			str = str[2:]
		case '"':
			buf.WriteByte('"')
			str = str[2:]
		case '\'':
			buf.WriteByte('\'')
			str = str[2:]
		case '\\':
			buf.WriteByte('\\')
			str = str[2:]
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9': // \nnn
			if found := reDecEscapeSeq.FindString(str); found != "" {
				if i, err := strconv.ParseInt(found[1:], 10, 32); err == nil && i <= 0xFF {
					buf.WriteByte(byte(i))
					str = str[len(found):]
				} else {
					l.error("[%s] decimal parse error , %s ,%d", found, err, i)
				}
			}
		case 'x': // \xFF
			if found := reHexEscapeSeq.FindString(str); found != "" {
				if i, err := strconv.ParseInt(found[2:], 16, 32); err == nil {
					buf.WriteByte(byte(i))
					str = str[len(found):]
				} else {
					l.error("[%s] hex parse error , %s", found, err)
				}
			}
		case 'u': // \u{FFF}
			if found := reUnicodeEscapeSeq.FindString(str); found != "" {
				if i, err := strconv.ParseInt(found[3:len(found)-1], 16, 32); err == nil && i <= 0x10FFFF {
					buf.WriteRune(rune(i))
					str = str[len(found):]
				} else {
					l.error("[%s] UTF-8 parse error , %s , %d", found, err, i)
				}
			}
		case 'z':
			i := 2
			for i < len(str) && checkWhiteSpace(str[i]) {
				i++
			}
			str = str[i:]
		}
	}

	return buf.String()
}

func (l *Lexer) skipComment() {
	//长注释
	if l.test("[") {
		l.scanLongString()
		return
	}

	//短注释
	for !l.empty() && !l.isNewLine() {
		l.next(1)
	}
}

// 跳过空白换行和注释
func (l *Lexer) skipWhiteSpaces() {
	for !l.empty() {
		if l.test("--") { //跳过注释
			l.next(2)
			l.skipComment()
		} else if l.test("\r\n") || l.test("\n\r") { //跳过windows换行
			l.next(2)
			l.line += 1
		} else if l.isNewLine() { //跳过linux换行
			l.next(1)
			l.line += 1
		} else if l.isWhiteSpace() { //其他符不需要增加行号
			l.next(1)
		} else {
			break
		}
	}
}

func (l *Lexer) NextToken() Token {
	//如果缓存着下一token的信息直接返回
	if l.cache != nil {
		l.line = l.cache.line
		l.i = l.cache.i
		c := l.cache
		l.cache = nil
		return *c
	}

	l.skipWhiteSpaces() //跳过空白和注释
	if l.empty() {
		return Token{TOKEN_EOF, "EOF", l.line, l.i}
	}

	errMsg := fmt.Sprintf("Can't resolve [%c]", l.char())
	switch l.char() {
	case ';':
		l.next(1)
		return Token{TOKEN_SEP_SEMI, ";", l.line, l.i}
	case ',':
		l.next(1)
		return Token{TOKEN_SEP_COMMA, ",", l.line, l.i}
	case '(':
		l.next(1)
		return Token{TOKEN_SEP_LPAREN, "(", l.line, l.i}
	case ')':
		l.next(1)
		return Token{TOKEN_SEP_RPAREN, ")", l.line, l.i}
	case '[':
		//存在注释或长字符串的可能
		if l.test("[[") || l.test("[=") {
			t := l.scanLongString()
			return Token{TOKEN_STRING, t, l.line, l.i}
		}
		l.next(1)
		return Token{TOKEN_SEP_LBRACK, "[", l.line, l.i}
	case ']':
		l.next(1)
		return Token{TOKEN_SEP_RBRACK, "]", l.line, l.i}
	case '{':
		l.next(1)
		return Token{TOKEN_SEP_LCURLY, "{", l.line, l.i}
	case '}':
		l.next(1)
		return Token{TOKEN_SEP_RCURLY, "}", l.line, l.i}
	case '+':
		l.next(1)
		return Token{TOKEN_OP_ADD, "+", l.line, l.i}
	case '-':
		l.next(1)
		return Token{TOKEN_OP_MINUS, "-", l.line, l.i}
	case '*':
		l.next(1)
		return Token{TOKEN_OP_MUL, "*", l.line, l.i}
	case '^':
		l.next(1)
		return Token{TOKEN_OP_POW, "^", l.line, l.i}
	case '%':
		l.next(1)
		return Token{TOKEN_OP_MOD, "%", l.line, l.i}
	case '&':
		l.next(1)
		return Token{TOKEN_OP_BAND, "&", l.line, l.i}
	case '|':
		l.next(1)
		return Token{TOKEN_OP_BOR, "|", l.line, l.i}
	case '#':
		l.next(1)
		return Token{TOKEN_OP_LEN, "#", l.line, l.i}
	case ':':
		if l.test("::") {
			l.next(2)
			return Token{TOKEN_SEP_LABEL, "::", l.line, l.i}
		}
		l.next(1)
		return Token{TOKEN_SEP_COLON, ":", l.line, l.i}
	case '/':
		if l.test("//") {
			l.next(2)
			return Token{TOKEN_OP_IDIV, "//", l.line, l.i}
		}
		l.next(1)
		return Token{TOKEN_OP_DIV, "/", l.line, l.i}
	case '~':
		if l.test("~=") {
			l.next(2)
			return Token{TOKEN_OP_NE, "~=", l.line, l.i}
		}
		l.next(1)
		return Token{TOKEN_OP_WAVE, "~", l.line, l.i} //需区分是not还是异或(xor)
	case '=':
		if l.test("==") {
			l.next(2)
			return Token{TOKEN_OP_EQ, "==", l.line, l.i}
		}
		l.next(1)
		return Token{TOKEN_OP_ASSIGN, "=", l.line, l.i} //赋值
	case '<':
		if l.test("<<") {
			l.next(2)
			return Token{TOKEN_OP_SHL, "<<", l.line, l.i}
		} else if l.test("<=") {
			l.next(2)
			return Token{TOKEN_OP_LE, "<=", l.line, l.i}
		}
		l.next(1)
		return Token{TOKEN_OP_LT, "<", l.line, l.i}
	case '>':
		if l.test(">>") {
			l.next(2)
			return Token{TOKEN_OP_SHR, ">>", l.line, l.i}
		} else if l.test(">=") {
			l.next(2)
			return Token{TOKEN_OP_GE, ">=", l.line, l.i}
		}
		l.next(1)
		return Token{TOKEN_OP_GT, ">", l.line, l.i}
	case '.':
		if l.test("...") {
			l.next(3)
			return Token{TOKEN_VARARG, "...", l.line, l.i} //可变参数
		} else if l.test("..") {
			l.next(2)
			return Token{TOKEN_OP_CONCAT, "..", l.line, l.i} //连接
		}
		l.next(1)
		return Token{TOKEN_SEP_DOT, ".", l.line, l.i}
	case '\'', '"': //短字符串开始标记
		t := l.scanShortString()
		return Token{TOKEN_STRING, t, l.line, l.i}
	}

	//数字字面量
	if checkNumber(l.char()) {
		if found := reNumber.FindString(l.string()); found != "" {
			l.next(len(found))
			return Token{TOKEN_NUMBER, found, l.line, l.i}
		}
		errMsg = fmt.Sprintf("Can't find number.near by [%c]", l.char())
	}

	//关键字或标识符
	if l.char() == '_' || checkLetter(l.char()) {
		if found := reIdentifier.FindString(l.string()); found != "" {
			l.next(len(found))
			if key, ok := keywords[found]; ok {
				return Token{key, found, l.line, l.i} //keyword
			}
			return Token{TOKEN_IDENTIFIER, found, l.line, l.i} //identifier
		}
		errMsg = fmt.Sprintf("Can't find keyword or identifier.near by [%c]", l.char())
	}

	return Token{0, errMsg, 0, 0}
}

// 查看下一个TOKEN但不改变词法分析器的状态
func (l *Lexer) LookToken() Token {
	if l.cache != nil {
		return *l.cache
	}
	preLine := l.line
	preI := l.i
	token := l.NextToken()
	l.cache = &token
	l.line = preLine
	l.i = preI
	return token
}

// 断言下一个TOKEN为kind类型
// 不是则终止
func (l *Lexer) AssertToken(kind int) *Token {
	t := l.LookToken()
	if t.kind != kind {
		panic(fmt.Sprintf("real type[%d] != excepted type[%d]", t.kind, kind))
	}
	return &t
}

func (l *Lexer) CheckToken(kind int) bool {
	t := l.LookToken()
	return t.Kind() == kind
}

func (l *Lexer) AssertIdentifier() Token {
	return *l.AssertToken(TOKEN_IDENTIFIER)
}

func (l *Lexer) AssertAndSkipToken(kind int) *Token {
	t := l.AssertToken(kind)
	l.NextToken()
	return t
}

func checkWhiteSpace(c byte) bool {
	switch c {
	case '\n', '\r', '\t', '\v', '\f', ' ':
		return true
	}
	return false
}

func checkNumber(c byte) bool {
	return c >= '0' && c <= '9'
}

func checkLetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}
