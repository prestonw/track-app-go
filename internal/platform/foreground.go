package platform

import "github.com/prestonw/track-app-go/internal/models"

// ForegroundMonitor samples the active application for project rules and activity log.
type ForegroundMonitor interface {
	CurrentForeground() models.ForegroundContext
	Trusted() bool
	TrustHint() string
}