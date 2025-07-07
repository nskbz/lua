package ast

/*
exp ::=
nil | false | true | Numeral | LiteralString |
‘...’ |
functiondef | tableconstructor |
exp binop exp | unop exp |
prefixexp

prefixexp ::= var | functioncall | ‘(’ exp ‘)’
var ::=  Name | prefixexp ‘[’ exp ‘]’ | prefixexp ‘.’ Name
functioncall ::=  prefixexp args | prefixexp ‘:’ Name args
*/
type Exp interface{}

/*
字面量表达式
exp ::=nil | false | true | Numeral | LiteralString
*/
type NilExp struct{ Line int }
type FalseExp struct{ Line int }
type TrueExp struct{ Line int }

// Numeral
type IntExp struct {
	Line int
	Val  int64
}
type FloatExp struct {
	Line int
	Val  float64
}

// LiteralString
type StringExp struct {
	Line int
	Str  string
} //字面量作值的类型
type NameExp struct {
	Line int
	Name string
} //字面量作标识符的类型

/*
vararg表达式
exp ::= '...'
*/
type VarargExp struct{ Line int }

/*
运算符表达式
exp ::=exp binop exp | unop exp
*/
type UnitaryOpExp struct {
	Line int
	Op   int //运算符标号
	A    Exp
} //单目运算符，只有一个表达式(操作数)A

type DualOpExp struct {
	Line int
	Op   int //运算符标号
	A    Exp
	B    Exp
} //双目运算符

type ConcatExp struct {
	Line int
	Exps []Exp //需要连接的表达式
} //连接运算符

/*
构造器表达式
exp ::=functiondef | tableconstructor
*/

// 表构造表达式
// tableconstructor ::= ‘{’ [fieldlist] ‘}’
// fieldlist ::= field {fieldsep field} [fieldsep]
// field ::= ‘[’ exp ‘]’ ‘=’ exp | Name ‘=’ exp | exp
// fieldsep ::= ‘,’ | ‘;’
//
// forexample：
// local arr = {10, 20, 30}
// local dict = {name = "Alice", age = 25}
//
//	local mixed = {
//	    "apple",  -- 自动分配索引1
//	    "banana", -- 索引2
//	    color = "red",  -- 键值对
//	    [5] = "five"    -- 显式指定索引5
//	}
type TableConstructExp struct {
	Line     int // line of '{',for debug
	LastLine int // line of  '}',for debug
	Keys     []Exp
	Vals     []Exp
}

// 函数定义表达式
// functiondef ::= function funcbody
// funcbody ::= ‘(’ [parlist] ‘)’ block end
// parlist ::= namelist [‘,’ ‘...’] | ‘...’
// namelist ::= Name {‘,’ Name}
//
// for example:
// local function max(num1, num2)
//
//	   if (num1 > num2) then
//	      result = num1;
//	   else
//	      result = num2;
//	   end
//		return result;
//
// end
type FuncDefExp struct {
	ArgList  []string
	IsVararg bool
	Block    *Block

	DefLine  int // line of  'function'
	LastLine int // line of 'end'
}

/*
前缀表达式
prefixexp ::= Name

	| '(' exp ')'
	| prefixexp '[' exp ']'
	| prefixexp '.' Name
	| prefixexp ':' Name args
	| prefixexp args
*/
type PrefixExp struct {
	Exp Exp
}

// 表访问表达式
// PrefixExp.CurrentExp
type TableAccessExp struct {
	LastLine   int // line of ']'
	PrefixExp  Exp //'.'前一个表达式
	CurrentExp Exp //当前表达式
}

// 函数调用表达式
// functioncall ::=prefixexp [ ':'  Name ]  args
// args ::='('  [explist]  ')'  |  tableconstructor  | LiteralString
//
// 例子：
// print "Hello"  -- 等同于 print("Hello")
//
// function show(t)
//
//	print(t[1], t[2])
//
// end
//
// show {10, 20}  -- 输出: 10 20
// show ({10, 20})  -- 输出: 10 20 与上面等价
//
// 所以可知在lua中函数调用的参数列表可以是标准的以'('开头
// 也可以直接以字符串或表构造式的形式开头
type FuncCallExp struct {
	Line     int   // line of '('
	LastLine int   // line of ')'
	Method   Exp   //可以是普通方法，也可以是类方法
	Exps     []Exp //参数表达式列表
	// TableConstructor *TableConstructExp
	// LiteralString    *StringExp
}

/*
圆括号表达式
圆括号会对 vararg ，函数调用和表构造的语义产生影响

对于vararg：
function getValues()

	return 1, 2, 3

end

function foo(...)

	local args = {...}
	print(#args)  -- 输出参数个数

end

foo(getValues())  -- 输出: 3 (传入了3个参数)
foo( ( getValues() ) )  -- 输出: 1 (传入了1个参数)

对于函数调用：
function sum(a, b, c)

	return a + b + c

end

function getNumbers()

	return 10, 20, 30

end

print(sum(getNumbers()))  -- 输出: 60 (三个值分别传给a,b,c)
print(sum( ( getNumbers() ) ))  -- 错误！只传入了第一个值10，b和c为nil
-- 相当于 sum(10, nil, nil) → 尝试对nil做算术运算会报错

对于表的构造：
function getValues()

	return 1, 2, 3

end

local t1 = {getValues()}  -- t1 = {1, 2, 3}
local t2 = {(getValues())} -- t2 = {1}

Lua 的这种设计有以下几个目的：

灵活性：让程序员可以自由控制是否展开多返回值，需要多值时（不加括号），需要单值时（加括号）
*/
type ParensExp struct {
	Exp Exp
}
