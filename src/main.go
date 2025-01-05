package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"nskbz.cn/lua/api"
	"nskbz.cn/lua/binchunk"
	"nskbz.cn/lua/instruction"
	"nskbz.cn/lua/state"
)

func main() {
	f, err := os.Open("../luac.out")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	datas, err := io.ReadAll(f)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	proto := binchunk.Undump(datas)
	luaLoop(proto)
}

func luaLoop(proto *binchunk.Prototype) {
	stackSize := proto.MaxStackSize
	vm := state.NewVM(int(stackSize), proto)
	vm.SetTop(6)
	for {
		pc := vm.PC()
		i := instruction.Instruction(vm.Fetch())
		if i.InstructionName() == "RETURN  " {
			break
		}
		i.Execute(vm)
		fmt.Printf("[%02d] %s ", pc+1, i.InstructionName())
		printStack(vm)
	}
}

func list(proto *binchunk.Prototype) {
	printHeader(proto)
	printCode(proto)
	printDetail(proto)
	for _, v := range proto.Protos {
		fmt.Println("===============================================================")
		list(v)
	}
}

func printHeader(proto *binchunk.Prototype) {
	funcType := "main"
	if proto.LineStart > 0 {
		funcType = "function"
	}
	varargFlag := "" //可变参数标记
	if proto.IsVararg != 0 {
		varargFlag = "+"
	}
	fmt.Printf("%s <%s:%d,%d> (%d instructions)\n", funcType, proto.Source,
		proto.LineStart, proto.LineEnd, len(proto.Codes))

	fmt.Printf("%d%s params, %d slots, %d upvalues, %d locals, %d constants, %d functions\n",
		proto.NumParams, varargFlag, proto.MaxStackSize, len(proto.Upvalues),
		len(proto.LocVars), len(proto.Constants), len(proto.Protos))
}

func printCode(proto *binchunk.Prototype) {
	for i, code := range proto.Codes {
		line := "-"
		if len(proto.LineInfo) != 0 {
			line = fmt.Sprintf("%d", proto.LineInfo[i])
		}
		info := codeInfo(code)
		fmt.Printf("\t%d\t[%s]\t%s\n", i+1, line, info)
	}
}

func codeInfo(code uint32) string {
	build := strings.Builder{}
	i := instruction.Instruction(code)
	switch i.OpMode() {
	case instruction.IABC:
		a, b, c := i.ABC()
		build.WriteString(fmt.Sprintf("%s %d", i.InstructionName(), a))
		if i.ModArgB(instruction.ArgU) {
			if b > 0xFF {
				b = -1 - b&0xFF
			}
			build.WriteString(fmt.Sprintf(" %d", b))
		}
		if i.ModArgC(instruction.ArgU) {
			if c > 0xFF {
				c = -1 - c&0xFF
			}
			build.WriteString(fmt.Sprintf(" %d", c))
		}
	case instruction.IABx:
		a, bx := i.ABx()
		build.WriteString(fmt.Sprintf("%s %d", i.InstructionName(), a))
		if i.ModArgB(instruction.ArgU) {
			if i.ModArgB(instruction.ArgK) {
				bx = -1 - bx
			}
			build.WriteString(fmt.Sprintf(" %d", bx))
		}
	case instruction.IAsBx:
		a, sbx := i.AsBx()
		build.WriteString(fmt.Sprintf("%s %d %d", i.InstructionName(), a, sbx))
	case instruction.IAx:
		ax := i.Ax()
		build.WriteString(fmt.Sprintf("%s %d", i.InstructionName(), -1-ax))
	}
	return build.String()
}

func printDetail(proto *binchunk.Prototype) {
	fmt.Printf("constants (%d):\n", len(proto.Constants))
	for i, constant := range proto.Constants {
		fmt.Printf("\t%d\t%s\n", i+1, constantToString(constant))
	}

	fmt.Printf("locals (%d):\n", len(proto.LocVars))

	for i, local := range proto.LocVars {
		fmt.Printf("\t%d\t%s\t%d\t%d\n", i,
			local.VarName, local.StartPC+1, local.EndPC+1)
	}

	fmt.Printf("upvalues (%d):\n", len(proto.Upvalues))

	for i, upvalue := range proto.Upvalues {
		fmt.Printf("\t%d\t%s\t%d\t%d\n", i,
			upvalueName(proto, int(upvalue.Idx)), upvalue.Instack, upvalue.Idx)
	}
}

func constantToString(constant interface{}) string {
	switch constant.(type) {
	case nil:
		return "nil"
	case bool:
		return fmt.Sprintf("%t", constant)
	case int64:
		return fmt.Sprintf("%d", constant)
	case float64:
		return fmt.Sprintf("%g", constant)
	case string:
		return fmt.Sprintf("%q", constant)
	default:
		return "?"
	}
}

func upvalueName(proto *binchunk.Prototype, index int) string {
	if len(proto.UpvalueNames) > 0 {
		return proto.UpvalueNames[index]
	}
	return "-"
}

func printStack(s api.LuaState) {
	top := s.GetTop()
	for i := 1; i <= top; i++ {
		tp := s.Type(i)
		switch tp {
		case api.LUAVALUE_BOOLEAN:
			fmt.Printf("[%t]", s.ToBoolean(i))
		case api.LUAVALUE_NUMBER:
			fmt.Printf("[%g]", s.ToFloat(i))
		case api.LUAVALUE_STRING:
			fmt.Printf("[%q]", s.ToString(i))
		default:
			fmt.Printf("[%s]", s.TypeName(tp))
		}
	}
	fmt.Println()
}
