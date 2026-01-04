package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseTableSeparatorAndData(t *testing.T) {
	input := `
| Name | Notes |
| --- | --- |
| Alice | uses---dashes |
`
	rows, err := parseTable(input)
	if err != nil {
		t.Fatalf("parseTable error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[1][1] != "uses---dashes" {
		t.Fatalf("unexpected cell value: %q", rows[1][1])
	}
}

func TestParseTableNormalizeRows(t *testing.T) {
	input := `
| A | B |
| --- | --- |
| 1 | 2 | 3 |
`
	rows, err := parseTable(input)
	if err != nil {
		t.Fatalf("parseTable error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if len(rows[0]) != 3 {
		t.Fatalf("expected header to be padded to 3 columns, got %d", len(rows[0]))
	}
	if rows[0][2] != "" {
		t.Fatalf("expected empty padding in header, got %q", rows[0][2])
	}
	if rows[1][2] != "3" {
		t.Fatalf("expected extra column to be preserved, got %q", rows[1][2])
	}
}

func TestParseTableLargeLine(t *testing.T) {
	longCell := strings.Repeat("a", 100000)
	input := fmt.Sprintf("| H |\n| --- |\n| %s |", longCell)
	rows, err := parseTable(input)
	if err != nil {
		t.Fatalf("parseTable error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[1][0] != longCell {
		t.Fatalf("large cell mismatch: got %d chars", len(rows[1][0]))
	}
}
