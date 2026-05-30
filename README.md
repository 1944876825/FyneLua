# FyneLua

Go + gopher-lua + Fyne v2 — 用 Lua 脚本驱动 Fyne GUI 应用。

## 特性

- 🎮 **丰富控件**: Button, Label, Entry, Slider, Select, Check, RadioGroup, ProgressBar, Hyperlink, Icon, Image, Rectangle
- 📋 **数据控件**: List（动态列表）, Tree（树形控件）, Table（表格）
- 📐 **布局容器**: VBox, HBox, GridWrap, Border, Scroll, Tabs/AppTabs, HSplit/VSplit（可拖拽分割）
- 🃏 **复合控件**: Card（卡片）, Accordion（折叠面板）, Form（表单）, Toolbar（工具栏）
- 💬 **对话框**: 消息框、确认框、输入框、进度对话框、文件打开/保存
- 📋 **菜单栏**: 窗口菜单，支持子菜单、分隔线、回调
- 🌙 **主题切换**: Light/Dark 一键切换
- 📋 **剪贴板**: 读写系统剪贴板
- 🌐 **网络请求**: 异步 HTTP GET/POST
- 📁 **文件 IO**: 读写、追加、目录列表、文件存在检测
- ⏱️ **定时器**: 延时执行、周期执行
- 🔥 **热重载**: 修改 .lua 文件自动刷新 UI，无需重新编译
- ❌ **错误弹窗**: Lua 脚本报错自动弹对话框，不用看控制台

## 快速开始

### 编译

```bash
# 需要 MSYS2 + mingw-w64-x86_64-gcc (Windows)
CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o golua-fyne.exe .
```

### 运行

```bash
./golua-fyne.exe scripts/main.lua
```

修改 `scripts/main.lua` 保存后窗口会自动刷新。

## 项目结构

```
├── main.go                # 入口 + 热重载
├── scripts/
│   └── main.lua           # 示例 Lua 脚本
├── bridge/
│   ├── bridge.go          # 核心桥接层（控件注册 + 方法分发）
│   ├── net.go             # fileio / net / timer 模块
│   ├── dialog.go          # 对话框模块
│   ├── menu.go            # 菜单栏辅助
│   ├── widgets2.go        # Scroll / Icon / Image / Rectangle 等
│   └── data_widgets.go    # List / Tree / Table / Split / Toolbar / Card / Accordion / Form
└── go.mod
```

## Lua API

### gui 模块

```lua
local gui = require("gui")

-- 窗口
local win = gui.Window("标题", 800, 600)
win:SetContent(content)
win:CenterOnScreen()
win:Show()
win:SetMainMenu({...})
win:OnClosed(function() end)
win:SetFixedSize(true/false)

-- 基础控件
local btn = gui.Button("点击")
btn:OnClick(function() end)
btn:SetText("新文字")
btn:SetImportance("high")  -- high/medium/low
btn:Enable() / btn:Disable()

local lbl = gui.Label("文字")
lbl:SetText("新文字")

local entry = gui.Entry()
entry:SetPlaceHolder("提示")
entry:OnSubmitted(function(text) end)
entry:OnChanged(function(text) end)

local slider = gui.Slider(0, 100)
slider:SetValue(50)
slider:OnChanged(function(v) end)

local sel = gui.Select({"选项1", "选项2"})
sel:SetSelected("选项1")
sel:OnChanged(function(s) end)

local chk = gui.Check("标签")
chk:SetChecked(true)
chk:OnChanged(function(checked) end)

local radio = gui.RadioGroup({"A", "B"})
radio:SetSelected("A")
radio:OnChanged(function(s) end)

local progress = gui.ProgressBar()
progress:SetMin(0)
progress:SetMax(100)
progress:SetValue(50)

local hl = gui.Hyperlink("链接", "https://example.com")
local sep = gui.Separator()

-- 多行输入 & 密码输入
local mle = gui.MultiLineEntry()
local pwd = gui.PasswordEntry()

-- 图标 & 图片 & 矩形
local icon = gui.Icon("search")      -- 主题图标名
local img = gui.Image("path.png")
img:FillMode("stretch")             -- stretch/contain/original
img:SetMinSize(100, 100)
local rect = gui.Rectangle("#FF0000")
rect:SetColor("#00FF00")

-- ===== 数据控件 =====

-- List 列表
local list = gui.List(
    function() return N end,           -- 返回条目数量
    function(i) return gui.Label(items[i+1]) end,  -- 创建/更新控件
    function(i) print("选中:", i) end  -- 点击回调（可选）
end)
list:Refresh()

-- Tree 树形
local tree = gui.Tree(
    function(uid) return children[uid] or {} end,  -- 子节点列表
    function(uid) return isBranch[uid] end,         -- 是否是分支
    function(isBranch) return gui.Label("...") end, -- 创建控件
    function(uid) print("选中:", uid) end           -- 点击回调（可选）
end)
tree:Open("uid")
tree:Close("uid")
tree:OpenAll()
tree:CloseAll()
tree:IsOpen("uid")

-- Table 表格
local tbl = gui.Table(
    function() return rows, cols end,                -- 返回行列数
    function(col, row) return gui.Label("...") end, -- 创建/更新单元格
    function(col, row) print(col, row) end           -- 点击回调（可选）
end)
tbl:SetColumnWidth(0, 150)

-- ===== 布局容器 =====
gui.VBox(widget1, widget2, ...)
gui.HBox(widget1, widget2, ...)
gui.Border(top, bottom, left, right, center)
gui.GridWrap(cols, widget1, ...)
gui.Scroll(content)
gui.HSplit(left, right)   -- 水平可拖拽分割
gui.VSplit(top, bottom)   -- 垂直可拖拽分割
split:SetOffset(0.5)       -- 设置分割比例 (0.0~1.0)
split:Offset()             -- 获取当前比例

-- Tab
gui.AppTabs(
    gui.TabItem("标签1", content1),
    gui.TabItem("标签2", content2)
)

-- ===== 复合控件 =====

-- Card 卡片
local card = gui.Card("标题", "副标题", content)
card:SetTitle("新标题")
card:SetSubTitle("新副标题")
card:Title()
card:SubTitle()

-- Accordion 折叠面板
local acc = gui.Accordion({
    {title = "标题1", content = widget, open = true},
    {title = "标题2", content = widget, open = false},
})
acc:Open(0)
acc:Close(0)
acc:OpenAll()
acc:CloseAll()

-- Form 表单
local form = gui.Form({
    {label = "姓名", widget = nameEntry},
    {label = "邮箱", widget = emailEntry},
})
form:OnSubmit(function() end)
form:OnCancel(function() end)
form:Append("新字段", newWidget)

-- Toolbar 工具栏
local toolbar = gui.Toolbar({
    {icon = "search",       action = function() end},
    {icon = "contentcopy",  action = function() end},
    {separator = true},
    {icon = "documentsave", action = function() end},
})

-- ===== 主题 =====
gui.SetDarkMode(true/false)
gui.IsDarkMode()  -- returns boolean

-- ===== 通用方法（所有控件可用） =====
widget:Resize(w, h)
widget:Show() / widget:Hide()
widget:Visible()
widget:Refresh()
widget:MinSize()  -- returns {w, h}
widget:Size()     -- returns {w, h}
```

### clipboard 模块

```lua
local clipboard = require("clipboard")

clipboard.Set("要复制的文本")
local text = clipboard.Get()
```

### dialog 模块

```lua
local dialog = require("dialog")

dialog.showInfo("标题", "内容")
dialog.showError("标题", "内容")
dialog.showConfirm("标题", "确认?", function(ok) end)
dialog.showEntry("标题", "提示", "默认值", function(text) end)

local pd = dialog.showProgress("处理中", "请等待...")
pd.setValue(0.5)
pd.hide()

dialog.showFileOpen(function(path) end)
dialog.showFileSave(function(path) end)
dialog.showFileOpenFiltered("描述", {".lua", ".txt"}, function(path) end)
```

### fileio 模块

```lua
local fileio = require("fileio")

local content, err = fileio.readFile("path")
local ok, err = fileio.writeFile("path", "content")
local ok, err = fileio.appendFile("path", "content")
local exists = fileio.exists("path")
local files = fileio.listDir(".")
local ok, err = fileio.remove("path")
local ok, err = fileio.mkdir("path")
```

### net 模块

```lua
local net = require("net")

net.get("http://example.com", function(resp, err) end)
net.post("http://example.com", '{"key":"value"}', function(resp, err) end)
-- resp = {status=number, body=string}
```

### timer 模块

```lua
local timer = require("timer")

timer.after(1000, function() end)
local id = timer.every(1000, function() end)
timer.cancel(id)
```

## 依赖

- Go 1.26+
- [Fyne v2.7](https://fyne.io/)
- [gopher-lua](https://github.com/yuin/gopher-lua) (纯 Go Lua 5.1 实现)
- [fsnotify](https://github.com/fsnotify/fsnotify) (文件监听)

## License

MIT
