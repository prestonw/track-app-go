package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

const (
	hudClockSize     float32 = 26
	hudTransportSize float32 = 36
)

// tapPad is an invisible expander that cycles HUD corners when clicked.
type tapPad struct {
	widget.BaseWidget
	onTap func()
}

func newTapPad(onTap func()) *tapPad {
	t := &tapPad{onTap: onTap}
	t.ExtendBaseWidget(t)
	return t
}

func (t *tapPad) Tapped(*fyne.PointEvent) {
	if t.onTap != nil {
		t.onTap()
	}
}

func (t *tapPad) MinSize() fyne.Size {
	return fyne.NewSize(20, 12)
}

func (t *tapPad) CreateRenderer() fyne.WidgetRenderer {
	r := canvas.NewRectangle(color.Transparent)
	r.SetMinSize(fyne.NewSize(20, 12))
	return widget.NewSimpleRenderer(r)
}

// hudClock shows a large monospace timer; tap toggles start/stop.
type hudClock struct {
	widget.BaseWidget
	text    *canvas.Text
	running bool
	onTap   func()
}

func newHUDClock(onTap func()) *hudClock {
	c := &hudClock{
		onTap: onTap,
		text: canvas.NewText("00:00:00", colorText),
	}
	c.text.TextSize = hudClockSize
	c.text.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	c.ExtendBaseWidget(c)
	return c
}

func (c *hudClock) Tapped(*fyne.PointEvent) {
	if c.onTap != nil {
		c.onTap()
	}
}

func (c *hudClock) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

func (c *hudClock) SetTime(text string, running bool) {
	c.running = running
	c.text.Text = text
	if running {
		c.text.Color = colorRunning
	} else {
		c.text.Color = colorText
	}
	c.text.Refresh()
}

func (c *hudClock) CreateRenderer() fyne.WidgetRenderer {
	c.text.Alignment = fyne.TextAlignCenter
	return widget.NewSimpleRenderer(c.text)
}

func (c *hudClock) MinSize() fyne.Size {
	c.text.TextSize = hudClockSize
	if c.text.Text == "" {
		c.text.Text = "00:00:00"
	}
	return c.text.MinSize()
}