package state

import (
	"math"

	"nskbz.cn/lua/api"
	"nskbz.cn/lua/number"
)

type intFunc func(int64, int64) int64
type floatFunc func(float64, float64) float64

var (
	iadd = func(a, b int64) int64 { return a + b }
	fadd = func(a, b float64) float64 { return a + b }
	isub = func(a, b int64) int64 { return a - b }
	fsub = func(a, b float64) float64 { return a - b }
	imul = func(a, b int64) int64 { return a * b }
	fmul = func(a, b float64) float64 { return a * b }
	imod = number.IntegerMod
	fmod = number.FloatMod
	pow  = math.Pow
	div  = func(a, b float64) float64 {
		if b == 0 {
			panic("The dividend cannot be zero")
		}
		return a / b
	}
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
	m string    //元方法
	i intFunc   //整形参数方法
	f floatFunc //浮点型参数方法
}

func doUnitaryArith(val luaValue, op operation, ls *luaState) (luaValue, bool) {
	if op.f == nil { //位运算 not
		if x, ok := convertToInteger(val); ok {
			return op.i(x, 0), true
		}
	} else { //数字运算 opposite
		if op.i != nil {
			if x, ok := val.(int64); ok {
				return op.i(x, 0), true
			}
		}
		if x, ok := convertToFloat(val); ok {
			return op.f(x, 0), true
		}
	}

	//上面转换不行则执行该类型的元方法
	if c := getMetaClosure(ls, op.m, val); c != nil {
		return callMetaClosure(ls, c, 1, val)[0], true
	}
	return nil, false
}

func doDualArith(a, b luaValue, op operation, ls *luaState) (luaValue, bool) {
	if op.f == nil { //位运算
		if x, ok := convertToInteger(a); ok {
			if y, ok := convertToInteger(b); ok {
				return op.i(x, y), true
			}
		}
	} else { //数字运算
		if op.i != nil { //add,sub,mul,mod,idiv,opposite
			if x, ok := a.(int64); ok {
				if y, ok := b.(int64); ok {
					return op.i(x, y), true
				}
			}
		}
		if x, ok := convertToFloat(a); ok {
			if y, ok := convertToFloat(b); ok {
				return op.f(x, y), true
			}
		}
	}

	//上面转换不行则执行该类型的元方法
	if c := getMetaClosure(ls, op.m, a, b); c != nil {
		return callMetaClosure(ls, c, 1, a, b)[0], true
	}
	return nil, false
}

var arith_operation = map[api.ArithOp]operation{
	api.ArithOp_ADD:      {META_ADD, iadd, fadd},
	api.ArithOp_SUB:      {META_SUB, isub, fsub},
	api.ArithOp_MUL:      {META_MUL, imul, fmul},
	api.ArithOp_MOD:      {META_MOD, imod, fmod},
	api.ArithOp_POW:      {META_POW, nil, pow},
	api.ArithOp_DIV:      {META_DIV, nil, div},
	api.ArithOp_IDIV:     {META_IDIV, iidiv, fidiv},
	api.ArithOp_AND:      {META_AND, and, nil},
	api.ArithOp_OR:       {META_OR, or, nil},
	api.ArithOp_XOR:      {META_XOR, xor, nil},
	api.ArithOp_SHL:      {META_SHL, shl, nil},
	api.ArithOp_SHR:      {META_SHR, shr, nil},
	api.ArithOp_OPPOSITE: {META_OPPOSITE, iopposite, fopposite},
	api.ArithOp_NOT:      {META_NOT, not, nil},
}
