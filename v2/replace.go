package anytime

import (
	"strings"
	"time"
)

type Direction int

const (
	Future = iota
	Past
)

// ReplaceAllRangesByFunc replaces all dates and date ranges within the string
// s by the result of calling f on the parsed date/range and the substring
// src that gave rise to it.
//
// Ranges include things like "this year". That is the range from Jan 1 to Dec
// 31 of the current year.
//
// Ranges also include things like "from last year to next year" or
// "last year to next year".
//
// In ambiguous cases like "December" that could be in the past or the future,
// the dir argument tells whether to choose the Past or Future instance of it.
func ReplaceAllRangesByFunc(s string, now time.Time, dir Direction, f func(src string, r Range) string) (string, error) {
	ls := strings.ToLower(s)
	var parts []string
	endOfPrevDate := 0
	p := 0
	for p < len(s) {
		sofw := findNextSignal(s, p)
		r, parsed, err := ParseRange(s[sofw:], ls[sofw:], now, dir)
		if err != nil {
			eofw := findNextNoise(s, sofw)
			p = eofw
			continue
		}
		parts = append(parts, s[endOfPrevDate:sofw])
		fr := f(parsed, r)
		parts = append(parts, fr)
		p = sofw + len(parsed)
		endOfPrevDate = p
	}
	parts = append(parts, s[endOfPrevDate:])
	return strings.Join(parts, ""), nil
}
