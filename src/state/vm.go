package state

import (
	"math"
)

func (s *luaState) PC() int {
	return s.stack.pc
}

func (s *luaState) AddPC(n int) {
	add := s.stack.pc + n
	if add < 0 || add > math.MaxInt {
		panic("pc overflow")
	}
	s.stack.pc = add
}

func (s *luaState) Fetch() uint32 {
	code := s.stack.closure.proto.Codes[s.PC()]
	s.AddPC(1)
	return code
}

func (s *luaState) GetConst(idx int) {
	if idx < 0 || idx >= len(s.stack.closure.proto.Constants) {
		panic("constant's index out of range")
	}
	c := s.stack.closure.proto.Constants[idx]
	s.stack.push(c)
}

func (s *luaState) GetRK(arg int) {
	if arg > 0xFF {
		//最高位不为0表示常量表索引
		s.GetConst(arg & 0xFF)
	} else {
		//最高位为0表示寄存器索引
		s.PushValue(arg + 1)
	}
}

func (s *luaState) LoadProto(idx int) {
	proto := s.stack.closure.proto.Protos[idx]
	closure := newClosure(proto)
	s.stack.push(closure)
}

func (s *luaState) RegisterCount() int {
	return int(s.stack.closure.proto.MaxStackSize)
}

func (s *luaState) LoadVarargs(n int) {
	if n < 0 {
		n = len(s.stack.varargs)
	}
	s.CheckStack(n)
	s.stack.pushN(s.stack.varargs, n)
}
