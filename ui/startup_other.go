//go:build !darwin

package ui

import (
	"fyne.io/fyne/v2"

	"github.com/prestonw/track-app-go/internal/app"
)

func ShowAccessibilityPromptIfNeeded(*app.TrackApp, fyne.Window) {}