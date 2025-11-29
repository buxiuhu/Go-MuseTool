package ui

import (
	"syscall"
	"unsafe"
)

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	procFindWindowW   = user32.NewProc("FindWindowW")
	procSetWindowPos  = user32.NewProc("SetWindowPos")
	procGetWindowRect = user32.NewProc("GetWindowRect")
	procGetWindowLong = user32.NewProc("GetWindowLongW")
	procSetWindowLong = user32.NewProc("SetWindowLongW")

	// DWM API for title bar color
	dwmapi                    = syscall.NewLazyDLL("dwmapi.dll")
	procDwmSetWindowAttribute = dwmapi.NewProc("DwmSetWindowAttribute")
)

type RECT struct {
	Left, Top, Right, Bottom int32
}

type POINT struct {
	X, Y int32
}

var (
	procGetCursorPos = user32.NewProc("GetCursorPos")
)

const (
	GWL_STYLE           = -16
	GWL_EXSTYLE         = -20
	WS_CAPTION          = 0x00C00000
	WS_THICKFRAME       = 0x00040000
	WS_MAXIMIZE         = 0x01000000
	WS_POPUP            = 0x80000000
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	SWP_NOSIZE          = 0x0001
	SWP_NOMOVE          = 0x0002
	SWP_NOZORDER        = 0x0004
	SWP_FRAMECHANGED    = 0x0020
	WS_SYSMENU          = 0x00080000
	HWND_TOPMOST        = -1
	HWND_NOTOPMOST      = -2
	SWP_NOACTIVATE      = 0x0010
)

// DWM constants
const (
	// DWMWA_CAPTION_COLOR 值为 35，用于设置标题栏背景色 (Windows 11/新版 Win10)
	DWMWA_CAPTION_COLOR = 35
)

func GetWindowHandle(title string) uintptr {
	ptr, _ := syscall.UTF16PtrFromString(title)
	hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(ptr)))
	return hwnd
}

func GetWindowRect(hwnd uintptr) (int, int, int, int) {
	var rect RECT
	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rect)))
	return int(rect.Left), int(rect.Top), int(rect.Right - rect.Left), int(rect.Bottom - rect.Top)
}

func SetWindowPos(hwnd uintptr, x, y int) {
	procSetWindowPos.Call(hwnd, 0, uintptr(x), uintptr(y), 0, 0, SWP_NOSIZE|SWP_NOZORDER)
}

// SetWindowAlwaysOnTop 设置窗口是否总在最前
func SetWindowAlwaysOnTop(hwnd uintptr, alwaysOnTop bool) {
	var target uintptr
	if alwaysOnTop {
		target = ^uintptr(0) // -1
	} else {
		target = ^uintptr(1) // -2
	}

	// 添加 SWP_NOACTIVATE 防止抢占焦点
	procSetWindowPos.Call(hwnd, target, 0, 0, 0, 0, SWP_NOMOVE|SWP_NOSIZE|SWP_NOACTIVATE)
}

func ResizeWindow(hwnd uintptr, w, h int) {
	procSetWindowPos.Call(hwnd, 0, 0, 0, uintptr(w), uintptr(h), SWP_NOMOVE|SWP_NOZORDER)
}

func MoveAndResizeWindow(hwnd uintptr, x, y, w, h int) {
	procSetWindowPos.Call(hwnd, 0, uintptr(x), uintptr(y), uintptr(w), uintptr(h), SWP_NOZORDER)
}

func GetWindowLong(hwnd uintptr, nIndex int) uintptr {
	ret, _, _ := procGetWindowLong.Call(hwnd, uintptr(nIndex))
	return ret
}

func SetWindowLong(hwnd uintptr, nIndex int, dwNewLong uintptr) uintptr {
	ret, _, _ := procSetWindowLong.Call(hwnd, uintptr(nIndex), dwNewLong)
	return ret
}

// SetTitleBarColor 使用 DWM API 设置标题栏颜色
// color 必须是 BGR (0xBBGGRR) 格式
func SetTitleBarColor(hwnd uintptr, color uint32) {
	attr := uintptr(DWMWA_CAPTION_COLOR)
	colorPtr := unsafe.Pointer(&color)

	procDwmSetWindowAttribute.Call(
		hwnd,
		attr,
		uintptr(colorPtr),
		unsafe.Sizeof(color),
	)
}

// DWM System Backdrop Constants
const (
	DWMWA_USE_IMMERSIVE_DARK_MODE = 20
	DWMWA_SYSTEMBACKDROP_TYPE     = 38
	DWMWA_COLOR_NONE              = 0xFFFFFFFE
	DWMWA_COLOR_DEFAULT           = 0xFFFFFFFF

	DWMSBT_AUTO            = 0
	DWMSBT_NONE            = 1
	DWMSBT_MAINWINDOW      = 2 // Mica
	DWMSBT_TRANSIENTWINDOW = 3 // Acrylic
	DWMSBT_TABBEDWINDOW    = 4 // Mica Alt
)

// SetSystemBackdrop 设置系统背景效果 (Mica/Acrylic)
func SetSystemBackdrop(hwnd uintptr, backdropType int32) {
	attr := uintptr(DWMWA_SYSTEMBACKDROP_TYPE)
	procDwmSetWindowAttribute.Call(
		hwnd,
		attr,
		uintptr(unsafe.Pointer(&backdropType)),
		unsafe.Sizeof(backdropType),
	)
}

// Resize directions (保留，虽然目前未使用，但保持兼容性)
const (
	ResizeLeft        = 1
	ResizeRight       = 2
	ResizeTop         = 3
	ResizeTopLeft     = 4
	ResizeTopRight    = 5
	ResizeBottom      = 6
	ResizeBottomLeft  = 7
	ResizeBottomRight = 8
)

func GetCursorPos() (int, int) {
	var pt POINT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	return int(pt.X), int(pt.Y)
}

var (
	procGetSystemMetrics      = user32.NewProc("GetSystemMetrics")
	procSystemParametersInfoW = user32.NewProc("SystemParametersInfoW")
)

const (
	SM_CXSCREEN     = 0
	SM_CYSCREEN     = 1
	SPI_GETWORKAREA = 0x0030
)

func GetScreenSize() (int, int) {
	w, _, _ := procGetSystemMetrics.Call(uintptr(SM_CXSCREEN))
	h, _, _ := procGetSystemMetrics.Call(uintptr(SM_CYSCREEN))
	return int(w), int(h)
}

// GetWorkArea 获取不包含任务栏的屏幕工作区
func GetWorkArea() (int, int, int, int) {
	var rect RECT
	procSystemParametersInfoW.Call(SPI_GETWORKAREA, 0, uintptr(unsafe.Pointer(&rect)), 0)
	return int(rect.Left), int(rect.Top), int(rect.Right), int(rect.Bottom)
}

// IsSystemDarkMode 检测 Windows 系统是否使用深色模式
func IsSystemDarkMode() bool {
	// 读取注册表：HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Themes\Personalize
	// 键名：AppsUseLightTheme
	// 值：0 = 深色模式，1 = 浅色模式

	advapi32 := syscall.NewLazyDLL("advapi32.dll")
	procRegOpenKeyEx := advapi32.NewProc("RegOpenKeyExW")
	procRegQueryValueEx := advapi32.NewProc("RegQueryValueExW")
	procRegCloseKey := advapi32.NewProc("RegCloseKey")

	const (
		HKEY_CURRENT_USER = 0x80000001
		KEY_READ          = 0x20019
	)

	// 打开注册表键
	keyPath, _ := syscall.UTF16PtrFromString(`Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`)
	var hKey uintptr
	ret, _, _ := procRegOpenKeyEx.Call(
		uintptr(HKEY_CURRENT_USER),
		uintptr(unsafe.Pointer(keyPath)),
		0,
		KEY_READ,
		uintptr(unsafe.Pointer(&hKey)),
	)

	if ret != 0 {
		// 无法打开注册表，默认返回深色模式
		return true
	}
	defer procRegCloseKey.Call(hKey)

	// 查询值
	valueName, _ := syscall.UTF16PtrFromString("AppsUseLightTheme")
	var dataType uint32
	var data uint32
	dataSize := uint32(4)

	ret, _, _ = procRegQueryValueEx.Call(
		hKey,
		uintptr(unsafe.Pointer(valueName)),
		0,
		uintptr(unsafe.Pointer(&dataType)),
		uintptr(unsafe.Pointer(&data)),
		uintptr(unsafe.Pointer(&dataSize)),
	)

	if ret != 0 {
		// 无法读取值，默认返回深色模式
		return true
	}

	// 0 = 深色模式，1 = 浅色模式
	return data == 0
}

// IsWindowMaximized 检测窗口是否处于最大化状态
func IsWindowMaximized(hwnd uintptr) bool {
	style := GetWindowLong(hwnd, GWL_STYLE)
	// 检查 WS_MAXIMIZE 标志
	return (style & WS_MAXIMIZE) != 0
}

// SetWindowOpacity sets the window opacity (0.0 = fully transparent, 1.0 = fully opaque)
func SetWindowOpacity(hwnd uintptr, opacity float64) {
	// Clamp opacity to valid range
	if opacity < 0.0 {
		opacity = 0.0
	}
	if opacity > 1.0 {
		opacity = 1.0
	}

	// Convert to alpha value (0-255)
	alpha := uint8(opacity * 255)

	const (
		WS_EX_LAYERED = 0x00080000
		LWA_ALPHA     = 0x00000002
	)

	procSetLayeredWindowAttributes := user32.NewProc("SetLayeredWindowAttributes")

	// Get current extended style
	exStyle := GetWindowLong(hwnd, GWL_EXSTYLE)

	// Add WS_EX_LAYERED style if not present
	if exStyle&WS_EX_LAYERED == 0 {
		SetWindowLong(hwnd, GWL_EXSTYLE, exStyle|WS_EX_LAYERED)
	}

	// Set the opacity
	procSetLayeredWindowAttributes.Call(hwnd, 0, uintptr(alpha), LWA_ALPHA)
}

// SetWindowNoTaskbar 设置窗口不在任务栏显示
func SetWindowNoTaskbar(hwnd uintptr) {
	const (
		WS_EX_TOOLWINDOW = 0x00000080
		WS_EX_APPWINDOW  = 0x00040000
	)

	// Get current extended style
	exStyle := GetWindowLong(hwnd, GWL_EXSTYLE)

	// Add WS_EX_TOOLWINDOW style and remove WS_EX_APPWINDOW to prevent showing in taskbar
	newStyle := (exStyle | WS_EX_TOOLWINDOW) &^ WS_EX_APPWINDOW
	SetWindowLong(hwnd, GWL_EXSTYLE, newStyle)
}