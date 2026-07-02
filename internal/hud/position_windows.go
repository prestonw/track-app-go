//go:build windows

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
	script := fmt.Sprintf(`
Add-Type @"
using System;
using System.Runtime.InteropServices;
public class HudWin32 {
  [DllImport("user32.dll")] public static extern IntPtr GetForegroundWindow();
  [DllImport("user32.dll")] public static extern bool MoveWindow(IntPtr hWnd, int X, int Y, int nWidth, int nHeight, bool bRepaint);
}
"@
$h = [HudWin32]::GetForegroundWindow()
[void][HudWin32]::MoveWindow($h, %d, %d, %d, %d, $true)
`, x, y, width, height)
	_ = exec.Command("powershell", "-NoProfile", "-Command", script).Run()
}

func screenSize() (int, int) {
	out, err := exec.Command("powershell", "-NoProfile", "-Command",
		`Add-Type -AssemblyName System.Windows.Forms; $a=[System.Windows.Forms.Screen]::PrimaryScreen.WorkingArea; "$($a.Width) $($a.Height)"`).Output()
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