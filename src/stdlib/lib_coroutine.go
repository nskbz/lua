package stdlib

import "nskbz.cn/lua/api"

var coroutineFuncs map[string]api.GoFunc = map[string]api.GoFunc{
	"create":      coroutineCreate,
	"resume":      coroutineResume,
	"yield":       coroutineYield,
	"status":      coroutineStatus,
	"running":     coroutineRunning,
	"isyieldable": coroutineIsYieldable,
	"wrap":        coroutineWrap,
}

func OpenCoroutineLib(vm api.LuaVM) int {
	vm.NewLib(coroutineFuncs)
	return 1
}

// coroutine.create(func)
// 接收一个函数作为参数，并返回一个协程对象（类型为 thread）
// 之后可以通过 coroutine.resume 启动或恢复该协程的执行
func coroutineCreate(vm api.LuaVM) int {
	vm.CheckType(1, api.LUAVALUE_FUNCTION)
	vm.NewCoroutine()
	vm.PushValue(1)                 //func_copy
	vm.XMove(vm.ToCoroutine(-1), 1) //将func移动到new coroutine的栈上[index=1]
	return 1
}

// success, value1, value2, ... = coroutine.resume(co, arg1, arg2, ...)
// 参数：
// co: 要恢复或启动的协程对象，由coroutine.create创建
// arg1, arg2, ...: 可选参数，作为协程函数的参数(首次resume)或yield的返回值
// 返回值：
// success: 布尔值，表示协程是否成功执行
// value1, value2, ...: 协程通过yield返回的值或协程函数的返回值
func coroutineResume(vm api.LuaVM) int {
	vm.CheckType(1, api.LUAVALUE_COROUTINE)
	co := vm.ToCoroutine(1)
	nArgs := vm.GetTop() - 1
	co.CheckStack(nArgs) //防止即将执行的coroutine装不下所有参数

	//已经死亡的coroutine不能再被恢复
	if co.Status() >= api.LUA_OK && co.Status() <= api.LUA_DEAD {
		vm.PushBoolean(false)
		vm.PushString("can't resume a dead coroutine")
		return 2
	}

	vm.XMove(co, nArgs) //转移参数到子协程
	//当正常返回或子协程调用yeild函数从而使父协程resume返回(以SUSPENDED返回)
	if status := vm.Resume(co, nArgs); status == api.LUA_OK || status == api.LUA_SUSPENDED {
		vm.PushBoolean(true)
	} else {
		vm.PushBoolean(false)
		co.XMove(vm, 1) //errmsg
		return 2
	}
	nRets := co.GetTop()
	vm.CheckStack(nRets) //父协程装不下所有返回值
	co.XMove(vm, nRets)  //将co的返回值转移到父协程中作为resume的返回值
	return nRets + 1
}

// value1, value2, ... = coroutine.yield(arg1, arg2, ...)
// 参数：
// arg1, arg2, ...: 可选参数，这些值将作为对应的resume调用的返回值
// 返回值：
// value1, value2, ...: 当下次恢复协程时，通过resume传递的参数将成为yield的返回值
func coroutineYield(vm api.LuaVM) int {
	return vm.Yield()
}

// coroutine.status(co)
// co必须为coroutine
// 获取co当前的状态信息
func coroutineStatus(vm api.LuaVM) int {
	vm.CheckType(1, api.LUAVALUE_COROUTINE)
	co := vm.ToCoroutine(1)
	switch co.Status() {
	case api.LUA_SUSPENDED:
		vm.PushString("SUSPENDED")
	case api.LUA_NORMAL:
		vm.PushString("NORMAL")
	case api.LUA_RUNNING:
		vm.PushString("RUNNING")
	default:
		vm.PushString("DEAD")
	}
	return 1
}

// co, ismain = coroutine.running()
// 返回值：
// co: 当前运行的协程对象；如果在主协程中调用，返回nil 如果在协程中调用，返回该协程对象
// ismain: 布尔值，表示当前是否在主线程中运行
//
// 用于获取当前正在运行的协程信息
func coroutineRunning(vm api.LuaVM) int {
	isMain := vm.PushCoroutine()
	if isMain {
		vm.Pop(1)
		vm.PushNil()
	}
	vm.PushBoolean(isMain)
	return 2
}

// can_yield = coroutine.isyieldable()
// can_yield:	true=>当前可以调用yield ； false=>当前不能调用yield
func coroutineIsYieldable(vm api.LuaVM) int {
	vm.PushBoolean(vm.IsYieldable())
	return 1
}

// wrapped_func = coroutine.wrap(f)
// f: 要作为协程执行的函数	wrapped_func: 一个可以直接调用的函数，每次调用相当于调用coroutine.resume
//
// coroutine.wrap 创建的协程函数在遇到错误时不会像 coroutine.resume 那样返回false加错误信息而是会直接将错误抛出，中断当前执行流程
func coroutineWrap(vm api.LuaVM) int {
	vm.CheckType(1, api.LUAVALUE_FUNCTION)
	co := vm.NewCoroutine()
	vm.PushValue(1) //func_copy
	vm.XMove(co, 1)
	vm.PushGoFunction(func(lv api.LuaVM) int {
		lv.PushValue(lv.UpvalueIndex(1))
		lv.Insert(1)
		nRets := coroutineResume(lv)
		//如果外部是以pcall的方式调用wrap的方法,需要特殊处理因为resume里面有pcall的逻辑
		//对于PCALL嵌套调用，这里需要将里层的err往上抛，暂时这样todo
		if !lv.ToBoolean(-(nRets - 1)) {
			panic(lv.ToString(-(nRets - 2)))
		}
		return nRets - 1
	}, 1)
	return 1
}
