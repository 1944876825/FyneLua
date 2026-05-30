package bridge

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/yuin/gopher-lua"
)

// ==================== List ====================
// gui.List(lengthFn, createFn [, onSelectFn])
//   lengthFn: function() -> number of items
//   createFn: function(index) -> CanvasObject (called to create template AND update)
//   onSelectFn: function(index) -- called on item tap (optional)
func lListFn(L *lua.LState) int {
	lengthFn := L.CheckFunction(1)
	createFn := L.CheckFunction(2)
	var onSelect *lua.LFunction
	if L.GetTop() >= 3 && L.Get(3) != lua.LNil {
		onSelect = L.CheckFunction(3)
	}

	list := widget.NewList(
		func() int {
			if CurrentL == nil {
				return 0
			}
			CurrentL.CallByParam(lua.P{Fn: lengthFn, NRet: 1, Protect: true})
			n, ok := CurrentL.Get(-1).(lua.LNumber)
			CurrentL.Pop(1)
			if !ok {
				return 0
			}
			return int(n)
		},
		func() fyne.CanvasObject {
			if CurrentL == nil {
				return widget.NewLabel("")
			}
			CurrentL.Push(lua.LNumber(0))
			CurrentL.CallByParam(lua.P{Fn: createFn, NRet: 1, Protect: true})
			ret := CurrentL.Get(-1)
			CurrentL.Pop(1)
			obj := extractObj(ret)
			if obj == nil {
				obj = widget.NewLabel("")
			}
			return obj
		},
		func(i widget.ListItemID, obj fyne.CanvasObject) {
			if CurrentL == nil {
				return
			}
			CurrentL.Push(lua.LNumber(i))
			CurrentL.CallByParam(lua.P{Fn: createFn, NRet: 1, Protect: true})
			ret := CurrentL.Get(-1)
			CurrentL.Pop(1)
			newObj := extractObj(ret)
			if newObj != nil {
				syncWidget(obj, newObj)
			}
		},
	)

	if onSelect != nil {
		list.OnSelected = func(id widget.ListItemID) {
			if CurrentL != nil {
				CurrentL.CallByParam(lua.P{Fn: onSelect, NRet: 0, Protect: true}, lua.LNumber(id))
			}
		}
	}

	pushWidget(L, &LuaWidget{Obj: list, Type: "List"})
	return 1
}

// List methods
func listMethod(L *lua.LState, w *LuaWidget, method string) {
	switch method {
	case "Refresh":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			if list, ok := w.Obj.(*widget.List); ok {
				list.Refresh()
			}
			return 0
		}))
	default:
		L.Push(lua.LNil)
	}
}

// ==================== Tree ====================
// gui.Tree(childUIDsFn, isBranchFn, createFn [, onSelectFn])
//   childUIDsFn: function(uid) -> table of child uids
//   isBranchFn: function(uid) -> bool
//   createFn: function(isBranch) -> CanvasObject
//   onSelectFn: function(uid) -- called on item tap (optional)
func lTreeFn(L *lua.LState) int {
	childFn := L.CheckFunction(1)
	isBranchFn := L.CheckFunction(2)
	createFn := L.CheckFunction(3)
	var onSelect *lua.LFunction
	if L.GetTop() >= 4 && L.Get(3+1) != lua.LNil {
		onSelect = L.CheckFunction(4)
	}

	tree := widget.NewTree(
		func(uid widget.TreeNodeID) []widget.TreeNodeID {
			if CurrentL == nil {
				return nil
			}
			CurrentL.Push(lua.LString(uid))
			CurrentL.CallByParam(lua.P{Fn: childFn, NRet: 1, Protect: true})
			ret := CurrentL.Get(-1)
			CurrentL.Pop(1)
			if tbl, ok := ret.(*lua.LTable); ok {
				var ids []widget.TreeNodeID
				tbl.ForEach(func(_, v lua.LValue) {
					ids = append(ids, widget.TreeNodeID(v.String()))
				})
				return ids
			}
			return nil
		},
		func(uid widget.TreeNodeID) bool {
			if CurrentL == nil {
				return false
			}
			CurrentL.Push(lua.LString(uid))
			CurrentL.CallByParam(lua.P{Fn: isBranchFn, NRet: 1, Protect: true})
			ret := CurrentL.Get(-1)
			CurrentL.Pop(1)
			return lua.LVAsBool(ret)
		},
		func(isBranch bool) fyne.CanvasObject {
			if CurrentL == nil {
				return widget.NewLabel("")
			}
			CurrentL.Push(lua.LBool(isBranch))
			CurrentL.CallByParam(lua.P{Fn: createFn, NRet: 1, Protect: true})
			ret := CurrentL.Get(-1)
			CurrentL.Pop(1)
			obj := extractObj(ret)
			if obj == nil {
				return widget.NewLabel("")
			}
			return obj
		},
		func(uid widget.TreeNodeID, isBranch bool, obj fyne.CanvasObject) {
			if CurrentL == nil {
				return
			}
			CurrentL.Push(lua.LBool(isBranch))
			CurrentL.CallByParam(lua.P{Fn: createFn, NRet: 1, Protect: true})
			ret := CurrentL.Get(-1)
			CurrentL.Pop(1)
			newObj := extractObj(ret)
			if newObj != nil {
				syncWidget(obj, newObj)
			}
		},
	)

	if onSelect != nil {
		tree.OnSelected = func(uid widget.TreeNodeID) {
			if CurrentL != nil {
				CurrentL.CallByParam(lua.P{Fn: onSelect, NRet: 0, Protect: true}, lua.LString(uid))
			}
		}
	}

	pushWidget(L, &LuaWidget{Obj: tree, Type: "Tree"})
	return 1
}

// Tree methods (Fyne v2.7 uses OpenBranch/CloseBranch/OpenAllBranches/CloseAllBranches)
func treeMethod(L *lua.LState, w *LuaWidget, method string) {
	switch method {
	case "Open":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			uid := L.CheckString(2)
			if tree, ok := w.Obj.(*widget.Tree); ok {
				tree.OpenBranch(uid)
			}
			return 0
		}))
	case "Close":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			uid := L.CheckString(2)
			if tree, ok := w.Obj.(*widget.Tree); ok {
				tree.CloseBranch(uid)
			}
			return 0
		}))
	case "OpenAll":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			if tree, ok := w.Obj.(*widget.Tree); ok {
				tree.OpenAllBranches()
			}
			return 0
		}))
	case "CloseAll":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			if tree, ok := w.Obj.(*widget.Tree); ok {
				tree.CloseAllBranches()
			}
			return 0
		}))
	case "IsOpen":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			uid := L.CheckString(2)
			if tree, ok := w.Obj.(*widget.Tree); ok {
				L.Push(lua.LBool(tree.IsBranchOpen(uid)))
			} else {
				L.Push(lua.LBool(false))
			}
			return 1
		}))
	default:
		L.Push(lua.LNil)
	}
}

// ==================== Table ====================
// gui.Table(lengthFn, createFn [, onSelectFn])
//   lengthFn: function() -> rows, cols  (returns two values)
//   createFn: function(col, row) -> CanvasObject
//   onSelectFn: function(col, row) -- called on cell tap (optional)
func lTableFn(L *lua.LState) int {
	lengthFn := L.CheckFunction(1)
	createFn := L.CheckFunction(2)
	var onSelect *lua.LFunction
	if L.GetTop() >= 3 && L.Get(3) != lua.LNil {
		onSelect = L.CheckFunction(3)
	}

	tableWidget := widget.NewTable(
		func() (int, int) {
			if CurrentL == nil {
				return 0, 0
			}
			CurrentL.CallByParam(lua.P{Fn: lengthFn, NRet: 2, Protect: true})
			ret2 := CurrentL.Get(-1)
			ret1 := CurrentL.Get(-2)
			CurrentL.Pop(2)
			rows := int(lua.LVAsNumber(ret1))
			cols := int(lua.LVAsNumber(ret2))
			return rows, cols
		},
		func() fyne.CanvasObject {
			if CurrentL == nil {
				return widget.NewLabel("")
			}
			CurrentL.Push(lua.LNumber(0))
			CurrentL.Push(lua.LNumber(0))
			CurrentL.CallByParam(lua.P{Fn: createFn, NRet: 1, Protect: true})
			ret := CurrentL.Get(-1)
			CurrentL.Pop(1)
			obj := extractObj(ret)
			if obj == nil {
				return widget.NewLabel("")
			}
			return obj
		},
		func(cell widget.TableCellID, obj fyne.CanvasObject) {
			if CurrentL == nil {
				return
			}
			CurrentL.Push(lua.LNumber(cell.Col))
			CurrentL.Push(lua.LNumber(cell.Row))
			CurrentL.CallByParam(lua.P{Fn: createFn, NRet: 1, Protect: true})
			ret := CurrentL.Get(-1)
			CurrentL.Pop(1)
			newObj := extractObj(ret)
			if newObj != nil {
				syncWidget(obj, newObj)
			}
		},
	)

	if onSelect != nil {
		tableWidget.OnSelected = func(id widget.TableCellID) {
			if CurrentL != nil {
				CurrentL.CallByParam(lua.P{Fn: onSelect, NRet: 0, Protect: true}, lua.LNumber(id.Col), lua.LNumber(id.Row))
			}
		}
	}

	pushWidget(L, &LuaWidget{Obj: tableWidget, Type: "Table"})
	return 1
}

// Table methods
func tableMethod(L *lua.LState, w *LuaWidget, method string) {
	switch method {
	case "SetColumnWidth":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			col := int(L.CheckNumber(2))
			width := float32(L.CheckNumber(3))
			if tw, ok := w.Obj.(*widget.Table); ok {
				tw.SetColumnWidth(col, width)
			}
			return 0
		}))
	default:
		L.Push(lua.LNil)
	}
}

// ==================== Toolbar ====================
// gui.Toolbar(items)
// items: {{icon="name", action=function()}, {separator=true}, ...}
func lToolbarFn(L *lua.LState) int {
	tbl := L.CheckTable(1)
	var items []widget.ToolbarItem

	tbl.ForEach(func(_, v lua.LValue) {
		item, ok := v.(*lua.LTable)
		if !ok {
			return
		}

		if sep := item.RawGetString("separator"); sep == lua.LTrue {
			items = append(items, widget.NewToolbarSeparator())
			return
		}

		iconName := item.RawGetString("icon")
		actionVal := item.RawGetString("action")

		var icon fyne.Resource
		if iconName != lua.LNil && iconName.String() != "" {
			icon = resolveIcon(iconName.String())
		}

		var action func()
		if fn, ok := actionVal.(*lua.LFunction); ok {
			action = func() {
				if CurrentL != nil {
					CurrentL.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true})
				}
			}
		}

		items = append(items, widget.NewToolbarAction(icon, action))
	})

	pushWidget(L, &LuaWidget{Obj: widget.NewToolbar(items...), Type: "Toolbar"})
	return 1
}

// ==================== Form ====================
// gui.Form(items)
// items: {{label="标签", widget=entry}, ...}
func lFormFn(L *lua.LState) int {
	tbl := L.CheckTable(1)
	var formItems []*widget.FormItem

	tbl.ForEach(func(_, v lua.LValue) {
		item, ok := v.(*lua.LTable)
		if !ok {
			return
		}
		label := item.RawGetString("label")
		widgetVal := item.RawGetString("widget")

		var w fyne.CanvasObject
		if ud, ok := widgetVal.(*lua.LUserData); ok {
			if lw, ok := ud.Value.(*LuaWidget); ok && lw.Obj != nil {
				w = lw.Obj
			}
		}

		text := ""
		if label != lua.LNil {
			text = label.String()
		}
		formItems = append(formItems, widget.NewFormItem(text, w))
	})

	pushWidget(L, &LuaWidget{Obj: widget.NewForm(formItems...), Type: "Form"})
	return 1
}

// Form methods
func formMethod(L *lua.LState, w *LuaWidget, method string) {
	switch method {
	case "OnSubmit":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			fn := L.CheckFunction(2)
			f := w.Obj.(*widget.Form)
			f.OnSubmit = func() {
				if CurrentL != nil {
					CurrentL.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true})
				}
			}
			return 0
		}))
	case "OnCancel":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			fn := L.CheckFunction(2)
			f := w.Obj.(*widget.Form)
			f.OnCancel = func() {
				if CurrentL != nil {
					CurrentL.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true})
				}
			}
			return 0
		}))
	case "Append":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			label := L.CheckString(2)
			child := checkWidget(L, 3)
			f := w.Obj.(*widget.Form)
			if child.Obj != nil {
				f.AppendItem(widget.NewFormItem(label, child.Obj))
				f.Refresh()
			}
			return 0
		}))
	default:
		L.Push(lua.LNil)
	}
}

// ==================== Card ====================
// gui.Card(title, subtitle, content1, content2, ...)
func lCardFn(L *lua.LState) int {
	title := ""
	subtitle := ""
	if L.GetTop() >= 1 {
		title = L.CheckString(1)
	}
	if L.GetTop() >= 2 {
		subtitle = L.CheckString(2)
	}

	var objs []fyne.CanvasObject
	for i := 3; i <= L.GetTop(); i++ {
		w := checkWidget(L, i)
		if w.Obj != nil {
			objs = append(objs, w.Obj)
		}
	}

	var content fyne.CanvasObject
	if len(objs) > 0 {
		content = container.NewVBox(objs...)
	} else {
		content = widget.NewLabel("")
	}

	pushWidget(L, &LuaWidget{Obj: widget.NewCard(title, subtitle, content), Type: "Card"})
	return 1
}

// Card methods
func cardMethod(L *lua.LState, w *LuaWidget, method string) {
	switch method {
	case "SetTitle":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			w.Obj.(*widget.Card).SetTitle(L.CheckString(2))
			return 0
		}))
	case "SetSubTitle":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			w.Obj.(*widget.Card).SetSubTitle(L.CheckString(2))
			return 0
		}))
	case "Title":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			L.Push(lua.LString(w.Obj.(*widget.Card).Title))
			return 1
		}))
	case "SubTitle":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			L.Push(lua.LString(w.Obj.(*widget.Card).Subtitle))
			return 1
		}))
	default:
		L.Push(lua.LNil)
	}
}

// ==================== Accordion ====================
// gui.Accordion(items)
// items: {{title="标题", content=widget, open=true/false}, ...}
func lAccordionFn(L *lua.LState) int {
	tbl := L.CheckTable(1)
	var accItems []*widget.AccordionItem
	var openIndices []int
	idx := 0

	tbl.ForEach(func(_, v lua.LValue) {
		item, ok := v.(*lua.LTable)
		if !ok {
			return
		}
		title := item.RawGetString("title")
		contentVal := item.RawGetString("content")
		openVal := item.RawGetString("open")

		var content fyne.CanvasObject
		if ud, ok := contentVal.(*lua.LUserData); ok {
			if lw, ok := ud.Value.(*LuaWidget); ok && lw.Obj != nil {
				content = lw.Obj
			}
		}
		if content == nil {
			content = widget.NewLabel("")
		}

		text := ""
		if title != lua.LNil {
			text = title.String()
		}

		accItems = append(accItems, widget.NewAccordionItem(text, content))
		if openVal == lua.LTrue {
			openIndices = append(openIndices, idx)
		}
		idx++
	})

	acc := widget.NewAccordion(accItems...)
	for _, i := range openIndices {
		acc.Open(i)
	}

	pushWidget(L, &LuaWidget{Obj: acc, Type: "Accordion"})
	return 1
}

// Accordion methods
func accordionMethod(L *lua.LState, w *LuaWidget, method string) {
	switch method {
	case "Open":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			idx := int(L.CheckNumber(2))
			if acc, ok := w.Obj.(*widget.Accordion); ok {
				if idx >= 0 && idx < len(acc.Items) {
					acc.Open(idx)
				}
			}
			return 0
		}))
	case "Close":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			idx := int(L.CheckNumber(2))
			if acc, ok := w.Obj.(*widget.Accordion); ok {
				if idx >= 0 && idx < len(acc.Items) {
					acc.Close(idx)
				}
			}
			return 0
		}))
	case "OpenAll":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			if acc, ok := w.Obj.(*widget.Accordion); ok {
				acc.OpenAll()
			}
			return 0
		}))
	case "CloseAll":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			if acc, ok := w.Obj.(*widget.Accordion); ok {
				acc.CloseAll()
			}
			return 0
		}))
	default:
		L.Push(lua.LNil)
	}
}

// ==================== Split ====================
// gui.HSplit(left, right)
// gui.VSplit(top, bottom)
func lHSplitFn(L *lua.LState) int {
	var left, right fyne.CanvasObject
	if w := checkWidget(L, 1); w.Obj != nil {
		left = w.Obj
	}
	if w := checkWidget(L, 2); w.Obj != nil {
		right = w.Obj
	}
	pushWidget(L, &LuaWidget{Obj: container.NewHSplit(left, right), Type: "Split"})
	return 1
}

func lVSplitFn(L *lua.LState) int {
	var top, bottom fyne.CanvasObject
	if w := checkWidget(L, 1); w.Obj != nil {
		top = w.Obj
	}
	if w := checkWidget(L, 2); w.Obj != nil {
		bottom = w.Obj
	}
	pushWidget(L, &LuaWidget{Obj: container.NewVSplit(top, bottom), Type: "Split"})
	return 1
}

// Split methods
func splitMethod(L *lua.LState, w *LuaWidget, method string) {
	switch method {
	case "SetOffset":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			offset := float64(L.CheckNumber(2))
			if s, ok := w.Obj.(*container.Split); ok {
				s.SetOffset(offset)
			}
			return 0
		}))
	case "Offset":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			if s, ok := w.Obj.(*container.Split); ok {
				L.Push(lua.LNumber(s.Offset))
			} else {
				L.Push(lua.LNumber(0))
			}
			return 1
		}))
	default:
		L.Push(lua.LNil)
	}
}

// ==================== Clipboard ====================
// gui.Clipboard.Set(text)
// gui.Clipboard.Get() -> string
func clipboardModule(L *lua.LState) *lua.LTable {
	mod := L.NewTable()
	wins := fyne.CurrentApp().Driver().AllWindows()
	getWin := func() fyne.Window {
		if len(wins) > 0 {
			return wins[0]
		}
		return nil
	}

	L.SetField(mod, "Set", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		if w := getWin(); w != nil {
			w.Clipboard().SetContent(text)
		}
		return 0
	}))

	L.SetField(mod, "Get", L.NewFunction(func(L *lua.LState) int {
		if w := getWin(); w != nil {
			L.Push(lua.LString(w.Clipboard().Content()))
		} else {
			L.Push(lua.LString(""))
		}
		return 1
	}))

	return mod
}

// ==================== Helpers ====================

// extractObj extracts a fyne.CanvasObject from a Lua value (userdata wrapping LuaWidget).
func extractObj(v lua.LValue) fyne.CanvasObject {
	if ud, ok := v.(*lua.LUserData); ok {
		if lw, ok := ud.Value.(*LuaWidget); ok && lw.Obj != nil {
			return lw.Obj
		}
	}
	return nil
}

// syncWidget copies text/content from a template/new widget to the live widget.
func syncWidget(live, template fyne.CanvasObject) {
	// Label
	if liveLabel, ok := live.(*widget.Label); ok {
		if t, ok := template.(*widget.Label); ok {
			liveLabel.SetText(t.Text)
			return
		}
	}
	// Button
	if liveBtn, ok := live.(*widget.Button); ok {
		if t, ok := template.(*widget.Button); ok {
			liveBtn.SetText(t.Text)
			return
		}
	}
	// Entry
	if liveEntry, ok := live.(*widget.Entry); ok {
		if t, ok := template.(*widget.Entry); ok {
			liveEntry.SetText(t.Text)
			return
		}
	}
	// Hyperlink
	if liveHL, ok := live.(*widget.Hyperlink); ok {
		if t, ok := template.(*widget.Hyperlink); ok {
			liveHL.SetText(t.Text)
			return
		}
	}
	// Icon
	if liveIcon, ok := live.(*widget.Icon); ok {
		if t, ok := template.(*widget.Icon); ok {
			liveIcon.SetResource(t.Resource)
			return
		}
	}
	// Checkbox
	if liveChk, ok := live.(*widget.Check); ok {
		if t, ok := template.(*widget.Check); ok {
			liveChk.SetChecked(t.Checked)
			return
		}
	}
	// Fallback: just refresh
	live.Refresh()
}
