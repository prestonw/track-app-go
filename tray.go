//go:build systray

package main

import (
	"github.com/getlantern/systray"

	"github.com/prestonw/track-app-go/internal/app"
	"github.com/prestonw/track-app-go/ui"
)

func runTray(core *app.TrackApp, hud *ui.HUD, mainWin *ui.MainWindow) {
	systray.Run(func() {
		systray.SetTitle("Track")
		systray.SetTooltip("Track App")
		mOpen := systray.AddMenuItem("Open Track App", "")
		mToday := systray.AddMenuItem("Today", "")
		mHUD := systray.AddMenuItem("Show Floating Timer", "")
		systray.AddSeparator()
		mQuit := systray.AddMenuItem("Quit", "")

		core.OnChange(func() {
			mHUD.SetTitle(hudMenuTitle(hud))
			systray.SetTooltip(core.Coordinator.StatusLine())
		})

		go func() {
			for {
				select {
				case <-mOpen.ClickedCh:
					mainWin.Show()
				case <-mToday.ClickedCh:
					mainWin.Show()
				case <-mHUD.ClickedCh:
					hud.Toggle()
				case <-mQuit.ClickedCh:
					systray.Quit()
					return
				}
			}
		}()
	}, nil)
}

func hudMenuTitle(hud *ui.HUD) string {
	if hud.Visible() {
		return "Hide Floating Timer"
	}
	return "Show Floating Timer"
}