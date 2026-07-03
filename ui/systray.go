package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/prestonw/track-app-go/internal/app"
)

// SetupSystray configures the Fyne desktop system tray when supported.
func SetupSystray(fyneApp fyne.App, core *app.TrackApp, hud *HUD, mainWin *MainWindow) {
	desk, ok := fyneApp.(desktop.App)
	if !ok {
		return
	}

	var setMenu func()
	setMenu = func() {
		hudLabel := "Show Floating Timer"
		if hud.Visible() {
			hudLabel = "Hide Floating Timer"
		}
		menu := fyne.NewMenu("Track App",
			fyne.NewMenuItem("Open Track App", func() { mainWin.Show() }),
			fyne.NewMenuItem("Today", func() {
				mainWin.Show()
				mainWin.nav.Select(0)
			}),
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