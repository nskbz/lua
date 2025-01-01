package number

import "math"

func AbsInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// 整除 向下取整 a/b
func IntegerDiv(a, b int64) int64 {
	if a%b == 0 || a*b > 0 {
		return a / b
	}
	return a/b - 1
}
func FloatDiv(a, b float64) float64 {
	return math.Floor(a / b)
}

// 取模 向下取整 a%b
func IntegerMod(a, b int64) int64   { return a - b*IntegerDiv(a, b) }
func FloatMod(a, b float64) float64 { return a - b*FloatDiv(a, b) }

func ShiftLeft(a, n int64) int64 {
	if n > 0 {
		return a << uint64(n)
	}
	return ShiftRight(a, -n)
}

func ShiftRight(a, n int64) int64 {
	if n > 0 {
		return int64(uint64(a) >> uint64(n))
	}
	return ShiftLeft(a, -n)
}

/*
类型转换
*/
func FloatToInteger(f float64) (int64, bool) {
	i := int64(f)
	return i, float64(i) == f
}
