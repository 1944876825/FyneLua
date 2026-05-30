package bridge

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/yuin/gopher-lua"
)

// LuaWidget wraps a Fyne widget or window for use inside Lua.
type LuaWidget struct {
	Win  fyne.Window      // set only when Type == "Window"
	Obj  fyne.CanvasObject // set for all non-Window types
	Type string            // "Window", "Button", "Label", "Entry", "VBox", "HBox", etc.
}

const widgetMT = "fyne_widget"

const tabItemMT = "fyne_tabitem"

// HotReload controls whether ShowAndRun behaves as non-blocking Show.
var HotReload = false

// CurrentL holds the current Lua LState for all async callbacks.
// Updated on each NewEngine call so hot-reload works correctly.
var CurrentL *lua.LState

// SetHotReload enables or disables hot-reload mode.
// When enabled, win:ShowAndRun() acts like win:Show() (non-blocking).
func SetHotReload(enabled bool) {
	HotReload = enabled
}

// NewEngine creates a new gopher-lua LState with the "gui" module preloaded.
// The returned LState must be used on the same goroutine that calls Fyne's Run().
func NewEngine(a fyne.App) *lua.LState {
	L := lua.NewState()
	CurrentL = L

	// Register metatable for all fyne widget userdata
	mt := L.NewTypeMetatable(widgetMT)
	L.SetField(mt, "__index", L.NewFunction(widgetIndex))

	// Register metatable for TabItem userdata
	tmt := L.NewTypeMetatable(tabItemMT)
	L.SetField(tmt, "__index", L.NewFunction(tabItemIndex))

	// Build the "gui" module table
	mod := L.NewTable()
	L.SetField(mod, "Window", L.NewFunction(lWindowFn(a)))
	L.SetField(mod, "NewWindow", L.NewFunction(lNewWindowFn(a)))
	L.SetField(mod, "Button", L.NewFunction(lButtonFn))
	L.SetField(mod, "Label", L.NewFunction(lLabelFn))
	L.SetField(mod, "Entry", L.NewFunction(lEntryFn))
	L.SetField(mod, "VBox", L.NewFunction(lVBoxFn))
	L.SetField(mod, "HBox", L.NewFunction(lHBoxFn))
	// New widgets
	L.SetField(mod, "Slider", L.NewFunction(lSliderFn))
	L.SetField(mod, "Select", L.NewFunction(lSelectFn))
	L.SetField(mod, "Check", L.NewFunction(lCheckFn))
	L.SetField(mod, "RadioGroup", L.NewFunction(lRadioGroupFn))
	L.SetField(mod, "ProgressBar", L.NewFunction(lProgressBarFn))
	L.SetField(mod, "Separator", L.NewFunction(lSeparatorFn))
	L.SetField(mod, "Hyperlink", L.NewFunction(lHyperlinkFn))
	// New layouts
	L.SetField(mod, "GridWrap", L.NewFunction(lGridWrapFn))
	L.SetField(mod, "Border", L.NewFunction(lBorderFn))
	L.SetField(mod, "Tabs", L.NewFunction(lTabsFn(false)))
	L.SetField(mod, "AppTabs", L.NewFunction(lTabsFn(true)))
	L.SetField(mod, "TabItem", L.NewFunction(lTabItemFn))
	// New widgets
	L.SetField(mod, "Scroll", L.NewFunction(lScrollFn))
	L.SetField(mod, "Icon", L.NewFunction(lIconFn))
	L.SetField(mod, "Image", L.NewFunction(lImageFn))
	L.SetField(mod, "Rectangle", L.NewFunction(lRectangleFn))
	L.SetField(mod, "MultiLineEntry", L.NewFunction(lMultiLineEntryFn))
	L.SetField(mod, "PasswordEntry", L.NewFunction(lPasswordEntryFn))
	// Data widgets
	L.SetField(mod, "List", L.NewFunction(lListFn))
	L.SetField(mod, "Tree", L.NewFunction(lTreeFn))
	L.SetField(mod, "Table", L.NewFunction(lTableFn))
	// Layout
	L.SetField(mod, "HSplit", L.NewFunction(lHSplitFn))
	L.SetField(mod, "VSplit", L.NewFunction(lVSplitFn))
	// Toolbar
	L.SetField(mod, "Toolbar", L.NewFunction(lToolbarFn))
	// Card
	L.SetField(mod, "Card", L.NewFunction(lCardFn))
	// Accordion
	L.SetField(mod, "Accordion", L.NewFunction(lAccordionFn))
	// Form
	L.SetField(mod, "Form", L.NewFunction(lFormFn))

	// Preload so require("gui") works
	L.PreloadModule("gui", func(L *lua.LState) int {
		L.Push(mod)
		return 1
	})
	// Preload io, net, timer modules
	L.PreloadModule("fileio", func(L *lua.LState) int {
		L.Push(IOModule(L))
		return 1
	})
	L.PreloadModule("net", func(L *lua.LState) int {
		L.Push(NetModule(L))
		return 1
	})
	L.PreloadModule("timer", func(L *lua.LState) int {
		L.Push(TimerModule())
		return 1
	})
	L.PreloadModule("dialog", func(L *lua.LState) int {
		L.Push(DialogModule())
		return 1
	})
	// Clipboard module
	L.PreloadModule("clipboard", func(L *lua.LState) int {
		L.Push(clipboardModule(L))
		return 1
	})

	// Global app functions
	L.SetField(mod, "SetDarkMode", L.NewFunction(func(L *lua.LState) int {
		dark := lua.LVAsBool(L.Get(1))
		a.Settings().SetTheme(func() fyne.Theme {
			if dark {
				return theme.DarkTheme()
			}
			return theme.LightTheme()
		}())
		return 0
	}))
	L.SetField(mod, "IsDarkMode", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LBool(a.Settings().ThemeVariant() == theme.VariantDark))
		return 1
	}))

	return L
}

// ---------- helpers ----------

func pushWidget(L *lua.LState, w *LuaWidget) {
	ud := L.NewUserData()
	ud.Value = w
	L.SetMetatable(ud, L.GetTypeMetatable(widgetMT))
	L.Push(ud)
}

func checkWidget(L *lua.LState, n int) *LuaWidget {
	ud := L.CheckUserData(n)
	if w, ok := ud.Value.(*LuaWidget); ok {
		return w
	}
	L.ArgError(n, "fyne widget expected")
	return nil
}

// pushTabItem pushes a TabItem as userdata.
func pushTabItem(L *lua.LState, tab *container.TabItem) {
	ud := L.NewUserData()
	ud.Value = tab
	L.SetMetatable(ud, L.GetTypeMetatable(tabItemMT))
	L.Push(ud)
}

// checkTabItem checks that the argument is a TabItem userdata.
func checkTabItem(L *lua.LState, n int) *container.TabItem {
	ud := L.CheckUserData(n)
	if tab, ok := ud.Value.(*container.TabItem); ok {
		return tab
	}
	L.ArgError(n, "TabItem expected")
	return nil
}

// lCheckStringTable converts a Lua table at the given index to []string.
func lCheckStringTable(L *lua.LState, idx int) []string {
	t := L.CheckTable(idx)
	var result []string
	t.ForEach(func(_, v lua.LValue) {
		result = append(result, v.String())
	})
	return result
}

// getOrNil returns the CanvasObject for a widget argument, or nil if the arg is nil.
func getOrNil(L *lua.LState, idx int) fyne.CanvasObject {
	if L.Get(idx) == lua.LNil {
		return nil
	}
	return checkWidget(L, idx).Obj
}

// pushSizeTable pushes a {w, h} Lua table from a fyne.Size.
func pushSizeTable(L *lua.LState, s fyne.Size) {
	t := L.NewTable()
	t.RawSetString("w", lua.LNumber(s.Width))
	t.RawSetString("h", lua.LNumber(s.Height))
	L.Push(t)
}

// ---------- __index metamethod ----------

func widgetIndex(L *lua.LState) int {
	w := checkWidget(L, 1)
	method := L.CheckString(2)

	// Universal methods for all CanvasObject widgets (but not Window)
	if w.Obj != nil {
		switch method {
		case "Resize":
			L.Push(L.NewFunction(func(L *lua.LState) int {
				width := float32(L.CheckNumber(2))
				height := float32(L.CheckNumber(3))
				w.Obj.Resize(fyne.NewSize(width, height))
				return 0
			}))
			return 1
		case "Show":
			L.Push(L.NewFunction(func(L *lua.LState) int {
				w.Obj.Show()
				return 0
			}))
			return 1
		case "Hide":
			L.Push(L.NewFunction(func(L *lua.LState) int {
				w.Obj.Hide()
				return 0
			}))
			return 1
		case "Visible":
			L.Push(L.NewFunction(func(L *lua.LState) int {
				L.Push(lua.LBool(w.Obj.Visible()))
				return 1
			}))
			return 1
		case "Refresh":
			L.Push(L.NewFunction(func(L *lua.LState) int {
				w.Obj.Refresh()
				return 0
			}))
			return 1
		case "MinSize":
			L.Push(L.NewFunction(func(L *lua.LState) int {
				pushSizeTable(L, w.Obj.MinSize())
				return 1
			}))
			return 1
		case "Size":
			L.Push(L.NewFunction(func(L *lua.LState) int {
				pushSizeTable(L, w.Obj.Size())
				return 1
			}))
			return 1
		}
	}

	// Type-specific methods
	switch w.Type {
	case "Window":
		windowMethod(L, w, method)
	case "Button":
		buttonMethod(L, w, method)
	case "Label":
		labelMethod(L, w, method)
	case "Entry":
		entryMethod(L, w, method)
	case "VBox", "HBox":
		boxMethod(L, w, method)
	case "Slider":
		sliderMethod(L, w, method)
	case "Select":
		selectMethod(L, w, method)
	case "Check":
		checkMethod(L, w, method)
	case "RadioGroup":
		radioGroupMethod(L, w, method)
	case "ProgressBar":
		progressBarMethod(L, w, method)
	case "GridWrap":
		boxMethod(L, w, method) // same Add/Remove pattern as VBox/HBox
	case "Border":
		boxMethod(L, w, method) // same Add/Remove pattern (underlying *fyne.Container)
	case "Tabs", "AppTabs":
		tabsMethod(L, w, method)
	case "Scroll":
		scrollMethod(L, w, method)
	case "Icon":
		// Icon has no methods currently
		L.Push(lua.LNil)
	case "Image":
		imageMethod(L, w, method)
	case "Rectangle":
		rectangleMethod(L, w, method)
	case "List":
		listMethod(L, w, method)
	case "Tree":
		treeMethod(L, w, method)
	case "Table":
		tableMethod(L, w, method)
	case "Toolbar":
		// Toolbar has no methods currently
		L.Push(lua.LNil)
	case "Form":
		formMethod(L, w, method)
	case "Card":
		cardMethod(L, w, method)
	case "Accordion":
		accordionMethod(L, w, method)
	case "Split":
		splitMethod(L, w, method)
	default:
		L.Push(lua.LNil)
	}
	return 1
}

// ---------- TabItem __index metamethod ----------

func tabItemIndex(L *lua.LState) int {
	_ = checkTabItem(L, 1)
	method := L.CheckString(2)
	// TabItem currently has no methods exposed to Lua; return nil
	if method == "" {
		L.Push(lua.LNil)
	}
	return 1
}

// ---------- constructors ----------

// ActiveWindow stores the currently active window for hot-reload reuse.
var ActiveWindow fyne.Window

// extraWindows stores additional windows created by Lua via gui.NewWindow().
// These are closed on hot-reload.
var extraWindows []fyne.Window

// ClearExtraWindows closes and removes all extra windows. Called before hot-reload.
func ClearExtraWindows() {
	for _, w := range extraWindows {
		w.Close()
	}
	extraWindows = nil
}

func lWindowFn(a fyne.App) lua.LGFunction {
	return func(L *lua.LState) int {
		title := L.CheckString(1)
		width := float32(L.CheckNumber(2))
		height := float32(L.CheckNumber(3))

		var win fyne.Window
		if ActiveWindow != nil {
			// Reuse existing window on hot-reload
			win = ActiveWindow
			win.SetTitle(title)
			win.Resize(fyne.NewSize(width, height))
		} else {
			win = a.NewWindow(title)
			win.Resize(fyne.NewSize(width, height))
		}
		ActiveWindow = win
		pushWidget(L, &LuaWidget{Win: win, Type: "Window"})
		return 1
	}
}

// lNewWindowFn creates a new independent window (not reused on hot-reload).
// gui.NewWindow("标题", w, h) -- creates a fresh window every call.
func lNewWindowFn(a fyne.App) lua.LGFunction {
	return func(L *lua.LState) int {
		title := L.CheckString(1)
		width := float32(L.CheckNumber(2))
		height := float32(L.CheckNumber(3))
		win := a.NewWindow(title)
		win.Resize(fyne.NewSize(width, height))
		extraWindows = append(extraWindows, win)
		pushWidget(L, &LuaWidget{Win: win, Type: "Window"})
		return 1
	}
}

func lButtonFn(L *lua.LState) int {
	text := L.CheckString(1)
	btn := widget.NewButton(text, func() {})
	pushWidget(L, &LuaWidget{Obj: btn, Type: "Button"})
	return 1
}

func lLabelFn(L *lua.LState) int {
	text := L.CheckString(1)
	lbl := widget.NewLabel(text)
	pushWidget(L, &LuaWidget{Obj: lbl, Type: "Label"})
	return 1
}

func lEntryFn(L *lua.LState) int {
	entry := widget.NewEntry()
	pushWidget(L, &LuaWidget{Obj: entry, Type: "Entry"})
	return 1
}

func lVBoxFn(L *lua.LState) int {
	var objs []fyne.CanvasObject
	n := L.GetTop()
	for i := 1; i <= n; i++ {
		w := checkWidget(L, i)
		if w.Obj != nil {
			objs = append(objs, w.Obj)
		}
	}
	pushWidget(L, &LuaWidget{Obj: container.NewVBox(objs...), Type: "VBox"})
	return 1
}

func lHBoxFn(L *lua.LState) int {
	var objs []fyne.CanvasObject
	n := L.GetTop()
	for i := 1; i <= n; i++ {
		w := checkWidget(L, i)
		if w.Obj != nil {
			objs = append(objs, w.Obj)
		}
	}
	pushWidget(L, &LuaWidget{Obj: container.NewHBox(objs...), Type: "HBox"})
	return 1
}

func lSliderFn(L *lua.LState) int {
	minVal := float64(L.CheckNumber(1))
	maxVal := float64(L.CheckNumber(2))
	s := widget.NewSlider(minVal, maxVal)
	pushWidget(L, &LuaWidget{Obj: s, Type: "Slider"})
	return 1
}

func lSelectFn(L *lua.LState) int {
	options := lCheckStringTable(L, 1)
	s := widget.NewSelect(options, nil)
	pushWidget(L, &LuaWidget{Obj: s, Type: "Select"})
	return 1
}

func lCheckFn(L *lua.LState) int {
	label := L.CheckString(1)
	c := widget.NewCheck(label, nil)
	pushWidget(L, &LuaWidget{Obj: c, Type: "Check"})
	return 1
}

func lRadioGroupFn(L *lua.LState) int {
	options := lCheckStringTable(L, 1)
	r := widget.NewRadioGroup(options, nil)
	pushWidget(L, &LuaWidget{Obj: r, Type: "RadioGroup"})
	return 1
}

func lProgressBarFn(L *lua.LState) int {
	p := widget.NewProgressBar()
	pushWidget(L, &LuaWidget{Obj: p, Type: "ProgressBar"})
	return 1
}

func lSeparatorFn(L *lua.LState) int {
	s := widget.NewSeparator()
	pushWidget(L, &LuaWidget{Obj: s, Type: "Separator"})
	return 1
}

func lHyperlinkFn(L *lua.LState) int {
	text := L.CheckString(1)
	rawURL := L.CheckString(2)
	u, _ := url.Parse(rawURL)
	h := widget.NewHyperlink(text, u)
	pushWidget(L, &LuaWidget{Obj: h, Type: "Hyperlink"})
	return 1
}

func lGridWrapFn(L *lua.LState) int {
	cols := float32(L.CheckNumber(1))
	var objs []fyne.CanvasObject
	n := L.GetTop()
	for i := 2; i <= n; i++ {
		w := checkWidget(L, i)
		if w.Obj != nil {
			objs = append(objs, w.Obj)
		}
	}
	pushWidget(L, &LuaWidget{Obj: container.NewGridWrap(fyne.NewSize(cols, cols), objs...), Type: "GridWrap"})
	return 1
}

func lBorderFn(L *lua.LState) int {
	top := getOrNil(L, 1)
	bottom := getOrNil(L, 2)
	left := getOrNil(L, 3)
	right := getOrNil(L, 4)
	center := getOrNil(L, 5)
	pushWidget(L, &LuaWidget{Obj: container.NewBorder(top, bottom, left, right, center), Type: "Border"})
	return 1
}

func lTabsFn(isAppTabs bool) lua.LGFunction {
	return func(L *lua.LState) int {
		n := L.GetTop()
		var tabs []*container.TabItem
		for i := 1; i <= n; i++ {
			tab := checkTabItem(L, i)
			tabs = append(tabs, tab)
		}
		// In Fyne v2.7, both Tabs and AppTabs use container.NewAppTabs
		obj := container.NewAppTabs(tabs...)
		typeName := "Tabs"
		if isAppTabs {
			typeName = "AppTabs"
		}
		pushWidget(L, &LuaWidget{Obj: obj, Type: typeName})
		return 1
	}
}

func lTabItemFn(L *lua.LState) int {
	label := L.CheckString(1)
	w := checkWidget(L, 2)
	tab := container.NewTabItem(label, w.Obj)
	pushTabItem(L, tab)
	return 1
}

// ---------- Window methods ----------

func windowMethod(L *lua.LState, w *LuaWidget, method string) {
	switch method {
	case "SetContent":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			child := checkWidget(L, 2)
			if child.Obj != nil {
				w.Win.SetContent(child.Obj)
			}
			return 0
		}))
	case "SetTitle":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			w.Win.SetTitle(L.CheckString(2))
			return 0
		}))
	case "Show":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			w.Win.Show()
			return 0
		}))
	case "ShowAndRun":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			if HotReload {
				w.Win.Show()
			} else {
				w.Win.ShowAndRun()
			}
			return 0
		}))
	case "OnClosed":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			fn := L.CheckFunction(2)
			w.Win.SetCloseIntercept(func() {
				if CurrentL != nil {
					CurrentL.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true})
				}
				w.Win.Close()
			})
			return 0
		}))
	case "CenterOnScreen":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			w.Win.CenterOnScreen()
			return 0
		}))
	case "SetFixedSize":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			fixed := lua.LVAsBool(L.Get(2))
			w.Win.SetFixedSize(fixed)
			return 0
		}))
	case "SetMainMenu":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			menu := checkLuaMenu(L, 2)
			var fyneMenu *fyne.MainMenu
			if menu != nil {
				fyneMenu = buildFyneMenu(L, menu)
			}
			w.Win.SetMainMenu(fyneMenu)
			return 0
		}))
	default:
		L.Push(lua.LNil)
	}
}

// ---------- Button methods ----------

func buttonMethod(L *lua.LState, w *LuaWidget, method string) {
	switch method {
	case "SetText":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			w.Obj.(*widget.Button).SetText(L.CheckString(2))
			return 0
		}))
	case "Text":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			L.Push(lua.LString(w.Obj.(*widget.Button).Text))
			return 1
		}))
	case "OnClick":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			fn := L.CheckFunction(2)
			btn := w.Obj.(*widget.Button)
			btn.OnTapped = func() {
				if CurrentL != nil {
					CurrentL.CallByParam(lua.P{
					Fn:      fn,
					NRet:    0,
					Protect: true,
					})
				}
			}
			return 0
		}))
	case "SetImportance":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			level := L.CheckString(2)
			btn := w.Obj.(*widget.Button)
			switch level {
			case "high":
				btn.Importance = widget.HighImportance
			case "medium":
				btn.Importance = widget.MediumImportance
			case "low":
				btn.Importance = widget.LowImportance
			}
			return 0
		}))
	case "Disable":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			w.Obj.(*widget.Button).Disable()
			return 0
		}))
	case "Enable":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			w.Obj.(*widget.Button).Enable()
			return 0
		}))
	default:
		L.Push(lua.LNil)
	}
}

// ---------- Label methods ----------

func labelMethod(L *lua.LState, w *LuaWidget, method string) {
	switch method {
	case "SetText":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			w.Obj.(*widget.Label).SetText(L.CheckString(2))
			return 0
		}))
	case "Text":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			L.Push(lua.LString(w.Obj.(*widget.Label).Text))
			return 1
		}))
	default:
		L.Push(lua.LNil)
	}
}

// ---------- Entry methods ----------

func entryMethod(L *lua.LState, w *LuaWidget, method string) {
	switch method {
	case "SetText":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			w.Obj.(*widget.Entry).SetText(L.CheckString(2))
			return 0
		}))
	case "Text":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			L.Push(lua.LString(w.Obj.(*widget.Entry).Text))
			return 1
		}))
	case "SetPlaceHolder":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			w.Obj.(*widget.Entry).SetPlaceHolder(L.CheckString(2))
			return 0
		}))
		// Resize is now a universal method handled in widgetIndex
	case "OnChanged":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			fn := L.CheckFunction(2)
			entry := w.Obj.(*widget.Entry)
			entry.OnChanged = func(s string) {
				if CurrentL != nil { CurrentL.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}, lua.LString(s)) }
			}
			return 0
		}))
	case "OnSubmitted":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			fn := L.CheckFunction(2)
			entry := w.Obj.(*widget.Entry)
			entry.OnSubmitted = func(s string) {
				if CurrentL != nil { CurrentL.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}, lua.LString(s)) }
			}
			return 0
		}))
	default:
		L.Push(lua.LNil)
	}
}

// ---------- VBox / HBox methods ----------

func boxMethod(L *lua.LState, w *LuaWidget, method string) {
	switch method {
	case "Add":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			child := checkWidget(L, 2)
			if child.Obj != nil {
				if c, ok := w.Obj.(*fyne.Container); ok {
					c.Add(child.Obj)
				}
			}
			return 0
		}))
	default:
		L.Push(lua.LNil)
	}
}

// ---------- Slider methods ----------

func sliderMethod(L *lua.LState, w *LuaWidget, method string) {
	switch method {
	case "SetValue":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			v := float64(L.CheckNumber(2))
			w.Obj.(*widget.Slider).SetValue(v)
			return 0
		}))
	case "Value":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			L.Push(lua.LNumber(w.Obj.(*widget.Slider).Value))
			return 1
		}))
	case "OnChanged":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			fn := L.CheckFunction(2)
			s := w.Obj.(*widget.Slider)
			s.OnChanged = func(v float64) {
				if CurrentL != nil { CurrentL.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}, lua.LNumber(v)) }
			}
			return 0
		}))
	default:
		L.Push(lua.LNil)
	}
}

// ---------- Select methods ----------

func selectMethod(L *lua.LState, w *LuaWidget, method string) {
	switch method {
	case "SetSelected":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			w.Obj.(*widget.Select).SetSelected(L.CheckString(2))
			return 0
		}))
	case "Selected":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			L.Push(lua.LString(w.Obj.(*widget.Select).Selected))
			return 1
		}))
	case "OnChanged":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			fn := L.CheckFunction(2)
			s := w.Obj.(*widget.Select)
			s.OnChanged = func(s string) {
				if CurrentL != nil { CurrentL.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}, lua.LString(s)) }
			}
			return 0
		}))
	default:
		L.Push(lua.LNil)
	}
}

// ---------- Check methods ----------

func checkMethod(L *lua.LState, w *LuaWidget, method string) {
	switch method {
	case "SetChecked":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			w.Obj.(*widget.Check).SetChecked(lua.LVAsBool(L.Get(2)))
			return 0
		}))
	case "Checked":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			L.Push(lua.LBool(w.Obj.(*widget.Check).Checked))
			return 1
		}))
	case "OnChanged":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			fn := L.CheckFunction(2)
			c := w.Obj.(*widget.Check)
			c.OnChanged = func(checked bool) {
				if CurrentL != nil { CurrentL.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}, lua.LBool(checked)) }
			}
			return 0
		}))
	default:
		L.Push(lua.LNil)
	}
}

// ---------- RadioGroup methods ----------

func radioGroupMethod(L *lua.LState, w *LuaWidget, method string) {
	switch method {
	case "SetSelected":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			w.Obj.(*widget.RadioGroup).SetSelected(L.CheckString(2))
			return 0
		}))
	case "Selected":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			L.Push(lua.LString(w.Obj.(*widget.RadioGroup).Selected))
			return 1
		}))
	case "OnChanged":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			fn := L.CheckFunction(2)
			r := w.Obj.(*widget.RadioGroup)
			r.OnChanged = func(s string) {
				if CurrentL != nil { CurrentL.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}, lua.LString(s)) }
			}
			return 0
		}))
	default:
		L.Push(lua.LNil)
	}
}

// ---------- ProgressBar methods ----------

func progressBarMethod(L *lua.LState, w *LuaWidget, method string) {
	switch method {
	case "SetValue":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			v := float64(L.CheckNumber(2))
			w.Obj.(*widget.ProgressBar).SetValue(v)
			return 0
		}))
	case "Value":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			L.Push(lua.LNumber(w.Obj.(*widget.ProgressBar).Value))
			return 1
		}))
	case "SetMax":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			v := float64(L.CheckNumber(2))
			w.Obj.(*widget.ProgressBar).Max = v
			return 0
		}))
	case "SetMin":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			v := float64(L.CheckNumber(2))
			w.Obj.(*widget.ProgressBar).Min = v
			return 0
		}))
	default:
		L.Push(lua.LNil)
	}
}

// ---------- Hyperlink methods ----------

func hyperlinkMethod(L *lua.LState, w *LuaWidget, method string) {
	switch method {
	case "SetText":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			w.Obj.(*widget.Hyperlink).SetText(L.CheckString(2))
			return 0
		}))
	default:
		L.Push(lua.LNil)
	}
}

// ---------- Tabs / AppTabs methods ----------

func tabsMethod(L *lua.LState, w *LuaWidget, method string) {
	switch method {
	case "Append":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			tab := checkTabItem(L, 2)
			if tc, ok := w.Obj.(*container.AppTabs); ok {
				tc.Append(tab)
			}
			return 0
		}))
	case "SelectTab":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			tab := checkTabItem(L, 2)
			if tc, ok := w.Obj.(*container.AppTabs); ok {
				tc.SelectTab(tab)
			}
			return 0
		}))
	default:
		L.Push(lua.LNil)
	}
}
