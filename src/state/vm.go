package state

import (
	"fmt"
	"math"

	"nskbz.cn/lua/instruction"
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

// 获取一条指令同时PC++
func (s *luaState) Fetch() uint32 {
	code := s.stack.closure.proto.Codes[s.PC()]
	s.AddPC(1)
	return uint32(code)
}

func (s *luaState) GetConst(idx int) {
	if idx < 0 || idx >= len(s.stack.closure.proto.Constants) {
		panic(fmt.Sprintf("constant index[%d] out of range", idx))
	}
	c := s.stack.closure.proto.Constants[idx]
	s.stack.push(c)
}

func (s *luaState) GetRK(rk int) {
	if rk < instruction.ConstantBase { //这里解释了为什么cg(代码生成)阶段,注册常量时需要加上256
		//最高位为0表示寄存器索引
		s.PushValue(rk + 1)
	} else {
		//最高位不为0表示常量表索引
		s.GetConst(rk & 0xFF)
	}
}

// 加载子函数
//
// 可以看出closure是通过proto生成的并辅以当前环境(Upvalue捕获)
// 所以任何的方法都是从proto定义出来的,即proto决定closure
func (s *luaState) LoadProto(idx int) {
	proto := s.stack.closure.proto.Protos[idx]
	c := newLuaClosure(proto)

	for i, v := range proto.Upvalues {
		uidx := int(v.Idx)
		if v.Instack == 0 { //==0 表示该捕获变量属于函数的外部
			c.upvals[i] = s.stack.closure.upvals[uidx]
		} else if v.Instack == 1 { //==1 表示该捕获变量属于函数的内部
			//记录打开的捕获变量
			if up, ok := s.stack.openuvs[uidx]; !ok {
				uv := upvalue{&s.stack.slots[uidx+1]} //UpValue会被捕获,要想统一修改,就只能是指针
				c.upvals[i] = uv
				s.stack.openuvs[uidx] = uv
			} else {
				c.upvals[i] = up
			}
		}
	}

	s.stack.push(c)
}

func (s *luaState) RegisterCount() int {
	return int(s.stack.closure.proto.MaxRegisterSize)
}

func (s *luaState) LoadVarargs(n int) {
	if n < 0 {
		n = len(s.stack.varargs)
	}
	s.CheckStack(n)
	s.stack.pushN(s.stack.varargs, n)
}
