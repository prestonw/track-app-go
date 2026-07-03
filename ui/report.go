package ui

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/prestonw/track-app-go/internal/export"
	"github.com/prestonw/track-app-go/internal/format"
	"github.com/prestonw/track-app-go/internal/models"
)

func (m *MainWindow) buildReport() fyne.CanvasObject {
	if m.reportRange == "" {
		m.reportRange = string(models.RangeWeek)
	}
	if m.reportType == "" {
		m.reportType = string(models.FilterAll)
	}

	rangeOpts := make([]string, len(models.AllReportRanges))
	for i, r := range models.AllReportRanges {
		rangeOpts[i] = string(r)
	}
	typeOpts := []string{string(models.FilterAll), string(models.FilterTracked), string(models.FilterManual)}

	dateFrom := widget.NewEntry()
	dateFrom.SetPlaceHolder("From YYYY-MM-DD")
	dateTo := widget.NewEntry()
	dateTo.SetPlaceHolder("To YYYY-MM-DD")
	if m.reportFromDate != "" {
		dateFrom.SetText(m.reportFromDate)
	}
	if m.reportToDate != "" {
		dateTo.SetText(m.reportToDate)
	}
	dateFrom.OnChanged = func(s string) { m.reportFromDate = s; m.refreshReport() }
	dateTo.OnChanged = func(s string) { m.reportToDate = s; m.refreshReport() }
	dateRow := container.NewHBox(dateFrom, dateTo)
	dateRow.Hide()

	rangeSel := widget.NewSelect(rangeOpts, func(s string) {
		m.reportRange = s
		if s == string(models.RangeCustom) {
			dateRow.Show()
		} else {
			dateRow.Hide()
		}
		m.refreshReport()
	})
	rangeSel.SetSelected(m.reportRange)
	if m.reportRange == string(models.RangeCustom) {
		dateRow.Show()
	}

	typeSel := widget.NewSelect(typeOpts, func(s string) {
		m.reportType = s
		m.refreshReport()
	})
	typeSel.SetSelected(m.reportType)

	clientSel := widget.NewSelect([]string{"All clients"}, func(s string) {
		m.reportClient = s
		m.refreshReport()
	})
	clientSel.SetSelected("All clients")

	tagSel := widget.NewSelect([]string{"All tags"}, func(s string) {
		m.reportTag = s
		m.refreshReport()
	})
	tagSel.SetSelected("All tags")

	timerSummary := container.NewVBox()
	sessionsList := container.NewVBox()

	m.reportRangeSel = rangeSel
	m.reportTypeSel = typeSel
	m.reportClientSel = clientSel
	m.reportTagSel = tagSel
	m.reportTimerSummary = timerSummary
	m.reportSessionsList = sessionsList
	m.reportDateFrom = dateFrom
	m.reportDateTo = dateTo
	m.reportDateRow = dateRow

	importBtn := widget.NewButton("Import…", func() { m.importReportFile() })
	exportCSV := widget.NewButton("Export CSV", func() { m.exportReportCSV() })
	exportDB := widget.NewButton("Export .sqlite", func() { m.exportReportDB() })
	deleteBtn := widget.NewButton("Delete selected", func() { m.deleteSelectedSessions() })
	deleteBtn.Importance = widget.DangerImportance

	filters := fluidCardTitled("Filters", "Range, type, and export", container.NewVBox(
		container.NewHBox(rangeSel, typeSel, clientSel, tagSel),
		dateRow,
		container.NewHBox(importBtn, exportCSV, exportDB, deleteBtn),
	), cardDefault)

	m.refreshReport()

	body := container.NewVBox(
		fluidCardTitled("By job timer", "", timerSummary, cardAccent),
		filters,
		sectionLabel("SESSIONS"),
		sessionsList,
	)
	return pageChrome("Report", "Job totals for the selected period, then individual sessions.", body)
}

func (m *MainWindow) refreshReport() {
	if m.reportTimerSummary == nil {
		return
	}

	r := models.ReportRange(m.reportRange)
	var fromPtr, toPtr *time.Time
	if r == models.RangeCustom {
		if t, err := time.Parse("2006-01-02", m.reportFromDate); err == nil {
			fromPtr = &t
		}
		if t, err := time.Parse("2006-01-02", m.reportToDate); err == nil {
			toPtr = &t
		}
	}
	from, to := format.ReportRangeBounds(r, fromPtr, toPtr)
	fromMs, toMs := from.UnixMilli(), to.UnixMilli()

	var sessions []models.Session
	for _, sess := range m.app.Store.Sessions {
		if sess.Start >= fromMs && sess.Start <= toMs {
			sessions = append(sessions, sess)
		}
	}

	m.updateReportFilterOptions(sessions)

	filterTag := ""
	if m.reportTag != "" && m.reportTag != "All tags" {
		filterTag = m.reportTag
	}
	filterClient := ""
	if m.reportClient != "" && m.reportClient != "All clients" {
		filterClient = m.reportClient
	}

	var filtered []models.Session
	for _, s := range sessions {
		if filterTag != "" && !containsTag(s.Tags, filterTag) {
			continue
		}
		if m.reportType == string(models.FilterManual) && !s.Manual {
			continue
		}
		if m.reportType == string(models.FilterTracked) && s.Manual {
			continue
		}
		if filterClient != "" && s.Client != filterClient {
			continue
		}
		filtered = append(filtered, s)
	}

	m.reportDisplayed = filtered
	m.updateReportTimerSummary(filtered)
	m.updateReportSessionsList(filtered)
}

func (m *MainWindow) updateReportFilterOptions(sessions []models.Session) {
	tagSet := map[string]struct{}{}
	clientSet := map[string]struct{}{}
	for _, s := range sessions {
		for _, t := range s.Tags {
			if t != "" {
				tagSet[t] = struct{}{}
			}
		}
		if s.Client != "" {
			clientSet[s.Client] = struct{}{}
		}
	}

	tagOpts := []string{"All tags"}
	for t := range tagSet {
		tagOpts = append(tagOpts, t)
	}
	clientOpts := []string{"All clients"}
	for c := range clientSet {
		clientOpts = append(clientOpts, c)
	}

	if m.reportTagSel != nil {
		cur := m.reportTag
		if cur == "" {
			cur = "All tags"
		}
		m.reportTagSel.Options = tagOpts
		if indexOf(tagOpts, cur) >= 0 {
			m.reportTagSel.SetSelected(cur)
		} else {
			m.reportTagSel.SetSelected("All tags")
			m.reportTag = "All tags"
		}
	}
	if m.reportClientSel != nil {
		cur := m.reportClient
		if cur == "" {
			cur = "All clients"
		}
		m.reportClientSel.Options = clientOpts
		if indexOf(clientOpts, cur) >= 0 {
			m.reportClientSel.SetSelected(cur)
		} else {
			m.reportClientSel.SetSelected("All clients")
			m.reportClient = "All clients"
		}
	}
}

func (m *MainWindow) updateReportTimerSummary(sessions []models.Session) {
	m.reportTimerSummary.Objects = nil
	byTimer := map[string]struct {
		name    string
		seconds int
		earned  map[string]float64
	}{}
	for _, sess := range sessions {
		row := byTimer[sess.TimerID]
		if row.name == "" {
			row.name = sess.Name
		}
		row.seconds += sess.Seconds
		if sess.Rate > 0 {
			if row.earned == nil {
				row.earned = map[string]float64{}
			}
			row.earned[sess.Currency] += float64(sess.Seconds) / 3600 * sess.Rate
		}
		byTimer[sess.TimerID] = row
	}
	if len(byTimer) == 0 {
		m.reportTimerSummary.Add(widget.NewLabel("No sessions in this period."))
		return
	}
	type row struct {
		name    string
		seconds int
		earned  map[string]float64
	}
	var rows []row
	for _, r := range byTimer {
		rows = append(rows, row{r.name, r.seconds, r.earned})
	}
	for i := 0; i < len(rows); i++ {
		for j := i + 1; j < len(rows); j++ {
			if rows[j].seconds > rows[i].seconds {
				rows[i], rows[j] = rows[j], rows[i]
			}
		}
	}
	for _, r := range rows {
		earned := earnedSummary(r.earned)
		line := fmt.Sprintf("%s   %s", r.name, format.HumanDuration(r.seconds))
		if earned != "" {
			line += "   ·   " + earned
		}
		m.reportTimerSummary.Add(monoLabel(line))
	}
}

func (m *MainWindow) updateReportSessionsList(sessions []models.Session) {
	m.reportSessionsList.Objects = nil
	if len(sessions) == 0 {
		m.reportSessionsList.Add(widget.NewLabel("No sessions match the current filters."))
		return
	}
	for _, sess := range sessions {
		sess := sess
		typ := "tracked"
		if sess.Manual {
			typ = "manual"
		}
		earned := "—"
		if sess.Rate > 0 {
			earned = format.Money(float64(sess.Seconds)/3600*sess.Rate, sess.Currency, m.app.Store.CustomCurrencies)
		}
		tags := strings.Join(sess.Tags, " ")
		if tags == "" {
			tags = "—"
		}
		client := sess.Client
		if client == "" {
			client = "—"
		}
		left := container.NewVBox(
			headingLabel(sess.Name),
			widget.NewLabel(fmt.Sprintf("%s · %s · %s", format.Date(sess.Start), typ, client)),
			widget.NewLabel(tags),
		)
		check := widget.NewCheck("", nil)
		check.SetChecked(m.reportSelected[sess.ID])
		check.OnChanged = func(on bool) {
			if m.reportSelected == nil {
				m.reportSelected = map[string]bool{}
			}
			if on {
				m.reportSelected[sess.ID] = true
			} else {
				delete(m.reportSelected, sess.ID)
			}
		}
		del := widget.NewButton("Delete", func() {
			m.app.Store.DeleteSession(sess.ID)
			delete(m.reportSelected, sess.ID)
			m.app.Notify()
		})
		del.Importance = widget.DangerImportance
		right := container.NewVBox(
			monoLabel(format.Duration(sess.Seconds)),
			widget.NewLabel(earned),
			container.NewHBox(check, del),
		)
		m.reportSessionsList.Add(fluidCard(container.NewBorder(nil, nil, left, right, nil), cardDefault))
	}
}

func (m *MainWindow) deleteSelectedSessions() {
	if len(m.reportSelected) == 0 {
		dialog.ShowInformation("Report", "Select one or more sessions first", m.window)
		return
	}
	for id := range m.reportSelected {
		m.app.Store.DeleteSession(id)
	}
	m.reportSelected = map[string]bool{}
	m.app.Notify()
}

func (m *MainWindow) exportReportCSV() {
	if len(m.reportDisplayed) == 0 {
		dialog.ShowInformation("Export", "No sessions to export", m.window)
		return
	}
	dialog.ShowFileSave(func(uc fyne.URIWriteCloser, err error) {
		if err != nil || uc == nil {
			return
		}
		defer uc.Close()
		csv := export.SessionsCSV(m.reportDisplayed, m.app.Store.CustomCurrencies)
		_, _ = io.WriteString(uc, csv)
	}, m.window)
}

func (m *MainWindow) exportReportDB() {
	dialog.ShowFileSave(func(uc fyne.URIWriteCloser, err error) {
		if err != nil || uc == nil {
			return
		}
		defer uc.Close()
		path := uc.URI().Path()
		if err := m.app.Store.ExportDatabase(path); err != nil {
			dialog.ShowError(err, m.window)
		}
	}, m.window)
}

func earnedSummary(byCurrency map[string]float64) string {
	if len(byCurrency) == 0 {
		return ""
	}
	var parts []string
	for code, amt := range byCurrency {
		if amt > 0 {
			parts = append(parts, format.Money(amt, code, nil))
		}
	}
	return strings.Join(parts, " + ")
}

func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (m *MainWindow) importReportFile() {
	dialog.ShowFileOpen(func(rc fyne.URIReadCloser, err error) {
		if err != nil || rc == nil {
			return
		}
		defer rc.Close()
		path := rc.URI().Path()
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".sqlite" || ext == ".db" {
			m.importSQLite(path)
			return
		}
		data, err := io.ReadAll(rc)
		if err != nil {
			dialog.ShowError(err, m.window)
			return
		}
		if err := m.app.Store.ImportCSV(string(data), models.ImportSkip); err != nil {
			dialog.ShowError(err, m.window)
			return
		}
		m.app.Notify()
		dialog.ShowInformation("Import", "Import complete", m.window)
	}, m.window)
}

func (m *MainWindow) importSQLite(path string) {
	conflicts, err := m.app.Store.DetectImportConflicts(path)
	if err != nil {
		dialog.ShowError(err, m.window)
		return
	}
	if conflicts.TimerCount == 0 && conflicts.SessionCount == 0 {
		m.finishSQLiteImport(path, models.ImportSkip)
		return
	}
	msg := fmt.Sprintf(
		"%d job timer(s) and %d session(s) already exist.\nChoose how to handle overlaps.",
		conflicts.TimerCount, conflicts.SessionCount,
	)
	var d dialog.Dialog
	closeAndImport := func(strategy models.ImportStrategy) {
		if d != nil {
			d.Hide()
		}
		m.finishSQLiteImport(path, strategy)
	}
	skip := widget.NewButton("Skip duplicates", func() { closeAndImport(models.ImportSkip) })
	merge := widget.NewButton("Merge", func() { closeAndImport(models.ImportMerge) })
	replace := widget.NewButton("Replace overlaps", func() { closeAndImport(models.ImportReplace) })
	body := container.NewVBox(
		widget.NewLabel(msg),
		container.NewHBox(skip, merge, replace),
	)
	d = dialog.NewCustom("Import database", "Cancel", body, m.window)
	d.Show()
}

func (m *MainWindow) finishSQLiteImport(path string, strategy models.ImportStrategy) {
	if err := m.app.Store.ImportDatabase(path, strategy); err != nil {
		dialog.ShowError(err, m.window)
		return
	}
	m.app.Notify()
	dialog.ShowInformation("Import", "Import complete", m.window)
}