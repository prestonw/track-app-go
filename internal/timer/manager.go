package timer

import (
	"sync"

	"github.com/prestonw/track-app-go/internal/format"
	"github.com/prestonw/track-app-go/internal/models"
	"github.com/prestonw/track-app-go/internal/store"
)

type Manager struct {
	store    *store.Store
	onChange func()
	mu       sync.Mutex
	ticker   *timeTicker
}

func NewManager(s *store.Store, onChange func()) *Manager {
	m := &Manager{store: s, onChange: onChange}
	if anyRunning(s) {
		m.startTicking()
	}
	return m
}

func (m *Manager) notify() {
	if m.onChange != nil {
		m.onChange()
	}
}

func (m *Manager) Start(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.store.Timers {
		if m.store.Timers[i].ID != id || m.store.Timers[i].Running {
			continue
		}
		now := models.NowMs()
		m.store.Timers[i].Running = true
		m.store.Timers[i].StartedAt = &now
		m.store.Timers[i].Adjustments = append(m.store.Timers[i].Adjustments,
			models.Adjustment{Ts: now, Type: "start", Detail: "Timer started"})
		_ = m.store.SaveAll()
		m.startTicking()
		m.notify()
		return
	}
}

func (m *Manager) Pause(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.store.Timers {
		t := &m.store.Timers[i]
		if t.ID != id || !t.Running || t.StartedAt == nil {
			continue
		}
		now := models.NowMs()
		started := *t.StartedAt
		sec := int((now - started) / 1000)
		if sec < 0 {
			sec = 0
		}
		t.Elapsed += sec
		t.Running = false
		t.StartedAt = nil
		if sec > 0 {
			m.store.Sessions = append([]models.Session{{
				ID: models.MakeID(), TimerID: t.ID, Name: t.Name, Tags: t.Tags,
				Rate: t.Rate, Currency: t.Currency, Notes: t.Notes,
				Start: started, End: now, Seconds: sec, Client: m.store.ClientName(t.ClientID),
			}}, m.store.Sessions...)
			t.Adjustments = append(t.Adjustments, models.Adjustment{
				Ts: now, Type: "stop", Detail: "Stopped — " + format.HumanDuration(sec) + " logged",
			})
		} else {
			t.Adjustments = append(t.Adjustments, models.Adjustment{Ts: now, Type: "stop", Detail: "Stopped — no time logged"})
		}
		_ = m.store.SaveAll()
		m.stopTickingIfNeeded()
		m.notify()
		return
	}
}

func (m *Manager) Reset(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.store.Timers {
		t := &m.store.Timers[i]
		if t.ID != id {
			continue
		}
		now := models.NowMs()
		if t.Running && t.StartedAt != nil {
			sec := int((now - *t.StartedAt) / 1000)
			t.Elapsed += max(0, sec)
			if sec > 0 {
				m.store.Sessions = append([]models.Session{{
					ID: models.MakeID(), TimerID: t.ID, Name: t.Name, Tags: t.Tags,
					Rate: t.Rate, Currency: t.Currency, Notes: t.Notes,
					Start: *t.StartedAt, End: now, Seconds: sec, Client: m.store.ClientName(t.ClientID),
				}}, m.store.Sessions...)
			}
		}
		was := t.Elapsed
		t.Elapsed = 0
		t.Running = false
		t.StartedAt = nil
		t.Adjustments = append(t.Adjustments, models.Adjustment{
			Ts: now, Type: "reset", Detail: "Reset to zero (was " + format.HumanDuration(was) + ")",
		})
		_ = m.store.SaveAll()
		m.stopTickingIfNeeded()
		m.notify()
		return
	}
}

func (m *Manager) Toggle(id string) {
	for _, t := range m.store.Timers {
		if t.ID == id {
			if t.Running {
				m.Pause(id)
			} else {
				m.Start(id)
			}
			return
		}
	}
}

func (m *Manager) CancelRunningBurst(id string) {
	for i := range m.store.Timers {
		if m.store.Timers[i].ID == id && m.store.Timers[i].Running {
			m.store.Timers[i].Running = false
			m.store.Timers[i].StartedAt = nil
			m.store.Timers[i].Adjustments = append(m.store.Timers[i].Adjustments,
				models.Adjustment{Ts: models.NowMs(), Type: "cancel", Detail: "Auto-start skipped"})
			_ = m.store.SaveAll()
			m.stopTickingIfNeeded()
			m.notify()
			return
		}
	}
}

func (m *Manager) RestoreElapsed(id string, seconds int) {
	for i := range m.store.Timers {
		if m.store.Timers[i].ID == id {
			if seconds < 0 {
				seconds = 0
			}
			m.store.Timers[i].Elapsed = seconds
			_ = m.store.SaveAll()
			m.notify()
			return
		}
	}
}

func (m *Manager) DeleteSessionsSince(timerID string, sinceMs int64) {
	var kept []models.Session
	for _, s := range m.store.Sessions {
		if s.TimerID == timerID && s.Start >= sinceMs {
			continue
		}
		kept = append(kept, s)
	}
	m.store.Sessions = kept
	_ = m.store.SaveAll()
	m.notify()
}

func (m *Manager) startTicking() {
	if m.ticker != nil {
		return
	}
	m.ticker = newTimeTicker(func() { m.notify() })
}

func (m *Manager) stopTickingIfNeeded() {
	if anyRunning(m.store) {
		return
	}
	if m.ticker != nil {
		m.ticker.Stop()
		m.ticker = nil
	}
}

func anyRunning(s *store.Store) bool {
	for _, t := range s.Timers {
		if t.Running {
			return true
		}
	}
	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}