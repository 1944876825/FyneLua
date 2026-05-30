package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/yuin/gopher-lua"

	"golua-fyne/bridge"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/fsnotify/fsnotify"
)

// 全局保持当前 LState 引用，回调需要它
var currentL *lua.LState

func main() {
	a := app.New()

	// Enable hot-reload mode: ShowAndRun() becomes non-blocking Show()
	bridge.SetHotReload(true)

	script := "scripts/main.lua"
	if len(os.Args) > 1 {
		script = os.Args[1]
	}

	// Run the Lua script for the first time
	if err := runScript(a, script); err != nil {
		errMsg := err.Error()
		fmt.Fprintf(os.Stderr, "Lua error: %v\n", errMsg)
		// Show error dialog if possible
		go func() {
			time.Sleep(500 * time.Millisecond) // wait for window to appear
			bridge.ShowLuaError("Lua 脚本错误", errMsg)
		}()
		// Still continue with a.Run() so the error dialog is visible
		watchDir := filepath.Dir(script)
		go startWatcher(watchDir, script)
		a.Run()
		return
	}

	// Start file watcher for hot-reload in background
	watchDir := filepath.Dir(script)
	go startWatcher(watchDir, script)

	// Block here running the Fyne event loop
	a.Run()
}

// runScript creates a fresh Lua engine, executes the script.
// The LState stays alive for callbacks to work.
// On subsequent calls (hot-reload), the old LState is closed first.
func runScript(a fyne.App, script string) error {
	// Close extra windows created by previous Lua run
	bridge.ClearExtraWindows()

	// Close old LState from previous run
	if currentL != nil {
		currentL.Close()
	}

	L := bridge.NewEngine(a)
	currentL = L

	return L.DoFile(script)
}

// startWatcher monitors a directory for .lua file changes and reloads the script.
func startWatcher(dir string, script string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("热重载监听失败: %v", err)
		return
	}
	defer watcher.Close()

	if err := watcher.Add(dir); err != nil {
		log.Printf("热重载监听目录失败: %v", err)
		return
	}

	log.Printf("🔥 热重载已启动，监听目录: %s", dir)

	// Debounce: file save may trigger multiple events
	var debounce *time.Timer
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				filename := filepath.Base(event.Name)
				if filepath.Ext(filename) == ".lua" {
					if debounce != nil {
						debounce.Stop()
					}
					debounce = time.AfterFunc(300*time.Millisecond, func() {
						log.Printf("🔄 检测到文件变更: %s，正在重载...", filename)
						fyne.Do(func() {
							if err := runScript(fyne.CurrentApp(), script); err != nil {
								log.Printf("❌ 重载失败: %v", err)
								bridge.ShowLuaError("重载失败", err.Error())
							} else {
								log.Printf("✅ 重载成功: %s", filename)
							}
						})
					})
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("热重载错误: %v", err)
		}
	}
}
