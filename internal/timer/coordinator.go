package timer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/prestonw/track-app-go/internal/format"
	"github.com/prestonw/track-app-go/internal/models"
	"github.com/prestonw/track-app-go/internal/store"
)

type Coordinator struct {
	store   *store.Store
	manager *Manager
	prefs   *prefsFile
	mu      sync.RWMutex
}

type prefsFile struct {
	path string
}

func NewCoordinator(s *store.Store, m *Manager) *Coordinator {
	dir := filepath.Dir(store.DefaultPath())
	return &Coordinator{
		store:   s,
		manager: m,
		prefs:   &prefsFile{path: filepath.Join(dir, "prefs.json")},
	}
}

type prefsData struct {
	PrimaryTimerID string                 `json:"primaryTimerId"`
	RecentTimerIDs []string               `json:"recentTimerIds"`
	AutoStart      *models.AutoStartPrompt `json:"autoStartPrompt,omitempty"`
	ShowHUDOnLaunch     *bool `json:"showHudOnLaunch,omitempty"`
	HUDCorner           int  `json:"hudCorner"`
	HUDHintDismissed    bool  `json:"hudHintDismissed"`
	OnboardingComplete  bool  `json:"onboardingComplete"`
}

func (c *Coordinator) loadPrefs() prefsData {
	var p prefsData
	showHUD := true
	p.ShowHUDOnLaunch = &showHUD
	p.HUDCorner = 1 // top-right default
	b, err := os.ReadFile(c.prefs.path)
	if err != nil {
		return p
	}
	_ = json.Unmarshal(b, &p)
	return p
}

func (c *Coordinator) savePrefs(p prefsData) {
	b, _ := json.MarshalIndent(p, "", "  ")
	_ = os.WriteFile(c.prefs.path, b, 0o644)
}

func (c *Coordinator) FocusTimer() *models.JobTimer {
	for i := range c.store.Timers {
		if c.store.Timers[i].Running {
			return &c.store.Timers[i]
		}
	}
	p := c.loadPrefs()
	if p.PrimaryTimerID != "" {
		for i := range c.store.Timers {
			if c.store.Timers[i].ID == p.PrimaryTimerID {
				return &c.store.Timers[i]
			}
		}
	}
	if len(c.store.Timers) > 0 {
		return &c.store.Timers[0]
	}
	return nil
}

func (c *Coordinator) IsRunning() bool {
	for _, t := range c.store.Timers {
		if t.Running {
			return true
		}
	}
	return false
}

func (c *Coordinator) SetFocus(timerID string, start bool) {
	p := c.loadPrefs()
	for _, t := range c.store.Timers {
		if t.Running && t.ID != timerID {
			c.manager.Pause(t.ID)
		}
	}
	p.PrimaryTimerID = timerID
	ids := p.RecentTimerIDs
	var filtered []string
	for _, id := range ids {
		if id != timerID {
			filtered = append(filtered, id)
		}
	}
	p.RecentTimerIDs = append([]string{timerID}, filtered...)
	if len(p.RecentTimerIDs) > 8 {
		p.RecentTimerIDs = p.RecentTimerIDs[:8]
	}
	c.savePrefs(p)
	if start {
		c.manager.Start(timerID)
	}
}

func (c *Coordinator) RecentTimerIDs(limit int) []string {
	p := c.loadPrefs()
	if limit <= 0 || limit > len(p.RecentTimerIDs) {
		return p.RecentTimerIDs
	}
	return p.RecentTimerIDs[:limit]
}

func (c *Coordinator) ToggleFocus() bool {
	t := c.FocusTimer()
	if t == nil {
		return false
	}
	c.SetFocus(t.ID, false)
	c.manager.Toggle(t.ID)
	return true
}

func (c *Coordinator) StopFocus() {
	for _, t := range c.store.Timers {
		if t.Running {
			c.manager.Pause(t.ID)
		}
	}
}

func (c *Coordinator) ResetFocus() {
	t := c.FocusTimer()
	if t == nil {
		return
	}
	c.SetFocus(t.ID, false)
	c.manager.Reset(t.ID)
}

func (c *Coordinator) StatusLine() string {
	t := c.FocusTimer()
	if t == nil {
		return "No job yet — add one from the menu"
	}
	client := c.store.ClientName(t.ClientID)
	prefix := ""
	if client != "" {
		prefix = client + " · "
	}
	mark := "⏸"
	if t.Running {
		mark = "▶"
	}
	return mark + " " + prefix + t.Name + " — " + format.Duration(t.CurrentElapsed(models.NowMs()))
}

func (c *Coordinator) DisplayContext() (title, subtitle string, elapsed int, running bool) {
	t := c.FocusTimer()
	if t == nil {
		return "Track App", "Add a job to begin", 0, false
	}
	client := c.store.ClientName(t.ClientID)
	sub := "Job timer"
	if client != "" {
		sub = client
	}
	return t.Name, sub, t.CurrentElapsed(models.NowMs()), t.Running
}

func (c *Coordinator) ShowHUDOnLaunch() bool {
	p := c.loadPrefs()
	if p.ShowHUDOnLaunch == nil {
		return true
	}
	return *p.ShowHUDOnLaunch
}

func (c *Coordinator) SetShowHUDOnLaunch(v bool) {
	p := c.loadPrefs()
	p.ShowHUDOnLaunch = &v
	c.savePrefs(p)
}

func (c *Coordinator) NeedsOnboarding() bool {
	p := c.loadPrefs()
	if p.OnboardingComplete {
		return false
	}
	return len(c.store.Timers) == 0
}

func (c *Coordinator) SetOnboardingComplete(v bool) {
	p := c.loadPrefs()
	p.OnboardingComplete = v
	c.savePrefs(p)
}

// ResetToDefaults removes saved preferences so factory defaults apply again.
func (c *Coordinator) ResetToDefaults() {
	_ = os.Remove(c.prefs.path)
}

func (c *Coordinator) HUDCorner() int {
	return c.loadPrefs().HUDCorner
}

func (c *Coordinator) SetHUDCorner(corner int) {
	p := c.loadPrefs()
	p.HUDCorner = corner % 4
	c.savePrefs(p)
}

func (c *Coordinator) CycleHUDCorner() int {
	p := c.loadPrefs()
	p.HUDCorner = (p.HUDCorner + 1) % 4
	c.savePrefs(p)
	return p.HUDCorner
}

func (c *Coordinator) HUDHintDismissed() bool {
	return c.loadPrefs().HUDHintDismissed
}

func (c *Coordinator) SetHUDHintDismissed(v bool) {
	p := c.loadPrefs()
	p.HUDHintDismissed = v
	c.savePrefs(p)
}

func (c *Coordinator) AutoStartPrompt() *models.AutoStartPrompt {
	p := c.loadPrefs()
	return p.AutoStart
}

func (c *Coordinator) SetAutoStartPrompt(prompt *models.AutoStartPrompt) {
	p := c.loadPrefs()
	p.AutoStart = prompt
	c.savePrefs(p)
}

func (c *Coordinator) ConfirmAutoStart() {
	p := c.loadPrefs()
	p.AutoStart = nil
	c.savePrefs(p)
}

func (c *Coordinator) SkipAutoStart() {
	p := c.loadPrefs()
	if p.AutoStart == nil {
		return
	}
	pr := p.AutoStart
	graceEnd := pr.PromptedAt + int64(models.GraceSeconds*1000)
	if models.NowMs() > graceEnd {
		p.AutoStart = nil
		c.savePrefs(p)
		return
	}
	c.manager.CancelRunningBurst(pr.TimerID)
	c.manager.RestoreElapsed(pr.TimerID, pr.PreStartElapsed)
	c.manager.DeleteSessionsSince(pr.TimerID, pr.PromptedAt)
	if proj := c.store.Project(pr.ProjectID); proj != nil {
		until := models.NowMs() + int64(models.CooldownSeconds*1000)
		proj.SkipCooldownUntil = &until
		c.store.UpdateProject(*proj)
	}
	p.AutoStart = nil
	c.savePrefs(p)
}

func (c *Coordinator) TodayRows() []TodayRow {
	from, to := format.ReportRangeBounds(models.RangeToday, nil, nil)
	fromMs := from.UnixMilli()
	toMs := to.UnixMilli()
	now := models.NowMs()
	byTimer := map[string]int{}
	for _, sess := range c.store.Sessions {
		if sess.Archived {
			continue
		}
		if sess.Start >= fromMs && sess.Start <= toMs {
			byTimer[sess.TimerID] += sess.Seconds
		}
	}
	for _, t := range c.store.Timers {
		if t.Archived {
			continue
		}
		if t.Running && t.StartedAt != nil && *t.StartedAt <= toMs {
			liveStart := max64(*t.StartedAt, fromMs)
			end := min64(toMs, now)
			live := int((end - liveStart) / 1000)
			if live > 0 {
				byTimer[t.ID] += live
			}
		}
	}
	var rows []TodayRow
	for id, sec := range byTimer {
		for _, t := range c.store.Timers {
			if t.ID == id && !t.Archived {
				rows = append(rows, TodayRow{
					TimerID: id, Name: t.Name, Client: c.store.ClientName(t.ClientID),
					Seconds: sec, Running: t.Running, Rate: t.Rate, Currency: t.Currency,
				})
				break
			}
		}
	}
	return rows
}

type TodayRow struct {
	TimerID  string
	Name     string
	Client   string
	Seconds  int
	Running  bool
	Rate     float64
	Currency string
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}