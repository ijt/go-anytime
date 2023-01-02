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

// TestParseAnyRange_fail tests parsing with inputs that are expected to fail.
func TestParseRange_fail(t *testing.T) {
	var badCases = []struct {
		input string
	}{
		{``},
		{`�`},
		{`a`},
		{`next thing`},
		{`not a date or a time`},
		{`Message me in 2 minutes`},
		{`Message me in 2 minutes from now`},
		{`Remind me in 1 hour`},
		{`Remind me in 1 hour from now`},
		{`Remind me in 1 hour and 3 minutes from now`},
		{`Remind me in an hour`},
		{`Remind me in an hour from now`},
		{`Remind me one day from now`},
		{`Remind me in a day`},
		{`Remind me in one day`},
		{`Remind me in one day from now`},
		{`Message me in a week`},
		{`Message me in one week`},
		{`Message me in one week from now`},
		{`Message me in two weeks from now`},
		{`Message me two weeks from now`},
		{`Message me in two weeks`},
		{`Remind me in 12 months from now at 6am`},
		{`Remind me in a month`},
		{`Remind me in 2 months`},
		{`Remind me in a month from now`},
		{`Remind me in 2 months from now`},
		{`Remind me in one year from now`},
		{`Remind me in a year`},
		{`Remind me in a year from now`},
		{`Restart the server in 2 days from now`},
		{`Remind me on the 5th of next month`},
		{`Remind me on the 5th of next month at 7am`},
		{`Remind me at 7am on the 5th of next month`},
		{`Remind me in one month from now`},
		{`Remind me in one month from now at 7am`},
		{`Remind me on the December 25th at 7am`},
		{`Remind me at 7am on December 25th`},
		{`Remind me on the 25th of December at 7am`},
		{`Check logs in the past 5 minutes`},

		// "1 minute" is a duration, not a time.
		{`1 minute`},

		// "one minute" is also a duration.
		{`one minute`},

		// "1 hour" is also a duration.
		{`1 hour`},

		// "1 day" is also a duration.
		{`1 day`},

		// "1 week" is also a duration.
		{`1 week`},

		// "1 month" is also a duration.
		{`1 month`},

		// "next 2 months" is a date range, not a time or a date.
		{`next 2 months`},

		// These are currently considered bad input, although they may
		{`10`},
		{`17`},

		// Bare years don't have enough context to be confidently parsed as dates.
		{`1999`},
		{`2008`},

		// Goofy input:
		{`10:am`},
	}
	for _, c := range badCases {
		t.Run(c.input, func(t *testing.T) {
			lower := strings.ToLower(c.input)
			_, _, err := ParseRange(c.input, lower, time.Time{}, Future)
			if err == nil {
				t.Error("parsing succeeded, want failure")
			}
		})
	}
}
