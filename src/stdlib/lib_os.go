package stdlib

import (
	"fmt"
	"strings"
	"time"

	"nskbz.cn/lua/api"
	"nskbz.cn/lua/tool"
)

var osFuncs map[string]api.GoFunc = map[string]api.GoFunc{
	"time": osTime,
	"date": osDate,
}

func OpenOsLib(vm api.LuaVM) int {
	vm.NewLib(osFuncs)
	return 1
}

// os.time([time_table])
//
//	local time_table = {
//	    year = 2023,
//	    month = 7,
//	    day = 15,
//	    hour = 14,
//	    min = 30,
//	    sec = 0
//	}
//
// 字段名	必填	默认值	取值范围
// year		是		无		1970+
// month	是		无		1-12
// day		是		无		1-31
// hour		否		12		0-23
// min		否		0		0-59
// sec		否		0		0-61
//
// 无参则返回当前时间戳，有参数则返回参数所对应的时间戳
func osTime(vm api.LuaVM) int {
	if vm.GetTop() == 0 { //无参数
		vm.PushInteger(time.Now().Unix())
		return 1
	}
	vm.CheckType(1, api.LUAVALUE_TABLE)
	year, month, day, hour, min, sec := 0, 0, 0, 12, 0, 0
	if api.LUAVALUE_NIL == vm.GetField(1, "year") {
		tool.Fatal(vm, "time_table not exist year")
	}
	year = int(vm.ToInteger(0))
	if year < 1970 {
		tool.Fatal(vm, fmt.Sprintf("year must between [1970,+), but %d", year))
	}
	if api.LUAVALUE_NIL == vm.GetField(1, "month") {
		tool.Fatal(vm, "time_table not exist month")
	}
	month = int(vm.ToInteger(0))
	if month < 1 || month > 12 {
		tool.Fatal(vm, fmt.Sprintf("month must between [1,12], but %d", month))
	}
	if api.LUAVALUE_NIL == vm.GetField(1, "day") {
		tool.Fatal(vm, "time_table not exist day")
	}
	day = int(vm.ToInteger(0))
	if day < 1 || day > 31 {
		tool.Fatal(vm, fmt.Sprintf("day must between [1,31], but %d", day))
	}
	if h := vm.GetField(1, "hour"); h != api.LUAVALUE_NIL {
		hour = int(vm.ToInteger(0))
		if hour < 0 || hour > 23 {
			tool.Fatal(vm, fmt.Sprintf("hour must between [0,23], but %d", hour))
		}
	}
	if m := vm.GetField(1, "min"); m != api.LUAVALUE_NIL {
		min = int(vm.ToInteger(0))
		if min < 0 || min > 59 {
			tool.Fatal(vm, fmt.Sprintf("min must between [0,59], but %d", min))
		}
	}
	if s := vm.GetField(1, "sec"); s != api.LUAVALUE_NIL {
		sec = int(vm.ToInteger(0))
		if sec < 0 || sec > 61 {
			tool.Fatal(vm, fmt.Sprintf("sec must between [0,61], but %d", sec))
		}
	}

	timestamp := time.Date(year, time.Month(month), day, hour, min, sec, 0, time.UTC).Unix()
	vm.PushInteger(timestamp)
	return 1
}

// os.date([format [, timestamp]])
// format = "%Y-%m-%d %H:%M:%S"
// 用于格式化时间戳为可读字符串或时间表的函数,如果没有timestamp则默认当前时间戳
// 如果format=="*t"则返回一个time_table
func osDate(vm api.LuaVM) int {
	vm.CheckString(1)
	format := vm.ToString(1)
	timestamp := time.Now()
	if format == "*t" {
		// todo
	}
	if vm.GetTop() > 1 {
		vm.CheckInteger(2)
		timestamp = time.Unix(vm.ToInteger(2), 0)
	}
	//golang的format: "2006-01-02 15:04:05"
	format = strings.ReplaceAll(format, "%Y", "2006")
	format = strings.ReplaceAll(format, "%m", "01")
	format = strings.ReplaceAll(format, "%d", "02")
	format = strings.ReplaceAll(format, "%H", "15")
	format = strings.ReplaceAll(format, "%M", "04")
	format = strings.ReplaceAll(format, "%S", "05")
	vm.PushString(timestamp.Format(format))
	return 1
}
