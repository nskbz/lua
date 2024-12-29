package binchunk

import (
	"encoding/binary"
	"math"
)

func Undump(datas []byte) *Prototype {
	r := reader{byteReader: byteReader{data: datas}}
	r.checkHeader()
	r.readByte() //跳过sizeUpvalues字段
	proto := r.readProto("")
	return proto
}

type reader struct {
	byteReader
}

func (r *reader) checkHeader() {
	if LUA_SIGNATURE != string(r.readBytes(4)) {
		panic("error signature")
	}
	if LUAC_VERSION != r.readByte() {
		panic("error version")
	}
	if LUAC_FORMAT != r.readByte() {
		panic("error format")
	}
	if LUAC_DATA != string(r.readBytes(6)) {
		panic("error luac_data")
	}
	if CINT_SIZE != r.readByte() {
		panic("error cint_size")
	}
	if CSIZET_SIZE != r.readByte() {
		panic("error csizet_size")
	}
	if INSTRUCTION_SIZE != r.readByte() {
		panic("error instruction_size")
	}
	if LUA_INTEGER_SIZE != r.readByte() {
		panic("error lua_int_size")
	}
	if LUA_FLOAT_SIZE != r.readByte() {
		panic("error lua_float_size")
	}
	if LUAC_INT != r.readLuaInteger() {
		panic("error lua_int")
	}
	if LUAC_FLOAT != r.readLuaFloat() {
		panic("error lua_float")
	}

}

func (r *reader) readProto(parentSource string) *Prototype {
	s := r.readString()
	if s == "" {
		s = parentSource
	}
	return &Prototype{
		Source:       s,
		LineStart:    r.readUint32(),
		LineEnd:      r.readUint32(),
		NumParams:    r.readByte(),
		IsVararg:     r.readByte(),
		MaxStackSize: r.readByte(),
		Codes:        r.readCodes(),
		Constants:    r.readConstants(),
		Upvalues:     r.readUpvalues(),
		Protos:       r.readProtos(s),
		LineInfo:     r.readLineInfo(),
		LocVars:      r.readLocVars(),
		UpvalueNames: r.readUpvalueNames(),
	}
}

func (r *reader) readLuaInteger() int64 {
	return int64(r.readUint64())
}

func (r *reader) readLuaFloat() float64 {
	return math.Float64frombits(r.readUint64())
}

func (r *reader) readString() string {
	b := r.readByte()
	if b == 0 {
		return ""
	}
	length := 0
	if b <= 0xFD {
		length = int(b) - 1
	} else {
		length = int(r.readUint64()) - 1
	}
	s := string(r.byteReader.readBytes(uint(length)))
	return s
}

func (r *reader) readCodes() []uint32 {
	length := r.readUint32()
	codes := make([]uint32, length)
	for i := 0; i < int(length); i++ {
		codes[i] = r.readUint32()
	}
	return codes
}

func (r *reader) readConstants() []interface{} {
	length := r.readUint32()
	constants := make([]interface{}, length)
	for i := 0; i < int(length); i++ {
		constants[i] = r.readConstant()
	}
	return constants
}

func (r *reader) readConstant() interface{} {
	tag := r.readByte()
	switch tag {
	case TAG_NIL:
		return nil
	case TAG_BOOLEAN:
		return r.readByte() != 0
	case TAG_INTEGER:
		return r.readLuaInteger()
	case TAG_FLOAT:
		return r.readLuaFloat()
	case TAG_SHORT_STR, TAG_LONG_STR:
		return r.readString()
	}
	panic("error")
}

func (r *reader) readUpvalues() []Upvalue {
	length := r.readUint32()
	uvs := make([]Upvalue, length)
	for i := 0; i < int(length); i++ {
		uvs[i] = Upvalue{
			Instack: r.readByte(),
			Idx:     r.readByte(),
		}
	}
	return uvs
}

func (r *reader) readProtos(parentSource string) []*Prototype {
	length := r.readUint32()
	ps := make([]*Prototype, length)
	for i := 0; i < int(length); i++ {
		ps[i] = r.readProto(parentSource)
	}
	return ps
}

func (r *reader) readLineInfo() []uint32 {
	length := r.readUint32()
	lis := make([]uint32, length)
	for i := 0; i < int(length); i++ {
		lis[i] = r.readUint32()
	}
	return lis
}

func (r *reader) readLocVars() []LocVar {
	length := r.readUint32()
	lvs := make([]LocVar, length)
	for i := 0; i < int(length); i++ {
		lvs[i] = LocVar{
			VarName: r.readString(),
			StartPC: r.readUint32(),
			EndPC:   r.readUint32(),
		}
	}
	return nil
}

func (r *reader) readUpvalueNames() []string {
	length := r.readUint32()
	ss := make([]string, length)
	for i := 0; i < int(length); i++ {
		ss[i] = r.readString()
	}
	return ss
}

type byteReader struct {
	data []byte
}

func (br *byteReader) readByte() byte {
	b := br.data[0]
	br.data = br.data[1:]
	return b
}

func (br *byteReader) readBytes(n uint) []byte {
	bs := make([]byte, n)
	copy(bs, br.data[:n])
	br.data = br.data[n:]
	return bs
}

func (br *byteReader) readUint32() uint32 {
	i := binary.LittleEndian.Uint32(br.data) //小端序读取
	br.data = br.data[4:]
	return i
}

func (br *byteReader) readUint64() uint64 {
	i := binary.LittleEndian.Uint64(br.data)
	br.data = br.data[8:]
	return i
}
