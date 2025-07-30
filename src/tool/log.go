package tool

import (
	"fmt"
	"strings"

	"nskbz.cn/lua/api"
)

var LogLevel int = LOG_DEFAULT

const (
	LOG_TRACE = -1 //只用于看指令执行
	LOG_ERROR = iota
	LOG_WARNING
	LOG_DEBUG
	LOG_DEFAULT
)

func Trace(format string, args ...interface{}) {
	format = strings.Trim(format, "\n")
	if LogLevel == LOG_TRACE {
		format = "[TRACE]: " + format + "\n"
		fmt.Printf(format, args...)
	}
}

func Debug(format string, args ...interface{}) {
	format = strings.Trim(format, "\n")
	if LogLevel <= LOG_DEBUG {
		format = "[DEBUG]: " + format + "\n"
		fmt.Printf(format, args...)
	}
}

func Warning(format string, args ...interface{}) {
	format = strings.Trim(format, "\n")
	if LogLevel <= LOG_DEBUG {
		format = "[WARNING]: " + format + "\n"
		fmt.Printf(format, args...)
	}
}

func Error(format string, args ...interface{}) {
	format = strings.Trim(format, "\n")
	if LogLevel <= LOG_DEBUG {
		format = "[ERROR]: " + format + "\n"
		fmt.Printf(format, args...)
	}
}

func Fatal(s api.LuaVM, msg string) {
	PrintStack(s)
	panic(msg)
}
