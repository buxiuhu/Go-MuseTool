package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ShortcutWidget is a custom widget that displays an icon above text (Windows 11 style)
type ShortcutWidget struct {
	widget.BaseWidget
	icon         *canvas.Image
	label        *widget.Label
	OnTapped     func()
	OnRightClick func(*fyne.PointEvent)
	onDragEnd    func(startPos, endPos fyne.Position)
	dragStartPos fyne.Position
	dragEndPos   fyne.Position
	isDragging   bool
}

// NewShortcutWidget creates a new vertical shortcut widget (Windows 11 style)
func NewShortcutWidget(labelText string, onLeft func(), onRight func(*fyne.PointEvent), onDragEnd func(startPos, endPos fyne.Position)) *ShortcutWidget {
	s := &ShortcutWidget{
		OnTapped:     onLeft,
		OnRightClick: onRight,
		onDragEnd:    onDragEnd,
	}

	s.icon = canvas.NewImageFromResource(theme.ComputerIcon())
	s.icon.FillMode = canvas.ImageFillContain
	s.icon.SetMinSize(fyne.NewSize(35, 35))

	s.label = widget.NewLabel(labelText)
	s.label.Alignment = fyne.TextAlignCenter
	s.label.Wrapping = fyne.TextWrapBreak

	s.ExtendBaseWidget(s)
	return s
}

// SetIcon sets the icon for the shortcut
func (s *ShortcutWidget) SetIcon(res fyne.Resource) {
	s.icon.Resource = res
	s.icon.Refresh()
}

// CreateRenderer creates the renderer for this widget
func (s *ShortcutWidget) CreateRenderer() fyne.WidgetRenderer {
	content := container.NewVBox(
		container.NewCenter(s.icon),
		s.label,
	)
	return widget.NewSimpleRenderer(container.NewStack(content))
}

// Tapped handles left click
func (s *ShortcutWidget) Tapped(_ *fyne.PointEvent) {
	if s.OnTapped != nil {
		s.OnTapped()
	}
}

// TappedSecondary handles right click
func (s *ShortcutWidget) TappedSecondary(e *fyne.PointEvent) {
	if s.OnRightClick != nil {
		s.OnRightClick(e)
	}
}

// Dragged 实现 fyne.Draggable 接口以支持拖拽
func (s *ShortcutWidget) Dragged(e *fyne.DragEvent) {
	if !s.isDragging {
		s.isDragging = true
		s.dragStartPos = e.Position
	}
	// 持续更新拖拽结束位置
	s.dragEndPos = e.Position
}

// DragEnd 实现 fyne.Draggable 接口，在拖拽结束时调用
func (s *ShortcutWidget) DragEnd() {
	if s.isDragging && s.onDragEnd != nil {
		s.onDragEnd(s.dragStartPos, s.dragEndPos)
	}
	s.isDragging = false
}
