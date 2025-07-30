package codegen

import "nskbz.cn/lua/binchunk"

func toProto(fi *funcInfo) *binchunk.Prototype {
	var isVararg byte = 0
	if fi.isVararg {
		isVararg = 1
	}
	proto := &binchunk.Prototype{
		LineStart:       uint32(fi.exp.DefLine),
		LineEnd:         uint32(fi.exp.LastLine),
		NumParams:       byte(fi.numParams),
		IsVararg:        isVararg,
		MaxRegisterSize: byte(fi.maxRegs),
		Codes:           fi.instructions,
		Constants:       _getConstants(fi),
		Upvalues:        nil,
		Protos:          _getProtos(fi),
		Source:          fi.funcName,       //debug
		LineInfo:        fi.lineOfIns,      //debug
		LocVars:         _getLocalVars(fi), //debug
		UpvalueNames:    nil,               //debug
	}
	proto.Upvalues, proto.UpvalueNames = _getUpvalues(fi)

	return proto
}

func _getConstants(fi *funcInfo) []interface{} {
	constants := make([]interface{}, len(fi.constants))
	for k, v := range fi.constants {
		constants[v] = k
	}
	return constants
}

func _getUpvalues(fi *funcInfo) ([]binchunk.Upvalue, []string) {
	upvalues := []binchunk.Upvalue{}
	upvalueNames := []string{}
	var inStack byte
	var idx byte
	for k, v := range fi.upvalVars {
		if v.localVarSlot >= 0 {
			inStack = 1
			idx = byte(v.localVarSlot)
		}
		if v.upvalIndex >= 0 {
			inStack = 0
			idx = byte(v.upvalIndex)
		}
		upvalues = append(upvalues, binchunk.Upvalue{
			Instack: inStack,
			Idx:     idx,
		})
		upvalueNames = append(upvalueNames, k)
	}
	return upvalues, upvalueNames
}

func _getProtos(fi *funcInfo) []*binchunk.Prototype {
	subProtos := make([]*binchunk.Prototype, len(fi.subFuncs))
	for i, v := range fi.subFuncs {
		subProtos[i] = toProto(v)
		subProtos[i].LineStart = uint32(v.exp.DefLine)
		subProtos[i].LineEnd = uint32(v.exp.LastLine)
	}
	return subProtos
}

func _getLocalVars(fi *funcInfo) []binchunk.LocVar {
	locVars := []binchunk.LocVar{}
	for _, v := range fi.localVars {
		locVars = append(locVars, binchunk.LocVar{
			VarName:   v.name,
			StartLine: uint32(v.startLine),
			EndLine:   uint32(v.endLine),
		})
	}
	return locVars
}
