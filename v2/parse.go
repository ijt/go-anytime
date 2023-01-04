package anytime

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var ErrNoRangeStartFound = errors.New("no start of range found just after `from`")
var ErrNoRangeEndFound = errors.New("no end of range found just after `to` or similar")

type ErrNoConnectorFound struct {
	ParsedStart    string
	WordAfterStart string
}

func (e *ErrNoConnectorFound) Error() string {
	return fmt.Sprintf("expected 'to|until|til|through' after %q, got %q", e.ParsedStart, e.WordAfterStart)
}

var ErrNoRangeFound = errors.New("no range found")
var ErrNoImplicitRangeFound = errors.New("no implicit range found")

// ParseRange parses either an explicit range or an implicit range starting
// at the beginning of s. A lower-cased version of s is given as ls.
func ParseRange(s string, now time.Time, dir Direction) (r Range, parsed string, err error) {
	eow1 := findNextNoise(s, 0)
	w1 := s[:eow1]

	// "from A to B" for implicit ranges A and B:
	if eq(w1, "from") {
		sow2 := findNextSignal(s, eow1)
		startRange, parsedStart, err := parseImplicitRange(s[sow2:], now, dir)
		if err != nil {
			return Range{}, "", ErrNoRangeStartFound
		}
		eoStart := sow2 + len(parsedStart)
		_, eoto, to := findSignalNoise(s, eoStart)
		if !isConnector(to) {
			return Range{}, "", &ErrNoConnectorFound{parsedStart, to}
		}
		soEnd := findNextSignal(s, eoto)
		endRange, parsedEnd, err := parseImplicitRange(s[soEnd:], now, dir)
		if err != nil {
			return Range{}, "", ErrNoRangeEndFound
		}
		r := Range{startRange.Start(), endRange.End().Sub(startRange.Start())}
		eoEnd := soEnd + len(parsedEnd)
		return r, s[:eoEnd], nil
	}

	// Either "A" or "A to B":
	r, parsed, err = parseImplicitRange(s, now, dir)
	if err != nil {
		return Range{}, "", ErrNoImplicitRangeFound
	}
	eor := len(parsed)
	_, eoto, to := findSignalNoise(s, eor)
	if !isConnector(to) {
		return r, parsed, nil
	}
	soEnd := findNextSignal(s, eoto)
	endRange, parsedEnd, err := parseImplicitRange(s[soEnd:], now, dir)
	if err != nil {
		// If we can't parse the end of the range, we'll just return the
		// start of the range.
		return r, parsed, nil
	}
	r = Range{r.Start(), endRange.End().Sub(r.Start())}
	eoEnd := soEnd + len(parsedEnd)
	return r, s[:eoEnd], nil
}

func isConnector(s string) bool {
	return s == "to" || s == "until" || s == "til" || s == "through"
}

func eq(a, b string) bool {
	return strings.EqualFold(a, b)
}

// parseImplicitRange parses an implicit date range from a string s.
// An implicit date range is something like "2022ad" as opposed to
// an explicit range like "from 2020ad to 2022ad".
//
// The lower-cased version of s is given as ls. The prefix of s that was parsed
// is also returned. If no range is found at the very beginning of s,
// ErrNoRangeFound is returned.
func parseImplicitRange(s string, now time.Time, dir Direction) (r Range, parsed string, err error) {
	// sofw is the start of the first word in s[p:].
	// eofw is the end of the first word in s[p:]
	// fw is the first word.
	sofw, eofw, fw := findSignalNoise(s, 0)
	if sofw == len(s) {
		return Range{}, "", ErrNoRangeFound
	}

	// Try for a match with "now", "today", etc.
	r, ok := oneWordStrToRange(fw, now)
	if ok {
		return r, s[sofw:eofw], nil
	}

	// Try for a match with "last week", "this month", "next year", etc.
	if eq(fw, "last") || eq(fw, "this") || eq(fw, "next") {
		// sosw is the start of the second word.
		// eosw is the end of the second word.
		// sw is the second word in s[p:].
		_, eosw, sw := findSignalNoise(s, eofw)
		fwsw := fw + " " + sw
		r, ok = lastThisNextStrToRange(fwsw, now)
		if ok {
			return r, s[sofw:eosw], nil
		}
	}

	// Try for a match with
	// "N days ago", "N days from now",
	// "N weeks ago"
	// "N months ago"
	// "N years ago"
	// etc.
	if i, ok := parseInt(fw); ok {
		_, eow2, w2 := findSignalNoise(s, eofw)
		_, eow3, w3 := findSignalNoise(s, eow2)
		_, eow4, w4 := findSignalNoise(s, eow3)
		if i >= 1000 && i <= 9999 && (eq(w2, "ad") || eq(w2, "ce")) {
			// Year
			r := truncateYear(time.Date(i, 1, 1, 0, 0, 0, 0, now.Location()))
			return r, s[sofw:eow2], nil
		}
		if eq(w2, "day") || eq(w2, "days") {
			if eq(w3, "ago") {
				r := truncateDay(now.AddDate(0, 0, -i))
				return r, s[sofw:eow3], nil
			}
			if eq(w3, "hence") {
				r := truncateDay(now.AddDate(0, 0, i))
				return r, s[sofw:eow3], nil
			}
			if eq(w3, "from") && (eq(w4, "now") || eq(w4, "today")) {
				r := truncateDay(now.AddDate(0, 0, i))
				return r, s[sofw:eow4], nil
			}
		}
		if eq(w2, "week") || eq(w2, "weeks") {
			if w3 == "ago" {
				r := truncateWeek(now.AddDate(0, 0, -i*7))
				return r, s[sofw:eow3], nil
			}
			if eq(w3, "hence") {
				r := truncateWeek(now.AddDate(0, 0, i*7))
				return r, s[sofw:eow3], nil
			}
			if eq(w3, "from") && (eq(w4, "now") || eq(w4, "today")) {
				r := truncateWeek(now.AddDate(0, 0, i*7))
				return r, s[sofw:eow4], nil
			}
		}
		if eq(w2, "month") || eq(w2, "months") {
			if eq(w3, "ago") {
				r := truncateMonth(now.AddDate(0, -i, 0))
				return r, s[sofw:eow3], nil
			}
			if eq(w3, "hence") {
				r := truncateMonth(now.AddDate(0, i, 0))
				return r, s[sofw:eow3], nil
			}
			if eq(w3, "from") && (eq(w4, "now") || eq(w4, "today")) {
				r := truncateMonth(now.AddDate(0, i, 0))
				return r, s[sofw:eow4], nil
			}
		}
		if eq(w2, "year") || eq(w2, "years") {
			if eq(w3, "ago") {
				r := truncateYear(now.AddDate(-i, 0, 0))
				return r, s[sofw:eow3], nil
			}
			if eq(w3, "hence") {
				r := truncateYear(now.AddDate(i, 0, 0))
				return r, s[sofw:eow3], nil
			}
			if eq(w3, "from") && (eq(w4, "now") || eq(w4, "today")) {
				r := truncateYear(now.AddDate(i, 0, 0))
				return r, s[sofw:eow4], nil
			}
		}
	}

	// Try for a match with "green october", "blue june", etc.
	if delta, ok := colorToDelta[fw]; ok {
		_, eosw, sw := findSignalNoise(s, eofw)
		r, ok = colorMonthToRange(delta, sw, now)
		if ok {
			return r, s[sofw:eosw], nil
		}
	}

	// Check for RFC3339 format.
	if rfc3339Rx.MatchString(s[sofw:eofw]) {
		t, err := time.Parse(time.RFC3339, s[sofw:eofw])
		if err == nil {
			r := Range{t, time.Second}
			return r, s[sofw:eofw], nil
		}
	}

	// Try parsing a more general, multi-word date...
	var d date
	// sow is the start of the current word.
	sow := sofw
	// eow is the end of the current word.
	eow := eofw
	// w is the current word, lower-cased.
	w := fw
	// eolgw is the end of the last good word, i.e., the end of the
	// last word that was successfully parsed.
	var eolgw int
	code := ""
	for sow < len(s) {
		prevD := d
		wCode, ok := parseDateWord(&d, w)
		if !ok {
			d = prevD
			break
		}
		// If the date word just parsed is one of the things already parsed
		// for this date then don't accept it and stop parsing.
		if strings.ContainsAny(code, wCode) {
			d = prevD
			break
		}
		// "d" has to come either at the beginning or after a month name,
		// or we ignore it and consider the implicit range to have ended
		// one word before.
		if wCode == "d" && !(len(code) == 0 || code[len(code)-1] == 'm') {
			d = prevD
			break
		}
		code = code + wCode
		eolgw = eow
		sow, eow, w = findSignalNoise(s, eow)
	}

	if strings.HasPrefix(code, "d") && !strings.HasPrefix(code, "dm") {
		// The only valid thing that can come after a day of month is a specific
		// month. Also, a day of month by itself is not enough to be
		// unambiguous. So bail.
		return Range{}, "", ErrNoRangeFound
	}

	r, ok = inferRange(d, now, dir, strings.ToLower(s[sofw:eolgw]))
	if !ok {
		// Not enough information was given, so skip it.
		return Range{}, "", ErrNoRangeFound
	}

	// Got enough information to specify an implicit date range.
	return r, s[sofw:eolgw], nil
}

// parseDateWord sets a field of d based on the given word w and returns
// true if it can. If no usable information is found, it returns false.
// It also returns a string signifying which type of thing was found:
// "y" for year, "m" for month, "d" for day, "ymd" for "YYYY/MM/DD" etc.
func parseDateWord(d *date, w string) (string, bool) {
	// Year
	if len(w) == 4 {
		y, err := strconv.Atoi(w)
		if err == nil && y >= 1000 && y <= 9999 {
			d.year = y
			return "y", true
		}
	}

	// Day of month
	if dom, ok := parseDayOfMonth(w); ok {
		d.dayOfMonth = dom
		return "d", true
	}

	// YYYY/MM/DD
	if sm := ymdRx.FindStringSubmatch(w); sm != nil {
		y, _ := strconv.Atoi(sm[1])
		m, _ := strconv.Atoi(sm[2])
		dom, _ := strconv.Atoi(sm[3])
		d.year = y
		d.month = time.Month(m)
		d.dayOfMonth = dom
		return "ymd", true
	}

	// DD/MM/YYYY
	if sm := dmyRx.FindStringSubmatch(w); sm != nil {
		dom, _ := strconv.Atoi(sm[1])
		m, _ := strconv.Atoi(sm[2])
		y, _ := strconv.Atoi(sm[3])
		d.year = y
		d.month = time.Month(m)
		d.dayOfMonth = dom
		return "dmy", true
	}

	// Month
	m, ok := monthNameToMonth[w]
	if ok {
		d.month = m
		return "m", true
	}

	// UTC time zone
	if w == "utc" {
		d.loc = time.UTC
		return "z", true
	}

	// Time zone like "utc+8"
	if (len(w) == len("utc+1") || len(w) == len("utc+10")) && w[:3] == "utc" {
		h, err := strconv.Atoi(w[3:])
		if err == nil && h >= -12 && h <= 12 {
			d.loc = fixedZone(h)
			return "z", true
		}
	}

	// 1999AD
	if len(w) == len("1999ad") && (w[4:] == "ad" || w[4:] == "ce") {
		y, err := strconv.Atoi(w[:4])
		if err == nil && y >= 1000 && y <= 9999 {
			d.year = y
			return "y", true
		}
	}

	return "", false
}

func parseDayOfMonth(w string) (int, bool) {
	i, ok := strToInt[w]
	if ok {
		if !okDayOfMonth(i) {
			return 0, false
		}
		return i, true
	}
	w = strings.TrimSuffix(w, "st")
	w = strings.TrimSuffix(w, "nd")
	w = strings.TrimSuffix(w, "rd")
	w = strings.TrimSuffix(w, "th")
	i, err := strconv.Atoi(w)
	if err == nil && okDayOfMonth(i) {
		return i, true
	}
	return 0, false
}

func okDayOfMonth(dom int) bool {
	return 1 <= dom && dom <= 31
}

func inferRange(d date, now time.Time, dir Direction, src string) (Range, bool) {
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
			24 * time.Hour,
		}, true

	// Year month
	case d.year != 0 && d.month != 0 && d.dayOfMonth == 0:
		s := time.Date(d.year, d.month, 1, 0, 0, 0, 0, loc)
		return truncateMonth(s), true

	// Year
	case d.year != 0 && d.month == 0 && d.dayOfMonth == 0 && (strings.HasSuffix(src, "ad") || strings.HasSuffix(src, "ce")):
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

func parseInt(s string) (int, bool) {
	i, ok := strToInt[s]
	if ok {
		return i, true
	}
	i, err := strconv.Atoi(s)
	if err == nil {
		return i, true
	}
	return 0, false
}

var strToInt = map[string]int{
	"a":         1,
	"one":       1,
	"two":       2,
	"three":     3,
	"four":      4,
	"five":      5,
	"six":       6,
	"seven":     7,
	"eight":     8,
	"nine":      9,
	"ten":       10,
	"eleven":    11,
	"twelve":    12,
	"thirteen":  13,
	"fourteen":  14,
	"fifteen":   15,
	"sixteen":   16,
	"seventeen": 17,
	"eighteen":  18,
	"nineteen":  19,
	"twenty":    20,
	"thirty":    30,
	"forty":     40,
	"fifty":     50,
	"sixty":     60,
	"seventy":   70,
	"eighty":    80,
	"ninety":    90,
}

func colorMonthToRange(colorDelta int, monthName string, now time.Time) (Range, bool) {
	m, ok := monthNameToMonth[monthName]
	if !ok {
		return Range{}, false
	}
	return truncateMonth(nextSpecificMonth(now, m).Start().AddDate(colorDelta, 0, 0)), true
}

var colorToDelta = map[string]int{
	"white":  0,
	"red":    1,
	"green":  2,
	"blue":   3,
	"gold":   4,
	"purple": 5,
	"orange": 6,
	"pink":   7,
	"silver": 8,
	"copper": 9,
}

func findSignalNoise(s string, start int) (int, int, string) {
	isig := findNextSignal(s, start)
	inoi := findNextNoise(s, isig)
	return isig, inoi, strings.ToLower(s[isig:inoi])
}

func findNextSignal(s string, start int) int {
	for i, roon := range s[start:] {
		if isSignal(roon) {
			return start + i
		}
	}
	return len(s)
}

func findNextNoise(s string, start int) int {
	for i, roon := range s[start:] {
		if !isSignal(roon) {
			return start + i
		}
	}
	return len(s)
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

var ymdRx = regexp.MustCompile(`(\d{4})[-/](\d{1,2})[-/](\d{1,2})`)
var dmyRx = regexp.MustCompile(`(\d{1,2})[-/](\d{1,2})[-/](\d{4})`)

func isSignal(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '/' || r == '-' || r == '+' || r == ':'
}

func oneWordStrToRange(w string, now time.Time) (Range, bool) {
	switch {
	case eq(w, "now"):
		return Range{now, time.Second}, true
	case eq(w, "yesterday"):
		return truncateDay(now.AddDate(0, 0, -1)), true
	case eq(w, "today"):
		return truncateDay(now), true
	case eq(w, "tomorrow"):
		return truncateDay(now.AddDate(0, 0, 1)), true
	}
	return Range{}, false
}

func lastThisNextStrToRange(w string, now time.Time) (Range, bool) {
	switch {
	case eq(w, "last week"):
		return truncateWeek(now.AddDate(0, 0, -7)), true
	case eq(w, "this week"):
		return truncateWeek(now), true
	case eq(w, "next week"):
		return truncateWeek(now.AddDate(0, 0, 7)), true
	case eq(w, "last month"):
		return truncateMonth(now.AddDate(0, -1, 0)), true
	case eq(w, "this month"):
		return truncateMonth(now), true
	case eq(w, "next month"):
		return truncateMonth(now.AddDate(0, 1, 0)), true
	case eq(w, "last year"):
		return truncateYear(now.AddDate(-1, 0, 0)), true
	case eq(w, "this year"):
		return truncateYear(now), true
	case eq(w, "next year"):
		return truncateYear(now.AddDate(1, 0, 0)), true
	}

	return Range{}, false
}

var rfc3339Rx = regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(?:\.(\d{1,9}))?(Z|([+-])(\d{2}):(\d{2}))?$`)
