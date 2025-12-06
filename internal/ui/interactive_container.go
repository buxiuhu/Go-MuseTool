package ui

import (
	"go-musetool/internal/language"
	"log"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// InteractiveContainer wraps the content to handle window dragging and auto-hide events.
type InteractiveContainer struct {
	widget.BaseWidget
	content fyne.CanvasObject
	window  fyne.Window

	hideTimer *time.Timer
	pollStop  chan struct{} // 用于停止轮询 Goroutine
	mutex     sync.Mutex

	dockSide         int    // 0: None, 1: Top, 2: Bottom, 3: Left, 4: Right
	isHidden         bool   // 是否处于隐藏状态
	lastX            int    // 上次窗口X位置
	lastY            int    // 上次窗口Y位置
	lastW            int    // 上次窗口宽度
	lastH            int    // 上次窗口高度
	saveStateFunc    func() // 保存状态的回调函数
	saveTimer        *time.Timer
	debugMode        bool
	isHostFullscreen *bool // 宿主窗口是否处于全屏模式
}

func (c *InteractiveContainer) SetDebugMode(enabled bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.debugMode = enabled
}

const (
	DockNone   = 0
	DockTop    = 1
	DockBottom = 2
	DockLeft   = 3
	DockRight  = 4

	SnapDistance     = 80 // 吸附到边缘的距离 (增加以提高灵活性)
	VisibleEdge      = 0  // 隐藏后可见的像素边缘 (0 表示完全隐藏)
	VisualAdjustment = 8  // 视觉校正偏移量 (用于消除 Windows 窗口阴影带来的视觉间隙)

	// BottomCenterZone 定义底部角落区域的大小（单位：像素）
	// 用于区分底部的"角落区域"和"中间区域"，以解决任务栏预览窗口导致的闪烁问题
	//
	// 工作原理：
	//   - 左下角：窗口左边缘距离屏幕左边缘 < BottomCenterZone
	//   - 右下角：窗口右边缘距离屏幕右边缘 < BottomCenterZone
	//   - 中间区域：不在左下角或右下角的底部区域
	//
	// 延迟策略：
	//   - 角落区域：50ms 快速响应（任务栏图标稀疏，预览窗口少）
	//   - 中间区域：300ms 延迟响应（任务栏图标密集，避免预览窗口干扰）
	//
	// 调整建议：
	//   - 100-150：小角落，大中间区域（适合任务栏图标非常集中的情况）
	//   - 200：默认值，平衡角落和中间区域
	//   - 300-400：大角落，小中间区域（适合只有屏幕正中央需要延迟的情况）
	//
	// 示例（假设屏幕宽度1920px）：
	//   BottomCenterZone = 200
	//   - 左下角：0-200px
	//   - 中间区域：200-1720px
	//   - 右下角：1720-1920px
	BottomCenterZone = 200
)

// NewInteractiveContainer creates a new container with the given window and content.
func NewInteractiveContainer(w fyne.Window, content fyne.CanvasObject, saveFunc func()) *InteractiveContainer {
	c := &InteractiveContainer{
		window:        w,
		content:       content,
		lastX:         -1,
		lastY:         -1,
		lastW:         -1,
		lastH:         -1,
		saveStateFunc: saveFunc,
	}
	c.ExtendBaseWidget(c)
	// 启动位置检查定时器
	c.startPositionCheck()
	return c
}

// CreateRenderer implements the fyne.Widget interface.
func (c *InteractiveContainer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.content)
}

// MouseIn implements the desktop.Hoverable interface.
func (c *InteractiveContainer) MouseIn(e *desktop.MouseEvent) {
	c.showWindow()
}

// MouseOut implements the desktop.Hoverable interface.
func (c *InteractiveContainer) MouseOut() {
	// 如果窗口处于全屏模式，则不执行任何操作
	if c.window.FullScreen() {
		return
	}

	// 增强全屏检测，通过Windows API检查窗口是否真正处于全屏状态
	hwnd := GetWindowHandle(language.T().WindowTitle)
	if hwnd != 0 {
		// 检查窗口是否最大化
		if IsWindowMaximized(hwnd) {
			return
		}

		// 获取窗口和屏幕信息
		wx, wy, ww, wh := GetWindowRect(hwnd)
		screenW, screenH := GetScreenSize()

		// 如果窗口尺寸接近全屏尺寸，则认为是全屏状态
		if ww >= screenW && wh >= screenH {
			return
		}

		// 检查窗口位置是否覆盖了整个屏幕（考虑任务栏）
		waLeft, waTop, waRight, waBottom := GetWorkArea()
		workAreaWidth := waRight - waLeft
		workAreaHeight := waBottom - waTop

		if ww >= workAreaWidth && wh >= workAreaHeight && wx <= waLeft && wy <= waTop {
			return
		}
	}

	c.mutex.Lock()
	dockSide := c.dockSide
	c.mutex.Unlock()

	// 只有当窗口已经吸附到边缘时，才在鼠标离开时隐藏窗口
	if dockSide != DockNone {
		go func() {
			// 底部居中位置使用更长延迟，避免任务栏预览窗口干扰
			// 左下角和右下角保持快速响应
			delay := 50 * time.Millisecond

			if dockSide == DockBottom {
				// 检查窗口是否在底部中间区域（非角落）
				hwnd := GetWindowHandle(language.T().WindowTitle)
				if hwnd != 0 {
					x, _, w, _ := GetWindowRect(hwnd)
					waLeft, _, waRight, _ := GetWorkArea()

					// 定义角落区域：距离左右边缘 BottomCenterZone 像素以内
					isLeftCorner := x < waLeft+BottomCenterZone
					isRightCorner := x+w > waRight-BottomCenterZone

					// 只有在中间区域才使用长延迟
					if !isLeftCorner && !isRightCorner {
						delay = 300 * time.Millisecond
					}
				}
			}

			time.Sleep(delay)

			hwnd := GetWindowHandle(language.T().WindowTitle)
			if hwnd == 0 {
				return
			}

			cx, cy := GetCursorPos()
			wx, wy, ww, wh := GetWindowRect(hwnd)

			if cx < wx || cx > wx+ww || cy < wy || cy > wy+wh {
				c.scheduleHide()
			}
		}()
	}
}

// MouseMoved implements the desktop.Hoverable interface.
func (c *InteractiveContainer) MouseMoved(e *desktop.MouseEvent) {
	// No op
}

// MouseDown implements the desktop.Mouseable interface.
func (c *InteractiveContainer) MouseDown(e *desktop.MouseEvent) {
	// 拖拽和调整大小现在由系统标题栏和边框处理
}

// MouseUp implements the desktop.Mouseable interface.
func (c *InteractiveContainer) MouseUp(e *desktop.MouseEvent) {
	// No op
}

// Dragged implements the fyne.Draggable interface.
func (c *InteractiveContainer) Dragged(e *fyne.DragEvent) {
	// Handled by System
}

// DragEnd implements the fyne.Draggable interface.
func (c *InteractiveContainer) DragEnd() {
	// 由 checkPosition 定时器处理
}

func (c *InteractiveContainer) showWindow() {
	c.mutex.Lock()

	if c.hideTimer != nil {
		c.hideTimer.Stop()
		c.hideTimer = nil
	}
	if c.pollStop != nil {
		close(c.pollStop)
		c.pollStop = nil
	}

	c.isHidden = false

	hwnd := GetWindowHandle(language.T().WindowTitle)
	if hwnd == 0 {
		c.mutex.Unlock()
		return
	}

	// 检查窗口是否处于全屏状态
	if c.window.FullScreen() {
		c.mutex.Unlock()
		return
	}

	// 通过Windows API进一步检查窗口是否处于全屏状态
	// 检查窗口是否最大化
	if IsWindowMaximized(hwnd) {
		c.mutex.Unlock()
		return
	}

	// 获取窗口和屏幕信息
	wx, wy, ww, wh := GetWindowRect(hwnd)
	screenW, screenH := GetScreenSize()

	// 如果窗口尺寸接近全屏尺寸，则认为是全屏状态
	if ww >= screenW && wh >= screenH {
		c.mutex.Unlock()
		return
	}

	// 检查窗口位置是否覆盖了整个屏幕（考虑任务栏）
	waLeft, waTop, waRight, waBottom := GetWorkArea()
	workAreaWidth := waRight - waLeft
	workAreaHeight := waBottom - waTop

	if ww >= workAreaWidth && wh >= workAreaHeight && wx <= waLeft && wy <= waTop {
		c.mutex.Unlock()
		return
	}

	x, y, w, h := GetWindowRect(hwnd)

	waLeft, waTop, waRight, waBottom = GetWorkArea()

	// 根据 dockSide 恢复到正确的可见位置
	// 左右两侧应用 VisualAdjustment 以消除视觉间隙
	// Capture dockSide before releasing lock
	dockSide := c.dockSide

	// Release lock before system calls to avoid deadlock
	c.mutex.Unlock()

	// 根据 dockSide 恢复到正确的可见位置
	// 左右两侧应用 VisualAdjustment 以消除视觉间隙
	switch dockSide {
	case DockTop:
		SetWindowPos(hwnd, x, waTop-VisualAdjustment)
		log.Printf("showWindow: DockTop, moving to (%d, %d)", x, waTop-VisualAdjustment)
	case DockBottom:
		SetWindowPos(hwnd, x, waBottom-h+8)
		log.Printf("showWindow: DockBottom, moving to (%d, %d)", x, waBottom-h+8)
	case DockLeft:
		SetWindowPos(hwnd, waLeft-VisualAdjustment, y)
		log.Printf("showWindow: DockLeft, moving to (%d, %d)", waLeft-VisualAdjustment, y)
	case DockRight:
		SetWindowPos(hwnd, waRight-w+VisualAdjustment, y)
		log.Printf("showWindow: DockRight, moving to (%d, %d)", waRight-w+VisualAdjustment, y)
	default:
		// 如果没有 dockSide，检查是否需要调整位置
		if y < waTop {
			SetWindowPos(hwnd, x, waTop)
		} else if y > waBottom-h {
			SetWindowPos(hwnd, x, waBottom-h)
		} else if x < waLeft {
			SetWindowPos(hwnd, waLeft-VisualAdjustment, y)
		} else if x > waRight-w {
			SetWindowPos(hwnd, waRight-w+VisualAdjustment, y)
		}
	}

}

func (c *InteractiveContainer) scheduleHide() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.scheduleHideLocked()
}

func (c *InteractiveContainer) scheduleHideLocked() {
	if c.hideTimer != nil {
		c.hideTimer.Stop()
	}
	c.hideTimer = time.AfterFunc(50*time.Millisecond, c.checkAndHide)
}

func (c *InteractiveContainer) checkAndHide() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	hwnd := GetWindowHandle(language.T().WindowTitle)
	if hwnd == 0 {
		return
	}

	// 检查窗口是否处于全屏状态
	if c.window.FullScreen() {
		return
	}

	// 通过Windows API进一步检查窗口是否处于全屏状态
	// 检查窗口是否最大化
	if IsWindowMaximized(hwnd) {
		return
	}

	// 获取窗口和屏幕信息
	wx, wy, ww, wh := GetWindowRect(hwnd)
	screenW, screenH := GetScreenSize()

	// 如果窗口尺寸接近全屏尺寸，则认为是全屏状态
	if ww >= screenW && wh >= screenH {
		return
	}

	// 检查窗口位置是否覆盖了整个屏幕（考虑任务栏）
	waLeft, waTop, waRight, waBottom := GetWorkArea()
	workAreaWidth := waRight - waLeft
	workAreaHeight := waBottom - waTop

	if ww >= workAreaWidth && wh >= workAreaHeight && wx <= waLeft && wy <= waTop {
		return
	}

	x, y, w, h := GetWindowRect(hwnd)
	waLeft, waTop, waRight, waBottom = GetWorkArea()

	if c.dockSide == DockNone {
		// 优先检查左右边缘，这样高窗口会优先识别为左右停靠
		if x <= waLeft {
			c.dockSide = DockLeft
		} else if x+w >= waRight {
			c.dockSide = DockRight
		} else if y <= waTop {
			c.dockSide = DockTop
		} else if y+h >= waBottom {
			c.dockSide = DockBottom
		}
	}

	// 调试日志
	// log.Printf("checkAndHide: pos=(%d,%d) size=(%dx%d) workArea=(%d,%d,%d,%d) dockSide=%d",
	// 	x, y, w, h, waLeft, waTop, waRight, waBottom, c.dockSide)

	hidden := false
	// 隐藏偏移量：足够大以确保阴影也完全不可见
	const hideOffset = 300

	switch c.dockSide {
	case DockTop:
		c.mutex.Unlock()
		SetWindowPos(hwnd, x, waTop-h-hideOffset)
		c.mutex.Lock()
		hidden = true
	case DockBottom:
		c.mutex.Unlock()
		SetWindowPos(hwnd, x, waBottom+hideOffset)
		c.mutex.Lock()
		hidden = true
	case DockLeft:
		c.mutex.Unlock()
		SetWindowPos(hwnd, waLeft-w-hideOffset, y)
		c.mutex.Lock()
		hidden = true
	case DockRight:
		c.mutex.Unlock()
		SetWindowPos(hwnd, waRight+hideOffset, y)
		c.mutex.Lock()
		hidden = true
	}

	if hidden {
		c.isHidden = true
		c.startPollingLocked()
	}
}

func (c *InteractiveContainer) startPollingLocked() {
	if c.pollStop != nil {
		return
	}
	c.pollStop = make(chan struct{})

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-c.pollStop:
				return
			case <-ticker.C:
				cx, cy := GetCursorPos()
				waLeft, waTop, waRight, waBottom := GetWorkArea()

				c.mutex.Lock()
				dockSide := c.dockSide
				c.mutex.Unlock()

				shouldShow := false
				// 触发区域：鼠标移动到边缘 20px 范围内即可唤出 (增加灵敏度)
				triggerZone := 1

				// 获取窗口位置信息，用于严格的位置检查
				hwnd := GetWindowHandle(language.T().WindowTitle)
				if hwnd == 0 {
					continue
				}
				_, wy, _, wh := GetWindowRect(hwnd)

				// Y轴容差：允许鼠标在窗口上下各50px范围内触发
				// 这样即使窗口隐藏了，鼠标在窗口原本位置附近也能触发
				yTolerance := 50

				switch dockSide {
				case DockTop:
					if cy <= waTop+triggerZone {
						shouldShow = true
					}
				case DockBottom:
					if cy >= waBottom-triggerZone {
						shouldShow = true
					}
				case DockLeft:
					// 严格检查：鼠标X坐标在左边缘 AND Y坐标在窗口范围内（含容差）
					if cx <= waLeft+triggerZone {
						// 检查鼠标Y坐标是否在窗口的垂直范围内
						if cy >= wy-yTolerance && cy <= wy+wh+yTolerance {
							shouldShow = true
						}
					}
				case DockRight:
					// 严格检查：鼠标X坐标在右边缘 AND Y坐标在窗口范围内（含容差）
					if cx >= waRight-triggerZone {
						// 检查鼠标Y坐标是否在窗口的垂直范围内
						if cy >= wy-yTolerance && cy <= wy+wh+yTolerance {
							shouldShow = true
						}
					}
				}

				if shouldShow {
					// log.Printf("Showing window from polling: dockSide=%d cursor=(%d,%d)", dockSide, cx, cy)
					c.showWindow()
					return
				}
			}
		}
	}()
}

func (c *InteractiveContainer) startPositionCheck() {
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			c.checkPosition()
		}
	}()
}

func (c *InteractiveContainer) checkPosition() {
	// 如果窗口处于全屏模式，则完全跳过所有位置检查和自动隐藏逻辑
	if c.window.FullScreen() {
		return
	}

	// 通过Windows API进一步检查窗口是否处于全屏状态
	hwnd := GetWindowHandle(language.T().WindowTitle)
	if hwnd != 0 {
		// 检查窗口是否最大化
		if IsWindowMaximized(hwnd) {
			return
		}

		// 获取窗口和屏幕信息
		wx, wy, ww, wh := GetWindowRect(hwnd)
		screenW, screenH := GetScreenSize()

		// 如果窗口尺寸接近全屏尺寸，则认为是全屏状态
		if ww >= screenW && wh >= screenH {
			return
		}

		// 检查窗口位置是否覆盖了整个屏幕（考虑任务栏）
		waLeft, waTop, waRight, waBottom := GetWorkArea()
		workAreaWidth := waRight - waLeft
		workAreaHeight := waBottom - waTop

		if ww >= workAreaWidth && wh >= workAreaHeight && wx <= waLeft && wy <= waTop {
			return
		}
	}

	// 1. Fast check for hidden state
	c.mutex.Lock()
	if c.isHidden {
		c.mutex.Unlock()
		return
	}
	c.mutex.Unlock()

	// 2. System calls (slow, do not hold lock)
	if hwnd == 0 {
		hwnd = GetWindowHandle(language.T().WindowTitle)
		if hwnd == 0 {
			return
		}
	}

	x, y, w, h := GetWindowRect(hwnd)
	waLeft, waTop, waRight, waBottom := GetWorkArea()

	// 3. Update state (hold lock)
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Re-check hidden state in case it changed
	if c.isHidden {
		return
	}

	posChanged := (x != c.lastX || y != c.lastY)
	sizeChanged := (w != c.lastW || h != c.lastH)

	// 检查窗口是否最大化
	isMaximized := IsWindowMaximized(hwnd)

	if !posChanged && !sizeChanged {
		// 如果窗口最大化，跳过所有位置调整
		if isMaximized {
			return
		}

		if c.dockSide != DockNone {
			shouldSnap := false
			targetX, targetY := x, y

			switch c.dockSide {
			case DockLeft:
				if x != waLeft-VisualAdjustment {
					targetX = waLeft - VisualAdjustment
					shouldSnap = true
				}
				// Corner Snap: Bottom-Left - 当接近底部时强制吸附
				if y+h >= waBottom-SnapDistance {
					if y != waBottom-h+8 {
						targetY = waBottom - h + 8
						shouldSnap = true
					}
				} else if y+h > waBottom+8 {
					targetY = waBottom - h + 8
					shouldSnap = true
				}
			case DockRight:
				if x != waRight-w+VisualAdjustment && x < waRight {
					targetX = waRight - w + VisualAdjustment
					shouldSnap = true
				}
				// Corner Snap: Bottom-Right - 当接近底部时强制吸附
				if y+h >= waBottom-SnapDistance {
					if y != waBottom-h+8 {
						targetY = waBottom - h + 8
						shouldSnap = true
					}
				} else if y+h > waBottom+8 {
					targetY = waBottom - h + 8
					shouldSnap = true
				}
			case DockTop:
				if y != waTop-VisualAdjustment {
					targetY = waTop - VisualAdjustment
					shouldSnap = true
				}
			case DockBottom:
				if y != waBottom-h+8 {
					targetY = waBottom - h + 8
					shouldSnap = true
				}
			}

			if shouldSnap {
				if c.debugMode {
					log.Println("[DEBUG] checkPosition: snapping")
				}
				c.mutex.Unlock() // Release lock before system call
				SetWindowPos(hwnd, targetX, targetY)
				c.mutex.Lock() // Re-acquire lock

				c.lastX = targetX
				c.lastY = targetY

				// 吸附后立即启动隐藏检查
				go func() {
					time.Sleep(200 * time.Millisecond) // 增加延迟，避免误触
					cx, cy := GetCursorPos()
					// 再次获取窗口位置，确保准确
					if wx, wy, ww, wh := GetWindowRect(hwnd); wx != 0 {
						if cx < wx || cx > wx+ww || cy < wy || cy > wy+wh {
							c.scheduleHideLocked()
						}
					}
				}()
			} else {
				// Watchdog: 监控鼠标是否离开窗口
				cx, cy := GetCursorPos()
				if cx < x || cx > x+w || cy < y || cy > y+h {
					if c.hideTimer == nil {
						// 延迟隐藏，防止抖动
						c.scheduleHideLocked()
					}
				}
			}
		} else {
			// 即使没有吸附，也要限制窗口不能超出工作区
			needsAdjustment := false
			targetX, targetY := x, y

			// 检查顶部边界（防止标题栏超出屏幕）
			if y < waTop {
				targetY = waTop
				needsAdjustment = true
				if c.debugMode {
					log.Printf("[DEBUG] checkPosition: restricting top boundary. y=%d, waTop=%d", y, waTop)
				}
			}

			// 检查底部边界（防止窗口超出任务栏上方）
			if y+h > waBottom {
				targetY = waBottom - h
				needsAdjustment = true
				if c.debugMode {
					log.Printf("[DEBUG] checkPosition: restricting bottom boundary. y=%d, h=%d, waBottom=%d", y, h, waBottom)
				}
			}

			// 检查左边界
			if x < waLeft {
				targetX = waLeft
				needsAdjustment = true
			}

			// 检查右边界
			if x+w > waRight {
				targetX = waRight - w
				needsAdjustment = true
			}

			if needsAdjustment {
				c.mutex.Unlock()
				SetWindowPos(hwnd, targetX, targetY)
				c.mutex.Lock()
				c.lastX = targetX
				c.lastY = targetY
			}
		}
		return
	}

	c.lastX = x
	c.lastY = y
	c.lastW = w
	c.lastH = h

	if (posChanged || sizeChanged) && c.saveStateFunc != nil {
		// Debounce save: wait for 1 second of inactivity before saving
		if c.saveTimer != nil {
			c.saveTimer.Stop()
		}
		c.saveTimer = time.AfterFunc(1*time.Second, func() {
			c.saveStateFunc()
		})
	}

	oldDockSide := c.dockSide
	c.dockSide = DockNone

	// 如果窗口最大化，跳过吸附检测（已在前面检测过）
	if isMaximized {
		// 窗口最大化时，清除 dockSide 并确保不会隐藏
		if oldDockSide != DockNone {
			c.isHidden = false
			if c.hideTimer != nil {
				c.hideTimer.Stop()
				c.hideTimer = nil
			}
			if c.pollStop != nil {
				close(c.pollStop)
				c.pollStop = nil
			}
		}
		return
	}

	// 增加吸附距离，使其更容易触发
	snapDist := 100

	// 优先级：左右 > 上下
	if x < waLeft+snapDist {
		c.dockSide = DockLeft
	} else if x+w > waRight-snapDist {
		c.dockSide = DockRight
	} else if y < waTop+snapDist {
		c.dockSide = DockTop
	} else if y+h > waBottom-snapDist {
		c.dockSide = DockBottom
	}

	if c.dockSide != oldDockSide && c.dockSide != DockNone {
		if c.hideTimer != nil {
			c.hideTimer.Stop()
			c.hideTimer = nil
		}
	} else if c.dockSide == DockNone && oldDockSide != DockNone {
		c.isHidden = false
		if c.hideTimer != nil {
			c.hideTimer.Stop()
			c.hideTimer = nil
		}
		if c.pollStop != nil {
			close(c.pollStop)
			c.pollStop = nil
		}
	}
}
