//go:build darwin

package platform

import (
	"os/exec"
	"strings"

	"github.com/prestonw/track-app-go/internal/models"
)

type darwinForeground struct{}

func newForeground() foregroundBackend { return darwinForeground{} }

func (d darwinForeground) level() ForegroundLevel {
	if d.Trusted() {
		return ForegroundFull
	}
	return ForegroundBasic
}

func (darwinForeground) Trusted() bool {
	_, err := exec.Command("osascript", "-e",
		`tell application "System Events" to get name of first application process whose frontmost is true`,
	).Output()
	return err == nil
}

func (d darwinForeground) TrustHint() string {
	if d.Trusted() {
		return "Accessibility enabled — app, bundle ID, and window title tracking active."
	}
	return "Grant Accessibility to Track App in System Settings → Privacy & Security → Accessibility."
}

func (darwinForeground) CurrentForeground() models.ForegroundContext {
	out, err := exec.Command("osascript", "-e", darwinForegroundScript).Output()
	if err != nil {
		return models.ForegroundContext{}
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	ctx := models.ForegroundContext{}
	if len(lines) > 0 {
		ctx.AppName = strings.TrimSpace(lines[0])
	}
	if len(lines) > 1 {
		ctx.BundleID = strings.TrimSpace(lines[1])
	}
	if len(lines) > 2 {
		ctx.WindowTitle = strings.TrimSpace(lines[2])
	}
	if len(lines) > 3 {
		ctx.DocumentPath = strings.TrimSpace(lines[3])
	}
	if ctx.BundleID == "" {
		ctx.BundleID = ctx.AppName
	}
	return ctx
}

const darwinForegroundScript = `
tell application "System Events"
	set frontApp to first application process whose frontmost is true
	set appName to name of frontApp
	set bundleId to bundle identifier of frontApp
	set winTitle to ""
	set docPath to ""
	try
		set winTitle to name of front window of frontApp
	end try
	try
		set docPath to POSIX path of (file of front window of frontApp)
	end try
end tell
return appName & linefeed & bundleId & linefeed & winTitle & linefeed & docPath
`