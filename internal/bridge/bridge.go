package bridge

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/emilijan-koteski/mktax/internal/buildinfo"
	"github.com/emilijan-koteski/mktax/internal/report"
	"github.com/emilijan-koteski/mktax/mktax"
	"github.com/emilijan-koteski/mktax/rates"
	"github.com/emilijan-koteski/mktax/t212csv"
)

const maxMonths = 120

type FileIn struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type ParseIn struct {
	Files     []FileIn        `json:"files"`
	FundUnits map[string]bool `json:"fund_units,omitempty"`
	Now       string          `json:"now,omitempty"`
}

type RatesNeeded struct {
	Months     []string `json:"months"`
	Currencies []string `json:"currencies"`
}

type Period struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type ParseOut struct {
	Import          report.Import `json:"import"`
	RatesNeeded     RatesNeeded   `json:"rates_needed"`
	SuggestedPeriod Period        `json:"suggested_period"`
}

type CalculateIn struct {
	Files     []FileIn        `json:"files"`
	Rates     rates.Table     `json:"rates"`
	From      string          `json:"from"`
	To        string          `json:"to"`
	Now       string          `json:"now,omitempty"`
	FundUnits map[string]bool `json:"fund_units,omitempty"`
}

type CalculateOut struct {
	Import report.Import `json:"import"`
	Report report.Report `json:"report"`
}

type MissingRate struct {
	Currency string `json:"currency"`
	Date     string `json:"date"`
}

type Error struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Missing []MissingRate `json:"missing,omitempty"`
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

func Version() string { return buildinfo.Version }

func Parse(in ParseIn) (ParseOut, *Error) {
	now, err := parseNow(in.Now)
	if err != nil {
		return ParseOut{}, &Error{Code: "bad_input", Message: err.Error()}
	}
	imp, perr := parseFiles(in.Files, in.FundUnits)
	if perr != nil {
		return ParseOut{}, perr
	}
	need, err := ratesNeeded(imp)
	if err != nil {
		return ParseOut{}, &Error{Code: "engine_error", Message: err.Error()}
	}
	return ParseOut{Import: report.ImportSummary(imp), RatesNeeded: need, SuggestedPeriod: suggested(imp, now)}, nil
}

// Runs the whole pipeline for the browser: parse the exports, validate the period, check every required rate is present, compute and shape the report. Errors carry a stable code so the UI can translate them.
func Calculate(in CalculateIn) (CalculateOut, *Error) {
	now, err := parseNow(in.Now)
	if err != nil {
		return CalculateOut{}, &Error{Code: "bad_input", Message: err.Error()}
	}
	imp, perr := parseFiles(in.Files, in.FundUnits)
	if perr != nil {
		return CalculateOut{}, perr
	}
	from, to, perr := period(in.From, in.To, suggested(imp, now), now)
	if perr != nil {
		return CalculateOut{}, perr
	}
	need, err := mktax.RequiredRates(imp.Trades, imp.Dividends, from, to)
	if err != nil {
		return CalculateOut{}, &Error{Code: "engine_error", Message: err.Error()}
	}
	if missing := missingRates(need, in.Rates); len(missing) > 0 {
		return CalculateOut{}, &Error{Code: "rates_missing", Message: fmt.Sprintf("%d NBRSM rates are missing", len(missing)), Missing: missing}
	}
	r, err := mktax.Compute(imp.Trades, imp.Dividends, in.Rates.RateFunc(), mktax.Options{From: from, To: to, Now: now})
	if err != nil {
		return CalculateOut{}, &Error{Code: "engine_error", Message: err.Error()}
	}
	return CalculateOut{Import: report.ImportSummary(imp), Report: report.New(r, imp, now)}, nil
}

func ParseJSON(in []byte) []byte {
	return envelope(func() (any, *Error) {
		var p ParseIn
		if err := json.Unmarshal(in, &p); err != nil {
			return nil, &Error{Code: "bad_input", Message: err.Error()}
		}
		return Parse(p)
	})
}

func CalculateJSON(in []byte) []byte {
	return envelope(func() (any, *Error) {
		var c CalculateIn
		if err := json.Unmarshal(in, &c); err != nil {
			return nil, &Error{Code: "bad_input", Message: err.Error()}
		}
		return Calculate(c)
	})
}

type envelopeOut struct {
	OK     bool   `json:"ok"`
	Result any    `json:"result,omitempty"`
	Error  *Error `json:"error,omitempty"`
}

func envelope(run func() (any, *Error)) (out []byte) {
	defer func() {
		if r := recover(); r != nil {
			out, _ = json.Marshal(envelopeOut{OK: false, Error: &Error{Code: "bad_input", Message: fmt.Sprint(r)}})
		}
	}()
	res, perr := run()
	if perr != nil {
		b, _ := json.Marshal(envelopeOut{OK: false, Error: perr})
		return b
	}
	b, err := json.Marshal(envelopeOut{OK: true, Result: res})
	if err != nil {
		b, _ = json.Marshal(envelopeOut{OK: false, Error: &Error{Code: "engine_error", Message: err.Error()}})
	}
	return b
}

func parseFiles(files []FileIn, fundUnits map[string]bool) (t212csv.Import, *Error) {
	in := make([]t212csv.File, 0, len(files))
	for _, f := range files {
		in = append(in, t212csv.File{Name: f.Name, Data: []byte(f.Data)})
	}
	imp, err := t212csv.Parse(in, t212csv.Options{FundUnits: fundUnits})
	if err != nil {
		return t212csv.Import{}, &Error{Code: "csv_error", Message: err.Error()}
	}
	return imp, nil
}

func ratesNeeded(imp t212csv.Import) (RatesNeeded, error) {
	out := RatesNeeded{Months: []string{}, Currencies: []string{}}
	if imp.Last.IsZero() {
		return out, nil
	}
	need, err := mktax.RequiredRates(imp.Trades, imp.Dividends, mktax.FirstMonth(), mktax.MonthOf(mktax.DateOf(imp.Last)))
	if err != nil {
		return out, err
	}
	months := map[string]bool{}
	for ccy, days := range need {
		out.Currencies = append(out.Currencies, ccy)
		for _, d := range days {
			months[mktax.MonthOf(d).String()] = true
		}
	}
	out.Months = slices.Sorted(maps.Keys(months))
	slices.Sort(out.Currencies)
	return out, nil
}

func suggested(imp t212csv.Import, now time.Time) Period {
	to := mktax.MonthOf(mktax.DateOf(now))
	from := to
	if !imp.First.IsZero() {
		from = mktax.MonthOf(mktax.DateOf(imp.First))
		if from.Compare(mktax.FirstMonth()) < 0 {
			from = mktax.FirstMonth()
		}
	}
	if from.Compare(to) > 0 {
		from = to
	}
	return Period{From: from.String(), To: to.String()}
}

func period(fromS, toS string, sug Period, now time.Time) (mktax.Month, mktax.Month, *Error) {
	if fromS == "" {
		fromS = sug.From
	}
	if toS == "" {
		toS = sug.To
	}
	from, err := mktax.ParseMonth(fromS)
	if err != nil {
		return mktax.Month{}, mktax.Month{}, &Error{Code: "invalid_period", Message: err.Error()}
	}
	to, err := mktax.ParseMonth(toS)
	if err != nil {
		return mktax.Month{}, mktax.Month{}, &Error{Code: "invalid_period", Message: err.Error()}
	}
	current := mktax.MonthOf(mktax.DateOf(now))
	var msg string
	switch {
	case from.Compare(mktax.FirstMonth()) < 0:
		msg = fmt.Sprintf("the rules apply from %s; got %s", mktax.FirstMonth(), from)
	case from.Compare(to) > 0:
		msg = fmt.Sprintf("from %s is after to %s", from, to)
	case to.Compare(current) > 0:
		msg = fmt.Sprintf("to %s is after the current month %s", to, current)
	case monthsBetween(from, to) > maxMonths:
		msg = fmt.Sprintf("at most %d months at a time", maxMonths)
	}
	if msg != "" {
		return mktax.Month{}, mktax.Month{}, &Error{Code: "invalid_period", Message: msg}
	}
	return from, to, nil
}

func monthsBetween(a, b mktax.Month) int {
	return (b.Year-a.Year)*12 + int(b.Month-a.Month) + 1
}

func missingRates(need map[string][]mktax.Date, tbl rates.Table) []MissingRate {
	var out []MissingRate
	for _, ccy := range slices.Sorted(maps.Keys(need)) {
		for _, d := range need[ccy] {
			if _, err := tbl.Lookup(ccy, d.Time()); err != nil {
				out = append(out, MissingRate{Currency: ccy, Date: d.String()})
			}
		}
	}
	return out
}

func parseNow(s string) (time.Time, error) {
	if s == "" {
		return time.Now(), nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("now: %w", err)
	}
	return t, nil
}
