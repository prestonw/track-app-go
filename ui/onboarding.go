package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/prestonw/track-app-go/internal/app"
	"github.com/prestonw/track-app-go/internal/format"
	"github.com/prestonw/track-app-go/internal/platform"
)

// ShowOnboardingIfNeeded presents the first-run wizard for new users.
func ShowOnboardingIfNeeded(core *app.TrackApp, fyneApp fyne.App, mainWin *MainWindow, hud *HUD) {
	if !core.Coordinator.NeedsOnboarding() {
		return
	}
	mainWin.window.Hide()
	w := fyneApp.NewWindow("Welcome to Track App")
	w.Resize(fyne.NewSize(520, 560))
	w.SetFixedSize(true)
	w.CenterOnScreen()
	w.SetIcon(AppIcon())

	step := 0
	jobName := widget.NewEntry()
	jobName.SetPlaceHolder("e.g. Client work, Writing, Meetings")
	showHUD := widget.NewCheck("Show floating timer on launch", nil)
	showHUD.SetChecked(core.Coordinator.ShowHUDOnLaunch())
	showHUD.OnChanged = func(v bool) {
		core.Coordinator.SetShowHUDOnLaunch(v)
		if v {
			hud.Show()
		} else {
			hud.Hide()
		}
	}

	stepTitle := widget.NewLabel("")
	stepTitle.TextStyle = fyne.TextStyle{Bold: true}
	stepBody := container.NewVBox()
	dots := widget.NewLabel("")

	var dlg dialog.Dialog
	var render func()

	render = func() {
		stepBody.Objects = nil
		switch step {
		case 0:
			stepTitle.SetText("Track time without friction")
			icon := canvas.NewImageFromResource(AppIcon())
			icon.FillMode = canvas.ImageFillContain
			icon.SetMinSize(fyne.NewSize(96, 96))
			stepBody.Add(container.NewCenter(icon))
			stepBody.Add(widget.NewLabel(""))
			stepBody.Add(mutedLabel("Track App logs time per job with a lightweight floating timer."))
			stepBody.Add(mutedLabel("Switch jobs from the overlay or main window — your data stays in a local SQLite database."))
		case 1:
			stepTitle.SetText("Create your first job")
			stepBody.Add(mutedLabel("A job timer is anything you bill or track time against."))
			stepBody.Add(widget.NewForm(widget.NewFormItem("Job name", jobName)))
		case 2:
			stepTitle.SetText("Floating timer")
			stepBody.Add(mutedLabel("The HUD stays on screen while you work. Click empty space on the panel to snap it between corners."))
			stepBody.Add(showHUD)
			if !hud.Visible() {
				stepBody.Add(primaryButton("Preview floating timer", func() { hud.Show() }))
			}
		case 3:
			stepTitle.SetText("Platform setup")
			cap := core.Platform.Capabilities()
			stepBody.Add(fluidCard(container.NewVBox(
				mutedLabel("OS: "+string(cap.OS)),
				widget.NewLabel(cap.ForegroundHint),
				widget.NewLabel(cap.WindowHint),
			), cardDefault))
			if cap.OS == platform.OSDarwin && !cap.ForegroundTrusted {
				stepBody.Add(mutedLabel("Grant Accessibility in System Settings, then refresh status in Settings later."))
			}
			refresh := widget.NewButton("Refresh status", func() {
				core.Platform.RefreshCapabilities()
				render()
			})
			stepBody.Add(refresh)
		case 4:
			stepTitle.SetText("You're ready")
			stepBody.Add(mutedLabel("Open Today to see live totals, or Projects to set up auto-tracking rules."))
			stepBody.Add(mutedLabel("Click the menu bar icon to open the app — Settings and Reset live there."))
		}
		dots.SetText(onboardingDots(step))
		stepBody.Refresh()
	}

	finish := func() {
		if dlg != nil {
			dlg.Hide()
		}
		w.Close()
		core.Coordinator.SetOnboardingComplete(true)
		if showHUD.Checked {
			hud.Show()
		}
		core.Notify()
	}

	back := widget.NewButton("Back", func() {
		if step > 0 {
			step--
			render()
		}
	})
	next := primaryButton("Continue", nil)
	next.OnTapped = func() {
		if step == 1 {
			if strings.TrimSpace(jobName.Text) == "" {
				dialog.ShowInformation("First job", "Enter a job name to continue", w)
				return
			}
			core.Store.AddTimer(jobName.Text, nil, 0, format.DisplayCurrency, "", "", 0)
			core.Coordinator.SetFocus(core.Store.Timers[0].ID, false)
			core.Notify()
		}
		if step < 4 {
			step++
			render()
			return
		}
		finish()
	}
	skip := widget.NewButton("Skip tour", finish)

	body := container.NewVBox(
		stepTitle,
		dots,
		widget.NewSeparator(),
		container.NewVScroll(stepBody),
		widget.NewSeparator(),
		container.NewBorder(nil, nil, container.NewHBox(back, skip), next, nil),
	)
	dlg = dialog.NewCustom("Onboarding", "", fluidCard(body, cardAccent), w)
	render()
	dlg.Show()
}

func onboardingDots(step int) string {
	var b strings.Builder
	for i := 0; i < 5; i++ {
		if i == step {
			b.WriteString("● ")
		} else {
			b.WriteString("○ ")
		}
	}
	return strings.TrimSpace(b.String())
}