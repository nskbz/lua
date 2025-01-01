package state

import (
	"math"

	"nskbz.cn/lua/api"
	"nskbz.cn/lua/number"
)

type intFunc func(int64, int64) int64
type floatFunc func(float64, float64) float64

var (
	iadd      = func(a, b int64) int64 { return a + b }
	fadd      = func(a, b float64) float64 { return a + b }
	isub      = func(a, b int64) int64 { return a - b }
	fsub      = func(a, b float64) float64 { return a - b }
	imul      = func(a, b int64) int64 { return a * b }
	fmul      = func(a, b float64) float64 { return a * b }
	imod      = number.IntegerMod
	fmod      = number.FloatMod
	pow       = math.Pow
	div       = func(a, b float64) float64 { return a / b }
	iidiv     = number.IntegerDiv
	fidiv     = number.FloatDiv
	and       = func(a, b int64) int64 { return a & b }
	or        = func(a, b int64) int64 { return a | b }
	xor       = func(a, b int64) int64 { return a ^ b }
	shl       = number.ShiftLeft
	shr       = number.ShiftRight
	iopposite = func(a, _ int64) int64 { return -a }
	fopposite = func(a, _ float64) float64 { return -a }
	not       = func(a, _ int64) int64 { return ^a }
)

type operation struct {
	i intFunc
	f floatFunc
}

func doUnitaryArith(val luaValue, op operation) luaValue {
	if op.f == nil { //位运算 not
		if x, ok := convertToInteger(val); ok {
			return op.i(x, 0)
		}
	} else { //数字运算 opposite
		if op.i != nil {
			if x, ok := val.(int64); ok {
				return op.i(x, 0)
			}
		}
		if x, ok := convertToFloat(val); ok {
			return op.f(x, 0)
		}
	}
	return nil
}

func doDualArith(a, b luaValue, op operation) luaValue {
	if op.f == nil { //位运算
		if x, ok := convertToInteger(a); ok {
			if y, ok := convertToInteger(b); ok {
				return op.i(x, y)
			}
		}
	} else { //数字运算
		if op.i != nil { //add,sub,mul,mod,idiv,opposite
			if x, ok := a.(int64); ok {
				if y, ok := b.(int64); ok {
					return op.i(x, y)
				}
			}
		}
		if x, ok := convertToFloat(a); ok {
			if y, ok := convertToFloat(b); ok {
				return op.f(x, y)
			}
		}
	}
	return nil
}

var arith_operation = map[api.ArithOp]operation{
	api.ArithOp_ADD:      {iadd, fadd},
	api.ArithOp_SUB:      {isub, fsub},
	api.ArithOp_MUL:      {imul, fmul},
	api.ArithOp_MOD:      {imod, fmod},
	api.ArithOp_POW:      {nil, pow},
	api.ArithOp_DIV:      {nil, div},
	api.ArithOp_IDIV:     {iidiv, fidiv},
	api.ArithOp_AND:      {and, nil},
	api.ArithOp_OR:       {or, nil},
	api.ArithOp_XOR:      {xor, nil},
	api.ArithOp_SHL:      {shl, nil},
	api.ArithOp_SHR:      {shr, nil},
	api.ArithOp_OPPOSITE: {iopposite, fopposite},
	api.ArithOp_NOT:      {not, nil},
}
