package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"nskbz.cn/lua/api"
	"nskbz.cn/lua/compile/lexer"
	"nskbz.cn/lua/compile/parser"
	"nskbz.cn/lua/state"
)

// var reDecEscapeSeq = regexp.MustCompile(`^\\[0-9]{1,3}`)          //十进制ASCII码
// var reHexEscapeSeq = regexp.MustCompile(`^\\x[0-9a-fA-F]{2}`)     //十六进制ASCII码
// var reUnicodeEscapeSeq = regexp.MustCompile(`^\\u{[0-9a-fA-F]+}`) //unicode码

func main() {
	var chunk, source string
	var run, compile bool
	flag.BoolVar(&run, "r", false, "...")
	flag.BoolVar(&compile, "c", false, "...")
	flag.StringVar(&chunk, "i", "../luac.out", "执行的chunk文件")
	flag.StringVar(&source, "s", "../test.lua", "需要编译的原文件")
	flag.Parse()
	if run {
		fmt.Printf("chunck file====>[%s]\n", chunk)
		f, err := os.Open(chunk)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		datas, err := io.ReadAll(f)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		vm := state.New()
		vm.Register("print", print)
		vm.Register("getmetatable", getMetaTable)
		vm.Register("setmetatable", setMetaTable)
		vm.Register("next", next)
		vm.Register("pairs", pairs)
		vm.Register("ipairs", iPairs)
		vm.Register("error", luaError)
		vm.Register("pcall", pCall)
		vm.Load(datas, "test.lua", "b")
		vm.Call(0, 0)
	} else if compile {
		fmt.Printf("source file====>[%s]\n", source)
		f, err := os.Open(source)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		datas, err := io.ReadAll(f)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		//testLexer(datas, source)
		testParser(datas, source)
	}
}

func testParser(data []byte, name string) {
	ast := parser.Parse(data, name)
	b, err := json.MarshalIndent(ast, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func testLexer(data []byte, name string) {
	l := lexer.NewLexer(data, name)

	for {
		t := l.NextToken()
		fmt.Printf("[%2d] [%-10s] %s\n", t.Line(), tokenKindToCategory(t.Kind()), t.Val())
		if t.Kind() == lexer.TOKEN_EOF {
			break
		}
	}
}

func tokenKindToCategory(kind int) string {
	switch {
	case kind < lexer.TOKEN_SEP_SEMI:
		return "other"
	case kind <= lexer.TOKEN_SEP_RCURLY:
		return "separator"
	case kind <= lexer.TOKEN_OP_NOT:
		return "operator"
	case kind <= lexer.TOKEN_KW_WHILE:
		return "keyword"
	case kind == lexer.TOKEN_IDENTIFIER:
		return "identifier"
	case kind == lexer.TOKEN_NUMBER:
		return "number"
	case kind == lexer.TOKEN_STRING:
		return "string"
	default:
		return "other"
	}
}

func print(vm api.LuaVM) int {
	nArgs := vm.GetTop()
	for i := 1; i <= nArgs; i++ {
		if vm.IsBoolean(i) {
			fmt.Printf("%t", vm.ToBoolean(i))
		} else if vm.IsString(i) {
			fmt.Print(vm.ToString(i))
		} else {
			fmt.Print(vm.TypeName(vm.Type(i)))
		}
		if i < nArgs {
			fmt.Print("\t")
		}
	}
	fmt.Println()
	return 0
}

func getMetaTable(vm api.LuaVM) int {
	if !vm.GetMetaTable(1) {
		vm.PushNil()
	}
	return 1
}

func setMetaTable(vm api.LuaVM) int {
	vm.SetMetaTable(1)
	return 1
}

func next(vm api.LuaVM) int {
	vm.SetTop(2)
	if vm.Next(1) {
		return 2
	} else {
		vm.PushNil()
		return 1
	}
}

// 用于map遍历
func pairs(vm api.LuaVM) int {
	vm.PushGoFunction(next, 0)
	vm.PushValue(1)
	vm.PushNil()
	return 3
}

// 用于数组遍历
func iPairs(vm api.LuaVM) int {
	vm.PushGoFunction(func(vm api.LuaVM) int {
		i := vm.ToInteger(2) + 1 //index先+1,所以下标从1开始
		vm.PushInteger(i)
		vm.Replace(2)         //保存index
		vm.PushInteger(i)     //压入index
		val := vm.GetTable(1) //压入stack[index]
		if val != api.LUAVALUE_NIL {
			return 2
		}
		return 1
	}, 0)
	vm.PushValue(1)
	vm.PushInteger(0) //index初始化为0
	return 3
}

func luaError(vm api.LuaVM) int {
	return vm.Error()
}

func pCall(vm api.LuaVM) int {
	nArgs := vm.GetTop() - 1
	status := vm.PCall(nArgs, -1, 0)
	vm.PushBoolean(status == api.LUA_OK)
	vm.Insert(1)
	return vm.GetTop()
}
