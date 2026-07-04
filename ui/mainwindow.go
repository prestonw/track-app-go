package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/prestonw/track-app-go/internal/app"
	"github.com/prestonw/track-app-go/internal/format"
	"github.com/prestonw/track-app-go/internal/models"
	"github.com/prestonw/track-app-go/internal/platform"
)

type MainWindow struct {
	app      *app.TrackApp
	fyneApp  fyne.App
	hud      *HUD
	window   fyne.Window
	content  *fyne.Container
	navBox   *fyne.Container
	sections []string

	selectedProject int
	ready             bool

	reportRange        string
	reportType         string
	reportClient       string
	reportTag          string
	reportSelected     map[string]bool
	reportDisplayed    []models.Session
	reportRangeSel     *widget.Select
	reportTypeSel      *widget.Select
	reportClientSel    *widget.Select
	reportTagSel       *widget.Select
	reportTimerSummary *fyne.Container
	reportSessionsList *fyne.Container
	reportFromDate     string
	reportToDate       string
	reportDateFrom     *widget.Entry
	reportDateTo       *widget.Entry
	reportDateRow      *fyne.Container
	reportRefreshing   bool

	activeSection  int
	todaySelected  bulkSet
	timerSelected  bulkSet
	activitySelected bulkSet
}

func NewMainWindow(a *app.TrackApp, fyneApp fyne.App, hud *HUD) *MainWindow {
	m := &MainWindow{
		app: a, fyneApp: fyneApp, hud: hud,
		sections: []string{"Today", "Job Timers", "Clients", "Projects", "Activity", "Report", "Settings"},
	}
	m.window = fyneApp.NewWindow("Track App")
	m.window.Resize(fyne.NewSize(1080, 720))
	m.window.SetMaster()
	m.window.SetIcon(AppIcon())
	m.todaySelected = newBulkSet()
	m.timerSelected = newBulkSet()
	m.activitySelected = newBulkSet()

	m.navBox = container.NewVBox()
	m.rebuildNav()
	m.content = container.NewMax()
	navScroll := container.NewVScroll(m.navBox)
	navScroll.SetMinSize(fyne.NewSize(sidebarW, 420))
	shell := container.NewBorder(nil, nil, sidebarPanel(navScroll), nil, m.content)

	inner := container.NewInnerWindow("Track App", shell)
	inner.Icon = AppIcon()
	inner.CloseIntercept = m.promptCloseOrTray
	inner.OnMinimized = m.minimizeToTray
	inner.SetPadded(false)
	m.window.SetContent(inner)
	m.window.SetCloseIntercept(m.promptCloseOrTray)

	a.OnChange(func() { onMain(m.refreshCurrent) })
	return m
}

// Window returns the underlying Fyne window (may be hidden).
func (m *MainWindow) Window() fyne.Window { return m.window }

func (m *MainWindow) Show() {
	if !m.ready {
		m.ready = true
		m.selectSection(0)
	}
	m.window.Show()
	m.window.RequestFocus()
	go func() {
		time.Sleep(120 * time.Millisecond)
		onMain(platform.MainWindowBorderless)
	}()
}

func (m *MainWindow) rebuildNav() {
	m.navBox.Objects = nil
	for i, section := range m.sections {
		i, section := i, section
		m.navBox.Add(navButton(navLabel(section), i == m.activeSection, func() { m.selectSection(i) }))
	}
	m.navBox.Refresh()
}

func (m *MainWindow) hideToTray() {
	platform.MainWindowHideAnimated()
	go func() {
		time.Sleep(220 * time.Millisecond)
		onMain(m.window.Hide)
	}()
}

func (m *MainWindow) minimizeToTray() { m.hideToTray() }

func (m *MainWindow) promptCloseOrTray() {
	var dlg dialog.Dialog
	body := container.NewVBox(
		mutedLabel("Keep Track App running in the menu bar, or quit entirely?"),
		container.NewHBox(
			primaryButton("Close to menu bar", func() {
				if dlg != nil {
					dlg.Hide()
				}
				m.hideToTray()
			}),
			widget.NewButton("Quit", func() {
				if dlg != nil {
					dlg.Hide()
				}
				m.fyneApp.Quit()
			}),
		),
	)
	dlg = dialog.NewCustom("Close Track App?", "Cancel", fluidCard(body, cardDefault), m.window)
	dlg.Show()
}

// OpenSection shows the main window on the named sidebar section.
func (m *MainWindow) OpenSection(name string) {
	m.Show()
	for i, s := range m.sections {
		if s == name {
			m.selectSection(i)
			return
		}
	}
}

func (m *MainWindow) selectSection(i int) {
	if i < 0 || i >= len(m.sections) {
		return
	}
	m.activeSection = i
	m.rebuildNav()
	m.showSection(i)
}

var currentSection int

func (m *MainWindow) showSection(i int) {
	currentSection = i
	switch m.sections[i] {
	case "Today":
		m.content.Objects = []fyne.CanvasObject{m.buildToday()}
	case "Job Timers":
		m.content.Objects = []fyne.CanvasObject{m.buildTimers()}
	case "Clients":
		m.content.Objects = []fyne.CanvasObject{m.buildClients()}
	case "Projects":
		m.content.Objects = []fyne.CanvasObject{m.buildProjects()}
	case "Activity":
		m.content.Objects = []fyne.CanvasObject{m.buildActivity()}
	case "Report":
		m.content.Objects = []fyne.CanvasObject{m.buildReport()}
	case "Settings":
		m.content.Objects = []fyne.CanvasObject{m.buildSettings()}
	}
	m.content.Refresh()
}

func (m *MainWindow) refreshCurrent() {
	if m.sections[currentSection] == "Report" && m.reportTimerSummary != nil {
		m.refreshReport()
		return
	}
	m.showSection(currentSection)
}

func (m *MainWindow) buildToday() fyne.CanvasObject {
	rows := m.app.Coordinator.TodayRows()
	total := 0
	ids := make([]string, 0, len(rows))
	for _, r := range rows {
		total += r.Seconds
		ids = append(ids, r.TimerID)
	}
	totalCard := fluidCard(container.NewVBox(
		sectionLabel("TOTAL TODAY"),
		monoLabel(format.HumanDuration(total)),
	), cardAccent)

	refresh := func() { m.showSection(currentSection) }
	bulk := bulkActionBar(ids, m.todaySelected, refresh,
		func(sel []string) {
			for _, id := range sel {
				m.app.Store.DeleteTimer(id)
			}
			m.app.Notify()
		},
		func(sel []string) {
			for _, id := range sel {
				m.app.Store.ArchiveTimer(id)
			}
			m.app.Notify()
		},
	)

	cards := container.NewVBox()
	if len(rows) == 0 {
		cards.Add(fluidCard(mutedLabel("No time tracked yet today. Start a job from the floating timer or add one under Job Timers."), cardDefault))
	}
	for _, r := range rows {
		r := r
		name := r.Name
		style := cardDefault
		if r.Running {
			name = "● " + name
			style = cardRunning
		}
		chk := rowCheckbox(r.TimerID, m.todaySelected, refresh)
		left := container.NewVBox(pageTitle(name), mutedLabel(r.Client))
		right := monoLabel(format.Duration(r.Seconds))
		row := container.NewBorder(nil, nil, chk, right, left)
		btn := widget.NewButton("Switch to job", func() { m.app.Coordinator.SetFocus(r.TimerID, false) })
		cards.Add(container.NewVBox(fluidCard(row, style), btn))
	}
	body := container.NewVBox(totalCard, bulk, widget.NewSeparator(), cards)
	return pageChrome("Today", "Live summary for every job timer — filtered to today.", body)
}

func (m *MainWindow) buildTimers() fyne.CanvasObject {
	name := widget.NewEntry()
	name.SetPlaceHolder("Job name")
	tags := widget.NewEntry()
	tags.SetPlaceHolder("Tags (comma separated)")
	rate := widget.NewEntry()
	rate.SetPlaceHolder("Rate per hour")
	clientOpts := []string{"— No client —"}
	clientIDs := []string{""}
	for _, c := range m.app.Store.Clients {
		clientOpts = append(clientOpts, c.Name)
		clientIDs = append(clientIDs, c.ID)
	}
	clientSel := widget.NewSelect(clientOpts, nil)
	if len(clientOpts) > 0 {
		clientSel.SetSelected(clientOpts[0])
	}
	addJob := func() {
		r, _ := strconv.ParseFloat(rate.Text, 64)
		tagList := splitTags(tags.Text)
		cid := ""
		if idx := indexOf(clientOpts, clientSel.Selected); idx >= 0 && idx < len(clientIDs) {
			cid = clientIDs[idx]
		}
		if strings.TrimSpace(name.Text) == "" {
			dialog.ShowInformation("Job timer", "Enter a job name", m.window)
			return
		}
		m.app.Store.AddTimer(name.Text, tagList, r, format.DisplayCurrency, "", cid, 0)
		m.app.Notify()
		name.SetText("")
		tags.SetText("")
		rate.SetText("")
	}
	form := fluidCardTitled("Add job", "Create a timer for billable or focus work", container.NewVBox(name, tags, rate, clientSel, primaryButton("Add job timer", addJob)), cardDefault)

	timers := m.app.Store.ActiveTimers()
	timerIDs := make([]string, 0, len(timers))
	for _, t := range timers {
		timerIDs = append(timerIDs, t.ID)
	}
	refresh := func() { m.showSection(currentSection) }
	bulk := bulkActionBar(timerIDs, m.timerSelected, refresh,
		func(sel []string) {
			for _, id := range sel {
				m.app.Store.DeleteTimer(id)
			}
			m.app.Notify()
		},
		func(sel []string) {
			for _, id := range sel {
				m.app.Store.ArchiveTimer(id)
			}
			m.app.Notify()
		},
	)

	list := container.NewVBox()
	for _, t := range timers {
		t := t
		elapsed := t.CurrentElapsed(models.NowMs())
		title := t.Name
		style := cardDefault
		if t.Running {
			title = "● " + title
			style = cardRunning
		}
		meta := m.app.Store.ClientName(t.ClientID)
		if len(t.Tags) > 0 {
			meta += " · " + strings.Join(t.Tags, ", ")
		}
		timeLbl := monoLabel(format.Duration(elapsed))
		startBtn := widget.NewButton("Start", func() { m.app.Manager.Start(t.ID) })
		if t.Running {
			startBtn.SetText("Pause")
			startBtn.OnTapped = func() { m.app.Manager.Pause(t.ID) }
		}
		chk := rowCheckbox(t.ID, m.timerSelected, refresh)
		actions := container.NewHBox(startBtn,
			widget.NewButton("Edit", func() { showEditTimer(m.window, m.app, t) }),
			widget.NewButton("Reset", func() { m.app.Manager.Reset(t.ID) }),
		)
		head := container.NewBorder(nil, nil, chk, timeLbl, container.NewVBox(pageTitle(title), mutedLabel(meta)))
		inner := container.NewVBox(head, actions)
		list.Add(fluidCard(inner, style))
	}
	body := container.NewVBox(form, bulk, widget.NewSeparator(), list)
	return pageChrome("Job Timers", "Manage jobs, rates, and clients. Start tracking from here or the floating timer.", body)
}

func (m *MainWindow) buildClients() fyne.CanvasObject {
	name := widget.NewEntry()
	name.SetPlaceHolder("Client or company name")
	addClient := func() {
		if _, err := m.app.Store.AddClient(name.Text); err != nil {
			dialog.ShowError(err, m.window)
			return
		}
		m.app.Notify()
		name.SetText("")
	}
	form := fluidCardTitled("Add client", "", container.NewHBox(name, primaryButton("Add client", addClient)), cardDefault)
	list := container.NewVBox()
	for _, c := range m.app.Store.Clients {
		c := c
		del := widget.NewButton("Delete", func() {
			if err := m.app.Store.DeleteClient(c.ID); err != nil {
				dialog.ShowInformation("Clients", "Remove client from jobs and projects first", m.window)
				return
			}
			m.app.Notify()
		})
		list.Add(fluidCard(container.NewBorder(nil, nil, pageTitle(c.Name), del, nil), cardDefault))
	}
	body := container.NewVBox(form, widget.NewSeparator(), list)
	return pageChrome("Clients", "Companies and contacts linked to job timers and projects.", body)
}

func (m *MainWindow) buildProjects() fyne.CanvasObject {
	name := widget.NewEntry()
	name.SetPlaceHolder("New project name")
	auto := widget.NewCheck("Auto-track when rules match", nil)
	clientOpts, clientIDs := m.popupOpts(m.app.Store.Clients, "(none)")
	timerOpts, timerIDs := m.timerOpts()
	clientSel := widget.NewSelect(clientOpts, nil)
	timerSel := widget.NewSelect(timerOpts, nil)
	if len(clientOpts) > 0 {
		clientSel.SetSelected(clientOpts[0])
	}
	if len(timerOpts) > 0 {
		timerSel.SetSelected(timerOpts[0])
	}
	addProject := func() {
		cid, tid := "", ""
		if i := indexOf(clientOpts, clientSel.Selected); i >= 0 {
			cid = clientIDs[i]
		}
		if i := indexOf(timerOpts, timerSel.Selected); i >= 0 {
			tid = timerIDs[i]
		}
		m.app.Store.AddProject(name.Text, cid, tid, auto.Checked, "")
		m.app.Notify()
		name.SetText("")
	}
	left := fluidCardTitled("Projects", m.app.Platform.Capabilities().ForegroundHint, container.NewVBox(name, clientSel, timerSel, auto, primaryButton("Add project", addProject)), cardDefault)

	projectList := widget.NewList(
		func() int { return len(m.app.Store.Projects) },
		func() fyne.CanvasObject { return widget.NewLabel("p") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			p := m.app.Store.Projects[i]
			o.(*widget.Label).SetText(p.Name)
		},
	)
	projectList.OnSelected = func(id widget.ListItemID) {
		m.selectedProject = int(id)
		m.content.Objects = []fyne.CanvasObject{m.buildProjects()}
		m.content.Refresh()
	}
	selected := m.selectedProject

	pattern := widget.NewEntry()
	pattern.SetPlaceHolder("Match pattern")
	kindOpts := []string{string(models.RuleAppBundle), string(models.RuleWindowTitle), string(models.RuleDocumentPath)}
	kindSel := widget.NewSelect(kindOpts, nil)
	kindSel.SetSelected(kindOpts[0])
	capture := widget.NewButton("Use current window", func() {
		ctx := m.app.Platform.Foreground().CurrentForeground()
		if ctx.BundleID != "" {
			kindSel.SetSelected(string(models.RuleAppBundle))
			pattern.SetText(ctx.BundleID)
		} else if ctx.WindowTitle != "" {
			kindSel.SetSelected(string(models.RuleWindowTitle))
			pattern.SetText(ctx.WindowTitle)
		}
	})
	addRule := widget.NewButton("Add rule", func() {
		if selected < 0 || selected >= len(m.app.Store.Projects) {
			dialog.ShowInformation("Projects", "Select a project first", m.window)
			return
		}
		pid := m.app.Store.Projects[selected].ID
		kind := models.RuleKind(kindSel.Selected)
		if _, err := m.app.Store.AddRule(pid, kind, pattern.Text); err != nil {
			dialog.ShowError(err, m.window)
			return
		}
		m.app.Notify()
		pattern.SetText("")
	})
	rulesList := container.NewVBox()
	if selected >= 0 && selected < len(m.app.Store.Projects) {
		for _, r := range m.app.Store.RulesFor(m.app.Store.Projects[selected].ID) {
			r := r
			rulesList.Add(container.NewHBox(
				widget.NewLabel(string(r.Kind)+": "+r.Pattern),
				widget.NewButton("Delete", func() { m.app.Store.DeleteRule(r.ID); m.app.Notify() }),
			))
		}
	}
	right := fluidCardTitled("Rules", "Select a project and add match rules", container.NewVBox(
		projectList, widget.NewSeparator(),
		container.NewHBox(kindSel, pattern), capture, addRule,
		container.NewVScroll(rulesList),
	), cardDefault)
	split := container.NewHSplit(left, right)
	split.SetOffset(0.42)
	return pageChrome("Projects", "Group foreground apps into projects and link them to job timers.", split)
}

func (m *MainWindow) buildActivity() fyne.CanvasObject {
	segs := m.app.Store.UngroupedActivity(7)
	total := 0
	segIDs := make([]string, 0, len(segs))
	for _, s := range segs {
		total += s.DurationSeconds(models.NowMs())
		segIDs = append(segIDs, s.ID)
	}
	header := widget.NewLabel(fmt.Sprintf("%d segments · %s unassigned", len(segs), format.HumanDuration(total)))

	refresh := func() { m.showSection(currentSection) }
	bulk := bulkActionBar(segIDs, m.activitySelected, refresh,
		func(sel []string) {
			m.app.Store.DeleteActivities(sel)
			m.app.Notify()
		},
		func(sel []string) {
			m.app.Store.ArchiveActivities(sel)
			m.app.Notify()
		},
	)

	rows := container.NewVBox()
	for _, s := range segs {
		s := s
		title := s.WindowTitle
		if title == "" {
			title = s.AppName
		}
		chk := rowCheckbox(s.ID, m.activitySelected, refresh)
		line := container.NewHBox(
			chk,
			widget.NewLabel(format.DateTime(s.StartedAt)),
			monoLabel(format.Duration(s.DurationSeconds(models.NowMs()))),
			widget.NewLabel(title),
		)
		rows.Add(fluidCard(line, cardDefault))
	}
	if len(segs) == 0 {
		rows.Add(fluidCard(mutedLabel("No unassigned activity in the last 7 days."), cardDefault))
	}

	projOpts, projIDs := m.projectOpts()
	projOpts = append([]string{"+ New project…"}, projOpts...)
	projIDs = append([]string{""}, projIDs...)
	projSel := widget.NewSelect(projOpts, nil)
	if len(projOpts) > 0 {
		projSel.SetSelected(projOpts[0])
	}
	assign := widget.NewButton("Assign to project", func() {
		if projSel.Selected == "+ New project…" {
			m.showNewProjectDialog(func(projID string) {
				if projID == "" {
					return
				}
				for _, id := range m.activitySelected.ids() {
					m.app.Store.AssignActivity(id, projID)
				}
				m.activitySelected.clear()
				m.app.Notify()
			})
			return
		}
		ids := m.activitySelected.ids()
		if len(ids) == 0 {
			dialog.ShowInformation("Activity", "Select one or more segments", m.window)
			return
		}
		idx := indexOf(projOpts, projSel.Selected)
		if idx < 0 {
			return
		}
		for _, id := range ids {
			m.app.Store.AssignActivity(id, projIDs[idx])
		}
		m.activitySelected.clear()
		m.app.Notify()
	})
	toolbar := fluidCard(container.NewHBox(projSel, assign), cardDefault)
	body := container.NewVBox(header, bulk, toolbar, widget.NewSeparator(), rows)
	return pageChrome("Activity", "Ungrouped foreground time from the last 7 days — assign to projects.", body)
}

func (m *MainWindow) showNewProjectDialog(onDone func(string)) {
	name := widget.NewEntry()
	name.SetPlaceHolder("Project name")
	auto := widget.NewCheck("Auto-track when rules match", nil)
	clientOpts, clientIDs := m.popupOpts(m.app.Store.Clients, "(none)")
	timerOpts, timerIDs := m.timerOpts()
	clientSel := widget.NewSelect(clientOpts, nil)
	timerSel := widget.NewSelect(timerOpts, nil)
	if len(clientOpts) > 0 {
		clientSel.SetSelected(clientOpts[0])
	}
	if len(timerOpts) > 0 {
		timerSel.SetSelected(timerOpts[0])
	}
	var dlg dialog.Dialog
	add := primaryButton("Add project", func() {
		n := strings.TrimSpace(name.Text)
		if n == "" {
			dialog.ShowInformation("New project", "Enter a project name", m.window)
			return
		}
		cid, tid := "", ""
		if i := indexOf(clientOpts, clientSel.Selected); i >= 0 {
			cid = clientIDs[i]
		}
		if i := indexOf(timerOpts, timerSel.Selected); i >= 0 {
			tid = timerIDs[i]
		}
		p := m.app.Store.AddProject(n, cid, tid, auto.Checked, "")
		if dlg != nil {
			dlg.Hide()
		}
		m.app.Notify()
		if onDone != nil {
			onDone(p.ID)
		}
	})
	form := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Name", name),
			widget.NewFormItem("Job timer", timerSel),
			widget.NewFormItem("Client", clientSel),
		),
		auto,
		add,
	)
	dlg = dialog.NewCustom("New project", "Cancel", fluidCard(form, cardDefault), m.window)
	dlg.Show()
}

func (m *MainWindow) buildSettings() fyne.CanvasObject {
	showHUD := widget.NewCheck("Show floating timer on launch", nil)
	showHUD.SetChecked(m.app.Coordinator.ShowHUDOnLaunch())
	showHUD.OnChanged = func(v bool) {
		m.app.Coordinator.SetShowHUDOnLaunch(v)
		if v {
			m.hud.Show()
		} else {
			m.hud.Hide()
		}
	}

	showNow := widget.NewButton("Show floating timer now", func() { m.hud.Show() })
	hideNow := widget.NewButton("Hide floating timer", func() { m.hud.Hide() })

	currOpts := []string{}
	for _, c := range format.AllCurrencies(m.app.Store.CustomCurrencies) {
		currOpts = append(currOpts, c.Code)
	}
	currSel := widget.NewSelect(currOpts, func(code string) { format.DisplayCurrency = code; m.app.Notify() })
	currSel.SetSelected(format.DisplayCurrency)

	cap := m.app.Platform.Capabilities()
	fgStatus := "not available"
	switch cap.Foreground {
	case platform.ForegroundFull:
		if cap.ForegroundTrusted {
			fgStatus = "full (trusted)"
		} else {
			fgStatus = "limited — permission required"
		}
	case platform.ForegroundBasic:
		fgStatus = "basic"
	}
	snapStatus := "unavailable"
	if cap.WindowCornerSnap {
		snapStatus = "available"
	}
	refreshPlat := widget.NewButton("Refresh platform status", func() {
		cap = m.app.Platform.RefreshCapabilities()
		m.app.Notify()
	})
	platformCard := fluidCardTitled("Platform", string(cap.OS)+" ("+cap.GoOS+")", container.NewVBox(
		mutedLabel("Foreground tracking: "+fgStatus),
		mutedLabel(cap.ForegroundHint),
		mutedLabel("HUD corner snap: "+snapStatus),
		mutedLabel(cap.WindowHint),
		refreshPlat,
	), cardDefault)
	hudCard := fluidCardTitled("Floating timer", "Overlay timer for quick job switching", container.NewVBox(showHUD, container.NewHBox(showNow, hideNow)), cardDefault)
	currCard := fluidCardTitled("Display currency", "", currSel, cardDefault)
	replayOnboarding := widget.NewButton("Show onboarding again", func() {
		m.app.Coordinator.SetOnboardingComplete(false)
		ShowOnboardingIfNeeded(m.app, m.fyneApp, m, m.hud)
	})
	resetApp := widget.NewButton("Reset app data…", func() { m.confirmResetApp() })
	dataCard := fluidCardTitled("Data", "Restore factory defaults for testing", container.NewVBox(
		mutedLabel("Deletes all jobs, projects, sessions, activity, and preferences. Cannot be undone."),
		resetApp,
		replayOnboarding,
	), cardDefault)
	aboutCard := fluidCard(container.NewVBox(
		mutedLabel("Build: "+BuildVersion()),
		mutedLabel("On macOS, use Track App.app from scripts/build-macos-app.sh — not the bare trackapp binary."),
	), cardDefault)
	body := container.NewVBox(platformCard, hudCard, currCard, dataCard, aboutCard)
	return pageChrome("Settings", "App preferences and platform integration status.", body)
}

func (m *MainWindow) confirmResetApp() {
	dialog.ShowConfirm(
		"Reset Track App",
		"Delete all jobs, projects, sessions, activity, and preferences?\n\nThis cannot be undone.",
		func(ok bool) {
			if !ok {
				return
			}
			if err := m.app.Store.ResetAll(); err != nil {
				dialog.ShowError(err, m.window)
				return
			}
			m.app.Coordinator.ResetToDefaults()
			format.DisplayCurrency = "GBP"
			if m.app.Coordinator.ShowHUDOnLaunch() {
				m.hud.Show()
			} else {
				m.hud.Hide()
			}
			m.app.Notify()
			m.showSection(currentSection)
			if m.app.Coordinator.NeedsOnboarding() {
				ShowOnboardingIfNeeded(m.app, m.fyneApp, m, m.hud)
			}
			dialog.ShowInformation("Reset complete", "App restored to factory defaults.", m.window)
		},
		m.window,
	)
}

func (m *MainWindow) popupOpts(clients []models.Client, none string) ([]string, []string) {
	opts := []string{none}
	ids := []string{""}
	for _, c := range clients {
		opts = append(opts, c.Name)
		ids = append(ids, c.ID)
	}
	return opts, ids
}

func (m *MainWindow) timerOpts() ([]string, []string) {
	opts := []string{"(none)"}
	ids := []string{""}
	for _, t := range m.app.Store.Timers {
		opts = append(opts, t.Name)
		ids = append(ids, t.ID)
	}
	return opts, ids
}

func (m *MainWindow) projectOpts() ([]string, []string) {
	opts := []string{}
	ids := []string{}
	for _, p := range m.app.Store.Projects {
		opts = append(opts, p.Name)
		ids = append(ids, p.ID)
	}
	return opts, ids
}

func splitTags(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func indexOf(slice []string, val string) int {
	for i, s := range slice {
		if s == val {
			return i
		}
	}
	return -1
}