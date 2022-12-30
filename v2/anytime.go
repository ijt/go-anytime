package anytime

import (
	"fmt"
	"regexp"
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

var wordSpaceRx = regexp.MustCompile(`(\w+)[,.\s]*`)

func ReplaceAllRangesByFunc(s string, now time.Time, dir Direction, f func(src string, normSrc string, r Range) string) (string, error) {
	ls := strings.ToLower(s)
	var parts []string
	endOfPrevDate := 0
	p := 0
	for p < len(s) {
		// Find the next possible date.
		mightBeDateStart := isSignal(s[p]) && (p == 0 || !isSignal(s[p-1]))
		if !mightBeDateStart {
			p++
			continue
		}
		// eofw is the end of the first word.
		eofw := p
		for eofw < len(s) && isSignal(s[eofw]) {
			eofw++
		}
		// fw is the first word.
		fw := ls[p:eofw]
		r, ok := oneWordStrToRange(fw, now)
		if ok {
			parts = append(parts, s[endOfPrevDate:p])
			fr := f(s[p:eofw], fw, r)
			parts = append(parts, fr)
			endOfPrevDate = eofw
			p = eofw
			continue
		}

		// Try for things like "last week", "this month", "next year".
		if fw == "last" || fw == "this" || fw == "next" {
			// sosw is the start of the second word.
			sosw := eofw
			for sosw < len(s) && !isSignal(s[sosw]) {
				sosw++
			}
			// eosw is the end of the second word.
			eosw := sosw
			if eosw < len(s) && isSignal(s[eosw]) {
				eosw++
			}
			sw := ls[sosw:eosw]
			// fwsw could make an allocation. Let's get rid of it.
			fwsw := fw + " " + sw
			r, ok = lastThisNextStrToRange(fwsw, now)
			if ok {
				parts = append(parts, s[endOfPrevDate:p])
				fr := f(s[p:eosw], fw, r)
				parts = append(parts, fr)
				endOfPrevDate = eosw
				p = eosw
				continue
			}
		}

		// Try parsing a more general date...

		// Nothing found. Skip over the first word.
		p = eofw
	}
	parts = append(parts, s[endOfPrevDate:])
	return strings.Join(parts, ""), nil
}

func isSignal(b byte) bool {
	r := rune(b)
	return unicode.IsLetter(r) || unicode.IsDigit(r) || b == '/' || b == '-'
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

func normalizedThreeWordStrToRange(normSrc string, _ time.Time, _ Direction) (Range, bool) {
	t, err := time.Parse("January 2 2006", normSrc)
	if err == nil {
		return truncateDay(t), true
	}
	t, err = time.Parse("Jan 2 2006", normSrc)
	if err == nil {
		return truncateDay(t), true
	}
	return Range{}, false
}

// normalize fills the given buf with a normalized version of s: lower-cased
// and with each instance of [\s.,]+ replaced by a single space.
func normalize(buf []byte, s string) []byte {
	buf = buf[:0]
	for i := 0; i < len(s); i++ {
		if isSpacePunct(s[i]) {
			// Represent however much space or punctuation in this chunk
			// with a single space in the normalized version in buf.
			buf = append(buf, ' ')
			for i < len(s) && isSpacePunct(s[i]) {
				i++
			}
			i--
			continue
		}
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		buf = append(buf, c)
	}
	return buf
}

func isSpacePunct(c byte) bool {
	return unicode.IsSpace(rune(c)) || c == ',' || c == '.'
}
