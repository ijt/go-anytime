package anytime

import (
	"math/rand"
	"reflect"
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
			gotRange, parsed, err := parseImplicitRange(tt.input, now, Future)
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
			gotRange, parsed, err = parseImplicitRange(tt.input, now, Past)
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

func Test_parseImplicitRange(t *testing.T) {
	type args struct {
		s   string
		now time.Time
		dir Direction
	}
	tests := []struct {
		name       string
		args       args
		wantR      Range
		wantParsed string
		wantErr    bool
	}{
		{
			name: "small int after year is ignored",
			args: args{
				s: "Jan 2017 1",
			},
			wantR:      truncateMonth(time.Date(2017, time.January, 1, 0, 0, 0, 0, time.UTC)),
			wantParsed: "Jan 2017",
		},
		{
			name: "small int before year at beginning causes failure",
			args: args{
				s: "1 2017 Jan",
			},
			wantErr: true,
		},
		{
			name: "oct nov",
			args: args{
				s:   "oct nov",
				now: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantR:      truncateMonth(time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)),
			wantParsed: "oct",
		},
		{
			name: "mar 1st",
			args: args{
				s:   "mar 1st",
				now: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantR:      truncateDay(time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC)),
			wantParsed: "mar 1st",
		},
		{
			name: "mar 2nd",
			args: args{
				s:   "mar 2nd",
				now: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantR:      truncateDay(time.Date(2023, 3, 2, 0, 0, 0, 0, time.UTC)),
			wantParsed: "mar 2nd",
		},
		{
			name: "mar 3rd",
			args: args{
				s:   "mar 3rd",
				now: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantR:      truncateDay(time.Date(2023, 3, 3, 0, 0, 0, 0, time.UTC)),
			wantParsed: "mar 3rd",
		},
		{
			name: "mar 4th",
			args: args{
				s:   "mar 4th",
				now: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantR:      truncateDay(time.Date(2023, 3, 4, 0, 0, 0, 0, time.UTC)),
			wantParsed: "mar 4th",
		},
		{
			name: "md date with dash",
			args: args{
				s:   "mar-24",
				now: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantR:      truncateDay(time.Date(2023, 3, 24, 0, 0, 0, 0, time.UTC)),
			wantParsed: "mar-24",
		},
		{
			name: "md date with dash, now month after mar",
			args: args{
				s:   "mar-24",
				now: time.Date(2022, 11, 1, 0, 0, 0, 0, time.UTC),
			},
			wantR:      truncateDay(time.Date(2023, 3, 24, 0, 0, 0, 0, time.UTC)),
			wantParsed: "mar-24",
		},
		{
			name: "december 40",
			args: args{
				s:   "december 40",
				now: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
				dir: Future,
			},
			wantR:      truncateMonth(time.Date(2022, 12, 1, 0, 0, 0, 0, time.UTC)),
			wantParsed: "december",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotR, gotParsed, err := parseImplicitRange(tt.args.s, tt.args.now, tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseImplicitRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotR, tt.wantR) {
				t.Errorf("parseImplicitRange() gotR = %v, want %v", gotR, tt.wantR)
			}
			if gotParsed != tt.wantParsed {
				t.Errorf("parseImplicitRange() gotParsed = %v, want %v", gotParsed, tt.wantParsed)
			}
		})
	}
}

func TestParseRange_fail(t *testing.T) {
	var badCases = []struct {
		input string
	}{
		{``},
		{`�`},
		{`a`},
		{`from`},
		{`from foo`},
		{`from yesterday`},
		{`from yesterday to`},
		{`from yesterday to beeblebrox`},
		{`red raspberry`},
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
			_, _, err := ParseRange(c.input, time.Time{}, Future)
			if err == nil {
				t.Error("parsing succeeded, want failure")
			}
		})
	}
}

func TestParseRange(t *testing.T) {
	now := time.UnixMilli(rand.Int63())
	type args struct {
		s   string
		now time.Time
		dir Direction
	}
	tests := []struct {
		name       string
		args       args
		wantR      Range
		wantParsed string
		wantErr    bool
	}{
		{
			name: "NOW TO evaluates to NOW",
			args: args{
				s:   "now to",
				now: now,
			},
			wantR: Range{
				start:    now,
				Duration: time.Second,
			},
			wantParsed: "now",
		},
		{
			name: "NOW TO BAR evaluates to NOW",
			args: args{
				s:   "now to bar",
				now: now,
			},
			wantR: Range{
				start:    now,
				Duration: time.Second,
			},
			wantParsed: "now",
		},
		{
			name: "from april to may",
			args: args{
				s:   "from april to may",
				now: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
				dir: Future,
			},
			wantR: RangeFromTimes(
				time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2022, 5, 1, 0, 0, 0, 0, time.UTC),
			),
			wantParsed: "from april to may",
			wantErr:    false,
		},
		{
			name: "april to may",
			args: args{
				s:   "april to may",
				now: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
				dir: Future,
			},
			wantR: RangeFromTimes(
				time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2022, 5, 1, 0, 0, 0, 0, time.UTC),
			),
			wantParsed: "april to may",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotR, gotParsed, err := ParseRange(tt.args.s, tt.args.now, tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotR, tt.wantR) {
				t.Errorf("ParseRange() gotR = %v, want %v", gotR, tt.wantR)
			}
			if gotParsed != tt.wantParsed {
				t.Errorf("ParseRange() gotParsed = %v, want %v", gotParsed, tt.wantParsed)
			}
		})
	}
}

func TestErrNoConnectorFound_Error(t *testing.T) {
	type fields struct {
		ParsedStart    string
		WordAfterStart string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "now foo",
			fields: fields{
				ParsedStart:    "now",
				WordAfterStart: "foo",
			},
			want: "expected 'to|until|til|through' after \"now\", got \"foo\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ErrNoConnectorFound{
				ParsedStart:    tt.fields.ParsedStart,
				WordAfterStart: tt.fields.WordAfterStart,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}
