package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// TappableBackground 是一个可以响应右键点击的背景矩形
type TappableBackground struct {
	widget.BaseWidget
	rect         *canvas.Rectangle
	onRightClick func(*fyne.PointEvent)
}

// NewTappableBackground 创建一个新的可点击背景
func NewTappableBackground(c color.Color, onRightClick func(*fyne.PointEvent)) *TappableBackground {
	bg := &TappableBackground{
		rect:         canvas.NewRectangle(c),
		onRightClick: onRightClick,
	}
	bg.ExtendBaseWidget(bg)
	return bg
}

// CreateRenderer 实现 fyne.Widget 接口
func (t *TappableBackground) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.rect)
}

// TappedSecondary 实现 desktop.Tappable 接口以支持右键点击
func (t *TappableBackground) TappedSecondary(e *fyne.PointEvent) {
	if t.onRightClick != nil {
		t.onRightClick(e)
	}
}

// SetColor 设置背景颜色
func (t *TappableBackground) SetColor(c color.Color) {
	t.rect.FillColor = c
	t.rect.Refresh()
}
