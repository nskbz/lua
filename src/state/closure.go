package state

import (
	"nskbz.cn/lua/api"
	"nskbz.cn/lua/binchunk"
)

type closure struct {
	proto  *binchunk.Prototype //lua函数实例
	goFunc api.GoFunc          //go函数实例
	upvals []upvalue           //捕获的变量列表,这里捕获的是变量的地址,所以可以借由该字段完成对捕获变量的修改
}

// 由于返回的luaValue有值类型，必须采用指针才能统一修改
// 所以包一层方便使用luaValue指针
type upvalue struct {
	val *luaValue
}

func newLuaClosure(proto *binchunk.Prototype) *closure {
	c := &closure{proto: proto}
	//c.upvals = make([]upvalue, len(proto.Upvalues)+1) //每个新的closure都会有'_ENV'
	if nUpvals := len(proto.Upvalues); nUpvals > 0 {
		c.upvals = make([]upvalue, nUpvals) //初始化交给API，这里只创建
	}
	return c
}

func newGoClosure(gf api.GoFunc, n int) *closure {
	gc := &closure{goFunc: gf}
	if n > 0 {
		gc.upvals = make([]upvalue, n)
	}
	return gc
}
