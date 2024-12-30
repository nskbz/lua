package state

import "nskbz.cn/lua/api"

type luaValue interface{}

func typeOf(val luaValue) api.LuaValueType {
	switch val.(type) {
	case nil:
		return api.LUAVALUE_NIL
	case bool:
		return api.LUAVALUE_BOOLEAN
	case int64, float64:
		return api.LUAVALUE_NUMBER
	case string:
		return api.LUAVALUE_STRING
	}
	panic("todo!")
}

func convertToBoolean(val luaValue) bool {
	switch v := val.(type) {
	case nil:
		return false
	case bool:
		return v
	}
	return true
}
