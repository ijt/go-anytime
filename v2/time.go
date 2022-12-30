package anytime

import (
	"fmt"
	"regexp"
	"time"
)

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

func setTimeMaybe(datePart Range, timePart *Range) Range {
	d := datePart.start
	if timePart == nil {
		return Range{d, 24 * time.Hour}
	}
	t := timePart.start
	return Range{
		time.Date(d.Year(), d.Month(), d.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location()),
		timePart.Duration,
	}
}

func setDayMaybe(r Range, maybeDay *int) Range {
	if maybeDay == nil {
		return r
	}
	t := r.start
	return Range{
		time.Date(t.Year(), t.Month(), *maybeDay, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location()),
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

// lastWeekdayFrom returns the previous week day relative to time t.
func lastWeekdayFrom(t time.Time, day time.Weekday) Range {
	d := t.Weekday() - day
	if d <= 0 {
		d += 7
	}
	return truncateDay(t.AddDate(0, 0, -int(d)))
}

// nextWeekdayFrom returns the next week day relative to time t.
func nextWeekdayFrom(t time.Time, day time.Weekday) Range {
	d := day - t.Weekday()
	if d <= 0 {
		d += 7
	}
	return truncateDay(t.AddDate(0, 0, int(d)))
}

// nextMonthDayTime returns the next month relative to time t, with given day of month and time of day.
func nextMonthDayTime(t time.Time, month time.Month, day int, hour int, min int, sec int) time.Time {
	nm := nextSpecificMonth(t, month)
	return time.Date(nm.start.Year(), nm.start.Month(), day, hour, min, sec, 0, t.Location())
}

// prevMonthDayTime returns the previous month relative to time t, with given day of month and time of day.
func prevMonthDayTime(t time.Time, month time.Month, day int, hour int, min int, sec int) time.Time {
	pm := lastSpecificMonth(t, month)
	return time.Date(pm.start.Year(), pm.start.Month(), day, hour, min, sec, 0, t.Location())
}

// nextMonth returns the next month relative to time now.
func nextMonth(now time.Time) Range {
	return truncateMonth(now.AddDate(0, 1, 0))
}

// thisMonth returns the current month relative to time now.
func thisMonth(now time.Time) Range {
	return truncateMonth(now)
}

// lastMonth returns the previous month relative to time now.
func lastMonth(now time.Time) Range {
	return truncateMonth(now.AddDate(0, -1, 0))
}

// nextYear returns the next year relative to time now.
func nextYear(now time.Time) Range {
	return truncateYear(now.AddDate(1, 0, 0))
}

// thisYear returns the current year relative to time now.
func thisYear(now time.Time) Range {
	return truncateYear(now)
}

// lastYear returns the previous year relative to time now.
func lastYear(now time.Time) Range {
	return truncateYear(now.AddDate(-1, 0, 0))
}

// nextSpecificMonth returns the next month relative to time t.
func nextSpecificMonth(t time.Time, month time.Month) Range {
	y := t.Year()
	if month-t.Month() <= 0 {
		y++
	}
	d := time.Date(y, month, 1, 0, 0, 0, 0, t.Location())
	e := d.AddDate(0, 1, 0)
	dur := e.Sub(d)
	return Range{d, dur}
}

// lastSpecificMonth returns the next month relative to time t.
func lastSpecificMonth(t time.Time, month time.Month) Range {
	y := t.Year()
	if t.Month()-month <= 0 {
		y--
	}
	d := time.Date(y, month, 1, 0, 0, 0, 0, t.Location())
	e := d.AddDate(0, 1, 0)
	dur := e.Sub(d)
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

// setRangeLocation returns the given range r in location loc.
func setRangeLocation(r Range, loc *time.Location) Range {
	return Range{setLocation(r.start, loc), r.Duration}
}

var nonDigitRx = regexp.MustCompile(`\D+`)

// fixDirectionWeek makes sure the time is in the given direction from now
// by possibly adding or subtracting a week.
func fixDirectionWeek(t time.Time, now time.Time, dir Direction) time.Time {
	if dir == Future {
		if t.Before(now) {
			t = t.AddDate(0, 0, 7)
		}
	} else if dir == Past {
		if t.After(now) {
			t = t.AddDate(0, 0, -7)
		}
	}
	return t
}

// fixDirectionYear makes sure the time is in the given direction from now
// by possibly adding or subtracting a year.
func fixDirectionYear(t time.Time, now time.Time, dir Direction) time.Time {
	if dir == Future {
		if t.Before(now) {
			t = t.AddDate(1, 0, 0)
		}
	} else if dir == Past {
		if t.After(now) {
			t = t.AddDate(-1, 0, 0)
		}
	}
	return t
}
