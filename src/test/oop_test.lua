-- 外层表
local company = {
    name = "Tech Corp",
    
    -- 公司级别的方法
    getCompanyName = function(self)
        return self.name
    end,
    
    -- 嵌套部门表
    departments = {
        Engineering = {
            head = "Alice",
            size = 50,
            
            -- 部门级别的方法
            getDepartmentInfo = function(self)
                return "Engineering".."(Head:"..self.head..", Size:"..self.size..")"
            end,
            
            -- 更深层的嵌套
            teams = {
                Backend = {
                    lead = "Bob",
                    getTeamLead = function(self) return self.lead end
                },
                Frontend = {
                    lead = "Charlie",
                    getTeamLead = function(self) return self.lead end
                }
            }
        },
        -- 另一个部门
        Marketing = {
            head = "David",
            size = 20,
            getDepartmentInfo = function(self)
                return "Marketing".."(Head:"..self.head..", Size:"..self.size..")"
            end
        }
    }
}

-- 使用冒号语法调用方法
print(company:getCompanyName())  -- 输出: Tech Corp

-- 访问嵌套表的方法
print(company.departments.Engineering:getDepartmentInfo())
print(company.departments.Engineering.teams.Backend:getTeamLead())

-- 动态添加方法
company.departments.Marketing.getCampaigns = function(self)
    return {"Summer Sale", "Black Friday"}
end

print(company.departments.Marketing:getCampaigns())