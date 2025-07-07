package instruction

import "nskbz.cn/lua/api"

// R(A) := {} (size = B,C)
func newTable(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a += 1
	vm.CreateTable(Fb2int(b), Fb2int(c))
	vm.Replace(a)
}

// R(A) := R(B)[RK(C)]
func getTable(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1
	vm.GetRK(c)
	vm.GetTable(b)
	vm.Replace(a)
}

// R(A)[RK(B)] := RK(C)
// 将RK(K)作为值映射为表R(A)中的RK(B)键
func setTable(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a += 1
	vm.GetRK(b)
	vm.GetRK(c)
	vm.SetTable(a)
}

// SETLIST 批大小
const LFIELDS_PER_FLUSH = 50

// SETLIST 指令用于将栈上的一组连续值批量赋值给表的数组部分，通常在初始化表或批量更新数组元素时使用。底层还是利用的SETTABLE指令。
// SETLIST A, B, C
// A：表所在的寄存器索引。
// B：栈上连续值的数量。
// C：决定目标表数组的起始索引(非实际索引),准确应该是批数。
func setList(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a += 1

	//判断该命令是否使用拓展
	if c > 0 {
		c = c - 1
	} else {
		c = Instruction(vm.Fetch()).Ax()
	}

	bIsZero := b == 0
	if bIsZero {
		b = int(vm.ToInteger(0)) - a - 1
		vm.Pop(1)
	}

	//把寄存器列表里的val添加进table
	vm.CheckStack(2)
	for i := 1; i <= b; i++ {
		vm.PushInteger(int64(c*LFIELDS_PER_FLUSH + i)) //C经过计算后才是实际索引
		vm.PushValue(a + i)
		vm.SetTable(a)
	}

	//如果b==0还需将栈上的val添加进table
	if bIsZero {
		for i := 1; i <= vm.GetTop()-vm.RegisterCount(); i++ {
			vm.PushInteger(int64(c*LFIELDS_PER_FLUSH + b + i))
			vm.PushValue(vm.RegisterCount() + i)
			vm.SetTable(a)
		}

		vm.SetTop(vm.RegisterCount()) //栈清零
	}
}

/*
** converts an integer to a "floating point byte", represented as
** (eeeeexxx), where the real value is (1xxx) * 2^(eeeee - 1) if
** eeeee != 0 and (xxx) otherwise.
 */
func Int2fb(x int) int {
	e := 0 /* exponent */
	if x < 8 {
		return x
	}
	for x >= (8 << 4) { /* coarse steps */
		x = (x + 0xf) >> 4 /* x = ceil(x / 16) */
		e += 4
	}
	for x >= (8 << 1) { /* fine steps */
		x = (x + 1) >> 1 /* x = ceil(x / 2) */
		e++
	}
	return ((e + 1) << 3) | (x - 8)
}

/* converts back */
func Fb2int(x int) int {
	if x < 8 {
		return x
	} else {
		return ((x & 7) + 8) << uint((x>>3)-1)
	}
}
