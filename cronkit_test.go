package cronkit

import (
	"testing"
	"time"
)

func TestParse_ValidExpressions(t *testing.T) {
	tests := []struct {
		expr string
	}{
		{"* * * * *"},
		{"*/15 9-17 * * 1-5"},
	}

	for _, tt := range tests {
		if _, err := Parse(tt.expr); err != nil {
			t.Errorf("Parse(%q) failed: %v", tt.expr, err)
		}
	}
}

func TestParse_InvalidExpressions(t *testing.T) {
	cases := []string{
		"* * * *",
		"*/a * * * *",
		"*/0 * * * *",
		"5- * * * *",
		"100 * * * *",
		"* * * * 8",
		"*/10/2 * * * *",
	}

	for _, expr := range cases {
		if _, err := Parse(expr); err == nil {
			t.Errorf("expected error for %q, got nil", expr)
		}
	}
}

func TestFieldSet_Has(t *testing.T) {
	f := fieldSet{vals: map[int]bool{1: true, 5: true}}
	if !f.has(1) {
		t.Error("expected true for 1")
	}
	if f.has(2) {
		t.Error("expected false for 2")
	}
	all := fieldSet{all: true}
	if !all.has(42) {
		t.Error("expected true when all=true")
	}
}

func TestParseField_Basic(t *testing.T) {
	tests := []struct {
		input    string
		lo, hi   int
		expected map[int]bool
	}{
		{"*", 0, 5, nil},
		{"1,3,5", 0, 5, map[int]bool{1: true, 3: true, 5: true}},
		{"1-3", 0, 5, map[int]bool{1: true, 2: true, 3: true}},
		{"*/2", 0, 5, map[int]bool{0: true, 2: true, 4: true}},
		{"1-4/2", 0, 5, map[int]bool{1: true, 3: true}},
	}
	for _, tt := range tests {
		fs, err := parseField(tt.input, tt.lo, tt.hi)
		if err != nil {
			t.Errorf("parseField(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if tt.input == "*" {
			if !fs.all {
				t.Errorf("%q: expected all=true", tt.input)
			}
			continue
		}
		for k := range tt.expected {
			if !fs.vals[k] {
				t.Errorf("%q missing %d", tt.input, k)
			}
		}
	}
}

func TestParseField_Invalid(t *testing.T) {
	cases := []string{
		"a", "*/0", "1-5/0", "1--3", "1-", "*/", "*/-1",
	}
	for _, c := range cases {
		if _, err := parseField(c, 0, 59); err == nil {
			t.Errorf("expected error for %q", c)
		}
	}
}

func TestNext_BasicMinuteIncrements(t *testing.T) {
	expr, _ := Parse("*/5 * * * *")
	base := time.Date(2025, 10, 26, 10, 7, 0, 0, time.UTC)
	next := expr.Next(base)
	expected := time.Date(2025, 10, 26, 10, 10, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("Next got %v, expected %v", next, expected)
	}
}

func TestNext_HourChange(t *testing.T) {
	expr, _ := Parse("0 * * * *")
	base := time.Date(2025, 10, 26, 10, 30, 0, 0, time.UTC)
	next := expr.Next(base)
	expected := time.Date(2025, 10, 26, 11, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("Next got %v, expected %v", next, expected)
	}
}

func TestNext_DayChange(t *testing.T) {
	expr, _ := Parse("0 0 * * 1")
	base := time.Date(2025, 10, 26, 0, 0, 0, 0, time.UTC)
	next := expr.Next(base)

	expected := time.Date(2025, 10, 27, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, next)
	}
}

func TestNext_MonthChange(t *testing.T) {
	expr, _ := Parse("0 0 1 * *")
	base := time.Date(2025, 1, 31, 10, 0, 0, 0, time.UTC)
	next := expr.Next(base)
	if next.Month() != time.February || next.Day() != 1 {
		t.Errorf("expected february 1, got %v", next)
	}
}

func TestNext_Edge31FebruaryLike(t *testing.T) {
	expr, _ := Parse("0 0 31 2 *")
	base := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	next := expr.Next(base)
	if next.Month() == time.February && next.Day() == 31 {
		t.Errorf("invalid february 31 detected")
	}
}

func TestNext_WeekdayOrDayLogic(t *testing.T) {
	expr, _ := Parse("* * 15 * 1")
	tm := time.Date(2025, 10, 14, 10, 0, 0, 0, time.UTC)
	next := expr.Next(tm)
	if !(next.Day() == 15 || next.Weekday() == time.Monday) {
		t.Errorf("expected 15th or monday, got %v", next)
	}
}
