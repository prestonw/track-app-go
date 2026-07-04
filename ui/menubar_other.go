//go:build !darwin

package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"

	"github.com/prestonw/track-app-go/internal/app"
)

func setupPlatformMenuBar(fyneApp fyne.App, core *app.TrackApp, hud *HUD, mainWin *MainWindow) {
	desk, ok := fyneApp.(desktop.App)
	if !ok {
		return
	}

	var setMenu func()
	setMenu = func() {
		hudLabel := "Show floating timer"
		if hud.Visible() {
			hudLabel = "Hide floating timer"
		}
		menu := fyne.NewMenu("Track App",
			fyne.NewMenuItem("Open Track App", func() { mainWin.Show() }),
			fyne.NewMenuItem("Today", func() { mainWin.OpenSection("Today") }),
			fyne.NewMenuItem("Job Timers", func() { mainWin.OpenSection("Job Timers") }),
			fyne.NewMenuItem("Settings…", func() { mainWin.OpenSection("Settings") }),
			fyne.NewMenuItem(hudLabel, func() {
				hud.Toggle()
				setMenu()
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Quit", func() { fyneApp.Quit() }),
		)
		desk.SetSystemTrayMenu(menu)
	}
	setMenu()
	desk.SetSystemTrayIcon(AppIcon())
	core.OnChange(func() { onMain(setMenu) })
}