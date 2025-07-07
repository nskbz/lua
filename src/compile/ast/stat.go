package ast

/*
stat ::=‘;’ |
varlist ‘=’ explist |
functioncall |
label |
break |
goto Name |
do block end |
while exp do block end |
repeat block until exp |
if exp then block {elseif exp then block} [else block] end |
for Name ‘=’ exp ‘,’ exp [‘,’ exp] do block end |
for namelist in explist do block end |
function funcname funcbody |
local function Name funcbody |
local namelist [‘=’ explist]
*/
type Stat interface{}

type EmptyStat struct{} // ';'

// stat ::=break
type BreakStat struct {
	Line int //break语句所在行数,用于debug
}

// stat ::=label
//
// for example:
// local a = 1
// ::label:: print("--- goto label ---")
// a = a+1
// if a < 3 then
// goto label
// end
type LabelStat struct {
	Name string //标签的值
}

// stat ::=goto
//
// for example:
// local b = true
// do
//
//	do
//	  print("作用域2")
//	  do
//	    print("作用域3")
//	    do
//	      print("作用域4")
//	      if b == true then
//	        goto myend
//	      end
//	      goto mylabel
//	      print("?")
//	    end
//	  end
//	end
//	::mylabel::
//	print("作用域1")
//
// end
// -- 下面不行,虽然下面也是作用域1但不是同一个(上面),所以会找不到mylabel标签
// -- 由此可知在退出作用域时会释放当前作用域下的所有标签
// -- do
// --   goto mylabel
// -- end
// ::myend::
// print("作用域0")
type GotoStat struct {
	Name string //goto语句跳转的标记

	Line int //用于debug,记录生成指令对应的行数
}

// stat ::= do block end
type DoStat struct {
	Block *Block //block 代码体
}

// stat ::= while exp do block end
type WhileStat struct {
	Exp   Exp
	Block *Block

	ExpLine int //用于debug
}

// stat ::=repeat block until exp
//
// for example:
// a = 10
// repeat
//
//	print("a的值为:", a)
//	a = a + 1
//
// until( a > 15 )
type RepeatStat struct {
	Exp   Exp
	Block *Block

	ExpLine int //repeat循环条件表达式所在的行数
}

// stat ::=if exp then block {elseif exp then block} [else block] end
// if exp then block == elseif exp then block
// else block == elseif (true) then block
type IfStat struct {
	Exps   []Exp
	Blocks []*Block
	//Exp与Block一一对应，索引为0表示if语句，其他表示elseif语句
}

// stat ::=for Name ‘=’ exp ‘,’ exp [‘,’ exp] do block end
//
//	for example:
//	for name = start, end, step do
//	-- 循环体
//	end
//
// for i = 10, 1, -2 do
// print(i)
// end
type ForNumStat struct {
	Name string
	//Init Limit Step可以是IntExp或FloatExp
	//如果Step是浮点型，则会将Init转换为浮点型，反之亦然
	Init  Exp    //Init可以是一个表达式包括字面量，初始化表达式
	Limit Exp    //同上，条件表达式或叫限制表达式
	Step  Exp    //可选，默认为1步长
	Block *Block //循环体

	LineOfFor int //for标准语句开始的行号，用于debug生成localvar作用域
	LineOfDo  int //for标准语句循环体开始的行号，用于debug
	LineOfEnd int
}

// stat ::=for namelist in explist do block end
// namelist ::=Name { ',' Name}
// explist ::=exp { ',' exp} //迭代器函数, 状态值, 初始控制变量
//
// for 变量列表 in 迭代器函数, 状态值, 初始控制变量 do block end
//
// for example:
// local colors = {"red", "green", "blue"}
// for idx, color in myIterator(colors) do
//
//	print(idx, color)
//
// end
// -- 输出：
// -- 1   red
// -- 2   green
// -- 3   blue
type ForInStat struct {
	NameList []string //namelist, 用于接受迭代器函数的返回值
	ExpList  []Exp    //迭代器函数, 状态值, 初始控制变量
	Block    *Block

	LineOfFor int
	LineOfDo  int //for in语句循环体开始的行号，用于debug记录作用域
	LineOfEnd int //同上
}

// stat ::=local namelist [‘=’ explist]
// namelist ::=Name { ',' Name}
// explist ::=exp { ',' exp}
//
// for example:
// local a
// local b, c, d = 1, 2, 3
//
// "区别与AssignStat作用于赋值，LocalVarStat则作用于定义"
type LocalVarStat struct {
	LastLine     int      //局部变量语句末尾行号，用于debug记录局部变量的作用范围行数
	LocalVarList []string //localVar,区别var,localVar不存在OOP的那种层级
	ExpList      []Exp    //explist
}

// stat ::=local function funcname funcbody
// funcname ::= Name --local定义方法的方法名就是单一标识符
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
type LocalFuncDefStat struct {
	Name string //可以为空，即匿名函数
	Body *FuncDefExp

	DefLine int //记录定义的函数，用于debug记录方法作用域
}

// 面向对象的方法定义
// stat ::= function funcname funcbody
// funcname ::= Name {'.' Name} [':' Name] --面向对象的方法定义伴随着'.'和':'
// for example:
// local M = {}
// function M.MyFunc(self, a)
// -- body
// end
type OopFuncDefStat struct {
	Name Exp
	Body *FuncDefExp

	DefLine int
}

// stat ::=varlist ‘=’ explist
// varlist ::=var { ',' var}
// var ::=Name | prefixexp '[' exp ']' | prefixexp '.' Name
// 变量考虑三种情况
// Name为普通的变量名
// prefixexp '[' exp ']' 为数组某下标变量的表示
// prefixexp '.' Name 为lua对象中的变量表示
//
// explist ::=exp { ',' exp}
//
// "区别与LocalVarStat作用于定义，AssignStat则作用于赋值"
//
// for example:
// local t = { x=1,y=2 }
// local t_copy=t
// print("x="..t.x.." y="..t.y)	--output:"x=1 y=2"
// t_copy.x=2
// t_copy.y=1
// print("x="..t.x.." y="..t.y) --output:"x=2 y=1"
type AssignStat struct {
	VarList []Exp //varlist
	ExpList []Exp //explist

	LastLine int //赋值语句末尾行号，用于debug
}

// FuncCallStat:
//
//	fun()	;语句只会以调用形式单独出现,不需要返回值
//
// LocalVarStat：
//
//	local a,b=1,fun()
//
// AssignStat:
//
//	table.a=fun()	;表tablea为之前已经定义声明了，该句为语法糖实质为table[a]=fun()
type FuncCallStat = FuncCallExp
