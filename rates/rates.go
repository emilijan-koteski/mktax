package rates

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/emilijan-koteski/mktax/mktax"
)

const DayLayout = time.DateOnly

var (
	ErrDayMissing      = errors.New("no NBRSM rate list")
	ErrCurrencyMissing = errors.New("currency not published")
	one                = decimal.NewFromInt(1)
)

type Rate struct {
	Units  int64           `json:"units"`
	Middle decimal.Decimal `json:"middle"`
}

func (r Rate) PerUnit() decimal.Decimal {
	if r.Units == 1 {
		return r.Middle
	}
	return r.Middle.Div(decimal.NewFromInt(r.Units))
}

type Table map[string]map[string]Rate

func DayKey(t time.Time) string {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC).Format(DayLayout)
}

func (t Table) Set(day, currency string, r Rate) {
	if t[day] == nil {
		t[day] = map[string]Rate{}
	}
	t[day][currency] = r
}

func (t Table) Merge(o Table) {
	for day, byCcy := range o {
		for ccy, r := range byCcy {
			t.Set(day, ccy, r)
		}
	}
}

func (t Table) Filter(currencies []string) Table {
	keep := map[string]bool{}
	for _, c := range currencies {
		keep[normalize(c)] = true
	}
	out := Table{}
	for day, byCcy := range t {
		for ccy, r := range byCcy {
			if len(keep) == 0 || keep[ccy] {
				out.Set(day, ccy, r)
			}
		}
	}
	return out
}

func (t Table) Days() []string {
	out := make([]string, 0, len(t))
	for day := range t {
		out = append(out, day)
	}
	slices.Sort(out)
	return out
}

func (t Table) Missing(from, to time.Time) []string {
	var out []string
	for d := DayKey(from); d <= DayKey(to); d = next(d) {
		if _, ok := t[d]; !ok {
			out = append(out, d)
		}
	}
	return out
}

func (t Table) Lookup(currency string, day time.Time) (decimal.Decimal, error) {
	ccy := normalize(currency)
	if ccy == "MKD" {
		return one, nil
	}
	k := DayKey(day)
	byCcy, ok := t[k]
	if !ok {
		return decimal.Zero, &lookupError{ErrDayMissing, fmt.Sprintf("no NBRSM rate list for %s (not published yet, or not prefetched)", k)}
	}
	r, ok := byCcy[ccy]
	if !ok {
		return decimal.Zero, &lookupError{ErrCurrencyMissing, fmt.Sprintf("NBRSM publishes no %s rate on %s", ccy, k)}
	}
	return r.PerUnit(), nil
}

type lookupError struct {
	kind error
	msg  string
}

func (e *lookupError) Error() string { return e.msg }

func (e *lookupError) Unwrap() error { return e.kind }

func (t Table) RateFunc() mktax.RateFunc {
	return func(currency string, d mktax.Date) (decimal.Decimal, error) {
		return t.Lookup(currency, d.Time())
	}
}

func normalize(currency string) string { return strings.ToUpper(strings.TrimSpace(currency)) }

func next(day string) string {
	t, _ := time.Parse(DayLayout, day)
	return t.AddDate(0, 0, 1).Format(DayLayout)
}
