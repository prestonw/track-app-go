//go:build windows

package platform

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type windowsWindow struct{}

func newWindowManager() windowBackend { return windowsWindow{} }

func (windowsWindow) canPlace() bool { return true }

func (windowsWindow) TrustHint() string {
	return "HUD corner snap uses Win32 MoveWindow on the foreground window."
}

func (w windowsWindow) PlaceHUD(corner Corner, width, height int, _ bool) {
	w.PlaceByTitle(HUDWindowTitle, corner, width, height)
}

func (w windowsWindow) PlaceByTitle(title string, corner Corner, width, height int) {
	sw, sh := w.screenSize()
	x, y := Origin(corner, sw, sh, width, height, 20)
	script := fmt.Sprintf(`
Add-Type @"
using System;
using System.Runtime.InteropServices;
using System.Text;
public class TrackHud {
  [DllImport("user32.dll")] public static extern IntPtr FindWindow(string lpClassName, string lpWindowName);
  [DllImport("user32.dll")] public static extern bool MoveWindow(IntPtr hWnd, int X, int Y, int nWidth, int nHeight, bool bRepaint);
}
"@
$h = [TrackHud]::FindWindow($null, %q)
if ($h -eq [IntPtr]::Zero) { $h = [System.Diagnostics.Process]::GetCurrentProcess().MainWindowHandle }
[void][TrackHud]::MoveWindow($h, %d, %d, %d, %d, $true)
`, title, x, y, width, height)
	_ = exec.Command("powershell", "-NoProfile", "-Command", script).Run()
}

func (windowsWindow) screenSize() (int, int) {
	out, err := exec.Command("powershell", "-NoProfile", "-Command",
		`Add-Type -AssemblyName System.Windows.Forms; $a=[System.Windows.Forms.Screen]::PrimaryScreen.WorkingArea; "$($a.Width) $($a.Height)"`,
	).Output()
	if err != nil {
		return 1920, 1080
	}
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) != 2 {
		return 1920, 1080
	}
	w, _ := strconv.Atoi(parts[0])
	h, _ := strconv.Atoi(parts[1])
	if w <= 0 || h <= 0 {
		return 1920, 1080
	}
	return w, h
}