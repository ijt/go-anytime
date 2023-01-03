package anytime

import "unicode/utf8"

// fixUTF8 salvages the valid UTF8 runes in the given string.
// https://gist.github.com/scbizu/23a7d3f3436a35ab3b79c3cce91ca8fe
func fixUTF8(s string) string {
	if utf8.Valid([]byte(s)) {
		return s
	}
	rs := make([]rune, 0, len(s))
	for i, r := range s {
		if r == utf8.RuneError {
			_, size := utf8.DecodeRuneInString(s[i:])
			if size == 1 {
				continue
			}
		}
		rs = append(rs, r)
	}
	return string(rs)
}
