//go:build darwin

package hud

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func MoveActiveWindow(corner, width, height int) {
	sw, sh := screenSize()
	x, y := cornerXY(corner, sw, sh, width, height)
	script := fmt.Sprintf(
		`tell application "System Events" to tell (first process whose frontmost is true) to set position of front window to {%d, %d}`,
		x, y,
	)
	_ = exec.Command("osascript", "-e", script).Run()
}

func screenSize() (int, int) {
	out, err := exec.Command("osascript", "-e", `tell application "Finder" to get bounds of window of desktop`).Output()
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

func cornerXY(corner, sw, sh, width, height int) (int, int) {
	margin := 16
	x, y := margin, margin
	switch corner % 4 {
	case 1:
		x = sw - width - margin
	case 2:
		y = sh - height - margin
	case 3:
		x = sw - width - margin
		y = sh - height - margin
	}
	return x, y
}