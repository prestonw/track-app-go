package platform

import "runtime"

// OS identifies the host operating system.
type OS string

const (
	OSDarwin  OS = "darwin"
	OSLinux   OS = "linux"
	OSWindows OS = "windows"
	OSUnknown OS = "unknown"
)

// ForegroundLevel describes how much context the OS integration can read.
type ForegroundLevel int

const (
	ForegroundNone ForegroundLevel = iota
	ForegroundBasic                 // app / process name
	ForegroundFull                  // app + window title + document path
)

// Capabilities reports what the current platform backend supports.
type Capabilities struct {
	OS                 OS
	GoOS               string
	Foreground         ForegroundLevel
	WindowCornerSnap   bool
	SystemTray         bool
	ForegroundTrusted  bool
	ForegroundHint     string
	WindowHint         string
}

func detectOS() OS {
	switch runtime.GOOS {
	case "darwin":
		return OSDarwin
	case "linux":
		return OSLinux
	case "windows":
		return OSWindows
	default:
		return OSUnknown
	}
}