// Package anytime parses dates, times and ranges without requiring a format.
package anytime

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	gp "github.com/ijt/goparsify"
)

// Range is a time range.
type Range struct {
	time.Time
	Duration time.Duration
}

func (r Range) Start() time.Time {
	return r.Time
}

func (r Range) End() time.Time {
	return r.Time.Add(r.Duration)
}

// RangeFromTimes returns a range given the start and end times.
func RangeFromTimes(start, end time.Time) Range {
	return Range{start, end.Sub(start)}
}

// String returns a string with the time and duration of the range.
func (r Range) String() string {
	return fmt.Sprintf("{time: %v, duration: %v}", r.Time, r.Duration)
}

type direction int

const (
	future = iota
	past
)

type opts struct {
	defaultDirection direction
}

// DefaultToFuture sets the option to default to the future in case of
// ambiguous dates.
func DefaultToFuture(o *opts) {
	o.defaultDirection = future
}

// DefaultToPast sets the option to default to the past in case of
// ambiguous dates.
func DefaultToPast(o *opts) {
	o.defaultDirection = past
}

// ReplaceDateRangesByFunc replaces all ranges with duration over one day in the
// given string s by running func f on each found range and the source string
// that defines it.
func ReplaceDateRangesByFunc(s string, ref time.Time, f func(source string, r Range) string, options ...func(o *opts)) (string, error) {
	rangeParser := RangeParser(ref, options...).Map(func(n *gp.Result) {
		r := n.Result.(Range)
		if r.Duration <= 24*time.Hour {
			// This time range is no more than one day, so it's not really a date
			// range.
			n.Result = n.Token
			return
		}
		n.Result = f(n.Token, r)
	})
	wordParser := gp.Regex(`\S+`).Map(func(n *gp.Result) {
		n.Result = n.Token
	})
	p := gp.Many(gp.AnyWithName("range or word", rangeParser, wordParser)).Map(func(n *gp.Result) {
		var results []string
		for _, c := range n.Child {
			r := c.Result.(string)
			results = append(results, r)
		}
		n.Result = strings.Join(results, " ")
	})
	result, _, err := gp.Run(p, s)
	if err != nil {
		return "", fmt.Errorf("parsing date ranges: %w", err)
	}
	s2 := result.(string)
	return s2, nil
}

// ReplaceTimesByFunc replaces all dates, times and datetimes found in the given
// string s by calling the func f. The ref and options arguments are the same as
// in Parse.
func ReplaceTimesByFunc(s string, ref time.Time, f func(time.Time) string, options ...func(o *opts)) (string, error) {
	tyme := Parser(ref, options...).Map(func(n *gp.Result) {
		r := n.Result.(Range)
		n.Result = f(r.Time)
	})
	word := gp.Regex(`\S+`).Map(func(n *gp.Result) {
		n.Result = n.Token
	})
	p := gp.Many(gp.AnyWithName("time or word", tyme, word)).Map(func(n *gp.Result) {
		var results []string
		for _, c := range n.Child {
			r := c.Result.(string)
			results = append(results, r)
		}
		n.Result = strings.Join(results, " ")
	})
	result, _, err := gp.Run(p, s)
	if err != nil {
		return "", fmt.Errorf("parsing: %w", err)
	}
	s2 := result.(string)
	return s2, nil
}

// ReplaceRangesByFunc replaces all ranges found in the given string s by
// calling the func f. The ref and options arguments are the same as in
// ParseRange. Ranges like "today" that can also be parsed as non-ranges
// are skipped over.
func ReplaceRangesByFunc(s string, ref time.Time, f func(Range) string, options ...func(o *opts)) (string, error) {
	rangeParser := RangeParser(ref, options...).Map(func(n *gp.Result) {
		r := n.Result.(Range)
		_, err := Parse(n.Token, ref, options...)
		if err == nil {
			// This "range" is also parseable as a regular date, so leave it
			// alone.
			n.Result = n.Token
			return
		}
		n.Result = f(r)
	})
	wordParser := gp.Regex(`\S+`).Map(func(n *gp.Result) {
		n.Result = n.Token
	})
	p := gp.Many(gp.AnyWithName("range or word", rangeParser, wordParser)).Map(func(n *gp.Result) {
		var results []string
		for _, c := range n.Child {
			r := c.Result.(string)
			results = append(results, r)
		}
		n.Result = strings.Join(results, " ")
	})
	result, _, err := gp.Run(p, s)
	if err != nil {
		return "", fmt.Errorf("parsing: %w", err)
	}
	s2 := result.(string)
	return s2, nil
}

// PartitionTimes returns a slice whose pieces are non-time pieces of s and time
// pieces of s, in order.
func PartitionTimes(s string, ref time.Time, options ...func(o *opts)) []any {
	var parts []any
	PartitionTimesByFuncs(s, ref,
		func(nonTimeChunk string) {
			parts = append(parts, nonTimeChunk)
		},
		func(timeChunk string, t time.Time) {
			parts = append(parts, t)
		},
		options...)
	return parts
}

// PartitionTimesByFuncs partitions the given string s into time and non-time
// parts, calling ntf on the non-time parts and tf on the time parts in
// succession.
func PartitionTimesByFuncs(s string, ref time.Time, ntf func(nonTimeChunk string), tf func(timeChunk string, t time.Time), options ...func(o *opts)) {
	tyme := Parser(ref, options...).Map(func(n *gp.Result) {
		r := n.Result.(Range)
		n.Result = r.Time
	})
	word := gp.Regex(`\S+`).Map(func(n *gp.Result) {
		n.Result = n.Token
	})
	p := gp.Many(gp.AnyWithName("time or word", tyme, word)).Map(func(n *gp.Result) {
		var strParts []string
		for _, c := range n.Child {
			switch v := c.Result.(type) {
			case string:
				strParts = append(strParts, v)
			case time.Time:
				if len(strParts) > 0 {
					ntf(strings.Join(strParts, " "))
					strParts = nil
				}
				tf(c.Token, v)
			default:
				panic(fmt.Sprintf("unexpected type %T in child list", v))
			}
		}
		if len(strParts) > 0 {
			ntf(strings.Join(strParts, " "))
			strParts = nil
		}
	})
	_, _, _ = gp.Run(p, s)
}

// Parse parses a string assumed to contain a date, a time, or a datetime
// in one of various formats.
func Parse(s string, ref time.Time, opts ...func(o *opts)) (time.Time, error) {
	p := Parser(ref, opts...)
	result, _, err := gp.Run(p, s, gp.UnicodeWhitespace)
	if err != nil {
		return time.Time{}, fmt.Errorf("running parser: %w", err)
	}
	t := result.(Range)
	return t.Time, nil
}

// Parser returns a parser of dates with a given reference time called ref.
// The result is a Range so that we have a time scale to work with, mainly
// for parsing implicit ranges within RangeParser().
func Parser(ref time.Time, options ...func(o *opts)) gp.Parser {
	var o opts
	for _, optFunc := range options {
		optFunc(&o)
	}

	sep := gp.Maybe(gp.AnyWithName("separator", "/", "-", ","))
	sladash := gp.AnyWithName("slash or dash", "/", "-")
	comma := gp.Maybe(",")

	now := gp.Bind(I("now"), Range{ref, time.Nanosecond})

	prevMo := gp.Seq(I("last"), I("month")).Map(func(n *gp.Result) {
		n.Result = truncateMonth(ref.AddDate(0, -1, 0))
	})

	thisMo := gp.Seq(I("this"), I("month")).Map(func(n *gp.Result) {
		n.Result = truncateMonth(ref)
	})

	nextMo := gp.Seq(I("next"), I("month")).Map(func(n *gp.Result) {
		n.Result = truncateMonth(ref.AddDate(0, 1, 0))
	})

	lastWeekParser := gp.Seq(I("last"), I("week")).Map(func(n *gp.Result) {
		n.Result = lastWeek(ref)
	})

	thisWeekParser := gp.Seq(I("this"), I("week")).Map(func(n *gp.Result) {
		n.Result = thisWeek(ref)
	})

	nextWeekParser := gp.Seq(I("next"), I("week")).Map(func(n *gp.Result) {
		n.Result = nextWeek(ref)
	})

	one := gp.Bind(I("one"), 1)
	a := gp.Bind(I("a"), 1)
	an := gp.Bind(I("an"), 1)
	two := gp.Bind(I("two"), 2)
	three := gp.Bind(I("three"), 3)
	four := gp.Bind(I("four"), 4)
	five := gp.Bind(I("five"), 5)
	six := gp.Bind(I("six"), 6)
	seven := gp.Bind(I("seven"), 7)
	eight := gp.Bind(I("eight"), 8)
	nine := gp.Bind(I("nine"), 9)
	ten := gp.Bind(I("ten"), 10)
	eleven := gp.Bind(I("eleven"), 11)
	twelve := gp.Bind(I("twelve"), 12)
	numeral := gp.Regex(`\d+`).Map(func(n *gp.Result) {
		num, err := strconv.Atoi(n.Token)
		if err != nil {
			panic(fmt.Sprintf("parsing numeral: %v", err))
		}
		n.Result = num
	})

	number := gp.AnyWithName("number", one, an, a, two, three, four, five, six, seven, eight, nine, ten, eleven, twelve, numeral).Map(func(n *gp.Result) {
		pass()
	})

	months := gp.Regex(`(?i)months?`)

	monthsAgo := gp.Seq(number, months, I("ago")).Map(func(n *gp.Result) {
		num := n.Child[0].Result.(int)
		s := ref.AddDate(0, -num, 0)
		dur := s.AddDate(0, 1, 0).Sub(s)
		n.Result = Range{s, dur}
	})

	monthsFromNow := gp.Seq(number, months, gp.Any(gp.Seq(I("from"), I("now")), I("hence"))).Map(func(n *gp.Result) {
		num := n.Child[0].Result.(int)
		s := ref.AddDate(0, num, 0)
		dur := s.AddDate(0, 1, 0).Sub(s)
		n.Result = Range{s, dur}
	})

	shortWeekdays := []string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"}
	var shortWeekdayParsers []gp.Parserish
	for _, swd := range shortWeekdays {
		shortWeekdayName := gp.Regex(`(?i)\b` + swd + `\b`)
		shortWeekdayParsers = append(shortWeekdayParsers, shortWeekdayName)
	}
	shortWeekday := gp.AnyWithName("short weekday", shortWeekdayParsers...)

	longWeekday := gp.AnyWithName("long weekday", I("monday"), I("tuesday"), I("wednesday"), I("thursday"), I("friday"), I("saturday"), I("sunday"))

	weekday := gp.AnyWithName("weekday", longWeekday, shortWeekday).Map(func(n *gp.Result) {
		m := map[string]time.Weekday{
			"sun": time.Sunday,
			"mon": time.Monday,
			"tue": time.Tuesday,
			"wed": time.Wednesday,
			"thu": time.Thursday,
			"fri": time.Friday,
			"sat": time.Saturday,
		}
		shortName := strings.ToLower(n.Token[:3])
		day := m[shortName]
		n.Result = day
	})

	longMonth := gp.AnyWithName("long month",
		I("january"), I("february"), I("march"), I("april"),
		/* may is already short */ I("june"), I("july"), I("august"), I("september"),
		I("october"), I("november"), I("december")).Map(func(n *gp.Result) {
		t, err := time.Parse("January", n.Token)
		if err != nil {
			panic(fmt.Sprintf("identifying month (long): %v", err))
		}
		n.Result = t.Month()
	})

	shortMonthNames := []string{
		"jan", "feb", "mar", "apr", "may", "jun", "jul", "aug", "sep", "oct", "nov", "dec",
	}
	var shortMonthParsers []gp.Parserish
	for _, mo := range shortMonthNames {
		shortMonth := gp.Regex(`(?i)\b` + mo + `\b`)
		shortMonthParsers = append(shortMonthParsers, shortMonth)
	}
	shortMonth := gp.AnyWithName("month", shortMonthParsers...).Map(func(n *gp.Result) {
		t, err := time.Parse("Jan", n.Token)
		if err != nil {
			panic(fmt.Sprintf("identifying month: %v", err))
		}
		n.Result = t.Month()
	})

	shortMonthMaybeDot := gp.Seq(shortMonth, gp.Maybe(".")).Map(func(n *gp.Result) {
		n.Result = n.Child[0].Result
	})

	monthNum := gp.Regex(`[01]?\d\b`).Map(func(n *gp.Result) {
		m, err := strconv.Atoi(n.Token)
		if err != nil {
			panic(fmt.Sprintf("parsing month number: %v", err))
		}
		n.Result = time.Month(m)
	})

	month := gp.AnyWithName("month", longMonth, shortMonthMaybeDot)

	lastSpecificMonth := gp.Seq(I("last"), month).Map(func(n *gp.Result) {
		m := n.Child[1].Result.(time.Month)
		n.Result = prevMonth(ref, m)
	})

	nextSpecificMonth := gp.Seq(I("next"), month).Map(func(n *gp.Result) {
		m := n.Child[1].Result.(time.Month)
		n.Result = nextMonth(ref, m)
	})

	dayOfMonthNum := gp.Regex(`[0-3]?\d`).Map(func(n *gp.Result) {
		d, err := strconv.Atoi(n.Token)
		if err != nil {
			panic(fmt.Sprintf("parsing day of month: %v", err))
		}
		n.Result = d
	})

	dayOfMonthEnding := gp.Regex(`(st|nd|rd|th)`).Map(func(n *gp.Result) {
		pass()
	})

	dayOfMonth := gp.Seq(dayOfMonthNum, gp.Maybe(dayOfMonthEnding)).Map(func(n *gp.Result) {
		n.Result = n.Child[0].Result
	})

	hour12 := gp.Regex(`[0-1]?\d`).Map(func(n *gp.Result) {
		h, err := strconv.Atoi(n.Token)
		if err != nil {
			panic(fmt.Sprintf("parsing hour (12h clock): %v", err))
		}
		n.Result = h
	})

	hour24 := gp.Regex(`[0-2]?\d\b`).Map(func(n *gp.Result) {
		h, err := strconv.Atoi(n.Token)
		if err != nil {
			panic(fmt.Sprintf("parsing hour (24h clock): %v", err))
		}
		n.Result = h
	})

	minute := gp.Regex(`[0-5]?\d`).Map(func(n *gp.Result) {
		m, err := strconv.Atoi(n.Token)
		if err != nil {
			panic(fmt.Sprintf("parsing minute: %v", err))
		}
		n.Result = m
	})

	// Second can go up to 60 because of leap seconds, for example
	// 1990-12-31T15:59:60-08:00.
	second := gp.Regex(`[0-6]?\d`).Map(func(n *gp.Result) {
		s, err := strconv.Atoi(n.Token)
		if err != nil {
			panic(fmt.Sprintf("parsing second: %v", err))
		}
		n.Result = s
	})

	amPM := gp.AnyWithName("AM or PM", I("am"), I("pm"))

	colonSecond := gp.Seq(":", second).Map(func(n *gp.Result) {
		n.Result = n.Child[1].Result
	})

	colonMinute := gp.Seq(":", minute).Map(func(n *gp.Result) {
		n.Result = n.Child[1].Result
	})

	colonMinuteColonSecond := gp.Seq(colonMinute, gp.Maybe(colonSecond)).Map(func(n *gp.Result) {
		m := n.Child[0].Result.(int)
		c1 := n.Child[1].Result
		dur := time.Minute
		s := 0
		if c1 != nil {
			s = c1.(int)
			dur = time.Second
		}
		n.Result = Range{
			time.Date(1, 1, 1, 0, m, s, 0, ref.Location()),
			dur,
		}
	})

	noon := gp.Regex(`(?i)\bnoon\b`).Map(func(n *gp.Result) {
		n.Result = Range{
			time.Date(ref.Year(), ref.Month(), ref.Day(), 12, 0, 0, 0, ref.Location()),
			time.Hour,
		}
	})

	hour12MinuteSecond := gp.Seq(hour12, gp.Maybe(colonMinuteColonSecond), amPM).Map(func(n *gp.Result) {
		h := n.Child[0].Result.(int)
		c1 := n.Child[1].Result
		dur := time.Hour
		m := 0
		s := 0
		ap := strings.ToLower(n.Child[2].Token)
		t, err := time.Parse("3pm", fmt.Sprintf("%d%s", h, ap))
		if err != nil {
			panic(err)
		}
		if c1 != nil {
			ms := n.Child[1].Result.(Range)
			m = ms.Minute()
			s = ms.Second()
			t, err = time.Parse("3:4:5pm", fmt.Sprintf("%d:%d:%d%s", h, m, s, ap))
			if err != nil {
				panic(err)
			}
			dur = ms.Duration
		}
		n.Result = Range{
			time.Date(ref.Year(), ref.Month(), ref.Day(), t.Hour(), t.Minute(), t.Second(), 0, ref.Location()),
			dur,
		}
	})

	hour24MinuteSecond := gp.Seq(hour24, colonMinute, gp.Maybe(colonSecond)).Map(func(n *gp.Result) {
		h := n.Child[0].Result.(int)
		m := n.Child[1].Result.(int)
		dur := time.Minute
		s := 0
		c2 := n.Child[2].Result
		if c2 != nil {
			s = c2.(int)
			dur = time.Second
		}
		n.Result = Range{
			time.Date(ref.Year(), ref.Month(), ref.Day(), h, m, s, 0, ref.Location()),
			dur,
		}
	})

	hourMinuteSecond := gp.AnyWithName("h:m:s", hour12MinuteSecond, hour24MinuteSecond)

	zoneHour := gp.Regex(`[-+][01]?\d`).Map(func(n *gp.Result) {
		h, err := strconv.Atoi(n.Token)
		if err != nil {
			panic(fmt.Sprintf("parsing time zone hour: %v", err))
		}
		n.Result = h
	})

	maybeColonMinute := gp.Seq(gp.Maybe(":"), minute).Map(func(n *gp.Result) {
		n.Result = n.Child[1].Result
	})

	zoneOffset := gp.Seq(zoneHour, gp.Maybe(maybeColonMinute)).Map(func(n *gp.Result) {
		h := n.Child[0].Result.(int)
		c1 := n.Child[1].Result
		m := 0
		if c1 != nil {
			m = c1.(int)
		}
		n.Result = fixedZoneHM(h, m)
	})

	zoneUTC := gp.Seq(I("utc"), gp.Maybe(zoneOffset)).Map(func(n *gp.Result) {
		c1 := n.Child[1].Result
		z := time.UTC
		if c1 != nil {
			z = c1.(*time.Location)
		}
		n.Result = z
	})

	zoneZ := gp.Bind(I("z"), time.UTC)

	zone := gp.AnyWithName("time zone", zoneUTC, zoneOffset, zoneZ).Map(func(n *gp.Result) {
		pass()
	})

	year := gp.Regex(`[12]\d{3}\b`).Map(func(n *gp.Result) {
		y, err := strconv.Atoi(n.Token)
		if err != nil {
			panic(fmt.Sprintf("parsing year: %v", err))
		}
		n.Result = y
	})

	ansiC := gp.Seq(weekday, month, dayOfMonth, hourMinuteSecond, year).Map(func(n *gp.Result) {
		m := n.Child[1].Result.(time.Month)
		d := n.Child[2].Result.(int)
		t := n.Child[3].Result.(Range)
		y := n.Child[4].Result.(int)
		n.Result = Range{
			time.Date(y, m, d, t.Hour(), t.Minute(), t.Second(), 0, ref.Location()),
			time.Second,
		}
	})

	rubyDate := gp.Seq(weekday, month, dayOfMonth, hourMinuteSecond, zone, year).Map(func(n *gp.Result) {
		m := n.Child[1].Result.(time.Month)
		d := n.Child[2].Result.(int)
		t := n.Child[3].Result.(Range)
		z := n.Child[4].Result.(*time.Location)
		y := n.Child[5].Result.(int)
		n.Result = Range{
			time.Date(y, m, d, t.Hour(), t.Minute(), t.Second(), 0, z),
			time.Second,
		}
	})

	rfc1123Z := gp.Seq(weekday, comma, dayOfMonth, month, year, hourMinuteSecond, gp.Cut(), zone).Map(func(n *gp.Result) {
		d := n.Child[2].Result.(int)
		m := n.Child[3].Result.(time.Month)
		y := n.Child[4].Result.(int)
		t := n.Child[5].Result.(Range)
		z := n.Child[7].Result.(*time.Location)
		n.Result = Range{
			time.Date(y, m, d, t.Hour(), t.Minute(), t.Second(), 0, z),
			time.Second,
		}
	})

	rfc3339 := gp.Regex(`(?i)[12]\d{3}-[01]\d-[0-3]\dt[0-2]\d:[0-5]\d:[0-6]\d(z|[-+][01]\d:\d\d)`).Map(func(n *gp.Result) {
		t, err := time.Parse(time.RFC3339, strings.ToUpper(n.Token))
		if err != nil {
			panic(fmt.Sprintf("parsing time in RFC3339 format: %v", err))
		}
		n.Result = Range{t, time.Second}
	})

	dmyDate := gp.Seq(dayOfMonth, gp.Maybe(gp.Any(I("of"), sep)), month, sep, year).Map(func(n *gp.Result) {
		d := n.Child[0].Result.(int)
		m := n.Child[2].Result.(time.Month)
		y := n.Child[4].Result.(int)
		n.Result = Range{
			time.Date(y, m, d, 0, 0, 0, 0, ref.Location()),
			24 * time.Hour,
		}
	})

	// "my" here stands for "month, year"
	myDate := gp.Seq(month, gp.Maybe(","), year).Map(func(n *gp.Result) {
		m := n.Child[0].Result.(time.Month)
		y := n.Child[2].Result.(int)
		d0 := time.Date(y, m, 1, 0, 0, 0, 0, ref.Location())
		d1 := d0.AddDate(0, 1, 0)
		dur := d1.Sub(d0)
		n.Result = Range{d0, dur}
	})

	ymDate := gp.Seq(year, month).Map(func(n *gp.Result) {
		y := n.Child[0].Result.(int)
		m := n.Child[1].Result.(time.Month)
		d0 := time.Date(y, m, 1, 0, 0, 0, 0, ref.Location())
		d1 := d0.AddDate(0, 1, 0)
		dur := d1.Sub(d0)
		n.Result = Range{d0, dur}
	})

	mdyDate := gp.Seq(month, sep, dayOfMonth, sep, year).Map(func(n *gp.Result) {
		m := n.Child[0].Result.(time.Month)
		d := n.Child[2].Result.(int)
		y := n.Child[4].Result.(int)
		n.Result = Range{
			time.Date(y, m, d, 0, 0, 0, 0, ref.Location()),
			24 * time.Hour,
		}
	})

	ymdDate := gp.Seq(year, sep, month, sep, dayOfMonth).Map(func(n *gp.Result) {
		y := n.Child[0].Result.(int)
		m := n.Child[2].Result.(time.Month)
		d := n.Child[4].Result.(int)
		n.Result = Range{
			time.Date(y, m, d, 0, 0, 0, 0, ref.Location()),
			24 * time.Hour,
		}
	})

	ymdNumDate := gp.Seq(year, sladash, monthNum, sladash, dayOfMonthNum).Map(func(n *gp.Result) {
		y := n.Child[0].Result.(int)
		m := n.Child[2].Result.(time.Month)
		d := n.Child[4].Result.(int)
		n.Result = Range{
			time.Date(y, m, d, 0, 0, 0, 0, ref.Location()),
			24 * time.Hour,
		}
	})

	dmyNumDate := gp.Seq(dayOfMonthNum, sep, monthNum, sep, year).Map(func(n *gp.Result) {
		d := n.Child[0].Result.(int)
		m := n.Child[2].Result.(time.Month)
		y := n.Child[4].Result.(int)
		n.Result = Range{
			time.Date(y, m, d, 0, 0, 0, 0, ref.Location()),
			24 * time.Hour,
		}
	})

	yearEra := gp.Regex(`(?i)[12]\d{3}\s*(ad|ce)\b`).Map(func(n *gp.Result) {
		s := n.Token
		s = strings.TrimSuffix(s, "ad")
		s = strings.TrimSuffix(s, "ce")
		s = strings.TrimSpace(s)
		y, _ := strconv.Atoi(s)
		d0 := time.Date(y, 1, 1, 0, 0, 0, 0, ref.Location())
		dur := d0.AddDate(1, 0, 0).Sub(d0)
		n.Result = Range{
			d0,
			dur,
		}
	})

	lastWeekday := gp.Seq(I("last"), weekday).Map(func(n *gp.Result) {
		day := n.Child[1].Result.(time.Weekday)
		d := prevWeekdayFrom(ref, day)
		n.Result = Range{d, 24 * time.Hour}
	})

	nextWeekday := gp.Seq(I("next"), weekday).Map(func(n *gp.Result) {
		day := n.Child[1].Result.(time.Weekday)
		d := nextWeekdayFrom(ref, day)
		n.Result = Range{d, 24 * time.Hour}
	})

	lastSpecificMonthDay := gp.Seq(I("last"), month, dayOfMonth).Map(func(n *gp.Result) {
		m := n.Child[1].Result.(time.Month)
		d := n.Child[2].Result.(int)
		pm := prevMonth(ref, m)
		t := time.Date(pm.Year(), pm.Month(), d, 0, 0, 0, 0, ref.Location())
		n.Result = Range{t, 24 * time.Hour}
	})

	nextSpecificMonthDay := gp.Seq(I("next"), month, dayOfMonth).Map(func(n *gp.Result) {
		m := n.Child[1].Result.(time.Month)
		d := n.Child[2].Result.(int)
		nm := nextMonth(ref, m)
		t := time.Date(nm.Year(), nm.Month(), d, 0, 0, 0, 0, ref.Location())
		n.Result = Range{t, 24 * time.Hour}
	})

	lastYear := gp.Seq(I("last"), I("year")).Map(func(n *gp.Result) {
		n.Result = truncateYear(ref.AddDate(-1, 0, 0))
	})

	thisYear := gp.Seq(I("this"), I("year")).Map(func(n *gp.Result) {
		n.Result = truncateYear(ref)
	})

	nextYear := gp.Seq(I("next"), I("year")).Map(func(n *gp.Result) {
		n.Result = truncateYear(ref.AddDate(1, 0, 0))
	})

	color := gp.AnyWithName("color",
		I("white"), I("red"), I("green"), I("blue"), I("gold"), I("purple"), I("orange"), I("pink"),
		I("silver"), I("copper"))

	colorMonth := gp.Seq(color, month).Map(func(n *gp.Result) {
		c := strings.ToLower(n.Child[0].Token)
		m := n.Child[1].Result.(time.Month)
		color2delta := map[string]int{
			"white":  0,
			"red":    1,
			"green":  2,
			"blue":   3,
			"gold":   4,
			"purple": 5,
			"orange": 6,
			"pink":   7,
			"silver": 8,
			"copper": 9,
		}
		delta := color2delta[c]
		t := nextMonth(ref, m)
		dur := t.AddDate(0, 1, 0).Sub(t.Time)
		n.Result = Range{t.AddDate(delta, 0, 0), dur}
	})

	monthNoYear := gp.Seq(month, gp.Maybe(dayOfMonth)).Map(func(n *gp.Result) {
		var d Range
		m := n.Child[0].Result.(time.Month)
		switch o.defaultDirection {
		case future:
			d = nextMonth(ref, m)
		case past:
			d = prevMonth(ref, m)
		default:
			panic(fmt.Sprintf("invalid default direction: %q", o.defaultDirection))
		}
		n.Result = setDayMaybe(d, n.Child[1].Result)
	})

	weekdayNoDirection := gp.Seq(weekday).Map(func(n *gp.Result) {
		w := n.Child[0].Result.(time.Weekday)
		var t time.Time
		switch o.defaultDirection {
		case future:
			t = nextWeekdayFrom(ref, w)
		case past:
			t = prevWeekdayFrom(ref, w)
		default:
			panic(fmt.Sprintf("invalid default direction: %q", o.defaultDirection))
		}
		n.Result = Range{t, 24 * time.Hour}
	})

	yesterday := gp.Bind(I("yesterday"), truncateDay(ref.AddDate(0, 0, -1)))
	today := gp.Bind(I("today"), truncateDay(ref))
	tomorrow := gp.Bind(I("tomorrow"), truncateDay(ref.AddDate(0, 0, 1)))

	yearsLabel := gp.Regex(`(?i)years?`)

	xYearsAgo := gp.Seq(number, yearsLabel, I("ago")).Map(func(n *gp.Result) {
		dy := n.Child[0].Result.(int)
		y := ref.AddDate(-dy, 0, 0)
		dur := y.AddDate(1, 0, 0).Sub(y)
		n.Result = Range{y, dur}
	})

	fromNowOrToday := gp.Any(I("hence"), gp.Seq(I("from"), gp.Any(I("now"), I("today"))))

	xYearsFromToday := gp.Seq(number, yearsLabel, fromNowOrToday).Map(func(n *gp.Result) {
		dy := n.Child[0].Result.(int)
		y := ref.AddDate(dy, 0, 0)
		dur := y.AddDate(1, 0, 0).Sub(y)
		n.Result = Range{y, dur}
	})

	daysLabel := gp.Regex(`(?i)days?`)

	xDaysAgo := gp.Seq(number, daysLabel, I("ago")).Map(func(n *gp.Result) {
		delta := n.Child[0].Result.(int)
		d := ref.AddDate(0, 0, -delta)
		n.Result = Range{d, 24 * time.Hour}
	})

	xDaysFromNow := gp.Seq(number, daysLabel, fromNowOrToday).Map(func(n *gp.Result) {
		delta := n.Child[0].Result.(int)
		d := ref.AddDate(0, 0, delta)
		n.Result = Range{d, 24 * time.Hour}
	})

	weeksLabel := gp.Regex(`(?i)weeks?`)

	xWeeksAgo := gp.Seq(number, weeksLabel, I("ago")).Map(func(n *gp.Result) {
		delta := n.Child[0].Result.(int)
		d := ref.AddDate(0, 0, -7*delta)
		n.Result = Range{d, 7 * 24 * time.Hour}
	})

	xWeeksFromNow := gp.Seq(number, weeksLabel, fromNowOrToday).Map(func(n *gp.Result) {
		delta := n.Child[0].Result.(int)
		d := ref.AddDate(0, 0, 7*delta)
		n.Result = Range{d, 7 * 24 * time.Hour}
	})

	date := gp.AnyWithName("date",
		yesterday, today, tomorrow,
		ymdDate, dmyDate, mdyDate, myDate, ymDate,
		ymdNumDate, dmyNumDate,
		lastSpecificMonthDay, nextSpecificMonthDay,
		lastSpecificMonth, nextSpecificMonth,
		lastYear, thisYear, nextYear,
		nextMo, thisMo, prevMo,
		lastWeekday, nextWeekday,
		lastWeekParser, thisWeekParser, nextWeekParser,
		colorMonth, monthNoYear,
		weekdayNoDirection, yearEra,
		xDaysAgo, xDaysFromNow,
		xWeeksAgo, xWeeksFromNow,
		monthsAgo, monthsFromNow,
		xYearsAgo, xYearsFromToday)

	on := gp.Regex(`(?i)\bon\b`)
	onDate := gp.Seq(gp.Maybe(on), date).Map(func(n *gp.Result) {
		n.Result = n.Child[1].Result
	})

	tyme := gp.AnyWithName("time", hourMinuteSecond, noon)

	at := gp.Regex(`(?i)\b(at|@)\b`)
	atTimeWithMaybeZone := gp.Seq(gp.Maybe(at), tyme, gp.Maybe(zone)).Map(func(n *gp.Result) {
		t := n.Child[1].Result.(Range)
		z := ref.Location()
		c2 := n.Child[2].Result
		if c2 != nil {
			z = c2.(*time.Location)
		}
		n.Result = Range{
			setLocation(t.Time, z),
			t.Duration,
		}
	})

	onDateAtTime := gp.Seq(onDate, comma, atTimeWithMaybeZone).Map(func(n *gp.Result) {
		d := n.Child[0].Result.(Range)
		n.Result = setTimeMaybe(d, n.Child[2].Result)
	})

	atTimeOnDate := gp.Seq(atTimeWithMaybeZone, onDate).Map(func(n *gp.Result) {
		d := n.Child[1].Result.(Range)
		n.Result = setTimeMaybe(d, n.Child[0].Result)
	})

	onDateZone := gp.Seq(onDate, zone).Map(func(n *gp.Result) {
		d := n.Child[0].Result.(Range)
		z := n.Child[1].Result.(*time.Location)
		n.Result =
			Range{
				setLocation(d.Time, z),
				d.Duration,
			}
	})

	minutesLabel := gp.Regex(`(?i)minutes?`)

	xMinutesAgo := gp.Seq(number, minutesLabel, I("ago")).Map(func(n *gp.Result) {
		m := n.Child[0].Result.(int)
		n.Result = Range{
			ref.Add(-time.Duration(m) * time.Minute),
			time.Minute,
		}
	})

	fromNow := gp.Any(I("hence"), gp.Seq(I("from"), I("now")))

	xMinutesFromNow := gp.Seq(number, minutesLabel, fromNow).Map(func(n *gp.Result) {
		m := n.Child[0].Result.(int)
		n.Result = Range{
			ref.Add(time.Duration(m) * time.Minute),
			time.Minute,
		}
	})

	hoursLabel := gp.Regex(`(?i)hours?`)

	xHoursAgo := gp.Seq(number, hoursLabel, I("ago")).Map(func(n *gp.Result) {
		h := n.Child[0].Result.(int)
		n.Result = Range{
			ref.Add(-time.Duration(h) * time.Hour),
			time.Hour,
		}
	})

	xHoursFromNow := gp.Seq(number, hoursLabel, fromNow).Map(func(n *gp.Result) {
		h := n.Child[0].Result.(int)
		n.Result = Range{
			ref.Add(time.Duration(h) * time.Hour),
			time.Hour,
		}
	})

	return gp.AnyWithName("natural date",
		now,
		ansiC, rubyDate, rfc1123Z, rfc3339,
		onDateZone, atTimeOnDate, onDateAtTime,
		onDate, atTimeWithMaybeZone,
		xMinutesAgo, xMinutesFromNow,
		xHoursAgo, xHoursFromNow,
		hourMinuteSecond).Map(func(n *gp.Result) {
		pass()
	})
}

func thisWeek(ref time.Time) Range {
	return truncateWeek(ref)
}

func lastWeek(ref time.Time) Range {
	minus7 := ref.AddDate(0, 0, -7)
	return truncateWeek(minus7)
}

func nextWeek(ref time.Time) Range {
	plus7 := ref.AddDate(0, 0, 7)
	return truncateWeek(plus7)
}

// ParseRange parses a string such as "from april 20 at 5pm to may 5 at 9pm"
// and returns a Range.
func ParseRange(s string, ref time.Time, opts ...func(o *opts)) (Range, error) {
	p := RangeParser(ref, opts...)
	result, _, err := gp.Run(p, s, gp.UnicodeWhitespace)
	if err != nil {
		return Range{}, fmt.Errorf("running range parser: %w", err)
	}
	r := result.(Range)
	return r, nil
}

// RangeParser takes a reference time ref and returns a parser for date ranges.
func RangeParser(ref time.Time, options ...func(o *opts)) gp.Parser {
	preposition := gp.AnyWithName("a preposition such as to or until", "to", "until", "through", "til", "'til", "till")
	toPart := gp.Seq(preposition, Parser(ref, options...)).Map(func(n *gp.Result) {
		n.Result = n.Child[1].Result
	})
	return gp.Seq(gp.Maybe("from"), Parser(ref, options...), gp.Maybe(toPart)).Map(func(n *gp.Result) {
		s := n.Child[1].Result.(Range)
		c2 := n.Child[2].Result
		if c2 != nil {
			// This is an explicit range like "from A until B"
			e := c2.(Range)
			dur := e.Sub(s.Time)
			n.Result = Range{s.Time, dur}
			return
		}
		n.Result = s
	})
}

func setTimeMaybe(datePart Range, timePart any) Range {
	d := datePart
	if timePart == nil {
		return d
	}
	t := timePart.(Range)
	return Range{
		time.Date(d.Year(), d.Month(), d.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location()),
		t.Duration,
	}
}

func setDayMaybe(t Range, dayAsAny any) Range {
	if dayAsAny == nil {
		return t
	}
	d := dayAsAny.(int)
	return Range{
		time.Date(t.Year(), t.Month(), d, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location()),
		24 * time.Hour,
	}
}

func fixedZoneHM(h, m int) *time.Location {
	offset := h*60*60 + m*60
	sign := "+"
	if h < 0 {
		sign = "-"
		h = -h
	}
	name := fmt.Sprintf("%s%02d:%02d", sign, h, m)
	return time.FixedZone(name, offset)
}

func fixedZone(offsetHours int) *time.Location {
	return fixedZoneHM(offsetHours, 0)
}

// prevWeekdayFrom returns the previous week day relative to time t.
func prevWeekdayFrom(t time.Time, day time.Weekday) time.Time {
	d := t.Weekday() - day
	if d <= 0 {
		d += 7
	}
	return truncateDay(t.AddDate(0, 0, -int(d))).Time
}

// nextWeekdayFrom returns the next week day relative to time t.
func nextWeekdayFrom(t time.Time, day time.Weekday) time.Time {
	d := day - t.Weekday()
	if d <= 0 {
		d += 7
	}
	return truncateDay(t.AddDate(0, 0, int(d))).Time
}

// nextMonthDayTime returns the next month relative to time t, with given day of month and time of day.
func nextMonthDayTime(t time.Time, month time.Month, day int, hour int, min int, sec int) time.Time {
	nm := nextMonth(t, month)
	return time.Date(nm.Year(), nm.Month(), day, hour, min, sec, 0, t.Location())
}

// prevMonthDayTime returns the previous month relative to time t, with given day of month and time of day.
func prevMonthDayTime(t time.Time, month time.Month, day int, hour int, min int, sec int) time.Time {
	pm := prevMonth(t, month)
	return time.Date(pm.Year(), pm.Month(), day, hour, min, sec, 0, t.Location())
}

// nextMonth returns the next month relative to time t.
func nextMonth(t time.Time, month time.Month) Range {
	y := t.Year()
	if month-t.Month() <= 0 {
		y++
	}
	d := time.Date(y, month, 1, 0, 0, 0, 0, t.Location())
	dur := d.AddDate(0, 1, 0).Sub(d)
	return Range{d, dur}
}

// prevMonth returns the next month relative to time t.
func prevMonth(t time.Time, month time.Month) Range {
	y := t.Year()
	if t.Month()-month <= 0 {
		y--
	}
	d := time.Date(y, month, 1, 0, 0, 0, 0, t.Location())
	dur := d.AddDate(0, 1, 0).Sub(d)
	return Range{d, dur}
}

// truncateDay returns a date truncated to the day.
func truncateDay(t time.Time) Range {
	y, m, d := t.Date()
	s := time.Date(y, m, d, 0, 0, 0, 0, t.Location())
	e := s.AddDate(0, 0, 1)
	return Range{s, e.Sub(s)}
}

// truncateWeek returns a date truncated to the week.
func truncateWeek(t time.Time) Range {
	for t.Weekday() != time.Sunday {
		t = t.AddDate(0, 0, -1)
	}
	s := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	e := s.AddDate(0, 0, 7)
	return Range{s, e.Sub(s)}
}

// truncateMonth returns a date truncated to the month.
func truncateMonth(t time.Time) Range {
	y, m, _ := t.Date()
	s := time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
	e := s.AddDate(0, 1, 0)
	return Range{s, e.Sub(s)}
}

// truncateYear returns a date truncated to the year.
func truncateYear(t time.Time) Range {
	s := time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
	e := s.AddDate(1, 0, 0)
	return Range{s, e.Sub(s)}
}

// setTime takes the date from d and the time from the remaining args and
// returns the combined result.
func setTime(d time.Time, h, m, s, ns int) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), h, m, s, ns, d.Location())
}

// setLocation returns the given time t in location loc.
func setLocation(t time.Time, loc *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
}

// pass is something to call so you can put a breakpoint in an empty func.
func pass() {
}

// I makes a gp.Insensitive parser from its string argument.
func I(s string) gp.Parser {
	return gp.Insensitive(s)
}
