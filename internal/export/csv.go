package export

import (
	"fmt"
	"strings"

	"github.com/prestonw/track-app-go/internal/format"
	"github.com/prestonw/track-app-go/internal/models"
)

// SessionsCSV exports sessions compatible with the Swift Track App CSV format.
func SessionsCSV(sessions []models.Session, custom []models.CustomCurrency) string {
	header := []string{
		"Date", "Job", "Client", "Tags", "Type", "Duration (hh:mm:ss)", "Seconds",
		"Currency", "Rate", "Earned", "Notes", "SessionID", "TimerID", "StartTS", "EndTS",
	}
	lines := []string{joinRow(header)}
	for _, s := range sessions {
		earned := ""
		if s.Rate > 0 {
			earned = fmt.Sprintf("%.2f", float64(s.Seconds)/3600*s.Rate)
		}
		notes := strings.TrimSpace(strings.ReplaceAll(s.Notes, "[manual entry]", ""))
		typ := "tracked"
		if s.Manual {
			typ = "manual"
		}
		row := []string{
			format.Date(s.Start), s.Name, s.Client, strings.Join(s.Tags, "; "),
			typ, format.Duration(s.Seconds), fmt.Sprintf("%d", s.Seconds), s.Currency,
			fmt.Sprintf("%.2f", s.Rate), earned, notes, s.ID, s.TimerID,
			fmt.Sprintf("%d", s.Start), fmt.Sprintf("%d", s.End),
		}
		lines = append(lines, joinRow(row))
	}
	return strings.Join(lines, "\n") + "\n"
}

func joinRow(cols []string) string {
	for i, c := range cols {
		cols[i] = escapeCSV(c)
	}
	return strings.Join(cols, ",")
}

func escapeCSV(s string) string {
	if strings.ContainsAny(s, ",\"\n\r") {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}