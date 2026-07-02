//go:build linux

package hud

import (
	"fmt"
	"os/exec"
	"strconv"
)

// MoveActiveWindow places the active window in a screen corner (requires xdotool).
func MoveActiveWindow(corner, width, height int) {
	if _, err := exec.LookPath("xdotool"); err != nil {
		return
	}
	geo, err := exec.Command("xdotool", "getdisplaygeometry").Output()
	if err != nil {
		return
	}
	var sw, sh int
	fmt.Sscanf(string(geo), "%d %d", &sw, &sh)
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
	_ = exec.Command("xdotool", "getactivewindow", "windowmove", strconv.Itoa(x), strconv.Itoa(y)).Run()
}