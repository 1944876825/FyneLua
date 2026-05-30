# FyneLua

Go + gopher-lua + Fyne v2 — 用 Lua 脚本驱动 Fyne GUI 应用。

## 特性

- 🎮 **丰富的控件**: Button, Label, Entry, Slider, Select, Check, RadioGroup, ProgressBar, Hyperlink, Icon, Image, Rectangle 等
- 📐 **布局容器**: VBox, HBox, GridWrap, Border, Scroll, Tabs/AppTabs
- 💬 **对话框**: 消息框、确认框、输入框、进度对话框、文件打开/保存
- 📋 **菜单栏**: 窗口菜单，支持子菜单、分隔线、回调
- 🌙 **主题切换**: Light/Dark 一键切换
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
├── main.go              # 入口 + 热重载
├── scripts/
│   └── main.lua         # 示例 Lua 脚本
├── bridge/
│   ├── bridge.go        # 核心桥接层（控件注册 + 方法分发）
│   ├── net.go           # fileio / net / timer 模块
│   ├── dialog.go        # 对话框模块
│   ├── menu.go          # 菜单栏辅助
│   └── widgets2.go      # Scroll / Icon / Image / Rectangle 等
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

-- 控件
local btn = gui.Button("点击")
btn:OnClick(function() end)
btn:SetText("新文字")
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
sel:OnChanged(function(s) end)

local chk = gui.Check("标签")
chk:OnChanged(function(checked) end)

local radio = gui.RadioGroup({"A", "B"})
radio:OnChanged(function(s) end)

local progress = gui.ProgressBar()
progress:SetValue(50)

-- 布局
gui.VBox(widget1, widget2, ...)
gui.HBox(widget1, widget2, ...)
gui.Border(top, bottom, left, right, center)
gui.GridWrap(cols, widget1, ...)
gui.Scroll(content)

-- Tab
gui.AppTabs(
    gui.TabItem("标签1", content1),
    gui.TabItem("标签2", content2)
)

-- 主题
gui.SetDarkMode(true/false)
gui.IsDarkMode()  -- returns boolean

-- 通用方法（所有控件可用）
widget:Resize(w, h)
widget:Show() / widget:Hide()
widget:Visible()
widget:Refresh()
widget:MinSize()  -- returns {w, h}
widget:Size()     -- returns {w, h}
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
