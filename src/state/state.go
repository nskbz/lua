package state

import (
	"fmt"

	"nskbz.cn/lua/api"
	"nskbz.cn/lua/binchunk"
	"nskbz.cn/lua/compile"
	"nskbz.cn/lua/instruction"
	"nskbz.cn/lua/number"
	"nskbz.cn/lua/tool"
)

type luaState struct {
	stack    *luaStack
	registry *table
}

func New() api.LuaVM {
	r := newTable(0, 0)                         //新建注册表
	r.put(api.LUA_GLOBALS_RIDX, newTable(0, 0)) //添加全局环境进注册表

	ls := &luaState{registry: r}
	ls.stack = newLuaStack(api.LUA_MIN_STACK, ls)
	return ls
}

func (s *luaState) GetTop() int {
	return s.stack.top
}

func (s *luaState) SetTop(idx int) {
	newtop := s.AbsIndex(idx)
	count := newtop - s.stack.top
	if newtop > s.stack.top {
		for i := 0; i < count; i++ {
			s.stack.push(nil)
		}
		return
	}
	for i := 0; i < -count; i++ {
		s.stack.pop()
	}
}

// 返回绝对索引，此方法保证返回的都是“可用索引”
//
//	AbsIndex(0)==top index
func (s *luaState) AbsIndex(idx int) int {
	if idx <= api.LUA_REGISTRY_INDEX {
		return idx
	}

	absidx := 0
	if idx > 0 {
		absidx = idx
	} else {
		absidx = s.stack.top + idx
	}
	if !s.isValidIdx(absidx) {
		panic("lua _state_stack overflow")
	}
	return absidx
}

/*
*	Upvalue支持
 */
func (s *luaState) UpvalueIndex(i int) int {
	return api.LUA_REGISTRY_INDEX - i
}

// 由于Upvalue是对LocalVar的引用，所以当LocalVar要释放时应当通过这个方法关闭引用它的Upvalue
func (s *luaState) CloseUpvalues(a int) {
	//to do 这里不是很懂
	//这里为什么要将拷贝Upvalue估计是为了支持并发操作，即使LocalVar释放了，但仍可以操作Upvalue不过无法影响原LocalVar
	for i, v := range s.stack.openuvs {
		if i >= a-1 {
			value := *v.val
			v.val = &value
			delete(s.stack.openuvs, i)
		}
	}
}

func (s *luaState) isValidIdx(absidx int) bool {
	if absidx == api.LUA_REGISTRY_INDEX {
		return true
	}
	if absidx <= 0 || absidx > s.stack.len() {
		return false
	}
	return true
}

func (s *luaState) CheckStack(n int) {
	if n < 0 {
		tool.Fatal(s, "stack can not expand")
	}
	available := s.stack.len() - s.stack.top
	s.stack.expand(n - available)
}

func (s *luaState) Pop(n int) {
	s.SetTop(-n)
}

func (s *luaState) Copy(fromIdx, toIdx int) {
	form := s.AbsIndex(fromIdx)
	to := s.AbsIndex(toIdx)
	fr := s.stack.get(form)
	s.stack.set(to, fr)
}

func (s *luaState) PushValue(idx int) {
	absidx := s.AbsIndex(idx)
	val := s.stack.get(absidx)
	s.stack.push(val)
}

func (s *luaState) Replace(idx int) {
	absidx := s.AbsIndex(idx)
	top := s.stack.top
	if top == idx {
		return
	}
	s.stack.set(absidx, s.stack.pop())
}

func (s *luaState) Rotate(idx, n int) {
	absidx := s.AbsIndex(idx)
	if absidx > s.stack.top {
		panic("can't rotate beyond of top")
	}
	size := s.stack.top - absidx + 1
	times := number.AbsInt(n)
	vals := make([]luaValue, size)
	if n < 0 {
		for i := 0; i < size; i++ {
			vals[i] = s.stack.get(s.stack.top - i) //n<0就反转
		}
		vs := s.doRotate(vals, times)
		for i := 0; i < size; i++ {
			s.stack.set(s.AbsIndex((size+times-i-1)%size+absidx), vs[size+times-i-1]) //将正向rotate的结果再反转回来
		}
		s.stack.reverse(absidx, s.stack.top)
		return
	}
	for i := 0; i < size; i++ {
		vals[i] = s.stack.get(absidx + i)
	}
	vs := s.doRotate(vals, times)
	for i := 0; i < size; i++ {
		s.stack.set(s.AbsIndex((times+i)%size+absidx), vs[times+i])
	}
}

func (s *luaState) doRotate(vals []luaValue, times int) []luaValue {
	size := len(vals)
	for i := 0; i < times; i++ {
		vals = append(vals, nil)
	}
	for i := size - 1; i >= 0; i-- {
		vals[i+times] = vals[i]
	}
	return vals
}

func (s *luaState) Insert(idx int) {
	absidx := s.AbsIndex(idx)
	s.Rotate(absidx, 1)
}

func (s *luaState) Remove(idx int) {
	absidx := s.AbsIndex(idx)
	s.Rotate(absidx, -1)
	s.Pop(1)
}

/*
*	压栈操作
 */
func (s *luaState) PushNil()              { s.stack.push(nil) }
func (s *luaState) PushBoolean(b bool)    { s.stack.push(b) }
func (s *luaState) PushInteger(n int64)   { s.stack.push(n) }
func (s *luaState) PushNumber(n float64)  { s.stack.push(n) } // todo 看能否改变方法名换为Float
func (s *luaState) PushString(str string) { s.stack.push(str) }

/*
*	栈元素访问
 */

func (s *luaState) TypeName(tp api.LuaValueType) string {
	switch tp {
	case api.LUAVALUE_NONE:
		return "no value"
	case api.LUAVALUE_NIL:
		return "nil"
	case api.LUAVALUE_BOOLEAN:
		return "bool"
	case api.LUAVALUE_NUMBER:
		return "number"
	case api.LUAVALUE_STRING:
		return "string"
	case api.LUAVALUE_TABLE:
		return "table"
	case api.LUAVALUE_FUNCTION:
		return "function"
	case api.LUAVALUE_THREAD:
		return "thread"
	}
	return "userdata"
}

func (s *luaState) Type(idx int) api.LuaValueType {
	absidx := s.AbsIndex(idx)
	return typeOf(s.stack.get(absidx))
}

func (s *luaState) IsNone(idx int) bool {
	absidx := s.AbsIndex(idx)
	return absidx > s.stack.top
}

func (s *luaState) IsNil(idx int) bool {
	return s.Type(idx) == api.LUAVALUE_NIL
}

func (s *luaState) IsNoneOrNil(idx int) bool {
	return s.IsNone(idx) || s.IsNil(idx)
}

func (s *luaState) IsBoolean(idx int) bool {
	return s.Type(idx) == api.LUAVALUE_BOOLEAN
}

func (s *luaState) IsInteger(idx int) bool {
	_, b := s.ToIntegerX(idx)
	return b
}

func (s *luaState) IsFloat(idx int) bool {
	_, b := s.ToFloatX(idx)
	return b
}

func (s *luaState) IsString(idx int) bool {
	tp := s.Type(idx)
	return tp == api.LUAVALUE_STRING || tp == api.LUAVALUE_NUMBER
}

func (s *luaState) IsTable(idx int) bool {
	return true
}

func (s *luaState) IsThread(idx int) bool {
	return true
}

func (s *luaState) IsFunction(idx int) bool {
	return true
}

func (s *luaState) ToBoolean(idx int) bool {
	absidx := s.AbsIndex(idx)
	val := s.stack.get(absidx)
	return convertToBoolean(val)
}

func (s *luaState) ToInteger(idx int) int64 {
	i, _ := s.ToIntegerX(idx)
	return i
}

func (s *luaState) ToIntegerX(idx int) (int64, bool) {
	absidx := s.AbsIndex(idx)
	val := s.stack.get(absidx)
	return convertToInteger(val)
}

func (s *luaState) ToFloat(idx int) float64 {
	n, _ := s.ToFloatX(idx)
	return n
}

func (s *luaState) ToFloatX(idx int) (float64, bool) {
	absidx := s.AbsIndex(idx)
	val := s.stack.get(absidx)
	return convertToFloat(val)
}

func (s *luaState) ToString(idx int) string {
	str, _ := s.ToStringX(idx)
	return str
}

func (s *luaState) ToStringX(idx int) (string, bool) {
	absidx := s.AbsIndex(idx)
	val := s.stack.get(absidx)
	return convertToString(val)
}

/*
*	运算操作
 */
func (s *luaState) Arith(op api.ArithOp) {
	operation, ok := arith_operation[op]
	if !ok {
		panic(fmt.Sprintf("no supported arith for %s", op.String()))
	}
	var result luaValue
	var success bool
	b := s.stack.pop()
	//区分一元运算与二元运算
	if op >= api.ArithOp_OPPOSITE {
		result, success = doUnitaryArith(b, operation, s)
	} else {
		a := s.stack.pop()
		result, success = doDualArith(a, b, operation, s)
	}

	if !success {
		panic(fmt.Sprintf("Arith error =>%s", op.String()))
	}
	s.stack.push(result) //结果压入栈
}

func (s *luaState) Compare(idx1, idx2 int, op api.CompareOp) bool {
	a := s.stack.get(s.AbsIndex(idx1))
	b := s.stack.get(s.AbsIndex(idx2))
	switch op {
	case api.CompareOp_EQ:
		return doEq(a, b, s)
	case api.CompareOp_LT:
		return doLt(a, b, s)
	case api.CompareOp_LE:
		return doLe(a, b, s)
	}
	panic(fmt.Sprintf("no supported compare for %s", op.String()))
}

func (s *luaState) Len(idx int) {
	absidx := s.AbsIndex(idx)
	val := s.stack.get(absidx)
	switch x := val.(type) {
	case string:
		s.stack.push(int64(len(x)))
	case *table:
		if c := getMetaClosure(s, META_LEN, val); c != nil {
			result := callMetaClosure(s, c, 1, val)
			s.stack.push(result[0])
			break
		}
		s.stack.push(int64(x.len()))
	default:
		panic(fmt.Sprintf("no supported length for %v", x))
	}
}

func (s *luaState) Concat(n int) {
	if n < 0 || n == 1 {
		return
	}
	if n == 0 {
		s.stack.push("")
		return
	}
	from := s.AbsIndex(-(n - 1))
	s.stack.reverse(from, s.stack.top)
	for i := n; i > 1; i-- { //当n==1时只剩一个元素故结束
		a := s.stack.pop()
		b := s.stack.pop()
		if x, ok := convertToString(a); ok {
			if y, ok := convertToString(b); ok {
				s.stack.push(x + y)
				continue
			}
		}

		//如果不能转换为string,则调用元方法
		if c := getMetaClosure(s, META_CONCAT, a, b); c != nil {
			result := callMetaClosure(s, c, 1, a, b)
			s.stack.push(result[0])
		} else {
			panic("do not found meta function for concat")
		}
	}
}

/*
*表相关操作
 */
func (s *luaState) NewTable() {
	s.CreateTable(0, 0)
}

func (s *luaState) CreateTable(nArr, nRec int) {
	table := newTable(nArr, nRec)
	s.stack.push(table)
}

func (s *luaState) GetTable(idx int) api.LuaValueType {
	absidx := s.AbsIndex(idx)
	t := s.stack.get(absidx)
	k := s.stack.pop()
	return s.getTableVal(t, k, false)
}

func (s *luaState) GetField(idx int, k string) api.LuaValueType {
	absidx := s.AbsIndex(idx)
	t := s.stack.get(absidx)
	return s.getTableVal(t, k, false)
}

func (s *luaState) GetI(idx int, i int64) api.LuaValueType {
	absidx := s.AbsIndex(idx)
	t := s.stack.get(absidx)
	return s.getTableVal(t, i, false)
}

// 获取t中键k的val的类型，并将val压入栈顶
func (s *luaState) getTableVal(t luaValue, k luaValue, raw bool) api.LuaValueType {
	if api.LUAVALUE_TABLE == typeOf(t) {
		tb := t.(*table)
		//是否采用元方法||t是表但t[k]是否存在||t没有META_INDEX元方法
		if raw || tb.get(k) != nil || !tb.hasMetaFunc(META_INDEX) {
			val := tb.get(k)
			s.stack.push(val)
			return typeOf(val)
		}
	}

	//采用元方法,t不是表或k在表中的值不存在
	if !raw {
		if val := getMetaClosure(s, META_INDEX, t); val != nil {
			switch x := val.(type) {
			case *table: //t是表但不存在k的值
				return s.getTableVal(x, k, false)
			case *closure: //t拥有META_INDEX函数
				result := callMetaClosure(s, val, 1, t, k)
				s.stack.push(result[0])
				return typeOf(result[0])
			}
		}
	}
	panic(fmt.Sprintf("type[%d] is not a table", typeOf(t)))
}

func (s *luaState) SetTable(idx int) {
	absidx := s.AbsIndex(idx)
	t := s.stack.get(absidx)
	val := s.stack.pop()
	key := s.stack.pop()
	s.setTableKV(t, key, val, false)
}

func (s *luaState) SetField(idx int, k string) {
	absidx := s.AbsIndex(idx)
	t := s.stack.get(absidx)
	val := s.stack.pop()
	s.setTableKV(t, k, val, false)
}

func (s *luaState) SetI(idx int, i int64) {
	absidx := s.AbsIndex(idx)
	t := s.stack.get(absidx)
	val := s.stack.pop()
	s.setTableKV(t, i, val, false)
}

func (s *luaState) setTableKV(t luaValue, k, v luaValue, raw bool) {
	if api.LUAVALUE_TABLE == typeOf(t) {
		tb := t.(*table)
		if raw || tb.get(k) != nil || !tb.hasMetaFunc(META_NEW_INDEX) {
			tb.put(k, v)
			return
		}
	}

	//采用元方法,t不是表或k在表中的值不存在
	if !raw {
		if val := getMetaClosure(s, META_NEW_INDEX, t); val != nil {
			switch x := val.(type) {
			case *table: //t是表但不存在k的值
				s.setTableKV(x, k, v, false)
				return
			case *closure: //t拥有META_NEW_INDEX函数
				callMetaClosure(s, x, 0, val, t, k, v)
				return
			}
		}
	}

	panic(fmt.Sprintf("type[%d] is not a table", typeOf(t)))
}

/*
*	函数调用栈扩展
*
* 	函数调用利用单向链表实现，链表头部是目前执行的函数(被调函数)，其prev是主调函数，main函数的prev为nil
* 	采用链表(prev)可以使return的时候方便找到主调函数
*	function test()
*   	print(123)
*	end
*
*	test()
*
*	call stack: 	nil<-main	(call main)
*	call stack: 	nil<-main<-test	(call test)
*	call stack: 	nil<-main<-test<-print	(call print)
*	call stack: 	nil<-main<-test	(print return)
*	call stack: 	nil<-main	(test return)
*	call stack: 	nil	(means obver)
 */
func (s *luaState) pushContext(f *luaStack) {
	f.prev = s.stack
	s.stack = f //切换执行函数
}

func (s *luaState) popContext() {
	outerCall := s.stack.prev
	s.stack.prev = nil
	s.stack = outerCall //切换执行函数
}

func (s *luaState) Load(chunk []byte, chunckName, mode string) int {
	var proto *binchunk.Prototype
	if binchunk.LUA_SIGNATURE == string(chunk[:4]) {
		proto = binchunk.Undump(chunk)
	} else {
		proto = compile.Compile(chunk, chunckName) //如果不是LUA二进制形式,则对其进行编译
	}
	c := newLuaClosure(proto)
	if len(c.upvals) > 0 {
		env := s.registry.get(api.LUA_GLOBALS_RIDX)
		c.upvals[0] = upvalue{&env}
	}
	s.stack.push(c)
	return api.LUA_OK
}

func (s *luaState) doLuaFunc(nResults int, c *closure, args []luaValue) {
	stackSize := int(c.proto.MaxRegisterSize)
	numParams := int(c.proto.NumParams)
	isVararg := c.proto.IsVararg == 1

	//初始化被调函数栈，即创建调用帧
	//栈大小为api.LUA_MIN_STACK+stackSize
	//下面会使stack的top指向stackSize
	//也就是说寄存器的数量为stackSize
	//api.LUA_MIN_STACK为临时空间,用于计算
	//[1,stacksize]&[stacksize+1,stacksize+api.LUA_MIN_STACK]
	//      |					|
	//  寄存器空间			   临时空间
	stack := newLuaStack(api.LUA_MIN_STACK+stackSize, s)
	stack.closure = c
	if isVararg && len(args) > numParams {
		stack.varargs = args[numParams:]
	}
	stack.pushN(args, numParams)
	stack.top = stackSize //这一步必须放在push参数后位置才对

	//切换上下文并调用函数
	s.pushContext(stack)
	s.doLuaFuncCall()
	s.popContext()

	//保存返回值至主调函数栈
	//nResults==0则不返回任何值
	//nResults<0则返回值全部压入,nResults>0则返回nResults个返回值
	if nResults != 0 {
		results := stack.popN(stack.top - stackSize)
		s.CheckStack(len(results))
		s.stack.pushN(results, nResults) //nResults<0则返回值全部压入,nResults>0则返回nResults个返回值
	}
}

func (s *luaState) doLuaFuncCall() {
	for {
		i := instruction.Instruction(s.Fetch()) //每次获取下一指令的同时会使PC++，即又指向下一个指令
		tool.Debug(i.Info())
		i.Execute(s)
		if i.Name() == "RETURN  " {
			break
		}
	}
}

func (s *luaState) doGoFunc(nResults int, c *closure, args []luaValue) {
	//准备Go函数调用帧，Go函数的调用帧栈不需要寄存器所以无需设置top值
	nArgs := len(args)
	stack := newLuaStack(api.LUA_MIN_STACK+nArgs, s)
	stack.pushN(args, nArgs)
	stack.closure = c

	//Go函数调用执行
	s.pushContext(stack)
	nr := c.goFunc(s)
	s.popContext()

	//nResults==0则不返回任何值
	//nResults<0则返回值全部压入,nResults>0则返回nResults个返回值
	if nResults != 0 {
		results := stack.popN(nr)
		s.CheckStack(len(results))
		s.stack.pushN(results, nResults) //nResults<0则返回值全部压入,nResults>0则返回nResults个返回值
	}
}

func (s *luaState) Call(nArgs, nResults int) {
	vals := s.stack.popN(nArgs + 1) //弹出1个func和nArgs个参数
	c, ok := vals[0].(*closure)
	//如若不能转换为函数则寻找该值的META_CALL元方法
	if !ok {
		if mc := getMetaClosure(s, META_CALL, vals[0]); mc == nil {
			//不是函数且没有META_CALL元方法报错
			tool.Fatal(s, fmt.Sprintf("[%s] is not a closure", typeOf(vals[0]).String()))
		} else {
			if _c, ok := mc.(*closure); ok {
				c = _c
				vals = append([]luaValue{nil}, vals...)
			}
		}
	}

	if c.goFunc != nil {
		s.doGoFunc(nResults, c, vals[1:])
	} else {
		s.doLuaFunc(nResults, c, vals[1:])
	}
}

/*
*	Go函数外部调用支持
 */
func (s *luaState) PushGoFunction(gf api.GoFunc, n int) {
	gc := newGoClosure(gf, n)
	for i := 0; i < n; i++ {
		val := s.stack.pop()
		gc.upvals[n-i-1] = upvalue{&val}
	}
	s.stack.push(gc)
}

func (s *luaState) IsGoFunction(idx int) bool {
	absidx := s.AbsIndex(idx)
	gf := s.stack.get(absidx)
	if c, ok := gf.(*closure); ok {
		return c.goFunc != nil
	}
	return false
}

func (s *luaState) ToGoFunction(idx int) api.GoFunc {
	absidx := s.AbsIndex(idx)
	gf := s.stack.get(absidx)
	if c, ok := gf.(*closure); ok {
		return c.goFunc
	}
	panic(fmt.Sprintf("[%s] is not a closure", typeOf(gf).String()))
}

/*
*	全局环境支持
 */
func (s *luaState) PushGlobalTable() {
	s.stack.push(s.registry.get(api.LUA_GLOBALS_RIDX))
}

func (s *luaState) GetGlobal(key string) api.LuaValueType {
	s.PushGlobalTable()
	s.PushString(key)
	return s.GetTable(-1)
}

func (s *luaState) SetGlobal(key string) {
	global := s.registry.get(api.LUA_GLOBALS_RIDX)
	val := s.stack.pop()
	s.setTableKV(global, key, val, true)
}

func (s *luaState) Register(key string, gf api.GoFunc) {
	s.stack.push(newGoClosure(gf, 0))
	s.SetGlobal(key)
}

/*
*	元编程支持
 */
func (s *luaState) GetMetaTable(idx int) bool {
	absidx := s.AbsIndex(idx)
	val := s.stack.get(absidx)
	if mt := getMetaTable(val, s); mt != nil {
		s.stack.push(mt)
		return true
	}
	return false
}

func (s *luaState) SetMetaTable(idx int) {
	absidx := s.AbsIndex(idx)
	val := s.stack.get(absidx)
	mt := s.stack.pop()
	if mt == nil {
		setMetaTable(val, nil, s)
	} else if t, ok := mt.(*table); ok {
		setMetaTable(val, t, s)
	} else {
		panic("table expected!")
	}
}

func (s *luaState) RawLen(idx int) uint {
	absidx := s.AbsIndex(idx)
	val := s.stack.get(absidx)
	switch x := val.(type) {
	case string:
		return uint(len(x))
	case *table:
		return uint(x.len())
	}
	return 0
}

func (s *luaState) RawEqual(idx1, idx2 int) bool {
	absidx1 := s.AbsIndex(idx1)
	absidx2 := s.AbsIndex(idx2)
	a := s.stack.get(absidx1)
	b := s.stack.get(absidx2)
	return doEq(a, b, nil)
}

func (s *luaState) RawGet(idx int) api.LuaValueType {
	absidx := s.AbsIndex(idx)
	t := s.stack.get(absidx)
	k := s.stack.pop()
	return s.getTableVal(t, k, true)
}

func (s *luaState) RawSet(idx int) {
	absidx := s.AbsIndex(idx)
	t := s.stack.get(absidx)
	v := s.stack.pop()
	k := s.stack.pop()
	s.setTableKV(t, k, v, true)
}

func (s *luaState) RawGetI(idx int, i int64) api.LuaValueType {
	absidx := s.AbsIndex(idx)
	t := s.stack.get(absidx)
	return s.getTableVal(t, i, true)
}

func (s *luaState) RawSetI(idx int, i int64) {
	absidx := s.AbsIndex(idx)
	t := s.stack.get(absidx)
	v := s.stack.pop()
	s.setTableKV(t, i, v, true)
}

/*
*	通用for循环支持
 */
func (s *luaState) Next(idx int) bool {
	absidx := s.AbsIndex(idx)
	val := s.stack.get(absidx)

	if t, ok := val.(*table); ok {
		k := s.stack.pop()
		nk := t.nextKey(k)
		nv := t.get(nk)
		if nk == nil {
			return false
		}
		s.stack.push(nk)
		s.stack.push(nv)
		return true
	}

	panic(fmt.Sprintf("stack[%d] is not a table", absidx))
}

/*
*	异常处理支持
 */
func (s *luaState) Error() int {
	err := s.stack.pop() //默认错误在栈顶
	panic(err)
}

func (s *luaState) PCall(nArgs, nResults, msgh int) (status int) {
	status = api.LUA_ERR_RUN
	caller := s.stack //存储调用函数栈
	defer func() {
		if err := recover(); err != nil {
			for s.stack != caller {
				s.popContext()
			} //恢复至调用函数上下文
			s.stack.push(err) //在调用函数栈中压入err
		}
	}()
	s.Call(nArgs, nResults)
	status = api.LUA_OK
	return
}
