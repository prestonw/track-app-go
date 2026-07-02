package ui

import (
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/prestonw/track-app-go/internal/app"
	"github.com/prestonw/track-app-go/internal/format"
	"github.com/prestonw/track-app-go/internal/models"
)

func showEditTimer(parent fyne.Window, core *app.TrackApp, timer models.JobTimer) {
	name := widget.NewEntry()
	name.SetText(timer.Name)

	tags := widget.NewEntry()
	tags.SetText(strings.Join(timer.Tags, ", "))

	rate := widget.NewEntry()
	if timer.Rate > 0 {
		rate.SetText(strconv.FormatFloat(timer.Rate, 'f', -1, 64))
	}

	notes := widget.NewMultiLineEntry()
	notes.SetText(timer.Notes)

	currOpts := []string{}
	for _, c := range format.AllCurrencies(core.Store.CustomCurrencies) {
		currOpts = append(currOpts, c.Code)
	}
	currSel := widget.NewSelect(currOpts, nil)
	currSel.SetSelected(timer.Currency)

	clientOpts, clientIDs := clientSelectOpts(core)
	clientSel := widget.NewSelect(clientOpts, nil)
	clientSel.SetSelected(clientLabel(core, timer.ClientID))

	manualH := widget.NewEntry()
	manualH.SetPlaceHolder("0")
	manualM := widget.NewEntry()
	manualM.SetPlaceHolder("0")
	manualS := widget.NewEntry()
	manualS.SetPlaceHolder("0")

	adjLines := container.NewVBox()
	if len(timer.Adjustments) == 0 {
		adjLines.Add(widget.NewLabel("No activity yet"))
	} else {
		for i := len(timer.Adjustments) - 1; i >= 0; i-- {
			a := timer.Adjustments[i]
			adjLines.Add(widget.NewLabel(format.DateTime(a.Ts) + "  " + a.Detail))
		}
	}

	form := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Job name", name),
			widget.NewFormItem("Tags", tags),
			widget.NewFormItem("Currency", currSel),
			widget.NewFormItem("Rate /hr", rate),
			widget.NewFormItem("Client", clientSel),
			widget.NewFormItem("Notes", notes),
			widget.NewFormItem("Adjust time", container.NewHBox(
				widget.NewLabel("H"), manualH,
				widget.NewLabel("M"), manualM,
				widget.NewLabel("S"), manualS,
			)),
		),
		headingLabel("Activity log"),
		container.NewVScroll(adjLines),
	)

	d := dialog.NewCustomConfirm("Edit Job Timer", "Save changes", "Cancel", form, func(ok bool) {
		if !ok {
			return
		}
		trimmed := strings.TrimSpace(name.Text)
		if trimmed == "" {
			dialog.ShowInformation("Edit Job Timer", "Enter a job name", parent)
			return
		}
		updated := timer
		updated.Name = trimmed
		updated.Tags = splitTags(tags.Text)
		updated.Rate, _ = strconv.ParseFloat(rate.Text, 64)
		updated.Currency = currSel.Selected
		updated.Notes = notes.Text
		if idx := indexOf(clientOpts, clientSel.Selected); idx >= 0 && idx < len(clientIDs) {
			updated.ClientID = clientIDs[idx]
		}
		manualSec := parseManualSeconds(manualH.Text, manualM.Text, manualS.Text)
		core.Store.UpdateTimer(updated, manualSec)
		core.Notify()
	}, parent)
	d.Resize(fyne.NewSize(460, 520))
	d.Show()
}

func clientSelectOpts(core *app.TrackApp) ([]string, []string) {
	opts := []string{"— No client —"}
	ids := []string{""}
	for _, c := range core.Store.Clients {
		opts = append(opts, c.Name)
		ids = append(ids, c.ID)
	}
	return opts, ids
}

func clientLabel(core *app.TrackApp, clientID string) string {
	if clientID == "" {
		return "— No client —"
	}
	return core.Store.ClientName(clientID)
}

func parseManualSeconds(h, m, s string) int {
	hv, _ := strconv.Atoi(strings.TrimSpace(h))
	mv, _ := strconv.Atoi(strings.TrimSpace(m))
	sv, _ := strconv.Atoi(strings.TrimSpace(s))
	return hv*3600 + mv*60 + sv
}