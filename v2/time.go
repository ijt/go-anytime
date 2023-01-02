package anytime

import (
	"fmt"
	"time"
)

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
