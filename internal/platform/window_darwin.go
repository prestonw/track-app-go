//go:build darwin

package platform

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type darwinWindow struct{}

func newWindowManager() windowBackend { return darwinWindow{} }

func (darwinWindow) canPlace() bool { return true }

func (darwinWindow) TrustHint() string {
	return "HUD corner snap uses AppleScript (Accessibility required for reliable placement)."
}

func (d darwinWindow) PlaceByTitle(title string, corner Corner, width, height int) {
	sw, sh := d.screenSize()
	x, y := Origin(corner, sw, sh, width, height, 16)
	script := fmt.Sprintf(`
tell application "System Events"
	repeat with proc in application processes
		try
			repeat with win in windows of proc
				if name of win is %q then
					set position of win to {%d, %d}
					return
				end if
			end repeat
		end try
	end repeat
end tell
`, title, x, y)
	_ = exec.Command("osascript", "-e", script).Run()
}

func (darwinWindow) screenSize() (int, int) {
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