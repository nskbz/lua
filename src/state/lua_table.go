package state

import (
	"math"

	"nskbz.cn/lua/number"
)

type table struct {
	metaTable *table                //元方法表
	_arr      []luaValue            //顺序数组下标的key
	_map      map[luaValue]luaValue //非顺序数组下标的key

	keys map[luaValue]luaValue //key的顺序
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

	//存map中的1.key不为整数2.key为整数但超过_arr的长度n个(n>1)
	//
	//table中不存在val为nil的键值对
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

	//改变table结构后key的顺序可能发生变化需要重新init
	t.keys = nil
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

func (t *table) hasMetaFunc(key string) bool {
	if t.metaTable == nil {
		return false
	}
	if val := t.metaTable.get(key); val != nil {
		return true
	}
	return false
}

// func (t *table) _initKeys() {
// 	//table结构并未发生改变不需要重新init
// 	if t.keys != nil {
// 		return
// 	}
// 	keys := make(map[luaValue]luaValue)
// 	var key luaValue = nil
// 	for i, v := range t._arr {
// 		if v != nil {
// 			keys[key] = int64(i + 1)
// 			key = int64(i + 1)
// 		}
// 	}
// 	for k, v := range t._map {
// 		if v != nil {
// 			keys[key] = k
// 			key = k
// 		}
// 	}

// 	t.keys = keys
// }

func (t *table) nextKey(k luaValue) luaValue {
	//table结构发生改变需要重新init
	if t.keys == nil {
		keys := make(map[luaValue]luaValue)
		var key luaValue = nil
		for i, v := range t._arr {
			if v != nil {
				keys[key] = int64(i + 1)
				key = int64(i + 1)
			}
		}
		for k, v := range t._map {
			if v != nil {
				keys[key] = k
				key = k
			}
		}

		t.keys = keys
	}
	return t.keys[k]
}
