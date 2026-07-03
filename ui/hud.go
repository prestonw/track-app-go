package ui

import (
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"github.com/prestonw/track-app-go/internal/app"
	"github.com/prestonw/track-app-go/internal/format"
	"github.com/prestonw/track-app-go/internal/platform"
)

type HUD struct {
	app      *app.TrackApp
	window   fyne.Window
	body     *fyne.Container
	banner   *fyne.Container
	clock    *hudClock
	jobBtn   *widget.Button
	play     *widget.Button
	visible  bool
	size     fyne.Size
}

func NewHUD(a *app.TrackApp, fyneApp fyne.App) *HUD {
	h := &HUD{app: a, size: fyne.NewSize(236, 80)}
	if drv, ok := fyneApp.Driver().(desktop.Driver); ok {
		h.window = drv.CreateSplashWindow()
		h.window.SetPadded(false)
	} else {
		h.window = fyneApp.NewWindow("")
		h.window.SetPadded(false)
	}
	h.window.SetFixedSize(true)
	h.window.Resize(h.size)
	h.window.SetTitle(platform.HUDWindowTitle)
	h.window.SetIcon(AppIcon())
	h.window.SetCloseIntercept(func() { h.Hide() })

	toggleTimer := func() {
		if a.Coordinator.FocusTimer() == nil {
			h.showQuickJobDialog()
			return
		}
		a.Coordinator.ToggleFocus()
	}

	h.clock = newHUDClock(toggleTimer)

	h.jobBtn = widget.NewButton("Select job ▾", func() { h.showJobMenu() })
	h.jobBtn.Importance = widget.LowImportance

	h.play = widget.NewButton("▶", toggleTimer)
	h.play.Importance = widget.HighImportance
	playBox := container.NewGridWrap(fyne.NewSize(hudTransportSize, hudTransportSize), h.play)

	clockRow := container.NewHBox(
		newTapPad(h.cycleCorner),
		container.NewCenter(h.clock),
		playBox,
		newTapPad(h.cycleCorner),
	)

	h.banner = container.NewHBox()

	top := h.jobBtn
	mid := newTapPad(h.cycleCorner)
	inner := container.NewBorder(nil, clockRow, nil, nil, container.NewVBox(h.banner, top, mid))
	bg := canvas.NewRectangle(colorSurface)
	bg.CornerRadius = cardRadius
	bg.StrokeColor = colorAccent
	bg.StrokeWidth = 1
	h.body = container.NewStack(bg, container.NewPadded(inner))
	h.window.SetContent(h.body)

	a.OnChange(func() { onMain(h.refresh) })
	return h
}

func (h *HUD) Show() {
	h.visible = true
	h.window.Show()
	h.refresh()
	h.placeHUDSoon()
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
	h.placeHUDSoon()
}

func (h *HUD) placeHUDSoon() {
	corner := h.app.Coordinator.HUDCorner()
	w, ht := int(h.size.Width), int(h.size.Height)
	title := platform.HUDWindowTitle
	go func() {
		delays := []time.Duration{0, 80 * time.Millisecond, 200 * time.Millisecond, 500 * time.Millisecond}
		for _, d := range delays {
			if d > 0 {
				time.Sleep(d)
			}
			if !h.visible {
				return
			}
			h.app.Platform.Window().PlaceByTitle(title, platform.CornerFromInt(corner), w, ht)
		}
	}()
}

func (h *HUD) refresh() {
	title, _, elapsed, running := h.app.DisplayContext()
	if title == "" {
		title = "Select job"
	}
	h.jobBtn.SetText(title + " ▾")
	h.clock.SetTime(format.Duration(elapsed), running)
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
		h.window.Resize(h.size)
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
	start := widget.NewButton("Start", func() { h.app.Coordinator.ConfirmAutoStart(); h.app.Notify(); h.refresh() })
	skip := widget.NewButton("Skip", func() { h.app.Coordinator.SkipAutoStart(); h.app.Notify(); h.refresh() })
	h.banner.Add(label)
	h.banner.Add(start)
	h.banner.Add(skip)
	h.banner.Refresh()
	h.window.Resize(fyne.NewSize(h.size.Width, h.size.Height+36))
	h.placeHUDSoon()
}

func (h *HUD) showJobMenu() {
	var menuItems []*fyne.MenuItem

	menuItems = append(menuItems, fyne.NewMenuItem("+ New job…", func() { h.showQuickJobDialog() }))
	menuItems = append(menuItems, fyne.NewMenuItem("+ Link to project…", func() { h.showQuickProjectDialog() }))
	menuItems = append(menuItems, fyne.NewMenuItemSeparator())

	seen := map[string]bool{}
	for _, id := range h.app.Coordinator.RecentTimerIDs(8) {
		if seen[id] {
			continue
		}
		seen[id] = true
		if item := h.timerMenuItem(id); item != nil {
			menuItems = append(menuItems, item)
		}
	}
	if len(menuItems) <= 2 {
		for _, t := range h.app.Store.Timers {
			if seen[t.ID] {
				continue
			}
			if item := h.timerMenuItem(t.ID); item != nil {
				menuItems = append(menuItems, item)
			}
		}
	}
	if len(menuItems) <= 2 {
		return
	}
	pop := widget.NewPopUpMenu(fyne.NewMenu("Jobs", menuItems...), h.window.Canvas())
	pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(h.jobBtn)
	pop.ShowAtPosition(pos.Add(fyne.NewPos(0, h.jobBtn.Size().Height)))
}

func (h *HUD) timerMenuItem(id string) *fyne.MenuItem {
	id := id
	for _, t := range h.app.Store.Timers {
		if t.ID != id {
			continue
		}
		label := t.Name
		if t.Running {
			label = "▶ " + label
		}
		return fyne.NewMenuItem(label, func() {
			h.app.Coordinator.SetFocus(id, false)
		})
	}
	return nil
}

func (h *HUD) showQuickJobDialog() {
	name := widget.NewEntry()
	name.SetPlaceHolder("e.g. Client website")
	hint := mutedLabel("Name your job — tracking starts right away.")

	var dlg dialog.Dialog
	add := primaryButton("Add & start", func() {
		n := strings.TrimSpace(name.Text)
		if n == "" {
			dialog.ShowInformation("New job", "Enter a job name", h.window)
			return
		}
		t := h.app.Store.AddTimer(n, nil, 0, format.DisplayCurrency, "", "", 0)
		h.app.Coordinator.SetFocus(t.ID, true)
		h.app.Notify()
		h.refresh()
		if dlg != nil {
			dlg.Hide()
		}
	})

	body := container.NewVBox(hint, name, add)
	dlg = dialog.NewCustom("New job", "Cancel", fluidCard(body, cardAccent), h.window)
	dlg.Show()
}

func (h *HUD) showQuickProjectDialog() {
	name := widget.NewEntry()
	name.SetPlaceHolder("Project name")
	auto := widget.NewCheck("Auto-track when rules match", nil)

	timerOpts := []string{"— No job —"}
	timerIDs := []string{""}
	for _, t := range h.app.Store.Timers {
		timerOpts = append(timerOpts, t.Name)
		timerIDs = append(timerIDs, t.ID)
	}
	timerSel := widget.NewSelect(timerOpts, nil)
	if focus := h.app.Coordinator.FocusTimer(); focus != nil {
		for i, id := range timerIDs {
			if id == focus.ID {
				timerSel.SetSelected(timerOpts[i])
				break
			}
		}
	} else if len(timerOpts) > 0 {
		timerSel.SetSelected(timerOpts[0])
	}

	clientOpts := []string{"— No client —"}
	clientIDs := []string{""}
	for _, c := range h.app.Store.Clients {
		clientOpts = append(clientOpts, c.Name)
		clientIDs = append(clientIDs, c.ID)
	}
	clientSel := widget.NewSelect(clientOpts, nil)
	if len(clientOpts) > 0 {
		clientSel.SetSelected(clientOpts[0])
	}

	hint := mutedLabel("Create a project and link it to a job timer. Add match rules in the main window.")

	var dlg dialog.Dialog
	add := primaryButton("Add project", func() {
		n := strings.TrimSpace(name.Text)
		if n == "" {
			dialog.ShowInformation("New project", "Enter a project name", h.window)
			return
		}
		cid, tid := "", ""
		if i := indexOf(clientOpts, clientSel.Selected); i >= 0 {
			cid = clientIDs[i]
		}
		if i := indexOf(timerOpts, timerSel.Selected); i >= 0 {
			tid = timerIDs[i]
		}
		h.app.Store.AddProject(n, cid, tid, auto.Checked, "")
		h.app.Notify()
		if dlg != nil {
			dlg.Hide()
		}
	})

	form := container.NewVBox(
		hint,
		widget.NewForm(
			widget.NewFormItem("Name", name),
			widget.NewFormItem("Job timer", timerSel),
			widget.NewFormItem("Client", clientSel),
		),
		auto,
		add,
	)
	dlg = dialog.NewCustom("Link to project", "Cancel", fluidCard(form, cardDefault), h.window)
	dlg.Show()
}