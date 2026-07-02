package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/prestonw/track-app-go/internal/app"
	hudpos "github.com/prestonw/track-app-go/internal/hud"
	"github.com/prestonw/track-app-go/internal/format"
)

type HUD struct {
	app      *app.TrackApp
	window   fyne.Window
	body     *fyne.Container
	banner   *fyne.Container
	clock    *widget.Label
	jobBtn   *widget.Button
	play     *widget.Button
	visible  bool
	size     fyne.Size
}

func NewHUD(a *app.TrackApp, fyneApp fyne.App) *HUD {
	h := &HUD{app: a, size: fyne.NewSize(260, 100)}
	h.window = fyneApp.NewWindow("")
	h.window.SetFixedSize(true)
	h.window.Resize(h.size)
	h.window.SetTitle("")
	h.window.SetPadded(true)
	h.window.SetCloseIntercept(func() { h.Hide() })

	h.clock = widget.NewLabel("00:00:00")
	h.clock.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}

	h.jobBtn = widget.NewButton("Select job ▾", func() { h.showJobMenu() })
	h.jobBtn.Importance = widget.LowImportance

	h.play = widget.NewButton("▶", nil)
	h.play.OnTapped = func() { a.Coordinator.ToggleFocus() }

	reset := widget.NewButton("×", func() { a.Coordinator.ResetFocus() })
	reset.Importance = widget.DangerImportance

	center := container.NewHBox(h.clock, h.play)

	cornerBtn := widget.NewButton("◢", func() { h.cycleCorner() })
	cornerBtn.Importance = widget.LowImportance

	h.banner = container.NewHBox()

	top := container.NewBorder(nil, nil, cornerBtn, reset, h.jobBtn)
	h.body = container.NewVBox(h.banner, container.NewBorder(top, nil, nil, nil, center))
	h.window.SetContent(h.body)

	a.OnChange(func() { h.refresh() })
	h.refresh()
	return h
}

func (h *HUD) Show() {
	h.visible = true
	h.window.Show()
	hudpos.MoveActiveWindow(h.app.Coordinator.HUDCorner(), int(h.size.Width), int(h.size.Height))
	h.refresh()
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

func (h *HUD) cycleCorner() {
	h.app.Coordinator.CycleHUDCorner()
	hudpos.MoveActiveWindow(h.app.Coordinator.HUDCorner(), int(h.size.Width), int(h.size.Height))
}

func (h *HUD) refresh() {
	title, _, elapsed, running := h.app.DisplayContext()
	h.jobBtn.SetText(title + " ▾")
	h.clock.SetText(format.Duration(elapsed))
	if running {
		h.play.SetText("⏸")
	} else {
		h.play.SetText("▶")
	}
	h.refreshBanner()
}

func (h *HUD) refreshBanner() {
	h.banner.Objects = nil
	prompt := h.app.Coordinator.AutoStartPrompt()
	if prompt == nil {
		h.banner.Hide()
		h.body.Refresh()
		return
	}
	h.banner.Show()
	proj := h.app.Store.Project(prompt.ProjectID)
	name := "this project"
	if proj != nil {
		name = proj.Name
	}
	label := widget.NewLabel("Start tracking " + name + "?")
	start := widget.NewButton("Start", func() { h.app.Coordinator.ConfirmAutoStart(); h.refresh() })
	skip := widget.NewButton("Skip", func() { h.app.Coordinator.SkipAutoStart(); h.refresh() })
	h.banner.Add(label)
	h.banner.Add(start)
	h.banner.Add(skip)
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
		for _, t := range h.app.Store.Timers {
			t := t
			menuItems = append(menuItems, fyne.NewMenuItem(t.Name, func() {
				h.app.Coordinator.SetFocus(t.ID, false)
			}))
		}
	}
	if len(menuItems) == 0 {
		return
	}
	pop := widget.NewPopUpMenu(fyne.NewMenu("Recent jobs", menuItems...), h.window.Canvas())
	pop.ShowAtPosition(fyne.NewPos(12, 40))
}