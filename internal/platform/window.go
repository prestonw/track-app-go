package platform

// HUDWindowTitle is the internal window title used to find the floating timer across OS APIs.
const HUDWindowTitle = "TrackApp HUD"

// WindowManager positions overlay windows. Implementations are best-effort per OS.
type WindowManager interface {
	PlaceHUD(corner Corner, width, height int, animate bool)
	PlaceByTitle(title string, corner Corner, width, height int)
	TrustHint() string
}