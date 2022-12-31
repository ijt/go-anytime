package anytime

import (
	"math/rand"
	"strings"
	"testing"
	"time"
)

func Test_parseImplicitRange_monthOnly(t *testing.T) {
	now := time.UnixMilli(rand.Int63())
	tests := []struct {
		input string
		month time.Month
	}{
		{"january", time.January},
		{"January", time.January},
		{"Jan", time.January},
		{"February", time.February},
		{"Feb", time.February},
		{"March", time.March},
		{"Mar", time.March},
		{"April", time.April},
		{"Apr", time.April},
		{"May", time.May},
		{"June", time.June},
		{"Jun", time.June},
		{"July", time.July},
		{"Jul", time.July},
		{"August", time.August},
		{"Aug", time.August},
		{"September", time.September},
		{"Sep", time.September},
		{"October", time.October},
		{"Oct", time.October},
		{"November", time.November},
		{"Nov", time.November},
		{"December", time.December},
		{"Dec", time.December},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// future
			gotRange, parsed, err := parseImplicitRange(tt.input, strings.ToLower(tt.input), now, Future)
			if err != nil {
				t.Fatal(err)
			}
			wantRange := nextSpecificMonth(now, tt.month)
			if !gotRange.Equal(wantRange) {
				t.Errorf("future: got range %v, want %v", gotRange, wantRange)
			}
			if parsed != tt.input {
				t.Errorf("future: parsed %q, want %q", parsed, tt.input)
			}

			// past
			gotRange, parsed, err = parseImplicitRange(tt.input, strings.ToLower(tt.input), now, Past)
			if err != nil {
				t.Fatal(err)
			}
			wantRange = lastSpecificMonth(now, tt.month)
			if !gotRange.Equal(wantRange) {
				t.Errorf("past: got range %v, want %v", gotRange, wantRange)
			}
			if parsed != tt.input {
				t.Errorf("past: parsed %q, want %q", parsed, tt.input)
			}
		})
	}
}
