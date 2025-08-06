-- package.searchpath(name, path [, path_mark [, exec_dir]])

-- 测试搜索
local test_path = "../?.lua;./?.lua;test/?.lua"
local path, err = package.searchpath("oop_test", test_path)
if path then
    print("找到模块路径:", path)  -- 应输出: test/oop_test.lua
else
    print("搜索失败:", err)
end

-- 使用 @ 代替默认的 ?
local path = package.searchpath("block", "./@.lua;compile/ast/@.go", "@")
print("自定义标记路径:", path)  -- 应输出: compile/ast/block.go

-- 测试下划线转换（使用!标记）
local path = package.searchpath("compile_ast", "./?/block.go", "?", "!")
print("下划线转换路径:", path)  -- 应输出: ./compile/ast/block.go