# Track App (Go)

Cross-platform time tracker — Linux, macOS, and Windows — sharing the same SQLite schema as the original Swift Track App.

## Features

- Floating timer HUD (borderless overlay)
- Main window: Today, Job Timers, Clients, Projects, Activity, Report, Settings
- SQLite persistence (`~/.config/TrackApp/track-app.sqlite` on Linux)
- Project rules and optional auto-track (platform-dependent foreground detection)
- Optional system tray (`-tags systray`)

## Build (Linux / NixOS)

```bash
make tidy
make build
make run
```

## macOS / Windows

Install Go 1.22+ and run:

```bash
go mod tidy
go build -o trackapp .
```

On macOS, grant Accessibility permission for window title capture.

## Database

Uses the same tables as the Swift app. Copy your existing database to:

- Linux: `~/.config/TrackApp/track-app.sqlite`
- macOS: `~/Library/Application Support/TrackApp/track-app.sqlite`
- Windows: `%AppData%\TrackApp\track-app.sqlite`