package ui

import (
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/prestonw/track-app-go/internal/app"
	"github.com/prestonw/track-app-go/internal/format"
	"github.com/prestonw/track-app-go/internal/platform"
)

type HUD struct {
	app          *app.TrackApp
	window       fyne.Window
	dialogParent fyne.Window
	body         *fyne.Container
	banner       *fyne.Container
	hintRow      *fyne.Container
	clock        *hudClock
	jobBtn       *widget.Button
	play         *widget.Button
	visible      bool
	size         fyne.Size
	compactSize  fyne.Size
	hintSize     fyne.Size
	placementGen int
}

func NewHUD(a *app.TrackApp, fyneApp fyne.App) *HUD {
	h := &HUD{
		app:         a,
		compactSize: fyne.NewSize(204, 54),
		hintSize:    fyne.NewSize(204, 68),
		size:        fyne.NewSize(204, 68),
	}
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
	h.play = newHUDTransport(toggleTimer)

	clockRow := container.NewHBox(
		newTapPad(h.cycleCorner),
		layout.NewSpacer(),
		h.clock,
		container.NewGridWrap(fyne.NewSize(hudTransportSize, hudTransportSize), h.play),
		newTapPad(h.cycleCorner),
	)

	h.banner = container.NewHBox()
	cornerBtn := widget.NewButton("◢", h.cycleCorner)
	cornerBtn.Importance = widget.LowImportance
	hint := mutedLabel("Tap empty space to move")
	h.hintRow = container.NewBorder(nil, nil, nil, cornerBtn, hint)

	top := container.NewVBox(h.jobBtn, h.hintRow)
	inner := container.NewBorder(nil, clockRow, nil, nil, container.NewVBox(h.banner, top))

	bg := canvas.NewRectangle(colorHUDSurface)
	bg.CornerRadius = 10
	bg.StrokeColor = colorHUDBorder
	bg.StrokeWidth = 1
	padded := container.New(layout.NewCustomPaddedLayout(hudPad, hudPad, hudPad, hudPad), inner)
	h.body = container.NewStack(bg, padded)
	h.window.SetContent(h.body)

	h.applyHintVisibility()
	a.OnChange(func() { onMain(h.refresh) })
	return h
}

func (h *HUD) applyHintVisibility() {
	if h.app.Coordinator.HUDHintDismissed() {
		h.hintRow.Hide()
		h.size = h.compactSize
	} else {
		h.hintRow.Show()
		h.size = h.hintSize
	}
	h.window.Resize(h.size)
	h.body.Refresh()
}

func (h *HUD) dismissHint() {
	if h.app.Coordinator.HUDHintDismissed() {
		return
	}
	h.app.Coordinator.SetHUDHintDismissed(true)
	h.applyHintVisibility()
	if h.visible {
		h.placeHUD(true)
	}
}

// Window returns the HUD Fyne window.
func (h *HUD) Window() fyne.Window { return h.window }

// SetDialogParent sets the window used for modal dialogs (main window is wider than the HUD).
func (h *HUD) SetDialogParent(w fyne.Window) { h.dialogParent = w }

func (h *HUD) dialogWindow() fyne.Window {
	if h.dialogParent != nil {
		return h.dialogParent
	}
	return h.window
}

func (h *HUD) prepareDialog() fyne.Window {
	parent := h.dialogWindow()
	if parent != nil && parent != h.window {
		parent.Show()
		parent.RequestFocus()
	}
	return parent
}

func (h *HUD) Show() {
	h.visible = true
	h.window.Show()
	onMain(func() {
		h.refresh()
		h.placeHUD(false)
	})
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
	h.dismissHint()
	h.placeHUD(true)
}

func (h *HUD) placeHUD(animate bool) {
	if !h.visible {
		return
	}
	h.placementGen++
	gen := h.placementGen
	corner := platform.CornerFromInt(h.app.Coordinator.HUDCorner())
	w, ht := int(h.size.Width), int(h.size.Height)
	win := h.app.Platform.Window()
	place := func() { win.PlaceHUD(corner, w, ht, animate) }
	place()
	if animate {
		return
	}
	go func() {
		for _, d := range []time.Duration{50 * time.Millisecond, 200 * time.Millisecond, 500 * time.Millisecond} {
			time.Sleep(d)
			if !h.visible || gen != h.placementGen {
				return
			}
			onMain(place)
		}
	}()
}

func (h *HUD) refresh() {
	title, _, elapsed, running := h.app.DisplayContext()
	if title == "" || title == "Track App" {
		title = "Select job"
	}
	h.jobBtn.SetText(title + " ▾")
	h.clock.SetTime(format.Duration(elapsed), running)
	setHUDTransport(h.play, running)
	h.refreshBanner()
}

func (h *HUD) refreshBanner() {
	h.banner.Objects = nil
	prompt := h.app.Coordinator.AutoStartPrompt()
	if prompt == nil {
		h.banner.Hide()
		base := h.compactSize
		if !h.app.Coordinator.HUDHintDismissed() {
			base = h.hintSize
		}
		h.size = base
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
	h.size = fyne.NewSize(h.compactSize.Width, h.compactSize.Height+36)
	if !h.app.Coordinator.HUDHintDismissed() {
		h.size = fyne.NewSize(h.hintSize.Width, h.hintSize.Height+36)
	}
	h.window.Resize(h.size)
	h.placeHUD(false)
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
		h.showQuickJobDialog()
		return
	}
	pop := widget.NewPopUpMenu(fyne.NewMenu("Jobs", menuItems...), h.window.Canvas())
	pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(h.jobBtn)
	pop.ShowAtPosition(pos.Add(fyne.NewPos(0, h.jobBtn.Size().Height)))
}

func (h *HUD) timerMenuItem(id string) *fyne.MenuItem {
	timerID := id
	for _, t := range h.app.Store.Timers {
		if t.ID != id {
			continue
		}
		label := t.Name
		if t.Running {
			label = "● " + label
		}
		return fyne.NewMenuItem(label, func() {
			h.app.Coordinator.SetFocus(timerID, false)
		})
	}
	return nil
}

func (h *HUD) showQuickJobDialog() {
	parent := h.prepareDialog()
	name := widget.NewEntry()
	name.SetPlaceHolder("e.g. Client website")
	hint := mutedLabel("Name your job — tracking starts right away.")

	var dlg dialog.Dialog
	add := primaryButton("Add & start", func() {
		n := strings.TrimSpace(name.Text)
		if n == "" {
			dialog.ShowInformation("New job", "Enter a job name", h.dialogWindow())
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
	dlg = dialog.NewCustom("New job", "Cancel", fluidCard(body, cardAccent), parent)
	dlg.Show()
}

func (h *HUD) showQuickProjectDialog() {
	parent := h.prepareDialog()
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
			dialog.ShowInformation("New project", "Enter a project name", h.dialogWindow())
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
	dlg = dialog.NewCustom("Link to project", "Cancel", fluidCard(form, cardDefault), parent)
	dlg.Show()
}