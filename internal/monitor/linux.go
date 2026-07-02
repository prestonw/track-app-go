//go:build linux

package monitor

import (
	"os"
	"os/exec"
	"strings"

	"github.com/prestonw/track-app-go/internal/models"
)

type linuxMonitor struct{}

func newPlatform() Monitor { return linuxMonitor{} }

func (linuxMonitor) Trusted() bool { return commandExists("xdotool") || commandExists("xprop") }

func (linuxMonitor) TrustHint() string {
	if commandExists("xdotool") {
		return "X11 tools detected — basic window title tracking available."
	}
	return "Install xdotool or xprop for window title tracking on X11."
}

func (linuxMonitor) CurrentForeground() models.ForegroundContext {
	ctx := models.ForegroundContext{}
	if commandExists("xdotool") {
		out, err := exec.Command("xdotool", "getactivewindow", "getwindowname").Output()
		if err == nil {
			ctx.WindowTitle = strings.TrimSpace(string(out))
			ctx.AppName = ctx.WindowTitle
		}
		pidOut, err := exec.Command("xdotool", "getactivewindow", "getwindowpid").Output()
		if err == nil {
			pid := strings.TrimSpace(string(pidOut))
			comm, err := os.ReadFile("/proc/" + pid + "/comm")
			if err == nil {
				ctx.AppName = strings.TrimSpace(string(comm))
				ctx.BundleID = ctx.AppName
			}
		}
	}
	return ctx
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}