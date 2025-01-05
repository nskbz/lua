package test

import (
	"fmt"
	"strings"
	"testing"

	"nskbz.cn/lua/api"
	"nskbz.cn/lua/state"
)

func newState() api.LuaState {
	state := state.NewState(20)
	state.PushString("JACK")
	state.PushInteger(1)
	state.PushInteger(2)
	state.PushInteger(3)
	state.PushInteger(4)
	state.PushInteger(5)
	return state
}

// success=true
func testState(s api.LuaState, vals ...interface{}) bool {
	return valString(vals...) == stackString(s)
}

func valString(vals ...interface{}) string {
	str := strings.Builder{}
	for _, v := range vals {
		switch x := v.(type) {
		case bool:
			str.WriteString(fmt.Sprintf("[%t]", x))
		case int, int64:
			str.WriteString(fmt.Sprintf("[%d]", x))
		case float64:
			str.WriteString(fmt.Sprintf("[%f]", x))
		case string:
			str.WriteString(fmt.Sprintf("[%q]", x))
		case nil:
			str.WriteString("[nil]")
		default:
			str.WriteString(fmt.Sprintf("[%s]", x))
		}
	}
	return str.String()
}

func stackString(s api.LuaState) string {
	str := strings.Builder{}
	top := s.GetTop()
	for i := 1; i <= top; i++ {
		tp := s.Type(i)
		switch tp {
		case api.LUAVALUE_BOOLEAN:
			str.WriteString(fmt.Sprintf("[%t]", s.ToBoolean(i)))
		case api.LUAVALUE_NUMBER:
			str.WriteString(fmt.Sprintf("[%g]", s.ToFloat(i)))
		case api.LUAVALUE_STRING:
			str.WriteString(fmt.Sprintf("[%q]", s.ToString(i)))
		default:
			str.WriteString(fmt.Sprintf("[%s]", s.TypeName(tp)))
		}
	}
	return str.String()
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
	if !testState(state, "JACK", "JACK", 2) {
		t.Fail()
	}
}

func TestRotate(t *testing.T) {
	state := newState()
	state.Insert(1)
	if !testState(state, 5, "JACK", 1, 2, 3, 4) {
		t.Fail()
	}
	state.Remove(1)
	if !testState(state, "JACK", 1, 2, 3, 4) {
		t.Fail()
	}
}

func TestPushXXX(t *testing.T) {
	state := newState()
	state.PushNil()
	state.PushBoolean(false)
	state.PushNumber(22.2)
	if !testState(state, "JACK", 1, 2, 3, 4, 5, nil, false, 22.2) {
		t.Fail()
	}
}

func TestState(t *testing.T) {
	s := state.NewState(20)
	s.PushBoolean(true)
	s.PushInteger(10)
	s.PushNil()
	s.PushString("hello")
	s.PushValue(-3)
	s.Replace(3)
	if !testState(s, true, 10, true, "hello") {
		t.FailNow()
	}
	s.SetTop(6)
	if !testState(s, true, 10, true, "hello", nil, nil) {
		t.FailNow()
	}
	s.Remove(-2)
	if !testState(s, true, 10, true, nil, nil) {
		t.FailNow()
	}
	s.SetTop(-4)
	if !testState(s, true) {
		t.FailNow()
	}
}

func TestOperation(t *testing.T) {
	ls := state.NewState(20)
	ls.PushInteger(1)
	ls.PushString("2.0")
	ls.PushString("3.0")
	ls.PushNumber(4.0)
	if !testState(ls, 1, "2.0", "3.0", 4) {
		t.FailNow()
	}
	ls.Arith(api.ArithOp_ADD)
	ls.Arith(api.ArithOp_NOT)
	ls.Len(2)
	if !testState(ls, 1, "2.0", -8, 3) {
		t.FailNow()
	}
	ls.Concat(3)
	if !testState(ls, 1, "2.0-83") {
		t.FailNow()
	}
	ls.PushBoolean(ls.Compare(1, 2, api.CompareOp_EQ))
	if !testState(ls, 1, "2.0-83", false) {
		t.FailNow()
	}
}
