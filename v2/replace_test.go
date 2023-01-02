package anytime

import (
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

func BenchmarkReplaceAllRangesByFunc_doItNowPoem(b *testing.B) {
	now := time.UnixMilli(rand.Int63())
	s := `
Do It Now
by Berton Braley
IF WITH PLEASURE you are viewing any work a man is doing,
If you like him or you love him, tell him now;
Don't withhold your approbation till the parson makes oration
And he lies with snowy lilies on his brow;
No matter how you shout it he won't really care about it;
He won't know how many teardrops you have shed;
If you think some praise is due him now's the time to slip it to him,
For he cannot read his tombstone when he's dead.`
	nowRx := regexp.MustCompile(`(?i)\bnow\b`)
	want := nowRx.ReplaceAllString(s, fmt.Sprintf("%v", now.UnixMilli()))
	f := func(_ string, r Range) string {
		return fmt.Sprintf("%v", r.Start().UnixMilli())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got, err := ReplaceAllRangesByFunc(s, now, Past, f)
		if err != nil {
			b.Fatal(err)
		}
		if got != want {
			b.Errorf("got = %v\n\nwant %v", got, want)
		}
	}
}

func TestReplaceAllRangesByFunc_nows(t *testing.T) {
	now := time.UnixMilli(rand.Int63())
	nowRx := regexp.MustCompile(`(?i)\bnow\b`)
	f := func(src string, r Range) string {
		return fmt.Sprintf("%v", r.Start().UnixMilli())
	}
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"one space", " "},
		{"verbiage", "This string contains no times or dates."},
		{"now", "now"},
		{"space now", " now"},
		{"now space", "now "},
		{"space now space", " now "},
		{"all caps now", "NOW"},
		{"now verbiage", "now is the time"},
		{"verbiage now", "the time is now"},
		{"verbiage now verbiage", "Without a doubt now is the time."},
		{"verbiage now punctuation verbiage", "If you don't know me by now, you will never know me."},
		{"now now", "now now"},
		{"nownow", "nownow"},
		{"noise with two nows", "a;slas 😅dflasdjfla now laksjdfsdf  xxc,mnv as2w0  @#R$@$ 😑now😵‍💫  ;xlc x;c,nv.s,hriop4qu-u98dsvfjkldfljs $!@@#$WERTwe5u682470sZ)(*&Y)*("},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReplaceAllRangesByFunc(tt.input, now, Future, f)
			if err != nil {
				t.Fatal(err)
			}
			want := nowRx.ReplaceAllString(tt.input, fmt.Sprintf("%v", now.UnixMilli()))
			if got != want {
				t.Errorf("\ngot\n%q\nwant\n%q", got, want)
			}
		})
	}
}

func TestReplaceAllRangesByFunc_noReplacementCrossTalk(t *testing.T) {
	input := "last year week"
	f := func(src string, r Range) string {
		return "maybe next"
	}
	want := "maybe next week"
	got, err := ReplaceAllRangesByFunc(input, time.Time{}, Future, f)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got = %q, want %q", got, want)
	}
}

func TestReplaceAllRangesByFunc_lastYearReplacements(t *testing.T) {
	now := time.UnixMilli(rand.Int63())
	ly := lastYear(now)
	f := func(src string, r Range) string {
		return fmt.Sprintf("%v", r.Start().UnixMilli())
	}
	inputs := []string{
		"last year",
		`"last year"`,
		"a last year",
		"last year a",
		"a last year a",
		"a last year last year a",
		"a (last year, last year) a",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			got, err := ReplaceAllRangesByFunc(input, now, Past, f)
			if err != nil {
				t.Fatal(err)
			}
			want := strings.ReplaceAll(input, "last year", fmt.Sprintf("%v", ly.Start().UnixMilli()))
			if got != want {
				t.Errorf("got = %v, want %v", got, want)
			}
		})
	}
}

func TestReplaceAllRangesByFunc_lastYearPlusVerbiage(t *testing.T) {
	now := time.UnixMilli(rand.Int63())
	wantRange := truncateYear(now.AddDate(-1, 0, 0))
	f := func(src string, r Range) string {
		return fmt.Sprintf("%v", r.Start().UnixMilli())
	}
	inputStr := `foo last year bar`
	gotStr, err := ReplaceAllRangesByFunc(inputStr, now, Past, f)
	if err != nil {
		t.Fatal(err)
	}
	unixStr := fmt.Sprintf("%v", wantRange.Start().UnixMilli())
	wantStr := strings.ReplaceAll(inputStr, "last year", unixStr)
	if gotStr != wantStr {
		t.Errorf("gotStr = %v, wantStr %v", gotStr, wantStr)
	}
}

func TestReplaceAllRangesByFunc_ok(t *testing.T) {
	now := time.Date(2022, 9, 29, 2, 48, 33, 123, time.UTC)

	var cases = []struct {
		Input     string
		WantRange Range
	}{
		// years
		{`Last year`, truncateYear(now.AddDate(-1, 0, 0))},
		{`last  year`, truncateYear(now.AddDate(-1, 0, 0))},
		{`This year`, truncateYear(now)},
		{`Next year`, truncateYear(now.AddDate(1, 0, 0))},

		// yesterday
		{`Yesterday`, truncateDay(now.AddDate(0, 0, -1))},

		// today
		{`Today`, truncateDay(now)},

		// tomorrow
		{`Tomorrow`, truncateDay(now.AddDate(0, 0, 1))},

		//// weeks
		{`Last week`, truncateWeek(now.AddDate(0, 0, -7))},
		{`This week`, truncateWeek(now)},
		{`Next week`, truncateWeek(now.AddDate(0, 0, 7))},

		// months
		{`Last month`, truncateMonth(now.AddDate(0, -1, 0))},
		{`This month`, truncateMonth(now)},
		{`Next month`, truncateMonth(now.AddDate(0, 1, 0))},

		// absolute dates
		{"January 2017", truncateMonth(time.Date(2017, 1, 1, 0, 0, 0, 0, now.Location()))},
		{"Jan 2017", truncateMonth(time.Date(2017, 1, 1, 0, 0, 0, 0, now.Location()))},
		{"March 31", truncateDay(time.Date(2023, 3, 31, 0, 0, 0, 0, now.Location()))},
		{"January, 2017", truncateMonth(time.Date(2017, 1, 1, 0, 0, 0, 0, now.Location()))},
		{"Jan, 2017", truncateMonth(time.Date(2017, 1, 1, 0, 0, 0, 0, now.Location()))},
		{"April 3 2017", truncateDay(time.Date(2017, 4, 3, 0, 0, 0, 0, now.Location()))},
		{"April 3, 2017", truncateDay(time.Date(2017, 4, 3, 0, 0, 0, 0, now.Location()))},
		{"Oct 7, 1970", truncateDay(time.Date(1970, 10, 7, 0, 0, 0, 0, now.Location()))},
		{"Oct 7, 1970 UTC+3", truncateDay(time.Date(1970, 10, 7, 0, 0, 0, 0, fixedZone(3)))},
		{"Oct 7 1970", truncateDay(time.Date(1970, 10, 7, 0, 0, 0, 0, now.Location()))},
		{"Oct. 7, 1970", truncateDay(time.Date(1970, 10, 7, 0, 0, 0, 0, now.Location()))},

		{"September 17, 2012 UTC+7", truncateDay(time.Date(2012, 9, 17, 10, 9, 0, 0, fixedZone(7)))},
		{"September 17, 2012 UTC-7", truncateDay(time.Date(2012, 9, 17, 10, 9, 0, 0, fixedZone(-7)))},
		{"September 17, 2012", truncateDay(time.Date(2012, 9, 17, 10, 9, 0, 0, now.Location()))},
		{"7 oct 1970", truncateDay(time.Date(1970, 10, 7, 0, 0, 0, 0, now.Location()))},
		{"7 oct, 1970", truncateDay(time.Date(1970, 10, 7, 0, 0, 0, 0, now.Location()))},
		{"03 February 2013", truncateDay(time.Date(2013, 2, 3, 0, 0, 0, 0, now.Location()))},
		{"2 July 2013", truncateDay(time.Date(2013, 7, 2, 0, 0, 0, 0, now.Location()))},
		{"2022 Feb 1", truncateDay(time.Date(2022, 2, 1, 0, 0, 0, 0, now.Location()))},
		//// yyyy/mm/dd, dd/mm/yyyy etc.
		{"2014/3/31", truncateDay(time.Date(2014, 3, 31, 0, 0, 0, 0, now.Location()))},
		{"2014/3/31 UTC", truncateDay(time.Date(2014, 3, 31, 0, 0, 0, 0, time.UTC))},
		{"2014/3/31 UTC+1", truncateDay(time.Date(2014, 3, 31, 0, 0, 0, 0, fixedZone(1)))},
		{"2014/03/31", truncateDay(time.Date(2014, 3, 31, 0, 0, 0, 0, now.Location()))},
		{"2014/03/31 UTC-1", truncateDay(time.Date(2014, 3, 31, 0, 0, 0, 0, fixedZone(-1)))},
		{"2014-04-26", truncateDay(time.Date(2014, 4, 26, 0, 0, 0, 0, now.Location()))},
		{"2014-4-26", truncateDay(time.Date(2014, 4, 26, 0, 0, 0, 0, now.Location()))},
		{"2014-4-6", truncateDay(time.Date(2014, 4, 6, 0, 0, 0, 0, now.Location()))},
		{"31/3/2014 UTC-8", truncateDay(time.Date(2014, 3, 31, 0, 0, 0, 0, fixedZone(-8)))},
		{"31-3-2014 UTC-8", truncateDay(time.Date(2014, 3, 31, 0, 0, 0, 0, fixedZone(-8)))},
		{"31/3/2014", truncateDay(time.Date(2014, 3, 31, 0, 0, 0, 0, now.Location()))},
		{"31-3-2014", truncateDay(time.Date(2014, 3, 31, 0, 0, 0, 0, now.Location()))},

		//// days
		{`One day ago`, truncateDay(now.Add(-24 * time.Hour))},
		{`1 day ago`, truncateDay(now.Add(-24 * time.Hour))},
		{`3 days ago`, truncateDay(now.Add(-3 * 24 * time.Hour))},
		{`Three days ago`, truncateDay(now.Add(-3 * 24 * time.Hour))},
		{`1 day from now`, truncateDay(now.Add(24 * time.Hour))},
		{`two days from now`, truncateDay(now.AddDate(0, 0, 2))},
		{`two days from today`, truncateDay(now.AddDate(0, 0, 2))},
		{`two days hence`, truncateDay(now.AddDate(0, 0, 2))},

		// weeks
		{`1 week ago`, truncateWeek(now.Add(-7 * 24 * time.Hour))},
		{`2 weeks ago`, truncateWeek(now.Add(-2 * 7 * 24 * time.Hour))},
		{`A week from now`, truncateWeek(now.Add(7 * 24 * time.Hour))},
		{`A week from today`, truncateWeek(now.Add(7 * 24 * time.Hour))},
		{`2 weeks hence`, truncateWeek(now.Add(2 * 7 * 24 * time.Hour))},

		// months
		{`A month ago`, truncateMonth(now.AddDate(0, -1, 0))},
		{`1 month ago`, truncateMonth(now.AddDate(0, -1, 0))},
		{`2 months ago`, truncateMonth(now.AddDate(0, -2, 0))},
		{`12 months ago`, truncateMonth(now.AddDate(0, -12, 0))},
		{`twelve months ago`, truncateMonth(now.AddDate(0, -12, 0))},
		{`A month from now`, truncateMonth(now.AddDate(0, 1, 0))},
		{`One month hence`, truncateMonth(now.AddDate(0, 1, 0))},
		{`1 month from now`, truncateMonth(now.AddDate(0, 1, 0))},
		{`2 months from now`, truncateMonth(now.AddDate(0, 2, 0))},

		// years
		{`One year ago`, truncateYear(now.AddDate(-1, 0, 0))},
		{`One year from now`, truncateYear(now.AddDate(1, 0, 0))},
		{`One year from today`, truncateYear(now.AddDate(1, 0, 0))},
		{`One year hence`, truncateYear(now.AddDate(1, 0, 0))},
		{`Two years ago`, truncateYear(now.AddDate(-2, 0, 0))},
		{`2 years ago`, truncateYear(now.AddDate(-2, 0, 0))},
		{`This year`, truncateYear(now)},
		{`1999AD`, truncateYear(time.Date(1999, 1, 1, 0, 0, 0, 0, now.Location()))},
		{`1999 AD`, truncateYear(time.Date(1999, 1, 1, 0, 0, 0, 0, now.Location()))},
		{`2008CE`, truncateYear(time.Date(2008, 1, 1, 0, 0, 0, 0, now.Location()))},
		{`2008 CE`, truncateYear(time.Date(2008, 1, 1, 0, 0, 0, 0, now.Location()))},

		// RFC3339
		{"2006-01-02T15:04:05Z", Range{time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC), time.Second}},
		{"1990-12-31T15:59:59-08:00", Range{time.Date(1990, 12, 31, 15, 59, 59, 0, time.FixedZone("", -8*60*60)), time.Second}},

		//// minutes
		//{`a minute from now`, now.Add(time.Minute)},
		//{`a minute ago`, now.Add(-time.Minute)},
		//{`1 minute ago`, now.Add(-time.Minute)},
		//
		//{`5 minutes ago`, now.Add(-5 * time.Minute)},
		//{`five minutes ago`, now.Add(-5 * time.Minute)},
		//{`   5    minutes  ago   `, now.Add(-5 * time.Minute)},
		//{`2 minutes from now`, now.Add(2 * time.Minute)},
		//{`two minutes from now`, now.Add(2 * time.Minute)},
		//
		//// hours
		//{`an hour from now`, now.Add(time.Hour)},
		//{`an hour ago`, now.Add(-time.Hour)},
		//{`1 hour ago`, now.Add(-time.Hour)},
		//{`6 hours ago`, now.Add(-6 * time.Hour)},
		//{`1 hour from now`, now.Add(time.Hour)},
		//
		//// times
		//{"noon", dateAtTime(today, 12, 0, 0)},
		//{"5:35:52pm", dateAtTime(today, 12+5, 35, 52)},
		//{`10am`, dateAtTime(today, 10, 0, 0)},
		//{`10 am`, dateAtTime(today, 10, 0, 0)},
		//{`5pm`, dateAtTime(today, 12+5, 0, 0)},
		//{`10:25am`, dateAtTime(today, 10, 25, 0)},
		//{`1:05pm`, dateAtTime(today, 12+1, 5, 0)},
		//{`10:25:10am`, dateAtTime(today, 10, 25, 10)},
		//{`1:05:10pm`, dateAtTime(today, 12+1, 5, 10)},
		//{`10:25`, dateAtTime(today, 10, 25, 0)},
		//{`10:25:30`, dateAtTime(today, 10, 25, 30)},
		//{`17:25:30`, dateAtTime(today, 17, 25, 30)},
		//
		//// dates with times
		//{"On Friday at noon UTC", timeInLocation(dateAtTime(nextWeekdayFrom(now, time.Friday).Time, 12, 0, 0), time.UTC)},
		//{"On Tuesday at 11am UTC", timeInLocation(dateAtTime(nextWeekdayFrom(now, time.Tuesday).Time, 11, 0, 0), time.UTC)},
		//{"On 3 feb 2025 at 5:35:52pm", time.Date(2025, time.February, 3, 12+5, 35, 52, 0, now.Location())},
		//{"3 feb 2025 at 5:35:52pm", time.Date(2025, time.February, 3, 12+5, 35, 52, 0, now.Location())},
		//{`3 days ago at 11:25am`, dateAtTime(now.Add(-3*24*time.Hour), 11, 25, 0)},
		//{`3 days from now at 14:26`, dateAtTime(now.Add(3*24*time.Hour), 14, 26, 0)},
		//{`2 weeks ago at 8am`, dateAtTime(now.Add(-2*7*24*time.Hour), 8, 0, 0)},
		//{`Today at 10am`, dateAtTime(now, 10, 0, 0)},
		//{`10am today`, dateAtTime(now, 10, 0, 0)},
		//{`Yesterday 10am`, dateAtTime(now.AddDate(0, 0, -1), 10, 0, 0)},
		//{`10am yesterday`, dateAtTime(now.AddDate(0, 0, -1), 10, 0, 0)},
		//{`Yesterday at 10am`, dateAtTime(now.AddDate(0, 0, -1), 10, 0, 0)},
		//{`Yesterday at 10:15am`, dateAtTime(now.AddDate(0, 0, -1), 10, 15, 0)},
		//{`Tomorrow 10am`, dateAtTime(now.AddDate(0, 0, 1), 10, 0, 0)},
		//{`10am tomorrow`, dateAtTime(now.AddDate(0, 0, 1), 10, 0, 0)},
		//{`Tomorrow at 10am`, dateAtTime(now.AddDate(0, 0, 1), 10, 0, 0)},
		//{`Tomorrow at 10:15am`, dateAtTime(now.AddDate(0, 0, 1), 10, 15, 0)},
		//{`10:15am tomorrow`, dateAtTime(now.AddDate(0, 0, 1), 10, 15, 0)},
		//{"Next dec 22nd at 3pm", timeInLocation(nextMonthDayTime(now, time.December, 22, 12+3, 0, 0), now.Location())},
		//{"Next December 25th at 7:30am UTC-7", timeInLocation(nextMonthDayTime(now, time.December, 25, 7, 30, 0), fixedZone(-7))},
		//{`Next December 23rd AT 5:25 PM`, nextMonthDayTime(now, time.December, 23, 12+5, 25, 0)},
		//{`Last December 23rd AT 5:25 PM`, prevMonthDayTime(now, time.December, 23, 12+5, 25, 0)},
		//{`Last sunday at 5:30pm`, dateAtTime(lastWeekdayFrom(now, time.Sunday).Time, 12+5, 30, 0)},
		//{`Next sunday at 22:45`, dateAtTime(nextWeekdayFrom(now, time.Sunday).Time, 22, 45, 0)},
		//{`Next sunday at 22:45`, dateAtTime(nextWeekdayFrom(now, time.Sunday).Time, 22, 45, 0)},
		//{`November 3rd, 1986 at 4:30pm`, time.Date(1986, 11, 3, 12+4, 30, 0, 0, now.Location())},
		//{"September 17, 2012 at 10:09am UTC", time.Date(2012, 9, 17, 10, 9, 0, 0, time.UTC)},
		//{"September 17, 2012 at 10:09am UTC-8", time.Date(2012, 9, 17, 10, 9, 0, 0, fixedZone(-8))},
		//{"September 17, 2012 at 10:09am UTC+8", time.Date(2012, 9, 17, 10, 9, 0, 0, fixedZone(8))},
		//{"September 17, 2012, 10:11:09", time.Date(2012, 9, 17, 10, 11, 9, 0, now.Location())},
		//{"September 17, 2012, 10:11", time.Date(2012, 9, 17, 10, 11, 0, 0, now.Location())},
		//{"September 17, 2012 10:11", time.Date(2012, 9, 17, 10, 11, 0, 0, now.Location())},
		//{"September 17 2012 10:11", time.Date(2012, 9, 17, 10, 11, 0, 0, now.Location())},
		//{"September 17 2012 at 10:11", time.Date(2012, 9, 17, 10, 11, 0, 0, now.Location())},

		// from A to B
		{
			"From 3 feb 2022 to 6 oct 2022",
			RangeFromTimes(
				time.Date(2022, 2, 3, 0, 0, 0, 0, now.Location()),
				time.Date(2022, 10, 7, 0, 0, 0, 0, now.Location()),
			),
		},
		// A to B
		{
			"3 feb 2022 to 6 oct 2022",
			RangeFromTimes(
				time.Date(2022, 2, 3, 0, 0, 0, 0, now.Location()),
				time.Date(2022, 10, 7, 0, 0, 0, 0, now.Location()),
			),
		},
		// A through B
		{
			"3 feb 2022 through 6 oct 2022",
			RangeFromTimes(
				time.Date(2022, 2, 3, 0, 0, 0, 0, now.Location()),
				time.Date(2022, 10, 7, 0, 0, 0, 0, now.Location()),
			),
		},
		// from A until B
		{
			"from 3 feb 2022 until 6 oct 2022",
			RangeFromTimes(
				time.Date(2022, 2, 3, 0, 0, 0, 0, now.Location()),
				time.Date(2022, 10, 7, 0, 0, 0, 0, now.Location()),
			),
		},
		//{
		//	"from tuesday at 5pm -12:00 until thursday 23:52 +14:00",
		//	RangeFromTimes(
		//		setLocation(setTime(nextWeekdayFrom(now, time.Tuesday).Time, 12+5, 0, 0, 0), fixedZone(-12)),
		//		setLocation(setTime(nextWeekdayFrom(now, time.Thursday).Time, 23, 52, 0, 0), fixedZone(14)),
		//	),
		//},
	}

	for _, c := range cases {
		t.Run(c.Input, func(t *testing.T) {
			var foundRanges []Range
			gotStr, err := ReplaceAllRangesByFunc(c.Input, now, Future, func(rs string, r Range) string {
				foundRanges = append(foundRanges, r)
				return "<range>"
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(foundRanges) == 0 {
				t.Fatal("no ranges found")
			}
			if len(foundRanges) > 1 {
				t.Fatalf("got %d ranges, want 1", len(foundRanges))
			}
			gotRange := foundRanges[0]
			if !gotRange.Equal(c.WantRange) {
				t.Fatalf("got range %v, want %v", gotRange, c.WantRange)
			}
			wantStr := "<range>"
			if gotStr != wantStr {
				t.Fatalf("got string %q, want %q", gotStr, wantStr)
			}
		})
	}
}

func TestReplaceAllRangesByFunc_colorMonth(t *testing.T) {
	now := time.UnixMilli(rand.Int63())
	var cases = []struct {
		Input     string
		WantRange Range
	}{
		// color month
		// http://www.jdawiseman.com/papers/trivia/futures.html
		{"White october", nextSpecificMonth(now, time.October)},
		{"Red october", truncateMonth(nextSpecificMonth(now, time.October).Start().AddDate(1, 0, 0))},
		{"Green october", truncateMonth(nextSpecificMonth(now, time.October).Start().AddDate(2, 0, 0))},
		{"Blue october", truncateMonth(nextSpecificMonth(now, time.October).Start().AddDate(3, 0, 0))},
		{"Gold october", truncateMonth(nextSpecificMonth(now, time.October).Start().AddDate(4, 0, 0))},
		{"Purple october", truncateMonth(nextSpecificMonth(now, time.October).Start().AddDate(5, 0, 0))},
		{"Orange october", truncateMonth(nextSpecificMonth(now, time.October).Start().AddDate(6, 0, 0))},
		{"Pink october", truncateMonth(nextSpecificMonth(now, time.October).Start().AddDate(7, 0, 0))},
		{"Silver october", truncateMonth(nextSpecificMonth(now, time.October).Start().AddDate(8, 0, 0))},
		{"Copper october", truncateMonth(nextSpecificMonth(now, time.October).Start().AddDate(9, 0, 0))},
	}

	for _, c := range cases {
		t.Run(c.Input, func(t *testing.T) {
			var gotRanges []Range
			gotStr, err := ReplaceAllRangesByFunc(c.Input, now, Past, func(src string, r Range) string {
				gotRanges = append(gotRanges, r)
				return "<range>"
			})
			if err != nil {
				t.Fatal(err)
			}
			wantStr := "<range>"
			if gotStr != wantStr {
				t.Fatalf("gotStr = %q, wantStr %q", gotStr, wantStr)
			}
			wantRanges := []Range{c.WantRange}
			if !reflect.DeepEqual(gotRanges, wantRanges) {
				t.Fatalf("gotRanges = %#v, wantRanges %#v", gotRanges, wantRanges)
			}
		})
	}
}

// TestParse_futurePast tests dates and times that are ambiguously in the past
// or the future.
func TestReplaceAllRangesByFunc_ambiguitiesResolvedByDirectionPreference(t *testing.T) {
	var now = time.Date(2022, 9, 29, 2, 48, 33, 123, time.UTC)
	tests := []struct {
		input      string
		wantFuture Range
		wantPast   Range
	}{
		{
			"December",
			truncateMonth(nextMonthDayTime(now, time.December, 20, 0, 0, 0)),
			truncateMonth(prevMonthDayTime(now, time.December, 20, 0, 0, 0)),
		},
		{
			"December 20",
			truncateDay(nextMonthDayTime(now, time.December, 20, 0, 0, 0)),
			truncateDay(prevMonthDayTime(now, time.December, 20, 0, 0, 0)),
		},
		//{
		//	"Thursday",
		//	nextWeekdayFrom(now, time.Thursday).Start(),
		//	lastWeekdayFrom(now, time.Thursday).Start(),
		//},
		//{
		//	"On thursday",
		//	nextWeekdayFrom(now, time.Thursday).Start(),
		//	lastWeekdayFrom(now, time.Thursday).Start(),
		//},
		//{
		//	"December 20 at 9pm",
		//	nextMonthDayTime(now, time.December, 20, 21, 0, 0),
		//	prevMonthDayTime(now, time.December, 20, 21, 0, 0),
		//},
		//{
		//	"Thursday at 23:59",
		//	setTime(nextWeekdayFrom(now, time.Thursday).Time, 23, 59, 0, 0),
		//	setTime(lastWeekdayFrom(now, time.Thursday).Time, 23, 59, 0, 0),
		//},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Future
			var gotRanges []Range
			_, err := ReplaceAllRangesByFunc(tt.input, now, Future, func(src string, r Range) string {
				gotRanges = append(gotRanges, r)
				return ""
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(gotRanges) != 1 {
				t.Fatalf("got %d ranges, want 1", len(gotRanges))
			}
			gotRange := gotRanges[0]
			if !gotRange.Equal(tt.wantFuture) {
				t.Fatalf("got range %v, want %v", gotRange, tt.wantFuture)
			}

			// Past
			gotRanges = nil
			_, err = ReplaceAllRangesByFunc(tt.input, now, Past, func(src string, r Range) string {
				gotRanges = append(gotRanges, r)
				return ""
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(gotRanges) != 1 {
				t.Fatalf("got %d ranges, want 1", len(gotRanges))
			}
			gotRange = gotRanges[0]
			if !gotRange.Equal(tt.wantPast) {
				t.Fatalf("got range %v, want %v", gotRange, tt.wantPast)
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
			if got := truncateWeek(tt.args.t).Start(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("truncateWeek() = %v, want %v", got, tt.want)
			}
		})
	}
}
