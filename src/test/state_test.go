package test

import (
	"fmt"
	"testing"

	"nskbz.cn/lua/api"
	"nskbz.cn/lua/state"
)

func newState() api.LuaState {
	state := state.New()
	state.PushString("JACK")
	state.PushInteger(1)
	state.PushInteger(2)
	state.PushInteger(3)
	state.PushInteger(4)
	state.PushInteger(5)
	return state
}

func printStack(s api.LuaState) {
	top := s.GetTop()
	for i := 1; i <= top; i++ {
		tp := s.Type(i)
		switch tp {
		case api.LUAVALUE_BOOLEAN:
			fmt.Printf("[%t]", s.ToBoolean(i))
		case api.LUAVALUE_NUMBER:
			fmt.Printf("[%g]", s.ToFloat(i))
		case api.LUAVALUE_STRING:
			fmt.Printf("[%q]", s.ToString(i))
		default:
			fmt.Printf("[%s]", s.TypeName(tp))
		}
	}
	fmt.Println()
}

func TestSetTop(t *testing.T) {
	state := newState()
	state.CheckStack(20)
	state.SetTop(26)
	state.SetTop(1)
	if state.GetTop() != 1 {
		t.Fail()
	}
}

func TestPop(t *testing.T) {
	state := newState()
	state.Pop(3)
	state.PushValue(1)
	state.Replace(2)
	//todo assert
}

func TestRotate(t *testing.T) {
	state := newState()
	state.Insert(1)
	state.Remove(1)
	//todo assert
}

func TestPushXXX(t *testing.T) {
	state := newState()
	state.PushNil()
	//todo assert
	state.PushBoolean(false)
	//todo assert
	state.PushNumber(22.2)
	//todo assert
}

func TestState(t *testing.T) {
	s := state.New()
	s.PushBoolean(true)
	printStack(s)
	s.PushInteger(10)
	printStack(s)
	s.PushNil()
	printStack(s)
	s.PushString("hello")
	printStack(s)
	s.PushValue(-3)
	printStack(s)
	s.Replace(3)
	printStack(s)
	s.SetTop(6)
	printStack(s)
	s.Remove(-2)
	printStack(s)
	s.SetTop(-4)
	printStack(s)
}
