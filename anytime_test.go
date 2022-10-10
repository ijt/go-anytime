package anytime

import (
	"fmt"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/tj/assert"
)

var now = time.Date(2022, 9, 29, 2, 48, 33, 123, time.Local)
var today = truncateDay(now)

func dateAtTime(dateFrom time.Time, hour int, min int, sec int) time.Time {
	t := dateFrom
	return time.Date(t.Year(), t.Month(), t.Day(), hour, min, sec, 0, t.Location())
}

// Test parsing on cases that are expected to parse successfully.
func TestParse_goodTimes(t *testing.T) {
	var cases = []struct {
		Input    string
		WantTime time.Time
	}{
		// now
		{`now`, now},

		// minutes
		{`a minute from now`, now.Add(time.Minute)},
		{`a minute ago`, now.Add(-time.Minute)},
		{`1 minute ago`, now.Add(-time.Minute)},

		{`5 minutes ago`, now.Add(-5 * time.Minute)},
		{`five minutes ago`, now.Add(-5 * time.Minute)},
		{`   5    minutes  ago   `, now.Add(-5 * time.Minute)},
		{`2 minutes from now`, now.Add(2 * time.Minute)},
		{`two minutes from now`, now.Add(2 * time.Minute)},

		// hours
		{`an hour from now`, now.Add(time.Hour)},
		{`an hour ago`, now.Add(-time.Hour)},
		{`1 hour ago`, now.Add(-time.Hour)},
		{`6 hours ago`, now.Add(-6 * time.Hour)},
		{`1 hour from now`, now.Add(time.Hour)},

		// times
		{"5:35:52pm", dateAtTime(now, 12+5, 35, 52)},

		// dates with times
		{"3 feb 2025 at 5:35:52pm", time.Date(2025, time.February, 3, 12+5, 35, 52, 0, now.Location())},
		{`3 days ago at 11:25am`, dateAtTime(now.Add(-3*24*time.Hour), 11, 25, 0)},
		{`3 days from now at 14:26`, dateAtTime(now.Add(3*24*time.Hour), 14, 26, 0)},
		{`2 weeks ago at 8am`, dateAtTime(now.Add(-2*7*24*time.Hour), 8, 0, 0)},
		{`today at 10am`, dateAtTime(now, 10, 0, 0)},
		{`10am today`, dateAtTime(now, 10, 0, 0)},
		{`yesterday 10am`, dateAtTime(now.AddDate(0, 0, -1), 10, 0, 0)},
		{`10am yesterday`, dateAtTime(now.AddDate(0, 0, -1), 10, 0, 0)},
		{`yesterday at 10am`, dateAtTime(now.AddDate(0, 0, -1), 10, 0, 0)},
		{`yesterday at 10:15am`, dateAtTime(now.AddDate(0, 0, -1), 10, 15, 0)},
		{`tomorrow 10am`, dateAtTime(now.AddDate(0, 0, 1), 10, 0, 0)},
		{`10am tomorrow`, dateAtTime(now.AddDate(0, 0, 1), 10, 0, 0)},
		{`tomorrow at 10am`, dateAtTime(now.AddDate(0, 0, 1), 10, 0, 0)},
		{`tomorrow at 10:15am`, dateAtTime(now.AddDate(0, 0, 1), 10, 15, 0)},
		{`10:15am tomorrow`, dateAtTime(now.AddDate(0, 0, 1), 10, 15, 0)},
		{"next December 25th at 7:30am UTC-7", timeInLocation(nextMonthDayTime(now, time.December, 25, 7, 30, 0), fixedZone(-7))},
		{`next December 23rd AT 5:25 PM`, nextMonthDayTime(now, time.December, 23, 12+5, 25, 0)},
		{`last December 23rd AT 5:25 PM`, prevMonthDayTime(now, time.December, 23, 12+5, 25, 0)},
		{`last sunday at 5:30pm`, dateAtTime(prevWeekdayFrom(now, time.Sunday), 12+5, 30, 0)},
		{`next sunday at 22:45`, dateAtTime(nextWeekdayFrom(now, time.Sunday), 22, 45, 0)},
		{`next sunday at 22:45`, dateAtTime(nextWeekdayFrom(now, time.Sunday), 22, 45, 0)},
		{`November 3rd, 1986 at 4:30pm`, time.Date(1986, 11, 3, 12+4, 30, 0, 0, now.Location())},
		{"September 17, 2012 at 10:09am UTC", time.Date(2012, 9, 17, 10, 9, 0, 0, time.UTC)},
		{"September 17, 2012 at 10:09am UTC-8", time.Date(2012, 9, 17, 10, 9, 0, 0, fixedZone(-8))},
		{"September 17, 2012 at 10:09am UTC+8", time.Date(2012, 9, 17, 10, 9, 0, 0, fixedZone(8))},
		{"September 17, 2012, 10:11:09", time.Date(2012, 9, 17, 10, 11, 9, 0, now.Location())},
		{"September 17, 2012, 10:11", time.Date(2012, 9, 17, 10, 11, 0, 0, now.Location())},
		{"September 17, 2012 10:11", time.Date(2012, 9, 17, 10, 11, 0, 0, now.Location())},
		{"September 17 2012 10:11", time.Date(2012, 9, 17, 10, 11, 0, 0, now.Location())},
		{"September 17 2012 at 10:11", time.Date(2012, 9, 17, 10, 11, 0, 0, now.Location())},

		// formats from the Go time package:
		// ANSIC
		{"Mon Jan  2 15:04:05 2006", time.Date(2006, 1, 2, 15, 4, 5, 0, now.Location())},
		// RubyDate
		{"Mon Jan 02 15:04:05 -0700 2006", time.Date(2006, 1, 2, 15, 4, 5, 0, fixedZone(-7))},
		// RFC1123Z
		{"Mon, 02 Jan 2006 15:04:05 -0700", time.Date(2006, 1, 2, 15, 4, 5, 0, fixedZone(-7))},
		{"Mon 02 Jan 2006 15:04:05 -0700", time.Date(2006, 1, 2, 15, 4, 5, 0, fixedZone(-7))},
		// RFC3339
		{"2006-01-02T15:04:05Z", time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)},
		{"1990-12-31T15:59:59-08:00", time.Date(1990, 12, 31, 15, 59, 59, 0, time.FixedZone("", -8*60*60))},

		// days
		{`one day ago`, now.Add(-24 * time.Hour)},
		{`1 day ago`, now.Add(-24 * time.Hour)},
		{`3 days ago`, now.Add(-3 * 24 * time.Hour)},
		{`three days ago`, now.Add(-3 * 24 * time.Hour)},
		{`1 day from now`, now.Add(24 * time.Hour)},

		// weeks
		{`1 week ago`, now.Add(-7 * 24 * time.Hour)},
		{`2 weeks ago`, now.Add(-2 * 7 * 24 * time.Hour)},
		{`a week from now`, now.Add(7 * 24 * time.Hour)},
		{`a week from today`, now.Add(7 * 24 * time.Hour)},

		// months
		{`a month ago`, now.AddDate(0, -1, 0)},
		{`1 month ago`, now.AddDate(0, -1, 0)},
		{`2 months ago`, now.AddDate(0, -2, 0)},
		{`12 months ago`, now.AddDate(0, -12, 0)},
		{`a month from now`, now.AddDate(0, 1, 0)},
		{`one month hence`, now.AddDate(0, 1, 0)},
		{`1 month from now`, now.AddDate(0, 1, 0)},
		{`2 months from now`, now.AddDate(0, 2, 0)},
		{`last january`, prevMonth(now, time.January)},
		{`next january`, nextMonth(now, time.January)},

		// years
		{`one year ago`, now.AddDate(-1, 0, 0)},
		{`one year from now`, now.AddDate(1, 0, 0)},
		{`one year from today`, now.AddDate(1, 0, 0)},
		{`two years ago`, now.AddDate(-2, 0, 0)},
		{`2 years ago`, now.AddDate(-2, 0, 0)},
		{`this year`, truncateYear(now)},
		{`1999`, time.Date(1999, 1, 1, 0, 0, 0, 0, now.Location())},
		{`2008`, time.Date(2008, 1, 1, 0, 0, 0, 0, now.Location())},
	}

	for _, c := range cases {
		t.Run(c.Input, func(t *testing.T) {
			v, err := Parse(c.Input, now)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, c.WantTime, v)
		})
	}
}

func TestParse_goodDays(t *testing.T) {
	var cases = []struct {
		Input    string
		WantTime time.Time
	}{
		// years
		{`last year`, truncateYear(now.AddDate(-1, 0, 0))},
		{`next year`, truncateYear(now.AddDate(1, 0, 0))},

		// today
		{`today`, now},

		// yesterday
		{`yesterday`, now.AddDate(0, 0, -1)},

		// tomorrow
		{`tomorrow`, now.AddDate(0, 0, 1)},

		// weeks
		{`last week`, truncateWeek(now.AddDate(0, 0, -7))},
		{`next week`, truncateWeek(now.AddDate(0, 0, 7))},

		// past weekdays
		{`last sunday`, prevWeekdayFrom(now, time.Sunday)},
		{`last monday`, prevWeekdayFrom(now, time.Monday)},
		{`last tuesday`, prevWeekdayFrom(now, time.Tuesday)},
		{`last wednesday`, prevWeekdayFrom(now, time.Wednesday)},
		{`last thursday`, prevWeekdayFrom(now, time.Thursday)},
		{`last friday`, prevWeekdayFrom(now, time.Friday)},
		{`last saturday`, prevWeekdayFrom(now, time.Saturday)},

		// future weekdays
		{`next sunday`, nextWeekdayFrom(now, time.Sunday)},
		{`next monday`, nextWeekdayFrom(now, time.Monday)},
		{`next tuesday`, nextWeekdayFrom(now, time.Tuesday)},
		{`next wednesday`, nextWeekdayFrom(now, time.Wednesday)},
		{`next thursday`, nextWeekdayFrom(now, time.Thursday)},
		{`next friday`, nextWeekdayFrom(now, time.Friday)},
		{`next saturday`, nextWeekdayFrom(now, time.Saturday)},

		// months
		{`last january`, prevMonth(now, time.January)},
		{`next january`, nextMonth(now, time.January)},
		{`last month`, truncateMonth(now.AddDate(0, -1, 0))},
		{`next month`, truncateMonth(now.AddDate(0, 1, 0))},

		// absolute dates
		{"january 2017", time.Date(2017, 1, 1, 0, 0, 0, 0, now.Location())},
		{"january, 2017", time.Date(2017, 1, 1, 0, 0, 0, 0, now.Location())},
		{"april 3 2017", time.Date(2017, 4, 3, 0, 0, 0, 0, now.Location())},
		{"april 3, 2017", time.Date(2017, 4, 3, 0, 0, 0, 0, now.Location())},
		{"oct 7, 1970", time.Date(1970, 10, 7, 0, 0, 0, 0, now.Location())},
		{"oct 7 1970", time.Date(1970, 10, 7, 0, 0, 0, 0, now.Location())},
		{"oct. 7, 1970", time.Date(1970, 10, 7, 0, 0, 0, 0, now.Location())},
		{"September 17, 2012 UTC+7", time.Date(2012, 9, 17, 10, 9, 0, 0, fixedZone(7))},
		{"September 17, 2012", time.Date(2012, 9, 17, 10, 9, 0, 0, now.Location())},
		{"7 oct 1970", time.Date(1970, 10, 7, 0, 0, 0, 0, now.Location())},
		{"7 oct, 1970", time.Date(1970, 10, 7, 0, 0, 0, 0, now.Location())},
		{"03 February 2013", time.Date(2013, 2, 3, 0, 0, 0, 0, now.Location())},
		{"2 July 2013", time.Date(2013, 7, 2, 0, 0, 0, 0, now.Location())},
		{"2022 Feb 1", time.Date(2022, 2, 1, 0, 0, 0, 0, now.Location())},
		// yyyy/mm/dd, dd/mm/yyyy etc.
		{"2014/3/31", time.Date(2014, 3, 31, 0, 0, 0, 0, now.Location())},
		{"2014/3/31 UTC", time.Date(2014, 3, 31, 0, 0, 0, 0, location("UTC"))},
		{"2014/3/31 UTC+1", time.Date(2014, 3, 31, 0, 0, 0, 0, fixedZone(1))},
		{"2014/03/31", time.Date(2014, 3, 31, 0, 0, 0, 0, now.Location())},
		{"2014/03/31 UTC-1", time.Date(2014, 3, 31, 0, 0, 0, 0, fixedZone(-1))},
		{"2014-04-26", time.Date(2014, 4, 26, 0, 0, 0, 0, now.Location())},
		{"2014-4-26", time.Date(2014, 4, 26, 0, 0, 0, 0, now.Location())},
		{"2014-4-6", time.Date(2014, 4, 6, 0, 0, 0, 0, now.Location())},
		{"31/3/2014 UTC-8", time.Date(2014, 3, 31, 0, 0, 0, 0, fixedZone(-8))},
		{"31-3-2014 UTC-8", time.Date(2014, 3, 31, 0, 0, 0, 0, fixedZone(-8))},
		{"31/3/2014", time.Date(2014, 3, 31, 0, 0, 0, 0, now.Location())},
		{"31-3-2014", time.Date(2014, 3, 31, 0, 0, 0, 0, now.Location())},

		// color month
		// http://www.jdawiseman.com/papers/trivia/futures.html
		{"white october", nextMonth(now, time.October)},
		{"red october", nextMonth(now, time.October).AddDate(1, 0, 0)},
		{"green october", nextMonth(now, time.October).AddDate(2, 0, 0)},
		{"blue october", nextMonth(now, time.October).AddDate(3, 0, 0)},
		{"gold october", nextMonth(now, time.October).AddDate(4, 0, 0)},
		{"purple october", nextMonth(now, time.October).AddDate(5, 0, 0)},
		{"orange october", nextMonth(now, time.October).AddDate(6, 0, 0)},
		{"pink october", nextMonth(now, time.October).AddDate(7, 0, 0)},
		{"silver october", nextMonth(now, time.October).AddDate(8, 0, 0)},
		{"copper october", nextMonth(now, time.October).AddDate(9, 0, 0)},
	}

	for _, c := range cases {
		t.Run(c.Input, func(t *testing.T) {
			v, err := Parse(c.Input, now)
			if err != nil {
				t.Fatal(err)
			}
			want := truncateDay(c.WantTime)
			assert.Equal(t, want, v)
		})
	}
}

// TestParse_futurePast tests dates and times that are ambiguously in the past
// or the future.
func TestParse_futurePast(t *testing.T) {
	tests := []struct {
		input      string
		wantFuture time.Time
		wantPast   time.Time
	}{
		{
			"december 20",
			nextMonthDayTime(now, time.December, 20, 0, 0, 0),
			prevMonthDayTime(now, time.December, 20, 0, 0, 0),
		},
		{
			"thursday",
			nextWeekdayFrom(now, time.Thursday),
			prevWeekdayFrom(now, time.Thursday),
		},
		{
			"december 20 at 9pm",
			nextMonthDayTime(now, time.December, 20, 21, 0, 0),
			prevMonthDayTime(now, time.December, 20, 21, 0, 0),
		},
		{
			"thursday at 23:59",
			setTime(nextWeekdayFrom(now, time.Thursday), 23, 59, 0, 0),
			setTime(prevWeekdayFrom(now, time.Thursday), 23, 59, 0, 0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// future
			got, err := Parse(tt.input, now, DefaultToFuture)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.wantFuture {
				t.Errorf("Parse(..., Future) = %v, want %v", got, tt.wantFuture)
			}

			// past
			got, err = Parse(tt.input, now, DefaultToPast)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.wantPast {
				t.Errorf("Parse(..., Past) = %v, want %v", got, tt.wantPast)
			}
		})
	}
}

func TestParse_monthOnly(t *testing.T) {
	tests := []struct {
		input string
		month time.Month
	}{
		{"january", time.January},
		{"february", time.February},
		{"march", time.March},
		{"april", time.April},
		{"may", time.May},
		{"june", time.June},
		{"july", time.July},
		{"august", time.August},
		{"september", time.September},
		{"october", time.October},
		{"november", time.November},
		{"december", time.December},
		{"December", time.December},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// future
			got, err := Parse(tt.input, now, DefaultToFuture)
			if err != nil {
				t.Errorf("Parse(..., DefaultToFuture) error = %v", err)
				return
			}
			futureDate := nextMonth(now, tt.month)
			if !reflect.DeepEqual(got, futureDate) {
				t.Errorf("Parse(..., DefaultToFuture) got = %v, want %v", got, futureDate)
			}

			// past
			got, err = Parse(tt.input, now, DefaultToPast)
			if err != nil {
				t.Errorf("Parse(..., DefaultToFuture) error = %v", err)
				return
			}
			pastDate := prevMonth(now, tt.month)
			if !reflect.DeepEqual(got, pastDate) {
				t.Errorf("Parse(..., DefaultToFuture) got = %v, want %v", got, pastDate)
			}
		})
	}
}

func TestParse_future(t *testing.T) {
	tests := []struct {
		input      string
		wantOutput time.Time
	}{
		{"january", nextMonth(now, time.January)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Parse(tt.input, now, DefaultToFuture)
			if err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.wantOutput) {
				t.Errorf("Parse() got = %v, want %v", got, tt.wantOutput)
			}
		})
	}
}

func location(locStr string) *time.Location {
	l, err := time.LoadLocation(locStr)
	if err != nil {
		panic(fmt.Sprintf("loading location %q: %v", locStr, err))
	}
	return l
}

func timeInLocation(t time.Time, l *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), l)
}

// Benchmark parsing.
func BenchmarkParse(b *testing.B) {
	b.SetBytes(1)
	for i := 0; i < b.N; i++ {
		_, err := Parse(`december 23rd 2022 at 5:25pm`, time.Time{})
		if err != nil {
			log.Fatalf("error: %s", err)
		}
	}
}

// Test parsing with inputs that are expected to result in errors.
func TestParse_bad(t *testing.T) {
	var badCases = []struct {
		input string
	}{
		{``},
		{`a`},
		{`not a date or a time`},
		{`right now`},
		{`  right  now  `},
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

		// Ambiguous 12-hour times:
		// These are ambiguous because they don't include the date.
		{`10am`},
		{`10 am`},
		{`5pm`},
		{`10:25am`},
		{`1:05pm`},
		{`10:25:10am`},
		{`1:05:10pm`},

		// Ambiguous 24-hour times:
		// These are ambiguous because they don't include the date.
		{`10`},
		{`10:25`},
		{`10:25:30`},
		{`17`},
		{`17:25:30`},

		// Goofy input:
		{`10:am`},
	}
	for _, c := range badCases {
		t.Run(c.input, func(t *testing.T) {
			now := time.Time{}
			v, err := Parse(c.input, now)
			if err == nil {
				t.Errorf("err is nil, result is %v", v)
			}
		})
	}
}

func Test_truncateWeek(t *testing.T) {
	type args struct {
		t time.Time
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "sunday",
			args: args{
				t: time.Date(2022, 10, 2, 23, 59, 59, 999999, time.UTC),
			},
			want: time.Date(2022, 10, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "monday",
			args: args{
				t: time.Date(2022, 10, 3, 23, 59, 59, 999999, time.UTC),
			},
			want: time.Date(2022, 10, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "saturday",
			args: args{
				t: time.Date(2022, 10, 15, 23, 59, 59, 999999, time.UTC),
			},
			want: time.Date(2022, 10, 9, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncateWeek(tt.args.t); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("truncateWeek() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRange(t *testing.T) {
	tests := []struct {
		input string
		want  Range
	}{
		// from A to B
		{
			"from 3 feb 2022 to 6 oct 2022",
			Range{
				Start: time.Date(2022, 2, 3, 0, 0, 0, 0, now.Location()),
				End:   time.Date(2022, 10, 6, 0, 0, 0, 0, now.Location()),
			},
		},
		// A to B
		{
			"3 feb 2022 to 6 oct 2022",
			Range{
				Start: time.Date(2022, 2, 3, 0, 0, 0, 0, now.Location()),
				End:   time.Date(2022, 10, 6, 0, 0, 0, 0, now.Location()),
			},
		},
		// from A until B
		{
			"from 3 feb 2022 until 6 oct 2022",
			Range{
				Start: time.Date(2022, 2, 3, 0, 0, 0, 0, now.Location()),
				End:   time.Date(2022, 10, 6, 0, 0, 0, 0, now.Location()),
			},
		},
		{
			"from tuesday at 5pm -12:00 until thursday 23:52 +14:00",
			Range{
				Start: setLocation(setTime(nextWeekdayFrom(now, time.Tuesday), 12+5, 0, 0, 0), fixedZone(-12)),
				End:   setLocation(setTime(nextWeekdayFrom(now, time.Thursday), 23, 52, 0, 0), fixedZone(14)),
			},
		},
		// yesterday
		{
			"yesterday",
			Range{
				Start: today.AddDate(0, 0, -1),
				End:   today,
			},
		},
		// today
		{
			"today",
			Range{
				Start: today,
				End:   today.AddDate(0, 0, 1),
			},
		},
		// tomorrow
		{
			"tomorrow",
			Range{
				Start: today.AddDate(0, 0, 1),
				End:   today.AddDate(0, 0, 2),
			},
		},
		// last week
		{
			"last week",
			Range{
				Start: time.Date(2022, 9, 18, 0, 0, 0, 0, now.Location()),
				End:   time.Date(2022, 9, 25, 0, 0, 0, 0, now.Location()),
			},
		},
		// this week
		{
			"this week",
			Range{
				Start: time.Date(2022, 9, 25, 0, 0, 0, 0, now.Location()),
				End:   time.Date(2022, 10, 2, 0, 0, 0, 0, now.Location()),
			},
		},
		// next week
		{
			"next week",
			Range{
				Start: time.Date(2022, 10, 3, 0, 0, 0, 0, now.Location()),
				End:   time.Date(2022, 10, 9, 0, 0, 0, 0, now.Location()),
			},
		},
		// last month
		{
			"last month",
			Range{
				Start: time.Date(2022, 8, 1, 0, 0, 0, 0, now.Location()),
				End:   time.Date(2022, 9, 1, 0, 0, 0, 0, now.Location()),
			},
		},
		// this month
		{
			"this month",
			Range{
				Start: time.Date(2022, 9, 1, 0, 0, 0, 0, now.Location()),
				End:   time.Date(2022, 10, 1, 0, 0, 0, 0, now.Location()),
			},
		},
		// next month
		{
			"next month",
			Range{
				Start: time.Date(2022, 10, 1, 0, 0, 0, 0, now.Location()),
				End:   time.Date(2022, 11, 1, 0, 0, 0, 0, now.Location()),
			},
		},
		// last year
		{
			"last year",
			Range{
				Start: time.Date(2021, 1, 1, 0, 0, 0, 0, now.Location()),
				End:   time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
			},
		},
		// this year
		{
			"this year",
			Range{
				Start: time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
				End:   time.Date(2023, 1, 1, 0, 0, 0, 0, now.Location()),
			},
		},
		// next year
		{
			"next year",
			Range{
				Start: time.Date(2023, 1, 1, 0, 0, 0, 0, now.Location()),
				End:   time.Date(2024, 1, 1, 0, 0, 0, 0, now.Location()),
			},
		},
		// absolute year
		{
			"2025",
			Range{
				Start: time.Date(2025, 1, 1, 0, 0, 0, 0, now.Location()),
				End:   time.Date(2026, 1, 1, 0, 0, 0, 0, now.Location()),
			},
		},
		// absolute month
		{
			"feb 2025",
			Range{
				Start: time.Date(2025, 2, 1, 0, 0, 0, 0, now.Location()),
				End:   time.Date(2025, 3, 1, 0, 0, 0, 0, now.Location()),
			},
		},
		// absolute day
		{
			"3 feb 2025",
			Range{
				Start: time.Date(2025, 2, 3, 0, 0, 0, 0, now.Location()),
				End:   time.Date(2025, 2, 4, 0, 0, 0, 0, now.Location()),
			},
		},
		// absolute hour
		{
			"3 feb 2025 at 5pm",
			Range{
				Start: time.Date(2025, 2, 3, 12+5, 0, 0, 0, now.Location()),
				End:   time.Date(2025, 2, 3, 12+5+1, 0, 0, 0, now.Location()),
			},
		},
		// absolute minute
		{
			"3 feb 2025 at 5:35pm",
			Range{
				Start: time.Date(2025, 2, 3, 12+5, 35, 0, 0, now.Location()),
				End:   time.Date(2025, 2, 3, 12+5, 36, 0, 0, now.Location()),
			},
		},
		// absolute second
		{
			"3 feb 2025 at 5:35:52pm",
			Range{
				Start: time.Date(2025, 2, 3, 12+5, 35, 52, 0, now.Location()),
				End:   time.Date(2025, 2, 3, 12+5, 35, 53, 0, now.Location()),
			},
		},
		// 2022 jan 1 0:0:0
		{
			"2022 jan 1 0:0:0",
			Range{
				Start: time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
				End:   time.Date(2022, 1, 1, 0, 0, 1, 0, now.Location()),
			},
		},
		// 2022 jan 1 0:0
		{
			"2022 jan 1 0:0",
			Range{
				Start: time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
				End:   time.Date(2022, 1, 1, 1, 0, 0, 0, now.Location()),
			},
		},
		// 2022 jan 1 0
		{
			"2022 jan 1 0",
			Range{
				Start: time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
				End:   time.Date(2022, 1, 1, 1, 0, 0, 0, now.Location()),
			},
		},
		// 2022 jan 1
		{
			"2022 jan 1",
			Range{
				Start: time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
				End:   time.Date(2022, 1, 2, 0, 0, 0, 0, now.Location()),
			},
		},
		// 2022 jan
		{
			"2022 jan",
			Range{
				Start: time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
				End:   time.Date(2022, 2, 1, 0, 0, 0, 0, now.Location()),
			},
		},
		// 2022
		{
			"2022",
			Range{
				Start: time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
				End:   time.Date(2023, 1, 1, 0, 0, 0, 0, now.Location()),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseRange(tt.input, now, DefaultToFuture)
			if err != nil {
				t.Errorf("ParseRange() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseRange() got =\n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}
