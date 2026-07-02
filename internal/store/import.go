package store

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/prestonw/track-app-go/internal/csvparse"
	"github.com/prestonw/track-app-go/internal/models"
)

var (
	ErrImportInvalidFile = errors.New("not a valid Track App database")
	ErrImportEmpty       = errors.New("no data found to import")
	ErrImportColumns     = errors.New("CSV is missing required columns (Job, Seconds)")
)

func (s *Store) DetectImportConflicts(path string) (models.ImportConflictSummary, error) {
	imp, err := s.openImportDB(path)
	if err != nil {
		return models.ImportConflictSummary{}, err
	}
	defer s.closeImportDB(imp)

	summary := models.ImportConflictSummary{}
	for _, t := range imp.Timers {
		for _, existing := range s.Timers {
			if existing.ID == t.ID {
				summary.TimerCount++
				break
			}
		}
	}
	for _, sess := range imp.Sessions {
		for _, existing := range s.Sessions {
			if existing.ID == sess.ID {
				summary.SessionCount++
				break
			}
		}
	}
	return summary, nil
}

func (s *Store) ImportDatabase(path string, strategy models.ImportStrategy) error {
	imp, err := s.openImportDB(path)
	if err != nil {
		return err
	}
	defer s.closeImportDB(imp)

	if len(imp.Timers) == 0 && len(imp.Sessions) == 0 {
		return ErrImportEmpty
	}
	s.applyImport(imp.Timers, imp.Sessions, imp.CustomCurrencies, imp.Clients, strategy)
	return nil
}

func (s *Store) ImportCSV(text string, strategy models.ImportStrategy) error {
	rows := csvparse.Parse(text)
	if len(rows) < 2 {
		return ErrImportEmpty
	}
	headers := rows[0]
	jobIdx := csvparse.HeaderIndex(headers, "job")
	secIdx := csvparse.HeaderIndex(headers, "seconds")
	if jobIdx < 0 || secIdx < 0 {
		return ErrImportColumns
	}

	var parsed []models.Session
	for _, row := range rows[1:] {
		if len(row) <= max(jobIdx, secIdx) {
			continue
		}
		seconds, _ := strconv.Atoi(strings.TrimSpace(row[secIdx]))
		if seconds <= 0 {
			continue
		}
		startTs := int64Field(row, csvparse.HeaderIndex(headers, "startts"))
		if startTs == 0 {
			startTs = models.NowMs()
		}
		endTs := int64Field(row, csvparse.HeaderIndex(headers, "endts"))
		if endTs == 0 {
			endTs = startTs + int64(seconds)*1000
		}
		tags := splitCSVField(row, csvparse.HeaderIndex(headers, "tags"), ";")
		typ := field(row, csvparse.HeaderIndex(headers, "type"))
		parsed = append(parsed, models.Session{
			ID:       fieldOrID(row, csvparse.HeaderIndex(headers, "sessionid")),
			TimerID:  fieldOrID(row, csvparse.HeaderIndex(headers, "timerid")),
			Name:     strings.TrimSpace(row[jobIdx]),
			Tags:     tags,
			Rate:     floatField(row, csvparse.HeaderIndex(headers, "rate")),
			Currency: defaultStr(field(row, csvparse.HeaderIndex(headers, "currency")), "GBP"),
			Notes:    field(row, csvparse.HeaderIndex(headers, "notes")),
			Start:    startTs,
			End:      endTs,
			Seconds:  seconds,
			Manual:   typ == "manual",
			Client:   field(row, csvparse.HeaderIndex(headers, "client")),
		})
	}
	if len(parsed) == 0 {
		return ErrImportEmpty
	}

	if strategy == models.ImportReplace {
		for _, sess := range parsed {
			if i := sessionIndex(s.Sessions, sess.ID); i >= 0 {
				s.Sessions[i] = sess
			} else {
				s.Sessions = append(s.Sessions, sess)
			}
		}
	} else {
		for _, sess := range parsed {
			if sessionIndex(s.Sessions, sess.ID) >= 0 {
				continue
			}
			s.Sessions = append(s.Sessions, sess)
			if i := timerIndex(s.Timers, sess.TimerID); i >= 0 {
				s.Timers[i].Elapsed += sess.Seconds
			} else {
				s.Timers = append(s.Timers, models.JobTimer{
					ID: sess.TimerID, Name: sess.Name, Tags: sess.Tags,
					Rate: sess.Rate, Currency: sess.Currency, Notes: sess.Notes,
					Elapsed: sess.Seconds,
				})
			}
		}
	}
	sortSessions(&s.Sessions)
	return s.SaveAll()
}

func (s *Store) openImportDB(path string) (*Store, error) {
	temp := filepath.Join(s.dir, "import-temp.sqlite")
	_ = os.Remove(temp)
	if err := copyFile(path, temp); err != nil {
		return nil, ErrImportInvalidFile
	}
	imp, err := Open(temp)
	if err != nil {
		_ = os.Remove(temp)
		return nil, ErrImportInvalidFile
	}
	return imp, nil
}

func (s *Store) closeImportDB(imp *Store) {
	if imp != nil {
		_ = imp.Close()
	}
	_ = os.Remove(filepath.Join(s.dir, "import-temp.sqlite"))
}

func (s *Store) applyImport(
	impTimers []models.JobTimer,
	impSessions []models.Session,
	impCurrencies []models.CustomCurrency,
	impClients []models.Client,
	strategy models.ImportStrategy,
) {
	timerMap := map[string]models.JobTimer{}
	for _, t := range s.Timers {
		timerMap[t.ID] = t
	}
	existingSessionIDs := map[string]struct{}{}
	for _, sess := range s.Sessions {
		existingSessionIDs[sess.ID] = struct{}{}
	}

	for _, imp := range impTimers {
		existing, ok := timerMap[imp.ID]
		if ok {
			switch strategy {
			case models.ImportReplace:
				timerMap[imp.ID] = imp
			case models.ImportMerge:
				merged := existing
				if imp.Elapsed > merged.Elapsed {
					merged.Elapsed = imp.Elapsed
				}
				adjTs := map[int64]struct{}{}
				for _, a := range merged.Adjustments {
					adjTs[a.Ts] = struct{}{}
				}
				for _, a := range imp.Adjustments {
					if _, seen := adjTs[a.Ts]; !seen {
						merged.Adjustments = append(merged.Adjustments, a)
					}
				}
				timerMap[imp.ID] = merged
			}
		} else {
			timerMap[imp.ID] = imp
		}
	}

	for _, imp := range impSessions {
		if _, exists := existingSessionIDs[imp.ID]; exists {
			if strategy == models.ImportReplace {
				for i, sess := range s.Sessions {
					if sess.ID == imp.ID {
						s.Sessions[i] = imp
						break
					}
				}
			}
			continue
		}
		s.Sessions = append(s.Sessions, imp)
		existingSessionIDs[imp.ID] = struct{}{}
		if t, ok := timerMap[imp.TimerID]; ok {
			t.Elapsed += imp.Seconds
			timerMap[imp.TimerID] = t
		} else {
			timerMap[imp.TimerID] = models.JobTimer{
				ID: imp.TimerID, Name: imp.Name, Tags: imp.Tags,
				Rate: imp.Rate, Currency: imp.Currency, Notes: imp.Notes,
				Elapsed: imp.Seconds,
			}
		}
	}

	s.Timers = make([]models.JobTimer, 0, len(timerMap))
	for _, t := range timerMap {
		s.Timers = append(s.Timers, t)
	}
	sortSessions(&s.Sessions)

	builtIn := map[string]struct{}{}
	for _, c := range models.BuiltInCurrencies {
		builtIn[c.Code] = struct{}{}
	}
	existingCodes := map[string]struct{}{}
	for code := range builtIn {
		existingCodes[code] = struct{}{}
	}
	for _, c := range s.CustomCurrencies {
		existingCodes[c.Code] = struct{}{}
	}
	for _, c := range impCurrencies {
		if _, ok := existingCodes[c.Code]; !ok {
			s.CustomCurrencies = append(s.CustomCurrencies, c)
			existingCodes[c.Code] = struct{}{}
		}
	}

	for _, c := range impClients {
		found := false
		for _, existing := range s.Clients {
			if existing.ID == c.ID {
				found = true
				break
			}
		}
		if !found {
			s.Clients = append(s.Clients, c)
		}
	}
	sort.Slice(s.Clients, func(i, j int) bool {
		return strings.ToLower(s.Clients[i].Name) < strings.ToLower(s.Clients[j].Name)
	})

	_ = s.SaveAll()
}

func sortSessions(sessions *[]models.Session) {
	sort.Slice(*sessions, func(i, j int) bool {
		return (*sessions)[i].Start > (*sessions)[j].Start
	})
}

func sessionIndex(sessions []models.Session, id string) int {
	for i, s := range sessions {
		if s.ID == id {
			return i
		}
	}
	return -1
}

func timerIndex(timers []models.JobTimer, id string) int {
	for i, t := range timers {
		if t.ID == id {
			return i
		}
	}
	return -1
}

func field(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

func fieldOrID(row []string, idx int) string {
	v := field(row, idx)
	if v != "" {
		return v
	}
	return models.MakeID()
}

func int64Field(row []string, idx int) int64 {
	v, _ := strconv.ParseInt(field(row, idx), 10, 64)
	return v
}

func floatField(row []string, idx int) float64 {
	v, _ := strconv.ParseFloat(field(row, idx), 64)
	return v
}

func splitCSVField(row []string, idx int, sep string) []string {
	raw := field(row, idx)
	if raw == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(raw, sep) {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func defaultStr(val, fallback string) string {
	if val == "" {
		return fallback
	}
	return val
}