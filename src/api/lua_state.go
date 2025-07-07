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

func (arith ArithOp) String() string {
	switch arith {
	case ArithOp_ADD:
		return "ADD"
	case ArithOp_SUB:
		return "SUB"
	case ArithOp_MUL:
		return "MUL"
	case ArithOp_MOD:
		return "MOD"
	case ArithOp_POW:
		return "POW"
	case ArithOp_DIV:
		return "DIV"
	case ArithOp_IDIV:
		return "Downward DIV"
	case ArithOp_AND:
		return "Bit:and"
	case ArithOp_OR:
		return "Bit:or"
	case ArithOp_XOR:
		return "Bit:xor"
	case ArithOp_SHL:
		return "Bit:left shift"
	case ArithOp_SHR:
		return "Bit:right shift"
	case ArithOp_OPPOSITE:
		return "Take the opposite number"
	case ArithOp_NOT:
		return "Bit:not"
	}
	return "Unknown"
}

func (compare CompareOp) String() string {
	switch compare {
	case CompareOp_EQ:
		return "Equal"
	case CompareOp_LT:
		return "Less than"
	case CompareOp_LE:
		return "Less or Equal"
	}
	return "Unknown"
}

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
	CheckStack(n int)        //检查是否可以容纳n个value,n>0
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
	ToBoolean(idx int) bool     //获取指定索引的bool值
	ToInteger(offset int) int64 //获取指定索引的int64值,idx为相对top的偏移
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
	GetTable(idx int) LuaValueType           //获取指定idx对应table的(stack[top])索引值类型并将其值压入栈(table=stack[idx])：table[stack[top]] ;该方法会弹出一个元素
	GetField(idx int, k string) LuaValueType //获取指定idx的table和指定字符串键的值类型并将其值压入栈：table[k]
	GetI(idx int, i int64) LuaValueType      //获取指定idx的table和指定整数键的值类型并将其值压入栈：table[i]
	SetTable(idx int)                        //设置指定idx的table的键值(table=stack[idx])：table[key]=val (val=stack[top],key=stack[top--]) ;该方法会弹出两个元素
	SetField(idx int, k string)              //设置指定idx的table和指定字符串键的值：table[k]=val (val=stack[top])
	SetI(idx int, i int64)                   //设置指定idx的table和指定整数键的值：table[i]=val (val=stack[top])
	/*
	*	函数调用
	 */
	Load(chunk []byte, chunckName, mode string) int //加载chunk获得对应的closure并将其压入栈
	//lua函数：将nArgs+1数量的val弹出作为函数及其参数，执行closure，最后将nResults数量的结果值压入栈(nResults<0则压入所有返回值)
	//go函数：将nArgs数量的val弹出作为外部Go函数的参数，执行Go函数并将所有返回值都压入栈中
	//nArgs为参数个数
	//nResults为返回值个数:==0则不返回任何值,<0则返回值全部压入,>0则返回nResults个返回值
	Call(nArgs, nResults int)
	/*
	*	Go函数支持
	 */
	PushGoFunction(gf GoFunc, n int) //弹出n个val作为gofunc的Upvalue(捕获变量)，然后压入gofunc
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
	CloseUpvalues(a int)    //取消对>=局部变量R[a-1]的Upvalue引用
	/*
	*	元编程支持
	 */
	GetMetaTable(idx int) bool             //如果指定idx值具备元表则将其压入栈并返回true;若没有元表则直接返回false
	SetMetaTable(idx int)                  //将栈顶的一个值弹出作为指定idx值的元表
	RawLen(idx int) uint                   //(不使用元方法)返回对应idx的val值的长度
	RawEqual(idx1, idx2 int) bool          //(不使用元方法)比较idx1和idx2是否相等
	RawGet(idx int) LuaValueType           //(不使用元方法)获取指定idx的table的stack[top]索引值类型并将其值压入栈(table=stack[idx])：table[stack[top]] ;该方法会弹出一个元素
	RawSet(idx int)                        //(不使用元方法)设置指定idx的table的键值(table=stack[idx])：table[key]=val (val=stack[top],key=stack[top--]) ;该方法会弹出两个元素
	RawGetI(idx int, i int64) LuaValueType //(不使用元方法)获取指定idx的table和指定整数键的值类型并将其值压入栈：table[i]
	RawSetI(idx int, i int64)              //(不使用元方法)设置指定idx的table和指定整数键的值：table[i]=val (val=stack[top])
	/*
	*	通用for循环支持
	 */
	//true：将指定idx对应table的key=stack[top]的下一个nextkey和其value压入栈顶(nextkey=top-1;value=top)
	//false：nextkey==nil,循环结束
	Next(idx int) bool
	/*
	*	异常处理支持
	 */
	Error() int //弹出栈顶值作为错误抛出
	PCall(nArgs, nResults, msgh int) int
}
