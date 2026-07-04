//go:build darwin

package platform

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework QuartzCore
void trackapp_main_window_borderless(void);
void trackapp_main_window_hide_animated(void);
*/
import "C"

func MainWindowBorderless() {
	C.trackapp_main_window_borderless()
}

func MainWindowHideAnimated() {
	C.trackapp_main_window_hide_animated()
}