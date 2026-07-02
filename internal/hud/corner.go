//go:build linux

package hud

func cornerXY(corner, sw, sh, width, height int) (int, int) {
	margin := 16
	x, y := margin, margin
	switch corner % 4 {
	case 1:
		x = sw - width - margin
	case 2:
		y = sh - height - margin
	case 3:
		x = sw - width - margin
		y = sh - height - margin
	}
	return x, y
}