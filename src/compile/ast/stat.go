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

type Empty struct{} // ';'

// stat ::=label
type LabelStat struct {
	Name string //标签的值
}

// stat ::=break
type BreakStat struct {
	Line int //break语句跳转的行数
}

// stat ::=goto
type GotoStat struct {
	Name string //goto语句跳转的标记
}

// stat ::= do block end
type DoStat struct {
	Block *Block //block 代码体
}

// stat ::= while exp do block end
type WhileStat struct {
	Exp   Exp
	Block *Block
}

// stat ::=repeat block until exp
type RepeatStat struct {
	Exp   Exp
	Block *Block
}

// stat ::=if exp then block {elseif exp then block} [else block] end
// if exp then block == elseif exp then block
// else block == elseif (true) then block
type IfStat struct {
	Exp   []Exp
	Block []*Block
	//Exp与Block一一对应，索引为0表示if语句，其他表示elseif语句
}

// stat ::=for Name ‘=’ exp ‘,’ exp [‘,’ exp] do block end
//
//	for example:
//	for name = start, end, step do
//	-- 循环体
//	end
type ForNumStat struct {
	LineOfFor int //for标准语句开始的行号，用于?
	LineOfDo  int //for标准语句循环体开始的行号，用于?
	Name      string
	Start     Exp    //start可以是一个表达式包括字面量
	End       Exp    //同上
	Step      Exp    //可选，默认为1步长
	Block     *Block //循环体
}

// stat ::=for namelist in explist do block end
// namelist ::=Name { ',' Name}
// explist ::=exp { ',' exp}
type ForInStat struct {
	LineOfDo int      //for in语句循环体开始的行号，用于?
	NameList []string //namelist
	ExpList  []Exp    //explist
	Block    *Block
}

// stat ::=local namelist [‘=’ explist]
// namelist ::=Name { ',' Name}
// explist ::=exp { ',' exp}
type LocalVarStat struct {
	LastLine int      //局部变量语句末尾行号，用于?
	NameList []string //namelist
	ExpList  []Exp    //explist
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
type AssignStat struct {
	LastLine int   //赋值语句末尾行号，用于?
	VarList  []Exp //varlist
	ExpList  []Exp //explist
}
