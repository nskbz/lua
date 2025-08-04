-- select (index, ···)

function testSelect(...)
    print("参数总数:", select('#', ...))  -- 获取参数个数
    
    -- 获取从第2个参数开始的所有参数
    local arg2_onwards = select(2, ...)
    print("从第2个参数开始:", arg2_onwards)
    
    -- 遍历所有参数
    for i = 1, select('#', ...) do
        print("参数"..i..":", select(i, ...))
    end
end

testSelect("a", "b", "c", "d")

-- load (chunk [, chunkname [, mode [, env]]])

local func=load("print(\"hello world\")")

func()

local lines = {
    "local a = 10",
    "local b = 20",
    "return a + b"
}

local i = 0
local func = load(function()
    i = i + 1
    return lines[i]
end, "数组代码示例")

print(func())  -- 输出: 30

local pow = function(a)
    return a*a
end

local math_env = {
    x = 2.5,
    pow = pow,
    show = print,
}

local expression = "show(\"x: \",x) return pow(x)"
local func = load(expression, "表达式", "t", math_env)
print("pow结果:", func()) --输出: 6.25

-- loadfile ([filename [, mode [, env]]])

print("enter lua txt:")
local func ,err = loadfile()
if err ~= nil then
    print(err)
else
    func()
end

local _env = {
    show = print
}
local func = loadfile("../test2.lua","bt",_env) --注意加载路径是否正确
func()

-- dofile ([filename])

print("enter lua txt:")
local func ,err = dofile()
if err ~= nil then
    print(err)
else
    func()
end

-- xpcall (func, errhandler [, arg1, ···])

function dangerous_function(a, b)
    if b == 0 then
        error("Division by zero!")
    end
    return a / b
end

function error_handler(err)
    print("发生错误:", err)
    -- 可以返回一个默认值或进行其他处理
    return -1  
end

local status, result = xpcall(dangerous_function, error_handler, 10, 0)
if status then
    print("结果:", result)
else
    print("错误处理结果:", result)  -- 输出 error_handler 的返回值
end

-- rawequal (v1, v2)

local t1 = {x = 10}
local t2 = {x = 10}
-- 普通比较
print(t1 == t2)  -- false（比较引用）
-- rawequal 比较
print(rawequal(t1, t2))  -- false（同样比较引用）
-- 设置 __eq 元方法
local mt = {
    __eq = function(a, b) return a.x == b.x end
}
setmetatable(t1, mt)
setmetatable(t2, mt)
print(t1 == t2)         -- true（调用元方法）
print(rawequal(t1, t2)) -- false（忽略元方法）

-- rawlen (v)

local t = {1, 2, 3}
-- 设置 __len 元方法
setmetatable(t, {
    __len = function() return 100 end
})
print(#t)          -- 输出 100（调用元方法）
print(rawlen(t))   -- 输出 3（忽略元方法）

-- rawget(table,index)

local t = {}
-- 设置 __index 元方法
setmetatable(t, {
    __index = function(t, k)
        return "Default " .. k
    end
})
print(t.name)        -- 输出 "Default name"（调用元方法）
print(rawget(t, "name"))  -- 输出 nil（忽略元方法）

-- rawset (table, index, value)

local t = {}
-- 设置 __newindex 元方法
setmetatable(t, {
    __newindex = function(t, k, v)
        error("不允许直接修改表!")
    end
})
-- 普通赋值会触发错误
t.name = "Lua"  -- 会抛出错误
-- 使用 rawset 可以绕过限制
rawset(t, "name", "Lua")  -- 成功设置
print(t.name)  -- 输出 "Lua"

-- type (v)

print(type(nil))            --> "nil"
print(type(true))           --> "boolean"
print(type(42))             --> "number"
print(type(3.14))           --> "number"
print(type("hello"))        --> "string"
print(type({}))             --> "table"
print(type(print))          --> "function"

-- tostring(v)

print(tostring(nil))        --> nil
print(tostring(true))       --> true
print(tostring(42.31))         --> 42.31
print(tostring("hello"))    --> hello
local t = {name = "Lua", version = 5.4}
print(tostring(t))  --> table: 0x7f8e5bc12340 (默认输出)
-- 自定义表的字符串表示
setmetatable(t, {
    __tostring = function(self)
        return "Lua" .. self.name .. " (version " .. self.version .. ")"
    end
})
print(tostring(t))  --> LuaLua (version 5.4)

-- tonumber (e [, base])

-- 默认10进制转换
print(tonumber("42"))         --> 42
print(tonumber("3.14"))       --> 3.14
print(tonumber(" 100  "))     --> 100（自动忽略前后空格）
print(tonumber("1e3"))        --> 1000（科学计数法）
-- 其他进制转换
print(tonumber("FF", 16))     --> 255（16进制）
print(tonumber("1010", 2))    --> 10（2进制）
print(tonumber("755", 8))     --> 493（8进制）
-- 转换失败
print(tonumber("abc"))        --> nil
print(tonumber("123.4.5"))    --> nil