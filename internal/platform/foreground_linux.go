//go:build linux

package platform

import (
	"os"
	"os/exec"
	"strings"

	"github.com/prestonw/track-app-go/internal/models"
)

type linuxForeground struct{}

func newForeground() foregroundBackend { return linuxForeground{} }

func (linuxForeground) level() ForegroundLevel {
	if commandExists("xdotool") {
		return ForegroundFull
	}
	if commandExists("xprop") {
		return ForegroundBasic
	}
	return ForegroundNone
}

func (linuxForeground) Trusted() bool {
	return commandExists("xdotool") || commandExists("xprop")
}

func (linuxForeground) TrustHint() string {
	if commandExists("xdotool") {
		return "X11 tools detected — window title and process tracking available."
	}
	return "Install xdotool (recommended) or xprop for foreground tracking on X11/Wayland XWayland."
}

func (linuxForeground) CurrentForeground() models.ForegroundContext {
	ctx := models.ForegroundContext{}
	if commandExists("xdotool") {
		out, err := exec.Command("xdotool", "getactivewindow", "getwindowname").Output()
		if err == nil {
			ctx.WindowTitle = strings.TrimSpace(string(out))
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
		if ctx.AppName == "" {
			ctx.AppName = ctx.WindowTitle
		}
	}
	return ctx
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}