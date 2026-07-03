package main

import (
	"log"

	fyneapp "fyne.io/fyne/v2/app"

	"github.com/prestonw/track-app-go/internal/app"
	"github.com/prestonw/track-app-go/ui"
)

func main() {
	core, err := app.New()
	if err != nil {
		log.Fatal(err)
	}
	defer core.Close()

	fyneApp := fyneapp.NewWithID("com.prestonw.trackapp")
	fyneApp.SetIcon(ui.AppIcon())
	fyneApp.Settings().SetTheme(ui.TrackTheme())

	hud := ui.NewHUD(core, fyneApp)
	mainWin := ui.NewMainWindow(core, fyneApp, hud)

	fyneApp.Lifecycle().SetOnStarted(func() {
		ui.SetupSystray(fyneApp, core, hud, mainWin)
		if core.Coordinator.ShowHUDOnLaunch() {
			hud.Show()
		}
		if core.Coordinator.NeedsOnboarding() {
			ui.ShowOnboardingIfNeeded(core, fyneApp, mainWin, hud)
		}
	})

	fyneApp.Run()
}