package cronkit

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Expression represents a parsed cron expression.
type Expression struct {
	minutes  fieldSet
	hours    fieldSet
	days     fieldSet
	months   fieldSet
	weekdays fieldSet
}

// Next returns the next time instant matching the expression strictly after `after`.
func (e Expression) Next(after time.Time) time.Time {
	t := after.Truncate(time.Minute).Add(time.Minute)

	for {
		if !e.months.has(int(t.Month())) {
			t = time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location())
			continue
		}

		dayMatch := e.days.has(t.Day())
		weekdayMatch := e.weekdays.has(int(t.Weekday()))
		isDayValid := dayMatch || weekdayMatch

		if !isDayValid {
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).AddDate(0, 0, 1)
			continue
		}

		if !e.hours.has(t.Hour()) {
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour()+1, 0, 0, 0, t.Location())
			continue
		}

		if !e.minutes.has(t.Minute()) {
			t = t.Add(time.Minute)
			continue
		}

		return t
	}
}

// Parse parses a standard 5-field cron expression (minute hour day month weekday).
func Parse(expr string) (Expression, error) {
	parts := strings.Fields(expr)
	if len(parts) != 5 {
		return Expression{}, fmt.Errorf("invalid cron: expected 5 fields, got %d", len(parts))
	}

	var e Expression
	var err error

	if e.minutes, err = parseField(parts[0], 0, 59); err != nil {
		return e, fmt.Errorf("minute: %v", err)
	}
	if e.hours, err = parseField(parts[1], 0, 23); err != nil {
		return e, fmt.Errorf("hour: %v", err)
	}
	if e.days, err = parseField(parts[2], 1, 31); err != nil {
		return e, fmt.Errorf("day: %v", err)
	}
	if e.months, err = parseField(parts[3], 1, 12); err != nil {
		return e, fmt.Errorf("month: %v", err)
	}
	if e.weekdays, err = parseField(parts[4], 0, 6); err != nil {
		return e, fmt.Errorf("weekday: %v", err)
	}

	return e, nil
}

type fieldSet struct {
	all  bool
	vals map[int]bool
}

func (f fieldSet) has(v int) bool {
	if f.all {
		return true
	}
	return f.vals[v]
}

func parseField(s string, lo, hi int) (fieldSet, error) {
	if s == "*" {
		return fieldSet{all: true}, nil
	}

	vals := map[int]bool{}

	for _, part := range strings.Split(s, ",") {
		if strings.Contains(part, "/") {
			sub := strings.Split(part, "/")
			if len(sub) != 2 {
				return fieldSet{}, fmt.Errorf("invalid step: %q", part)
			}
			base, stepStr := sub[0], sub[1]
			step, err := strconv.Atoi(stepStr)
			if err != nil || step <= 0 {
				return fieldSet{}, fmt.Errorf("bad step %q", stepStr)
			}

			start, end := lo, hi
			if base != "*" {
				if strings.Contains(base, "-") {
					r := strings.Split(base, "-")
					if len(r) != 2 || r[0] == "" || r[1] == "" {
						return fieldSet{}, fmt.Errorf("bad range %q", base)
					}
					start, _ = strconv.Atoi(r[0])
					end, _ = strconv.Atoi(r[1])
				} else {
					start, _ = strconv.Atoi(base)
				}
			}
			for i := start; i <= end; i += step {
				vals[i] = true
			}
			continue
		}

		if strings.Contains(part, "-") {
			r := strings.Split(part, "-")
			if len(r) != 2 || r[0] == "" || r[1] == "" {
				return fieldSet{}, fmt.Errorf("bad range %q", part)
			}
			start, _ := strconv.Atoi(r[0])
			end, _ := strconv.Atoi(r[1])
			for i := start; i <= end; i++ {
				vals[i] = true
			}
			continue
		}

		v, err := strconv.Atoi(part)
		if err != nil {
			return fieldSet{}, fmt.Errorf("invalid number %q", part)
		}
		if v < lo || v > hi {
			return fieldSet{}, fmt.Errorf("value %d out of range [%d-%d]", v, lo, hi)
		}
		vals[v] = true
	}

	return fieldSet{vals: vals}, nil
}
