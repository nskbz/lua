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
	LUAVALUE_THREAD
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

const LUA_MIN_STACK = 20
const LUA_MAX_STACK = 1000000
const LUA_REGISTRY_INDEX = -LUA_MAX_STACK - 1000
const LUA_GLOBALS_RIDX int64 = 2 //lua虚拟机执行入口的Load会压入lua默认main函数(index=1)，所以全局表放在index=2
