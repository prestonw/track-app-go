package monitor

import "github.com/prestonw/track-app-go/internal/models"

// Monitor samples foreground application context for auto-tracking rules.
type Monitor interface {
	CurrentForeground() models.ForegroundContext
	Trusted() bool
	TrustHint() string
}

func New() Monitor {
	return newPlatform()
}