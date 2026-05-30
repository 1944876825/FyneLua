local gui = require("gui")
local fileio = require("fileio")

local win = gui.Window("文件 IO 测试", 400, 300)

local statusLabel = gui.Label("状态: 等待操作")

local btnWrite = gui.Button("写入文件")
btnWrite:OnClick(function()
    local ok, err = fileio.writeFile("test.txt", "Hello! " .. os.date())
    statusLabel:SetText(ok and "✅ 写入成功" or "❌ 写入失败: " .. tostring(err))
end)

local btnRead = gui.Button("读取文件")
btnRead:OnClick(function()
    local content, err = fileio.readFile("test.txt")
    if err then
        statusLabel:SetText("❌ 读取失败: " .. tostring(err))
    else
        statusLabel:SetText("📄 " .. tostring(content))
    end
end)

local btnCheck = gui.Button("检查文件")
btnCheck:OnClick(function()
    local exists = fileio.exists("test.txt")
    statusLabel:SetText("文件存在: " .. tostring(exists))
end)

local content = gui.VBox(
    gui.Label("📁 文件 IO 最简测试"),
    gui.Separator(),
    gui.HBox(btnWrite, btnRead, btnCheck),
    gui.Separator(),
    statusLabel
)

win:SetContent(content)
win:ShowAndRun()
