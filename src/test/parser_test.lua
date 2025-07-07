-- 单行注释
--[[多行
注释]]

-- 变量与数据类型
local num = 42
local str = "hello"
local bool = true
local tbl = {1, "two", x=3}
local func = function() end

-- 运算符
local a, b = 10, 3
print(a + b, a // b, a ^ b, #str)  -- 算术/长度
print(a == b, a ~= b, bool and false or true)  -- 关系/逻辑

-- 控制结构
if a > b then
    print("if")
elseif a == b then
    print("elseif")
else
    print("else")
end

while a > 0 do
    a = a - 1
    if a % 2 == 0 then break end
end

repeat
    b = b - 1
until b == 0

for i = 1, 3 do
    print(i)
end

for k, v in pairs(tbl) do
    print(k, v)
end

-- 函数
local function add(x, y)
    return x + y, "extra"  -- 多返回值
end
print(add(2, 3))

local function f()


  local a, b = 1, 2; print(a, b)   -->	1	2
  local a, b = 3, 4; print(a, b)   -->	3	4
  do
    print(a, b)                    -->	3	4
    local a, b = 5, 6; print(a, b) -->	5	6
  end
  print(a, b)                      -->	3	4

  return function(a)
    print(a)
  end
end

f()("jack")

-- 表操作
tbl[1] = "one"
print(tbl.x, tbl[1])

-- 元表（简化）
local mt = {__add = function(x,y) return x.val + y.val end}
local x = {val=10}; setmetatable(x, mt)
local y = {val=20}; setmetatable(y, mt)
print(x + y)

-- 协程（简化）
local co = coroutine.create(function()
    coroutine.yield("pause")
    return "done"
end)
print(coroutine.resume(co))
print(coroutine.resume(co))

-- 标准库示例
print(string.upper("lua"), math.pi, table.concat({"a","b"}, ","))

-- 标签与循环控制
local b = false
do
  do
    print("作用域2")
    do
      print("作用域3")
      do
        print("作用域4")
        if b == true then
          goto myend  
        end
        goto mylabel
        print("?")
      end
    end
  end
  ::mylabel::
  print("作用域1")  
end

do
  do
    goto mylabel
  end
  ::mylabel::
  print("2.作用域1")
end

::myend::
print("作用域0")

-- 可变参数测试
do
    local function test(...)
        print("--- Vararg Test ---")
        print("Raw ...:", ...)
        print("Count:", select('#', ...))
        print("As table:", table.concat({...}, "|"))
        print("Packed nils:", table.pack(...)[2])
    end

    test("A", nil, "C", 42)
end

local function test_concat()
    -- 基础测试
    assert("1" .. "2" == "12")
    assert("A" .. 1 == "A1")
    
    -- 边界测试
    assert("" .. "B" == "B")
    -- assert("X" .. nil)  -- 应报错
    
    -- 元表测试
    local mt = {__tostring = function() return "META" end}
    local obj = setmetatable({}, mt)
    assert("OBJ:" .. obj == "OBJ:META")
    
    print("All concatenation tests passed!")
end

test_concat()