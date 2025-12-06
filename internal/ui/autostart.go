package ui

import (
	"log"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

const (
	runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
	appName    = "GoMuseTool"
)

// SetAutoStart enables or disables auto-start on Windows boot
func SetAutoStart(enable bool) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	if enable {
		// Get the executable path
		exePath, err := os.Executable()
		if err != nil {
			return err
		}
		// Convert to absolute path
		absPath, err := filepath.Abs(exePath)
		if err != nil {
			return err
		}
		// Set the registry value
		err = k.SetStringValue(appName, absPath)
		if err != nil {
			return err
		}
		log.Printf("Auto-start enabled: %s", absPath)
	} else {
		// Remove the registry value
		err = k.DeleteValue(appName)
		if err != nil && err != registry.ErrNotExist {
			return err
		}
		log.Printf("Auto-start disabled")
	}

	return nil
}

// IsAutoStartEnabled checks if auto-start is currently enabled
func IsAutoStartEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()

	_, _, err = k.GetStringValue(appName)
	return err == nil
}

// MinimizeWindow minimizes a window to the taskbar
func MinimizeWindow(hwnd uintptr) {
	const SW_MINIMIZE = 6
	showWindow := syscall.NewLazyDLL("user32.dll").NewProc("ShowWindow")
	showWindow.Call(hwnd, SW_MINIMIZE)
}

// HideWindow hides a window completely
func HideWindow(hwnd uintptr) {
	const SW_HIDE = 0
	showWindow := syscall.NewLazyDLL("user32.dll").NewProc("ShowWindow")
	showWindow.Call(hwnd, SW_HIDE)
}

// ShowWindow shows a window
func ShowWindowNormal(hwnd uintptr) {
	const SW_SHOW = 5
	const SW_RESTORE = 9
	showWindow := syscall.NewLazyDLL("user32.dll").NewProc("ShowWindow")
	// First restore if minimized
	showWindow.Call(hwnd, SW_RESTORE)
	// Then show
	showWindow.Call(hwnd, SW_SHOW)
	// Bring to foreground
	setForegroundWindow := syscall.NewLazyDLL("user32.dll").NewProc("SetForegroundWindow")
	setForegroundWindow.Call(hwnd)
}

// IsWindowVisible checks if a window is currently visible
func IsWindowVisible(hwnd uintptr) bool {
	isWindowVisible := syscall.NewLazyDLL("user32.dll").NewProc("IsWindowVisible")
	ret, _, _ := isWindowVisible.Call(hwnd)
	return ret != 0
}

// BringWindowToFront brings a window to the foreground
func BringWindowToFront(hwnd uintptr) {
	// Remove topmost flag temporarily
	SetWindowAlwaysOnTop(hwnd, false)

	// Show and restore the window
	const SW_RESTORE = 9
	showWindow := syscall.NewLazyDLL("user32.dll").NewProc("ShowWindow")
	showWindow.Call(hwnd, SW_RESTORE)

	// Set foreground
	setForegroundWindow := syscall.NewLazyDLL("user32.dll").NewProc("SetForegroundWindow")
	setForegroundWindow.Call(hwnd)

	// Restore topmost flag
	SetWindowAlwaysOnTop(hwnd, true)
}

// FlashWindow flashes the window to get user attention
func FlashWindow(hwnd uintptr) {
	const FLASHW_ALL = 3
	const FLASHW_TIMERNOFG = 12

	type FLASHWINFO struct {
		cbSize    uint32
		hwnd      uintptr
		dwFlags   uint32
		uCount    uint32
		dwTimeout uint32
	}

	flashWindowEx := syscall.NewLazyDLL("user32.dll").NewProc("FlashWindowEx")

	fwi := FLASHWINFO{
		cbSize:    uint32(unsafe.Sizeof(FLASHWINFO{})),
		hwnd:      hwnd,
		dwFlags:   FLASHW_ALL | FLASHW_TIMERNOFG,
		uCount:    3,
		dwTimeout: 0,
	}

	flashWindowEx.Call(uintptr(unsafe.Pointer(&fwi)))
}
