local gui = require("gui")
local fileio = require("fileio")
local net = require("net")
local timer = require("timer")
local dialog = require("dialog")
local clipboard = require("clipboard")

-- ========== 创建主窗口 ==========
local win = gui.Window("FyneLua 完整演示", 850, 650)

-- ========== 菜单栏 ==========
local isDark = gui.IsDarkMode()

win:SetMainMenu({
    {
        label = "文件",
        items = {
            { label = "打开文件...", action = function()
                dialog.showFileOpen(function(path)
                    if path then
                        local content, err = fileio.readFile(path)
                        if content then
                            info:SetText("📄 已打开: " .. path .. " (" .. #content .. " bytes)")
                        else
                            dialog.showError("读取失败", tostring(err))
                        end
                    end
                end)
            end },
            { label = "-" },
            { label = "退出", action = function()
                os.exit(0)
            end },
        }
    },
    {
        label = "查看",
        items = {
            { label = "切换主题", action = function()
                isDark = not isDark
                gui.SetDarkMode(isDark)
                themeLabel:SetText("主题: " .. (isDark and "🌙 Dark" or "☀️ Light"))
            end },
            { label = "打开新窗口", action = function()
                local win2 = gui.NewWindow("独立窗口", 400, 300)
                local lbl = gui.Label("🎉 这是一个独立窗口！\n\n热重载时会被关闭重建。")
                local btnClose = gui.Button("关闭此窗口")
                btnClose:OnClick(function()
                    win2:Close()
                end)
                win2:SetContent(gui.VBox(lbl, btnClose))
                win2:CenterOnScreen()
                win2:Show()
            end },
        }
    },
    {
        label = "帮助",
        items = {
            { label = "关于", action = function()
                dialog.showInfo("关于", "FyneLua 演示项目\nGo + gopher-lua + Fyne v2\n\n新增: List, Tree, Table, Split, Toolbar,\nCard, Accordion, Form, Clipboard")
            end },
        }
    },
})

-- ========== 标题 ==========
local title = gui.Label("🎉 FyneLua 完整演示")
local info  = gui.Label("List · Tree · Table · Split · Toolbar · Card · Accordion · Form · Clipboard")
local themeLabel = gui.Label("主题: " .. (isDark and "🌙 Dark" or "☀️ Light"))

-- ========== 按钮区 ==========
local btnClick  = gui.Button("点我计数")
local btnReset  = gui.Button("重置")
local btnDanger = gui.Button("禁用我")
btnDanger:SetImportance("high")
btnDanger:Disable()

local clickCount = 0
local statusLabel = gui.Label("等待点击...")

btnClick:OnClick(function()
    clickCount = clickCount + 1
    statusLabel:SetText("已点击 " .. clickCount .. " 次")
    if clickCount >= 10 then
        btnDanger:Enable()
        statusLabel:SetText("已点击 " .. clickCount .. " 次 — 高级按钮已解锁！")
    end
end)

btnReset:OnClick(function()
    clickCount = 0
    statusLabel:SetText("已重置")
    btnDanger:Disable()
end)

btnDanger:OnClick(function()
    statusLabel:SetText("💥 高级按钮被触发了！")
    btnDanger:Disable()
end)

-- ========== 输入框 ==========
local entry = gui.Entry()
entry:SetPlaceHolder("在这里输入文字...")
entry:OnSubmitted(function(text)
    info:SetText("你提交了: " .. text)
end)

-- ========== 滑块 ==========
local sliderValue = gui.Label("滑块值: 50")
local slider = gui.Slider(0, 100)
slider:SetValue(50)
slider:OnChanged(function(v)
    sliderValue:SetText("滑块值: " .. math.floor(v))
end)

-- ========== 选择器 ==========
local selectStatus = gui.Label("选择: 无")
local sel = gui.Select({"苹果", "香蕉", "橙子", "葡萄", "西瓜"})
sel:OnChanged(function(s)
    selectStatus:SetText("选择: " .. s)
end)

-- ========== 复选框 ==========
local checkLabel = gui.Label("复选框状态: 未选中")
local chk = gui.Check("同意协议")
chk:OnChanged(function(checked)
    checkLabel:SetText("复选框状态: " .. (checked and "✅ 已选中" or "未选中"))
end)

-- ========== 单选组 ==========
local radioLabel = gui.Label("选择语言: 无")
local radio = gui.RadioGroup({"Go", "Python", "Lua", "Rust"})
radio:OnChanged(function(s)
    radioLabel:SetText("选择语言: " .. s)
end)

-- ========== 进度条 ==========
local progress = gui.ProgressBar()
progress:SetMin(0)
progress:SetMax(100)
progress:SetValue(30)

local btnProgress = gui.Button("加载进度")
btnProgress:OnClick(function()
    for i = 1, 10 do
        progress:SetValue(i * 10)
    end
    statusLabel:SetText("进度加载完成！")
end)

-- ========== Tab 1: 基础控件 ==========
local buttonRow = gui.HBox(btnClick, btnReset, btnDanger, btnProgress)

local leftPanel = gui.VBox(
    gui.Label("📝 文字输入"),
    entry,
    gui.Separator(),
    gui.Label("🎚️ 滑块"),
    slider,
    sliderValue,
    gui.Separator(),
    gui.Label("🍎 水果选择"),
    sel,
    selectStatus
)

local rightPanel = gui.VBox(
    checkLabel,
    chk,
    radioLabel,
    radio,
    gui.Separator(),
    gui.Label("📊 进度条"),
    progress,
    statusLabel
)

local tab1Content = gui.VBox(buttonRow, gui.HBox(leftPanel, rightPanel))

-- ========== Tab 2: List 列表 ==========
local items = {"Go", "Python", "Lua", "Rust", "JavaScript", "TypeScript", "C++", "Java", "Kotlin", "Swift"}
local listStatusLabel = gui.Label("点击列表项查看")

local listWidget = gui.List(
    function() return #items end,
    function(i)
        return gui.Label("  📌 " .. items[i + 1])
    end,
    function(i)
        listStatusLabel:SetText("✅ 选中: " .. items[i + 1])
    end
)

local btnAddItem = gui.Button("添加项")
btnAddItem:OnClick(function()
    local name = "新语言 " .. (#items + 1)
    table.insert(items, name)
    listWidget:Refresh()
    listStatusLabel:SetText("➕ 已添加: " .. name)
end)

local listPanel = gui.VBox(
    gui.Label("📋 List 列表演示"),
    gui.Separator(),
    gui.HBox(btnAddItem),
    listWidget,
    listStatusLabel
)

-- ========== Tab 3: Tree 树形控件 ==========
local treeStatusLabel = gui.Label("点击树节点查看")

local treeData = {
    root = {"src", "docs", "scripts"},
    src = {"main.go", "bridge", "go.mod"},
    bridge = {"bridge.go", "dialog.go", "net.go", "widgets2.go", "data_widgets.go"},
    docs = {"README.md", "LICENSE"},
    scripts = {"main.lua"},
}

local isBranchMap = {
    root = true, src = true, bridge = true, docs = true, scripts = true,
}

local treeWidget = gui.Tree(
    function(uid)
        return treeData[uid] or {}
    end,
    function(uid)
        return isBranchMap[uid] or false
    end,
    function(isBranch)
        return gui.Label("  📄 " .. (isBranch and "📁" or "📄") .. " ...")
    end,
    function(uid)
        treeStatusLabel:SetText("📂 选中: " .. uid)
    end
)

local treePanel = gui.VBox(
    gui.Label("🌳 Tree 树形控件演示"),
    gui.Separator(),
    treeWidget,
    treeStatusLabel
)

-- ========== Tab 4: Table 表格 ==========
local tableData = {
    {"Go",      "1.24",  "静态编译", "⭐⭐⭐⭐⭐"},
    {"Python",  "3.12",  "解释执行", "⭐⭐⭐⭐"},
    {"Lua",     "5.4",   "嵌入式",   "⭐⭐⭐⭐⭐"},
    {"Rust",    "1.80",  "内存安全", "⭐⭐⭐⭐"},
    {"Java",    "21",    "跨平台",   "⭐⭐⭐"},
}

local tableStatusLabel = gui.Label("点击表格单元格查看")

local tableWidget = gui.Table(
    function() return #tableData, 4 end,
    function(col, row)
        local r = tableData[row + 1]
        if r then
            return gui.Label("  " .. (r[col + 1] or ""))
        end
        return gui.Label("")
    end,
    function(col, row)
        local r = tableData[row + 1]
        if r then
            tableStatusLabel:SetText("📊 [" .. (row+1) .. "," .. (col+1) .. "] " .. (r[col + 1] or ""))
        end
    end
)

local tablePanel = gui.VBox(
    gui.Label("📊 Table 表格演示"),
    gui.Separator(),
    tableWidget,
    tableStatusLabel
)

-- ========== Tab 5: Split + Toolbar ==========
local leftContent = gui.VBox(
    gui.Label("◀ 左侧面板"),
    gui.Separator(),
    gui.Label("可拖拽分割线"),
    gui.Label("调整两侧大小"),
    gui.Label("左侧"),
    gui.Label("左侧"),
    gui.Label("左侧")
)

local rightContent = gui.VBox(
    gui.Label("右侧面板 ▶"),
    gui.Separator(),
    gui.Label("这是右侧内容"),
    gui.Label("可以放任何控件"),
    gui.Label("右侧"),
    gui.Label("右侧"),
    gui.Label("右侧")
)

local splitWidget = gui.HSplit(leftContent, rightContent)

local toolbarStatusLabel = gui.Label("点击工具栏按钮")

local toolbar = gui.Toolbar({
    {icon = "contentcopy",  action = function()
        clipboard.Set("Hello from FyneLua!")
        toolbarStatusLabel:SetText("📋 已复制到剪贴板")
    end},
    {icon = "contentpaste", action = function()
        local text = clipboard.Get()
        toolbarStatusLabel:SetText("📋 剪贴板内容: " .. text)
    end},
    {separator = true},
    {icon = "search",       action = function()
        toolbarStatusLabel:SetText("🔍 搜索功能...")
    end},
    {icon = "contentadd",   action = function()
        toolbarStatusLabel:SetText("➕ 添加功能...")
    end},
    {icon = "documentsave", action = function()
        toolbarStatusLabel:SetText("💾 保存功能...")
    end},
    {separator = true},
    {icon = "visibility",   action = function()
        toolbarStatusLabel:SetText("👁️ 显示/隐藏...")
    end},
})

local splitPanel = gui.VBox(
    gui.Label("🔀 Split 分割 + Toolbar 工具栏"),
    gui.Separator(),
    toolbar,
    gui.Separator(),
    splitWidget,
    toolbarStatusLabel
)

-- ========== Tab 6: Card + Accordion + Form ==========
local card1 = gui.Card("欢迎使用", "FyneLua 演示项目", gui.Label("这是一个用 Lua 驱动的 Fyne GUI 应用框架"))
local card2 = gui.Card("新功能", "v2.0 新增", gui.Label("List, Tree, Table, Split, Toolbar, Card, Accordion, Form, Clipboard"))

local cardPanel = gui.VBox(
    gui.Label("🃏 Card 卡片演示"),
    gui.Separator(),
    gui.HBox(card1, card2)
)

local accordionContent1 = gui.Label("这是第一项的详细内容。\n\nAccordion 可以折叠展开，\n适合做 FAQ 或设置面板。")
local accordionContent2 = gui.Label("第二项内容。\n\n支持 Open/Close/OpenAll/CloseAll 方法。")
local accordionContent3 = gui.Label("第三项内容。")

local acc = gui.Accordion({
    {title = "📖 什么是 FyneLua？",     content = accordionContent1, open = true},
    {title = "🔧 如何使用？",           content = accordionContent2, open = false},
    {title = "💡 提示与技巧",           content = accordionContent3, open = false},
})

local accordionPanel = gui.VBox(
    gui.Label("🪗 Accordion 折叠面板演示"),
    gui.Separator(),
    acc
)

local formName = gui.Entry()
formName:SetPlaceHolder("输入姓名...")
local formEmail = gui.Entry()
formEmail:SetPlaceHolder("输入邮箱...")
local formBio = gui.MultiLineEntry()
formBio:SetPlaceHolder("个人简介...")

local form = gui.Form({
    {label = "姓名", widget = formName},
    {label = "邮箱", widget = formEmail},
    {label = "简介", widget = formBio},
})
form:OnSubmit(function()
    dialog.showInfo("表单提交", "姓名: " .. formName:Text() .. "\n邮箱: " .. formEmail:Text())
end)
form:OnCancel(function()
    formName:SetText("")
    formEmail:SetText("")
    formBio:SetText("")
end)

local formPanel = gui.VBox(
    gui.Label("📝 Form 表单演示"),
    gui.Separator(),
    form
)

local tab6Content = gui.VBox(
    cardPanel,
    gui.Separator(),
    accordionPanel,
    gui.Separator(),
    formPanel
)

-- ========== Tab 7: 网络 ==========
local netStatusLabel = gui.Label("点击按钮发送 HTTP 请求")

local btnHttpGet = gui.Button("HTTP GET")
btnHttpGet:OnClick(function()
    netStatusLabel:SetText("⏳ 请求中...")
    net.get("http://httpbin.org/get", function(resp, err)
        if err then
            netStatusLabel:SetText("❌ GET 失败: " .. tostring(err))
        else
            local short = string.sub(tostring(resp.body), 1, 100)
            netStatusLabel:SetText("✅ GET: " .. resp.status .. " | " .. short)
        end
    end)
end)

local btnHttpPost = gui.Button("HTTP POST")
btnHttpPost:OnClick(function()
    netStatusLabel:SetText("⏳ POST 请求中...")
    net.post("http://httpbin.org/post", '{"hello":"golua-fyne"}', function(resp, err)
        if err then
            netStatusLabel:SetText("❌ POST 失败: " .. tostring(err))
        else
            local short = string.sub(tostring(resp.body), 1, 100)
            netStatusLabel:SetText("✅ POST: " .. resp.status .. " | " .. short)
        end
    end)
end)

local netPanel = gui.VBox(
    gui.Label("🌐 HTTP 请求演示"),
    gui.Separator(),
    gui.HBox(btnHttpGet, btnHttpPost),
    netStatusLabel
)

-- ========== Tab 8: 定时器 ==========
local timerLabel = gui.Label("⏱️ 倒计时: 10 秒")
local countdownTimerId = nil

local btnStartTimer = gui.Button("开始倒计时")
btnStartTimer:OnClick(function()
    if countdownTimerId then
        timer.cancel(countdownTimerId)
    end
    local remaining = 10
    timerLabel:SetText("⏱️ 倒计时: " .. remaining .. " 秒")
    countdownTimerId = timer.every(1000, function()
        remaining = remaining - 1
        if remaining <= 0 then
            timerLabel:SetText("🎉 倒计时结束！")
            timer.cancel(countdownTimerId)
            countdownTimerId = nil
        else
            timerLabel:SetText("⏱️ 倒计时: " .. remaining .. " 秒")
        end
    end)
end)

local btnStopTimer = gui.Button("停止")
btnStopTimer:OnClick(function()
    if countdownTimerId then
        timer.cancel(countdownTimerId)
        countdownTimerId = nil
        timerLabel:SetText("⏱️ 已停止")
    end
end)

local afterLabel = gui.Label("")
local btnAfter = gui.Button("3秒后执行")
btnAfter:OnClick(function()
    afterLabel:SetText("⏳ 等待 3 秒...")
    timer.after(3000, function()
        afterLabel:SetText("✅ 3秒后执行了！")
    end)
end)

local timerPanel = gui.VBox(
    gui.Label("⏱️ 定时器演示"),
    gui.Separator(),
    gui.HBox(btnStartTimer, btnStopTimer, btnAfter),
    timerLabel,
    afterLabel
)

-- ========== Tab 9: 对话框 ==========
local dialogStatus = gui.Label("点击按钮测试各种对话框")

local btnMsgBox = gui.Button("消息框")
btnMsgBox:OnClick(function()
    dialog.showInfo("提示", "这是一个消息框！\nFyne + Lua 真的很方便。")
end)

local btnErrorBox = gui.Button("错误框")
btnErrorBox:OnClick(function()
    dialog.showError("错误", "这是一个错误示例！\n用于显示严重问题。")
end)

local btnConfirmBox = gui.Button("确认框")
btnConfirmBox:OnClick(function()
    dialog.showConfirm("确认", "你确定要执行此操作吗？", function(ok)
        dialogStatus:SetText(ok and "✅ 用户点击了确认" or "❌ 用户点击了取消")
    end)
end)

local btnInputBox = gui.Button("输入框")
btnInputBox:OnClick(function()
    dialog.showEntry("输入", "请输入你的名字:", "张三", function(text)
        if text then
            dialogStatus:SetText("👋 你好, " .. text .. "!")
        end
    end)
end)

local btnOpenFile = gui.Button("打开文件")
btnOpenFile:OnClick(function()
    dialog.showFileOpenFiltered("Lua 脚本", {".lua", ".txt"}, function(path)
        if path then
            dialogStatus:SetText("📂 选择文件: " .. path)
        end
    end)
end)

local btnSaveFile = gui.Button("保存文件")
btnSaveFile:OnClick(function()
    dialog.showFileSave(function(path)
        if path then
            dialogStatus:SetText("💾 保存到: " .. path)
        end
    end)
end)

local dialogPanel = gui.VBox(
    gui.Label("💬 对话框演示"),
    gui.Separator(),
    gui.HBox(btnMsgBox, btnErrorBox, btnConfirmBox),
    gui.HBox(btnInputBox, btnOpenFile, btnSaveFile),
    gui.Separator(),
    dialogStatus
)

-- ========== 组装 Tabs ==========
local tabs = gui.AppTabs(
    gui.TabItem("🎮 控件", tab1Content),
    gui.TabItem("📋 List", listPanel),
    gui.TabItem("🌳 Tree", treePanel),
    gui.TabItem("📊 Table", tablePanel),
    gui.TabItem("🔀 Split+Toolbar", splitPanel),
    gui.TabItem("🃏 Card+Accordion+Form", tab6Content),
    gui.TabItem("🌐 网络", netPanel),
    gui.TabItem("⏱️ 定时器", timerPanel),
    gui.TabItem("💬 对话框", dialogPanel)
)

-- ========== 最终布局 ==========
local mainContent = gui.VBox(
    title,
    info,
    themeLabel,
    gui.Separator(),
    tabs
)

-- ========== 窗口关闭事件 ==========
win:OnClosed(function()
    print("窗口关闭了，共点击了 " .. clickCount .. " 次")
end)

win:SetContent(mainContent)
win:CenterOnScreen()
win:Show()
