# Track App (Go)

Cross-platform time tracker for Linux, macOS, and Windows — same SQLite schema as the Swift Track App.

## Architecture

Platform-specific code lives behind `internal/platform`. Everything else (store, timers, tracker, UI) is OS-agnostic.

```
internal/platform/
  platform.go          # Default bundle: capabilities + backends
  foreground_*.go      # ForegroundMonitor (build-tagged per OS)
  window_*.go          # WindowManager for HUD corner snap
  corner.go            # Shared corner math (all platforms)
```

| Backend | Foreground | HUD snap |
|---------|------------|----------|
| **macOS** | AppleScript + System Events (Accessibility) | AppleScript by window title |
| **Linux** | xdotool / proc | xdotool search by title |
| **Windows** | Win32 via PowerShell | FindWindow + MoveWindow |

`app.TrackApp` holds `Platform *platform.Default`. UI and tracker never import OS packages directly.

## Features

- Floating timer HUD (Fyne splash window)
- Main window: Today, Job Timers, Clients, Projects, Activity, Report, Settings
- Auto-track with foreground polling, Start/Skip prompt
- CSV + SQLite import/export
- Fyne desktop system tray
- SQLite: `~/.config/TrackApp/track-app.sqlite` on Linux

## Build

### Linux (NixOS / dev)

```bash
make tidy
make build
make run
```

### macOS (primary test target)

```bash
go mod tidy
go build -o trackapp .
./trackapp
```

Grant **Accessibility** in System Settings → Privacy & Security → Accessibility so foreground rules and HUD snap work.

Cross-compile from Linux:

```bash
GOOS=darwin GOARCH=arm64 go build -o trackapp-darwin-arm64 .
GOOS=darwin GOARCH=amd64 go build -o trackapp-darwin-amd64 .
```

### Windows

```bash
go build -o trackapp.exe .
```

## Database paths

| OS      | Path |
|---------|------|
| Linux   | `~/.config/TrackApp/track-app.sqlite` |
| macOS   | `~/Library/Application Support/TrackApp/track-app.sqlite` |
| Windows | `%AppData%\TrackApp\track-app.sqlite` |

Copy an existing Swift Track App database to migrate data.

## Adding a new platform

1. Add `foreground_<os>.go` and `window_<os>.go` with `//go:build <os>` implementing the two interfaces.
2. Extend `detectOS()` in `capabilities.go` if needed.
3. Run `go test ./internal/platform/...` — corner math tests are OS-independent.
4. No UI changes required if capabilities are reported correctly.