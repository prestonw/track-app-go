package ui

import (
	"fyne.io/fyne/v2"

	"github.com/prestonw/track-app-go/internal/app"
)

// SetupSystray configures the menu bar (macOS) or system tray (other platforms).
func SetupSystray(fyneApp fyne.App, core *app.TrackApp, hud *HUD, mainWin *MainWindow) {
	setupPlatformMenuBar(fyneApp, core, hud, mainWin)
}