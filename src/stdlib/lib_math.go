package stdlib

import (
	"math"

	"nskbz.cn/lua/api"
)

var mathFuncs map[string]api.GoFunc = map[string]api.GoFunc{
	"min":   mathMin,
	"max":   mathMax,
	"abs":   mathAbs,
	"ceil":  mathCeil,
	"floor": mathFloor,
	"pow":   mathPow,
	"sqrt":  mathSqrt,
}

// 被RequireF调用,带唯一参数modname,具体于该方法的函数栈stack[1]=="math"
// 用于加载math库的方法,最后应当生成一个table于栈顶返回
func OpenMathLib(vm api.LuaVM) int {
	vm.NewLib(mathFuncs)
	return 1
}

// math.min(x1, x2, ...)
// 参数：可以传入任意数量的数值（整数或浮点数）。
// 返回值：返回所有参数中最小的数值。
func mathMin(vm api.LuaVM) int {
	nArgs := vm.GetTop()
	if nArgs == 0 {
		return vm.Error2("there is no args!")
	}
	var min float64 = math.MaxFloat64

	for i := 1; i <= nArgs; i++ {
		if f, ok := vm.ToFloatX(i); !ok {
			return vm.Error2("%s can't convert to float64", vm.Type(i).String())
		} else if f < min {
			min = f
		}
	}
	if min != math.MaxFloat64 {
		vm.PushFloat(min)
		return 1
	}
	return vm.Error2("do min error")
}

// math.max(x1, x2, ...)
// 参数：可以传入任意数量的数值（整数或浮点数）。
// 返回值：返回所有参数中最大的数值。
func mathMax(vm api.LuaVM) int {
	nArgs := vm.GetTop()
	if nArgs == 0 {
		return vm.Error2("there is no args!")
	}
	var max float64 = -math.MaxFloat64

	for i := 1; i <= nArgs; i++ {
		if f, ok := vm.ToFloatX(i); !ok {
			return vm.Error2("%s can't convert to float64", vm.Type(i).String())
		} else if f > max {
			max = f
		}
	}
	if max != -math.MaxFloat64 {
		vm.PushFloat(max)
		return 1
	}
	return vm.Error2("do min error")
}

// math.abs(x)
// 返回 x 的绝对值	math.abs(-3.5) → 3.5
func mathAbs(vm api.LuaVM) int {
	vm.CheckAny(1)
	p := vm.ToPointer(1)
	switch v := p.(type) {
	case int64:
		if v < 0 {
			v = -v
		}
		vm.PushInteger(v)
		return 1
	case float64:
		if v < 0 {
			v = -v
		}
		vm.PushFloat(v)
		return 1
	}
	return vm.Error2("expected number ,not support %s", vm.Type(1).String())
}

// math.ceil(x)
// 向上取整	math.ceil(3.2) → 4
func mathCeil(vm api.LuaVM) int {
	vm.CheckAny(1)
	if f, ok := vm.ToFloatX(1); ok {
		f = math.Ceil(f)
		vm.PushFloat(f)
		return 1
	}
	return vm.Error2("ceil fail!")
}

// math.floor(x)
// 向下取整	math.floor(3.7) → 3
func mathFloor(vm api.LuaVM) int {
	vm.CheckAny(1)
	if f, ok := vm.ToFloatX(1); ok {
		f = math.Floor(f)
		vm.PushFloat(f)
		return 1
	}
	return vm.Error2("ceil fail!")
}

// math.pow(x, y)
// 计算 x 的 y 次方	math.pow(2, 3) → 8
func mathPow(vm api.LuaVM) int {
	vm.CheckAny(1)
	vm.CheckAny(2)
	if x, ok := vm.ToFloatX(1); ok {
		if y, ok := vm.ToFloatX(2); ok {
			vm.PushFloat(math.Pow(x, y))
			return 1
		}
	}
	return vm.Error2("get pow error!")
}

// math.sqrt(x)
// 计算 x 的平方根	math.sqrt(9) → 3
func mathSqrt(vm api.LuaVM) int {
	vm.CheckAny(1)
	if x, ok := vm.ToFloatX(1); ok {
		vm.PushFloat(math.Sqrt(x))
		return 1
	}
	return vm.Error2("get sqrt error!")
}
