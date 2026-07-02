//go:build windows

package monitor

import "github.com/prestonw/track-app-go/internal/models"

type winMonitor struct{}

func newPlatform() Monitor { return winMonitor{} }

func (winMonitor) Trusted() bool { return true }

func (winMonitor) TrustHint() string {
	return "Window tracking uses Win32 foreground APIs."
}

func (winMonitor) CurrentForeground() models.ForegroundContext {
	// Extended Win32 implementation can be added; manual rules still work.
	return models.ForegroundContext{AppName: "Windows", BundleID: "windows"}
}