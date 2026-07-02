//go:build darwin

package monitor

import (
	"os/exec"
	"strings"

	"github.com/prestonw/track-app-go/internal/models"
)

type darwinMonitor struct{}

func newPlatform() Monitor { return darwinMonitor{} }

func (darwinMonitor) Trusted() bool { return true }

func (darwinMonitor) TrustHint() string {
	return "Enable Accessibility in System Settings for full window title tracking."
}

func (darwinMonitor) CurrentForeground() models.ForegroundContext {
	ctx := models.ForegroundContext{}
	out, err := exec.Command("osascript", "-e",
		`tell application "System Events" to get name of first application process whose frontmost is true`).Output()
	if err == nil {
		ctx.AppName = strings.TrimSpace(string(out))
		ctx.BundleID = ctx.AppName
	}
	return ctx
}