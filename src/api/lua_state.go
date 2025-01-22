package api

type LuaValueType int //Lua的数据类型

func (tp LuaValueType) String() string {
	switch tp {
	case LUAVALUE_NONE:
		return "none"
	case LUAVALUE_NIL:
		return "nil"
	case LUAVALUE_BOOLEAN:
		return "boolean"
	case LUAVALUE_LIGHTUSERDATA:
		return "Luserdata"
	case LUAVALUE_NUMBER:
		return "number"
	case LUAVALUE_STRING:
		return "string"
	case LUAVALUE_TABLE:
		return "table"
	case LUAVALUE_FUNCTION:
		return "function"
	case LUAVALUE_USERDATA:
		return "userdata"
	case LUAVALUE_THREAD:
		return "thread"
	}
	return "unknown"
}

type ArithOp int   //算术运算与位运算
type CompareOp int //比较运算

// Go函数;return返回值的个数
type GoFunc func(LuaVM) int

/*		Stack					Index
*		|			nil			 |	  7 		  -							  -
*		|			nil			 |	  6			 |	无效索引		|
*		|			nil			 |	  5			-							|
* top|	LuaValue	|		4		  -							|可接受索引
*		|	LuaValue	|		3		 |	有效			  |
*		|	LuaValue	|		2		|	索引			|
*		|	LuaValue	|		1	   -					   -
 *
 *		LuaState主要是基于luaStack上的功能封装
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
	ToBoolean(idx int) bool  //获取指定索引的bool值
	ToInteger(idx int) int64 //获取指定索引的int64值
	ToIntegerX(idx int) (int64, bool)
	ToFloat(idx int) float64 //获取指定索引的float64值
	ToFloatX(idx int) (float64, bool)
	ToString(idx int) string //获取指定索引的string值
	ToStringX(idx int) (string, bool)
	/*
	*	运算操作
	 */
	Arith(op ArithOp)                          //进行算术运算
	Compare(idx1, idx2 int, op CompareOp) bool //对两索引的val进行比较，不改变栈结构
	Len(idx int)                               //将指定索引的val的长度压入栈顶
	Concat(n int)                              //从栈顶弹出n个val进行字符串拼接，结果压入栈
	/*
	*	表相关操作
	 */
	NewTable()
	CreateTable(nArr, nRec int)              //创建table并将其压入栈顶
	GetTable(idx int) LuaValueType           //获取指定idx的table的stack[top]索引值类型并将其值压入栈(table=stack[idx])：table[stack[top]] ;该方法会弹出一个元素
	GetField(idx int, k string) LuaValueType //获取指定idx的table和指定字符串键的值类型并将其值压入栈：table[k]
	GetI(idx int, i int64) LuaValueType      //获取指定idx的table和指定整数键的值类型并将其值压入栈：table[i]
	SetTable(idx int)                        //设置指定idx的table的键值(table=stack[idx])：table[key]=val (val=stack[top],key=stack[top--]) ;该方法会弹出两个元素
	SetField(idx int, k string)              //设置指定idx的table和指定字符串键的值：table[k]=val (val=stack[top])
	SetI(idx int, i int64)                   //设置指定idx的table和指定整数键的值：table[i]=val (val=stack[top])
	/*
	*	函数调用
	 */
	Load(chunk []byte, chunckName, mode string) int //加载chunk获得对应的closure并将其压入栈
	//lua函数：将nArgs+1数量的val弹出作为函数及其参数，执行closure，最后将nResults数量的结果值压入栈
	//go函数：将nArgs数量的val弹出作为外部Go函数的参数，执行Go函数并将所有返回值都压入栈中
	Call(nArgs, nResults int)
	/*
	*	Go函数支持
	 */
	PushGoFunction(gf GoFunc, n int)
	IsGoFunction(idx int) bool
	ToGoFunction(idx int) GoFunc
	/*
	*	全局环境
	 */
	PushGlobalTable()                  //将全局环境压入栈顶
	GetGlobal(key string) LuaValueType //获取key=name的全局环境
	SetGlobal(key string)              //设置全局环境key=val(val=stack[top])
	Register(key string, gf GoFunc)    //注册外部Go函数
	/*
	*	Upvalue支持
	 */
	UpvalueIndex(i int) int //获取Upvalue索引
	CloseUpvalues(a int)    //取消对>=Upvalue[a-1]的引用
}
