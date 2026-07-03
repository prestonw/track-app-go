//go:build darwin

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const hudMargin = 20

type darwinWindow struct{}

func newWindowManager() windowBackend { return darwinWindow{} }

func (darwinWindow) canPlace() bool { return true }

func (darwinWindow) TrustHint() string {
	return "HUD corner snap uses AppleScript (Accessibility required for reliable placement)."
}

func (d darwinWindow) PlaceByTitle(title string, corner Corner, width, height int) {
	exe, _ := os.Executable()
	procName := strings.TrimSuffix(filepath.Base(exe), filepath.Ext(exe))
	script := fmt.Sprintf(`
use framework "AppKit"
use scripting additions

set cornerIdx to %d
set winW to %d
set winH to %d
set margin to %d
set titleNeedle to %q
set procNeedle to %q

set vf to current application's NSScreen's mainScreen's visibleFrame()
set vfX to (vf's origin's x) as integer
set vfY to (vf's origin's y) as integer
set vfW to (vf's |size|'s width) as integer
set vfH to (vf's |size|'s height) as integer
set screenFrame to current application's NSScreen's mainScreen's frame()
set screenH to (screenFrame's |size|'s height) as integer

set cocoaX to vfX + margin
set cocoaY to vfY + vfH - winH - margin
if cornerIdx is 1 then
	set cocoaX to vfX + vfW - winW - margin
	set cocoaY to vfY + vfH - winH - margin
else if cornerIdx is 2 then
	set cocoaX to vfX + margin
	set cocoaY to vfY + margin
else if cornerIdx is 3 then
	set cocoaX to vfX + vfW - winW - margin
	set cocoaY to vfY + margin
end if

set posX to cocoaX
set posY to screenH - cocoaY - winH

tell application "System Events"
	repeat with proc in application processes
		if background only of proc is false then
			try
				set procName to name of proc as text
				if procName contains procNeedle or procName is procNeedle then
					repeat with win in windows of proc
						set winName to name of win as text
						if winName is titleNeedle or winName contains titleNeedle or winName contains "TrackApp" or winName contains "HUD" then
							set position of win to {posX, posY}
							return
						end if
					end repeat
				end if
			end try
		end if
	end repeat
	repeat with proc in application processes
		if background only of proc is false then
			try
				repeat with win in windows of proc
					set winName to name of win as text
					if winName is titleNeedle or winName contains titleNeedle then
						set position of win to {posX, posY}
						return
					end if
				end repeat
			end try
		end if
	end repeat
end tell
`, int(corner), width, height, hudMargin, title, procName)
	_ = exec.Command("osascript", "-e", script).Run()
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