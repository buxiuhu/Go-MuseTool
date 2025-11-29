package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// setupEscapeKeyClose 为窗口设置ESC键关闭功能
// closeFunc 是关闭窗口时要执行的函数
func setupEscapeKeyClose(win fyne.Window, closeFunc func()) {
	// 创建一个自定义的Canvas来捕获键盘事件
	canvas := win.Canvas()

	// 设置键盘快捷键
	canvas.SetOnTypedKey(func(ev *fyne.KeyEvent) {
		if ev.Name == fyne.KeyEscape {
			if closeFunc != nil {
				closeFunc()
			}
		}
	})
}

// setupEscapeKeyCloseWithShortcut 为窗口设置ESC键关闭功能(使用快捷键方式)
// closeFunc 是关闭窗口时要执行的函数
func setupEscapeKeyCloseWithShortcut(win fyne.Window, closeFunc func()) {
	// 创建ESC快捷键
	escShortcut := &desktop.CustomShortcut{
		KeyName: fyne.KeyEscape,
	}

	// 添加快捷键处理
	win.Canvas().AddShortcut(escShortcut, func(shortcut fyne.Shortcut) {
		if closeFunc != nil {
			closeFunc()
		}
	})
}
