package state

import "nskbz.cn/lua/binchunk"

type closure struct {
	proto *binchunk.Prototype
}

func newClosure(proto *binchunk.Prototype) *closure {
	return &closure{proto: proto}
}
