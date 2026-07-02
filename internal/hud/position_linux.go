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
	x, y := cornerXY(corner, sw, sh, width, height)
	_ = exec.Command("xdotool", "getactivewindow", "windowmove", strconv.Itoa(x), strconv.Itoa(y)).Run()
}