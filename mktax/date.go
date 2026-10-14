package mktax

import (
	"fmt"
	"time"
	_ "time/tzdata"
)

var skopje = loadSkopje()

func loadSkopje() *time.Location {
	loc, err := time.LoadLocation("Europe/Skopje")
	if err != nil {
		panic("mktax: Europe/Skopje zone unavailable: " + err.Error())
	}
	return loc
}

func Location() *time.Location { return skopje }

type Date struct {
	Year  int
	Month time.Month
	Day   int
}

func DateOf(t time.Time) Date {
	l := t.In(skopje)
	return Date{l.Year(), l.Month(), l.Day()}
}

func dateOfUTC(t time.Time) Date { return Date{t.Year(), t.Month(), t.Day()} }

func (d Date) Time() time.Time { return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC) }

func (d Date) AddDate(years, months, days int) Date {
	return dateOfUTC(d.Time().AddDate(years, months, days))
}

func (d Date) Compare(o Date) int { return d.Time().Compare(o.Time()) }
func (d Date) Before(o Date) bool { return d.Compare(o) < 0 }
func (d Date) After(o Date) bool  { return d.Compare(o) > 0 }
func (d Date) IsZero() bool       { return d == Date{} }
func (d Date) String() string {
	if d.IsZero() {
		return ""
	}
	return d.Time().Format(time.DateOnly)
}

type Month struct {
	Year  int
	Month time.Month
}

func MonthOf(d Date) Month { return Month{d.Year, d.Month} }

func ParseMonth(s string) (Month, error) {
	t, err := time.Parse("2006-01", s)
	if err != nil || len(s) != len("2006-01") {
		return Month{}, fmt.Errorf("invalid month %q: use YYYY-MM", s)
	}
	return Month{t.Year(), t.Month()}, nil
}

func (m Month) First() Date { return Date{m.Year, m.Month, 1} }
func (m Month) Last() Date  { return m.AddMonths(1).First().AddDate(0, 0, -1) }

func (m Month) AddMonths(n int) Month {
	t := time.Date(m.Year, m.Month, 1, 0, 0, 0, 0, time.UTC).AddDate(0, n, 0)
	return Month{t.Year(), t.Month()}
}

func (m Month) Compare(o Month) int { return m.First().Compare(o.First()) }
func (m Month) IsZero() bool        { return m == Month{} }
func (m Month) String() string {
	if m.IsZero() {
		return ""
	}
	return fmt.Sprintf("%04d-%02d", m.Year, int(m.Month))
}

// Returns the first date on which a lot acquired on the given date is exempt: more than two calendar years later.
func TaxFreeFrom(acquired Date) Date { return acquired.AddDate(exemptYears, 0, 0).AddDate(0, 0, 1) }

func fileBy(m Month) Date { return m.AddMonths(1).First().AddDate(0, 0, 9) }
func payBy(m Month) Date  { return m.AddMonths(1).First().AddDate(0, 0, 14) }
