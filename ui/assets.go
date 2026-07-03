package ui

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed embedded/icon.png
var iconPNG []byte

// AppIcon returns the bundled Track App stopwatch icon.
func AppIcon() fyne.Resource {
	return fyne.NewStaticResource("icon.png", iconPNG)
}