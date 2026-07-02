package app

import (
	"github.com/prestonw/track-app-go/internal/models"
)

// DisplayContext returns HUD title/subtitle and timer state.
func (a *TrackApp) DisplayContext() (title, subtitle string, elapsed int, running bool) {
	t := a.Coordinator.FocusTimer()
	if t == nil {
		return "Track App", "Add a job to begin", 0, false
	}
	client := a.Store.ClientName(t.ClientID)
	sub := "Job timer"
	if client != "" {
		sub = client
	}
	if a.Tracker != nil {
		if p := a.Tracker.ActiveProject(); p != nil {
			sub = p.Name
			if client != "" {
				sub += " · " + client
			}
		}
	}
	return t.Name, sub, t.CurrentElapsed(models.NowMs()), t.Running
}

// StatusLine is the menu bar / tray status text.
func (a *TrackApp) StatusLine() string {
	return a.Coordinator.StatusLine()
}