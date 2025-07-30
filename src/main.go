package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"nskbz.cn/lua/binchunk"
	"nskbz.cn/lua/compile"
	"nskbz.cn/lua/compile/lexer"
	"nskbz.cn/lua/compile/parser"
	"nskbz.cn/lua/state"
	"nskbz.cn/lua/tool"
)

func main() {

	var c bool
	flag.BoolVar(&c, "c", false, "是否只是编译")
	flag.IntVar(&tool.LogLevel, "d", tool.LOG_DEFAULT, "log输出信息级别")
	flag.Parse()
	if len(flag.Args()) == 0 {
		panic("no specified file!!!")
	}
	chunk := flag.Arg(0) //第一个非'-'参数必须为文件名

	f, err := os.Open(chunk)
	if err != nil {
		panic(err.Error())
	}
	data, err := io.ReadAll(f)
	if err != nil {
		panic(err.Error())
	}

	if c { //只进行编译操作，则于stdout输出json格式
		proto := compile.Compile(data, chunk)
		// to do 生成可供LUA虚拟机执行的二进制文件
		// suffix := chunk[strings.LastIndex(chunk, "."):]
		// target := strings.Replace(chunk, suffix, ".out", 1)
		// f, err := os.Open(target)
		// if err != nil {
		// 	panic(err.Error())
		// }
		// f.Write(proto.ToBytes())
		protoInfo := binchunk.ProtoToProtoInfo(proto)
		fmt.Printf("chunck file====>[%s]\n\n", chunk)
		s, err := json.Marshal(protoInfo)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println(string(s))
		return
	}

	//没有-c参数,则文件有可能是lua二进制文件,也有可能是lua源文件
	vm := state.New()
	vm.OpenLibs()
	vm.LoadFile(chunk)
	vm.Call(0, -1)
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
