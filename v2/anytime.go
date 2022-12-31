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

func ReplaceAllRangesByFunc(s string, now time.Time, dir Direction, f func(src string, r Range) string) (string, error) {
	ls := strings.ToLower(s)
	var parts []string
	endOfPrevDate := 0
	p := 0
	for p < len(s) {
		sofw := findNextSignal(s, p)
		r, parsed, err := parseAnyRange(s[sofw:], ls[sofw:], now, dir)
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
