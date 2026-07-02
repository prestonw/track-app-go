package csvparse

import "strings"

// Parse reads RFC4180-style CSV rows from text.
func Parse(text string) [][]string {
	var rows [][]string
	var row []string
	cur := ""
	inQuotes := false
	i := 0
	for i < len(text) {
		ch := text[i]
		if ch == '"' && inQuotes && i+1 < len(text) && text[i+1] == '"' {
			cur += `"`
			i += 2
			continue
		}
		if ch == '"' {
			inQuotes = !inQuotes
			i++
			continue
		}
		if !inQuotes && ch == ',' {
			row = append(row, cur)
			cur = ""
			i++
			continue
		}
		if !inQuotes && (ch == '\n' || ch == '\r') {
			if ch == '\r' && i+1 < len(text) && text[i+1] == '\n' {
				i++
			}
			row = append(row, cur)
			cur = ""
			if rowHasData(row) {
				rows = append(rows, row)
			}
			row = nil
			i++
			continue
		}
		cur += string(ch)
		i++
	}
	if cur != "" || len(row) > 0 {
		row = append(row, cur)
		if rowHasData(row) {
			rows = append(rows, row)
		}
	}
	return rows
}

func rowHasData(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return true
		}
	}
	return false
}

func HeaderIndex(headers []string, name string) int {
	norm := strings.ToLower(strings.ReplaceAll(name, " ", ""))
	for i, h := range headers {
		if strings.ToLower(strings.ReplaceAll(h, " ", "")) == norm {
			return i
		}
	}
	return -1
}