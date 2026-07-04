package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type bulkSet map[string]bool

func newBulkSet() bulkSet { return bulkSet{} }

func (b bulkSet) has(id string) bool { return b[id] }

func (b bulkSet) set(id string, on bool) {
	if on {
		b[id] = true
	} else {
		delete(b, id)
	}
}

func (b bulkSet) ids() []string {
	out := make([]string, 0, len(b))
	for id := range b {
		out = append(out, id)
	}
	return out
}

func (b bulkSet) clear() {
	for id := range b {
		delete(b, id)
	}
}

func bulkActionBar(allIDs []string, selected bulkSet, onChanged func(), deleteFn, archiveFn func([]string)) fyne.CanvasObject {
	selectAll := visibleCheck("Select all", false, func(on bool) {
		for _, id := range allIDs {
			selected.set(id, on)
		}
		if onChanged != nil {
			onChanged()
		}
	})

	del := widget.NewButton("Delete", func() {
		ids := selected.ids()
		if len(ids) == 0 {
			return
		}
		deleteFn(ids)
		selected.clear()
		if onChanged != nil {
			onChanged()
		}
	})
	del.Importance = widget.DangerImportance

	archive := widget.NewButton("Archive", func() {
		ids := selected.ids()
		if len(ids) == 0 {
			return
		}
		archiveFn(ids)
		selected.clear()
		if onChanged != nil {
			onChanged()
		}
	})

	return fluidCard(container.NewHBox(selectAll, widget.NewLabel("·"), archive, del), cardDefault)
}

func rowCheckbox(id string, selected bulkSet, onChanged func()) fyne.CanvasObject {
	return visibleCheck("", selected.has(id), func(on bool) {
		selected.set(id, on)
		if onChanged != nil {
			onChanged()
		}
	})
}

func visibleCheck(label string, checked bool, onChanged func(bool)) fyne.CanvasObject {
	chk := widget.NewCheck(label, onChanged)
	chk.SetChecked(checked)
	bg := canvas.NewRectangle(colorCheckFill)
	bg.CornerRadius = 5
	bg.StrokeColor = colorCheckBorder
	bg.StrokeWidth = 1.5
	return container.NewStack(bg, container.NewPadded(chk))
}