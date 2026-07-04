//go:build darwin

package platform

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework QuartzCore -framework ApplicationServices
void trackapp_place_hud_window(int corner, int width, int height, double margin, int animate);
*/
import "C"

import (
	"os/exec"
	"strconv"
	"strings"
)

const hudMargin = 20

type darwinWindow struct{}

func newWindowManager() windowBackend { return darwinWindow{} }

func (darwinWindow) canPlace() bool { return true }

func (darwinWindow) TrustHint() string {
	return "HUD corner snap uses native Cocoa window placement (Accessibility helps foreground tracking)."
}

func (darwinWindow) PlaceHUD(corner Corner, width, height int, animate bool) {
	anim := C.int(0)
	if animate {
		anim = 1
	}
	C.trackapp_place_hud_window(C.int(corner), C.int(width), C.int(height), C.double(hudMargin), anim)
}

func (d darwinWindow) PlaceByTitle(title string, corner Corner, width, height int) {
	d.PlaceHUD(corner, width, height, false)
}

func (darwinWindow) screenSize() (int, int) {
	out, err := exec.Command("osascript", "-e",
		`use framework "AppKit"
set vf to current application's NSScreen's mainScreen's visibleFrame()
return ((vf's |size|'s width) as integer) & "," & ((vf's |size|'s height) as integer)`,
	).Output()
	if err != nil {
		return screenSizeFallback()
	}
	parts := strings.Split(strings.TrimSpace(string(out)), ",")
	if len(parts) != 2 {
		return screenSizeFallback()
	}
	w, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	h, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
	if w <= 0 || h <= 0 {
		return screenSizeFallback()
	}
	return w, h
}

func screenSizeFallback() (int, int) {
	out, err := exec.Command("osascript", "-e",
		`tell application "Finder" to get bounds of window of desktop`,
	).Output()
	if err != nil {
		return 1920, 1080
	}
	parts := strings.Split(strings.TrimSpace(string(out)), ", ")
	if len(parts) != 4 {
		return 1920, 1080
	}
	right, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
	bottom, _ := strconv.Atoi(strings.TrimSpace(parts[3]))
	if right <= 0 || bottom <= 0 {
		return 1920, 1080
	}
	return right, bottom
}