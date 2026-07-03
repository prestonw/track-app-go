//go:build darwin

package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	"github.com/prestonw/track-app-go/internal/app"
	"github.com/prestonw/track-app-go/internal/platform"
)

// ShowAccessibilityPromptIfNeeded reminds the user to grant Accessibility on macOS.
func ShowAccessibilityPromptIfNeeded(core *app.TrackApp, parent fyne.Window) {
	cap := core.Platform.Capabilities()
	if cap.OS != platform.OSDarwin || cap.ForegroundTrusted {
		return
	}
	dialog.ShowInformation(
		"Enable Accessibility",
		"Track App needs Accessibility permission for the floating timer to snap to screen corners and for project auto-tracking.\n\n"+
			"Open System Settings → Privacy & Security → Accessibility, then turn on Track App (or trackapp).\n\n"+
			"After enabling, use Settings → Refresh platform status.",
		parent,
	)
}