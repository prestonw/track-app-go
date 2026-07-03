//go:build linux

package platform

import (
	"fmt"
	"os/exec"
	"strconv"
)

type linuxWindow struct{}

func newWindowManager() windowBackend { return linuxWindow{} }

func (linuxWindow) canPlace() bool {
	_, err := exec.LookPath("xdotool")
	return err == nil
}

func (w linuxWindow) TrustHint() string {
	if w.canPlace() {
		return "HUD corner snap uses xdotool."
	}
	return "Install xdotool for HUD corner snap on X11."
}

func (w linuxWindow) PlaceHUD(corner Corner, width, height int, _ bool) {
	w.PlaceByTitle(HUDWindowTitle, corner, width, height)
}

func (w linuxWindow) PlaceByTitle(title string, corner Corner, width, height int) {
	if !w.canPlace() {
		return
	}
	// Search by window name, fall back to active window after HUD is shown.
	idOut, err := exec.Command("xdotool", "search", "--name", title).Output()
	if err == nil && len(idOut) > 0 {
		winID := string(idOut)
		if idx := indexByte(winID, '\n'); idx >= 0 {
			winID = winID[:idx]
		}
		placeWindowID(winID, corner, width, height)
		return
	}
	placeActiveWindow(corner, width, height)
}

func placeActiveWindow(corner Corner, width, height int) {
	geo, err := exec.Command("xdotool", "getdisplaygeometry").Output()
	if err != nil {
		return
	}
	var sw, sh int
	fmt.Sscanf(string(geo), "%d %d", &sw, &sh)
	x, y := Origin(corner, sw, sh, width, height, 20)
	_ = exec.Command("xdotool", "getactivewindow", "windowmove", strconv.Itoa(x), strconv.Itoa(y)).Run()
}

func placeWindowID(winID string, corner Corner, width, height int) {
	geo, err := exec.Command("xdotool", "getdisplaygeometry").Output()
	if err != nil {
		return
	}
	var sw, sh int
	fmt.Sscanf(string(geo), "%d %d", &sw, &sh)
	x, y := Origin(corner, sw, sh, width, height, 20)
	_ = exec.Command("xdotool", "windowmove", winID, strconv.Itoa(x), strconv.Itoa(y)).Run()
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}