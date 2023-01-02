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

func FuzzReplaceAllRangesByFunc_stringsUnchangedWhenFReturnsSrc(f *testing.F) {
	f.Add("", rand.Int63(), false)
	f.Add("", rand.Int63(), true)
	f.Add("2022ad", rand.Int63(), true)
	f.Add("December", rand.Int63(), false)
	f.Fuzz(func(t *testing.T, s string, nowMillis int64, defaultToFuture bool) {
		var dir Direction = Past
		if defaultToFuture {
			dir = Future
		}
		now := time.UnixMilli(nowMillis)
		s2, err := ReplaceAllRangesByFunc(s, now, dir, func(src string, r Range) string {
			return src
		})
		if err != nil {
			t.Fatal(err)
		}
		if s2 != s {
			t.Fatalf("got %q, want %q", s2, s)
		}
	})
}
