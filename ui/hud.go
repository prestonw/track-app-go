package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/prestonw/track-app-go/internal/app"
	"github.com/prestonw/track-app-go/internal/format"
)

type HUD struct {
	app     *app.TrackApp
	window  fyne.Window
	clock   *widget.Label
	jobBtn  *widget.Button
	play    *widget.Button
	visible bool
}

func NewHUD(a *app.TrackApp, fyneApp fyne.App) *HUD {
	h := &HUD{app: a}
	h.window = fyneApp.NewWindow("")
	h.window.SetFixedSize(true)
	h.window.Resize(fyne.NewSize(240, 88))
	h.window.SetTitle("")
	h.window.SetPadded(true)

	h.clock = widget.NewLabel("00:00:00")
	h.clock.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}

	h.jobBtn = widget.NewButton("Select job ▾", func() { h.showJobMenu() })
	h.jobBtn.Importance = widget.LowImportance

	h.play = widget.NewButton("▶", nil)
	reset := widget.NewButton("×", func() { a.Coordinator.ResetFocus() })
	reset.Importance = widget.DangerImportance
	h.play.OnTapped = func() { a.Coordinator.ToggleFocus() }

	top := container.NewBorder(nil, nil, nil, reset, h.jobBtn)
	center := container.NewHBox(h.clock, h.play)
	h.window.SetContent(container.NewBorder(top, nil, nil, nil, center))

	a.OnChange(func() { h.refresh() })
	h.refresh()
	return h
}

func (h *HUD) Show() {
	h.visible = true
	h.window.Show()
}
func (h *HUD) Hide() {
	h.visible = false
	h.window.Hide()
}
func (h *HUD) Visible() bool { return h.visible }
func (h *HUD) Toggle() {
	if h.Visible() {
		h.Hide()
	} else {
		h.Show()
	}
}

func (h *HUD) refresh() {
	title, _, elapsed, running := h.app.Coordinator.DisplayContext()
	h.jobBtn.SetText(title + " ▾")
	h.clock.SetText(format.Duration(elapsed))
	if running {
		h.play.SetText("⏸")
	} else {
		h.play.SetText("▶")
	}
}

func (h *HUD) showJobMenu() {
	var menuItems []*fyne.MenuItem
	for _, id := range h.app.Coordinator.RecentTimerIDs(8) {
		id := id
		for _, t := range h.app.Store.Timers {
			if t.ID != id {
				continue
			}
			label := t.Name
			if t.Running {
				label = "▶ " + label
			}
			menuItems = append(menuItems, fyne.NewMenuItem(label, func() {
				h.app.Coordinator.SetFocus(id, false)
			}))
			break
		}
	}
	if len(menuItems) == 0 {
		return
	}
	pop := widget.NewPopUpMenu(fyne.NewMenu("Recent jobs", menuItems...), h.window.Canvas())
	pop.ShowAtPosition(fyne.NewPos(20, 20))
}