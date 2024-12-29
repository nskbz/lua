package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"nskbz.cn/lua/binchunk"
	"nskbz.cn/lua/vm"
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
	list(proto)
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
	i := vm.Instruction(code)
	op := i.Create()
	switch i.OpMode() {
	case vm.IABC:
		abc := op.(vm.ABC)
		build.WriteString(fmt.Sprintf("%s %d", abc.InstructionName(), abc.A()))
		if abc.UsedArgB() {
			b := abc.B()
			if b > 0xFF {
				b = -1 - b&0xFF
			}
			build.WriteString(fmt.Sprintf(" %d", b))
		}
		if abc.UsedArgC() {
			c := abc.C()
			if c > 0xFF {
				c = -1 - c&0xFF
			}
			build.WriteString(fmt.Sprintf(" %d", c))
		}
	case vm.IABx:
		abx := op.(vm.ABx)
		build.WriteString(fmt.Sprintf("%s %d", abx.InstructionName(), abx.A()))
		if abx.UsedArgB() {
			bx := abx.Bx()
			if abx.BisArgK() {
				bx = -1 - bx
			}
			build.WriteString(fmt.Sprintf(" %d", bx))
		}
	case vm.IAsBx:
		asbx := op.(vm.AsBx)
		build.WriteString(fmt.Sprintf("%s %d %d", asbx.InstructionName(), asbx.A(), asbx.Sbx()))
	case vm.IAx:
		ax := op.(vm.Ax)
		build.WriteString(fmt.Sprintf("%s %d", ax.InstructionName(), -1-ax.Ax()))
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
