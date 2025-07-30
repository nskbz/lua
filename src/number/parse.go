package number

import (
	"strconv"

	"nskbz.cn/lua/tool"
)

func ParseInteger(str string) (int64, bool) {
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		tool.Debug("%s", err.Error())
	}
	return i, err == nil
}

func ParseFloat(str string) (float64, bool) {
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		tool.Debug("%s", err.Error())
	}
	return f, err == nil
}
