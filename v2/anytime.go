package anytime

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
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

func ReplaceAllRangesByFunc(s string, now time.Time, dir Direction, f func(src string, r Range) string) (string, error) {
	ls := strings.ToLower(s)
	var parts []string
	endOfPrevDate := 0
	p := 0
	for p < len(s) {
		// sofw is the start of the first word in s[p:].
		sofw := len(s)
		for i, roon := range s[p:] {
			if isSignal(roon) {
				sofw = p + i
				break
			}
		}
		if sofw == len(s) {
			break
		}

		// eofw is the end of the first word in s[p:]
		eofw := len(s)
		for i, roon := range s[sofw:] {
			if !isSignal(roon) {
				eofw = sofw + i
				break
			}
		}

		// fw is the first word.
		fw := ls[sofw:eofw]

		// Try for a match with "now", "today", etc.
		r, ok := oneWordStrToRange(fw, now)
		if ok {
			parts = append(parts, s[endOfPrevDate:sofw])
			fr := f(s[p:eofw], r)
			parts = append(parts, fr)
			endOfPrevDate = eofw
			p = eofw
			continue
		}

		// Try for a match with "last week", "this month", "next year", etc.
		if fw == "last" || fw == "this" || fw == "next" {
			// sosw is the start of the second word.
			sosw := len(s)
			for i, roon := range s[eofw:] {
				if isSignal(roon) {
					sosw = eofw + i
					break
				}
			}
			// eosw is the end of the second word.
			eosw := len(s)
			for i, roon := range s[sosw:] {
				if !isSignal(roon) {
					eosw = sosw + i
					break
				}
			}

			// sw is the second word in s[p:].
			sw := ls[sosw:eosw]

			// 🚨: fwsw could make an allocation.
			fwsw := fw + " " + sw

			r, ok = lastThisNextStrToRange(fwsw, now)
			if ok {
				parts = append(parts, s[endOfPrevDate:p])
				fr := f(s[p:eosw], r)
				parts = append(parts, fr)
				endOfPrevDate = eosw
				p = eosw
				continue
			}
		}

		// Try parsing a more general date...
		var d date
		// sow is the start of the current word.
		sow := sofw
		// eow is the end of the current word.
		eow := eofw
		for sow < len(s) {
			// w is the current word, lower-cased.
			w := ls[sow:eow]
			ok = parseDateWord(&d, w)
			if !ok {
				break
			}
			sow = len(s)
			for i, roon := range s[eow:] {
				if isSignal(roon) {
					sow = eow + i
					break
				}
			}
			eow = len(s)
			for i, roon := range s[sow:] {
				if !isSignal(roon) {
					eow = sow + i
					break
				}
			}
		}

		r, ok = inferRange(d, now, dir)
		if !ok {
			// Not enough information was given, so skip it.
			p = eow
			continue
		}

		// Got enough information to specify an implicit date range, so
		// append that to the result:
		// Add non-date stuff before the current date.
		parts = append(parts, s[endOfPrevDate:p])
		// Add the current date, mogrified by the user-provided f.
		fr := f(s[p:eow], r)
		parts = append(parts, fr)
		p = eow
	}
	parts = append(parts, s[endOfPrevDate:])
	return strings.Join(parts, ""), nil
}

var monthNameToMonth = map[string]time.Month{
	"jan": time.January,
	"feb": time.February,
	"mar": time.March,
	"apr": time.April,
	"may": time.May,
	"jun": time.June,
	"jul": time.July,
	"aug": time.August,
	"sep": time.September,
	"oct": time.October,
	"nov": time.November,
	"dec": time.December,

	"january":   time.January,
	"february":  time.February,
	"march":     time.March,
	"april":     time.April,
	"june":      time.June,
	"july":      time.July,
	"august":    time.August,
	"september": time.September,
	"october":   time.October,
	"november":  time.November,
	"december":  time.December,
}

// parseDateWord sets a field of d based on the given word w and returns
// true if it can. If no usable information is found, it returns false.
func parseDateWord(d *date, w string) bool {
	// Year
	if len(w) == 4 {
		y, err := strconv.Atoi(w)
		if err == nil && y >= 1000 && y <= 9999 {
			d.year = y
			return true
		}
	}

	// Day of month
	if len(w) == 1 || len(w) == 2 {
		dom, err := strconv.Atoi(w)
		if err == nil && dom >= 1 && dom <= 31 {
			d.dayOfMonth = dom
			return true
		}
	}

	// Month
	m, ok := monthNameToMonth[w]
	if ok {
		d.month = m
		return true
	}

	// Time zone like "utc+8"
	if (len(w) == len("utc+1") || len(w) == len("utc+10")) && w[:3] == "utc" {
		h, err := strconv.Atoi(w[3:])
		if err == nil && h >= -12 && h <= 12 {
			d.loc = fixedZone(h)
			return true
		}
	}

	return false
}

func inferRange(d date, now time.Time, dir Direction) (Range, bool) {
	if d.year == 0 && d.month == 0 {
		return Range{}, false
	}

	loc := d.loc
	if loc == nil {
		loc = now.Location()
	}

	// Infer the concrete implicit date range from the information given.
	switch {
	// Year month dayOfMonth
	case d.year != 0 && d.month != 0 && d.dayOfMonth != 0:
		return Range{
			time.Date(d.year, d.month, d.dayOfMonth, 0, 0, 0, 0, loc),
			time.Duration(24 * time.Hour),
		}, true

	// Year month
	case d.year != 0 && d.month != 0 && d.dayOfMonth == 0:
		s := time.Date(d.year, d.month, 1, 0, 0, 0, 0, loc)
		return truncateMonth(s), true

	// Year
	case d.year != 0 && d.month == 0 && d.dayOfMonth == 0:
		s := time.Date(d.year, 1, 1, 0, 0, 0, 0, loc)
		return truncateYear(s), true

	// Month dayOfMonth
	case d.year == 0 && d.month != 0 && d.dayOfMonth != 0:
		var r Range
		if dir == Future {
			r = nextSpecificMonth(now, d.month)
		} else {
			r = lastSpecificMonth(now, d.month)
		}
		s := r.start
		s2 := time.Date(s.Year(), d.month, d.dayOfMonth, 0, 0, 0, 0, loc)
		return truncateDay(s2), true

	// Month
	case d.year == 0 && d.month != 0 && d.dayOfMonth == 0:
		var r Range
		if dir == Future {
			r = nextSpecificMonth(now, d.month)
		} else {
			r = lastSpecificMonth(now, d.month)
		}
		s := r.start
		s2 := time.Date(s.Year(), d.month, 1, 0, 0, 0, 0, loc)
		return truncateMonth(s2), true

	default:
		return Range{}, false
	}
}

type date struct {
	year       int
	month      time.Month
	dayOfMonth int
	loc        *time.Location
}

func isSignal(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '/' || r == '-'
}

func oneWordStrToRange(normSrc string, now time.Time) (Range, bool) {
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

func lastThisNextStrToRange(normSrc string, now time.Time) (Range, bool) {
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

	t, err := time.Parse("January 2006", normSrc)
	if err == nil {
		return truncateMonth(t), true
	}
	t, err = time.Parse("Jan 2006", normSrc)
	if err == nil {
		return truncateMonth(t), true
	}

	return Range{}, false
}
