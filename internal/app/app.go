package app

import (
	"sync"

	"github.com/prestonw/track-app-go/internal/monitor"
	"github.com/prestonw/track-app-go/internal/store"
	"github.com/prestonw/track-app-go/internal/timer"
)

// TrackApp is the shared application core used by UI and systray.
type TrackApp struct {
	Store       *store.Store
	Manager     *timer.Manager
	Coordinator *timer.Coordinator
	Monitor     monitor.Monitor

	mu        sync.RWMutex
	listeners []func()
}

func New() (*TrackApp, error) {
	s, err := store.Open("")
	if err != nil {
		return nil, err
	}
	core := &TrackApp{Store: s, Monitor: monitor.New()}
	core.Manager = timer.NewManager(s, func() { core.Notify() })
	core.Coordinator = timer.NewCoordinator(s, core.Manager)
	return core, nil
}

func (a *TrackApp) OnChange(fn func()) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.listeners = append(a.listeners, fn)
}

func (a *TrackApp) Notify() {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for _, fn := range a.listeners {
		fn()
	}
}

func (a *TrackApp) Close() error {
	return a.Store.Close()
}