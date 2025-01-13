package api

type LuaVM interface {
	LuaState
	PC() int          //返回当前PC
	AddPC(n int)      //修改PC
	Fetch() uint32    //取出当前指令并将PC指向下一条指令
	GetConst(idx int) //将指定索引的常量val推入栈顶
	GetRK(arg int)    //将ArgR或ArgK的值推入栈顶(该arg不用+1)
}
