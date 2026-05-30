local gui = require("gui")
local fileio = require("fileio")
local net = require("net")
local timer = require("timer")
local dialog = require("dialog")

-- ========== 创建主窗口 ==========
local win = gui.Window("Fyne + Lua 完整演示", 750, 620)

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
        }
    },
    {
        label = "帮助",
        items = {
            { label = "关于", action = function()
                dialog.showInfo("关于", "Fyne + Lua 演示项目\nGo + gopher-lua + Fyne v2\n\n支持热重载、对话框、菜单、主题切换")
            end },
        }
    },
})

-- ========== 标签 ==========
local title = gui.Label("🎉 Fyne + Lua 完整演示")
local info  = gui.Label("网络 · 文件IO · 定时器 · 对话框 · 菜单 · 主题")
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

-- ========== Tab 1: 控件 ==========
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

-- ========== Tab 2: 网络 ==========
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

local btnLocalTest = gui.Button("本地测试")
btnLocalTest:OnClick(function()
    netStatusLabel:SetText("✅ 按钮点击正常！网络请求可能被墙")
end)

local netPanel = gui.VBox(
    gui.Label("🌐 HTTP 请求演示"),
    gui.Separator(),
    gui.HBox(btnHttpGet, btnHttpPost, btnLocalTest),
    netStatusLabel
)

-- ========== Tab 3: 文件 IO ==========
local fileStatusLabel = gui.Label("点击按钮测试文件读写")

local btnFileWrite = gui.Button("写入测试文件")
btnFileWrite:OnClick(function()
    local ok, err = fileio.writeFile("test_golua.txt", "Hello from golua-fyne! 🚀\n时间: " .. os.date())
    fileStatusLabel:SetText(ok and "✅ 文件写入成功" or "❌ 写入失败: " .. tostring(err))
end)

local btnFileRead = gui.Button("读取测试文件")
btnFileRead:OnClick(function()
    local content, err = fileio.readFile("test_golua.txt")
    if err then
        fileStatusLabel:SetText("❌ 读取失败: " .. tostring(err))
    else
        fileStatusLabel:SetText("📄 内容: " .. tostring(content))
    end
end)

local btnFileExists = gui.Button("检查文件")
local fileExistsLabel = gui.Label("")
btnFileExists:OnClick(function()
    local exists = fileio.exists("test_golua.txt")
    fileExistsLabel:SetText("文件存在: " .. tostring(exists))
end)

local btnFileList = gui.Button("列出目录")
local fileListLabel = gui.Label("")
btnFileList:OnClick(function()
    local files = fileio.listDir(".")
    fileListLabel:SetText("📁 文件数: " .. #files)
end)

local filePanel = gui.VBox(
    gui.Label("📁 文件 IO 演示"),
    gui.Separator(),
    gui.HBox(btnFileWrite, btnFileRead),
    gui.HBox(btnFileExists, btnFileList),
    gui.Separator(),
    fileStatusLabel,
    fileExistsLabel,
    fileListLabel
)

-- ========== Tab 4: 定时器 ==========
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

-- ========== Tab 5: 对话框 ==========
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

local btnProgressDialog = gui.Button("进度对话框")
btnProgressDialog:OnClick(function()
    local pd = dialog.showProgress("处理中", "正在下载...")
    local v = 0
    local tid = timer.every(200, function()
        v = v + 10
        pd.setValue(v)
        if v >= 100 then
            timer.cancel(tid)
            pd.hide()
            dialogStatus:SetText("✅ 进度对话框处理完成！")
        end
    end)
end)

local dialogPanel = gui.VBox(
    gui.Label("💬 对话框演示"),
    gui.Separator(),
    gui.HBox(btnMsgBox, btnErrorBox, btnConfirmBox),
    gui.HBox(btnInputBox, btnOpenFile, btnSaveFile),
    gui.HBox(btnProgressDialog),
    gui.Separator(),
    dialogStatus
)

-- ========== 组装 Tabs ==========
local tabs = gui.AppTabs(
    gui.TabItem("🎮 控件", tab1Content),
    gui.TabItem("🌐 网络", netPanel),
    gui.TabItem("📁 文件 IO", filePanel),
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
