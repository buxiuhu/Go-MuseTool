package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// TabButton 是一个支持右键菜单和拖拽的按钮
type TabButton struct {
	widget.Button
	onRightClick func(*fyne.PointEvent)
	onDragEnd    func(startPos, endPos fyne.Position)
	dragStartPos fyne.Position
	dragEndPos   fyne.Position
	isDragging   bool
}

// NewTabButton 创建一个新的 TabButton
func NewTabButton(text string, icon fyne.Resource, tapped func(), onRightClick func(*fyne.PointEvent), onDragEnd func(startPos, endPos fyne.Position)) *TabButton {
	btn := &TabButton{
		onRightClick: onRightClick,
		onDragEnd:    onDragEnd,
	}
	btn.Text = text
	btn.Icon = icon
	btn.OnTapped = tapped
	btn.ExtendBaseWidget(btn)
	return btn
}

// TappedSecondary 实现 desktop.Tappable 接口以支持右键点击
func (b *TabButton) TappedSecondary(e *fyne.PointEvent) {
	if b.onRightClick != nil {
		b.onRightClick(e)
	}
}

// Dragged 实现 fyne.Draggable 接口以支持拖拽
func (b *TabButton) Dragged(e *fyne.DragEvent) {
	if !b.isDragging {
		b.isDragging = true
		b.dragStartPos = e.Position
	}
	// 持续更新拖拽结束位置
	b.dragEndPos = e.Position
}

// DragEnd 实现 fyne.Draggable 接口，在拖拽结束时调用
func (b *TabButton) DragEnd() {
	if b.isDragging && b.onDragEnd != nil {
		b.onDragEnd(b.dragStartPos, b.dragEndPos)
	}
	b.isDragging = false
}
