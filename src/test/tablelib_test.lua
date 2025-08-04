local printTable = function(t)
    for k, v in pairs(t) do
        print(k, v)
    end 
end

-- table.move( a1 , f , e , t , [a2] )

local t = {1, 2, 3, 4, 5}
table.move(t, 1, 3, 4)  -- 将前3个元素复制到从第4位开始的位置
-- t 变为 {1, 2, 3, 1, 2, 3, 4, 5}
printTable(t)

local src = {10, 20, 30}
local dest = {}
table.move(src, 1, 3, 1, dest)
-- dest 变为 {10, 20, 30}
printTable(dest)

local src = {'a', 'b', 'c', 'd'}
local sub = {}
table.move(src, 2, 4, 1, sub)
-- sub 变为 {'b', 'c', 'd'}
printTable(sub)

-- table.pack(...)

function process_args(...)
    local args = table.pack(...)
    print("Received " .. args.n .. " arguments")
    for i = 1, args.n do
        print(i, args[i])
    end
end

process_args("a", nil, "c")

-- table.unpack(list [, i [, j]])

local colors = {"red", "green", "blue", "yellow"}
local first, second = table.unpack(colors, 1, 2)
print(first, second)  -- 输出: red   green

function sum(a, b, c)
    return a + b + c
end

local nums = {5, 10, 15}
print(sum(table.unpack(nums)))  -- 输出: 30

-- table.concat(list [, sep] [, i [, j]])

local colors = {"red", "green", "blue", "yellow"}
print(table.concat(colors, ", ", 2, 4))  -- 输出: "green, blue, yellow"

local mixed = {1, "text", 3.14}
print(table.concat(mixed, "|"))  -- 输出: "1|text|3.14"

