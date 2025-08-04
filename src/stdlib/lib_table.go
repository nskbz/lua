package stdlib

import (
	"fmt"
	"strings"

	"nskbz.cn/lua/api"
	"nskbz.cn/lua/tool"
)

var tableFuncs map[string]api.GoFunc = map[string]api.GoFunc{
	"insert": tableInsert,
	"remove": tableRemove,
	"sort":   tableSort,
	"move":   tableMove,
	"pack":   tablePack,
	"unpack": tableUnpack,
	"concat": tableConcat,
}

func OpenTableLib(vm api.LuaVM) int {
	vm.NewTable()
	for k, v := range tableFuncs {
		vm.PushGoFunction(v, 0)
		vm.SetField(-1, k)
	}
	return 1
}

// table.insert(tbl, [pos,] value)
// 在表 tbl 的 pos 位置插入 value（默认插入末尾），没有返回值
func tableInsert(vm api.LuaVM) int {
	vm.CheckType(1, api.LUAVALUE_TABLE)
	nArgs := vm.GetTop()
	tbLen := vm.Len2(1)
	if nArgs == 3 {
		if i, ok := vm.ToIntegerX(2); !ok {
			return vm.Error2("expectd 2th arg is integer, but %s", vm.Type(2).String())
		} else {
			for j := tbLen; j >= i; j-- {
				vm.GetI(1, j)
				vm.SetI(1, j+1)
			}
			vm.SetI(1, i)
			return 0
		}
	} else if nArgs == 2 {
		vm.SetI(1, tbLen+1)
		return 0
	}
	return vm.Error2("table insert error!")
}

// table.remove(tbl, [pos])
// 移除表 tbl 中 pos 位置的元素（默认移除最后一个），返回删除的值
func tableRemove(vm api.LuaVM) int {
	vm.CheckType(1, api.LUAVALUE_TABLE)
	nArgs := vm.GetTop()
	tbLen := vm.Len2(1)
	idx := tbLen
	if nArgs == 2 {
		if i, ok := vm.ToIntegerX(2); !ok {
			return vm.Error2("expectd 2th arg is integer, but %s", vm.Type(2).String())
		} else {
			idx = i
		}
	}
	vm.RemoveI(1, idx)
	return 1
}

// table.sort(tbl, [comp])
// 对表 tbl 进行排序（可自定义比较函数 comp）
func tableSort(vm api.LuaVM) int {
	vm.CheckType(1, api.LUAVALUE_TABLE)
	var compare func(interface{}, interface{}) bool
	if vm.GetTop() > 1 && vm.Type(2) == api.LUAVALUE_FUNCTION {
		compare = func(i1, i2 interface{}) bool {
			vm.PushValue(2) //压入luafunc
			vm.PushBasic(i1)
			vm.PushBasic(i2)
			vm.Call(2, 1)
			return vm.ToBoolean(0) //栈顶的返回值就是compare的比较结果
		}
	}
	vm.SortI(1, compare)
	return 0
}

// table.move(a1, f, e, t [, a2 ])
// 将表 a1 的 [f, e] 范围元素复制到 a2 的 t 位置，a2未指定则默认a1
// 返回目标表 a2（如果指定）或 a1（如果未指定 a2）
func tableMove(vm api.LuaVM) int {
	nArgs := vm.GetTop()
	vm.CheckType(1, api.LUAVALUE_TABLE)
	vm.CheckInteger(2)
	vm.CheckInteger(3)
	vm.CheckInteger(4)
	src := 1
	dest := src
	f := vm.ToInteger(2)
	e := vm.ToInteger(3)
	t := vm.ToInteger(4)
	if nArgs > 4 {
		vm.CheckType(5, api.LUAVALUE_TABLE)
		dest = 5
	}

	span := int(e - f + 1)

	//先使dest表扩容以至于可以装载迁移的数据
	for i := 0; i < span; i++ {
		vm.PushString("move占位")
		l := vm.RawLen(dest) //获取table._arr的长度
		vm.SetI(dest, int64(l+1))
	}
	//如果是自生转移还需要将元素后挪
	for i := t + int64(span); i <= int64(vm.RawLen(dest)); i++ {
		vm.GetI(dest, i-int64(span))
		vm.SetI(dest, i)
	}

	//挪动[f,e]区间内的元素至目的位置
	for i, j := f, t; i <= e; i++ {
		vm.GetI(src, i)
		vm.SetI(dest, j)
		j++
	}

	vm.PushValue(dest) //返回目标表
	return 1
}

// table.pack(...)
// 用于将可变数量的参数打包成一个表，并自动记录参数个数
//
// 返回一个包含以下内容的表：
// 所有传入的参数作为表的数组部分（索引从1开始）
// 一个额外的字段 n，记录参数的总个数
func tablePack(vm api.LuaVM) int {
	nArgs := vm.GetTop()

	vm.NewTable()
	for i := 1; i <= nArgs; i++ {
		vm.PushValue(i)
		vm.SetI(-1, int64(i))
	}

	vm.PushInteger(int64(nArgs))
	vm.SetField(-1, "n") //添加'n'字段记录元素个数

	return 1
}

// table.unpack(list [, i [, j]])
func tableUnpack(vm api.LuaVM) int {
	nArgs := vm.GetTop()
	vm.CheckType(1, api.LUAVALUE_TABLE)
	start := 1
	end := vm.RawLen(1)
	if nArgs > 1 {
		vm.CheckInteger(2)
		vm.CheckInteger(3)
		start = int(vm.ToInteger(2))
		end = int(vm.ToInteger(3))
	}
	n := end - start + 1
	for start <= end {
		vm.GetI(1, int64(start))
		start++
	}

	if n > 0 {
		return n
	}
	return 0
}

// table.concat(list [, sep] [, i [, j]])
//
// list	要连接的数组(table)	必填
// sep	分隔符(字符串)	空字符串""
// i	起始索引	1
// j	结束索引	#list
// 返回连接后的字符串
//
// 用于将数组(table)中的元素连接成字符串
func tableConcat(vm api.LuaVM) int {
	nArgs := vm.GetTop()
	vm.CheckType(1, api.LUAVALUE_TABLE)
	separator := " "
	start := 1
	end := vm.RawLen(1)
	switch nArgs {
	case 2, 4:
		vm.CheckString(2)
		separator = vm.ToString(2)
		if nArgs == 4 {
			vm.CheckInteger(3)
			vm.CheckInteger(4)
			start = int(vm.ToInteger(3))
			end = int(vm.ToInteger(4))
		}
	}
	n := end - start + 1
	sb := strings.Builder{}
	for ; start <= end; start++ {
		vm.GetI(1, int64(start))
		if s, ok := vm.ToStringX(0); !ok {
			tool.Fatal(vm, fmt.Sprintf("table concat error,not support %s", vm.Type(0).String()))
		} else {
			sb.WriteString(s)
			sb.WriteString(separator)
		}
	}
	str := sb.String()
	if n > 0 {
		str = strings.TrimSuffix(str, separator)
	}
	vm.PushString(str)
	return 1
}
