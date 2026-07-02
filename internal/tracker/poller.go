package tracker

import (
	"time"

	"github.com/prestonw/track-app-go/internal/models"
	"github.com/prestonw/track-app-go/internal/platform"
	"github.com/prestonw/track-app-go/internal/store"
)

// Poller samples foreground context periodically (like ActivityMonitor).
type Poller struct {
	Monitor platform.ForegroundMonitor
	Store   *store.Store
	Tracker *ProjectTracker

	stop     chan struct{}
	lastCtx  models.ForegroundContext
	hasLast  bool
}

func NewPoller(mon platform.ForegroundMonitor, s *store.Store, tr *ProjectTracker) *Poller {
	return &Poller{Monitor: mon, Store: s, Tracker: tr, stop: make(chan struct{})}
}

func (p *Poller) Start() {
	go func() {
		tick := time.NewTicker(2 * time.Second)
		defer tick.Stop()
		p.sample()
		for {
			select {
			case <-tick.C:
				p.sample()
			case <-p.stop:
				return
			}
		}
	}()
}

func (p *Poller) Stop() {
	select {
	case <-p.stop:
	default:
		close(p.stop)
	}
}

func (p *Poller) sample() {
	ctx := p.Monitor.CurrentForeground()
	if p.hasLast && ctx == p.lastCtx {
		return
	}
	p.lastCtx = ctx
	p.hasLast = true

	project := p.Store.MatchProject(ctx)
	pid := ""
	if project != nil {
		pid = project.ID
	}
	p.Store.StartActivitySegment(ctx, pid)
	p.Tracker.ActiveProjectChanged(project)
}