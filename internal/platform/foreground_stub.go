//go:build !linux && !darwin && !windows

package platform

import "github.com/prestonw/track-app-go/internal/models"

type stubForeground struct{}

func newForeground() foregroundBackend { return stubForeground{} }

func (stubForeground) level() ForegroundLevel { return ForegroundNone }

func (stubForeground) Trusted() bool { return false }

func (stubForeground) TrustHint() string {
	return "Foreground tracking is not available on this platform."
}

func (stubForeground) CurrentForeground() models.ForegroundContext {
	return models.ForegroundContext{AppName: "Unknown"}
}