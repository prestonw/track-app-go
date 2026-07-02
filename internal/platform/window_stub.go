//go:build !linux && !darwin && !windows

package platform

type stubWindow struct{}

func newWindowManager() windowBackend { return stubWindow{} }

func (stubWindow) canPlace() bool { return false }

func (stubWindow) TrustHint() string {
	return "HUD corner snap is not available on this platform."
}

func (stubWindow) PlaceByTitle(string, Corner, int, int) {}