package number

import (
	"fmt"
	"strconv"
)

func ParseInteger(str string) (int64, bool) {
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		fmt.Printf("\n\n%s\n\n", err.Error())
	}
	return i, err == nil
}

func ParseFloat(str string) (float64, bool) {
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		fmt.Printf("\n\n%s\n\n", err.Error())
	}
	return f, err == nil
}
