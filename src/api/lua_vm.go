package api

type LuaVM interface {
	LuaState
	PC() int          //返回当前PC
	AddPC(n int)      //修改PC
	Fetch() uint32    //取出当前指令并将PC指向下一条指令
	GetConst(idx int) //将常量表中指定索引的常量压入栈顶
	GetRK(rk int)     //将ArgR或ArgK的值推入栈顶(该arg不用+1)

	LoadProto(idx int)  //将函数原型表中指定索引的原型压入栈顶
	RegisterCount() int //返回函数所操作寄存器的数量
	LoadVarargs(n int)  //将n个vararg压入栈顶，n<0表示将全部varargs压入栈顶
}
