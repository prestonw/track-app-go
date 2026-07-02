package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type trackTheme struct{}

var _ fyne.Theme = (*trackTheme)(nil)

func (trackTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 22, G: 24, B: 28, A: 255}
	case theme.ColorNameButton:
		return color.NRGBA{R: 45, G: 110, B: 245, A: 255}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 55, G: 58, B: 64, A: 255}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 230, G: 232, B: 235, A: 255}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 32, G: 35, B: 40, A: 255}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 120, G: 125, B: 135, A: 255}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 72, G: 140, B: 255, A: 255}
	case theme.ColorNameHover:
		return color.NRGBA{R: 40, G: 44, B: 50, A: 255}
	case theme.ColorNameSelection:
		return color.NRGBA{R: 45, G: 90, B: 180, A: 120}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 60, G: 64, B: 72, A: 180}
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 80}
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (t trackTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t trackTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t trackTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInnerPadding:
		return 6
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 20
	}
	return theme.DefaultTheme().Size(name)
}