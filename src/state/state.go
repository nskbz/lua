package state

import (
	"fmt"
	"strings"

	"nskbz.cn/lua/api"
	"nskbz.cn/lua/number"
)

const DefaultStackSize = 20

type luaState struct {
	stack *luaStack
}

func New() *luaState {
	return &luaState{
		stack: newLuaStack(DefaultStackSize),
	}
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

func (s *luaState) isValidIdx(absidx int) bool {
	if absidx <= 0 || absidx > s.stack.len() {
		return false
	}
	return true
}

func (s *luaState) CheckStack(n int) bool {
	if n < 0 {
		return false
	}
	available := s.stack.len() - s.stack.top
	s.stack.expand(n - available)
	return true
}

func (s *luaState) Pop(n int) {
	s.SetTop(-n)
}

func (s *luaState) Copy(fromIdx, toIdx int) {
	form := s.AbsIndex(fromIdx)
	to := s.AbsIndex(toIdx)
	fv := s.stack.get(form)
	s.stack.set(to, fv)
}

func (s *luaState) PushValue(idx int) {
	absidx := s.AbsIndex(idx)
	val := s.stack.get(absidx)
	s.stack.push(val)
}

func (s *luaState) Replace(idx int) {
	absidx := s.AbsIndex(idx)
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
		panic(fmt.Sprintf("no supported arith for %d", op))
	}
	var result luaValue
	b := s.stack.pop()
	//区分一元运算与二元运算
	if op >= api.ArithOp_OPPOSITE {
		result = doUnitaryArith(b, operation)
	} else {
		a := s.stack.pop()
		result = doDualArith(a, b, operation)
	}
	if result == nil {
		panic(fmt.Sprintf("Arith error =>%d", op))
	}
	s.stack.push(result)
}

func (s *luaState) Compare(idx1, idx2 int, op api.CompareOp) bool {
	a := s.stack.get(s.AbsIndex(idx1))
	b := s.stack.get(s.AbsIndex(idx2))
	switch op {
	case api.CompareOp_EQ:
		return doEq(a, b)
	case api.CompareOp_LT:
		return doLt(a, b)
	case api.CompareOp_LE:
		return doLe(a, b)
	}
	panic(fmt.Sprintf("no supported compare for %d", op))
}

func (s *luaState) Len(idx int) {
	absidx := s.AbsIndex(idx)
	val := s.stack.get(absidx)
	switch x := val.(type) {
	case string:
		s.stack.push(int64(len(x)))
	default:
		panic(fmt.Sprintf("no supported length for %v", x))
	}
}

func (s *luaState) Concat(n int) {
	if n < 0 {
		return
	}
	str := strings.Builder{}
	from := s.AbsIndex(-(n - 1))
	s.stack.reverse(from, s.stack.top)
	for i := 0; i < n; i++ {
		val := s.stack.pop()
		s, ok := convertToString(val)
		if !ok {
			panic("error for concat")
		}
		str.WriteString(s)
	}
	s.stack.push(str.String())
}
