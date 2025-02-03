package state

func doEq(a, b luaValue, ls *luaState) bool {
	switch x := a.(type) {
	case nil:
		return b == nil
	case bool:
		y, ok := b.(bool)
		return ok && y == x
	case string:
		y, ok := b.(string)
		return ok && y == x
	case int64:
		switch y := b.(type) {
		case int64:
			return x == y
		case float64:
			return float64(x) == y
		default:
			return false
		}
	case float64:
		switch y := b.(type) {
		case float64:
			return x == y
		case int64:
			return int64(x) == y
		default:
			return false
		}
	case *table: //表类型支持元方法
		//ls!=nil用于判断是否采用元方法。当ls不为nil时采用元方法;ls为nil时不采用元方法
		if y, ok := b.(*table); ok && x != y && ls != nil {
			if c := getMetaClosure(ls, META_EQ, a, b); c != nil {
				result := callMetaClosure(ls, c, 1, a, b)
				return convertToBoolean(result[0])
			}
		}
	}
	return a == b
}

// 能比较大小的只有string,int64,float64
func doLt(a, b luaValue, ls *luaState) bool {
	switch x := a.(type) {
	case string:
		if y, ok := b.(string); ok {
			return x < y
		}
	case int64:
		switch y := b.(type) {
		case int64:
			return x < y
		case float64:
			return float64(x) < y
		}
	case float64:
		switch y := b.(type) {
		case float64:
			return x < y
		case int64:
			return int64(x) < y
		}
	case *table:
		//ls!=nil用于判断是否采用元方法。当ls不为nil时采用元方法;ls为nil时不采用元方法
		if y, ok := b.(*table); ok && x != y && ls != nil {
			if c := getMetaClosure(ls, META_LT, a, b); c != nil {
				result := callMetaClosure(ls, c, 1, a, b)
				return convertToBoolean(result[0])
			}
		}
	}
	panic("can't compare_lt")
}

func doLe(a, b luaValue, ls *luaState) bool {
	switch x := a.(type) {
	case string:
		if y, ok := b.(string); ok {
			return x <= y
		}
	case int64:
		switch y := b.(type) {
		case int64:
			return x <= y
		case float64:
			return float64(x) <= y
		}
	case float64:
		switch y := b.(type) {
		case float64:
			return x <= y
		case int64:
			return int64(x) <= y
		}
	case *table:
		//ls!=nil用于判断是否采用元方法。当ls不为nil时采用元方法;ls为nil时不采用元方法
		if y, ok := b.(*table); ok && x != y && ls != nil {
			if c := getMetaClosure(ls, META_LE, a, b); c != nil {
				result := callMetaClosure(ls, c, 1, a, b)
				return convertToBoolean(result[0])
			} else if c := getMetaClosure(ls, META_LT, a, b); c != nil {
				//a<=b equal !(b<a)
				result := callMetaClosure(ls, c, 1, b, a)
				return !convertToBoolean(result[0])
			}
		}
	}
	panic("can't compare_le")
}
