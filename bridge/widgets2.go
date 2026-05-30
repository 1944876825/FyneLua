package bridge

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/yuin/gopher-lua"
)

// ---------- Scroll constructor ----------

func lScrollFn(L *lua.LState) int {
	child := checkWidget(L, 1)
	obj := container.NewScroll(child.Obj)
	pushWidget(L, &LuaWidget{Obj: obj, Type: "Scroll"})
	return 1
}

// ---------- Icon constructor ----------
// gui.Icon(name) -- name is a theme icon name like "search", "confirm", "cancel", etc.

func lIconFn(L *lua.LState) int {
	name := L.CheckString(1)
	icon := resolveIcon(name)
	pushWidget(L, &LuaWidget{Obj: widget.NewIcon(icon), Type: "Icon"})
	return 1
}

// ---------- Image constructor ----------
// gui.Image(path) -- loads image from file path

func lImageFn(L *lua.LState) int {
	path := L.CheckString(1)
	img := canvas.NewImageFromFile(path)
	pushWidget(L, &LuaWidget{Obj: img, Type: "Image"})
	return 1
}

// ---------- Rectangle constructor ----------
// gui.Rectangle(color_hex) -- e.g. "#FF0000"

func lRectangleFn(L *lua.LState) int {
	hex := L.CheckString(1)
	c := parseHexColor(hex)
	rect := canvas.NewRectangle(c)
	pushWidget(L, &LuaWidget{Obj: rect, Type: "Rectangle"})
	return 1
}

// ---------- MultiLineEntry constructor ----------

func lMultiLineEntryFn(L *lua.LState) int {
	entry := widget.NewEntry()
	entry.MultiLine = true
	pushWidget(L, &LuaWidget{Obj: entry, Type: "Entry"})
	return 1
}

// ---------- PasswordEntry constructor ----------

func lPasswordEntryFn(L *lua.LState) int {
	entry := widget.NewPasswordEntry()
	pushWidget(L, &LuaWidget{Obj: entry, Type: "Entry"})
	return 1
}

// ---------- Scroll methods ----------

func scrollMethod(L *lua.LState, w *LuaWidget, method string) {
	switch method {
	case "ScrollToTop":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			if s, ok := w.Obj.(*container.Scroll); ok {
				s.Offset = fyne.NewPos(0, 0)
				s.Refresh()
			}
			return 0
		}))
	case "ScrollToBottom":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			if s, ok := w.Obj.(*container.Scroll); ok {
				s.Offset = fyne.NewPos(0, 99999)
				s.Refresh()
			}
			return 0
		}))
	case "ScrollToOffset":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			offset := float32(L.CheckNumber(2))
			if s, ok := w.Obj.(*container.Scroll); ok {
				s.Offset = fyne.NewPos(0, offset)
				s.Refresh()
			}
			return 0
		}))
	case "Direction":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			// Direction is not directly settable on Fyne Scroll widget
			// Use NewHScroll or NewVScroll for single-direction scrolling
			return 0
		}))
	case "Content":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			if s, ok := w.Obj.(*container.Scroll); ok {
				cw := checkWidget(L, 2)
				s.Content = cw.Obj
				s.Refresh()
			}
			return 0
		}))
	default:
		L.Push(lua.LNil)
	}
}

// ---------- Image methods ----------

func imageMethod(L *lua.LState, w *LuaWidget, method string) {
	switch method {
	case "FillMode":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			mode := L.CheckString(2)
			img := w.Obj.(*canvas.Image)
			switch strings.ToLower(mode) {
			case "original", "fit":
				img.FillMode = canvas.ImageFillOriginal
			case "contain":
				img.FillMode = canvas.ImageFillContain
			case "stretch", "fill":
				img.FillMode = canvas.ImageFillStretch
			}
			img.Refresh()
			return 0
		}))
	case "SetMinSize":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			wi := float32(L.CheckNumber(2))
			hi := float32(L.CheckNumber(3))
			w.Obj.(*canvas.Image).SetMinSize(fyne.NewSize(wi, hi))
			return 0
		}))
	default:
		L.Push(lua.LNil)
	}
}

// ---------- Rectangle methods ----------

func rectangleMethod(L *lua.LState, w *LuaWidget, method string) {
	switch method {
	case "SetColor":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			hex := L.CheckString(2)
			w.Obj.(*canvas.Rectangle).FillColor = parseHexColor(hex)
			w.Obj.Refresh()
			return 0
		}))
	case "SetMinSize":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			wi := float32(L.CheckNumber(2))
			hi := float32(L.CheckNumber(3))
			w.Obj.(*canvas.Rectangle).SetMinSize(fyne.NewSize(wi, hi))
			return 0
		}))
	default:
		L.Push(lua.LNil)
	}
}

// ---------- Helpers ----------

// resolveIcon maps string names to Fyne theme icons.
func resolveIcon(name string) fyne.Resource {
	m := map[string]fyne.Resource{
		"cancel":           theme.CancelIcon(),
		"confirm":          theme.ConfirmIcon(),
		"delete":           theme.DeleteIcon(),
		"search":           theme.SearchIcon(),
		"settings":         theme.SettingsIcon(),
		"home":             theme.HomeIcon(),
		"info":             theme.InfoIcon(),
		"warning":          theme.WarningIcon(),
		"error":            theme.ErrorIcon(),
		"question":         theme.QuestionIcon(),
		"help":             theme.HelpIcon(),
		"menu":             theme.MenuIcon(),
		"navigateback":     theme.NavigateBackIcon(),
		"navigateforward":  theme.NavigateNextIcon(),
		"download":         theme.DownloadIcon(),
		"computer":         theme.ComputerIcon(),
		"document":         theme.DocumentIcon(),
		"documentcreate":   theme.DocumentCreateIcon(),
		"documentsave":     theme.DocumentSaveIcon(),
		"folder":           theme.FolderIcon(),
		"folderopen":       theme.FolderOpenIcon(),
		"foldernew":        theme.FolderNewIcon(),
		"file":             theme.FileIcon(),
		"filetext":         theme.FileTextIcon(),
		"fileaudio":        theme.FileAudioIcon(),
		"fileimage":        theme.FileImageIcon(),
		"filevideo":        theme.FileVideoIcon(),
		"music":            theme.MediaMusicIcon(),
		"photo":            theme.MediaPhotoIcon(),
		"video":            theme.MediaVideoIcon(),
		"play":             theme.MediaPlayIcon(),
		"pause":            theme.MediaPauseIcon(),
		"stop":             theme.MediaStopIcon(),
		"record":           theme.MediaRecordIcon(),
		"fastforward":      theme.MediaFastForwardIcon(),
		"fastrewind":       theme.MediaFastRewindIcon(),
		"skipnext":         theme.MediaSkipNextIcon(),
		"skipprevious":     theme.MediaSkipPreviousIcon(),
		"mail":             theme.MailComposeIcon(),
		"mailsend":         theme.MailSendIcon(),
		"visibility":       theme.VisibilityIcon(),
		"visibilityoff":    theme.VisibilityOffIcon(),
		"check":            theme.CheckButtonIcon(),
		"radio":            theme.RadioButtonIcon(),
		"arrowup":          theme.MoveUpIcon(),
		"arrowdown":        theme.MoveDownIcon(),
		"contentadd":       theme.ContentAddIcon(),
		"contentremove":    theme.ContentRemoveIcon(),
		"contentcopy":      theme.ContentCopyIcon(),
		"contentpaste":     theme.ContentPasteIcon(),
		"contentcut":       theme.ContentCutIcon(),
		"contentclear":     theme.ContentClearIcon(),
		"contentundo":      theme.ContentUndoIcon(),
		"contentredo":      theme.ContentRedoIcon(),
		"fullscreen":       theme.ViewFullScreenIcon(),
		"viewrefresh":      theme.ViewRefreshIcon(),
		"morehorizontal":   theme.MoreHorizontalIcon(),
		"morevertical":     theme.MoreVerticalIcon(),
		"zoomin":           theme.ZoomInIcon(),
		"zoomout":          theme.ZoomOutIcon(),
		"zoomfit":          theme.ZoomFitIcon(),
		"history":          theme.HistoryIcon(),
		"logincolor":       theme.LoginIcon(),
		"logoutcolor":      theme.LogoutIcon(),
		"account":          theme.AccountIcon(),
		"brokenimage":      theme.BrokenImageIcon(),
	}
	if r, ok := m[strings.ToLower(name)]; ok {
		return r
	}
	return theme.InfoIcon()
}

// parseHexColor parses "#RRGGBB" or "#RGB" into a color.Color.
func parseHexColor(hex string) color.Color {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		return color.Black
	}
	var r, g, b uint8
	for i := 0; i < 6; i += 2 {
		val := 0
		for j := 0; j < 2; j++ {
			c := hex[i+j]
			val *= 16
			if c >= '0' && c <= '9' {
				val += int(c - '0')
			} else if c >= 'a' && c <= 'f' {
				val += int(c-'a') + 10
			} else if c >= 'A' && c <= 'F' {
				val += int(c-'A') + 10
			}
		}
		if i == 0 {
			r = uint8(val)
		} else if i == 2 {
			g = uint8(val)
		} else {
			b = uint8(val)
		}
	}
	return color.NRGBA{R: r, G: g, B: b, A: 255}
}
