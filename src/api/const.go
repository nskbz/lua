package api

const (
	LUAVALUE_NONE LuaValueType = iota - 1 // -1
	LUAVALUE_NIL
	LUAVALUE_BOOLEAN
	LUAVALUE_LIGHTUSERDATA
	LUAVALUE_NUMBER
	LUAVALUE_STRING
	LUAVALUE_TABLE
	LUAVALUE_FUNCTION
	LUAVALUE_USERDATA
	LUAVALUE_COROUTINE
)

const (
	ArithOp_ADD  ArithOp = iota //+
	ArithOp_SUB                 //-
	ArithOp_MUL                 //*
	ArithOp_MOD                 //%
	ArithOp_POW                 //^
	ArithOp_DIV                 // 普通除法
	ArithOp_IDIV                //向下整除
	ArithOp_AND                 //&
	ArithOp_OR                  //|
	ArithOp_XOR                 //~
	ArithOp_SHL                 //<<
	ArithOp_SHR                 //>>
	//一元运算
	ArithOp_OPPOSITE //取相反数，添负号
	ArithOp_NOT      //位取反
)

const (
	CompareOp_EQ CompareOp = iota //==
	CompareOp_LT                  //<
	CompareOp_LE                  //<=
)

/*
*	Index:
*	[-∞,-LUA_MAX_STACK - 1000) ∪ [-LUA_MAX_STACK ,LUA_MAX_STACK ]
 *	UpvalIdx ∪ AbsIdx
*/
const LUA_MIN_STACK = 20
const LUA_MAX_STACK = 1000000
const LUA_REGISTRY_INDEX = -LUA_MAX_STACK - 1000

const LUA_MAIN_COROUTING_RIDX int64 = 1
const LUA_GLOBALS_RIDX int64 = 2 //全局表默认key为2放在注册表中

// as a key; in the registry, for table of loaded modules
const LUA_LOADED_TABLE = "_LOADED"

// as a key; in the registry, for table of preloaded loaders
const LUA_PRELOAD_TABLE = "_PRELOAD"

const LUA_MULTRET = -1

/* thread status */
const (
	//LUA_OK到LUA_ERR_FILE作为协程的返回值,都属于LUA_DEAD状态
	LUA_OK = iota
	LUA_ERR_RUN
	LUA_ERR_SYNTAX
	LUA_ERR_MEM
	LUA_ERR_GCMM
	LUA_ERR_ERR
	LUA_ERR_FILE
	LUA_DEAD      //协程死亡状态,即该协程已经执行完毕
	LUA_SUSPENDED //协程挂起状态,协程初始状态或则由该协程通过yeild方法主动让出控制权
	LUA_RUNNING   //协程执行状态,即当前协程目前具有控制权
	LUA_NORMAL    //协程正常执行状态,协程被恢复,不过未运行;区别于RUNNING的点是该协程在执行过程中调用了resume方法交出了控制权
)
