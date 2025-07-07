package instruction

import "nskbz.cn/lua/api"

// R(A) := R(B)
func move(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	vm.Copy(b+1, a+1)
}

// pc+=sBx; if (A) close all upvalues >= R(A - 1)
func jump(i Instruction, vm api.LuaVM) {
	a, sbx := i.AsBx()
	vm.AddPC(sbx)
	if a != 0 {
		// close all upvalues >= R(A - 1)：关闭所有从寄存器R(A-1)开始的upvalue。这是因为：
		// 当离开一个作用域时，该作用域内的局部变量将不再可用
		// 这些变量可能已经被嵌套函数捕获为upvalue
		// 需要关闭这些upvalue以确保它们被正确释放
		// forexample:
		// do
		//  lcal x = 10  -- 这个变量可能被嵌套函数捕获
		//  -- ...一些代码...
		//  goto label    -- 这里会产生JMP指令
		//  -- x的作用域在这里结束
		// end
		// ::label::
		// 在这种情况下，JMP跳出do-end块时，需要关闭x的upvalue（如果它被任何闭包捕获）。
		vm.CloseUpvalues(a)
	}
}

func valLen(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	vm.Len(b + 1)
	vm.Replace(a + 1)
}

// R(A) := R(B).. ... ..R(C)
func concat(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1
	c += 1
	n := c - b + 1
	vm.CheckStack(n)
	for i := b; i <= c; i++ {
		vm.PushValue(i)
	}
	vm.Concat(n)
	vm.Replace(a)
}
