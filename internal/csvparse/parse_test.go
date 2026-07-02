package csvparse

import "testing"

func TestParseQuotedCSV(t *testing.T) {
	rows := Parse("Job,Seconds\n\"Hello, world\",120\n")
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[1][0] != "Hello, world" || rows[1][1] != "120" {
		t.Fatalf("unexpected row: %#v", rows[1])
	}
}

func TestHeaderIndex(t *testing.T) {
	idx := HeaderIndex([]string{"Date", "Duration (hh:mm:ss)", "Seconds"}, "duration(hh:mm:ss)")
	if idx != 1 {
		t.Fatalf("expected index 1, got %d", idx)
	}
}