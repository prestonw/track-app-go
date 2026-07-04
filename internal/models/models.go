package models

import (
	"fmt"
	"math/rand"
	"time"
)

type CurrencyInfo struct {
	Code   string `json:"code"`
	Symbol string `json:"symbol"`
}

var BuiltInCurrencies = []CurrencyInfo{
	{Code: "GBP", Symbol: "£"},
	{Code: "USD", Symbol: "$"},
	{Code: "CAD", Symbol: "C$"},
	{Code: "EUR", Symbol: "€"},
}

type Client struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Adjustment struct {
	Ts     int64  `json:"ts"`
	Type   string `json:"type"`
	Detail string `json:"detail"`
}

type JobTimer struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Tags        []string     `json:"tags"`
	Rate        float64      `json:"rate"`
	Currency    string       `json:"currency"`
	Notes       string       `json:"notes"`
	Elapsed     int          `json:"elapsed"`
	Running     bool         `json:"running"`
	StartedAt   *int64       `json:"startedAt,omitempty"`
	Adjustments []Adjustment `json:"adjustments"`
	ClientID    string       `json:"clientId"`
	Archived    bool         `json:"archived,omitempty"`
}

func (t JobTimer) CurrentElapsed(nowMs int64) int {
	if !t.Running || t.StartedAt == nil {
		return t.Elapsed
	}
	live := int((nowMs - *t.StartedAt) / 1000)
	if live < 0 {
		live = 0
	}
	return t.Elapsed + live
}

type Session struct {
	ID       string   `json:"id"`
	TimerID  string   `json:"timerId"`
	Name     string   `json:"name"`
	Tags     []string `json:"tags"`
	Rate     float64  `json:"rate"`
	Currency string   `json:"currency"`
	Notes    string   `json:"notes"`
	Start    int64    `json:"start"`
	End      int64    `json:"end"`
	Seconds  int      `json:"seconds"`
	Manual   bool     `json:"manual"`
	Client   string   `json:"client"`
	Archived bool     `json:"archived,omitempty"`
}

type CustomCurrency struct {
	Code   string `json:"code"`
	Symbol string `json:"symbol"`
}

type Project struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	ClientID          string `json:"clientId"`
	TimerID           string `json:"timerId"`
	AutoTrack         bool   `json:"autoTrack"`
	Notes             string `json:"notes"`
	SkipCooldownUntil *int64 `json:"skipCooldownUntil,omitempty"`
}

type AutoStartPrompt struct {
	ProjectID       string `json:"projectId"`
	TimerID         string `json:"timerId"`
	PreStartElapsed int    `json:"preStartElapsed"`
	PromptedAt      int64  `json:"promptedAt"`
}

const GraceSeconds = 60
const CooldownSeconds = 15 * 60

type RuleKind string

const (
	RuleAppBundle     RuleKind = "App"
	RuleWindowTitle   RuleKind = "Window title"
	RuleDocumentPath  RuleKind = "Document path"
)

var AllRuleKinds = []RuleKind{RuleAppBundle, RuleWindowTitle, RuleDocumentPath}

type ProjectRule struct {
	ID        string   `json:"id"`
	ProjectID string   `json:"projectId"`
	Kind      RuleKind `json:"kind"`
	Pattern   string   `json:"pattern"`
}

type ActivitySegment struct {
	ID           string `json:"id"`
	StartedAt    int64  `json:"startedAt"`
	EndedAt      *int64 `json:"endedAt,omitempty"`
	AppName      string `json:"appName"`
	BundleID     string `json:"bundleId"`
	WindowTitle  string `json:"windowTitle"`
	DocumentPath string `json:"documentPath"`
	ProjectID    string `json:"projectId,omitempty"`
	Archived     bool   `json:"archived,omitempty"`
}

func (s ActivitySegment) IsOpen() bool { return s.EndedAt == nil }

func (s ActivitySegment) DurationSeconds(nowMs int64) int {
	end := nowMs
	if s.EndedAt != nil {
		end = *s.EndedAt
	}
	sec := int((end - s.StartedAt) / 1000)
	if sec < 0 {
		return 0
	}
	return sec
}

type ForegroundContext struct {
	AppName      string `json:"appName"`
	BundleID     string `json:"bundleId"`
	WindowTitle  string `json:"windowTitle"`
	DocumentPath string `json:"documentPath"`
}

type ReportRange string

const (
	RangeToday  ReportRange = "Today"
	RangeWeek   ReportRange = "This week"
	RangeMonth  ReportRange = "This month"
	RangeAll    ReportRange = "All time"
	RangeCustom ReportRange = "Custom range"
)

var AllReportRanges = []ReportRange{RangeToday, RangeWeek, RangeMonth, RangeAll, RangeCustom}

type SessionTypeFilter string

const (
	FilterAll     SessionTypeFilter = "All types"
	FilterTracked SessionTypeFilter = "Tracked"
	FilterManual  SessionTypeFilter = "Manual"
)

type ImportStrategy string

const (
	ImportSkip    ImportStrategy = "skip"
	ImportMerge   ImportStrategy = "merge"
	ImportReplace ImportStrategy = "replace"
)

type ImportConflictSummary struct {
	TimerCount   int
	SessionCount int
}

func MakeID() string {
	ts := time.Now().UnixMilli()
	r := rand.Intn(0x10000)
	return fmt.Sprintf("%s%04x", formatBase36(ts), r)
}

func formatBase36(n int64) string {
	if n == 0 {
		return "0"
	}
	const digits = "0123456789abcdefghijklmnopqrstuvwxyz"
	var out []byte
	for n > 0 {
		out = append([]byte{digits[n%36]}, out...)
		n /= 36
	}
	return string(out)
}

func NowMs() int64 { return time.Now().UnixMilli() }