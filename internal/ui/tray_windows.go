package ui

import (
	"fmt"
	"go-musetool/internal/language"
	"go-musetool/internal/logger"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"unsafe"
)

var (
	onShow func()
	onExit func()
)

// Windows API definitions
var (
	// Reuse user32 from windows_utils.go
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")

	procGetModuleHandleW    = kernel32.NewProc("GetModuleHandleW")
	procRegisterClassExW    = user32.NewProc("RegisterClassExW")
	procCreateWindowExW     = user32.NewProc("CreateWindowExW")
	procDefWindowProcW      = user32.NewProc("DefWindowProcW")
	procLoadImageW          = user32.NewProc("LoadImageW")
	procShellNotifyIconW    = shell32.NewProc("Shell_NotifyIconW")
	procPostQuitMessage     = user32.NewProc("PostQuitMessage")
	procGetMessageW         = user32.NewProc("GetMessageW")
	procTranslateMessage    = user32.NewProc("TranslateMessage")
	procDispatchMessageW    = user32.NewProc("DispatchMessageW")
	procCreatePopupMenu     = user32.NewProc("CreatePopupMenu")
	procAppendMenuW         = user32.NewProc("AppendMenuW")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procTrackPopupMenu      = user32.NewProc("TrackPopupMenu")
	procDestroyWindow       = user32.NewProc("DestroyWindow")
)

const (
	IMAGE_ICON      = 0x1
	LR_LOADFROMFILE = 0x00000010
	LR_DEFAULTSIZE  = 0x00000040
	NIF_ICON        = 0x00000002
	NIF_MESSAGE     = 0x00000001
	NIF_TIP         = 0x00000004
	WM_TRAYICON     = 0x0400
	WM_LBUTTONUP    = 0x0202
	WM_RBUTTONUP    = 0x0205
	WM_COMMAND      = 0x0111
	WM_DESTROY      = 0x0002
	ID_TRAY_SHOW    = 1001
	ID_TRAY_EXIT    = 1002
	NIM_ADD         = 0x00000000
	NIM_DELETE      = 0x00000002
	MF_STRING       = 0x00000000
	MF_SEPARATOR    = 0x00000800
	TPM_BOTTOMALIGN = 0x00000020
	TPM_LEFTALIGN   = 0x00000000
)

type WNDCLASSEX struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   syscall.Handle
	Icon       syscall.Handle
	Cursor     syscall.Handle
	Background syscall.Handle
	MenuName   *uint16
	ClassName  *uint16
	SmallIcon  syscall.Handle
}

type NOTIFYICONDATA struct {
	Size            uint32
	Wnd             syscall.Handle
	ID              uint32
	Flags           uint32
	CallbackMessage uint32
	Icon            syscall.Handle
	Tip             [128]uint16
}

var (
	trayHwnd syscall.Handle
	trayIcon syscall.Handle
)

// InitTray initializes the native Windows system tray
func InitTray(iconPath string, showCallback func(), exitCallback func()) {
	initTrayWithIcon(iconPath, nil, showCallback, exitCallback)
}

// InitTrayWithData initializes the native Windows system tray with embedded icon data
func InitTrayWithData(iconData []byte, showCallback func(), exitCallback func()) {
	initTrayWithIcon("", iconData, showCallback, exitCallback)
}

// extractIconToDirectory extracts embedded icon data to the icons directory
func extractIconToDirectory(iconData []byte) (string, error) {
	// Get executable directory
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	exeDir := filepath.Dir(exePath)

	// Create icons directory in the same directory as the executable
	iconsDir := filepath.Join(exeDir, "icons")
	if err := os.MkdirAll(iconsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create icons directory: %w", err)
	}

	// Target icon path - use real_icon.ico which is a true ICO file
	iconPath := filepath.Join(iconsDir, "real_icon.ico")

	// Check if icon already exists and is valid
	if existingData, err := os.ReadFile(iconPath); err == nil {
		if len(existingData) == len(iconData) {
			logger.Info("[Tray] Icon already exists at %s (size: %d bytes), skipping extraction", iconPath, len(existingData))
			return iconPath, nil
		}
	}

	// Write icon data to file
	if err := os.WriteFile(iconPath, iconData, 0644); err != nil {
		return "", fmt.Errorf("failed to write icon file: %w", err)
	}

	logger.Info("[Tray] Extracted icon to %s (size: %d bytes)", iconPath, len(iconData))
	return iconPath, nil
}

// initTrayWithWindow initializes the native Windows system tray with icon path or embedded data
func initTrayWithIcon(iconPath string, iconData []byte, showCallback func(), exitCallback func()) {
	onShow = showCallback
	onExit = exitCallback

	go func() {
		// Lock OS thread for Windows message loop
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		className, _ := syscall.UTF16PtrFromString("GoMuseToolTrayClass")
		windowName, _ := syscall.UTF16PtrFromString("GoMuseToolTray")
		hInstance, _, _ := procGetModuleHandleW.Call(0)

		// Register Window Class
		var wc WNDCLASSEX
		wc.Size = uint32(unsafe.Sizeof(wc))
		wc.WndProc = syscall.NewCallback(wndProc)
		wc.Instance = syscall.Handle(hInstance)
		wc.ClassName = className

		procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

		// Create Window (Hidden)
		hwnd, _, _ := procCreateWindowExW.Call(
			0,
			uintptr(unsafe.Pointer(className)),
			uintptr(unsafe.Pointer(windowName)),
			0,
			0, 0, 0, 0,
			0, 0, uintptr(hInstance), 0,
		)
		trayHwnd = syscall.Handle(hwnd)

		var hIcon uintptr
		candidatePaths := make([]string, 0, 4)

		// 如果提供了嵌入的图标数据，解压到 icons 目录
		if len(iconData) > 0 {
			logger.Info("[Tray] Received embedded icon data, size: %d bytes", len(iconData))
			extractedPath, err := extractIconToDirectory(iconData)
			if err != nil {
				logger.Error("[Tray] Failed to extract icon: %v", err)
			} else {
				candidatePaths = append(candidatePaths, extractedPath)
			}
		} else {
			logger.Info("[Tray] Warning: No embedded icon data provided!")
		}

		// 添加其他候选路径
		if iconPath != "" {
			candidatePaths = append(candidatePaths, iconPath)
		}

		// Get executable directory for fallback paths
		exePath, _ := os.Executable()
		if exePath != "" {
			exeDir := filepath.Dir(exePath)
			// 首先尝试根目录的 real_icon.ico (真正的 ICO 文件)
			candidatePaths = append(candidatePaths,
				filepath.Join(exeDir, "real_icon.ico"),
				filepath.Join(exeDir, "icons", "real_icon.ico"),
			)
		}

		candidatePaths = append(candidatePaths,
			"real_icon.ico",
			filepath.Join("icons", "real_icon.ico"),
			filepath.Join("internal", "assets", "real_icon.ico"), // For dev environment
		)

		// 尝试加载图标
		seen := make(map[string]struct{})
		for _, candidate := range candidatePaths {
			if candidate == "" {
				continue
			}
			if _, ok := seen[candidate]; ok {
				continue
			}
			seen[candidate] = struct{}{}

			absPath, err := filepath.Abs(candidate)
			if err != nil {
				absPath = candidate
			}

			logger.Debug("[Tray] Trying to load icon from: %s", absPath)

			// Check if file exists
			if _, err := os.Stat(absPath); err != nil {
				logger.Debug("[Tray] File does not exist or not accessible: %v", err)
				continue
			}

			iconPathPtr, err := syscall.UTF16PtrFromString(absPath)
			if err != nil {
				logger.Error("[Tray] Failed to convert path to UTF16: %v", err)
				continue
			}

			// Try different loading strategies
			// Strategy 1: Load with default size
			var lastErr error
			hIcon, _, lastErr = procLoadImageW.Call(
				0,
				uintptr(unsafe.Pointer(iconPathPtr)),
				IMAGE_ICON,
				0, 0,
				LR_LOADFROMFILE|LR_DEFAULTSIZE,
			)

			// Strategy 2: If failed, try with specific size (16x16 for tray)
			if hIcon == 0 {
				logger.Debug("[Tray] First attempt failed, trying with 16x16 size...")
				hIcon, _, lastErr = procLoadImageW.Call(
					0,
					uintptr(unsafe.Pointer(iconPathPtr)),
					IMAGE_ICON,
					16, 16,
					LR_LOADFROMFILE,
				)
			}

			// Strategy 3: If still failed, try with 32x32 size
			if hIcon == 0 {
				logger.Debug("[Tray] Second attempt failed, trying with 32x32 size...")
				hIcon, _, lastErr = procLoadImageW.Call(
					0,
					uintptr(unsafe.Pointer(iconPathPtr)),
					IMAGE_ICON,
					32, 32,
					LR_LOADFROMFILE,
				)
			}

			if hIcon != 0 {
				logger.Info("[Tray] Successfully loaded icon from: %s", absPath)
				break
			} else {
				logger.Debug("[Tray] LoadImageW failed for %s, error: %v", absPath, lastErr)
			}
		}

		if hIcon == 0 {
			logger.Error("[Tray] ERROR: Failed to load any icon! Tray icon will not be visible.")
		}

		trayIcon = syscall.Handle(hIcon)

		// Add Tray Icon
		var nid NOTIFYICONDATA
		nid.Size = uint32(unsafe.Sizeof(nid))
		nid.Wnd = trayHwnd
		nid.ID = 1
		nid.Flags = NIF_ICON | NIF_MESSAGE | NIF_TIP
		nid.CallbackMessage = WM_TRAYICON
		nid.Icon = trayIcon

		tip := "Go MuseTool"
		tipUTF16, _ := syscall.UTF16FromString(tip)
		copy(nid.Tip[:], tipUTF16)

		procShellNotifyIconW.Call(NIM_ADD, uintptr(unsafe.Pointer(&nid)))

		// Message Loop
		var msg struct {
			Hwnd    syscall.Handle
			Message uint32
			WParam  uintptr
			LParam  uintptr
			Time    uint32
			Pt      POINT
		}

		for {
			ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
			if ret == 0 {
				break
			}
			procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
			procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
		}
	}()
}

func wndProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_TRAYICON:
		switch lParam {
		case WM_LBUTTONUP:
			// Left click: Show window
			if onShow != nil {
				onShow()
			}
		case WM_RBUTTONUP:
			// Right click: Show menu
			showTrayMenu()
		}
	case WM_COMMAND:
		id := int(wParam & 0xffff)
		switch id {
		case ID_TRAY_SHOW:
			if onShow != nil {
				onShow()
			}
		case ID_TRAY_EXIT:
			RemoveTray()
			if onExit != nil {
				onExit()
			}
		}
	case WM_DESTROY:
		procPostQuitMessage.Call(0)
	default:
		ret, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
		return ret
	}
	return 0
}

func showTrayMenu() {
	hMenu, _, _ := procCreatePopupMenu.Call()

	// Add menu items
	showText, _ := syscall.UTF16PtrFromString(language.T().TrayShow)
	exitText, _ := syscall.UTF16PtrFromString(language.T().TrayExit)

	procAppendMenuW.Call(hMenu, MF_STRING, ID_TRAY_SHOW, uintptr(unsafe.Pointer(showText)))
	procAppendMenuW.Call(hMenu, MF_SEPARATOR, 0, 0)
	procAppendMenuW.Call(hMenu, MF_STRING, ID_TRAY_EXIT, uintptr(unsafe.Pointer(exitText)))

	var pt POINT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

	// Required for the menu to disappear when clicking outside
	procSetForegroundWindow.Call(uintptr(trayHwnd))

	procTrackPopupMenu.Call(
		hMenu,
		TPM_BOTTOMALIGN|TPM_LEFTALIGN,
		uintptr(pt.X),
		uintptr(pt.Y),
		0,
		uintptr(trayHwnd),
		0,
	)

	procDestroyWindow.Call(hMenu)
}

// RemoveTray removes the tray icon and cleans up
func RemoveTray() {
	var nid NOTIFYICONDATA
	nid.Size = uint32(unsafe.Sizeof(nid))
	nid.Wnd = trayHwnd
	nid.ID = 1

	procShellNotifyIconW.Call(NIM_DELETE, uintptr(unsafe.Pointer(&nid)))
	procDestroyWindow.Call(uintptr(trayHwnd))
}
