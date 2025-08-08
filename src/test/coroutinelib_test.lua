-- 作者测试用例

-- function permgen (a, n)
--   n = n or #a          -- default for 'n' is size of 'a'
--   if n <= 1 then       -- nothing to change?
--     coroutine.yield(a)
--   else
--     for i = 1, n do
--       -- put i-th element as the last one
--       a[n], a[i] = a[i], a[n]
--       -- generate all permutations of the other elements
--       permgen(a, n - 1)
--       -- restore i-th element
--       a[n], a[i] = a[i], a[n]
--     end
--   end
-- end

-- function permutations (a)
--   local co = coroutine.create(function () permgen(a) end)
--   return function ()   -- iterator
--     local code, res = coroutine.resume(co)
--     return res
--   end
-- end

-- for p in permutations({"a", "b", "c"}) do
--   print(table.concat(p, ","))
-- end

-- 测试用例

local coA = coroutine.create(function()
    print("协程A 开始")
    local val1 = coroutine.yield(999)
    print("协程A 的状态:", coroutine.status(coA))
    local val2,val3 = coroutine.yield(-1)
    return val2/val3
end)

local ret , n = coroutine.resume(coA)
print(ret,n)

print("协程A 的状态:", coroutine.status(coA))

local ret,n = coroutine.resume(coA,"jack")
print(ret,n)

print("协程A 的状态:", coroutine.status(coA))

local ret,n = coroutine.resume(coA,10,0)
print(ret,n)

print("协程A 的状态:", coroutine.status(coA))

-- coroutine.wrap(f)

local co_func = coroutine.wrap(function(a, b)
    print("第一次调用:", a, b)
    local x = coroutine.yield(a + b)
    print("第二次调用:", x)
    return "结束"
end)

local sum = co_func(10, 20)  -- 输出: 第一次调用: 10 20
print(sum)                   -- 输出: 30
local result = co_func(100)  -- 输出: 第二次调用: 100
print(result)                -- 输出: 结束

-- 错误处理示例
local faulty = coroutine.wrap(function()
    error("出错了!")
end)

local ok, err = pcall(faulty) --以保护模式调用
print(ok, err)  -- 输出: false  出错了!

-- faulty() --直接调用则会panic