package instruction

import "nskbz.cn/lua/api"

// R(A) := closure(KPROTO[Bx])
func closure(i Instruction, vm api.LuaVM) {
	a, bx := i.ABx()
	a += 1

	vm.LoadProto(bx) // 捕获外部变量装载到closure
	vm.Replace(a)
}

// 将func及其入参压入栈，并返回入参个数
func _pushFuncAndArgs(a, b int, vm api.LuaVM) int {
	nArgs := 0
	if b > 0 {
		vm.CheckStack(b)
		for i := a; i < a+b; i++ {
			vm.PushValue(i) // 1个func，b-1个参数
		}
		nArgs = b - 1
	} else if b == 0 {
		//b==0表示压入的参数包括某个被调函数的全部返回值 例:f(1,2,g())
		x := int(vm.ToInteger(0))
		vm.Pop(1)

		//压入除去被调函数的其他参数
		vm.CheckStack(x - a)
		for i := a; i < x; i++ {
			vm.PushValue(i)
		}
		vm.Rotate(vm.RegisterCount()+1, x-a) //调整参数的顺序=>[其他参数，被调函数返回值]
		nArgs = vm.GetTop() - vm.RegisterCount() - 1
	}
	return nArgs
}

// c>1返回值数量为c-1
// c==1返回值数量为0
// c<=0返回所有返回值 例:f(1,2,g()),g函数调用需返回所有返回值于f函数,所以调用g函数的CALL指令c为0
func _popResults(a, c int, vm api.LuaVM) {
	if c == 1 {
	} else if c > 1 {
		for i := a + c - 2; i >= a; i-- {
			vm.Replace(i) //从尾部开始替换，栈顶存放的是最后的返回值
		}
	} else if c == 0 {
		vm.CheckStack(1)
		vm.PushInteger(int64(a)) //压入
	}
}

// R(A), ... ,R(A+C-2) := R(A)(R(A+1), ... ,R(A+B-1))
// c>1返回值数量为c-1
// C==1则不会返回任何返回值
// c<=0返回所有返回值
func call(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a += 1

	//将参数压入
	nArgs := _pushFuncAndArgs(a, b, vm)

	//执行函数
	vm.Call(nArgs, c-1)

	//设置返回值
	_popResults(a, c, vm)
}

func tailcall(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	nArgs := _pushFuncAndArgs(a, b, vm)
	vm.Call(nArgs, -1)
	_popResults(a, 0, vm)
}

// return R(A),...,R(A+B-2)
func luaReturn(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	a += 1

	//b==1不需要返回值
	//b>1返回b-1个返回值
	//b==0表示返回被调函数所有返回值 例:return f()
	if b == 1 {
	} else if b > 1 {
		vm.CheckStack(b - 1)
		for i := a; i <= a+b-2; i++ {
			vm.PushValue(i)
		}
	} else if b == 0 {
		x := int(vm.ToInteger(0)) //获取除被调函数返回值的起始寄存器索引
		vm.Pop(1)

		//压入除去被调函数的其他参数
		vm.CheckStack(x - a)
		for i := a; i < x; i++ {
			vm.PushValue(i)
		}
		vm.Rotate(vm.RegisterCount()+1, x-a)
	}
}

// R(A), R(A+1), ..., R(A+B-2) = vararg
func vararg(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	if b < 0 {
		panic("VARARG error param b <0")
	}
	if b > 1 {
		vm.LoadVarargs(b - 1)
		for i := a + b - 2; i >= a; i-- {
			vm.Replace(i)
		}
	} else if b == 0 {
		vm.LoadVarargs(-1) //全部load
		vm.CheckStack(1)
		vm.PushInteger(int64(a)) //压入起始索引号
	}
}

// R(A+1) := R(B); R(A) := R(B)[RK(C)]
// SELF指令主要用于方法调用(method call)的准备工作，特别是在面向对象编程OOP中用于传递隐式的self参数
// 指令格式：SELF A, B, C
// A：目标寄存器，存储方法函数。
// B：接收者（self）所在的寄存器。
// C：方法名的常量表索引（通常是字符串，如 "method"）
// 当调用形如 obj:method(args)
// R(A+1) = R(B)                 -- 将接收者（self）存入下一个寄存器
// R(A) = R(B)[RK(C)]  			-- 从 R(B) 中获取名为 K(C) 的方法
func self(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1

	vm.PushValue(b)
	vm.Replace(a + 1)
	vm.GetRK(c)
	vm.GetTable(b)
	vm.Replace(a)
}
