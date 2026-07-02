package platform

import "testing"

func TestCornerOrigin(t *testing.T) {
	x, y := Origin(TopRight, 1920, 1080, 280, 108, 16)
	if x != 1920-280-16 || y != 16 {
		t.Fatalf("top-right: got %d,%d", x, y)
	}
	x, y = Origin(BottomRight, 1920, 1080, 280, 108, 16)
	if x != 1920-280-16 || y != 1080-108-16 {
		t.Fatalf("bottom-right: got %d,%d", x, y)
	}
}

func TestCornerNext(t *testing.T) {
	if TopLeft.Next() != TopRight {
		t.Fatal("corner cycle broken")
	}
	if BottomRight.Next() != TopLeft {
		t.Fatal("corner wrap broken")
	}
}