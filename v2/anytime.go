package anytime

import (
	"fmt"
	"regexp"
	"time"
)

//go:generate pigeon -o grammar.go grammar.peg

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

// RangeFromTimes returns a range given the start and end times.
func RangeFromTimes(start, end time.Time) Range {
	return Range{start, end.Sub(start)}
}

type Direction int

const (
	Future = iota
	Past
)

var candidateRx = regexp.MustCompile(`\b[a-zA-Z0-9]`)

func ReplaceAllRangesByFunc(s string, ref time.Time, f func(source string, r Range) string, dir Direction) (string, error) {
	indexes := candidateRx.FindAllStringIndex(s, -1)
	type stringWithPos struct {
		s string
		p int
	}
	var timeStrs []stringWithPos
	for len(indexes) > 0 {
		startEnd := indexes[0]
		start := startEnd[0]
		indexes = indexes[1:]
		filename := fmt.Sprintf("input string starting at position %v", start+1)
		locRangeAsAny, err := Parse(filename, []byte(s[start:]))
		if err != nil {
			break
		}
		locRange := locRangeAsAny.(LocatedRange)
		r := locRange.RangeFn(ref, dir)
		fr := f(string(locRange.Text), r)
		timeStrs = append(timeStrs, stringWithPos{fr, start})
	}

	return "", nil
}
