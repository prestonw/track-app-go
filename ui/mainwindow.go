package ui

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/prestonw/track-app-go/internal/app"
	"github.com/prestonw/track-app-go/internal/format"
	"github.com/prestonw/track-app-go/internal/models"
)

type MainWindow struct {
	app      *app.TrackApp
	hud      *HUD
	window   fyne.Window
	content  *fyne.Container
	nav      *widget.List
	sections []string

	selectedProject int

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
}

func NewMainWindow(a *app.TrackApp, fyneApp fyne.App, hud *HUD) *MainWindow {
	m := &MainWindow{
		app: a, hud: hud,
		sections: []string{"Today", "Job Timers", "Clients", "Projects", "Activity", "Report", "Settings"},
	}
	m.window = fyneApp.NewWindow("Track App")
	m.window.Resize(fyne.NewSize(1000, 700))
	m.window.SetMaster()

	m.nav = widget.NewList(
		func() int { return len(m.sections) },
		func() fyne.CanvasObject { return widget.NewLabel("section") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(m.sections[i])
		},
	)
	m.nav.OnSelected = func(id widget.ListItemID) { m.showSection(int(id)) }

	m.content = container.NewMax()
	body := container.NewBorder(nil, nil, container.NewGridWrap(fyne.NewSize(180, 36), m.nav), nil, m.content)
	m.window.SetContent(body)

	a.OnChange(func() { m.refreshCurrent() })
	m.nav.Select(0)
	return m
}

func (m *MainWindow) Show() { m.window.Show() }

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

func (m *MainWindow) refreshCurrent() { m.showSection(currentSection) }

func (m *MainWindow) buildToday() fyne.CanvasObject {
	rows := m.app.Coordinator.TodayRows()
	total := 0
	for _, r := range rows {
		total += r.Seconds
	}
	header := monoLabel(format.HumanDuration(total))
	sub := widget.NewLabel("Live summary for today")
	cards := container.NewVBox()
	if len(rows) == 0 {
		cards.Add(widget.NewLabel("No time tracked yet today. Start a job from the floating timer."))
	}
	for _, r := range rows {
		r := r
		name := r.Name
		if r.Running {
			name = "● " + name
		}
		left := container.NewVBox(headingLabel(name), widget.NewLabel(r.Client))
		right := widget.NewLabel(format.Duration(r.Seconds))
		row := container.NewBorder(nil, nil, nil, right, left)
		btn := widget.NewButton("Switch to job", func() { m.app.Coordinator.SetFocus(r.TimerID, false) })
		cards.Add(container.NewVBox(widget.NewCard("", "", row), btn))
	}
	return container.NewPadded(container.NewVBox(
		headingLabel("Today"),
		sub, header, widget.NewSeparator(), cards,
	))
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
	addBtn := widget.NewButton("Add job timer", func() {
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
	})
	form := widget.NewCard("Add job", "", container.NewVBox(name, tags, rate, clientSel, addBtn))

	list := container.NewVBox()
	for _, t := range m.app.Store.Timers {
		t := t
		elapsed := t.CurrentElapsed(models.NowMs())
		title := t.Name
		if t.Running {
			title = "● " + title
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
		actions := container.NewHBox(startBtn,
			widget.NewButton("Edit", func() { showEditTimer(m.window, m.app, t) }),
			widget.NewButton("Reset", func() { m.app.Manager.Reset(t.ID) }),
			widget.NewButton("Delete", func() { m.app.Store.DeleteTimer(t.ID); m.app.Notify() }),
		)
		list.Add(widget.NewCard(title, meta, container.NewBorder(nil, nil, nil, timeLbl, actions)))
	}
	scroll := container.NewVScroll(list)
	return container.NewBorder(
		headingLabel("Job Timers"), nil, nil, nil,
		container.NewVBox(form, widget.NewSeparator(), scroll),
	)
}

func (m *MainWindow) buildClients() fyne.CanvasObject {
	name := widget.NewEntry()
	name.SetPlaceHolder("Client or company name")
	add := widget.NewButton("Add client", func() {
		if _, err := m.app.Store.AddClient(name.Text); err != nil {
			dialog.ShowError(err, m.window)
			return
		}
		m.app.Notify()
		name.SetText("")
	})
	form := widget.NewCard("Add client", "", container.NewHBox(name, add))
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
		list.Add(widget.NewCard(c.Name, "", container.NewBorder(nil, nil, nil, del, layout.NewSpacer())))
	}
	return container.NewVBox(headingLabel("Clients"), form, container.NewVScroll(list))
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
	addProj := widget.NewButton("Add project", func() {
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
	})
	left := widget.NewCard("Projects", m.app.Monitor.TrustHint(), container.NewVBox(name, clientSel, timerSel, auto, addProj))

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
		m.showSection(currentSection)
	}
	selected := m.selectedProject

	pattern := widget.NewEntry()
	pattern.SetPlaceHolder("Match pattern")
	kindOpts := []string{string(models.RuleAppBundle), string(models.RuleWindowTitle), string(models.RuleDocumentPath)}
	kindSel := widget.NewSelect(kindOpts, nil)
	kindSel.SetSelected(kindOpts[0])
	capture := widget.NewButton("Use current window", func() {
		ctx := m.app.Monitor.CurrentForeground()
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
	right := widget.NewCard("Rules", "Select a project and add match rules", container.NewVBox(
		projectList, widget.NewSeparator(),
		container.NewHBox(kindSel, pattern), capture, addRule,
		container.NewVScroll(rulesList),
	))
	return container.NewHSplit(left, right)
}

func (m *MainWindow) buildActivity() fyne.CanvasObject {
	segs := m.app.Store.UngroupedActivity(7)
	total := 0
	for _, s := range segs {
		total += s.DurationSeconds(models.NowMs())
	}
	header := widget.NewLabel(fmt.Sprintf("%d segments · %s unassigned", len(segs), format.HumanDuration(total)))
	list := widget.NewList(
		func() int { return len(segs) },
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewLabel("a"), widget.NewLabel("b"), widget.NewLabel("c"))
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			s := segs[i]
			box := o.(*fyne.Container)
			box.Objects[0].(*widget.Label).SetText(format.DateTime(s.StartedAt))
			box.Objects[1].(*widget.Label).SetText(format.Duration(s.DurationSeconds(models.NowMs())))
			title := s.WindowTitle
			if title == "" {
				title = s.AppName
			}
			box.Objects[2].(*widget.Label).SetText(title)
		},
	)
	projOpts, projIDs := m.projectOpts()
	projSel := widget.NewSelect(projOpts, nil)
	if len(projOpts) > 0 {
		projSel.SetSelected(projOpts[0])
	}
	selectedRow := -1
	list.OnSelected = func(id widget.ListItemID) { selectedRow = int(id) }
	assign := widget.NewButton("Assign to project", func() {
		if selectedRow < 0 || selectedRow >= len(segs) {
			dialog.ShowInformation("Activity", "Select a segment", m.window)
			return
		}
		idx := indexOf(projOpts, projSel.Selected)
		if idx < 0 {
			return
		}
		m.app.Store.AssignActivity(segs[selectedRow].ID, projIDs[idx])
		m.app.Notify()
	})
	return container.NewBorder(header, container.NewHBox(projSel, assign), nil, nil, list)
}

func (m *MainWindow) buildSettings() fyne.CanvasObject {
	showHUD := widget.NewCheck("Show floating timer on launch", nil)
	showHUD.SetChecked(m.app.Coordinator.ShowHUDOnLaunch())
	showHUD.OnChanged = func(v bool) { m.app.Coordinator.SetShowHUDOnLaunch(v) }

	showNow := widget.NewButton("Show floating timer now", func() { m.hud.Show() })
	hideNow := widget.NewButton("Hide floating timer", func() { m.hud.Hide() })

	currOpts := []string{}
	for _, c := range format.AllCurrencies(m.app.Store.CustomCurrencies) {
		currOpts = append(currOpts, c.Code)
	}
	currSel := widget.NewSelect(currOpts, func(code string) { format.DisplayCurrency = code; m.app.Notify() })
	currSel.SetSelected(format.DisplayCurrency)

	return container.NewVBox(
		headingLabel("Settings"),
		widget.NewCard("Floating timer", "", container.NewVBox(showHUD, container.NewHBox(showNow, hideNow))),
		widget.NewCard("Display currency", "", currSel),
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