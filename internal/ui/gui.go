package ui

import (
	_ "embed"
	"fmt"
	"image/color"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go-musetool/internal/language"
	"go-musetool/internal/launcher"
	"go-musetool/internal/logger"
	"go-musetool/internal/model"
	"go-musetool/internal/storage"
	"go-musetool/internal/version"

	nativeDialog "github.com/sqweek/dialog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type LauncherApp struct {
	App                        fyne.App
	Window                     fyne.Window
	Config                     *model.Config
	ConfigPath                 string
	CurrentGroup               string
	SettingsWindow             fyne.Window // 设置窗口引用
	ShortcutWindow             fyne.Window // 快捷方式窗口引用
	AddGroupWindow             fyne.Window // 新增分组窗口引用
	EditGroupWindow            fyne.Window // 编辑分组窗口引用
	DeleteGroupWindow          fyne.Window // 删除分组窗口引用
	EditGroupDialogForWindow   fyne.Window // 编辑分组对话框(For)窗口引用
	DeleteGroupDialogForWindow fyne.Window // 删除分组对话框(For)窗口引用
	DeleteShortcutWindow       fyne.Window // 删除快捷方式确认窗口引用
	AboutWindow                fyne.Window // 关于窗口引用
	MainWindowIconData         []byte
	isTrueFullscreen           bool
	preFullscreenState         struct {
		x, y, w, h int
		style      uintptr
	}
}

// SetMainWindowIconData 设置主窗口图标数据
func (l *LauncherApp) SetMainWindowIconData(data []byte) {
	l.MainWindowIconData = data
}

// getMainWindowIconData 获取主窗口图标数据（用于托盘）
func (l *LauncherApp) getMainWindowIconData() []byte {
	return l.MainWindowIconData
}

// getTitleBarColor 获取当前主题对应的标题栏颜色（RGB格式）
func (l *LauncherApp) getTitleBarColor() (uint8, uint8, uint8) {
	var color uint32
	if l.Config.TitleBarColor != 0 {
		color = l.Config.TitleBarColor
	} else {
		switch l.Config.ThemePreference {
		case model.ThemeDark:
			color = 0x202020 // BGR format - 深灰色
		case model.ThemeLight:
			color = 0xF0F0F0 // BGR format - 浅灰色
		default:
			// 跟随系统：检测系统深色模式
			if IsSystemDarkMode() {
				color = 0x202020 // 深色模式
			} else {
				color = 0xF0F0F0 // 浅色模式
			}
		}
	}
	// Convert BGR to RGB
	b := uint8(color & 0xFF)
	g := uint8((color >> 8) & 0xFF)
	r := uint8((color >> 16) & 0xFF)
	return r, g, b
}

// applyWindowStyle 应用窗口样式（置顶 + 标题栏颜色）
func (l *LauncherApp) applyWindowStyle(windowTitle string) {
	// 如果主窗口处于全屏模式，则不应用任何样式更改
	if l.Window != nil && l.Window.FullScreen() {
		return
	}
	go func() {
		time.Sleep(100 * time.Millisecond)
		hwnd := GetWindowHandle(windowTitle)
		if hwnd != 0 {
			// 设置窗口总在最上层（先关闭再打开以确保正确的z-order）
			SetWindowAlwaysOnTop(hwnd, false)
			SetWindowAlwaysOnTop(hwnd, true)

			// 根据当前主题设置标题栏颜色
			var color uint32
			switch l.Config.ThemePreference {
			case model.ThemeDark:
				color = 0x202020 // 深灰色
			case model.ThemeLight:
				color = 0xF0F0F0 // 浅灰色
			default:
				if IsSystemDarkMode() {
					color = 0x202020
				} else {
					color = 0xF0F0F0
				}
			}
			SetTitleBarColor(hwnd, color)

			// 设置窗口不在任务栏显示
			SetWindowNoTaskbar(hwnd)
		}
	}()
}

func NewLauncherApp(config *model.Config, configPath string, iconData []byte) *LauncherApp {
	log.Println("creating fyne app...")
	a := app.New()

	l := &LauncherApp{
		App:                a,
		Config:             config,
		ConfigPath:         configPath,
		MainWindowIconData: iconData, // 在初始化时就设置图标数据
	}

	// 0. 加载语言配置
	if err := language.Load(config.Language); err != nil {
		log.Printf("failed to load language: %v", err)
		// Fallback to English if loading fails
		_ = language.Load("en")
	}

	// 1. 在创建窗口之前应用主题设置
	l.applyTheme()

	log.Println("creating window...")
	w := a.NewWindow(language.T().WindowTitle)
	l.Window = w

	/* w.SetOnFullScreenChanged(func(fullscreen bool) {
		hwnd := GetWindowHandle(language.T().WindowTitle)
		if hwnd == 0 {
			return
		}

		if fullscreen {
			// Entering true fullscreen
			l.isTrueFullscreen = true

			// Save original window style and rect
			x, y, w, h := GetWindowRect(hwnd)
			l.preFullscreenState.x = x
			l.preFullscreenState.y = y
			l.preFullscreenState.w = w
			l.preFullscreenState.h = h
			l.preFullscreenState.style = GetWindowLong(hwnd, GWL_STYLE)

			// Get screen dimensions for the current monitor
			screenWidth, screenHeight := GetScreenSize()

			// Change window style to a borderless popup
			SetWindowLong(hwnd, GWL_STYLE, l.preFullscreenState.style&^WS_OVERLAPPEDWINDOW|WS_POPUP)

			// Use SetWindowPos to resize, move, and set topmost in one atomic call
			procSetWindowPos.Call(hwnd, uintptr(HWND_TOPMOST), uintptr(0), uintptr(0), uintptr(screenWidth), uintptr(screenHeight), SWP_FRAMECHANGED)

		} else {
			// Exiting true fullscreen
			if !l.isTrueFullscreen {
				return
			}
			l.isTrueFullscreen = false

			// Restore original window style
			SetWindowLong(hwnd, GWL_STYLE, l.preFullscreenState.style)

			// Restore original window size, position, and Z-order
			procSetWindowPos.Call(hwnd, uintptr(HWND_NOTOPMOST), uintptr(l.preFullscreenState.x), uintptr(l.preFullscreenState.y), uintptr(l.preFullscreenState.w), uintptr(l.preFullscreenState.h), SWP_FRAMECHANGED)
		}
	}) */

	// 2. 设置系统托盘 (使用原生 Windows API 以支持左键点击)
	// 注意: 我们不再使用 Fyne 的 desk.SetSystemTrayMenu，而是使用自定义的 InitTray
	// 使用与主窗口相同的图标数据,确保托盘图标和主窗口图标一致
	// Create a channel to handle show requests from tray
	// This decouples the OS-locked tray thread from Fyne's thread handling
	trayShowChan := make(chan struct{}, 1)
	go func() {
		for range trayShowChan {
			// fyne.Do schedules the function on the main UI thread
			fyne.Do(func() {
				if l.Window != nil {
					l.Window.Show()
					l.Window.RequestFocus()
				}
			})
		}
	}()

	InitTrayWithData(l.getMainWindowIconData(), func() {
		// OnShow callback (Left click or Menu Show)
		// Non-blocking send to channel
		select {
		case trayShowChan <- struct{}{}:
		default:
		}
	}, func() {
		// OnExit callback (Menu Exit)
		l.saveWindowState()
		a.Quit()
	})
	log.Println("Native Windows system tray initialized")

	// 4. 从配置加载窗口大小和位置
	// 注意：必须使用 Windows API 恢复大小，因为 Fyne 的 Resize 设置的是内容大小，而我们保存的是窗口总大小
	if config.WindowWidth > 0 && config.WindowHeight > 0 && config.WindowX >= 0 && config.WindowY >= 0 {
		go func() {
			// 等待窗口创建完成
			time.Sleep(200 * time.Millisecond)
			hwnd := GetWindowHandle(language.T().WindowTitle)
			if hwnd == 0 {
				return
			}

			// 验证位置是否在屏幕范围内
			sw, sh := GetScreenSize()
			x, y := config.WindowX, config.WindowY
			w, h := config.WindowWidth, config.WindowHeight

			// 确保窗口至少有一部分在屏幕内
			minVisible := 100
			if x > sw-minVisible {
				x = sw - minVisible
			}
			if x < -w+minVisible {
				x = -w + minVisible
			}
			if y > sh-minVisible {
				y = sh - minVisible
			}
			if y < 0 {
				y = 0
			}

			// 使用 WinAPI 同时设置位置和大小
			MoveAndResizeWindow(hwnd, x, y, w, h)
		}()
	} else {
		// No valid window geometry, use default size
		w.Resize(fyne.NewSize(800, 600))
		w.CenterOnScreen()
	}

	// 启动定期强制置顶的 goroutine (独立于窗口位置恢复逻辑)
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond) // 提高检查频率以获得更好的响应性
		for range ticker.C {
			// 如果窗口处于全屏模式，则跳过所有置顶逻辑
			if l.Window != nil && l.Window.FullScreen() {
				continue
			}

			// 检查是否有任何子窗口存在 (使用内部引用而不是系统调用)
			hasChildWindow := l.SettingsWindow != nil ||
				l.ShortcutWindow != nil ||
				l.AddGroupWindow != nil ||
				l.EditGroupWindow != nil ||
				l.DeleteGroupWindow != nil ||
				l.EditGroupDialogForWindow != nil ||
				l.DeleteGroupDialogForWindow != nil ||
				l.DeleteShortcutWindow != nil ||
				l.AboutWindow != nil

			// 获取主窗口句柄
			mainHwnd := GetWindowHandle(language.T().WindowTitle)
			if mainHwnd == 0 {
				continue
			}

			if hasChildWindow {
				// 如果有子窗口存在，显式禁用主窗口的置顶状态
				SetWindowAlwaysOnTop(mainHwnd, false)
			} else {
				// 只有在所有子窗口都不存在时，才强制主窗口置顶
				SetWindowAlwaysOnTop(mainHwnd, true)
			}
		}
	}()

	// 3. 设置窗口关闭事件处理
	// 使用 SetCloseIntercept 而不是 SetOnClosed,避免递归调用问题
	w.SetCloseIntercept(func() {
		// 使用 logger.Debug 代替直接写入标准输出,避免非调试模式下的噪音
		logger.Debug("SetCloseIntercept: CloseDialogShown=%v", l.Config.CloseDialogShown)

		// 如果用户已经选择过行为,直接执行
		if l.Config.CloseDialogShown {
			logger.Debug("Dialog already shown, executing direct action. MinimizeToTray=%v", l.Config.MinimizeToTray)
			if l.Config.MinimizeToTray {
				// 最小化到托盘
				l.Window.Hide()
			} else {
				// 退出程序
				l.saveWindowState()
				l.App.Quit()
			}
			return
		}

		// 首次关闭,显示选择对话框
		logger.Debug("SetCloseIntercept: Showing close dialog for first time")

		rememberCheck := widget.NewCheck(language.T().CloseDialogRemember, func(b bool) {})
		rememberCheck.SetChecked(true) // 默认选中记住

		var closeDialog dialog.Dialog

		content := container.NewVBox(
			widget.NewLabel(language.T().CloseDialogMessage),
			rememberCheck,
			container.NewHBox(
				layout.NewSpacer(),
				widget.NewButton(language.T().CloseDialogMinimize, func() {
					logger.Debug("SetCloseIntercept: User selected Minimize to Tray")
					if rememberCheck.Checked {
						l.Config.CloseDialogShown = true
						l.Config.MinimizeToTray = true
						storage.SaveConfig(l.ConfigPath, l.Config)
					}
					closeDialog.Hide()
					// 执行最小化
					l.Window.Hide()
				}),
				widget.NewButton(language.T().CloseDialogExit, func() {
					logger.Debug("SetCloseIntercept: User selected Exit")
					if rememberCheck.Checked {
						l.Config.CloseDialogShown = true
						l.Config.MinimizeToTray = false
						storage.SaveConfig(l.ConfigPath, l.Config)
					}
					closeDialog.Hide()
					// 执行退出
					l.saveWindowState()
					l.App.Quit()
				}),
				layout.NewSpacer(),
			),
		)

		closeDialog = dialog.NewCustom(
			language.T().CloseDialogTitle,
			language.T().Cancel,
			content,
			w,
		)
		closeDialog.SetDismissText(language.T().Cancel)
		logger.Debug("SetCloseIntercept: About to show close dialog")
		closeDialog.Show()
		logger.Debug("SetCloseIntercept: Close dialog shown")
	})

	log.Println("setting up ui content...")
	l.setupUI()
	log.Println("ui content setup complete.")
	return l
}

// 辅助函数：根据配置应用主题
func (l *LauncherApp) applyTheme() {
	// 1. 设置 Fyne 内部主题 (Must run on UI thread)
	// Use fyne.Do to schedule on UI thread as suggested by error message
	fyne.Do(func() {
		switch l.Config.ThemePreference {
		case model.ThemeDark:
			l.App.Settings().SetTheme(theme.DarkTheme())
		case model.ThemeLight:
			l.App.Settings().SetTheme(theme.LightTheme())
		default: // 默认跟随系统
			l.App.Settings().SetTheme(theme.DefaultTheme())
		}

		// 3. 强制重新绘制窗口 (Must run on UI thread)
		if l.Window != nil && l.Window.Content() != nil {
			l.Window.Canvas().Refresh(l.Window.Content())
		}
	})

	// 2. 设置系统标题栏颜色 (Runs in its own goroutine, independent of Fyne UI thread)
	if l.Window != nil {
		go func() {
			// 等待窗口创建并获取句柄
			// 重试几次以确保窗口已创建
			// 重试几次以确保窗口已创建
			var hwnd uintptr
			for i := 0; i < 10; i++ {
				hwnd = GetWindowHandle(language.T().WindowTitle)
				if hwnd != 0 {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}

			if hwnd != 0 {
				var color uint32
				// Windows DWM 颜色使用 BGR 格式 (0xBBGGRR)
				if l.Config.TitleBarColor != 0 {
					color = l.Config.TitleBarColor
				} else {
					switch l.Config.ThemePreference {
					case model.ThemeDark:
						// 深灰色 (RGB 0x202020 -> BGR 0x202020)
						color = 0x202020
					case model.ThemeLight:
						// 浅灰色 (RGB 0xF0F0F0 -> BGR 0xF0F0F0)
						color = 0xF0F0F0
					default:
						// 跟随系统：检测系统深色模式
						if IsSystemDarkMode() {
							color = 0x202020 // 深色模式
						} else {
							color = 0xF0F0F0 // 浅色模式
						}
					}
				}

				SetTitleBarColor(hwnd, color)
				log.Printf("applied title bar color: 0x%X", color)

				// 应用配置中的窗口透明度
				opacity := l.Config.Opacity
				if opacity <= 0 || opacity > 1.0 {
					opacity = 1.0 // 默认不透明（仅当未设置或无效时）
				}
				SetWindowOpacity(hwnd, opacity)
				log.Printf("applied window opacity: %.2f", opacity)

				// 如果窗口不是全屏模式，则设置主窗口不在任务栏显示，避免与全屏模式冲突
				if !l.Window.FullScreen() {
					SetWindowNoTaskbar(hwnd)
					log.Printf("SetWindowNoTaskbar applied (not fullscreen)")
				} else {
					log.Printf("SetWindowNoTaskbar skipped (fullscreen)")
				}

				// 确保窗口置顶
				SetWindowAlwaysOnTop(hwnd, true)
			}
		}()
	}
}

// setupUI 设置主界面布局
func (l *LauncherApp) setupUI() {
	// 确保在每次重新创建 UI 时都应用主题
	l.applyTheme()

	// 1. 创建内容区域容器
	contentContainer := container.NewMax()

	// 2. 确保 CurrentGroup 有效
	if len(l.Config.Groups) == 0 {
		l.Config.Groups = append(l.Config.Groups, model.Group{Name: "Default"})
	}
	groupExists := false
	for _, g := range l.Config.Groups {
		if g.Name == l.CurrentGroup {
			groupExists = true
			break
		}
	}
	if !groupExists && len(l.Config.Groups) > 0 {
		l.CurrentGroup = l.Config.Groups[0].Name
	}

	// 3. 定义切换分组的函数
	var refreshTabBar func() // 前向声明
	switchGroup := func(groupName string) {
		l.CurrentGroup = groupName
		// 更新内容区域
		for _, g := range l.Config.Groups {
			if g.Name == groupName {
				contentContainer.Objects = []fyne.CanvasObject{l.createGroupContent(g)}
				contentContainer.Refresh()
				break
			}
		}
		// 更新 Tab 栏选中状态
		if refreshTabBar != nil {
			refreshTabBar()
		}
	}

	// 初始化显示当前分组内容
	switchGroup(l.CurrentGroup)

	// 4. 创建自定义 Tab 栏
	var tabBarContainer *fyne.Container

	createTabBar := func() {
		var buttons []fyne.CanvasObject

		for i, g := range l.Config.Groups {
			groupName := g.Name
			groupIndex := i // 捕获索引
			// 使用自定义的 TabButton 支持右键菜单
			// 捕获 groupName 到局部变量避免闭包问题
			targetGroup := groupName
			btn := NewTabButton(groupName, nil, func() {
				switchGroup(targetGroup)
			}, func(e *fyne.PointEvent) {
				// 右键菜单：编辑、删除、移动
				var menuItems []*fyne.MenuItem

				// 向左移动
				if groupIndex > 0 {
					menuItems = append(menuItems, fyne.NewMenuItem(language.T().ContextMenuMoveLeft, func() {
						l.reorderGroup(groupIndex, groupIndex-1)
					}))
				}

				// 向右移动
				if groupIndex < len(l.Config.Groups)-1 {
					menuItems = append(menuItems, fyne.NewMenuItem(language.T().ContextMenuMoveRight, func() {
						l.reorderGroup(groupIndex, groupIndex+1)
					}))
				}

				menuItems = append(menuItems,
					fyne.NewMenuItem(language.T().ContextMenuRenameGroup, func() {
						l.showEditGroupDialogFor(targetGroup)
					}),
					fyne.NewMenuItem(language.T().ContextMenuDeleteGroup, func() {
						l.showDeleteGroupDialogFor(targetGroup)
					}),
				)

				menu := fyne.NewMenu("", menuItems...)
				widget.ShowPopUpMenuAtPosition(menu, l.Window.Canvas(), e.AbsolutePosition)
			}, func(startPos, endPos fyne.Position) {
				// 拖拽结束回调：计算目标索引
				// 根据拖拽方向和距离判断目标位置
				var dragDistance float32
				if l.Config.TabPosition == "left" || l.Config.TabPosition == "right" {
					dragDistance = endPos.Y - startPos.Y
				} else {
					dragDistance = endPos.X - startPos.X
				}

				// 估算每个按钮的大小（这是一个简化的实现）
				buttonSize := float32(100) // 假设每个按钮约100像素
				offset := int(dragDistance / buttonSize)

				if offset != 0 {
					targetIndex := groupIndex + offset
					if targetIndex < 0 {
						targetIndex = 0
					}
					if targetIndex >= len(l.Config.Groups) {
						targetIndex = len(l.Config.Groups) - 1
					}
					l.reorderGroup(groupIndex, targetIndex)
				}
			})

			// 选中状态样式：HighImportance (填充)，未选中：LowImportance (扁平)
			if groupName == l.CurrentGroup {
				btn.Importance = widget.HighImportance
			} else {
				btn.Importance = widget.LowImportance
			}
			buttons = append(buttons, btn)
		}

		// 创建按钮容器
		var buttonsContainer *fyne.Container
		if l.Config.TabPosition == "left" || l.Config.TabPosition == "right" {
			buttonsContainer = container.NewVBox(buttons...)
		} else {
			buttonsContainer = container.NewHBox(buttons...)
		}

		// 创建可右键点击的背景，颜色与标题栏一致
		r, g, b := l.getTitleBarColor()
		background := NewTappableBackground(color.NRGBA{R: r, G: g, B: b, A: 255}, func(e *fyne.PointEvent) {
			// 右键菜单：新建分组
			menu := fyne.NewMenu("",
				fyne.NewMenuItem(language.T().ContextMenuNewGroup, func() {
					l.showAddGroupDialog()
				}),
			)
			widget.ShowPopUpMenuAtPosition(menu, l.Window.Canvas(), e.AbsolutePosition)
		})

		// 使用 Max 容器叠加背景和按钮
		tabBarContainer = container.NewMax(background, buttonsContainer)
	}

	// 首次创建
	createTabBar()

	// 绑定刷新函数，以便点击时更新按钮样式
	refreshTabBar = func() {
		var newButtons []fyne.CanvasObject
		for i, g := range l.Config.Groups {
			groupName := g.Name
			groupIndex := i // 捕获索引
			btn := NewTabButton(groupName, nil, func() {
				switchGroup(groupName)
			}, func(e *fyne.PointEvent) {
				// 右键菜单：编辑、删除、移动
				var menuItems []*fyne.MenuItem

				// 向左移动
				if groupIndex > 0 {
					menuItems = append(menuItems, fyne.NewMenuItem(language.T().ContextMenuMoveLeft, func() {
						l.reorderGroup(groupIndex, groupIndex-1)
					}))
				}

				// 向右移动
				if groupIndex < len(l.Config.Groups)-1 {
					menuItems = append(menuItems, fyne.NewMenuItem(language.T().ContextMenuMoveRight, func() {
						l.reorderGroup(groupIndex, groupIndex+1)
					}))
				}

				menuItems = append(menuItems,
					fyne.NewMenuItem(language.T().ContextMenuRenameGroup, func() {
						l.showEditGroupDialogFor(groupName)
					}),
					fyne.NewMenuItem(language.T().ContextMenuDeleteGroup, func() {
						l.showDeleteGroupDialogFor(groupName)
					}),
				)

				menu := fyne.NewMenu("", menuItems...)
				widget.ShowPopUpMenuAtPosition(menu, l.Window.Canvas(), e.AbsolutePosition)
			}, func(startPos, endPos fyne.Position) {
				// 拖拽结束回调：计算目标索引
				var dragDistance float32
				if l.Config.TabPosition == "left" || l.Config.TabPosition == "right" {
					dragDistance = endPos.Y - startPos.Y
				} else {
					dragDistance = endPos.X - startPos.X
				}

				buttonSize := float32(100)
				offset := int(dragDistance / buttonSize)

				if offset != 0 {
					targetIndex := groupIndex + offset
					if targetIndex < 0 {
						targetIndex = 0
					}
					if targetIndex >= len(l.Config.Groups) {
						targetIndex = len(l.Config.Groups) - 1
					}
					l.reorderGroup(groupIndex, targetIndex)
				}
			})

			if groupName == l.CurrentGroup {
				btn.Importance = widget.HighImportance
			} else {
				btn.Importance = widget.LowImportance
			}
			newButtons = append(newButtons, btn)
		}

		// 重新创建按钮容器
		var buttonsContainer *fyne.Container
		if l.Config.TabPosition == "left" || l.Config.TabPosition == "right" {
			buttonsContainer = container.NewVBox(newButtons...)
		} else {
			buttonsContainer = container.NewHBox(newButtons...)
		}

		// 更新背景颜色
		r, g, b := l.getTitleBarColor()
		background := NewTappableBackground(color.NRGBA{R: r, G: g, B: b, A: 255}, func(e *fyne.PointEvent) {
			menu := fyne.NewMenu("",
				fyne.NewMenuItem(language.T().ContextMenuNewGroup, func() {
					l.showAddGroupDialog()
				}),
			)
			widget.ShowPopUpMenuAtPosition(menu, l.Window.Canvas(), e.AbsolutePosition)
		})

		// 更新容器内容
		tabBarContainer.Objects = []fyne.CanvasObject{background, buttonsContainer}
		tabBarContainer.Refresh()
	}

	// 增加内边距 (只保留一层，使图标更靠边)
	paddedContent := container.NewPadded(contentContainer)

	// 创建工具栏 (自定义菜单栏)
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), func() { l.showShortcutDialog(nil) }),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.FolderNewIcon(), func() { l.showAddGroupDialog() }),
		widget.NewToolbarAction(theme.DocumentCreateIcon(), func() { l.showEditGroupDialog() }),
		widget.NewToolbarAction(theme.DeleteIcon(), func() { l.showDeleteGroupDialog() }),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.SettingsIcon(), func() { l.showSettingsDialog() }),
		widget.NewToolbarAction(theme.HelpIcon(), func() { l.showAboutDialog() }),
		widget.NewToolbarSpacer(),
	)

	// 使用 Border 布局
	// 核心布局：Tab 栏 + 内容
	var mainLayout fyne.CanvasObject
	switch l.Config.TabPosition {
	case "left":
		mainLayout = container.NewBorder(nil, nil, tabBarContainer, nil, paddedContent)
	case "right":
		mainLayout = container.NewBorder(nil, nil, nil, tabBarContainer, paddedContent)
	case "bottom":
		mainLayout = container.NewBorder(nil, tabBarContainer, nil, nil, paddedContent)
	default: // "top"
		mainLayout = container.NewBorder(tabBarContainer, nil, nil, nil, paddedContent)
	}

	// 外层布局：工具栏固定在底部
	borderLayout := container.NewBorder(nil, toolbar, nil, nil, mainLayout)
	finalContent := NewInteractiveContainer(l.Window, borderLayout, l.saveWindowState)
	finalContent.SetDebugMode(l.Config.DebugMode)

	l.Window.SetContent(finalContent)
}

func (l *LauncherApp) createGroupContent(group model.Group) fyne.CanvasObject {
	var items []fyne.CanvasObject
	for i, s := range group.Shortcuts {
		shortcut := s      // capture loop variable
		shortcutIndex := i // 捕获索引
		btn := NewShortcutWidget(shortcut.Name, func() {
			log.Printf("launching: %s (%s)", shortcut.Name, shortcut.Path)
			if err := launcher.Open(shortcut.Path); err != nil {
				log.Printf("error launching %s: %v", shortcut.Name, err)
			}
		}, func(e *fyne.PointEvent) {
			// 构建菜单项
			menuItems := []*fyne.MenuItem{
				fyne.NewMenuItem(language.T().ContextMenuOpenLocation, func() {
					targetPath := shortcut.Path
					// 如果是 .lnk 快捷方式，解析真实路径
					if strings.HasSuffix(strings.ToLower(targetPath), ".lnk") {
						resolved := ResolveLnkTarget(targetPath)
						if resolved != "" {
							targetPath = resolved
						}
					}

					// 获取所在目录
					dirPath := filepath.Dir(targetPath)

					// check path exists
					if _, err := os.Stat(dirPath); err == nil {
						// 使用 explorer 打开目录
						exec.Command("explorer", dirPath).Start()
					} else {
						log.Printf("Directory not found: %s", dirPath)
					}
				}),
				fyne.NewMenuItem(language.T().ShortcutEdit, func() { l.showShortcutDialog(&shortcut) }),
				fyne.NewMenuItem(language.T().ShortcutDelete, func() { l.showDeleteShortcutDialog(group.Name, shortcut.Name) }),
			}

			// 如果有多个分组，添加"移动到分组"子菜单
			if len(l.Config.Groups) > 1 {
				// 创建子菜单项列表
				var moveToItems []*fyne.MenuItem
				for _, g := range l.Config.Groups {
					if g.Name != group.Name { // 排除当前分组
						targetGroupName := g.Name // 捕获变量
						moveToItems = append(moveToItems, fyne.NewMenuItem(targetGroupName, func() {
							l.moveShortcutToGroup(group.Name, targetGroupName, shortcut.Name)
						}))
					}
				}

				// 只有在有其他分组时才添加"移动到分组"菜单
				if len(moveToItems) > 0 {
					moveToMenu := fyne.NewMenuItem(language.T().ShortcutMoveTo, nil)
					moveToMenu.ChildMenu = fyne.NewMenu("", moveToItems...)
					menuItems = append(menuItems, moveToMenu)
				}
			}

			menu := fyne.NewMenu("Shortcut", menuItems...)
			widget.ShowPopUpMenuAtPosition(menu, l.Window.Canvas(), e.AbsolutePosition)
		}, func(startPos, endPos fyne.Position) {
			// 拖拽结束回调：计算目标索引
			// 网格布局中，每个图标大小约为 90x90
			gridSize := float32(90)

			// 计算拖拽的行列偏移
			colOffset := int((endPos.X - startPos.X) / gridSize)
			rowOffset := int((endPos.Y - startPos.Y) / gridSize)

			// 估算每行的列数（这是一个简化的实现，实际列数取决于窗口宽度）
			// 假设窗口宽度约为 800，减去边距后约 700，每个图标 90，约 7-8 列
			colsPerRow := 7

			// 计算总偏移
			offset := rowOffset*colsPerRow + colOffset

			if offset != 0 {
				targetIndex := shortcutIndex + offset
				if targetIndex < 0 {
					targetIndex = 0
				}
				if targetIndex >= len(group.Shortcuts) {
					targetIndex = len(group.Shortcuts) - 1
				}
				l.reorderShortcut(group.Name, shortcutIndex, targetIndex)
			}
		})
		if shortcut.IconPath != "" {
			if res, err := fyne.LoadResourceFromPath(shortcut.IconPath); err == nil {
				btn.SetIcon(res)
			} else {
				log.Printf("failed to load icon: %v", err)
			}
		}
		items = append(items, btn)
	}

	// 网格布局设置
	grid := container.New(layout.NewGridWrapLayout(fyne.NewSize(90, 90)), items...)

	paddedGrid := container.NewPadded(
		container.NewPadded(grid),
	)

	return container.NewScroll(paddedGrid)
}

// showSettingsDialog 用于主题配置
func (l *LauncherApp) showSettingsDialog() {
	// Ensure single settings window
	if l.SettingsWindow != nil {
		// 重新应用窗口样式确保置顶
		l.applyWindowStyle(language.T().SettingsTitle)
		l.SettingsWindow.Show()
		l.SettingsWindow.RequestFocus()
		return
	}

	// Theme Settings
	themeOptions := []string{language.T().ThemeSystem, language.T().ThemeLight, language.T().ThemeDark}
	themeSelect := widget.NewSelect(themeOptions, func(selected string) {})
	currentTheme := l.Config.ThemePreference
	if currentTheme == "" {
		currentTheme = model.ThemeSystem
	}
	// Map internal values to display values
	switch currentTheme {
	case model.ThemeSystem:
		themeSelect.SetSelected(language.T().ThemeSystem)
	case model.ThemeLight:
		themeSelect.SetSelected(language.T().ThemeLight)
	case model.ThemeDark:
		themeSelect.SetSelected(language.T().ThemeDark)
	}

	// Tab Position Settings
	posOptions := []string{language.T().TabPosTop, language.T().TabPosLeft, language.T().TabPosRight}
	posSelect := widget.NewSelect(posOptions, func(selected string) {})
	// Map internal values to display values
	switch l.Config.TabPosition {
	case "top":
		posSelect.SetSelected(language.T().TabPosTop)
	case "left":
		posSelect.SetSelected(language.T().TabPosLeft)
	case "right":
		posSelect.SetSelected(language.T().TabPosRight)
	default:
		posSelect.SetSelected(language.T().TabPosTop)
	}

	// Language Settings
	langOptions := []string{"English", "中文"}
	langSelect := widget.NewSelect(langOptions, func(selected string) {})
	if l.Config.Language == "zh" {
		langSelect.SetSelected("中文")
	} else {
		langSelect.SetSelected("English")
	}

	// Create independent window for settings
	settingsWin := l.App.NewWindow(language.T().SettingsTitle)
	settingsWin.Resize(fyne.NewSize(300, 650))
	settingsWin.CenterOnScreen()
	settingsWin.SetIcon(nil)

	// Opacity Settings
	currentOpacity := l.Config.Opacity
	if currentOpacity <= 0 || currentOpacity > 1.0 {
		currentOpacity = 1.0 // 默认不透明（仅当未设置或无效时）
	}

	opacitySlider := widget.NewSlider(0.3, 1.0)
	opacitySlider.Step = 0.05

	opacityLabel := widget.NewLabel(fmt.Sprintf(language.T().SettingsOpacity, currentOpacity*100))
	// opacityHint removed - no longer displayed

	// Update opacity slider callback - set this BEFORE setting the value
	opacitySlider.OnChanged = func(value float64) {
		opacityLabel.SetText(fmt.Sprintf(language.T().SettingsOpacity, value*100))
		// 实时预览透明度效果 - 只应用到主窗口
		go func() {
			mainHwnd := GetWindowHandle("Go MuseTool")
			if mainHwnd != 0 {
				SetWindowOpacity(mainHwnd, value)
			}
		}()
	}

	// Set the value AFTER OnChanged is registered
	opacitySlider.SetValue(currentOpacity)

	// Apply current theme to settings window
	l.applyWindowStyle(language.T().SettingsTitle)

	// Debug Mode
	debugCheck := widget.NewCheck(language.T().SettingsDebugLog, func(checked bool) {})
	debugCheck.SetChecked(l.Config.DebugMode)

	// Auto Start
	autoStartCheck := widget.NewCheck(language.T().SettingsAutoStart, func(checked bool) {})
	autoStartCheck.SetChecked(l.Config.AutoStart)

	// Minimize to Tray
	minimizeToTrayCheck := widget.NewCheck(language.T().SettingsMinimizeToTray, func(checked bool) {})
	minimizeToTrayCheck.SetChecked(l.Config.MinimizeToTray)

	// Reset Close Dialog Button
	// resetCloseDialogDesc removed - no longer displayed

	resetCloseDialogBtn := widget.NewButton(language.T().SettingsResetCloseDialog, func() {
		l.Config.CloseDialogShown = false
		if err := storage.SaveConfig(l.ConfigPath, l.Config); err != nil {
			log.Printf("error saving config: %v", err)
			dialog.ShowError(fmt.Errorf("failed to save config: %w", err), settingsWin)
		} else {
			dialog.ShowInformation(language.T().Success, language.T().SettingsResetCloseDialogDesc, settingsWin)
		}
	})

	// Create dialog content with all settings
	dialogContent := container.NewVBox(
		widget.NewLabel(language.T().SettingsTheme),
		themeSelect,
		widget.NewSeparator(),
		widget.NewLabel(language.T().SettingsTabPosition),
		posSelect,
		widget.NewSeparator(),
		widget.NewLabel(language.T().SettingsLanguage),
		langSelect,
		widget.NewSeparator(),
		widget.NewLabel(language.T().SettingsOpacityTitle),
		opacityLabel,
		opacitySlider,
		widget.NewSeparator(),
		debugCheck,
		autoStartCheck,
		minimizeToTrayCheck,
		widget.NewSeparator(),
		widget.NewLabel(language.T().SettingsDataManagement),
		container.NewHBox(
			widget.NewButton(language.T().SettingsExport, func() {
				// 临时禁用设置窗口的置顶状态，确保文件对话框显示在前面
				settingsHwnd := GetWindowHandle(language.T().SettingsTitle)
				if settingsHwnd != 0 {
					SetWindowAlwaysOnTop(settingsHwnd, false)
				}

				filename, err := nativeDialog.File().Title(language.T().SettingsExport).Filter("ZIP Archive", "zip").Save()

				// 恢复设置窗口的置顶状态
				if settingsHwnd != 0 {
					SetWindowAlwaysOnTop(settingsHwnd, true)
				}

				if err == nil && filename != "" {
					if !strings.HasSuffix(strings.ToLower(filename), ".zip") {
						filename += ".zip"
					}
					if err := storage.ExportConfigWithIcons(filename, l.Config); err != nil {
						dialog.ShowError(err, settingsWin)
					}
					// 导出成功后不显示提示对话框
				}
			}),
			widget.NewButton(language.T().SettingsImport, func() {
				// 临时禁用设置窗口的置顶状态，确保文件对话框显示在前面
				settingsHwnd := GetWindowHandle(language.T().SettingsTitle)
				if settingsHwnd != 0 {
					SetWindowAlwaysOnTop(settingsHwnd, false)
				}

				filename, err := nativeDialog.File().Title(language.T().SettingsImport).Filter("ZIP Archive", "zip").Load()

				// 恢复设置窗口的置顶状态
				if settingsHwnd != 0 {
					SetWindowAlwaysOnTop(settingsHwnd, true)
				}

				if err == nil && filename != "" {
					// Get app data directory from config path
					appDataDir := filepath.Dir(l.ConfigPath)

					newConfig, err := storage.ImportConfigWithIcons(filename, appDataDir)
					if err != nil {
						dialog.ShowError(fmt.Errorf("%s: %v", language.T().SettingsImportError, err), settingsWin)
						return
					}

					// Update config
					l.Config = newConfig
					// Save to default path
					if err := storage.SaveConfig(l.ConfigPath, l.Config); err != nil {
						log.Printf("error saving imported config: %v", err)
					}

					dialog.ShowInformation(language.T().Success, language.T().SettingsImportSuccess, settingsWin)

					// Refresh UI
					l.setupUI()
					// Close settings window
					settingsWin.Close()
				}
			}),
		),
		widget.NewSeparator(),
		widget.NewLabel(language.T().SettingsResetCloseDialog),
		resetCloseDialogBtn,
	)

	saveBtn := widget.NewButton(language.T().SettingsSave, func() {
		// Save Theme
		var newTheme string
		switch themeSelect.Selected {
		case language.T().ThemeSystem:
			newTheme = model.ThemeSystem
		case language.T().ThemeLight:
			newTheme = model.ThemeLight
		case language.T().ThemeDark:
			newTheme = model.ThemeDark
		}

		// Save Tab Position (as lowercase)
		var newPos string
		switch posSelect.Selected {
		case language.T().TabPosTop:
			newPos = "top"
		case language.T().TabPosLeft:
			newPos = "left"
		case language.T().TabPosRight:
			newPos = "right"
		default:
			newPos = "top"
		}

		// Save Opacity
		newOpacity := opacitySlider.Value

		// Save Language
		newLang := "en"
		if langSelect.Selected == "中文" {
			newLang = "zh"
		}

		changed := false
		if newTheme != l.Config.ThemePreference {
			l.Config.ThemePreference = newTheme
			changed = true
		}
		if newPos != l.Config.TabPosition {
			l.Config.TabPosition = newPos
			changed = true
		}
		if newOpacity != l.Config.Opacity {
			l.Config.Opacity = newOpacity
			changed = true
			// Apply opacity immediately
			go func() {
				time.Sleep(100 * time.Millisecond)
				hwnd := GetWindowHandle("Go MuseTool")
				if hwnd != 0 {
					SetWindowOpacity(hwnd, newOpacity)
				}
			}()
		}
		if newLang != l.Config.Language {
			l.Config.Language = newLang
			language.Load(newLang)
			changed = true
		}

		// Save Debug Mode
		if debugCheck.Checked != l.Config.DebugMode {
			l.Config.DebugMode = debugCheck.Checked
			logger.SetDebugEnabled(l.Config.DebugMode)
			changed = true
		}

		// Save Auto Start
		if autoStartCheck.Checked != l.Config.AutoStart {
			l.Config.AutoStart = autoStartCheck.Checked
			if err := SetAutoStart(l.Config.AutoStart); err != nil {
				log.Printf("Failed to set auto-start: %v", err)
				dialog.ShowError(fmt.Errorf("failed to set auto-start: %w", err), settingsWin)
			} else {
				changed = true
			}
		}

		// Save Minimize to Tray
		if minimizeToTrayCheck.Checked != l.Config.MinimizeToTray {
			l.Config.MinimizeToTray = minimizeToTrayCheck.Checked
			changed = true
		}

		if changed {
			if err := storage.SaveConfig(l.ConfigPath, l.Config); err != nil {
				log.Printf("error saving config: %v", err)
				dialog.ShowError(fmt.Errorf("failed to save config: %w", err), settingsWin)
				return
			}
			// 重新设置 UI 以应用更改
			l.setupUI()
		}
		l.SettingsWindow = nil
		settingsWin.Close()
	})

	cancelBtn := widget.NewButton(language.T().SettingsCancel, func() {
		l.SettingsWindow = nil
		settingsWin.Close()
	})

	buttons := container.NewHBox(
		layout.NewSpacer(),
		cancelBtn,
		saveBtn,
	)

	content := container.NewBorder(nil, buttons, nil, nil, container.NewVScroll(dialogContent))
	settingsWin.SetContent(content)
	// keep reference to enforce singleton
	l.SettingsWindow = settingsWin
	settingsWin.SetOnClosed(func() {
		l.SettingsWindow = nil
	})
	// 添加ESC键关闭功能
	setupEscapeKeyCloseWithShortcut(settingsWin, func() {
		settingsWin.Close()
	})
	settingsWin.Show()
	// 窗口显示后应用样式,确保子窗口在主窗口前面
	l.applyWindowStyle(language.T().SettingsTitle)
}

// showAboutDialog 显示关于对话框
func (l *LauncherApp) showAboutDialog() {
	// Ensure only one about window exists
	if l.AboutWindow != nil {
		// 重新应用窗口样式确保置顶
		l.applyWindowStyle(language.T().AboutTitle)
		l.AboutWindow.Show()
		l.AboutWindow.RequestFocus()
		return
	}

	// Create independent window for about
	aboutWin := l.App.NewWindow(language.T().AboutTitle)
	aboutWin.Resize(fyne.NewSize(400, 250))
	aboutWin.CenterOnScreen()
	aboutWin.SetIcon(nil)

	// Version
	versionLabel := widget.NewLabel(fmt.Sprintf(language.T().AboutVersion, version.GetVersion()))
	versionLabel.Alignment = fyne.TextAlignCenter

	// Author
	authorLabel := widget.NewLabel(fmt.Sprintf(language.T().AboutAuthor, version.Author))
	authorLabel.Alignment = fyne.TextAlignCenter

	// Project Link
	projectLinkLabel := widget.NewLabel(language.T().AboutLink + ":")
	projectLinkLabel.Alignment = fyne.TextAlignCenter

	projectLinkBtn := widget.NewButton(version.GitHubURL, func() {
		// 使用 Windows API 打开 URL
		import_cmd := exec.Command("cmd", "/c", "start", version.GitHubURL)
		_ = import_cmd.Start()
	})

	// Close button
	closeBtn := widget.NewButton(language.T().Close, func() {
		l.AboutWindow = nil
		aboutWin.Close()
	})

	// Layout - 使用 VBox 布局使三行内容间距一致
	content := container.NewVBox(
		container.NewCenter(authorLabel),
		container.NewCenter(container.NewVBox(
			projectLinkLabel,
			projectLinkBtn,
		)),
		container.NewCenter(versionLabel),
		layout.NewSpacer(),
		container.NewCenter(closeBtn),
	)

	aboutWin.SetContent(content)
	// keep reference to enforce singleton
	l.AboutWindow = aboutWin
	aboutWin.SetOnClosed(func() {
		l.AboutWindow = nil
	})
	// 添加ESC键关闭功能
	setupEscapeKeyCloseWithShortcut(aboutWin, func() {
		aboutWin.Close()
	})
	aboutWin.Show()
	// 窗口显示后应用样式，确保子窗口在主窗口前面
	l.applyWindowStyle(language.T().AboutTitle)
}

// --- 群组管理逻辑 ---

func (l *LauncherApp) showAddGroupDialog() {
	// Ensure only one add group window exists
	if l.AddGroupWindow != nil {
		// 重新应用窗口样式确保置顶
		l.applyWindowStyle(language.T().GroupAddTitle)
		l.AddGroupWindow.Show()
		l.AddGroupWindow.RequestFocus()
		return
	}

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder(language.T().GroupAddLabel)

	// Use an independent window like shortcut/settings dialogs
	addWin := l.App.NewWindow(language.T().GroupAddTitle)
	addWin.Resize(fyne.NewSize(400, 160))
	addWin.CenterOnScreen()
	addWin.SetIcon(nil)

	saveBtn := widget.NewButton(language.T().GroupAddButton, func() {
		name := nameEntry.Text
		if name == "" {
			dialog.ShowInformation(language.T().Error, language.T().GroupNameEmpty, l.Window)
			return
		}

		for _, g := range l.Config.Groups {
			if g.Name == name {
				dialog.ShowInformation(language.T().Error, fmt.Sprintf(language.T().GroupAlreadyExists, name), l.Window)
				return
			}
		}

		l.Config.Groups = append(l.Config.Groups, model.Group{Name: name, Shortcuts: []model.Shortcut{}})
		if err := storage.SaveConfig(l.ConfigPath, l.Config); err != nil {
			log.Printf("error saving config: %v", err)
		}
		l.CurrentGroup = name
		log.Printf("group added successfully: %s", name)
		l.setupUI()
		// clear singleton reference before closing
		l.AddGroupWindow = nil
		addWin.Close()
	})

	cancelBtn := widget.NewButton(language.T().SettingsCancel, func() {
		l.AddGroupWindow = nil
		addWin.Close()
	})

	buttons := container.NewHBox(
		layout.NewSpacer(),
		cancelBtn,
		saveBtn,
	)

	content := container.NewBorder(nil, buttons, nil, nil, container.NewVBox(
		widget.NewLabel(language.T().GroupAddLabel),
		nameEntry,
	))

	addWin.SetContent(content)
	// keep reference to enforce single window
	l.AddGroupWindow = addWin
	// ensure reference cleared on close
	addWin.SetOnClosed(func() {
		l.AddGroupWindow = nil
	})
	// 添加ESC键关闭功能
	setupEscapeKeyCloseWithShortcut(addWin, func() {
		addWin.Close()
	})
	addWin.Show()
	// 窗口显示后应用样式，确保子窗口在主窗口前面
	l.applyWindowStyle(language.T().GroupAddTitle)
}

func (l *LauncherApp) showDeleteGroupDialog() {
	// Ensure single group window
	if l.DeleteGroupWindow != nil {
		// 重新应用窗口样式确保置顶
		l.applyWindowStyle(language.T().GroupDeleteTitle)
		l.DeleteGroupWindow.Show()
		l.DeleteGroupWindow.RequestFocus()
		return
	}

	if len(l.Config.Groups) <= 1 {
		dialog.ShowInformation(language.T().GroupCannotDelete, language.T().GroupMustHaveOne, l.Window)
		return
	}

	delWin := l.App.NewWindow(language.T().GroupDeleteTitle)
	delWin.Resize(fyne.NewSize(420, 160))
	delWin.CenterOnScreen()
	delWin.SetIcon(nil)

	confirmBtn := widget.NewButton(language.T().Confirm, func() {
		newGroups := []model.Group{}
		for _, g := range l.Config.Groups {
			if g.Name != l.CurrentGroup {
				newGroups = append(newGroups, g)
				continue
			}
		}
		l.Config.Groups = newGroups

		if err := storage.SaveConfig(l.ConfigPath, l.Config); err != nil {
			log.Printf("error saving config: %v", err)
		}

		if len(l.Config.Groups) > 0 {
			l.CurrentGroup = l.Config.Groups[0].Name
		}

		log.Printf("group deleted successfully")
		l.setupUI()
		l.DeleteGroupWindow = nil
		delWin.Close()
	})

	cancelBtn := widget.NewButton(language.T().Cancel, func() {
		l.DeleteGroupWindow = nil
		delWin.Close()
	})

	btns := container.NewHBox(layout.NewSpacer(), cancelBtn, confirmBtn)
	content := container.NewBorder(nil, btns, nil, nil, container.NewVBox(widget.NewLabel(fmt.Sprintf(language.T().GroupDeleteConfirm, l.CurrentGroup))))

	delWin.SetContent(content)
	l.DeleteGroupWindow = delWin
	delWin.SetOnClosed(func() {
		l.DeleteGroupWindow = nil
	})
	// 添加ESC键关闭功能
	setupEscapeKeyCloseWithShortcut(delWin, func() {
		delWin.Close()
	})
	delWin.Show()
	// 窗口显示后应用样式，确保子窗口在主窗口前面
	l.applyWindowStyle(language.T().GroupDeleteTitle)
}

func (l *LauncherApp) showEditGroupDialog() {
	// Ensure single group window
	if l.EditGroupWindow != nil {
		// 重新应用窗口样式确保置顶
		l.applyWindowStyle(language.T().GroupRenameTitle)
		l.EditGroupWindow.Show()
		l.EditGroupWindow.RequestFocus()
		return
	}

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder(language.T().GroupRenameLabel)
	nameEntry.SetText(l.CurrentGroup)

	editWin := l.App.NewWindow(language.T().GroupRenameTitle)
	editWin.Resize(fyne.NewSize(420, 160))
	editWin.CenterOnScreen()
	editWin.SetIcon(nil)

	saveBtn := widget.NewButton(language.T().SettingsSave, func() {
		newName := nameEntry.Text
		if newName == "" {
			dialog.ShowInformation(language.T().Error, language.T().GroupNameEmpty, l.Window)
			return
		}

		for _, g := range l.Config.Groups {
			if g.Name == newName && newName != l.CurrentGroup {
				dialog.ShowInformation(language.T().Error, fmt.Sprintf(language.T().GroupAlreadyExists, newName), l.Window)
				return
			}
		}

		for i := range l.Config.Groups {
			if l.Config.Groups[i].Name == l.CurrentGroup {
				l.Config.Groups[i].Name = newName
				break
			}
		}

		l.CurrentGroup = newName
		if err := storage.SaveConfig(l.ConfigPath, l.Config); err != nil {
			log.Printf("error saving config: %v", err)
		}
		log.Printf("group renamed successfully to: %s", newName)
		l.setupUI()
		l.EditGroupWindow = nil
		editWin.Close()
	})

	cancelBtn := widget.NewButton(language.T().SettingsCancel, func() {
		l.EditGroupWindow = nil
		editWin.Close()
	})

	buttons := container.NewHBox(
		layout.NewSpacer(),
		cancelBtn,
		saveBtn,
	)

	content := container.NewBorder(nil, buttons, nil, nil, container.NewVBox(
		widget.NewLabel(fmt.Sprintf(language.T().GroupRenameLabel, l.CurrentGroup)),
		nameEntry,
	))

	editWin.SetContent(content)
	l.EditGroupWindow = editWin
	editWin.SetOnClosed(func() {
		l.EditGroupWindow = nil
	})
	// 添加ESC键关闭功能
	setupEscapeKeyCloseWithShortcut(editWin, func() {
		editWin.Close()
	})
	editWin.Show()
	// 窗口显示后应用样式，确保子窗口在主窗口前面
	l.applyWindowStyle(language.T().GroupRenameTitle)
}

// showEditGroupDialogFor 为指定分组显示重命名对话框
func (l *LauncherApp) showEditGroupDialogFor(groupName string) {
	// Ensure single group window
	if l.EditGroupDialogForWindow != nil {
		// 重新应用窗口样式确保置顶
		l.applyWindowStyle(language.T().GroupRenameTitle)
		l.EditGroupDialogForWindow.Show()
		l.EditGroupDialogForWindow.RequestFocus()
		return
	}

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder(language.T().GroupRenameLabel)
	nameEntry.SetText(groupName)

	editWin := l.App.NewWindow(language.T().GroupRenameTitle)
	editWin.Resize(fyne.NewSize(420, 160))
	editWin.CenterOnScreen()
	editWin.SetIcon(nil)

	saveBtn := widget.NewButton(language.T().SettingsSave, func() {
		newName := nameEntry.Text
		if newName == "" {
			dialog.ShowInformation(language.T().Error, language.T().GroupNameEmpty, l.Window)
			return
		}

		for _, g := range l.Config.Groups {
			if g.Name == newName && newName != groupName {
				dialog.ShowInformation(language.T().Error, fmt.Sprintf(language.T().GroupAlreadyExists, newName), l.Window)
				return
			}
		}

		for i := range l.Config.Groups {
			if l.Config.Groups[i].Name == groupName {
				l.Config.Groups[i].Name = newName
				break
			}
		}

		// 如果重命名的是当前分组，更新 CurrentGroup
		if l.CurrentGroup == groupName {
			l.CurrentGroup = newName
		}

		if err := storage.SaveConfig(l.ConfigPath, l.Config); err != nil {
			log.Printf("error saving config: %v", err)
		}
		log.Printf("group renamed successfully to: %s", newName)
		l.setupUI()
		l.EditGroupDialogForWindow = nil
		editWin.Close()
	})

	cancelBtn := widget.NewButton(language.T().SettingsCancel, func() {
		l.EditGroupDialogForWindow = nil
		editWin.Close()
	})

	buttons := container.NewHBox(
		layout.NewSpacer(),
		cancelBtn,
		saveBtn,
	)

	content := container.NewBorder(nil, buttons, nil, nil, container.NewVBox(
		widget.NewLabel(fmt.Sprintf(language.T().GroupRenameLabel, groupName)),
		nameEntry,
	))

	editWin.SetContent(content)
	l.EditGroupDialogForWindow = editWin
	editWin.SetOnClosed(func() {
		l.EditGroupDialogForWindow = nil
	})
	// 添加ESC键关闭功能
	setupEscapeKeyCloseWithShortcut(editWin, func() {
		editWin.Close()
	})
	editWin.Show()
	// 窗口显示后应用样式，确保子窗口在主窗口前面
	l.applyWindowStyle(language.T().GroupRenameTitle)
}

// showDeleteGroupDialogFor 为指定分组显示删除对话框
func (l *LauncherApp) showDeleteGroupDialogFor(groupName string) {
	// Ensure single group window
	if l.DeleteGroupDialogForWindow != nil {
		// 重新应用窗口样式确保置顶
		l.applyWindowStyle(language.T().GroupDeleteTitle)
		l.DeleteGroupDialogForWindow.Show()
		l.DeleteGroupDialogForWindow.RequestFocus()
		return
	}

	if len(l.Config.Groups) <= 1 {
		dialog.ShowInformation(language.T().GroupCannotDelete, language.T().GroupMustHaveOne, l.Window)
		return
	}

	delWin := l.App.NewWindow(language.T().GroupDeleteTitle)
	delWin.Resize(fyne.NewSize(420, 160))
	delWin.CenterOnScreen()
	delWin.SetIcon(nil)

	confirmBtn := widget.NewButton(language.T().Confirm, func() {
		newGroups := []model.Group{}
		for _, g := range l.Config.Groups {
			if g.Name != groupName {
				newGroups = append(newGroups, g)
			}
		}
		l.Config.Groups = newGroups

		// 如果删除的是当前分组，切换到第一个分组
		if l.CurrentGroup == groupName && len(newGroups) > 0 {
			l.CurrentGroup = newGroups[0].Name
		}

		if err := storage.SaveConfig(l.ConfigPath, l.Config); err != nil {
			log.Printf("error saving config: %v", err)
		}

		log.Printf("group deleted successfully")
		l.setupUI()
		l.DeleteGroupDialogForWindow = nil
		delWin.Close()
	})

	cancelBtn := widget.NewButton(language.T().Cancel, func() {
		l.DeleteGroupDialogForWindow = nil
		delWin.Close()
	})

	btns := container.NewHBox(layout.NewSpacer(), cancelBtn, confirmBtn)
	content := container.NewBorder(nil, btns, nil, nil, container.NewVBox(widget.NewLabel(fmt.Sprintf(language.T().GroupDeleteConfirm, groupName))))

	delWin.SetContent(content)
	l.DeleteGroupDialogForWindow = delWin
	delWin.SetOnClosed(func() {
		l.DeleteGroupDialogForWindow = nil
	})
	// 添加ESC键关闭功能
	setupEscapeKeyCloseWithShortcut(delWin, func() {
		delWin.Close()
	})
	delWin.Show()
	// 窗口显示后应用样式，确保子窗口在主窗口前面
	l.applyWindowStyle(language.T().GroupDeleteTitle)
}

// --- 快捷方式和运行时逻辑 ---

func (l *LauncherApp) showShortcutDialog(editing *model.Shortcut) {
	// Ensure single shortcut window
	if l.ShortcutWindow != nil {
		// 重新应用窗口样式确保置顶
		var title string
		if editing != nil {
			title = language.T().ShortcutEditTitle
		} else {
			title = language.T().ShortcutAddTitle
		}
		l.applyWindowStyle(title)
		l.ShortcutWindow.Show()
		l.ShortcutWindow.RequestFocus()
		return
	}

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder(language.T().ShortcutName)
	pathEntry := widget.NewEntry()
	pathEntry.SetPlaceHolder(language.T().ShortcutPath)
	iconEntry := widget.NewEntry()
	iconEntry.SetPlaceHolder(language.T().ShortcutIcon)

	title := language.T().ShortcutAddTitle
	btnText := language.T().ShortcutAdd
	var originalName string
	isEditing := false

	if editing != nil {
		originalName = editing.Name
		isEditing = true
		title = language.T().ShortcutEditTitle
		btnText = language.T().ShortcutSave
		nameEntry.SetText(editing.Name)
		pathEntry.SetText(editing.Path)
		iconEntry.SetText(editing.IconPath)
	}

	// Create independent window
	shortcutWin := l.App.NewWindow(title)
	shortcutWin.Resize(fyne.NewSize(400, 300))
	shortcutWin.CenterOnScreen()
	shortcutWin.SetIcon(nil)

	// Apply current theme (title bar color + TopMost)
	l.applyWindowStyle(title)

	browseBtn := widget.NewButton(language.T().ShortcutBrowse, func() {
		filename, err := nativeDialog.File().Title(language.T().ShortcutBrowseExe).
			Filter("Executable/Shortcut Files", "exe", "lnk").
			Filter("All Files", "*").
			Load()
		if err == nil && filename != "" {
			pathEntry.SetText(filename)

			// 直接调用同包下的函数
			if nameEntry.Text == "" {
				name, _ := GetExecutableInfo(filename)
				nameEntry.SetText(name)
			}

			iconPath := ""
			filePathLower := strings.ToLower(filename)
			if strings.HasSuffix(filePathLower, ".exe") {
				iconPath = ExtractIconFromExe(filename)
			} else if strings.HasSuffix(filePathLower, ".lnk") {
				exePath := ResolveLnkTarget(filename)
				if exePath != "" && strings.HasSuffix(strings.ToLower(exePath), ".exe") {
					iconPath = ExtractIconFromExe(exePath)
				}
			}
			iconEntry.SetText(iconPath)
		}
	})

	saveBtn := widget.NewButton(btnText, func() {
		name := nameEntry.Text
		path := pathEntry.Text
		icon := iconEntry.Text
		if name == "" || path == "" {
			log.Println("name or path is empty, cannot save shortcut")
			return
		}
		newShortcut := model.Shortcut{Name: name, Path: path, IconPath: icon}
		var err error

		if isEditing {
			err = storage.UpdateShortcut(l.Config, l.CurrentGroup, originalName, newShortcut)
		} else {
			err = storage.AddShortcut(l.Config, l.CurrentGroup, newShortcut)
		}
		if err != nil {
			log.Printf("error saving shortcut: %v", err)
			dialog.ShowError(fmt.Errorf("failed to save shortcut: %w", err), shortcutWin)
			return
		}
		if err = storage.SaveConfig(l.ConfigPath, l.Config); err != nil {
			log.Printf("error saving config: %v", err)
		}
		log.Printf("shortcut saved successfully: %s", name)
		l.setupUI()
		l.ShortcutWindow = nil
		shortcutWin.Close()
	})

	cancelBtn := widget.NewButton(language.T().SettingsCancel, func() {
		l.ShortcutWindow = nil
		shortcutWin.Close()
	})

	buttons := container.NewHBox(
		layout.NewSpacer(),
		cancelBtn,
		saveBtn,
	)

	content := container.NewBorder(
		nil,
		buttons,
		nil,
		nil,
		container.NewVBox(
			widget.NewLabel(title),
			nameEntry,
			container.NewHBox(pathEntry, browseBtn),
			iconEntry,
		),
	)

	shortcutWin.SetContent(content)
	l.ShortcutWindow = shortcutWin
	shortcutWin.SetOnClosed(func() {
		l.ShortcutWindow = nil
	})
	// 添加ESC键关闭功能
	setupEscapeKeyCloseWithShortcut(shortcutWin, func() {
		shortcutWin.Close()
	})
	shortcutWin.Show()
}

func (l *LauncherApp) deleteShortcut(groupName, shortcutName string) {
	if err := storage.RemoveShortcut(l.Config, groupName, shortcutName); err != nil {
		log.Printf("error removing shortcut: %v", err)
		return
	}
	if err := storage.SaveConfig(l.ConfigPath, l.Config); err != nil {
		log.Printf("error saving config: %v", err)
	}
	l.setupUI()
}

// showDeleteShortcutDialog 显示删除快捷方式确认对话框
func (l *LauncherApp) showDeleteShortcutDialog(groupName, shortcutName string) {
	// Ensure only one delete shortcut window exists
	if l.DeleteShortcutWindow != nil {
		// 重新应用窗口样式确保置顶
		l.applyWindowStyle(language.T().ShortcutDeleteTitle)
		l.DeleteShortcutWindow.Show()
		l.DeleteShortcutWindow.RequestFocus()
		return
	}

	delWin := l.App.NewWindow(language.T().ShortcutDeleteTitle)
	delWin.Resize(fyne.NewSize(350, 140))
	delWin.CenterOnScreen()
	delWin.SetIcon(nil)

	confirmBtn := widget.NewButton(language.T().Confirm, func() {
		if err := storage.RemoveShortcut(l.Config, groupName, shortcutName); err != nil {
			log.Printf("error removing shortcut: %v", err)
			dialog.ShowError(fmt.Errorf("failed to delete shortcut: %w", err), l.Window)
			return
		}
		if err := storage.SaveConfig(l.ConfigPath, l.Config); err != nil {
			log.Printf("error saving config: %v", err)
		}
		log.Printf("shortcut deleted successfully: %s", shortcutName)
		l.setupUI()
		l.DeleteShortcutWindow = nil
		delWin.Close()
	})

	cancelBtn := widget.NewButton(language.T().Cancel, func() {
		l.DeleteShortcutWindow = nil
		delWin.Close()
	})

	btns := container.NewHBox(layout.NewSpacer(), cancelBtn, confirmBtn)
	content := container.NewBorder(nil, btns, nil, nil, container.NewVBox(
		widget.NewLabel(fmt.Sprintf(language.T().ShortcutDeleteConfirm, shortcutName)),
	))

	delWin.SetContent(content)
	// keep reference to enforce singleton
	l.DeleteShortcutWindow = delWin
	delWin.SetOnClosed(func() {
		l.DeleteShortcutWindow = nil
	})
	// 添加ESC键关闭功能
	setupEscapeKeyCloseWithShortcut(delWin, func() {
		delWin.Close()
	})
	delWin.Show()
	// 窗口显示后应用样式，确保子窗口在主窗口前面
	l.applyWindowStyle(language.T().ShortcutDeleteTitle)
}

// reorderGroup 重新排序分组
func (l *LauncherApp) reorderGroup(fromIndex, toIndex int) {
	if fromIndex < 0 || fromIndex >= len(l.Config.Groups) || toIndex < 0 || toIndex >= len(l.Config.Groups) {
		return
	}
	if fromIndex == toIndex {
		return
	}

	// 移动分组
	group := l.Config.Groups[fromIndex]
	l.Config.Groups = append(l.Config.Groups[:fromIndex], l.Config.Groups[fromIndex+1:]...)

	// 插入到新位置
	l.Config.Groups = append(l.Config.Groups[:toIndex], append([]model.Group{group}, l.Config.Groups[toIndex:]...)...)

	// 保存配置
	if err := storage.SaveConfig(l.ConfigPath, l.Config); err != nil {
		log.Printf("error saving config: %v", err)
	}

	// 刷新UI
	l.setupUI()
}

// reorderShortcut 重新排序快捷方式
func (l *LauncherApp) reorderShortcut(groupName string, fromIndex, toIndex int) {
	// 查找分组
	for i := range l.Config.Groups {
		if l.Config.Groups[i].Name == groupName {
			shortcuts := l.Config.Groups[i].Shortcuts
			if fromIndex < 0 || fromIndex >= len(shortcuts) || toIndex < 0 || toIndex >= len(shortcuts) {
				return
			}
			if fromIndex == toIndex {
				return
			}

			// 移动快捷方式
			shortcut := shortcuts[fromIndex]
			l.Config.Groups[i].Shortcuts = append(shortcuts[:fromIndex], shortcuts[fromIndex+1:]...)

			// 插入到新位置
			l.Config.Groups[i].Shortcuts = append(l.Config.Groups[i].Shortcuts[:toIndex], append([]model.Shortcut{shortcut}, l.Config.Groups[i].Shortcuts[toIndex:]...)...)

			// 保存配置
			if err := storage.SaveConfig(l.ConfigPath, l.Config); err != nil {
				log.Printf("error saving config: %v", err)
			}

			// 刷新UI
			l.setupUI()
			return
		}
	}
}

// moveShortcutToGroup 将快捷方式从一个分组移动到另一个分组
func (l *LauncherApp) moveShortcutToGroup(fromGroup, toGroup, shortcutName string) {
	if fromGroup == toGroup {
		return
	}

	var shortcutToMove *model.Shortcut
	var fromGroupIndex int = -1
	var toGroupIndex int = -1

	// 查找源分组和目标分组
	for i := range l.Config.Groups {
		if l.Config.Groups[i].Name == fromGroup {
			fromGroupIndex = i
		}
		if l.Config.Groups[i].Name == toGroup {
			toGroupIndex = i
		}
	}

	if fromGroupIndex == -1 || toGroupIndex == -1 {
		log.Printf("source or target group not found")
		return
	}

	// 从源分组中查找并删除快捷方式
	for j, s := range l.Config.Groups[fromGroupIndex].Shortcuts {
		if s.Name == shortcutName {
			shortcutToMove = &s
			l.Config.Groups[fromGroupIndex].Shortcuts = append(
				l.Config.Groups[fromGroupIndex].Shortcuts[:j],
				l.Config.Groups[fromGroupIndex].Shortcuts[j+1:]...,
			)
			break
		}
	}

	if shortcutToMove == nil {
		log.Printf("shortcut not found in source group")
		return
	}

	// 添加到目标分组
	l.Config.Groups[toGroupIndex].Shortcuts = append(l.Config.Groups[toGroupIndex].Shortcuts, *shortcutToMove)

	// 保存配置
	if err := storage.SaveConfig(l.ConfigPath, l.Config); err != nil {
		log.Printf("error saving config: %v", err)
	}

	// 刷新UI
	l.setupUI()
}

func (l *LauncherApp) Run() {
	l.Window.SetOnDropped(l.Dropped)
	log.Println("calling window.showandrun()...")
	l.Window.ShowAndRun()
	log.Println("window closed.")
}

// Dropped handles the file drop event
func (l *LauncherApp) Dropped(_ fyne.Position, uris []fyne.URI) {
	addedCount := 0
	for _, uri := range uris {
		filePath := uri.Path()

		// 直接调用同包下的函数
		name, _ := GetExecutableInfo(filePath)

		iconPath := ""
		filePathLower := strings.ToLower(filePath)

		if strings.HasSuffix(filePathLower, ".exe") {
			iconPath = ExtractIconFromExe(filePath)
		} else if strings.HasSuffix(filePathLower, ".lnk") {
			exePath := ResolveLnkTarget(filePath)
			if exePath != "" && strings.HasSuffix(strings.ToLower(exePath), ".exe") {
				iconPath = ExtractIconFromExe(exePath)
			}
		}

		newShortcut := model.Shortcut{
			Name:     name,
			Path:     filePath,
			IconPath: iconPath,
		}

		if err := storage.AddShortcut(l.Config, l.CurrentGroup, newShortcut); err != nil {
			log.Printf("error adding dropped shortcut %s: %v", name, err)
			// Continue to next file instead of stopping
			continue
		}
		addedCount++
		log.Printf("shortcut added successfully via drag and drop: %s", name)
	}

	if addedCount > 0 {
		if err := storage.SaveConfig(l.ConfigPath, l.Config); err != nil {
			log.Printf("error saving config after drop: %v", err)
			dialog.ShowError(fmt.Errorf("failed to save config: %w", err), l.Window)
		}
		l.setupUI()
	}
}

// saveWindowState 保存窗口状态到配置文件
func (l *LauncherApp) saveWindowState() {
	// 如果窗口处于全屏模式，则不保存窗口状态
	if l.Window.FullScreen() {
		log.Println("window is in fullscreen mode, skipping state save.")
		return
	}

	hwnd := GetWindowHandle(language.T().WindowTitle)
	if hwnd == 0 {
		log.Println("failed to get window handle for saving state")
		return
	}

	// 获取窗口位置和大小
	x, y, w, h := GetWindowRect(hwnd)

	// 更新配置
	l.Config.WindowX = x
	l.Config.WindowY = y
	l.Config.WindowWidth = w
	l.Config.WindowHeight = h

	// 保存到文件
	if err := storage.SaveConfig(l.ConfigPath, l.Config); err != nil {
		log.Printf("error saving window state: %v", err)
	} else {
		log.Printf("window state saved: x=%d, y=%d, w=%d, h=%d", x, y, w, h)
	}
}

// Temporary file to append restart function
// restartApplication restarts the application
func (l *LauncherApp) restartApplication() {
	// Save window state before restarting
	l.saveWindowState()

	// Get the current executable path
	exePath, err := os.Executable()
	if err != nil {
		log.Printf("failed to get executable path: %v", err)
		dialog.ShowError(fmt.Errorf("无法重启程序: %v", err), l.Window)
		return
	}

	// Start a new instance of the application
	cmd := exec.Command(exePath)
	cmd.Dir = filepath.Dir(exePath)

	if err := cmd.Start(); err != nil {
		log.Printf("failed to start new instance: %v", err)
		dialog.ShowError(fmt.Errorf("无法重启程序: %v", err), l.Window)
		return
	}

	// Exit the current instance
	l.App.Quit()
}
