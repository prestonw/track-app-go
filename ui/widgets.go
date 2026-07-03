package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func headingLabel(text string) *widget.Label {
	return pageTitle(text)
}

func monoLabel(text string) *widget.Label {
	l := widget.NewLabel(text)
	l.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	return l
}