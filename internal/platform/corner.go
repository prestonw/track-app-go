package platform

// Corner identifies which screen corner the HUD is snapped to.
// Order matches the Swift app: top-left → top-right → bottom-left → bottom-right.
type Corner int

const (
	TopLeft Corner = iota
	TopRight
	BottomLeft
	BottomRight
)

func (c Corner) Next() Corner {
	return Corner((int(c) + 1) % 4)
}

// Origin returns the top-left window position for a corner snap.
func Origin(corner Corner, screenW, screenH, windowW, windowH, margin int) (x, y int) {
	if margin <= 0 {
		margin = 16
	}
	x, y = margin, margin
	switch corner {
	case TopRight:
		x = screenW - windowW - margin
	case BottomLeft:
		y = screenH - windowH - margin
	case BottomRight:
		x = screenW - windowW - margin
		y = screenH - windowH - margin
	}
	return x, y
}

// CornerFromInt maps persisted prefs (0–3) to a Corner.
func CornerFromInt(v int) Corner {
	return Corner(v % 4)
}