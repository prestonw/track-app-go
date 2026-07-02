package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/prestonw/track-app-go/internal/format"
	"github.com/prestonw/track-app-go/internal/models"
	_ "modernc.org/sqlite"
)

type Store struct {
	db  *sql.DB
	dir string

	Timers           []models.JobTimer
	Sessions         []models.Session
	Clients          []models.Client
	CustomCurrencies []models.CustomCurrency
	Projects         []models.Project
	ProjectRules     []models.ProjectRule
	ActivityLog      []models.ActivitySegment
}

func DefaultPath() string {
	base, _ := os.UserConfigDir()
	return filepath.Join(base, "TrackApp", "track-app.sqlite")
}

func Open(path string) (*Store, error) {
	if path == "" {
		path = DefaultPath()
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)")
	if err != nil {
		return nil, err
	}
	s := &Store{db: db, dir: dir}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	if err := s.Reload(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS timers (
    id TEXT PRIMARY KEY, name TEXT NOT NULL, tags TEXT DEFAULT '[]',
    rate REAL DEFAULT 0, currency TEXT DEFAULT 'GBP', notes TEXT DEFAULT '',
    elapsed INTEGER DEFAULT 0, running INTEGER DEFAULT 0,
    started_at INTEGER, adjustments TEXT DEFAULT '[]', client_id TEXT DEFAULT ''
);
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY, timer_id TEXT, name TEXT, tags TEXT DEFAULT '[]',
    rate REAL DEFAULT 0, currency TEXT DEFAULT 'GBP', notes TEXT DEFAULT '',
    start INTEGER, end_ts INTEGER, seconds INTEGER DEFAULT 0, manual INTEGER DEFAULT 0,
    client TEXT DEFAULT ''
);
CREATE TABLE IF NOT EXISTS custom_currencies (code TEXT PRIMARY KEY, symbol TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS clients (id TEXT PRIMARY KEY, name TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY, name TEXT NOT NULL, client_id TEXT DEFAULT '',
    timer_id TEXT DEFAULT '', auto_track INTEGER DEFAULT 0, notes TEXT DEFAULT '',
    skip_cooldown_until INTEGER
);
CREATE TABLE IF NOT EXISTS project_rules (
    id TEXT PRIMARY KEY, project_id TEXT NOT NULL, kind TEXT NOT NULL, pattern TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS activity_log (
    id TEXT PRIMARY KEY, started_at INTEGER NOT NULL, ended_at INTEGER,
    app_name TEXT, bundle_id TEXT, window_title TEXT, document_path TEXT, project_id TEXT
);
`)
	if err != nil {
		return err
	}
	_, _ = s.db.Exec("ALTER TABLE timers ADD COLUMN client_id TEXT DEFAULT ''")
	_, _ = s.db.Exec("ALTER TABLE sessions ADD COLUMN client TEXT DEFAULT ''")
	_, _ = s.db.Exec("ALTER TABLE projects ADD COLUMN skip_cooldown_until INTEGER")
	return nil
}

func (s *Store) Reload() error {
	var err error
	if s.Timers, err = s.fetchTimers(); err != nil {
		return err
	}
	if s.Sessions, err = s.fetchSessions(); err != nil {
		return err
	}
	if s.Clients, err = s.fetchClients(); err != nil {
		return err
	}
	if s.CustomCurrencies, err = s.fetchCurrencies(); err != nil {
		return err
	}
	if s.Projects, err = s.fetchProjects(); err != nil {
		return err
	}
	if s.ProjectRules, err = s.fetchRules(); err != nil {
		return err
	}
	if s.ActivityLog, err = s.fetchActivity(); err != nil {
		return err
	}
	return nil
}

func (s *Store) SaveAll() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmts := []string{
		"DELETE FROM timers", "DELETE FROM sessions", "DELETE FROM clients",
		"DELETE FROM custom_currencies", "DELETE FROM projects", "DELETE FROM project_rules",
	}
	for _, q := range stmts {
		if _, err := tx.Exec(q); err != nil {
			return err
		}
	}

	for _, t := range s.Timers {
		started := ""
		if t.StartedAt != nil {
			started = itoa64(*t.StartedAt)
		}
		_, err := tx.Exec(`INSERT INTO timers VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
			t.ID, t.Name, encode(t.Tags), t.Rate, t.Currency, t.Notes,
			t.Elapsed, bool01(t.Running), started, encode(t.Adjustments), t.ClientID)
		if err != nil {
			return err
		}
	}
	for _, sess := range s.Sessions {
		_, err := tx.Exec(`INSERT INTO sessions VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
			sess.ID, sess.TimerID, sess.Name, encode(sess.Tags), sess.Rate, sess.Currency, sess.Notes,
			sess.Start, sess.End, sess.Seconds, bool01(sess.Manual), sess.Client)
		if err != nil {
			return err
		}
	}
	for _, c := range s.Clients {
		if _, err := tx.Exec(`INSERT INTO clients VALUES (?,?)`, c.ID, c.Name); err != nil {
			return err
		}
	}
	for _, c := range s.CustomCurrencies {
		if _, err := tx.Exec(`INSERT INTO custom_currencies VALUES (?,?)`, c.Code, c.Symbol); err != nil {
			return err
		}
	}
	for _, p := range s.Projects {
		cool := ""
		if p.SkipCooldownUntil != nil {
			cool = itoa64(*p.SkipCooldownUntil)
		}
		_, err := tx.Exec(`INSERT INTO projects VALUES (?,?,?,?,?,?,?)`,
			p.ID, p.Name, p.ClientID, p.TimerID, bool01(p.AutoTrack), p.Notes, cool)
		if err != nil {
			return err
		}
	}
	for _, r := range s.ProjectRules {
		_, err := tx.Exec(`INSERT INTO project_rules VALUES (?,?,?,?)`,
			r.ID, r.ProjectID, string(r.Kind), r.Pattern)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ClientName(id string) string {
	for _, c := range s.Clients {
		if c.ID == id {
			return c.Name
		}
	}
	return ""
}

func (s *Store) AddClient(name string) (*models.Client, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, errors.New("empty name")
	}
	for _, c := range s.Clients {
		if strings.EqualFold(c.Name, trimmed) {
			return &c, nil
		}
	}
	c := models.Client{ID: models.MakeID(), Name: trimmed}
	s.Clients = append(s.Clients, c)
	return &c, s.SaveAll()
}

func (s *Store) DeleteClient(id string) error {
	for _, t := range s.Timers {
		if t.ClientID == id {
			return errors.New("in use")
		}
	}
	for _, p := range s.Projects {
		if p.ClientID == id {
			return errors.New("in use")
		}
	}
	s.Clients = filterClients(s.Clients, id)
	return s.SaveAll()
}

func (s *Store) AddCustomCurrency(code, symbol string) bool {
	code = strings.ToUpper(strings.TrimSpace(code))
	symbol = strings.TrimSpace(symbol)
	if code == "" || symbol == "" {
		return false
	}
	for _, c := range models.BuiltInCurrencies {
		if c.Code == code {
			return false
		}
	}
	for _, c := range s.CustomCurrencies {
		if c.Code == code {
			return false
		}
	}
	if len(symbol) > 4 {
		symbol = symbol[:4]
	}
	s.CustomCurrencies = append(s.CustomCurrencies, models.CustomCurrency{Code: code, Symbol: symbol})
	_ = s.SaveAll()
	return true
}

func (s *Store) AddTimer(name string, tags []string, rate float64, currency, notes, clientID string, manualSec int) models.JobTimer {
	var adjs []models.Adjustment
	elapsed := 0
	if manualSec != 0 {
		detail := "Created with " + format.HumanDuration(abs(manualSec)) + " pre-loaded"
		if manualSec < 0 {
			detail = "Created with " + format.HumanDuration(abs(manualSec)) + " subtracted"
		}
		adjs = append(adjs, models.Adjustment{Ts: models.NowMs(), Type: "add", Detail: detail})
		elapsed = max(0, manualSec)
	}
	t := models.JobTimer{
		ID: models.MakeID(), Name: strings.TrimSpace(name), Tags: tags, Rate: rate,
		Currency: currency, Notes: notes, Elapsed: elapsed, Adjustments: adjs, ClientID: clientID,
	}
	s.Timers = append([]models.JobTimer{t}, s.Timers...)
	if manualSec > 0 {
		s.Sessions = append([]models.Session{s.makeManualSession(t, manualSec)}, s.Sessions...)
	}
	_ = s.SaveAll()
	return t
}

func (s *Store) UpdateTimer(t models.JobTimer, manualSec int) {
	for i := range s.Timers {
		if s.Timers[i].ID != t.ID {
			continue
		}
		if manualSec != 0 {
			t.Elapsed = max(0, t.Elapsed+manualSec)
			if manualSec > 0 {
				s.Sessions = append([]models.Session{s.makeManualSession(t, manualSec)}, s.Sessions...)
			}
			t.Adjustments = append(t.Adjustments, models.Adjustment{
				Ts: models.NowMs(), Type: "add",
				Detail: "Manual adjust: " + format.HumanDuration(abs(manualSec)),
			})
		}
		s.Timers[i] = t
		break
	}
	_ = s.SaveAll()
}

func (s *Store) DeleteTimer(id string) {
	s.Timers = filterTimers(s.Timers, id)
	_ = s.SaveAll()
}

func (s *Store) DeleteSession(id string) {
	for i, sess := range s.Sessions {
		if sess.ID != id {
			continue
		}
		for j := range s.Timers {
			if s.Timers[j].ID == sess.TimerID {
				s.Timers[j].Elapsed = max(0, s.Timers[j].Elapsed-sess.Seconds)
			}
		}
		s.Sessions = append(s.Sessions[:i], s.Sessions[i+1:]...)
		break
	}
	_ = s.SaveAll()
}

func (s *Store) AddProject(name, clientID, timerID string, autoTrack bool, notes string) models.Project {
	p := models.Project{
		ID: models.MakeID(), Name: strings.TrimSpace(name),
		ClientID: clientID, TimerID: timerID, AutoTrack: autoTrack, Notes: notes,
	}
	s.Projects = append(s.Projects, p)
	_ = s.SaveAll()
	return p
}

func (s *Store) UpdateProject(p models.Project) {
	for i := range s.Projects {
		if s.Projects[i].ID == p.ID {
			s.Projects[i] = p
			break
		}
	}
	_ = s.SaveAll()
}

func (s *Store) DeleteProject(id string) {
	s.Projects = filterProjects(s.Projects, id)
	s.ProjectRules = filterRulesByProject(s.ProjectRules, id)
	_ = s.SaveAll()
}

func (s *Store) Project(id string) *models.Project {
	for i := range s.Projects {
		if s.Projects[i].ID == id {
			return &s.Projects[i]
		}
	}
	return nil
}

func (s *Store) RulesFor(projectID string) []models.ProjectRule {
	var out []models.ProjectRule
	for _, r := range s.ProjectRules {
		if r.ProjectID == projectID {
			out = append(out, r)
		}
	}
	return out
}

func (s *Store) AddRule(projectID string, kind models.RuleKind, pattern string) (*models.ProjectRule, error) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return nil, errors.New("empty pattern")
	}
	r := models.ProjectRule{ID: models.MakeID(), ProjectID: projectID, Kind: kind, Pattern: pattern}
	s.ProjectRules = append(s.ProjectRules, r)
	return &r, s.SaveAll()
}

func (s *Store) DeleteRule(id string) {
	s.ProjectRules = filterRules(s.ProjectRules, id)
	_ = s.SaveAll()
}

func (s *Store) ProjectsLinkedTo(timerID string) []models.Project {
	var out []models.Project
	for _, p := range s.Projects {
		if p.TimerID == timerID {
			out = append(out, p)
		}
	}
	return out
}

func (s *Store) MatchProject(ctx models.ForegroundContext) *models.Project {
	for _, p := range s.Projects {
		rules := s.RulesFor(p.ID)
		if len(rules) == 0 {
			continue
		}
		for _, r := range rules {
			if s.MatchesRule(r, ctx) {
				return &p
			}
		}
	}
	return nil
}

func (s *Store) IsInSkipCooldown(p models.Project) bool {
	if p.SkipCooldownUntil == nil {
		return false
	}
	return *p.SkipCooldownUntil > models.NowMs()
}

func (s *Store) MatchesRule(r models.ProjectRule, ctx models.ForegroundContext) bool {
	pat := strings.ToLower(strings.TrimSpace(r.Pattern))
	if pat == "" {
		return false
	}
	switch r.Kind {
	case models.RuleAppBundle:
		return strings.Contains(strings.ToLower(ctx.BundleID), pat)
	case models.RuleWindowTitle:
		return strings.Contains(strings.ToLower(ctx.WindowTitle), pat)
	case models.RuleDocumentPath:
		return strings.Contains(strings.ToLower(ctx.DocumentPath), pat)
	}
	return false
}

func (s *Store) UngroupedActivity(days int) []models.ActivitySegment {
	cutoff := models.NowMs() - int64(days*86400000)
	var out []models.ActivitySegment
	for _, seg := range s.ActivityLog {
		if seg.ProjectID != "" {
			continue
		}
		if seg.StartedAt >= cutoff {
			out = append(out, seg)
		}
	}
	return out
}

func (s *Store) AssignActivity(id, projectID string) {
	for i := range s.ActivityLog {
		if s.ActivityLog[i].ID == id {
			s.ActivityLog[i].ProjectID = projectID
			_, _ = s.db.Exec(`UPDATE activity_log SET project_id=? WHERE id=?`, projectID, id)
			break
		}
	}
	_ = s.SaveAll()
}

func (s *Store) makeManualSession(t models.JobTimer, seconds int) models.Session {
	ts := models.NowMs()
	notes := t.Notes
	if notes != "" {
		notes += " "
	}
	notes += "[manual entry]"
	return models.Session{
		ID: models.MakeID(), TimerID: t.ID, Name: t.Name, Tags: t.Tags,
		Rate: t.Rate, Currency: t.Currency, Notes: notes,
		Start: ts, End: ts, Seconds: seconds, Manual: true, Client: s.ClientName(t.ClientID),
	}
}

// fetch helpers

func (s *Store) fetchTimers() ([]models.JobTimer, error) {
	rows, err := s.db.Query(`SELECT id,name,tags,rate,currency,notes,elapsed,running,started_at,adjustments,client_id FROM timers ORDER BY rowid`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.JobTimer
	for rows.Next() {
		var t models.JobTimer
		var tags, adjs, started, running string
		if err := rows.Scan(&t.ID, &t.Name, &tags, &t.Rate, &t.Currency, &t.Notes, &t.Elapsed, &running, &started, &adjs, &t.ClientID); err != nil {
			return nil, err
		}
		decode(tags, &t.Tags)
		decode(adjs, &t.Adjustments)
		t.Running = running == "1"
		if started != "" {
			v := parseInt64(started)
			t.StartedAt = &v
		}
		if t.Currency == "" {
			t.Currency = "GBP"
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) fetchSessions() ([]models.Session, error) {
	rows, err := s.db.Query(`SELECT id,timer_id,name,tags,rate,currency,notes,start,end_ts,seconds,manual,client FROM sessions ORDER BY start DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Session
	for rows.Next() {
		var sess models.Session
		var tags, manual string
		if err := rows.Scan(&sess.ID, &sess.TimerID, &sess.Name, &tags, &sess.Rate, &sess.Currency, &sess.Notes,
			&sess.Start, &sess.End, &sess.Seconds, &manual, &sess.Client); err != nil {
			return nil, err
		}
		decode(tags, &sess.Tags)
		sess.Manual = manual == "1"
		out = append(out, sess)
	}
	return out, rows.Err()
}

func (s *Store) fetchClients() ([]models.Client, error) {
	rows, err := s.db.Query(`SELECT id,name FROM clients ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Client
	for rows.Next() {
		var c models.Client
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) fetchCurrencies() ([]models.CustomCurrency, error) {
	rows, err := s.db.Query(`SELECT code,symbol FROM custom_currencies`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.CustomCurrency
	for rows.Next() {
		var c models.CustomCurrency
		if err := rows.Scan(&c.Code, &c.Symbol); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) fetchProjects() ([]models.Project, error) {
	rows, err := s.db.Query(`SELECT id,name,client_id,timer_id,auto_track,notes,skip_cooldown_until FROM projects ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Project
	for rows.Next() {
		var p models.Project
		var auto, cool sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &p.ClientID, &p.TimerID, &auto, &p.Notes, &cool); err != nil {
			return nil, err
		}
		p.AutoTrack = auto.String != "0" && auto.String != ""
		if cool.Valid && cool.String != "" {
			v := parseInt64(cool.String)
			p.SkipCooldownUntil = &v
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) fetchRules() ([]models.ProjectRule, error) {
	rows, err := s.db.Query(`SELECT id,project_id,kind,pattern FROM project_rules ORDER BY rowid`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.ProjectRule
	for rows.Next() {
		var r models.ProjectRule
		var kind string
		if err := rows.Scan(&r.ID, &r.ProjectID, &kind, &r.Pattern); err != nil {
			return nil, err
		}
		r.Kind = models.RuleKind(kind)
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) fetchActivity() ([]models.ActivitySegment, error) {
	rows, err := s.db.Query(`SELECT id,started_at,ended_at,app_name,bundle_id,window_title,document_path,project_id FROM activity_log ORDER BY started_at DESC LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.ActivitySegment
	for rows.Next() {
		var seg models.ActivitySegment
		var ended, proj sql.NullString
		if err := rows.Scan(&seg.ID, &seg.StartedAt, &ended, &seg.AppName, &seg.BundleID, &seg.WindowTitle, &seg.DocumentPath, &proj); err != nil {
			return nil, err
		}
		if ended.Valid && ended.String != "" {
			v := parseInt64(ended.String)
			seg.EndedAt = &v
		}
		if proj.Valid {
			seg.ProjectID = proj.String
		}
		out = append(out, seg)
	}
	return out, rows.Err()
}

func encode(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func decode(src string, dest any) {
	_ = json.Unmarshal([]byte(src), dest)
}

func bool01(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func itoa64(n int64) string { return strconv.FormatInt(n, 10) }

func parseInt64(s string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return n
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func filterTimers(in []models.JobTimer, id string) []models.JobTimer {
	var out []models.JobTimer
	for _, t := range in {
		if t.ID != id {
			out = append(out, t)
		}
	}
	return out
}

func filterClients(in []models.Client, id string) []models.Client {
	var out []models.Client
	for _, c := range in {
		if c.ID != id {
			out = append(out, c)
		}
	}
	return out
}

func filterProjects(in []models.Project, id string) []models.Project {
	var out []models.Project
	for _, p := range in {
		if p.ID != id {
			out = append(out, p)
		}
	}
	return out
}

func filterRules(in []models.ProjectRule, id string) []models.ProjectRule {
	var out []models.ProjectRule
	for _, r := range in {
		if r.ID != id {
			out = append(out, r)
		}
	}
	return out
}

func filterRulesByProject(in []models.ProjectRule, pid string) []models.ProjectRule {
	var out []models.ProjectRule
	for _, r := range in {
		if r.ProjectID != pid {
			out = append(out, r)
		}
	}
	return out
}