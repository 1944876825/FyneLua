package bridge

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"

	"github.com/yuin/gopher-lua"
)

// ==================== IO Module ====================

// IOModule returns a Lua table with file I/O functions.
func IOModule(L *lua.LState) *lua.LTable {
	mod := L.NewTable()

	// io.readFile(path) -> string | nil, err
	L.SetField(mod, "readFile", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		data, err := os.ReadFile(path)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LString(string(data)))
		return 1
	}))

	// io.writeFile(path, content) -> true | false, err
	L.SetField(mod, "writeFile", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		content := L.CheckString(2)
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	// io.appendFile(path, content) -> true | false, err
	L.SetField(mod, "appendFile", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		content := L.CheckString(2)
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		defer f.Close()
		if _, err := f.WriteString(content); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	// io.listDir(path) -> table | nil, err
	L.SetField(mod, "listDir", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		entries, err := os.ReadDir(path)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		tbl := L.NewTable()
		for _, entry := range entries {
			tbl.Append(lua.LString(entry.Name()))
		}
		L.Push(tbl)
		return 1
	}))

	// io.exists(path) -> bool
	L.SetField(mod, "exists", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		_, err := os.Stat(path)
		L.Push(lua.LBool(err == nil))
		return 1
	}))

	// io.remove(path) -> true | false, err
	L.SetField(mod, "remove", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		err := os.Remove(path)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	// io.mkdir(path) -> true | false, err
	L.SetField(mod, "mkdir", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		err := os.MkdirAll(path, 0755)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	return mod
}

// ==================== Net Module (Async) ====================

// NetModule returns a Lua table with async HTTP functions.
// net.get(url [, headers], callback)
// net.post(url, body [, headers], callback)
// callback: function(resp) where resp = {status=number, body=string} or function(nil, err_string)
func NetModule(L *lua.LState) *lua.LTable {
	mod := L.NewTable()

	parseHeaders := func(L *lua.LState, idx int) http.Header {
		h := make(http.Header)
		if L.Get(idx) == lua.LNil {
			return h
		}
		t := L.CheckTable(idx)
		t.ForEach(func(k, v lua.LValue) {
			h.Set(k.String(), v.String())
		})
		return h
	}

	// net.get(url [, headers], callback)
	// callback: function(resp, err)  -- resp={status,body} or nil
	L.SetField(mod, "get", L.NewFunction(func(L *lua.LState) int {
		rawURL := L.CheckString(1)

		// 判断参数：第二个参数可能是 headers 或 callback
		var headers http.Header
		var callback *lua.LFunction
		n := L.GetTop()
		if n >= 3 {
			headers = parseHeaders(L, 2)
			callback = L.CheckFunction(3)
		} else {
			callback = L.CheckFunction(2)
		}

		req, _ := http.NewRequest("GET", rawURL, nil)
		if req != nil {
			req.Header = headers
		}

		go func() {
			client := &http.Client{Timeout: 15 * time.Second}
			resp, err := client.Do(req)
			fyne.Do(func() {
				if CurrentL == nil {
					return
				}
				if err != nil {
					CurrentL.CallByParam(lua.P{Fn: callback, NRet: 0, Protect: true}, lua.LNil, lua.LString(err.Error()))
					return
				}
				defer resp.Body.Close()
				body, _ := io.ReadAll(resp.Body)
				tbl := CurrentL.NewTable()
				tbl.RawSetString("status", lua.LNumber(resp.StatusCode))
				tbl.RawSetString("body", lua.LString(string(body)))
				CurrentL.CallByParam(lua.P{Fn: callback, NRet: 0, Protect: true}, tbl, lua.LNil)
			})
		}()

		return 0
	}))

	// net.post(url, body [, headers], callback)
	L.SetField(mod, "post", L.NewFunction(func(L *lua.LState) int {
		rawURL := L.CheckString(1)
		body := L.CheckString(2)

		var headers http.Header
		var callback *lua.LFunction
		n := L.GetTop()
		if n >= 4 {
			headers = parseHeaders(L, 3)
			callback = L.CheckFunction(4)
		} else {
			callback = L.CheckFunction(3)
		}

		req, _ := http.NewRequest("POST", rawURL, bytes.NewBufferString(body))
		if req != nil {
			req.Header = headers
			if req.Header.Get("Content-Type") == "" {
				req.Header.Set("Content-Type", "application/json")
			}
		}

		go func() {
			client := &http.Client{Timeout: 15 * time.Second}
			resp, err := client.Do(req)
			fyne.Do(func() {
				if CurrentL == nil {
					return
				}
				if err != nil {
					CurrentL.CallByParam(lua.P{Fn: callback, NRet: 0, Protect: true}, lua.LNil, lua.LString(err.Error()))
					return
				}
				defer resp.Body.Close()
				bodyData, _ := io.ReadAll(resp.Body)
				tbl := CurrentL.NewTable()
				tbl.RawSetString("status", lua.LNumber(resp.StatusCode))
				tbl.RawSetString("body", lua.LString(string(bodyData)))
				CurrentL.CallByParam(lua.P{Fn: callback, NRet: 0, Protect: true}, tbl, lua.LNil)
			})
		}()

		return 0
	}))

	return mod
}

// ==================== Timer Module ====================

var timerIDCounter uint64

// activeTimers stores running tickers by ID for cancellation.
var activeTimers sync.Map

// TimerModule returns a Lua table with timer functions.
// Callbacks are dispatched via fyne.Do to run on the main goroutine.
func TimerModule() *lua.LTable {
	mod := new(lua.LTable)

	// timer.after(ms, callback)
	mod.RawSetString("after", &lua.LFunction{IsG: true, GFunction: func(L *lua.LState) int {
		ms := int(L.CheckNumber(1))
		fn := L.CheckFunction(2)

		time.AfterFunc(time.Duration(ms)*time.Millisecond, func() {
			fyne.Do(func() {
				if CurrentL != nil {
					CurrentL.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true})
				}
			})
		})

		return 0
	}})

	// timer.every(ms, callback) -> id
	mod.RawSetString("every", &lua.LFunction{IsG: true, GFunction: func(L *lua.LState) int {
		ms := int(L.CheckNumber(1))
		fn := L.CheckFunction(2)

		id := atomic.AddUint64(&timerIDCounter, 1)
		ticker := time.NewTicker(time.Duration(ms) * time.Millisecond)
		go func() {
			for range ticker.C {
				fyne.Do(func() {
					if CurrentL != nil {
						CurrentL.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true})
					}
				})
			}
		}()

		activeTimers.Store(id, ticker)

		L.Push(lua.LNumber(id))
		return 1
	}})

	// timer.cancel(id)
	mod.RawSetString("cancel", &lua.LFunction{IsG: true, GFunction: func(L *lua.LState) int {
		id := uint64(L.CheckNumber(1))
		if v, ok := activeTimers.Load(id); ok {
			v.(*time.Ticker).Stop()
			activeTimers.Delete(id)
		}
		return 0
	}})

	return mod
}
