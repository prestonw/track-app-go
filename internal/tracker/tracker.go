package tracker

import (
	"github.com/prestonw/track-app-go/internal/models"
	"github.com/prestonw/track-app-go/internal/store"
	"github.com/prestonw/track-app-go/internal/timer"
)

// ProjectTracker handles foreground project matches and optional auto-start.
type ProjectTracker struct {
	Store       *store.Store
	Manager     *timer.Manager
	Coordinator *timer.Coordinator
	OnChange    func()

	activeProjectID string
	autoTimerID     string
}

func New(s *store.Store, m *timer.Manager, c *timer.Coordinator, onChange func()) *ProjectTracker {
	return &ProjectTracker{Store: s, Manager: m, Coordinator: c, OnChange: onChange}
}

func (t *ProjectTracker) ActiveProject() *models.Project {
	return t.Store.Project(t.activeProjectID)
}

func (t *ProjectTracker) ActiveProjectChanged(project *models.Project) {
	newID := ""
	if project != nil {
		newID = project.ID
	}
	if newID == t.activeProjectID {
		return
	}

	if t.autoTimerID != "" {
		for i := range t.Store.Timers {
			if t.Store.Timers[i].ID == t.autoTimerID && t.Store.Timers[i].Running {
				t.Manager.Pause(t.autoTimerID)
			}
		}
	}
	t.autoTimerID = ""
	t.activeProjectID = newID

	if project == nil || project.TimerID == "" {
		t.notify()
		return
	}

	t.Coordinator.SetFocus(project.TimerID, false)

	shouldAuto := project.AutoTrack && !t.Store.IsInSkipCooldown(*project)
	if shouldAuto {
		for i := range t.Store.Timers {
			timer := &t.Store.Timers[i]
			if timer.ID == project.TimerID && !timer.Running {
				pre := timer.Elapsed
				t.Manager.Start(project.TimerID)
				t.autoTimerID = project.TimerID
				t.Coordinator.SetAutoStartPrompt(&models.AutoStartPrompt{
					ProjectID: project.ID, TimerID: project.TimerID,
					PreStartElapsed: pre, PromptedAt: models.NowMs(),
				})
				break
			}
		}
	}
	t.notify()
}

func (t *ProjectTracker) notify() {
	if t.OnChange != nil {
		t.OnChange()
	}
}