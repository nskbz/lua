package api

type LuaValueType int //Lua的数据类型
type ArithOp int      //算术运算与位运算
type CompareOp int    //比较运算

/*		Stack					Index
*		|			nil			 |	  7 		  -							  -
*		|			nil			 |	  6			 |	无效索引		|
*		|			nil			 |	  5			-							|
* top|	LuaValue	|		4		  -							|可接受索引
*		|	LuaValue	|		3		 |	有效			  |
*		|	LuaValue	|		2		|	索引			|
*		|	LuaValue	|		1	   -					   -
 */

type LuaState interface {
	/*
	*	基础栈操作
	 */
	GetTop() int             //获取top index
	SetTop(idx int)          //设置top index并将其后的出栈
	AbsIndex(idx int) int    //获取绝对index，返回的合法index应当大于0
	CheckStack(n int) bool   //检查是否可以容纳n个value,n>0
	Pop(n int)               //弹出n个val
	Copy(fromIdx, toIdx int) //将from的val复制到to
	PushValue(idx int)       //将指定索引的val压入栈顶
	Replace(idx int)         //将栈顶val弹出并替换指定索引位置的val
	Rotate(idx, n int)       //将[idx,top]的val进行轮转；n>0向栈顶轮转，n<0向栈底轮转
	Insert(idx int)          //将栈顶val弹出并插入到指定索引，整个过程栈长度不变
	Remove(idx int)          //删除指定索引并将后续的val依次顺移填充，栈长度-1
	/*
	*	进栈操作
	 */
	PushNil()
	PushBoolean(b bool)
	PushInteger(n int64)
	PushNumber(n float64)
	PushString(s string)
	/*
	*	栈元素访问
	 */
	TypeName(tp LuaValueType) string //获取对应luaval的名称
	Type(idx int) LuaValueType       //返回对应索引的val
	IsNone(idx int) bool
	IsNil(idx int) bool
	IsNoneOrNil(idx int) bool
	IsBoolean(idx int) bool
	IsInteger(idx int) bool
	IsFloat(idx int) bool
	IsString(idx int) bool
	IsTable(idx int) bool
	IsThread(idx int) bool
	IsFunction(idx int) bool
	ToBoolean(idx int) bool
	ToInteger(idx int) int64
	ToIntegerX(idx int) (int64, bool)
	ToFloat(idx int) float64
	ToFloatX(idx int) (float64, bool)
	ToString(idx int) string
	ToStringX(idx int) (string, bool)
	/*
	*运算操作
	 */
	Arith(op ArithOp)                          //进行算术运算
	Compare(idx1, idx2 int, op CompareOp) bool //对两索引的val进行比较，不改变栈结构
	Len(idx int)                               //将指定索引的val的长度压入栈顶
	Concat(n int)                              //从栈顶弹出n个val进行字符串拼接，结果压入栈
	/*
	*表相关操作
	 */
	NewTable()
	CreateTable(nArr, nRec int)              //创建table并将其压入栈顶
	GetTable(idx int) LuaValueType           //获取指定索引table的值类型并将其值压入栈(table=stack[idx])：table[stack[top]]
	GetField(idx int, k string) LuaValueType //获取指定索引table和指定字符串键的值类型并将其值压入栈：table[k]
	GetI(idx int, i int64) LuaValueType      //获取指定索引table和指定整数键的值类型并将其值压入栈：table[i]
	SetTable(idx int)                        //设置指定索引table的键值(table=stack[idx])：table[key]=val (val=stack[top],key=stack[top--])
	SetField(idx int, k string)              //设置指定索引table和指定字符串键的值：table[k]=val (val=stack[top])
	SetI(idx int, i int64)                   //设置指定索引table和指定整数键的值：table[i]=val (val=stack[top])
}
