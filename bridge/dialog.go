package bridge

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"github.com/yuin/gopher-lua"
)

// DialogModule returns a Lua table with dialog functions.
func DialogModule() *lua.LTable {
	mod := new(lua.LTable)

	parent := func() fyne.Window {
		windows := fyne.CurrentApp().Driver().AllWindows()
		if len(windows) > 0 {
			return windows[0]
		}
		return nil
	}

	// dialog.showInfo(title, message)
	mod.RawSetString("showInfo", &lua.LFunction{IsG: true, GFunction: func(L *lua.LState) int {
		dialog.ShowInformation(L.CheckString(1), L.CheckString(2), parent())
		return 0
	}})

	// dialog.showError(title, message)
	mod.RawSetString("showError", &lua.LFunction{IsG: true, GFunction: func(L *lua.LState) int {
		dialog.ShowError(fmt.Errorf("%s", L.CheckString(2)), parent())
		return 0
	}})

	// dialog.showConfirm(title, message, callback)
	// callback(bool) -- true = OK, false = Cancel
	mod.RawSetString("showConfirm", &lua.LFunction{IsG: true, GFunction: func(L *lua.LState) int {
		title := L.CheckString(1)
		message := L.CheckString(2)
		callback := L.CheckFunction(3)
		dialog.ShowConfirm(title, message, func(confirmed bool) {
			if CurrentL != nil {
				CurrentL.CallByParam(lua.P{Fn: callback, NRet: 0, Protect: true}, lua.LBool(confirmed))
			}
		}, parent())
		return 0
	}})

	// dialog.showEntry(title, message, onConfirm)
	// onConfirm(text) or onConfirm(nil) if cancelled
	mod.RawSetString("showEntry", &lua.LFunction{IsG: true, GFunction: func(L *lua.LState) int {
		title := L.CheckString(1)
		message := L.CheckString(2)
		callback := L.CheckFunction(3)
		d := dialog.NewEntryDialog(title, message, func(text string) {
			if CurrentL != nil {
				CurrentL.CallByParam(lua.P{Fn: callback, NRet: 0, Protect: true}, lua.LString(text))
			}
		}, parent())
		d.Show()
		return 0
	}})

	// dialog.showProgress(title, message) -> {setValue(v), hide()}
	mod.RawSetString("showProgress", &lua.LFunction{IsG: true, GFunction: func(L *lua.LState) int {
		pd := dialog.NewProgress(L.CheckString(1), L.CheckString(2), parent())
		pd.Show()
		ctl := L.NewTable()
		L.SetField(ctl, "setValue", L.NewFunction(func(L *lua.LState) int {
			pd.SetValue(float64(L.CheckNumber(2)))
			return 0
		}))
		L.SetField(ctl, "hide", L.NewFunction(func(L *lua.LState) int {
			pd.Hide()
			return 0
		}))
		L.Push(ctl)
		return 1
	}})

	// dialog.showFileOpen(callback)
	// callback(path) or callback(nil) if cancelled
	mod.RawSetString("showFileOpen", &lua.LFunction{IsG: true, GFunction: func(L *lua.LState) int {
		callback := L.CheckFunction(1)
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if CurrentL == nil {
				return
			}
			if err != nil || reader == nil {
				CurrentL.CallByParam(lua.P{Fn: callback, NRet: 0, Protect: true}, lua.LNil)
				return
			}
			path := reader.URI().Path()
			reader.Close()
			CurrentL.CallByParam(lua.P{Fn: callback, NRet: 0, Protect: true}, lua.LString(path))
		}, parent())
		return 0
	}})

	// dialog.showFileOpenFiltered(desc, extensions, callback)
	// extensions: {".txt", ".lua"} -- with leading dot
	mod.RawSetString("showFileOpenFiltered", &lua.LFunction{IsG: true, GFunction: func(L *lua.LState) int {
		desc := L.CheckString(1)
		exts := lCheckStringTable(L, 2)
		callback := L.CheckFunction(3)
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if CurrentL == nil {
				return
			}
			if err != nil || reader == nil {
				CurrentL.CallByParam(lua.P{Fn: callback, NRet: 0, Protect: true}, lua.LNil)
				return
			}
			path := reader.URI().Path()
			reader.Close()
			CurrentL.CallByParam(lua.P{Fn: callback, NRet: 0, Protect: true}, lua.LString(path))
		}, parent())
		fd.SetFilter(storage.NewExtensionFileFilter(exts))
		fd.SetFileName(desc)
		fd.Show()
		return 0
	}})

	// dialog.showFileSave(callback)
	mod.RawSetString("showFileSave", &lua.LFunction{IsG: true, GFunction: func(L *lua.LState) int {
		callback := L.CheckFunction(1)
		dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
			if CurrentL == nil {
				return
			}
			if err != nil || writer == nil {
				CurrentL.CallByParam(lua.P{Fn: callback, NRet: 0, Protect: true}, lua.LNil)
				return
			}
			path := writer.URI().Path()
			writer.Close()
			CurrentL.CallByParam(lua.P{Fn: callback, NRet: 0, Protect: true}, lua.LString(path))
		}, parent())
		return 0
	}})

	return mod
}

// ShowLuaError displays a Lua error in a dialog window.
func ShowLuaError(title, errMsg string) {
	fyne.Do(func() {
		windows := fyne.CurrentApp().Driver().AllWindows()
		if len(windows) > 0 {
			dialog.ShowError(fmt.Errorf("%s", errMsg), windows[0])
		}
	})
}
