//go:build !systray

package main

import (
	"github.com/prestonw/track-app-go/internal/app"
	"github.com/prestonw/track-app-go/ui"
)

func runTray(_ *app.TrackApp, _ *ui.HUD, _ *ui.MainWindow) {}