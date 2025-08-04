-- os.time([time_table])

print(os.time())
print(os.time({
    year = 2023, 
    month = 7,
    day = 15,
    hour = 14,
    min = 30,
    sec = 0
}))

-- os.date([format [, timestamp]])

local time_str = os.date("%Y-%d-%m-%m %H:%M:%S", os.time())
print(time_str)
time_str = os.date("%H:%M:%S")
print(time_str)
