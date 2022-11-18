package anytime

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"
	"time"

	gp "github.com/ijt/goparsify"
	"github.com/tj/assert"
)

var now = time.Date(2022, 9, 29, 2, 48, 33, 123, time.UTC)
var today = truncateDay(now).Time

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
		{"noon", dateAtTime(today, 12, 0, 0)},
		{"5:35:52pm", dateAtTime(today, 12+5, 35, 52)},
		{`10am`, dateAtTime(today, 10, 0, 0)},
		{`10 am`, dateAtTime(today, 10, 0, 0)},
		{`5pm`, dateAtTime(today, 12+5, 0, 0)},
		{`10:25am`, dateAtTime(today, 10, 25, 0)},
		{`1:05pm`, dateAtTime(today, 12+1, 5, 0)},
		{`10:25:10am`, dateAtTime(today, 10, 25, 10)},
		{`1:05:10pm`, dateAtTime(today, 12+1, 5, 10)},
		{`10:25`, dateAtTime(today, 10, 25, 0)},
		{`10:25:30`, dateAtTime(today, 10, 25, 30)},
		{`17:25:30`, dateAtTime(today, 17, 25, 30)},

		// dates with times
		{"On Friday at noon UTC", timeInLocation(dateAtTime(nextWeekdayFrom(now, time.Friday), 12, 0, 0), time.UTC)},
		{"On Tuesday at 11am UTC", timeInLocation(dateAtTime(nextWeekdayFrom(now, time.Tuesday), 11, 0, 0), time.UTC)},
		{"On 3 feb 2025 at 5:35:52pm", time.Date(2025, time.February, 3, 12+5, 35, 52, 0, now.Location())},
		{"3 feb 2025 at 5:35:52pm", time.Date(2025, time.February, 3, 12+5, 35, 52, 0, now.Location())},
		{`3 days ago at 11:25am`, dateAtTime(now.Add(-3*24*time.Hour), 11, 25, 0)},
		{`3 days from now at 14:26`, dateAtTime(now.Add(3*24*time.Hour), 14, 26, 0)},
		{`2 weeks ago at 8am`, dateAtTime(now.Add(-2*7*24*time.Hour), 8, 0, 0)},
		{`Today at 10am`, dateAtTime(now, 10, 0, 0)},
		{`10am today`, dateAtTime(now, 10, 0, 0)},
		{`Yesterday 10am`, dateAtTime(now.AddDate(0, 0, -1), 10, 0, 0)},
		{`10am yesterday`, dateAtTime(now.AddDate(0, 0, -1), 10, 0, 0)},
		{`Yesterday at 10am`, dateAtTime(now.AddDate(0, 0, -1), 10, 0, 0)},
		{`Yesterday at 10:15am`, dateAtTime(now.AddDate(0, 0, -1), 10, 15, 0)},
		{`Tomorrow 10am`, dateAtTime(now.AddDate(0, 0, 1), 10, 0, 0)},
		{`10am tomorrow`, dateAtTime(now.AddDate(0, 0, 1), 10, 0, 0)},
		{`Tomorrow at 10am`, dateAtTime(now.AddDate(0, 0, 1), 10, 0, 0)},
		{`Tomorrow at 10:15am`, dateAtTime(now.AddDate(0, 0, 1), 10, 15, 0)},
		{`10:15am tomorrow`, dateAtTime(now.AddDate(0, 0, 1), 10, 15, 0)},
		{"Next dec 22nd at 3pm", timeInLocation(nextMonthDayTime(now, time.December, 22, 12+3, 0, 0), now.Location())},
		{"Next December 25th at 7:30am UTC-7", timeInLocation(nextMonthDayTime(now, time.December, 25, 7, 30, 0), fixedZone(-7))},
		{`Next December 23rd AT 5:25 PM`, nextMonthDayTime(now, time.December, 23, 12+5, 25, 0)},
		{`Last December 23rd AT 5:25 PM`, prevMonthDayTime(now, time.December, 23, 12+5, 25, 0)},
		{`Last sunday at 5:30pm`, dateAtTime(prevWeekdayFrom(now, time.Sunday), 12+5, 30, 0)},
		{`Next sunday at 22:45`, dateAtTime(nextWeekdayFrom(now, time.Sunday), 22, 45, 0)},
		{`Next sunday at 22:45`, dateAtTime(nextWeekdayFrom(now, time.Sunday), 22, 45, 0)},
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
		{`One day ago`, now.Add(-24 * time.Hour)},
		{`1 day ago`, now.Add(-24 * time.Hour)},
		{`3 days ago`, now.Add(-3 * 24 * time.Hour)},
		{`Three days ago`, now.Add(-3 * 24 * time.Hour)},
		{`1 day from now`, now.Add(24 * time.Hour)},

		// weeks
		{`1 week ago`, now.Add(-7 * 24 * time.Hour)},
		{`2 weeks ago`, now.Add(-2 * 7 * 24 * time.Hour)},
		{`A week from now`, now.Add(7 * 24 * time.Hour)},
		{`A week from today`, now.Add(7 * 24 * time.Hour)},

		// months
		{`A month ago`, now.AddDate(0, -1, 0)},
		{`1 month ago`, now.AddDate(0, -1, 0)},
		{`2 months ago`, now.AddDate(0, -2, 0)},
		{`12 months ago`, now.AddDate(0, -12, 0)},
		{`A month from now`, now.AddDate(0, 1, 0)},
		{`One month hence`, now.AddDate(0, 1, 0)},
		{`1 month from now`, now.AddDate(0, 1, 0)},
		{`2 months from now`, now.AddDate(0, 2, 0)},
		{`Last January`, prevMonth(now, time.January).Time},
		{`Last january`, prevMonth(now, time.January).Time},
		{`Next january`, nextMonth(now, time.January).Time},

		// years
		{`One year ago`, now.AddDate(-1, 0, 0)},
		{`One year from now`, now.AddDate(1, 0, 0)},
		{`One year from today`, now.AddDate(1, 0, 0)},
		{`Two years ago`, now.AddDate(-2, 0, 0)},
		{`2 years ago`, now.AddDate(-2, 0, 0)},
		{`This year`, truncateYear(now).Time},
		{`1999AD`, time.Date(1999, 1, 1, 0, 0, 0, 0, now.Location())},
		{`1999 AD`, time.Date(1999, 1, 1, 0, 0, 0, 0, now.Location())},
		{`2008CE`, time.Date(2008, 1, 1, 0, 0, 0, 0, now.Location())},
		{`2008 CE`, time.Date(2008, 1, 1, 0, 0, 0, 0, now.Location())},
	}

	for _, c := range cases {
		t.Run(c.Input, func(t *testing.T) {
			v, err := Parse(c.Input, now)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, c.WantTime, v)

			// Run the parser at a lower level and make sure the token it
			// returns matches the input string.
			dp := Parser(now, DefaultToFuture)
			node, _ := runParser(c.Input, dp)
			want := strings.TrimSpace(c.Input)
			if node.Token != want {
				t.Errorf("parsed token = %q, want %q", node.Token, want)
			}
		})
	}
}

func TestParse_goodDays(t *testing.T) {
	var cases = []struct {
		Input    string
		WantTime time.Time
	}{
		// years
		{`Last year`, truncateYear(now.AddDate(-1, 0, 0)).Time},
		{`Next year`, truncateYear(now.AddDate(1, 0, 0)).Time},

		// today
		{`Today`, now},

		// yesterday
		{`Yesterday`, now.AddDate(0, 0, -1)},

		// tomorrow
		{`Tomorrow`, now.AddDate(0, 0, 1)},

		// weeks
		{`Last week`, truncateWeek(now.AddDate(0, 0, -7)).Time},
		{`Next week`, truncateWeek(now.AddDate(0, 0, 7)).Time},

		// past weekdays
		{`Last sunday`, prevWeekdayFrom(now, time.Sunday)},
		{`Last monday`, prevWeekdayFrom(now, time.Monday)},
		{`Last Monday`, prevWeekdayFrom(now, time.Monday)},
		{`Last tuesday`, prevWeekdayFrom(now, time.Tuesday)},
		{`Last wednesday`, prevWeekdayFrom(now, time.Wednesday)},
		{`Last Thursday`, prevWeekdayFrom(now, time.Thursday)},
		{`Last Friday`, prevWeekdayFrom(now, time.Friday)},
		{`Last saturday`, prevWeekdayFrom(now, time.Saturday)},

		// future weekdays
		{`Next sunday`, nextWeekdayFrom(now, time.Sunday)},
		{`Next monday`, nextWeekdayFrom(now, time.Monday)},
		{`Next tuesday`, nextWeekdayFrom(now, time.Tuesday)},
		{`Next wednesday`, nextWeekdayFrom(now, time.Wednesday)},
		{`Next thursday`, nextWeekdayFrom(now, time.Thursday)},
		{`Next friday`, nextWeekdayFrom(now, time.Friday)},
		{`Next saturday`, nextWeekdayFrom(now, time.Saturday)},

		// months
		{`Last january`, prevMonth(now, time.January).Time},
		{`Next january`, nextMonth(now, time.January).Time},
		{`Last month`, truncateMonth(now.AddDate(0, -1, 0)).Time},
		{`Next month`, truncateMonth(now.AddDate(0, 1, 0)).Time},

		// absolute dates
		{"January 2017", time.Date(2017, 1, 1, 0, 0, 0, 0, now.Location())},
		{"January, 2017", time.Date(2017, 1, 1, 0, 0, 0, 0, now.Location())},
		{"April 3 2017", time.Date(2017, 4, 3, 0, 0, 0, 0, now.Location())},
		{"April 3, 2017", time.Date(2017, 4, 3, 0, 0, 0, 0, now.Location())},
		{"Oct 7, 1970", time.Date(1970, 10, 7, 0, 0, 0, 0, now.Location())},
		{"Oct 7 1970", time.Date(1970, 10, 7, 0, 0, 0, 0, now.Location())},
		{"Oct. 7, 1970", time.Date(1970, 10, 7, 0, 0, 0, 0, now.Location())},
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
		{"White october", nextMonth(now, time.October).Time},
		{"Red october", nextMonth(now, time.October).AddDate(1, 0, 0)},
		{"Green october", nextMonth(now, time.October).AddDate(2, 0, 0)},
		{"Blue october", nextMonth(now, time.October).AddDate(3, 0, 0)},
		{"Gold october", nextMonth(now, time.October).AddDate(4, 0, 0)},
		{"Purple october", nextMonth(now, time.October).AddDate(5, 0, 0)},
		{"Orange october", nextMonth(now, time.October).AddDate(6, 0, 0)},
		{"Pink october", nextMonth(now, time.October).AddDate(7, 0, 0)},
		{"Silver october", nextMonth(now, time.October).AddDate(8, 0, 0)},
		{"Copper october", nextMonth(now, time.October).AddDate(9, 0, 0)},
	}

	for _, c := range cases {
		t.Run(c.Input, func(t *testing.T) {
			v, err := Parse(c.Input, now)
			if err != nil {
				t.Fatal(err)
			}
			want := truncateDay(c.WantTime).Time
			assert.Equal(t, want, v)

			// Run the parser at a lower level and make sure the token it
			// returns matches the input string.
			dp := Parser(now, DefaultToFuture)
			node, _ := runParser(c.Input, dp)
			wantTok := strings.TrimSpace(c.Input)
			if node.Token != wantTok {
				t.Errorf("parsed token = %q, want %q", node.Token, wantTok)
			}
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
			"December 20",
			nextMonthDayTime(now, time.December, 20, 0, 0, 0),
			prevMonthDayTime(now, time.December, 20, 0, 0, 0),
		},
		{
			"Thursday",
			nextWeekdayFrom(now, time.Thursday),
			prevWeekdayFrom(now, time.Thursday),
		},
		{
			"On thursday",
			nextWeekdayFrom(now, time.Thursday),
			prevWeekdayFrom(now, time.Thursday),
		},
		{
			"December 20 at 9pm",
			nextMonthDayTime(now, time.December, 20, 21, 0, 0),
			prevMonthDayTime(now, time.December, 20, 21, 0, 0),
		},
		{
			"Thursday at 23:59",
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
		{"January", time.January},
		{"February", time.February},
		{"March", time.March},
		{"April", time.April},
		{"May", time.May},
		{"June", time.June},
		{"July", time.July},
		{"August", time.August},
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
			futureDate := nextMonth(now, tt.month).Time
			if !reflect.DeepEqual(got, futureDate) {
				t.Errorf("Parse(..., DefaultToFuture) got = %v, want %v", got, futureDate)
			}

			// past
			got, err = Parse(tt.input, now, DefaultToPast)
			if err != nil {
				t.Errorf("Parse(..., DefaultToFuture) error = %v", err)
				return
			}
			pastDate := prevMonth(now, tt.month).Time
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
		{"january", nextMonth(now, time.January).Time},
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
			name: "Sunday",
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
			if got := truncateWeek(tt.args.t).Time; !reflect.DeepEqual(got, tt.want) {
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
			"From 3 feb 2022 to 6 oct 2022",
			RangeFromTimes(
				time.Date(2022, 2, 3, 0, 0, 0, 0, now.Location()),
				time.Date(2022, 10, 6, 0, 0, 0, 0, now.Location()),
			),
		},
		// A to B
		{
			"3 feb 2022 to 6 oct 2022",
			RangeFromTimes(
				time.Date(2022, 2, 3, 0, 0, 0, 0, now.Location()),
				time.Date(2022, 10, 6, 0, 0, 0, 0, now.Location()),
			),
		},
		// A through B
		{
			"3 feb 2022 through 6 oct 2022",
			RangeFromTimes(
				time.Date(2022, 2, 3, 0, 0, 0, 0, now.Location()),
				time.Date(2022, 10, 6, 0, 0, 0, 0, now.Location()),
			),
		},
		// from A until B
		{
			"from 3 feb 2022 until 6 oct 2022",
			RangeFromTimes(
				time.Date(2022, 2, 3, 0, 0, 0, 0, now.Location()),
				time.Date(2022, 10, 6, 0, 0, 0, 0, now.Location()),
			),
		},
		{
			"from tuesday at 5pm -12:00 until thursday 23:52 +14:00",
			RangeFromTimes(
				setLocation(setTime(nextWeekdayFrom(now, time.Tuesday), 12+5, 0, 0, 0), fixedZone(-12)),
				setLocation(setTime(nextWeekdayFrom(now, time.Thursday), 23, 52, 0, 0), fixedZone(14)),
			),
		},
		// yesterday
		{
			"Yesterday",
			RangeFromTimes(
				today.AddDate(0, 0, -1),
				today.Add(-time.Second),
			),
		},
		// today
		{
			"Today",
			RangeFromTimes(
				today,
				today.AddDate(0, 0, 1).Add(-time.Second),
			),
		},
		// tomorrow
		{
			"Tomorrow",
			RangeFromTimes(
				today.AddDate(0, 0, 1),
				today.AddDate(0, 0, 2).Add(-time.Second),
			),
		},
		{
			"From today until next thursday",
			RangeFromTimes(
				today,
				nextWeekdayFrom(today, time.Thursday),
			),
		},
		{
			"From tomorrow until next tuesday",
			RangeFromTimes(
				today.AddDate(0, 0, 1),
				nextWeekdayFrom(today, time.Tuesday),
			),
		},
		// last week
		{
			"Last week",
			RangeFromTimes(
				time.Date(2022, 9, 18, 0, 0, 0, 0, now.Location()),
				time.Date(2022, 9, 25, 0, 0, 0, 0, now.Location()).Add(-time.Second),
			),
		},
		// this week
		{
			"This week",
			RangeFromTimes(
				time.Date(2022, 9, 25, 0, 0, 0, 0, now.Location()),
				time.Date(2022, 10, 2, 0, 0, 0, 0, now.Location()).Add(-time.Second),
			),
		},
		// next week
		{
			"next week",
			RangeFromTimes(
				time.Date(2022, 10, 2, 0, 0, 0, 0, now.Location()),
				time.Date(2022, 10, 9, 0, 0, 0, 0, now.Location()).Add(-time.Second),
			),
		},
		// last month
		{
			"Last month",
			RangeFromTimes(
				time.Date(2022, 8, 1, 0, 0, 0, 0, now.Location()),
				time.Date(2022, 9, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second),
			),
		},
		// this month
		{
			"This month",
			RangeFromTimes(
				time.Date(2022, 9, 1, 0, 0, 0, 0, now.Location()),
				time.Date(2022, 10, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second),
			),
		},
		// next month
		{
			"Next month",
			RangeFromTimes(
				time.Date(2022, 10, 1, 0, 0, 0, 0, now.Location()),
				time.Date(2022, 11, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second),
			),
		},
		// last year
		{
			"Last year",
			RangeFromTimes(
				time.Date(2021, 1, 1, 0, 0, 0, 0, now.Location()),
				time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second),
			),
		},
		// this year
		{
			"This year",
			RangeFromTimes(
				time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
				time.Date(2023, 1, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second),
			),
		},
		// next year
		{
			"Next year",
			RangeFromTimes(
				time.Date(2023, 1, 1, 0, 0, 0, 0, now.Location()),
				time.Date(2024, 1, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second),
			),
		},
		// absolute year
		{
			"2025Ad",
			RangeFromTimes(
				time.Date(2025, 1, 1, 0, 0, 0, 0, now.Location()),
				time.Date(2026, 1, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second),
			),
		},
		// absolute month
		{
			"Feb 2025",
			RangeFromTimes(
				time.Date(2025, 2, 1, 0, 0, 0, 0, now.Location()),
				time.Date(2025, 3, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second),
			),
		},
		// absolute day
		{
			"3 feb 2025",
			RangeFromTimes(
				time.Date(2025, 2, 3, 0, 0, 0, 0, now.Location()),
				time.Date(2025, 2, 4, 0, 0, 0, 0, now.Location()).Add(-time.Second),
			),
		},
		// absolute hour
		{
			"3 feb 2025 at 5PM",
			RangeFromTimes(
				time.Date(2025, 2, 3, 12+5, 0, 0, 0, now.Location()),
				time.Date(2025, 2, 3, 12+5+1, 0, 0, 0, now.Location()).Add(-time.Second),
			),
		},
		// absolute minute
		{
			"3 feb 2025 at 5:35pm",
			RangeFromTimes(
				time.Date(2025, 2, 3, 12+5, 35, 0, 0, now.Location()),
				time.Date(2025, 2, 3, 12+5, 36, 0, 0, now.Location()).Add(-time.Second),
			),
		},
		// absolute second
		{
			"3 Feb 2025 at 5:35:52pm",
			RangeFromTimes(
				time.Date(2025, 2, 3, 12+5, 35, 52, 0, now.Location()),
				time.Date(2025, 2, 3, 12+5, 35, 53, 0, now.Location()),
			),
		},
		// 2022 jan 1 0:0:0
		{
			"2022 jan 1 0:0:0",
			RangeFromTimes(
				time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
				time.Date(2022, 1, 1, 0, 0, 1, 0, now.Location()),
			),
		},
		// 2022 jan 1 0:0
		{
			"2022 jan 1 0:0",
			RangeFromTimes(
				time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
				time.Date(2022, 1, 1, 0, 1, 0, 0, now.Location()).Add(-time.Second),
			),
		},
		// 2022 jan 1 12am
		{
			"2022 jan 1 12am",
			RangeFromTimes(
				time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
				time.Date(2022, 1, 1, 1, 0, 0, 0, now.Location()).Add(-time.Second),
			),
		},
		// 2022 jan 1 0am
		{
			"2022 jan 1 0am",
			RangeFromTimes(
				time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
				time.Date(2022, 1, 1, 1, 0, 0, 0, now.Location()).Add(-time.Second),
			),
		},
		// 2022 jan 1
		{
			"2022 jan 1",
			RangeFromTimes(
				time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
				time.Date(2022, 1, 2, 0, 0, 0, 0, now.Location()).Add(-time.Second),
			),
		},
		// 2022 jan
		{
			"2022 jan",
			RangeFromTimes(
				time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
				time.Date(2022, 2, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second),
			),
		},
		// 2022
		{
			"2022ce",
			RangeFromTimes(
				time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
				time.Date(2023, 1, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second),
			),
		},
		// 2022
		{
			"2022CE",
			RangeFromTimes(
				time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
				time.Date(2023, 1, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second),
			),
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

			// Run the parser at a lower level and make sure the token it
			// returns matches the input string.
			rp := RangeParser(now, DefaultToFuture)
			node, _ := runParser(tt.input, rp)
			want := strings.TrimSpace(tt.input)
			if node.Token != want {
				t.Errorf("parsed token = %q, want %q", node.Token, want)
			}
		})
	}
}

func Test_nextWeek(t *testing.T) {
	type args struct {
		ref time.Time
	}
	tests := []struct {
		name string
		args args
		want Range
	}{
		{
			name: "2022-9-30",
			args: args{
				ref: now,
			},
			want: Range{
				Time:     time.Date(2022, 10, 2, 0, 0, 0, 0, now.Location()),
				Duration: 7*24*time.Hour - time.Second,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nextWeek(tt.args.ref); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nextWeek() = \n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}

func TestRange_String(t *testing.T) {
	type fields struct {
		Time     time.Time
		Duration time.Duration
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "zeros",
			fields: fields{
				Time:     time.Time{},
				Duration: 0,
			},
			want: "{time: 0001-01-01 00:00:00 +0000 UTC, duration: 0s}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Range{
				Time:     tt.fields.Time,
				Duration: tt.fields.Duration,
			}
			if got := r.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_prevWeekdayFrom(t *testing.T) {
	tests := []struct {
		day  time.Weekday
		want time.Time
	}{
		{time.Thursday, time.Date(2022, 9, 22, 0, 0, 0, 0, now.Location())},
		{time.Friday, time.Date(2022, 9, 23, 0, 0, 0, 0, now.Location())},
		{time.Saturday, time.Date(2022, 9, 24, 0, 0, 0, 0, now.Location())},
		{time.Sunday, time.Date(2022, 9, 25, 0, 0, 0, 0, now.Location())},
		{time.Monday, time.Date(2022, 9, 26, 0, 0, 0, 0, now.Location())},
		{time.Tuesday, time.Date(2022, 9, 27, 0, 0, 0, 0, now.Location())},
		{time.Wednesday, time.Date(2022, 9, 28, 0, 0, 0, 0, now.Location())},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s", tt.day), func(t *testing.T) {
			if got := prevWeekdayFrom(now, tt.day); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("prevWeekdayFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nextWeekdayFrom(t *testing.T) {
	tests := []struct {
		day  time.Weekday
		want time.Time
	}{
		{time.Friday, time.Date(2022, 9, 30, 0, 0, 0, 0, now.Location())},
		{time.Saturday, time.Date(2022, 10, 1, 0, 0, 0, 0, now.Location())},
		{time.Sunday, time.Date(2022, 10, 2, 0, 0, 0, 0, now.Location())},
		{time.Monday, time.Date(2022, 10, 3, 0, 0, 0, 0, now.Location())},
		{time.Tuesday, time.Date(2022, 10, 4, 0, 0, 0, 0, now.Location())},
		{time.Wednesday, time.Date(2022, 10, 5, 0, 0, 0, 0, now.Location())},
		{time.Thursday, time.Date(2022, 10, 6, 0, 0, 0, 0, now.Location())},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s", tt.day), func(t *testing.T) {
			if got := nextWeekdayFrom(now, tt.day); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("prevWeekdayFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}

func runParser(input string, parser gp.Parser) (gp.Result, *gp.State) {
	ps := gp.NewState(input)
	result := gp.Result{}
	parser(ps, &result)
	return result, ps
}

func TestReplaceTimesByFunc(t *testing.T) {
	type args struct {
		s       string
		ref     time.Time
		f       func(time.Time) string
		options []func(o *opts)
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "empty",
			args: args{
				s:       "",
				ref:     time.Time{},
				f:       nil,
				options: nil,
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "issue 14 example without prefix",
			args: args{
				s:   "on Tuesday at 11am UTC",
				ref: time.Date(2022, time.Month(10), 24, 0, 0, 0, 0, time.UTC),
				f: func(t time.Time) string {
					return t.String()
				},
				options: nil,
			},
			want:    "2022-10-25 11:00:00 +0000 UTC",
			wantErr: false,
		},
		{
			name: "issue 14 example",
			args: args{
				s:   "Let's meet on Tuesday at 11am UTC if that works for you",
				ref: time.Date(2022, time.Month(10), 24, 0, 0, 0, 0, time.UTC),
				f: func(t time.Time) string {
					return t.String()
				},
				options: nil,
			},
			want:    "Let's meet 2022-10-25 11:00:00 +0000 UTC if that works for you",
			wantErr: false,
		},
		{
			name: "two dates",
			args: args{
				s:   "Let's meet on Tuesday at 11am UTC or monday if you like",
				ref: time.Date(2022, time.Month(10), 24, 0, 0, 0, 0, time.UTC),
				f: func(t time.Time) string {
					return t.String()
				},
				options: nil,
			},
			want:    "Let's meet 2022-10-25 11:00:00 +0000 UTC or 2022-10-31 00:00:00 +0000 UTC if you like",
			wantErr: false,
		},
		{
			name: "one range shows up as two dates",
			args: args{
				s:   "from 3 feb 2022 to 6 oct 2022",
				ref: now,
				f: func(t time.Time) string {
					return "DATE"
				},
				options: nil,
			},
			want:    "from DATE to DATE",
			wantErr: false,
		},
		{
			name: "market doesn't get mistaken for month of mar followed by ket",
			args: args{
				s:       "market",
				ref:     time.Time{},
				f:       func(t time.Time) string { return t.String() },
				options: nil,
			},
			want:    "market",
			wantErr: false,
		},
		{
			name: "month doesn't get mistaken for mon-day followed by th",
			args: args{
				s:   "month",
				ref: time.Time{},
				f: func(t time.Time) string {
					return t.String()
				},
				options: nil,
			},
			want:    "month",
			wantErr: false,
		},
		{
			name: "1000 widgets",
			args: args{
				s:   "1000 widgets",
				ref: time.Time{},
				f: func(t time.Time) string {
					return t.String()
				},
				options: nil,
			},
			want:    "1000 widgets",
			wantErr: false,
		},
		{
			name: "1000AD",
			args: args{
				s:   "1000AD",
				ref: time.Time{},
				f: func(t time.Time) string {
					return t.String()
				},
				options: nil,
			},
			want:    time.Date(1000, time.Month(1), 1, 0, 0, 0, 0, time.UTC).String(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReplaceTimesByFunc(tt.args.s, tt.args.ref, tt.args.f, tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReplaceTimesByFunc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReplaceTimesByFunc() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReplaceRangesByFunc(t *testing.T) {
	type args struct {
		s       string
		ref     time.Time
		f       func(Range) string
		options []func(o *opts)
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "empty",
			args: args{
				s:       "",
				ref:     now,
				f:       nil,
				options: nil,
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "two dates without a connecting preposition get left alone",
			args: args{
				s:       "Let's meet on Tuesday at 11am UTC or monday if you like",
				ref:     time.Date(2022, time.Month(10), 24, 0, 0, 0, 0, time.UTC),
				f:       nil,
				options: nil,
			},
			want:    "Let's meet on Tuesday at 11am UTC or monday if you like",
			wantErr: false,
		},
		{
			name: "one range shows up as a range",
			args: args{
				s:   "from 3 feb 2022 to 6 oct 2022",
				ref: now,
				f: func(r Range) string {
					ymd := "2006/01/02"
					return fmt.Sprintf("%s - %s", r.Start().Format(ymd), r.End().Format(ymd))
				},
				options: nil,
			},
			want:    "2022/02/03 - 2022/10/06",
			wantErr: false,
		},
		{
			name: "stuff then range then stuff then range then stuff",
			args: args{
				s:   "twas brillig from 3 feb 2022 to 6 oct 2022 and the slithy toves from april until may did gyre and gimble in the wabe",
				ref: now,
				f: func(r Range) string {
					return "RANGE"
				},
				options: nil,
			},
			want:    "twas brillig RANGE and the slithy toves RANGE did gyre and gimble in the wabe",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReplaceRangesByFunc(tt.args.s, tt.args.ref, tt.args.f, tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReplaceRangesByFunc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReplaceRangesByFunc() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPartitionTimes(t *testing.T) {
	d1 := time.Date(2022, 11, 2, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2023, 12, 30, 0, 0, 0, 0, time.UTC)
	type args struct {
		s       string
		ref     time.Time
		options []func(o *opts)
	}
	tests := []struct {
		name string
		args args
		want []any
	}{
		{
			name: "empty",
			args: args{
				s:       "",
				ref:     time.Time{},
				options: nil,
			},
			want: nil,
		},
		{
			name: "a",
			args: args{
				s:       "a",
				ref:     time.Time{},
				options: nil,
			},
			want: []any{"a"},
		},
		{
			name: "a b",
			args: args{
				s:       "a b",
				ref:     time.Time{},
				options: nil,
			},
			want: []any{"a b"},
		},
		{
			name: "2022/11/2",
			args: args{
				s:       "2022/11/2",
				ref:     time.Time{},
				options: nil,
			},
			want: []any{d1},
		},
		{
			name: "a 2022/11/2 b",
			args: args{
				s:       "a 2022/11/2 b",
				ref:     time.Time{},
				options: nil,
			},
			want: []any{"a", d1, "b"},
		},
		{
			name: "2022/11/2 vs 2023/12/30",
			args: args{
				s:       "2022/11/2 vs 2023/12/30",
				ref:     time.Time{},
				options: nil,
			},
			want: []any{d1, "vs", d2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PartitionTimes(tt.args.s, tt.args.ref, tt.args.options...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PartitionTimes() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestPartitionTimesByFuncs(t *testing.T) {
	type args struct {
		s       string
		ref     time.Time
		ntf     func(nonTimeChunk string)
		tf      func(timeChunk string, t time.Time)
		options []func(o *opts)
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "check for empty string underlying time",
			args: args{
				s:   "2022/11/2",
				ref: time.Time{},
				ntf: nil,
				tf: func(chunk string, t time.Time) {
					if chunk == "" {
						panic("empty chunk for time")
					}
				},
				options: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			PartitionTimesByFuncs(tt.args.s, tt.args.ref, tt.args.ntf, tt.args.tf, tt.args.options...)
		})
	}
}

func TestReplaceDateRangesByFunc(t *testing.T) {
	type args struct {
		s       string
		ref     time.Time
		f       func(source string, r Range) string
		options []func(o *opts)
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "empty",
			args: args{
				s:       "",
				ref:     time.Time{},
				f:       nil,
				options: nil,
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "this week",
			args: args{
				s:   "this week",
				ref: now,
				f: func(source string, r Range) string {
					return fmt.Sprintf("%s to %s", r.Start(), r.End())
				},
				options: nil,
			},
			want:    "2022-09-25 00:00:00 +0000 UTC to 2022-10-01 23:59:59 +0000 UTC",
			wantErr: false,
		},
		{
			name: "several ranges in a row",
			args: args{
				s:   "last week this month next month next week",
				ref: now,
				f: func(source string, r Range) string {
					return fmt.Sprintf("%s to %s", r.Start(), r.End())
				},
				options: nil,
			},
			want:    "2022-09-18 00:00:00 +0000 UTC to 2022-09-24 23:59:59 +0000 UTC 2022-09-01 00:00:00 +0000 UTC to 2022-09-30 23:59:59 +0000 UTC 2022-10-01 00:00:00 +0000 UTC to 2022-10-31 23:59:59 +0000 UTC 2022-10-02 00:00:00 +0000 UTC to 2022-10-08 23:59:59 +0000 UTC",
			wantErr: false,
		},
		{
			name: "days and smaller get skipped",
			args: args{
				s:   "last week today next week now",
				ref: now,
				f: func(source string, r Range) string {
					return fmt.Sprintf("%s to %s", r.Start(), r.End())
				},
				options: nil,
			},
			want:    "2022-09-18 00:00:00 +0000 UTC to 2022-09-24 23:59:59 +0000 UTC today 2022-10-02 00:00:00 +0000 UTC to 2022-10-08 23:59:59 +0000 UTC now",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReplaceDateRangesByFunc(tt.args.s, tt.args.ref, tt.args.f, tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReplaceDateRangesByFunc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReplaceDateRangesByFunc() got = %v, want %v", got, tt.want)
			}
		})
	}
}
