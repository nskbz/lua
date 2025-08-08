package api

type FuncReg map[string]GoFunc

/*
* auxiliary interface
 */
type AuxLib interface {
	/* Error-report functions */
	Error2(fmt string, a ...interface{}) int //压入fmt格式串错误信息并触发panic,方法不会返回
	ArgError(idx int, extraMsg string) int   //方法调用时参数错误,方法不会返回
	/* Argument check functions */
	ArgCheck(cond bool, idx int, extraMsg string) //检查cond是否为true。如果不是则抛出带有标准消息的错误
	CheckAny(idx int)                             //检查R(idx)是否为LuaValueType
	CheckType(idx int, t LuaValueType)            //检查R(idx)是否为指定的LuaValueType
	CheckInteger(idx int) int64                   //检查R(idx)是否能转换为整型,能转换则返回转换的整型,不能则ArgError
	CheckFloat(idx int) float64                   //检查R(idx)是否能转换为浮点型,能转换则返回转换的浮点型,不能则ArgError
	CheckString(idx int) string                   //检查R(idx)是否能转换为字符串,能转换则返回转换的字符串,不能则ArgError
	OptInteger(idx int, d int64) int64            //如果R(idx)是整数（或可转换为整数），则返回该整数。如果此参数不存在或为nil，则返回d。否则，引发错误。
	OptFloat(idx int, d float64) float64          //如果R(idx)是整数（或可转换为浮点数），则返回该浮点数。如果此参数不存在或为nil，则返回d。否则，引发错误。
	OptString(idx int, d string) string           //如果R(idx)是字符串（或可转换为整数），则返回该字符串。如果此参数不存在或为nil，则返回d。否则，引发错误。
	/* Load functions */
	DoFile(filename string) bool         //加载并运行给定的文件。如果没有错误则返回false；反之返回true
	LoadFile(filename string) int        //实质为LoadFileX但提供空mode
	LoadFileX(filename, mode string) int //将文件加载为Lua块。如果filename为NULL，则从标准输入加载。如果文件中的第一行以#开头，则忽略它。
	DoString(str string) bool            //加载并运行给定的字符串。如果没有错误则返回false；反之返回true
	LoadString(s string) int             //成功返回LUA_OK并压入对应的closure
	/* Other functions */
	TypeName2(idx int) string                    //获取指定索引的类型名称
	ToString2(idx int) string                    //将给定索引的LuaValue转换字符串型。结果字符串压入堆栈，并由函数返回。如果该LuaValue存在元方法"__tostring",则应调用元方法
	Len2(idx int) int64                          //获取指定索引的LuaValue的长度
	GetSubTable(idx int, fname string) bool      //确保获取表中的表元素。table=R(idx) and type(table[fname])==table。如果fname对应的键值是table则返回true并将其压入栈，反之类型不是table则返回false并创建table
	GetMetafield(obj int, e string) LuaValueType //将索引为obj的对象的元表中的字段e压入堆栈，并返回压入值的类型。如果对象没有元表，或者元表没有此字段，则不推送任何内容并返回LUA_TNIL。
	CallMeta(obj int, e string) bool             //调用元方法;如果索引obj处的对象有元表，并且这个元表有字段e，则此函数调用该字段，并将该对象作为其唯一参数。在本例中，该函数返回true并将调用返回的值压入堆栈。如果没有元表或元方法，则此函数返回false（不向堆栈上压入任何值）。

	OpenLibs()
	//确保modname模块加载
	//如果modname不存在于包中package.loaded,则以字符串modname作为参数调用函数openf，并在包中package.loaded设置调用结果
	//如果glb为true，还将模块存储到全局modname中。
	//在堆栈上留下模块的副本。
	RequireF(modname string, openf GoFunc, glb bool)

	NewLib(funcs map[string]GoFunc) //通过funcs创建一个模块(table),并将其压入栈顶
}
