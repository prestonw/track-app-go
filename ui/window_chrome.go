package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/prestonw/track-app-go/internal/platform"
)

const titleBarH float32 = 28

// windowDragHandle fills the title bar and moves the native window when dragged.
type windowDragHandle struct {
	widget.BaseWidget
}

func newWindowDragHandle() *windowDragHandle {
	d := &windowDragHandle{}
	d.ExtendBaseWidget(d)
	return d
}

func (d *windowDragHandle) Dragged(ev *fyne.DragEvent) {
	platform.MainWindowMoveBy(ev.Dragged.X, ev.Dragged.Y)
}

func (d *windowDragHandle) DragEnd() {}

func (d *windowDragHandle) MinSize() fyne.Size {
	return fyne.NewSize(120, titleBarH)
}

func (d *windowDragHandle) CreateRenderer() fyne.WidgetRenderer {
	r := canvas.NewRectangle(color.Transparent)
	r.SetMinSize(fyne.NewSize(120, titleBarH))
	return widget.NewSimpleRenderer(r)
}

func windowChrome(content fyne.CanvasObject, onClose, onMin func()) fyne.CanvasObject {
	close := widget.NewButtonWithIcon("", theme.WindowCloseIcon(), onClose)
	close.Importance = widget.MediumImportance
	min := widget.NewButtonWithIcon("", theme.WindowMinimizeIcon(), onMin)
	min.Importance = widget.MediumImportance

	drag := newWindowDragHandle()
	controls := container.NewHBox(close, min)
	barInner := container.NewBorder(nil, nil, controls, nil, drag)

	barBg := canvas.NewRectangle(colorSurfaceAlt)
	barBg.CornerRadius = 0
	bar := container.NewStack(barBg, container.NewPadded(barInner))

	shellBg := canvas.NewRectangle(colorBG)
	return container.NewStack(shellBg, container.NewBorder(bar, nil, nil, nil, content))
}