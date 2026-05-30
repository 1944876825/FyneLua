package bridge

import (
	"fyne.io/fyne/v2"
	"github.com/yuin/gopher-lua"
)

// LuaMenuItem represents a menu structure from Lua:
// { label = "File", items = {
//   { label = "Open", action = function() end },
//   { label = "Save", action = function() end },
//   { label = "-" },  -- separator
//   { label = "Quit", action = function() end },
// }}
type LuaMenuItem struct {
	Label  string
	Items  []*LuaMenuItem
	Action *lua.LFunction
}

// checkLuaMenu parses a Lua table into a menu structure.
// Expected: { {label="File", items={{label="Open", action=fn}, ...}}, ... }
func checkLuaMenu(L *lua.LState, idx int) []*LuaMenuItem {
	t := L.CheckTable(idx)
	var menus []*LuaMenuItem
	t.ForEach(func(_, v lua.LValue) {
		if tbl, ok := v.(*lua.LTable); ok {
			menus = append(menus, parseMenuItem(L, tbl))
		}
	})
	return menus
}

func parseMenuItem(L *lua.LState, t *lua.LTable) *LuaMenuItem {
	item := &LuaMenuItem{}
	if v := t.RawGetString("label"); v != lua.LNil {
		item.Label = v.String()
	}
	if v := t.RawGetString("action"); v != nil {
		if fn, ok := v.(*lua.LFunction); ok {
			item.Action = fn
		}
	}
	if v := t.RawGetString("items"); v != nil {
		if itemsTbl, ok := v.(*lua.LTable); ok {
			itemsTbl.ForEach(func(_, iv lua.LValue) {
				if it, ok := iv.(*lua.LTable); ok {
					item.Items = append(item.Items, parseMenuItem(L, it))
				}
			})
		}
	}
	return item
}

// buildFyneMenu converts LuaMenuItem slice to *fyne.MainMenu.
func buildFyneMenu(L *lua.LState, items []*LuaMenuItem) *fyne.MainMenu {
	var menus []*fyne.Menu
	for _, item := range items {
		var fyneItems []*fyne.MenuItem
		for _, sub := range item.Items {
			if sub.Label == "-" {
				fyneItems = append(fyneItems, fyne.NewMenuItemSeparator())
			} else {
				fn := sub.Action // capture
				mi := fyne.NewMenuItem(sub.Label, func() {
					if fn != nil && CurrentL != nil {
						CurrentL.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true})
					}
				})
				fyneItems = append(fyneItems, mi)
			}
		}
		if len(fyneItems) > 0 {
			menus = append(menus, fyne.NewMenu(item.Label, fyneItems...))
		}
	}
	return &fyne.MainMenu{Items: menus}
}
