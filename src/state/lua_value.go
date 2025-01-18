package state

import (
	"fmt"

	"nskbz.cn/lua/api"
	"nskbz.cn/lua/number"
)

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
	case *table:
		return api.LUAVALUE_TABLE
	case *closure:
		return api.LUAVALUE_FUNCTION
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

func convertToFloat(val luaValue) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	case string:
		return number.ParseFloat(v)
	}
	return 0, false
}

func convertToInteger(val luaValue) (int64, bool) {
	switch v := val.(type) {
	case int64:
		return v, true
	case float64:
		return number.FloatToInteger(v)
	case string:
		//如果字符串可以转换为整数
		if i, ok := number.ParseInteger(v); ok {
			return i, ok
		}
		//如果字符串可以转换为小数
		if f, ok := number.ParseFloat(v); ok {
			return number.FloatToInteger(f) //再转换成整数
		}
	}
	return 0, false
}

func convertToString(val luaValue) (string, bool) {
	switch v := val.(type) {
	case string:
		return v, true
	case float64, int64:
		sv := fmt.Sprintf("%v", v)
		return sv, true
	}
	return "", false
}
