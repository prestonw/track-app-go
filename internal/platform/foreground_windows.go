//go:build windows

package platform

import (
	"os/exec"
	"strings"

	"github.com/prestonw/track-app-go/internal/models"
)

type windowsForeground struct{}

func newForeground() foregroundBackend { return windowsForeground{} }

func (windowsForeground) level() ForegroundLevel { return ForegroundFull }

func (windowsForeground) Trusted() bool { return true }

func (windowsForeground) TrustHint() string {
	return "Win32 foreground window tracking is active."
}

func (windowsForeground) CurrentForeground() models.ForegroundContext {
	out, err := exec.Command("powershell", "-NoProfile", "-Command", windowsForegroundScript).Output()
	if err != nil {
		return models.ForegroundContext{AppName: "Windows", BundleID: "windows"}
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	ctx := models.ForegroundContext{}
	if len(lines) > 0 {
		ctx.WindowTitle = strings.TrimSpace(lines[0])
	}
	if len(lines) > 1 {
		ctx.AppName = strings.TrimSpace(lines[1])
	}
	if len(lines) > 2 {
		ctx.BundleID = strings.TrimSpace(lines[2])
	}
	if ctx.AppName == "" {
		ctx.AppName = ctx.WindowTitle
	}
	if ctx.BundleID == "" {
		ctx.BundleID = ctx.AppName
	}
	return ctx
}

const windowsForegroundScript = `
Add-Type @"
using System;
using System.Runtime.InteropServices;
using System.Text;
public class TrackFg {
  [DllImport("user32.dll")] public static extern IntPtr GetForegroundWindow();
  [DllImport("user32.dll", CharSet=CharSet.Unicode)] public static extern int GetWindowText(IntPtr hWnd, StringBuilder lpString, int nMaxCount);
  [DllImport("user32.dll")] public static extern uint GetWindowThreadProcessId(IntPtr hWnd, out uint lpdwProcessId);
}
"@
$h = [TrackFg]::GetForegroundWindow()
$sb = New-Object System.Text.StringBuilder 512
[void][TrackFg]::GetWindowText($h, $sb, 512)
$title = $sb.ToString()
$pid = [uint32]0
[void][TrackFg]::GetWindowThreadProcessId($h, [ref]$pid)
$proc = Get-Process -Id $pid -ErrorAction SilentlyContinue
$name = if ($proc) { $proc.ProcessName } else { "" }
$bundle = if ($proc) { $proc.ProcessName } else { "" }
Write-Output $title
Write-Output $name
Write-Output $bundle
`