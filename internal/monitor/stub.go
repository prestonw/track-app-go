//go:build !linux && !darwin && !windows

package monitor

import "github.com/prestonw/track-app-go/internal/models"

type stub struct{}

func newPlatform() Monitor { return stub{} }

func (stub) CurrentForeground() models.ForegroundContext {
	return models.ForegroundContext{AppName: "Unknown", WindowTitle: ""}
}

func (stub) Trusted() bool { return false }

func (stub) TrustHint() string {
	return "Foreground tracking is not available on this platform yet."
}