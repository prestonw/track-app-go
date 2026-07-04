package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	pagePad    float32 = 28
	cardRadius float32 = 14
	sidebarW   float32 = 220
)

type cardStyle int

const (
	cardDefault cardStyle = iota
	cardAccent
	cardRunning
)

func pageTitle(text string) *widget.Label {
	l := widget.NewLabel(text)
	l.TextStyle = fyne.TextStyle{Bold: true}
	return l
}

func pageSubtitle(text string) *widget.Label {
	l := widget.NewLabel(text)
	l.Importance = widget.LowImportance
	return l
}

func sectionLabel(text string) *widget.Label {
	l := widget.NewLabel(text)
	l.TextStyle = fyne.TextStyle{Bold: true}
	l.Importance = widget.LowImportance
	return l
}

func mutedLabel(text string) *widget.Label {
	l := widget.NewLabel(text)
	l.Importance = widget.LowImportance
	return l
}

func statLabel(text string) *widget.RichText {
	rt := widget.NewRichTextFromMarkdown("## " + text)
	return rt
}

// pageChrome returns a titled page with subtitle and scrollable body.
func pageChrome(title, subtitle string, body fyne.CanvasObject) fyne.CanvasObject {
	header := container.NewVBox(
		pageTitle(title),
		pageSubtitle(subtitle),
		widget.NewSeparator(),
	)
	if body == nil {
		body = widget.NewLabel("")
	}
	scroll := container.NewVScroll(body)
	return container.NewBorder(header, nil, nil, nil, container.NewPadded(scroll))
}

func navButton(label string, selected bool, tapped func()) fyne.CanvasObject {
	btn := widget.NewButton(label, tapped)
	if selected {
		btn.Importance = widget.HighImportance
	} else {
		btn.Importance = widget.LowImportance
	}
	bg := canvas.NewRectangle(colorAccentSoft)
	bg.CornerRadius = 8
	if !selected {
		bg.Hide()
	}
	return container.NewStack(bg, btn)
}

func fluidCard(content fyne.CanvasObject, style cardStyle) fyne.CanvasObject {
	bg := canvas.NewRectangle(colorSurface)
	bg.CornerRadius = cardRadius
	bg.StrokeWidth = 1
	switch style {
	case cardRunning:
		bg.StrokeColor = colorRunning
		bg.FillColor = colorRunningSoft
	case cardAccent:
		bg.StrokeColor = colorAccent
		bg.FillColor = colorAccentSoft
	default:
		bg.StrokeColor = colorBorder
		bg.FillColor = colorSurface
	}
	padded := container.NewPadded(content)
	return container.NewStack(bg, padded)
}

func fluidCardTitled(title, subtitle string, content fyne.CanvasObject, style cardStyle) fyne.CanvasObject {
	var top fyne.CanvasObject
	if subtitle != "" {
		top = container.NewVBox(sectionLabel(title), mutedLabel(subtitle))
	} else if title != "" {
		top = sectionLabel(title)
	}
	inner := container.NewBorder(top, nil, nil, nil, content)
	return fluidCard(inner, style)
}

func primaryButton(label string, tapped func()) *widget.Button {
	b := widget.NewButton(label, tapped)
	b.Importance = widget.HighImportance
	return b
}

func sidebarBrand() fyne.CanvasObject {
	icon := canvas.NewImageFromResource(AppIcon())
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSize(44, 44))
	title := widget.NewLabel("Track App")
	title.TextStyle = fyne.TextStyle{Bold: true}
	tag := mutedLabel("Time per job")
	return container.NewPadded(container.NewVBox(
		container.NewCenter(icon),
		container.NewCenter(title),
		container.NewCenter(tag),
		widget.NewSeparator(),
	))
}

func navLabel(section string) string {
	switch section {
	case "Today":
		return "◷  Today"
	case "Job Timers":
		return "◔  Job Timers"
	case "Clients":
		return "◫  Clients"
	case "Projects":
		return "◆  Projects"
	case "Activity":
		return "◉  Activity"
	case "Report":
		return "▤  Report"
	case "Settings":
		return "⚙  Settings"
	default:
		return section
	}
}

func sidebarPanel(nav fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(colorSurface)
	bg.CornerRadius = 0
	bg.FillColor = color.NRGBA{R: 20, G: 28, B: 38, A: 255}
	body := container.NewBorder(sidebarBrand(), nil, nil, nil, nav)
	return container.NewStack(bg, container.NewPadded(body))
}