package state

import "fmt"

//以1为起始索引的lua栈
//栈顶(top)指向最新的val

type luaStack struct {
	slots []luaValue
	top   int
}

func newLuaStack(size int) *luaStack {
	return &luaStack{
		slots: make([]luaValue, size+1),
		top:   0,
	}
}

func (s *luaStack) len() int {
	return len(s.slots) - 1
}

func (s *luaStack) empty() bool {
	return s.top == 0
}

func (s *luaStack) full() bool {
	return s.top == s.len()
}

// 扩容
func (s *luaStack) expand(n int) {
	for i := 0; i < n; i++ {
		s.slots = append(s.slots, nil)
	}
}

func (s *luaStack) checkIdx(absidx int) {
	if absidx <= 0 || absidx > s.top {
		panic(fmt.Sprintf("stack access[%d] out of limit[1,%d]!!", absidx, s.top))
	}
}

func (s *luaStack) get(absidx int) luaValue {
	s.checkIdx(absidx)
	return s.slots[absidx]
}

func (s *luaStack) set(absidx int, val luaValue) {
	s.checkIdx(absidx)
	s.slots[absidx] = val
}

func (s *luaStack) pop() luaValue {
	if s.empty() {
		panic("stack empty")
	}
	topval := s.slots[s.top]
	s.slots[s.top] = nil
	s.top--
	return topval
}

func (s *luaStack) push(val luaValue) {
	if s.full() {
		panic("stack full")
	}
	s.top++
	s.slots[s.top] = val
}

func (s *luaStack) reverse(from, to int) {
	if from > to {
		s.reverse(to, from)
		return
	}
	for from < to {
		temp := s.slots[from]
		s.slots[from] = s.slots[to]
		s.slots[to] = temp
		from++
		to--
	}
}
