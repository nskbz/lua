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
type UnitaryArithExp struct {
	Line int
	Op   int //运算符标号
	A    Exp
} //单目运算符，只有一个表达式(操作数)A

type DualArithExp struct {
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
	Line     int // line of '{',for what?
	LastLine int // line of  '}',for what?
	Key      []Exp
	Val      []Exp
}

// 函数构造表达式
// functiondef ::= function funcbody
// funcbody ::= ‘(’ [parlist] ‘)’ block end
// parlist ::= namelist [‘,’ ‘...’] | ‘...’
// namelist ::= Name {‘,’ Name}
type FuncConstructExp struct {
	Line     int // line of  'function'
	LastLine int // line of 'end'
	ParList  []string
	//Vararg   *VarargExp
	IsVararg bool
	Block    *Block
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
type ParensExp struct {
	Exp Exp
}

// 表访问表达式
type TableAccessExp struct {
	LastLine  int // line of ']'
	PrefixExp Exp
	KeyExp    Exp
}

// 函数调用表达式
// functioncall ::=prefixexp [ ':'  Name ]  args
// args ::='('  [explist]  ')'  |  tableconstructor  | LiteralString
type FuncCallExp struct {
	Line      int // line of '('
	LastLine  int // line of ')'
	PrefixExp Exp
	NameExp   *StringExp
	Args      []Exp
	// TableConstructor *TableConstructExp
	// LiteralString    *StringExp
}
