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

var dateTimeRx = regexp.MustCompile(
	`(?i)\bnow|yesterday|today|tomorrow|(last|this|next) (week|month|year)\b`)

func ReplaceAllRangesByFunc(inputStr string, now time.Time, f func(src, normSrc string, r Range) string, dir Direction) (string, error) {
	var errStrs []string
	s2 := dateTimeRx.ReplaceAllStringFunc(inputStr, func(src string) string {
		normSrc := normalize(src)
		r, err := normalizedStrToRange(normSrc, now, dir)
		if err != nil {
			errStrs = append(errStrs, err.Error())
			return ""
		}
		return f(src, normSrc, r)
	})
	if len(errStrs) > 0 {
		return "", fmt.Errorf(strings.Join(errStrs, ", "))
	}
	return s2, nil
}

func normalizedStrToRange(normSrc string, now time.Time, _ Direction) (Range, error) {
	switch normSrc {
	case "now":
		return Range{now, time.Second}, nil
	case "yesterday":
		return truncateDay(now.AddDate(0, 0, -1)), nil
	case "today":
		return truncateDay(now), nil
	case "tomorrow":
		return truncateDay(now.AddDate(0, 0, 1)), nil
	case "last week":
		return truncateWeek(now.AddDate(0, 0, -7)), nil
	case "this week":
		return truncateWeek(now), nil
	case "next week":
		return truncateWeek(now.AddDate(0, 0, 7)), nil
	case "last month":
		return truncateMonth(now.AddDate(0, -1, 0)), nil
	case "this month":
		return truncateMonth(now), nil
	case "next month":
		return truncateMonth(now.AddDate(0, 1, 0)), nil
	}
	return Range{}, fmt.Errorf("unrecognized date/time %q", normSrc)
}

var spaceRx = regexp.MustCompile(`\s+`)

func normalize(s string) string {
	s = spaceRx.ReplaceAllString(s, " ")
	s = strings.ToLower(s)
	return s
}
