package state

import (
	"nskbz.cn/lua/api"
	"nskbz.cn/lua/binchunk"
)

type closure struct {
	proto  *binchunk.Prototype //lua函数实例
	goFunc api.GoFunc          //go函数实例
}

func newClosure(proto *binchunk.Prototype) *closure {
	return &closure{proto: proto}
}

func newGoClosure(gf api.GoFunc) *closure {
	return &closure{goFunc: gf}
}
