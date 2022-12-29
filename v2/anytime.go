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

var wordsRx = regexp.MustCompile(`\b\w+(\s*\w+)*\b`)
var twoWordRx = regexp.MustCompile(`\b(\w+)\s+(\w+)\b`)
var oneWordRx = regexp.MustCompile(`\b(\w+)\b`)
var wordSpaceRx = regexp.MustCompile(`^\w+\s*`)

func ReplaceAllRangesByFunc(inputStr string, now time.Time, dir Direction, f func(src string, normSrc string, r Range) string) (string, error) {
	inputStr = wordsRx.ReplaceAllStringFunc(inputStr, func(s string) string {
		replaceTwoWordString := func(s string) string {
			ns := normalize(s)
			r, ok := normalizedTwoWordStrToRange(ns, now, dir)
			if !ok {
				return s
			}
			return f(s, ns, r)
		}
		s = twoWordRx.ReplaceAllStringFunc(s, replaceTwoWordString)

		// Handle the odd pairs. If the string is like "a b c" then the
		// first call to twoWordRx.ReplaceAllStringFunc will replace "a b"
		// and leave "c" alone. So we need to call it again on "b c".
		loc := wordSpaceRx.FindStringIndex(s)
		if loc == nil {
			return s
		}
		s2 := twoWordRx.ReplaceAllStringFunc(s[loc[1]:], replaceTwoWordString)
		return s[:loc[1]] + s2
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

	// last $longMonth
	case "last january":
		return lastSpecificMonth(now, time.January), true
	case "last february":
		return lastSpecificMonth(now, time.February), true
	case "last march":
		return lastSpecificMonth(now, time.March), true
	case "last april":
		return lastSpecificMonth(now, time.April), true
	case "last may":
		return lastSpecificMonth(now, time.May), true
	case "last june":
		return lastSpecificMonth(now, time.June), true
	case "last july":
		return lastSpecificMonth(now, time.July), true
	case "last august":
		return lastSpecificMonth(now, time.August), true
	case "last september":
		return lastSpecificMonth(now, time.September), true
	case "last october":
		return lastSpecificMonth(now, time.October), true
	case "last november":
		return lastSpecificMonth(now, time.November), true
	case "last december":
		return lastSpecificMonth(now, time.December), true

	// last $shortMonth
	case "last jan":
		return lastSpecificMonth(now, time.January), true
	case "last feb":
		return lastSpecificMonth(now, time.February), true
	case "last mar":
		return lastSpecificMonth(now, time.March), true
	case "last apr":
		return lastSpecificMonth(now, time.April), true
	case "last jun":
		return lastSpecificMonth(now, time.June), true
	case "last jul":
		return lastSpecificMonth(now, time.July), true
	case "last aug":
		return lastSpecificMonth(now, time.August), true
	case "last sep":
		return lastSpecificMonth(now, time.September), true
	case "last oct":
		return lastSpecificMonth(now, time.October), true
	case "last nov":
		return lastSpecificMonth(now, time.November), true
	case "last dec":
		return lastSpecificMonth(now, time.December), true

	// next $longMonth
	case "next january":
		return nextSpecificMonth(now, time.January), true
	case "next february":
		return nextSpecificMonth(now, time.February), true
	case "next march":
		return nextSpecificMonth(now, time.March), true
	case "next april":
		return nextSpecificMonth(now, time.April), true
	case "next may":
		return nextSpecificMonth(now, time.May), true
	case "next june":
		return nextSpecificMonth(now, time.June), true
	case "next july":
		return nextSpecificMonth(now, time.July), true
	case "next august":
		return nextSpecificMonth(now, time.August), true
	case "next september":
		return nextSpecificMonth(now, time.September), true
	case "next october":
		return nextSpecificMonth(now, time.October), true
	case "next november":
		return nextSpecificMonth(now, time.November), true
	case "next december":
		return nextSpecificMonth(now, time.December), true

	// next $shortMonth
	case "next jan":
		return nextSpecificMonth(now, time.January), true
	case "next feb":
		return nextSpecificMonth(now, time.February), true
	case "next mar":
		return nextSpecificMonth(now, time.March), true
	case "next apr":
		return nextSpecificMonth(now, time.April), true
	case "next jun":
		return nextSpecificMonth(now, time.June), true
	case "next jul":
		return nextSpecificMonth(now, time.July), true
	case "next aug":
		return nextSpecificMonth(now, time.August), true
	case "next sep":
		return nextSpecificMonth(now, time.September), true
	case "next oct":
		return nextSpecificMonth(now, time.October), true
	case "next nov":
		return nextSpecificMonth(now, time.November), true
	case "next dec":
		return nextSpecificMonth(now, time.December), true

	}
	return Range{}, false
}

func normalize(s string) string {
	fs := strings.Fields(s)
	s = strings.Join(fs, " ")
	s = strings.ToLower(s)
	return s
}
