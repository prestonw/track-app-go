# Track App Go — session memory

**Last updated:** 2026-07-02  
**Repo:** `/home/will/track-app-go`  
**Swift reference:** `/home/will/track-app` (macOS native, same SQLite schema)  
**Primary test target:** macOS (build on Mac; Linux used for dev/CGO-free logic only)

---

## Resume here

1. Read this file.
2. On **Mac**:
   ```bash
   cd track-app-go
   go mod tidy
   go build -o trackapp .
   ./trackapp
   ```
3. Grant **Accessibility** (macOS) for foreground rules + HUD corner snap.
4. First run shows **onboarding** if no job timers exist (`prefs.json` → `onboardingComplete`).
5. App icon + tray icon: bundled `ui/embedded/icon.png` (from Swift `AppIcon`).

**Next work:** polish onboarding on Mac, always-on-top HUD, menu-bar native app (macOS), screenshots phase.

---

## What this project is

Cross-platform **Go + Fyne** rewrite of Track App — one codebase for macOS, Linux, Windows.

- **Bundle / app ID:** `com.prestonw.trackapp`
- **Stack:** Go 1.25, Fyne v2.7, modernc SQLite
- **DB:** same schema as Swift app (portable between apps)

| OS | DB path |
|----|---------|
| macOS | `~/Library/Application Support/TrackApp/track-app.sqlite` |
| Linux | `~/.config/TrackApp/track-app.sqlite` |
| Windows | `%AppData%\TrackApp\track-app.sqlite` |

---

## Architecture

```
internal/platform/     # OS backends (foreground + HUD window placement)
internal/store/        # SQLite CRUD, import/export
internal/timer/        # Manager + Coordinator (focus, prefs, onboarding)
internal/tracker/      # Auto-track poller + project matcher
ui/                    # Fyne UI, fluid theme, onboarding
```

**Do not** import OS packages from UI — use `app.Platform`.

### Platform backends

| OS | Foreground | HUD snap |
|----|------------|----------|
| darwin | AppleScript / System Events | AppleScript by window title |
| linux | xdotool + /proc | xdotool search |
| windows | Win32 via PowerShell | FindWindow + MoveWindow |

---

## UI (fluid style)

- **Theme:** `ui/theme.go` + `ui/colors.go` (teal accent `#5BC0BE`, dark navy bg)
- **Layout helpers:** `ui/style.go` — `pageChrome`, `fluidCard`, `sidebarPanel`, `primaryButton`
- **Icon:** `ui/assets.go` embeds `ui/embedded/icon.png`
- **Onboarding:** `ui/onboarding.go` — 5 steps; skip if timers already exist
- **Sections:** sidebar nav with glyphs; all pages use `pageChrome`

### HUD

- Fyne `CreateSplashWindow` (borderless on desktop)
- Internal title: `TrackApp HUD` (for platform window placement)
- Rounded card background with accent border

---

## Build

```bash
# Linux dev (needs display for GUI)
make build && make run

# macOS (required for release binaries — Fyne/CGO)
go build -o trackapp .

# Tests (no display needed)
make test
```

Cross-compile macOS **from Linux fails** (no Xcode toolchain). Build on Mac.

---

## Git / commits (recent)

| Commit | Summary |
|--------|---------|
| `c357f9e` | `internal/platform` multiplatform refactor |
| `da531e7` | CSV/SQLite import, edit timer, HUD platform snap |
| `8014516` | Auto-track, report UI, Fyne systray |
| `30406bf` | Initial Go scaffold |

**Latest:** pushed to **https://github.com/prestonw/track-app-go** (`main`). Remote uses HTTPS + `gh` credentials.

---

## Feature parity vs Swift

| Feature | Go | Swift |
|---------|----|-------|
| Floating HUD | ✓ | ✓ |
| Auto-track + Start/Skip | ✓ | ✓ |
| Projects + rules | ✓ | ✓ |
| Activity log | ✓ | ✓ |
| Report + CSV/SQLite IO | ✓ | ✓ |
| Onboarding | ✓ | — |
| Menu bar (macOS) | tray only | native |
| Always-on-top HUD | partial | ✓ |
| Screenshots | — | planned |

---

## Prefs (`prefs.json` next to DB)

```json
{
  "showHudOnLaunch": true,
  "hudCorner": 0,
  "onboardingComplete": false,
  "primaryTimerId": "...",
  "recentTimerIds": []
}
```

Reset onboarding: Settings → "Show onboarding again", or delete `onboardingComplete` / clear timers.

---

## Common fixes

- **No app icon:** ensure `ui/embedded/icon.png` exists; `go:embed` must be path inside `ui/` package.
- **Foreground empty on Mac:** grant Accessibility; Settings → Refresh platform status.
- **HUD wrong corner:** click ◢ in HUD; uses `platform.Window().PlaceByTitle`.
- **Onboarding loops:** `SetOnboardingComplete(true)` on finish; skipped when `len(timers) > 0`.