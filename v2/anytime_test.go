package anytime

import (
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"testing"
	"time"
)

func TestReplaceAllRangesByFunc_nows(t *testing.T) {
	now := time.UnixMilli(rand.Int63())
	nowRx := regexp.MustCompile(`(?i)\bnow\b`)
	f := func(source string, r Range) string {
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
			got, err := ReplaceAllRangesByFunc(tt.input, now, f, Future)
			if err != nil {
				t.Fatal(err)
			}
			want := nowRx.ReplaceAllString(tt.input, fmt.Sprintf("%v", now.UnixMilli()))
			if got != want {
				t.Errorf("got = %v, want %v", got, want)
			}
		})
	}
}

func BenchmarkReplaceAllRangesByFunc(b *testing.B) {
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
	f := func(source string, r Range) string {
		return fmt.Sprintf("%v", r.Start().UnixMilli())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got, err := ReplaceAllRangesByFunc(s, now, f, Past)
		if err != nil {
			b.Fatal(err)
		}
		if got != want {
			b.Errorf("got = %v\n\nwant %v", got, want)
		}
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
		{`This year`, truncateYear(now)},
		{`Next year`, truncateYear(now.AddDate(1, 0, 0))},
		//
		//// today
		//{`Today`, truncateDay(now)},
		//
		//// yesterday
		//{`Yesterday`, truncateDay(now.AddDate(0, 0, -1))},
		//
		//// tomorrow
		//{`Tomorrow`, truncateDay(now.AddDate(0, 0, 1))},
		//
		//// weeks
		//{`Last week`, truncateWeek(now.AddDate(0, 0, -7))},
		//{`Next week`, truncateWeek(now.AddDate(0, 0, 7))},
		//
		//// months
		//{`Last month`, truncateMonth(now.AddDate(0, -1, 0))},
		//{`Next month`, truncateMonth(now.AddDate(0, 1, 0))},
		//
		//// past weekdays
		//{`Last sunday`, lastWeekdayFrom(now, time.Sunday)},
		//{`Last monday`, lastWeekdayFrom(now, time.Monday)},
		//{`Last Monday`, lastWeekdayFrom(now, time.Monday)},
		//{`Last tuesday`, lastWeekdayFrom(now, time.Tuesday)},
		//{`Last wednesday`, lastWeekdayFrom(now, time.Wednesday)},
		//{`Last Thursday`, lastWeekdayFrom(now, time.Thursday)},
		//{`Last Friday`, lastWeekdayFrom(now, time.Friday)},
		//{`Last saturday`, lastWeekdayFrom(now, time.Saturday)},
		//
		//// future weekdays
		//{`Next sunday`, nextWeekdayFrom(now, time.Sunday)},
		//{`Next monday`, nextWeekdayFrom(now, time.Monday)},
		//{`Next tuesday`, nextWeekdayFrom(now, time.Tuesday)},
		//{`Next wednesday`, nextWeekdayFrom(now, time.Wednesday)},
		//{`Next thursday`, nextWeekdayFrom(now, time.Thursday)},
		//{`Next friday`, nextWeekdayFrom(now, time.Friday)},
		//{`Next saturday`, nextWeekdayFrom(now, time.Saturday)},
		//
		//// months
		//{`Last january`, lastSpecificMonth(now, time.January)},
		//{`Next january`, nextSpecificMonth(now, time.January)},

		//// absolute dates
		//{"January 2017", time.Date(2017, 1, 1, 0, 0, 0, 0, now.Location())},
		//// {"March 31", time.Date(2022, 3, 31, 0, 0, 0, 0, now.Location())},
		//{"January, 2017", time.Date(2017, 1, 1, 0, 0, 0, 0, now.Location())},
		//{"April 3 2017", time.Date(2017, 4, 3, 0, 0, 0, 0, now.Location())},
		//{"April 3, 2017", time.Date(2017, 4, 3, 0, 0, 0, 0, now.Location())},
		//{"Oct 7, 1970", time.Date(1970, 10, 7, 0, 0, 0, 0, now.Location())},
		//{"Oct 7 1970", time.Date(1970, 10, 7, 0, 0, 0, 0, now.Location())},
		//{"Oct. 7, 1970", time.Date(1970, 10, 7, 0, 0, 0, 0, now.Location())},
		////{"September 17, 2012 UTC+7", time.Date(2012, 9, 17, 10, 9, 0, 0, fixedZone(7))},
		////{"September 17, 2012 UTC-7", time.Date(2012, 9, 17, 10, 9, 0, 0, fixedZone(-7))},
		//{"September 17, 2012", time.Date(2012, 9, 17, 10, 9, 0, 0, now.Location())},
		//{"7 oct 1970", time.Date(1970, 10, 7, 0, 0, 0, 0, now.Location())},
		//{"7 oct, 1970", time.Date(1970, 10, 7, 0, 0, 0, 0, now.Location())},
		//{"03 February 2013", time.Date(2013, 2, 3, 0, 0, 0, 0, now.Location())},
		//{"2 July 2013", time.Date(2013, 7, 2, 0, 0, 0, 0, now.Location())},
		//{"2022 Feb 1", time.Date(2022, 2, 1, 0, 0, 0, 0, now.Location())},
		//// yyyy/mm/dd, dd/mm/yyyy etc.
		//{"2014/3/31", time.Date(2014, 3, 31, 0, 0, 0, 0, now.Location())},
		////{"2014/3/31 UTC", time.Date(2014, 3, 31, 0, 0, 0, 0, location("UTC"))},
		////{"2014/3/31 UTC+1", time.Date(2014, 3, 31, 0, 0, 0, 0, fixedZone(1))},
		//{"2014/03/31", time.Date(2014, 3, 31, 0, 0, 0, 0, now.Location())},
		////{"2014/03/31 UTC-1", time.Date(2014, 3, 31, 0, 0, 0, 0, fixedZone(-1))},
		//{"2014-04-26", time.Date(2014, 4, 26, 0, 0, 0, 0, now.Location())},
		//{"2014-4-26", time.Date(2014, 4, 26, 0, 0, 0, 0, now.Location())},
		//{"2014-4-6", time.Date(2014, 4, 6, 0, 0, 0, 0, now.Location())},
		////{"31/3/2014 UTC-8", time.Date(2014, 3, 31, 0, 0, 0, 0, fixedZone(-8))},
		////{"31-3-2014 UTC-8", time.Date(2014, 3, 31, 0, 0, 0, 0, fixedZone(-8))},
		//{"31/3/2014", time.Date(2014, 3, 31, 0, 0, 0, 0, now.Location())},
		//{"31-3-2014", time.Date(2014, 3, 31, 0, 0, 0, 0, now.Location())},

		// color month
		// http://www.jdawiseman.com/papers/trivia/futures.html
		//{"White october", nextSpecificMonth(now, time.October)},
		//{"Red october", nextSpecificMonth(now, time.October).Time.AddDate(1, 0, 0)},
		//{"Green october", nextSpecificMonth(now, time.October).Time.AddDate(2, 0, 0)},
		//{"Blue october", nextSpecificMonth(now, time.October).Time.AddDate(3, 0, 0)},
		//{"Gold october", nextSpecificMonth(now, time.October).Time.AddDate(4, 0, 0)},
		//{"Purple october", nextSpecificMonth(now, time.October).Time.AddDate(5, 0, 0)},
		//{"Orange october", nextSpecificMonth(now, time.October).Time.AddDate(6, 0, 0)},
		//{"Pink october", nextSpecificMonth(now, time.October).Time.AddDate(7, 0, 0)},
		//{"Silver october", nextSpecificMonth(now, time.October).Time.AddDate(8, 0, 0)},
		//{"Copper october", nextSpecificMonth(now, time.October).Time.AddDate(9, 0, 0)},
		//// days
		//{`One day ago`, truncateDay(now.Add(-24 * time.Hour))},
		//{`1 day ago`, truncateDay(now.Add(-24 * time.Hour))},
		//{`3 days ago`, truncateDay(now.Add(-3 * 24 * time.Hour))},
		//{`Three days ago`, truncateDay(now.Add(-3 * 24 * time.Hour))},
		//{`1 day from now`, truncateDay(now.Add(24 * time.Hour))},
		//{`two days from now`, truncateDay(now.AddDate(0, 0, 2))},
		//{`two days from today`, truncateDay(now.AddDate(0, 0, 2))},

		//// weeks
		//{`1 week ago`, truncateWeek(now.Add(-7 * 24 * time.Hour))},
		//{`2 weeks ago`, truncateWeek(now.Add(-2 * 7 * 24 * time.Hour))},
		//{`A week from now`, truncateWeek(now.Add(7 * 24 * time.Hour))},
		//{`A week from today`, truncateWeek(now.Add(7 * 24 * time.Hour))},
		//
		//// months
		//{`A month ago`, truncateMonth(now.AddDate(0, -1, 0))},
		//{`1 month ago`, truncateMonth(now.AddDate(0, -1, 0))},
		//{`2 months ago`, truncateMonth(now.AddDate(0, -2, 0))},
		//{`12 months ago`, truncateMonth(now.AddDate(0, -12, 0))},
		//{`twelve months ago`, truncateMonth(now.AddDate(0, -12, 0))},
		//{`A month from now`, truncateMonth(now.AddDate(0, 1, 0))},
		//{`One month hence`, truncateMonth(now.AddDate(0, 1, 0))},
		//{`1 month from now`, truncateMonth(now.AddDate(0, 1, 0))},
		//{`2 months from now`, truncateMonth(now.AddDate(0, 2, 0))},
		//{`Last January`, lastSpecificMonth(now, time.January)},
		//{`Last january`, lastSpecificMonth(now, time.January)},
		//{`Next january`, nextSpecificMonth(now, time.January)},
		//
		//// years
		//{`One year ago`, truncateYear(now.AddDate(-1, 0, 0))},
		//{`One year from now`, truncateYear(now.AddDate(1, 0, 0))},
		//{`One year from today`, truncateYear(now.AddDate(1, 0, 0))},
		//{`Two years ago`, truncateYear(now.AddDate(-2, 0, 0))},
		//{`2 years ago`, truncateYear(now.AddDate(-2, 0, 0))},
		//{`This year`, truncateYear(now)},
		//{`1999AD`, truncateYear(time.Date(1999, 1, 1, 0, 0, 0, 0, now.Location()))},
		//{`1999 AD`, truncateYear(time.Date(1999, 1, 1, 0, 0, 0, 0, now.Location()))},
		//{`2008CE`, truncateYear(time.Date(2008, 1, 1, 0, 0, 0, 0, now.Location()))},
		//{`2008 CE`, truncateYear(time.Date(2008, 1, 1, 0, 0, 0, 0, now.Location()))},

		// formats from the Go time package:
		// ANSIC
		//{"Mon Jan  2 15:04:05 2006", time.Date(2006, 1, 2, 15, 4, 5, 0, now.Location())},
		//// RubyDate
		//{"Mon Jan 02 15:04:05 -0700 2006", time.Date(2006, 1, 2, 15, 4, 5, 0, fixedZone(-7))},
		//// RFC1123Z
		//{"Mon, 02 Jan 2006 15:04:05 -0700", time.Date(2006, 1, 2, 15, 4, 5, 0, fixedZone(-7))},
		//{"Mon 02 Jan 2006 15:04:05 -0700", time.Date(2006, 1, 2, 15, 4, 5, 0, fixedZone(-7))},
		//// RFC3339
		//{"2006-01-02T15:04:05Z", time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)},
		//{"1990-12-31T15:59:59-08:00", time.Date(1990, 12, 31, 15, 59, 59, 0, time.FixedZone("", -8*60*60))},

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
	}

	for _, c := range cases {
		t.Run(c.Input, func(t *testing.T) {
			var foundRanges []Range
			_, err := ReplaceAllRangesByFunc(c.Input, now, func(_ string, r Range) string {
				foundRanges = append(foundRanges, r)
				return ""
			}, Future)
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
		})
	}
}

//// TestParse_futurePast tests dates and times that are ambiguously in the past
//// or the future.
//func TestParse_futurePast(t *testing.T) {
//	var now = time.Date(2022, 9, 29, 2, 48, 33, 123, time.UTC)
//	tests := []struct {
//		input      string
//		wantFuture time.Time
//		wantPast   time.Time
//	}{
//		{
//			"December 20",
//			nextMonthDayTime(now, time.December, 20, 0, 0, 0),
//			prevMonthDayTime(now, time.December, 20, 0, 0, 0),
//		},
//		{
//			"Thursday",
//			nextWeekdayFrom(now, time.Thursday).Start(),
//			lastWeekdayFrom(now, time.Thursday).Start(),
//		},
//		{
//			"On thursday",
//			nextWeekdayFrom(now, time.Thursday).Start(),
//			lastWeekdayFrom(now, time.Thursday).Start(),
//		},
//		//{
//		//	"December 20 at 9pm",
//		//	nextMonthDayTime(now, time.December, 20, 21, 0, 0),
//		//	prevMonthDayTime(now, time.December, 20, 21, 0, 0),
//		//},
//		//{
//		//	"Thursday at 23:59",
//		//	setTime(nextWeekdayFrom(now, time.Thursday).Time, 23, 59, 0, 0),
//		//	setTime(lastWeekdayFrom(now, time.Thursday).Time, 23, 59, 0, 0),
//		//},
//	}
//	for _, tt := range tests {
//		t.Run(tt.input, func(t *testing.T) {
//			// future
//			v, err := Parse(tt.input)
//			if err != nil {
//				t.Fatal(err)
//			}
//			got, err := verbiageToTime(v, now, future)
//			if err != nil {
//				t.Fatal(err)
//			}
//			if !got.Equal(tt.wantFuture) {
//				t.Errorf("got = %v, want %v", got, tt.wantFuture)
//			}
//
//			// past
//			got, err = verbiageToTime(v, now, past)
//			if err != nil {
//				t.Fatal(err)
//			}
//			if !got.Equal(tt.wantPast) {
//				t.Errorf("got %v, want %v", got, tt.wantPast)
//			}
//		})
//	}
//}

//func TestParse_monthOnly(t *testing.T) {
//	tests := []struct {
//		input string
//		month time.Month
//	}{
//		{"January", time.January},
//		{"Jan", time.January},
//		{"February", time.February},
//		{"Feb", time.February},
//		{"March", time.March},
//		{"Mar", time.March},
//		{"April", time.April},
//		{"Apr", time.April},
//		{"May", time.May},
//		{"June", time.June},
//		{"Jun", time.June},
//		{"July", time.July},
//		{"Jul", time.July},
//		{"August", time.August},
//		{"Aug", time.August},
//		{"september", time.September},
//		{"sep", time.September},
//		{"october", time.October},
//		{"oct", time.October},
//		{"november", time.November},
//		{"nov", time.November},
//		{"december", time.December},
//		{"dec", time.December},
//	}
//	for _, tt := range tests {
//		t.Run(tt.input, func(t *testing.T) {
//			got, err := Parse(tt.input)
//
//			// future
//			if err != nil {
//				t.Fatal(err)
//				return
//			}
//			gotTime, err := verbiageToTime(got, now, future)
//			if err != nil {
//				t.Fatal(err)
//			}
//			futureDate := nextSpecificMonth(now, tt.month).Time
//			if !gotTime.Equal(futureDate) {
//				t.Errorf("got %v, want %v", gotTime, futureDate)
//			}
//
//			// past
//			gotTime, err = verbiageToTime(got, now, past)
//			if err != nil {
//				t.Fatal(err)
//			}
//			pastDate := lastSpecificMonth(now, tt.month).Time
//			if !gotTime.Equal(pastDate) {
//				t.Errorf("got %v, want %v", gotTime, pastDate)
//			}
//		})
//	}
//}

//func TestParse_future(t *testing.T) {
//	tests := []struct {
//		input      string
//		wantOutput time.Time
//	}{
//		{"january", nextSpecificMonth(now, time.January).Time},
//	}
//	for _, tt := range tests {
//		t.Run(tt.input, func(t *testing.T) {
//			got, err := Parse(tt.input)
//			if err != nil {
//				t.Fatal(err)
//			}
//			gotTime, err := verbiageToTime(got, now, future)
//			if err != nil {
//				t.Fatal(err)
//			}
//			if !gotTime.Equal(tt.wantOutput) {
//				t.Errorf("got %v, want %v", gotTime, tt.wantOutput)
//			}
//		})
//	}
//}

//// Benchmark parsing.
//func BenchmarkParse(b *testing.B) {
//	b.SetBytes(1)
//	want := time.Date(2022, time.December, 23, 0, 0, 0, 0, now.Location())
//	for i := 0; i < b.N; i++ {
//		v, err := Parse(`december 23, 2022`)
//		gotTime, err := verbiageToTime(v, now, future)
//		if err != nil {
//			b.Fatal(err)
//		}
//		if !gotTime.Equal(want) {
//			b.Errorf("got %v, want %v", gotTime, want)
//		}
//	}
//}

// TestParse_nonTemporal tests parsing with inputs that are not expected to
// result in time-like output.
//func TestParse_nonTemporal(t *testing.T) {
//	var badCases = []struct {
//		input string
//	}{
//		{``},
//		{`next thing`},
//		{`a`},
//		{`not a date or a time`},
//		{`right now`},
//		{`  right  now  `},
//		{`Message me in 2 minutes`},
//		{`Message me in 2 minutes from now`},
//		{`Remind me in 1 hour`},
//		{`Remind me in 1 hour from now`},
//		{`Remind me in 1 hour and 3 minutes from now`},
//		{`Remind me in an hour`},
//		{`Remind me in an hour from now`},
//		{`Remind me one day from now`},
//		{`Remind me in a day`},
//		{`Remind me in one day`},
//		{`Remind me in one day from now`},
//		{`Message me in a week`},
//		{`Message me in one week`},
//		{`Message me in one week from now`},
//		{`Message me in two weeks from now`},
//		{`Message me two weeks from now`},
//		{`Message me in two weeks`},
//		{`Remind me in 12 months from now at 6am`},
//		{`Remind me in a month`},
//		{`Remind me in 2 months`},
//		{`Remind me in a month from now`},
//		{`Remind me in 2 months from now`},
//		{`Remind me in one year from now`},
//		{`Remind me in a year`},
//		{`Remind me in a year from now`},
//		{`Restart the server in 2 days from now`},
//		{`Remind me on the 5th of next month`},
//		{`Remind me on the 5th of next month at 7am`},
//		{`Remind me at 7am on the 5th of next month`},
//		{`Remind me in one month from now`},
//		{`Remind me in one month from now at 7am`},
//		{`Remind me on the December 25th at 7am`},
//		{`Remind me at 7am on December 25th`},
//		{`Remind me on the 25th of December at 7am`},
//		{`Check logs in the past 5 minutes`},
//
//		// "1 minute" is a duration, not a time.
//		{`1 minute`},
//
//		// "one minute" is also a duration.
//		{`one minute`},
//
//		// "1 hour" is also a duration.
//		{`1 hour`},
//
//		// "1 day" is also a duration.
//		{`1 day`},
//
//		// "1 week" is also a duration.
//		{`1 week`},
//
//		// "1 month" is also a duration.
//		{`1 month`},
//
//		// "next 2 months" is a date range, not a time or a date.
//		{`next 2 months`},
//
//		// These are currently considered bad input, although they may
//		{`10`},
//		{`17`},
//
//		// Bare years don't have enough context to be confidently parsed as dates.
//		{`1999`},
//		{`2008`},
//
//		// Goofy input:
//		{`10:am`},
//	}
//	for _, c := range badCases {
//		t.Run(c.input, func(t *testing.T) {
//			v, err := Parse(c.input)
//			if err == nil {
//				t.Errorf("err is nil, result is %v", v)
//			}
//		})
//	}
//}

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

//func TestParseRange(t *testing.T) {
//	tests := []struct {
//		input string
//		want  Range
//	}{
//		// from A to B
//		{
//			"From 3 feb 2022 to 6 oct 2022",
//			RangeFromTimes(
//				time.Date(2022, 2, 3, 0, 0, 0, 0, now.Location()),
//				time.Date(2022, 10, 6, 23, 59, 59, 0, now.Location()),
//			),
//		},
//		//// A to B
//		//{
//		//	"3 feb 2022 to 6 oct 2022",
//		//	RangeFromTimes(
//		//		time.Date(2022, 2, 3, 0, 0, 0, 0, now.Location()),
//		//		time.Date(2022, 10, 6, 23, 59, 59, 0, now.Location()),
//		//	),
//		//},
//		//// A through B
//		//{
//		//	"3 feb 2022 through 6 oct 2022",
//		//	RangeFromTimes(
//		//		time.Date(2022, 2, 3, 0, 0, 0, 0, now.Location()),
//		//		time.Date(2022, 10, 6, 23, 59, 59, 0, now.Location()),
//		//	),
//		//},
//		// from A until B
//		{
//			"from 3 feb 2022 until 6 oct 2022",
//			RangeFromTimes(
//				time.Date(2022, 2, 3, 0, 0, 0, 0, now.Location()),
//				time.Date(2022, 10, 6, 23, 59, 59, 0, now.Location()),
//			),
//		},
//		//{
//		//	"from tuesday at 5pm -12:00 until thursday 23:52 +14:00",
//		//	RangeFromTimes(
//		//		setLocation(setTime(nextWeekdayFrom(now, time.Tuesday).Time, 12+5, 0, 0, 0), fixedZone(-12)),
//		//		setLocation(setTime(nextWeekdayFrom(now, time.Thursday).Time, 23, 52, 0, 0), fixedZone(14)),
//		//	),
//		//},
//		// yesterday
//		{
//			"Yesterday",
//			RangeFromTimes(
//				today.AddDate(0, 0, -1),
//				today.Add(-time.Second),
//			),
//		},
//		// today
//		{
//			"Today",
//			RangeFromTimes(
//				today,
//				today.AddDate(0, 0, 1).Add(-time.Second),
//			),
//		},
//		// tomorrow
//		{
//			"Tomorrow",
//			RangeFromTimes(
//				today.AddDate(0, 0, 1),
//				today.AddDate(0, 0, 2).Add(-time.Second),
//			),
//		},
//		{
//			"From today until next thursday",
//			RangeFromDays(
//				today,
//				nextWeekdayFrom(today, time.Thursday).Time,
//			),
//		},
//		{
//			"From tomorrow until next tuesday",
//			RangeFromDays(
//				today.AddDate(0, 0, 1),
//				nextWeekdayFrom(today, time.Tuesday).Time,
//			),
//		},
//		// last week
//		{
//			"Last week",
//			RangeFromTimes(
//				time.Date(2022, 9, 18, 0, 0, 0, 0, now.Location()),
//				time.Date(2022, 9, 25, 0, 0, 0, 0, now.Location()).Add(-time.Second),
//			),
//		},
//		// this week
//		{
//			"This week",
//			RangeFromTimes(
//				time.Date(2022, 9, 25, 0, 0, 0, 0, now.Location()),
//				time.Date(2022, 10, 2, 0, 0, 0, 0, now.Location()).Add(-time.Second),
//			),
//		},
//		// next week
//		{
//			"next week",
//			RangeFromTimes(
//				time.Date(2022, 10, 2, 0, 0, 0, 0, now.Location()),
//				time.Date(2022, 10, 9, 0, 0, 0, 0, now.Location()).Add(-time.Second),
//			),
//		},
//		// last month
//		{
//			"Last month",
//			RangeFromTimes(
//				time.Date(2022, 8, 1, 0, 0, 0, 0, now.Location()),
//				time.Date(2022, 9, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second),
//			),
//		},
//		// this month
//		{
//			"This month",
//			RangeFromTimes(
//				time.Date(2022, 9, 1, 0, 0, 0, 0, now.Location()),
//				time.Date(2022, 10, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second),
//			),
//		},
//		// next month
//		{
//			"Next month",
//			RangeFromTimes(
//				time.Date(2022, 10, 1, 0, 0, 0, 0, now.Location()),
//				time.Date(2022, 11, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second),
//			),
//		},
//		// last year
//		{
//			"Last year",
//			RangeFromTimes(
//				time.Date(2021, 1, 1, 0, 0, 0, 0, now.Location()),
//				time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second),
//			),
//		},
//		// this year
//		{
//			"This year",
//			RangeFromTimes(
//				time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
//				time.Date(2023, 1, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second),
//			),
//		},
//		// next year
//		{
//			"Next year",
//			RangeFromTimes(
//				time.Date(2023, 1, 1, 0, 0, 0, 0, now.Location()),
//				time.Date(2024, 1, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second),
//			),
//		},
//		// absolute year
//		{
//			"2025Ad",
//			RangeFromTimes(
//				time.Date(2025, 1, 1, 0, 0, 0, 0, now.Location()),
//				time.Date(2026, 1, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second),
//			),
//		},
//		{
//			"2025ce",
//			RangeFromTimes(
//				time.Date(2025, 1, 1, 0, 0, 0, 0, now.Location()),
//				time.Date(2026, 1, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second),
//			),
//		},
//		// absolute month
//		{
//			"Feb 2025",
//			RangeFromTimes(
//				time.Date(2025, 2, 1, 0, 0, 0, 0, now.Location()),
//				time.Date(2025, 3, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second),
//			),
//		},
//		// absolute day
//		{
//			"3 feb 2025",
//			RangeFromTimes(
//				time.Date(2025, 2, 3, 0, 0, 0, 0, now.Location()),
//				time.Date(2025, 2, 4, 0, 0, 0, 0, now.Location()).Add(-time.Second),
//			),
//		},
//		//// absolute hour
//		//{
//		//	"3 feb 2025 at 5PM",
//		//	RangeFromTimes(
//		//		time.Date(2025, 2, 3, 12+5, 0, 0, 0, now.Location()),
//		//		time.Date(2025, 2, 3, 12+5+1, 0, 0, 0, now.Location()).Add(-time.Second),
//		//	),
//		//},
//		//// absolute minute
//		//{
//		//	"3 feb 2025 at 5:35pm",
//		//	RangeFromTimes(
//		//		time.Date(2025, 2, 3, 12+5, 35, 0, 0, now.Location()),
//		//		time.Date(2025, 2, 3, 12+5, 36, 0, 0, now.Location()).Add(-time.Second),
//		//	),
//		//},
//		//// absolute second
//		//{
//		//	"3 Feb 2025 at 5:35:52pm",
//		//	RangeFromTimes(
//		//		time.Date(2025, 2, 3, 12+5, 35, 52, 0, now.Location()),
//		//		time.Date(2025, 2, 3, 12+5, 35, 53, 0, now.Location()),
//		//	),
//		//},
//		//// 2022 jan 1 0:0:0
//		//{
//		//	"2022 jan 1 0:0:0",
//		//	RangeFromTimes(
//		//		time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
//		//		time.Date(2022, 1, 1, 0, 0, 1, 0, now.Location()),
//		//	),
//		//},
//		//// 2022 jan 1 0:0
//		//{
//		//	"2022 jan 1 0:0",
//		//	RangeFromTimes(
//		//		time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
//		//		time.Date(2022, 1, 1, 0, 1, 0, 0, now.Location()).Add(-time.Second),
//		//	),
//		//},
//		//// 2022 jan 1 12am
//		//{
//		//	"2022 jan 1 12am",
//		//	RangeFromTimes(
//		//		time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
//		//		time.Date(2022, 1, 1, 1, 0, 0, 0, now.Location()).Add(-time.Second),
//		//	),
//		//},
//		//// 2022 jan 1 0am
//		//{
//		//	"2022 jan 1 0am",
//		//	RangeFromTimes(
//		//		time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
//		//		time.Date(2022, 1, 1, 1, 0, 0, 0, now.Location()).Add(-time.Second),
//		//	),
//		//},
//		// 2022 jan 1
//		{
//			"2022 jan 1",
//			RangeFromTimes(
//				time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
//				time.Date(2022, 1, 2, 0, 0, 0, 0, now.Location()).Add(-time.Second),
//			),
//		},
//		// 2022 jan
//		{
//			"2022 jan",
//			RangeFromTimes(
//				time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
//				time.Date(2022, 2, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second),
//			),
//		},
//		// 2022
//		{
//			"2022ce",
//			RangeFromTimes(
//				time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
//				time.Date(2023, 1, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second),
//			),
//		},
//		// 2022
//		{
//			"2022CE",
//			RangeFromTimes(
//				time.Date(2022, 1, 1, 0, 0, 0, 0, now.Location()),
//				time.Date(2023, 1, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second),
//			),
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.input, func(t *testing.T) {
//			got, err := Parse(tt.input)
//			if err != nil {
//				t.Fatal(err)
//			}
//			gotRange, err := verbiageToRange(got, now, future)
//			if err != nil {
//				t.Fatal(err)
//			}
//			if !gotRange.Equal(tt.want) {
//				t.Errorf("got = %v, want %v", gotRange, tt.want)
//			}
//		})
//	}
//}

//func Test_nextWeek(t *testing.T) {
//	type args struct {
//		ref time.Time
//	}
//	tests := []struct {
//		name string
//		args args
//		want Range
//	}{
//		{
//			name: "2022-9-30",
//			args: args{
//				ref: now,
//			},
//			want: Range{
//				Time:     time.Date(2022, 10, 2, 0, 0, 0, 0, now.Location()),
//				Duration: 7*24*time.Hour - time.Second,
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := nextWeek(tt.args.ref); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("nextWeek() = \n%v\nwant\n%v", got, tt.want)
//			}
//		})
//	}
//}

//func TestRange_String(t *testing.T) {
//	type fields struct {
//		Time     time.Time
//		Duration time.Duration
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		want   string
//	}{
//		{
//			name: "zeros",
//			fields: fields{
//				Time:     time.Time{},
//				Duration: 0,
//			},
//			want: "{time: 0001-01-01 00:00:00 +0000 UTC, duration: 0s}",
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			r := Range{
//				Time:     tt.fields.Time,
//				Duration: tt.fields.Duration,
//			}
//			if got := r.String(); got != tt.want {
//				t.Errorf("String() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

//func Test_prevWeekdayFrom(t *testing.T) {
//	tests := []struct {
//		day  time.Weekday
//		want time.Time
//	}{
//		{time.Thursday, time.Date(2022, 9, 22, 0, 0, 0, 0, now.Location())},
//		{time.Friday, time.Date(2022, 9, 23, 0, 0, 0, 0, now.Location())},
//		{time.Saturday, time.Date(2022, 9, 24, 0, 0, 0, 0, now.Location())},
//		{time.Sunday, time.Date(2022, 9, 25, 0, 0, 0, 0, now.Location())},
//		{time.Monday, time.Date(2022, 9, 26, 0, 0, 0, 0, now.Location())},
//		{time.Tuesday, time.Date(2022, 9, 27, 0, 0, 0, 0, now.Location())},
//		{time.Wednesday, time.Date(2022, 9, 28, 0, 0, 0, 0, now.Location())},
//	}
//	for _, tt := range tests {
//		t.Run(fmt.Sprintf("%s", tt.day), func(t *testing.T) {
//			if got := lastWeekdayFrom(now, tt.day); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("lastWeekdayFrom() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

//func Test_nextWeekdayFrom(t *testing.T) {
//	tests := []struct {
//		day  time.Weekday
//		want time.Time
//	}{
//		{time.Friday, time.Date(2022, 9, 30, 0, 0, 0, 0, now.Location())},
//		{time.Saturday, time.Date(2022, 10, 1, 0, 0, 0, 0, now.Location())},
//		{time.Sunday, time.Date(2022, 10, 2, 0, 0, 0, 0, now.Location())},
//		{time.Monday, time.Date(2022, 10, 3, 0, 0, 0, 0, now.Location())},
//		{time.Tuesday, time.Date(2022, 10, 4, 0, 0, 0, 0, now.Location())},
//		{time.Wednesday, time.Date(2022, 10, 5, 0, 0, 0, 0, now.Location())},
//		{time.Thursday, time.Date(2022, 10, 6, 0, 0, 0, 0, now.Location())},
//	}
//	for _, tt := range tests {
//		t.Run(fmt.Sprintf("%s", tt.day), func(t *testing.T) {
//			if got := nextWeekdayFrom(now, tt.day); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("lastWeekdayFrom() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

//func TestReplaceTimesByFunc(t *testing.T) {
//	type args struct {
//		s       string
//		ref     time.Time
//		f       func(string, time.Time) string
//		options []func(o *opts)
//	}
//	tests := []struct {
//		name    string
//		args    args
//		want    string
//		wantErr bool
//	}{
//		{
//			name: "empty",
//			args: args{
//				s:       "",
//				ref:     time.Time{},
//				f:       nil,
//				options: nil,
//			},
//			want:    "",
//			wantErr: false,
//		},
//		{
//			name: "issue 14 example without prefix",
//			args: args{
//				s:   "on Tuesday at 11am UTC",
//				ref: time.Date(2022, time.Month(10), 24, 0, 0, 0, 0, time.UTC),
//				f: func(src string, t time.Time) string {
//					return t.String()
//				},
//				options: nil,
//			},
//			want:    "2022-10-25 11:00:00 +0000 UTC",
//			wantErr: false,
//		},
//		{
//			name: "issue 14 example",
//			args: args{
//				s:   "Let's meet on Tuesday at 11am UTC if that works for you",
//				ref: time.Date(2022, time.Month(10), 24, 0, 0, 0, 0, time.UTC),
//				f: func(src string, t time.Time) string {
//					return t.String()
//				},
//				options: nil,
//			},
//			want:    "Let's meet 2022-10-25 11:00:00 +0000 UTC if that works for you",
//			wantErr: false,
//		},
//		{
//			name: "two dates",
//			args: args{
//				s:   "Let's meet on Tuesday at 11am UTC or monday if you like",
//				ref: time.Date(2022, time.Month(10), 24, 0, 0, 0, 0, time.UTC),
//				f: func(src string, t time.Time) string {
//					return t.String()
//				},
//				options: nil,
//			},
//			want:    "Let's meet 2022-10-25 11:00:00 +0000 UTC or 2022-10-31 00:00:00 +0000 UTC if you like",
//			wantErr: false,
//		},
//		{
//			name: "one range shows up as two dates",
//			args: args{
//				s:   "from 3 feb 2022 to 6 oct 2022",
//				ref: now,
//				f: func(src string, t time.Time) string {
//					return "DATE"
//				},
//				options: nil,
//			},
//			want:    "from DATE to DATE",
//			wantErr: false,
//		},
//		{
//			name: "market doesn't get mistaken for month of mar followed by ket",
//			args: args{
//				s:       "market",
//				ref:     time.Time{},
//				f:       func(src string, t time.Time) string { return t.String() },
//				options: nil,
//			},
//			want:    "market",
//			wantErr: false,
//		},
//		{
//			name: "month doesn't get mistaken for mon-day followed by th",
//			args: args{
//				s:   "month",
//				ref: time.Time{},
//				f: func(src string, t time.Time) string {
//					return t.String()
//				},
//				options: nil,
//			},
//			want:    "month",
//			wantErr: false,
//		},
//		{
//			name: "1000 widgets",
//			args: args{
//				s:   "1000 widgets",
//				ref: time.Time{},
//				f: func(src string, t time.Time) string {
//					return t.String()
//				},
//				options: nil,
//			},
//			want:    "1000 widgets",
//			wantErr: false,
//		},
//		{
//			name: "1000AD",
//			args: args{
//				s:   "1000AD",
//				ref: time.Time{},
//				f: func(src string, t time.Time) string {
//					return t.String()
//				},
//				options: nil,
//			},
//			want:    time.Date(1000, time.Month(1), 1, 0, 0, 0, 0, time.UTC).String(),
//			wantErr: false,
//		},
//		{
//			name: "issue 38: want december 40 to only get december",
//			args: args{
//				s:   "december 40",
//				ref: now,
//				f: func(src string, t time.Time) string {
//					return t.Format("Jan 2")
//				},
//				options: nil,
//			},
//			want:    "Dec 1 40",
//			wantErr: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, err := ReplaceTimesByFunc(tt.args.s, tt.args.ref, tt.args.f, tt.args.options...)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("ReplaceTimesByFunc() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if got != tt.want {
//				t.Errorf("ReplaceTimesByFunc() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

//func TestReplaceRangesByFunc(t *testing.T) {
//	type args struct {
//		s       string
//		ref     time.Time
//		f       func(Range) string
//		options []func(o *opts)
//	}
//	tests := []struct {
//		name    string
//		args    args
//		want    string
//		wantErr bool
//	}{
//		{
//			name: "empty",
//			args: args{
//				s:       "",
//				ref:     now,
//				f:       nil,
//				options: nil,
//			},
//			want:    "",
//			wantErr: false,
//		},
//		//{
//		//	name: "two dates without a connecting preposition get left alone",
//		//	args: args{
//		//		s:       "Let's meet on Tuesday at 11am UTC or monday if you like",
//		//		ref:     time.Date(2022, time.Month(10), 24, 0, 0, 0, 0, time.UTC),
//		//		f:       nil,
//		//		options: nil,
//		//	},
//		//	want:    "Let's meet on Tuesday at 11am UTC or monday if you like",
//		//	wantErr: false,
//		//},
//		//{
//		//	name: "one range shows up as a range",
//		//	args: args{
//		//		s:   "from 3 feb 2022 to 6 oct 2022",
//		//		ref: now,
//		//		f: func(src string, r Range) string {
//		//			ymd := "2006/01/02"
//		//			return fmt.Sprintf("%s - %s", r.Start().Format(ymd), r.End().Format(ymd))
//		//		},
//		//		options: nil,
//		//	},
//		//	want:    "2022/02/03 - 2022/10/06",
//		//	wantErr: false,
//		//},
//		//{
//		//	name: "stuff then range then stuff then range then stuff",
//		//	args: args{
//		//		s:   "twas brillig from 3 feb 2022 to 6 oct 2022 and the slithy toves from april until may did gyre and gimble in the wabe",
//		//		ref: now,
//		//		f: func(src string, r Range) string {
//		//			return "RANGE"
//		//		},
//		//		options: nil,
//		//	},
//		//	want:    "twas brillig RANGE and the slithy toves RANGE did gyre and gimble in the wabe",
//		//	wantErr: false,
//		//},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, err := ReplaceRangesByFunc(tt.args.s, tt.args.ref, tt.args.f, tt.args.options...)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("ReplaceRangesByFunc() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if got != tt.want {
//				t.Errorf("ReplaceRangesByFunc() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

//func TestReplaceDateRangesByFunc(t *testing.T) {
//	type args struct {
//		s       string
//		ref     time.Time
//		f       func(r Range) string
//		options []func(o *opts)
//	}
//	tests := []struct {
//		name    string
//		args    args
//		want    string
//		wantErr bool
//	}{
//		{
//			name: "empty",
//			args: args{
//				s:       "",
//				ref:     time.Time{},
//				f:       nil,
//				options: nil,
//			},
//			want:    "",
//			wantErr: false,
//		},
//		{
//			name: "this week",
//			args: args{
//				s:   "this week",
//				ref: now,
//				f: func(r Range) string {
//					return fmt.Sprintf("from %s to %s", r.Start(), r.End())
//				},
//				options: nil,
//			},
//			want:    "from 2022-09-25 00:00:00 +0000 UTC to 2022-10-01 23:59:59 +0000 UTC",
//			wantErr: false,
//		},
//		{
//			name: "now now",
//			args: args{
//				s:   "now now",
//				ref: now,
//				f: func(r Range) string {
//					return "n"
//				},
//				options: nil,
//			},
//			want: "n n",
//		},
//		{
//			name: "several ranges in a row",
//			args: args{
//				s:   "last week this month next month next week",
//				ref: now,
//				f: func(r Range) string {
//					return fmt.Sprintf("%s to %s", r.Start(), r.End())
//				},
//				options: nil,
//			},
//			want:    "2022-09-18 00:00:00 +0000 UTC to 2022-09-24 23:59:59 +0000 UTC 2022-09-01 00:00:00 +0000 UTC to 2022-09-30 23:59:59 +0000 UTC 2022-10-01 00:00:00 +0000 UTC to 2022-10-31 23:59:59 +0000 UTC 2022-10-02 00:00:00 +0000 UTC to 2022-10-08 23:59:59 +0000 UTC",
//			wantErr: false,
//		},
//		{
//			name: "days and smaller get skipped",
//			args: args{
//				s:   "last week today next week now",
//				ref: now,
//				f: func(r Range) string {
//					return fmt.Sprintf("%s to %s", r.Start(), r.End())
//				},
//				options: nil,
//			},
//			want:    "2022-09-18 00:00:00 +0000 UTC to 2022-09-24 23:59:59 +0000 UTC today 2022-10-02 00:00:00 +0000 UTC to 2022-10-08 23:59:59 +0000 UTC now",
//			wantErr: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, err := ReplaceRangesByFunc(tt.args.s, tt.args.ref, tt.args.f, tt.args.options...)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("ReplaceDateRangesByFunc() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if got != tt.want {
//				t.Errorf("ReplaceDateRangesByFunc() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
