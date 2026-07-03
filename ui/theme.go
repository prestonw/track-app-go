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
		return colorBG
	case theme.ColorNameButton:
		return colorPrimary
	case theme.ColorNameDisabledButton:
		return colorSurfaceAlt
	case theme.ColorNameForeground:
		return colorText
	case theme.ColorNameInputBackground:
		return colorSurface
	case theme.ColorNamePlaceHolder:
		return colorTextDim
	case theme.ColorNamePrimary:
		return colorAccent
	case theme.ColorNameHover:
		return colorSurfaceAlt
	case theme.ColorNameSelection:
		return colorAccentSoft
	case theme.ColorNameScrollBar:
		return colorBorder
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 100}
	case theme.ColorNameHeaderBackground:
		return colorSurface
	case theme.ColorNameMenuBackground:
		return colorSurface
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 8, G: 12, B: 18, A: 220}
	case theme.ColorNameSeparator:
		return colorBorder
	case theme.ColorNameDisabled:
		return colorTextDim
	case theme.ColorNameInputBorder:
		return colorBorder
	case theme.ColorNameFocus:
		return colorAccent
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
		return 12
	case theme.SizeNameInnerPadding:
		return 10
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 22
	case theme.SizeNameCaptionText:
		return 12
	case theme.SizeNameInlineIcon:
		return 18
	}
	return theme.DefaultTheme().Size(name)
}