package anytime

import (
	"fmt"
	"time"
)

//go:generate pigeon -o grammar.go grammar.peg

func toIfaceSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	return v.([]interface{})
}

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

type opts struct {
	defaultDirection Direction
}

// DefaultToFuture sets the option to default to the Future in case of
// ambiguous dates.
func DefaultToFuture(o *opts) {
	o.defaultDirection = Future
}

// DefaultToPast sets the option to default to the past in case of
// ambiguous dates.
func DefaultToPast(o *opts) {
	o.defaultDirection = Past
}
