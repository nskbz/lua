package state

import (
	"math"

	"nskbz.cn/lua/number"
)

type table struct {
	_arr []luaValue
	_map map[luaValue]luaValue
}

func newTable(nArr, nRec int) *table {
	t := table{}
	if nArr > 0 {
		t._arr = make([]luaValue, nArr)
	}
	if nRec > 0 {
		t._map = make(map[luaValue]luaValue, nRec)
	}
	return &t
}

func (t *table) get(key luaValue) luaValue {
	idx := keyToInt(key)
	if idx >= 1 && idx <= int64(t.len()) {
		return t._arr[idx-1]
	}
	return t._map[key]
}

// 返回0表示key不能转换为int
func keyToInt(key luaValue) int64 {
	var idx int64 = 0
	if key == nil {
		panic("error table key [nil]")
	}
	if v, ok := key.(float64); ok {
		if math.IsNaN(v) {
			panic("error table key [NaN]")
		}
		if i, ok := number.FloatToInteger(v); ok {
			idx = i
		}
	}
	if v, ok := key.(int64); ok {
		idx = v
	}
	return idx
}

func (t *table) put(key, value luaValue) {
	//
	//存数组中的
	k := keyToInt(key)
	if k >= 1 && k <= int64(t.len()) {
		//lua表是1为起始所以映射为数组时要减1
		t._arr[k-1] = value
		//如果put的元素刚好为arr的末尾且value等于nil进行shrink
		if k == int64(t.len()) && value == nil {
			t.shrink()
		}
		return
	}
	//如果刚好put的是arr末尾的下一索引且value不为nil则扩充arr并进行expand
	if k == int64(t.len())+1 {
		delete(t._map, key) //此时两种情况1.map中有该键值(之前put过的)2.map中没有该键值(从未put过的)。为了保持数据一致性，这里需要对map中该键值进行删除
		if value != nil {
			t._arr = append(t._arr, value)
			t.expand()
		}
		return
	}

	//
	//存map中的1.key不为整数2.key为整数但超过_arr的长度n个(n>1)
	if value == nil {
		delete(t._map, key) //节省空间
		return
	}
	//如果该table还未创建map则创建
	//arr不需要考虑为空问题，因为 var is []int = nil;is = append(is, 1) 这里append可以直接初始化并添加
	if t._map == nil {
		t._map = make(map[luaValue]luaValue, 8)
	}
	t._map[key] = value
}

// 消除arr尾部nil元素
func (t *table) shrink() {
	i := t.len() - 1
	for i >= 0 && t._arr[i] == nil {
		i--
	}
	t._arr = t._arr[0 : i+1]
}

func (t *table) expand() {
	i := t.len()
	for {
		//lua表是1为起始所以要加1查询
		if v, ok := t._map[i+1]; ok {
			delete(t._map, i)
			t._arr = append(t._arr, v)
		} else {
			break
		}
		i++
	}
}

func (t *table) len() int {
	return len(t._arr)
}
