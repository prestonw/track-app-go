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
	fyneApp.Settings().SetTheme(ui.TrackTheme())

	hud := ui.NewHUD(core, fyneApp)
	mainWin := ui.NewMainWindow(core, fyneApp, hud)

	if core.Coordinator.ShowHUDOnLaunch() {
		hud.Show()
	}
	mainWin.Show()

	go runTray(core, hud, mainWin)

	fyneApp.Run()
}