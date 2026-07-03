package ui

import (
	"fyne.io/fyne/v2"

	"github.com/prestonw/track-app-go/internal/app"
)

// SetupSystray configures the menu bar / system tray for the current platform.
func SetupSystray(fyneApp fyne.App, core *app.TrackApp, hud *HUD, mainWin *MainWindow) {
	setupPlatformMenuBar(fyneApp, core, hud, mainWin)
}