package state

import (
	"fmt"

	"nskbz.cn/lua/api"
	"nskbz.cn/lua/number"
)

/*
*	Meta Function definition
 */

const (
	META_ADD       = "__add"
	META_SUB       = "__sub"
	META_MUL       = "__mul"
	META_MOD       = "__mod"
	META_POW       = "__pow"
	META_DIV       = "__div"
	META_IDIV      = "__idiv"
	META_AND       = "__band"
	META_OR        = "__bor"
	META_XOR       = "__bxor"
	META_SHL       = "__shl"
	META_SHR       = "__shr"
	META_OPPOSITE  = "__unm"
	META_NOT       = "__bnot"
	META_LEN       = "__len"
	META_CONCAT    = "__concat"
	META_EQ        = "__eq"
	META_LT        = "__lt"
	META_LE        = "__le"
	META_INDEX     = "__index"
	META_NEW_INDEX = "__newindex"
	META_CALL      = "__call"
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

/*
* 类型元表->[K（元方法名）,V（元方法实现）]
 */

// 设置类型元表
func setMetaTable(target luaValue, mt *table, ls *luaState) {
	//如果target是table
	if t, ok := target.(*table); ok {
		t.metaTable = mt
		return
	}
	//如果target是非table类型,则每一种类型对应一个mt
	key := metaKey(target)
	ls.registry.put(key, mt)
}

// 获取类型元表
func getMetaTable(target luaValue, ls *luaState) *table {
	//如果target是table
	if t, ok := target.(*table); ok {
		return t.metaTable
	}
	//如果target是非table类型
	key := metaKey(target)
	if t, ok := ls.registry.get(key).(*table); ok {
		return t
	}
	return nil
}

// 从vals中依次尝试获取元方法，如若vals中都没有元方法则返回nil
func getMetaClosure(ls *luaState, key string, vals ...luaValue) luaValue {
	for _, v := range vals {
		mt := getMetaTable(v, ls)
		if mt != nil {
			return mt.get(key)
		}
	}
	return nil
}

func metaKey(val luaValue) string {
	return fmt.Sprintf("_MT_%s", typeOf(val).String())
}

// 调用元方法并弹出nResult个返回值
func callMetaClosure(ls *luaState, function luaValue, nResult int, params ...luaValue) []luaValue {
	var c *closure
	//如果function为nil或其他不为*closure的类型panic
	if closure, ok := function.(*closure); !ok {
		panic("call meta func error!")
	} else {
		c = closure
	}
	nArgs := len(params)
	ls.CheckStack(1 + nArgs)
	ls.stack.push(c)           //推入元方法
	for _, v := range params { //推入参数
		ls.stack.push(v)
	}
	ls.Call(nArgs, nResult)
	return ls.stack.popN(nResult)
}
