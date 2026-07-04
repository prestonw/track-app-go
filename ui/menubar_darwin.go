//go:build darwin

package ui

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>
#include "menubar_darwin.h"
*/
import "C"
import (
	"time"
	"unsafe"

	"fyne.io/fyne/v2"

	"github.com/prestonw/track-app-go/internal/app"
)

type menuBarState struct {
	fyneApp fyne.App
	core    *app.TrackApp
	hud     *HUD
	mainWin *MainWindow
}

var activeMenuBar *menuBarState

//export trackapp_menu_open
func trackapp_menu_open() {
	if activeMenuBar != nil && activeMenuBar.mainWin != nil {
		onMain(activeMenuBar.mainWin.Show)
	}
}

//export trackapp_menu_today
func trackapp_menu_today() {
	if activeMenuBar != nil && activeMenuBar.mainWin != nil {
		onMain(func() { activeMenuBar.mainWin.OpenSection("Today") })
	}
}

//export trackapp_menu_jobs
func trackapp_menu_jobs() {
	if activeMenuBar != nil && activeMenuBar.mainWin != nil {
		onMain(func() { activeMenuBar.mainWin.OpenSection("Job Timers") })
	}
}

//export trackapp_menu_settings
func trackapp_menu_settings() {
	if activeMenuBar != nil && activeMenuBar.mainWin != nil {
		onMain(func() { activeMenuBar.mainWin.OpenSection("Settings") })
	}
}

//export trackapp_menu_toggle_hud
func trackapp_menu_toggle_hud() {
	if activeMenuBar == nil || activeMenuBar.hud == nil {
		return
	}
	onMain(func() {
		activeMenuBar.hud.Toggle()
		refreshNativeMenuBar()
	})
}

//export trackapp_menu_quit
func trackapp_menu_quit() {
	if activeMenuBar != nil && activeMenuBar.fyneApp != nil {
		onMain(activeMenuBar.fyneApp.Quit)
	}
}

func setupPlatformMenuBar(fyneApp fyne.App, core *app.TrackApp, hud *HUD, mainWin *MainWindow) {
	activeMenuBar = &menuBarState{fyneApp: fyneApp, core: core, hud: hud, mainWin: mainWin}
	installNativeMenuBar()
	// Fyne may not have finished wiring NSApplication on the first tick — retry once.
	go func() {
		time.Sleep(400 * time.Millisecond)
		onMain(installNativeMenuBar)
	}()
	core.OnChange(func() { onMain(refreshNativeMenuBar) })
}

func installNativeMenuBar() {
	if activeMenuBar == nil {
		return
	}
	C.trackapp_menubar_install()
	icon := AppIcon().Content()
	if len(icon) > 0 {
		C.trackapp_menubar_set_icon((*C.uchar)(unsafe.Pointer(&icon[0])), C.int(len(icon)))
	}
	refreshNativeMenuBar()
}

func refreshNativeMenuBar() {
	if activeMenuBar == nil || activeMenuBar.core == nil {
		return
	}
	line := C.CString(activeMenuBar.core.Coordinator.StatusLine())
	defer C.free(unsafe.Pointer(line))
	C.trackapp_menubar_set_status(line)

	hudLabel := "Show Floating Timer"
	if activeMenuBar.hud != nil && activeMenuBar.hud.Visible() {
		hudLabel = "Hide Floating Timer"
	}
	hl := C.CString(hudLabel)
	defer C.free(unsafe.Pointer(hl))
	C.trackapp_menubar_set_hud_label(hl)
}