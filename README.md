# Track App (Go)

Cross-platform time tracker for Linux, macOS, and Windows — same SQLite schema as the Swift Track App.

## Features

- **Floating timer HUD** — compact overlay with job picker, play/pause, reset, corner snap (Linux via xdotool), and auto-start prompt banner
- **Main window** — Today, Job Timers, Clients, Projects (rules + auto-track), Activity, Report, Settings
- **Auto-track** — foreground polling matches project rules, switches focus timer, optional Start/Skip prompt (60s grace, 15min cooldown)
- **Report** — per-job totals, session list with filters, CSV export, SQLite backup
- **System tray** — Fyne desktop tray on supported platforms (Open, Today, Show/Hide HUD, Quit)
- **SQLite** — `~/.config/TrackApp/track-app.sqlite` on Linux (compatible with Swift app DB)

## Build (Linux / NixOS)

```bash
make tidy
make build
make run
```

Requires a display for the GUI. On Linux, install `xdotool` for HUD corner positioning and foreground window detection.

## macOS / Windows

```bash
go mod tidy
go build -o trackapp .
```

On macOS, grant Accessibility permission for window title capture.

## Database paths

| OS      | Path |
|---------|------|
| Linux   | `~/.config/TrackApp/track-app.sqlite` |
| macOS   | `~/Library/Application Support/TrackApp/track-app.sqlite` |
| Windows | `%AppData%\TrackApp\track-app.sqlite` |

Copy an existing Swift Track App database to the path above to keep your data.

## Parity vs Swift app

| Feature | Go | Swift |
|---------|----|-------|
| Floating HUD | ✓ | ✓ |
| Auto-track + Start/Skip | ✓ | ✓ |
| Project rules | ✓ | ✓ |
| Activity log | ✓ | ✓ |
| Report + CSV export | ✓ | ✓ |
| CSV import | — | ✓ |
| Always-on-top HUD | partial | ✓ |
| Menu bar (macOS) | tray | ✓ |