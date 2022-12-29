package anytime

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// LocatedRange is a Range with information about where it was found within the
// input string.
type LocatedRange struct {
	// RangeFn returns the time range found in the input text.
	RangeFn RangeFunc

	// Pos is the position of the range in the input text.
	Pos int

	// Text is the text that gave rise to the range (e.g. "tomorrow").
	Text []byte
}

// RangeFunc is a function that can make a Range, given a reference time
// and a default direction.
type RangeFunc func(ref time.Time, dir Direction) Range

// Range is a time range as a half-open interval.
type Range struct {
	start    time.Time
	Duration time.Duration
}

// Start is when the range begins, inclusive.
func (r Range) Start() time.Time {
	return r.start
}

// End is when the range ends, exclusive.
func (r Range) End() time.Time {
	return r.start.Add(r.Duration)
}

// EndInclusiveDay returns the beginning of the last day of the range.
// The range is assumed to be long enough for that to make sense.
func (r Range) EndInclusiveDay() time.Time {
	return r.End().Add(-24 * time.Hour)
}

// String returns a string with the time and duration of the range.
func (r Range) String() string {
	return fmt.Sprintf("{start: %v, duration: %v}", r.start, r.Duration)
}

// Equal returns true if the two ranges are equal.
func (r Range) Equal(other Range) bool {
	return r.start.Equal(other.start) && r.Duration == other.Duration
}

type Direction int

const (
	Future = iota
	Past
)

var twoWordRx = regexp.MustCompile(`\b(\w+)\s+(\w+)\b`)
var oneWordRx = regexp.MustCompile(`\b(\w+)\b`)

func ReplaceAllRangesByFunc(inputStr string, now time.Time, f func(src, normSrc string, r Range) string, dir Direction) (string, error) {
	inputStr = twoWordRx.ReplaceAllStringFunc(inputStr, func(s string) string {
		ns := normalize(s)
		r, ok := normalizedTwoWordStrToRange(ns, now, dir)
		if !ok {
			return s
		}
		return f(s, ns, r)
	})
	inputStr = oneWordRx.ReplaceAllStringFunc(inputStr, func(s string) string {
		ns := strings.ToLower(s)
		r, ok := normalizedOneWordStrToRange(ns, now, dir)
		if !ok {
			return s
		}
		return f(s, ns, r)
	})
	return inputStr, nil
}

func normalizedOneWordStrToRange(normSrc string, now time.Time, _ Direction) (Range, bool) {
	switch normSrc {
	case "now":
		return Range{now, time.Second}, true
	case "yesterday":
		return truncateDay(now.AddDate(0, 0, -1)), true
	case "today":
		return truncateDay(now), true
	case "tomorrow":
		return truncateDay(now.AddDate(0, 0, 1)), true
	}
	return Range{}, false
}

func normalizedTwoWordStrToRange(normSrc string, now time.Time, _ Direction) (Range, bool) {
	switch normSrc {
	case "last week":
		return truncateWeek(now.AddDate(0, 0, -7)), true
	case "this week":
		return truncateWeek(now), true
	case "next week":
		return truncateWeek(now.AddDate(0, 0, 7)), true
	case "last month":
		return truncateMonth(now.AddDate(0, -1, 0)), true
	case "this month":
		return truncateMonth(now), true
	case "next month":
		return truncateMonth(now.AddDate(0, 1, 0)), true
	case "last year":
		return truncateYear(now.AddDate(-1, 0, 0)), true
	case "this year":
		return truncateYear(now), true
	case "next year":
		return truncateYear(now.AddDate(1, 0, 0)), true
	case "last january":
		return lastSpecificMonth(now, time.January), true
	case "last jan":
		return lastSpecificMonth(now, time.January), true
	case "next january":
		return nextSpecificMonth(now, time.January), true
	case "next jan":
		return nextSpecificMonth(now, time.January), true
	}
	return Range{}, false
}

func normalize(s string) string {
	fs := strings.Fields(s)
	s = strings.Join(fs, " ")
	s = strings.ToLower(s)
	return s
}
