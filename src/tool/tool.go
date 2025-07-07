package tool

import (
	"fmt"
	"strings"

	"nskbz.cn/lua/api"
)

var DebugFlag bool = false

func Debug(format string, args ...interface{}) {
	if DebugFlag {
		if len(args) == 0 {
			fmt.Println(format)
			return
		}
		fmt.Printf(format, args...)
	}
}

func Fatal(s api.LuaVM, msg string) {
	PrintStack(s)
	panic(msg)
}

func PrintStack(s api.LuaVM) {
	//fmt.Printf("register count=%d\n", s.RegisterCount())
	for i := 1; i <= s.GetTop(); i++ {
		tp := s.Type(i)
		switch tp {
		case api.LUAVALUE_BOOLEAN:
			fmt.Printf("[%t]", s.ToBoolean(i))
		case api.LUAVALUE_NUMBER:
			fmt.Printf("[%g]", s.ToFloat(i))
		case api.LUAVALUE_STRING:
			fmt.Printf("[%q]", s.ToString(i))
		default:
			fmt.Printf("[%s]", s.TypeName(tp))
		}
	}
	fmt.Println()
}

func ReplaceTabToSpace(str string, tabSpan int) string {
	sb := strings.Builder{}
	idx := 0
	for _, s := range str {
		if s == '\t' {
			nSpace := tabSpan - (idx % tabSpan)
			sb.WriteString(strings.Repeat(" ", nSpace))
			idx += nSpace
		} else {
			sb.WriteRune(s)
		}
		idx++
		if s == '\n' {
			idx = 0
		}
	}
	return sb.String()
}
