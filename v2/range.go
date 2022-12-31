package anytime

import (
	"fmt"
	"time"
)

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

// RangeFromTimes returns a Range from two times.
func RangeFromTimes(start, end time.Time) Range {
	return Range{start, end.Sub(start)}
}
