package format

import (
	"fmt"
	"strings"
	"time"

	"github.com/prestonw/track-app-go/internal/models"
)

var DisplayCurrency = "GBP"

func Symbol(code string, custom []models.CustomCurrency) string {
	for _, c := range models.BuiltInCurrencies {
		if c.Code == code {
			return c.Symbol
		}
	}
	for _, c := range custom {
		if c.Code == code {
			return c.Symbol
		}
	}
	return code
}

func AllCurrencies(custom []models.CustomCurrency) []models.CurrencyInfo {
	out := append([]models.CurrencyInfo{}, models.BuiltInCurrencies...)
	for _, c := range custom {
		out = append(out, models.CurrencyInfo{Code: c.Code, Symbol: c.Symbol})
	}
	return out
}

func Duration(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func HumanDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	h := seconds / 3600
	m := (seconds % 3600) / 60
	if h == 0 {
		return fmt.Sprintf("%dm", m)
	}
	if m > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dh", h)
}

func Money(amount float64, code string, custom []models.CustomCurrency) string {
	sym := Symbol(code, custom)
	formatted := fmt.Sprintf("%.2f", amount)
	parts := strings.SplitN(formatted, ".", 2)
	whole := parts[0]
	if len(whole) > 3 {
		var grouped strings.Builder
		for i, ch := range whole {
			if i > 0 && (len(whole)-i)%3 == 0 {
				grouped.WriteByte(',')
			}
			grouped.WriteRune(ch)
		}
		whole = grouped.String()
	}
	frac := "00"
	if len(parts) > 1 {
		frac = parts[1]
	}
	return sym + whole + "." + frac
}

func Date(ts int64) string {
	return time.UnixMilli(ts).Format("2 Jan 2006")
}

func DateTime(ts int64) string {
	return time.UnixMilli(ts).Format("2 Jan, 15:04")
}

func ReportRangeBounds(r models.ReportRange, from, to *time.Time) (time.Time, time.Time) {
	now := time.Now()
	switch r {
	case models.RangeToday:
		start := startOfDay(now)
		return start, endOfDay(now)
	case models.RangeWeek:
		start := startOfDay(now.AddDate(0, 0, -int(now.Weekday())))
		if now.Weekday() == time.Sunday {
			start = startOfDay(now.AddDate(0, 0, -6))
		}
		return start, endOfDay(now)
	case models.RangeMonth:
		y, m, _ := now.Date()
		start := time.Date(y, m, 1, 0, 0, 0, 0, now.Location())
		return start, endOfDay(now)
	case models.RangeCustom:
		if from != nil && to != nil {
			return startOfDay(*from), endOfDay(*to)
		}
		fallthrough
	default:
		return time.Unix(0, 0), endOfDay(now)
	}
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func endOfDay(t time.Time) time.Time {
	return startOfDay(t).Add(24*time.Hour - time.Second)
}