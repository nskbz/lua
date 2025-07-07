package instruction

import (
	"nskbz.cn/lua/api"
)

/*
	FORPREP和FORLOOP指令协同工作共同实现lua中数值for的功能
	FORPREP负责初始化循环参数，在循环开始时执行一次 | R(A)-=R(A+2); pc+=sBx
	FORLOOP负责执行循环迭代，每次循环结束后执行 | R(A)+=R(A+2); if R(A) <?= R(A+1) then { pc+=sBx; R(A+3)=R(A) }

	FORPREP → [循环体] → FORLOOP → [循环体] → FORLOOP → ... → 退出循环
	forexample：

	for i = 1, 10, 2 do
    	print(i)
	end

	1. FORPREP  (初始化 i=1, limit=10, step=2)
	2. [循环体开始]
	3.   GETGLOBAL "print"
	4.   MOVE      (i的值)
	5.   CALL
	6. FORLOOP  (i += 2; 检查 i <= 10; 决定跳转或退出)

    1       [1]     LOADK           0 -1    ; 1  加载R(A)
    2       [1]     LOADK           1 -2    ; 10 加载R(A+1)
    3       [1]     LOADK           2 -3    ; 2	 加载R(A+2)
    4       [1]     FORPREP         0 3     ; to 8
    5       [2]     GETTABUP        4 0 -4  ; _ENV "print"
    6       [2]     MOVE            5 3		; 这里move的是R(A+3),用户不能直接更改循环控制变量R(A)~R(A+2),只能读R(A+3)
    7       [2]     CALL            4 2 1	; 这里调用print(i)
    8       [1]     FORLOOP         0 -4    ; to 5
    9       [3]     RETURN          0 1
*/

// R(A)-=R(A+2); pc+=sBx
func forPrep(i Instruction, vm api.LuaVM) {
	a, sbx := i.AsBx()
	a += 1
	vm.PushValue(a)
	vm.PushValue(a + 2)
	vm.Arith(api.ArithOp_SUB)
	vm.Replace(a)
	vm.AddPC(sbx)
}

// R(A)+=R(A+2); if R(A) <?= R(A+1) then { pc+=sBx; R(A+3)=R(A) }
// R(A) → 循环变量 var（初始值为 init,,,,存储循环变量的“工作值”，由虚拟机直接更新（+= step），但用户代码不直接访问它
// R(A+1) → 结束值 limit
// R(A+2) → 步长 step
// R(A+3) → 循环体内可访问的 var 的副本 ,,,,存储循环变量的“暴露值”，供用户代码（循环体）读取，每次迭代后同步更新
// PS：
// 为什么 Lua 不直接让用户代码读取 R(A)？
// 安全性：防止用户代码意外修改 R(A)（破坏循环控制逻辑）。
// 清晰性：分离“控制变量”和“用户变量”职责，符合虚拟机设计原则。
func forLoop(i Instruction, vm api.LuaVM) {
	a, sbx := i.AsBx()
	a += 1
	vm.PushValue(a + 2)
	vm.PushValue(a)
	vm.Arith(api.ArithOp_ADD)
	vm.Replace(a)
	idx1, idx2 := 0, 0
	if vm.ToFloat(a+2) < 0 {
		//step<0
		idx1 = a + 1
		idx2 = a
	} else {
		//step>=0
		idx1 = a
		idx2 = a + 1
	}
	if !vm.Compare(idx1, idx2, api.CompareOp_LE) {
		return
	}
	vm.Copy(a, a+3)
	vm.AddPC(sbx)
}

/*
	TFORCALL和TFORLOOP指令协同工作共同实现lua中通用for的功能
	TFORCALL 负责调用迭代器即pairs方法并获取返回值 | R(A+3), ... ,R(A+2+C) := R(A)(R(A+1), R(A+2))
	TFORLOOP 决定是否继续循环 | if R(A+1) ~= nil then { R(A)=R(A+1); pc += sBx }

	for 变量列表 in 迭代器函数, 状态值, 初始控制变量 do
    -- 循环体
	end
	迭代器函数：每次调用时返回下一个值,R(A)
	状态值：通常是被遍历的对象（如表、字符串），在迭代过程中保持不变,R(A+1)
	初始控制变量：首次调用迭代器时的第二个参数（通常为 nil 或初始值）,R(A+2);每次迭代后，控制变量会被更新为上一次迭代器函数返回的第一个值,如果迭代器函数返回的第一个值为nil则停止循环

	"in 后必须是一个迭代器三元组（函数 + 状态 + 初始值）,可以是函数调用返回这三个东西"
	"pairs函数就是返回三个即next，table对象，nil"
	"pairs函数返回键对值，所以变量列表可以用k,v接受"

	for example:
	local arr = {"a", "b"}
	for k, v in pairs(arr) do
    	print(k, v)
	end

    1       [1]     NEWTABLE        0 2 0
    2       [1]     LOADK           1 -1    ; "a"
    3       [1]     LOADK           2 -2    ; "b"
    4       [1]     SETLIST         0 2 1   ; 1
    5       [2]     GETTABUP        1 0 -3  ; _ENV "pairs" 获取pairs函数
    6       [2]     MOVE            2 0		; 将表arr作为pairs的参数
											; 调用后的寄存器栈(0索引开始递增):[arr]->[pairs]->[arr_reference]
    7       [2]     CALL            1 2 4	; R(A),...,R(A+C-2):=R(A)(R(A+1),...,R(A+B-1)),
											; 调用pairs函数并将next迭代器，arr，nil作为返回值存入R(A),R(A+1),R(A+2)
											; 调用后的寄存器栈(0索引开始递增):[arr]->[next]->[arr_reference]->[nil]
	8       [2]     JMP             0 4     ; to 13
    9       [3]     GETTABUP        6 0 -4  ; _ENV "print" [6]=print
    10      [3]     MOVE            7 4		; move [k] to [7]
    11      [3]     MOVE            8 5		; move [v] to [8]
											; 调用后的寄存器栈(0索引开始递增):[arr]->[next]->[arr_reference]->[nil]->[k]->[v]->[print]->[7;k_copy]->[8;v_copy]
    12      [3]     CALL            6 3 1	; 调用[6]print函数,参数1[7]k,参数2[8]v
											; 这里的C==1所以此函数调用没有返回值
											; 调用后的寄存器栈(0索引开始递增):[arr]->[next]->[arr_reference]->[nil]->[k]->[v]
    13      [2]     TFORCALL        1   2	; R(A+3), ... ,R(A+2+C) := R(A)(R(A+1), R(A+2))
											; 调用后的寄存器栈(0索引开始递增):[arr]->[next]->[arr_reference]->[nil]->[k]->[v]
    14      [2]     TFORLOOP        3 -6    ; to 9 检查返回值，决定是否跳转 if R(A+1) ~= nil then { R(A)=R(A+1); pc += sBx }
											; 调用后的寄存器栈(0索引开始递增):[arr]->[next]->[arr_reference]->[k_copy;k!=nil]->[k]->[v]
    15      [4]     RETURN          0 1
*/

// R(A+3), ... ,R(A+2+C) := R(A)(R(A+1), R(A+2))
func tForCall(i Instruction, vm api.LuaVM) {
	a, _, c := i.ABC()
	a += 1

	vm.CheckStack(3)
	vm.PushValue(a)     //推入迭代器函数
	vm.PushValue(a + 1) //推入状态值
	vm.PushValue(a + 2) //推入初始值
	vm.Call(2, c)       //调用迭代器函数

	for i := a + c + 2; i >= a+3; i-- {
		vm.Replace(i)
	}
}

// if R(A+1) ~= nil then { R(A)=R(A+1); pc += sBx }
func tForLoop(i Instruction, vm api.LuaVM) {
	a, sbx := i.AsBx()
	a += 1

	//如果迭代器函数的第一个返回值为nil
	if !vm.IsNil(a + 1) {
		vm.Copy(a+1, a)
		vm.AddPC(sbx)
	}
}
