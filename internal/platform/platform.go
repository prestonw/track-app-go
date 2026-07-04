package platform

import "runtime"

// Default is the host platform integration used by the app core.
type Default struct {
	Cap        Capabilities
	foreground foregroundBackend
	window     windowBackend
}

type foregroundBackend interface {
	ForegroundMonitor
	level() ForegroundLevel
}

type windowBackend interface {
	WindowManager
	canPlace() bool
}

// New builds platform backends for the OS this binary was compiled for.
func New() *Default {
	fg := newForeground()
	win := newWindowManager()
	cap := Capabilities{
		OS:                detectOS(),
		GoOS:              runtime.GOOS,
		Foreground:        fg.level(),
		WindowCornerSnap:  win.canPlace(),
		ForegroundTrusted: fg.Trusted(),
		ForegroundHint:    fg.TrustHint(),
		WindowHint:        win.TrustHint(),
	}
	switch cap.OS {
	case OSLinux, OSWindows:
		cap.SystemTray = true
	case OSDarwin:
		cap.SystemTray = false // menu bar only on macOS
	}
	return &Default{Cap: cap, foreground: fg, window: win}
}

// Foreground returns the foreground monitor backend.
func (p *Default) Foreground() ForegroundMonitor { return p.foreground }

// Window returns the overlay window manager backend.
func (p *Default) Window() WindowManager { return p.window }

// Capabilities returns a snapshot of platform support.
func (p *Default) Capabilities() Capabilities { return p.Cap }

// RefreshCapabilities re-probes trust state (e.g. after user grants Accessibility).
func (p *Default) RefreshCapabilities() Capabilities {
	p.Cap.ForegroundTrusted = p.foreground.Trusted()
	p.Cap.ForegroundHint = p.foreground.TrustHint()
	p.Cap.WindowCornerSnap = p.window.canPlace()
	p.Cap.WindowHint = p.window.TrustHint()
	return p.Cap
}