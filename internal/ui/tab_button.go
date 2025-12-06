package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// TabButton is a custom button for tabs with non-bold text
type TabButton struct {
	widget.BaseWidget
	Text       string
	Icon       fyne.Resource
	Importance widget.Importance
	OnTapped   func()

	onRightClick func(*fyne.PointEvent)
	onDragEnd    func(startPos, endPos fyne.Position)
	dragStartPos fyne.Position
	dragEndPos   fyne.Position
	isDragging   bool
	hovered      bool
}

// NewTabButton 创建一个新的 TabButton
func NewTabButton(text string, icon fyne.Resource, tapped func(), onRightClick func(*fyne.PointEvent), onDragEnd func(startPos, endPos fyne.Position)) *TabButton {
	b := &TabButton{
		Text:         text,
		Icon:         icon,
		OnTapped:     tapped,
		onRightClick: onRightClick,
		onDragEnd:    onDragEnd,
		Importance:   widget.LowImportance,
	}
	b.ExtendBaseWidget(b)
	return b
}

// CreateRenderer creates the renderer for this widget
func (b *TabButton) CreateRenderer() fyne.WidgetRenderer {
	text := canvas.NewText(b.Text, theme.ForegroundColor())
	text.Alignment = fyne.TextAlignCenter
	// Ensure not bold
	text.TextStyle = fyne.TextStyle{Bold: false}

	var icon *widget.Icon
	if b.Icon != nil {
		icon = widget.NewIcon(b.Icon)
	}

	bg := canvas.NewRectangle(color.Transparent)

	// Create content container
	var content fyne.CanvasObject
	if icon != nil {
		content = container.NewHBox(icon, text)
	} else {
		content = text
	}

	// Use Padded container to give some space like a button
	padded := container.NewPadded(content)

	return &tabButtonRenderer{
		b:      b,
		bg:     bg,
		text:   text,
		icon:   icon,
		layout: padded,
	}
}

// Tapped handles left click
func (b *TabButton) Tapped(_ *fyne.PointEvent) {
	if b.OnTapped != nil {
		b.OnTapped()
	}
}

// TappedSecondary handles right click
func (b *TabButton) TappedSecondary(e *fyne.PointEvent) {
	if b.onRightClick != nil {
		b.onRightClick(e)
	}
}

// MouseIn handles mouse hover enter
func (b *TabButton) MouseIn(_ *desktop.MouseEvent) {
	b.hovered = true
	b.Refresh()
}

// MouseOut handles mouse hover leave
func (b *TabButton) MouseOut() {
	b.hovered = false
	b.Refresh()
}

// MouseMoved handles mouse movement (required for Hoverable)
func (b *TabButton) MouseMoved(_ *desktop.MouseEvent) {
}

// Dragged implements fyne.Draggable
func (b *TabButton) Dragged(e *fyne.DragEvent) {
	if !b.isDragging {
		b.isDragging = true
		b.dragStartPos = e.Position
	}
	b.dragEndPos = e.Position
}

// DragEnd implements fyne.Draggable
func (b *TabButton) DragEnd() {
	if b.isDragging && b.onDragEnd != nil {
		b.onDragEnd(b.dragStartPos, b.dragEndPos)
	}
	b.isDragging = false
}

// Renderer implementation
type tabButtonRenderer struct {
	b      *TabButton
	bg     *canvas.Rectangle
	text   *canvas.Text
	icon   *widget.Icon
	layout fyne.CanvasObject
}

func (r *tabButtonRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	r.layout.Resize(size)
}

func (r *tabButtonRenderer) MinSize() fyne.Size {
	return r.layout.MinSize()
}

func (r *tabButtonRenderer) Refresh() {
	r.text.Text = r.b.Text
	r.text.TextSize = theme.TextSize()
	r.text.TextStyle = fyne.TextStyle{Bold: false} // Enforce non-bold

	// Update background and text color based on state
	if r.b.Importance == widget.HighImportance {
		// Selected state
		r.bg.FillColor = theme.PrimaryColor()
		r.text.Color = theme.BackgroundColor() // Usually provides good contrast on Primary
	} else {
		// Unselected state
		r.text.Color = theme.ForegroundColor()
		if r.b.hovered {
			r.bg.FillColor = theme.HoverColor()
		} else {
			r.bg.FillColor = color.Transparent
		}
	}

	canvas.Refresh(r.bg)
	canvas.Refresh(r.text)

	if r.icon != nil {
		r.icon.Refresh()
	}
}

func (r *tabButtonRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.layout}
}

func (r *tabButtonRenderer) Destroy() {
}
