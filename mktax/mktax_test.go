package mktax

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func d(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func at(s string) time.Time {
	v, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return v
}

func month(s string) Month {
	m, err := ParseMonth(s)
	if err != nil {
		panic(err)
	}
	return m
}

func eq(t *testing.T, name string, got decimal.Decimal, want string) {
	t.Helper()
	if !got.Equal(d(want)) {
		t.Errorf("%s = %s, want %s", name, got, want)
	}
}

func flat(rates map[string]string) RateFunc {
	return func(ccy string, _ Date) (decimal.Decimal, error) {
		r, ok := rates[ccy]
		if !ok {
			return decimal.Zero, errors.New("no rate for " + ccy)
		}
		return d(r), nil
	}
}

func dated(rates map[string]string) RateFunc {
	return func(ccy string, day Date) (decimal.Decimal, error) {
		r, ok := rates[ccy+" "+day.String()]
		if !ok {
			return decimal.Zero, errors.New("no rate for " + ccy + " " + day.String())
		}
		return d(r), nil
	}
}

func buy(isin, when, qty, value, ccy string) Trade {
	return Trade{Ticker: isin + "_T", ISIN: isin, Time: at(when), Side: Buy, Quantity: d(qty), Value: d(value), Currency: ccy}
}

func sell(isin, when, qty, value, ccy string) Trade {
	t := buy(isin, when, qty, value, ccy)
	t.Side = Sell
	return t
}

func opts(from, to, now string) Options {
	return Options{From: month(from), To: month(to), Now: at(now)}
}

func onlySale(t *testing.T, r Report) Sale {
	t.Helper()
	var sales []Sale
	for _, m := range r.Months {
		sales = append(sales, m.Sales...)
	}
	if len(sales) != 1 {
		t.Fatalf("got %d sales, want 1: %+v", len(sales), sales)
	}
	return sales[0]
}

func TestFIFOFractionalPartialLot(t *testing.T) {
	trades := []Trade{
		sell("US0000000001", "2026-05-15T10:00:00Z", "2.5", "400", "USD"),
		buy("US0000000001", "2025-01-10T10:00:00Z", "1.5", "150", "USD"),
		buy("US0000000001", "2025-03-10T10:00:00Z", "5.5", "660", "USD"),
	}
	r, err := Compute(trades, nil, flat(map[string]string{"USD": "50"}), opts("2026-05", "2026-05", "2026-06-20T12:00:00Z"))
	if err != nil {
		t.Fatal(err)
	}
	s := onlySale(t, r)
	if len(s.Lots) != 2 {
		t.Fatalf("lots = %+v", s.Lots)
	}
	l1, l2 := s.Lots[0], s.Lots[1]
	if l1.Acquired != (Date{2025, time.January, 10}) || l2.Acquired != (Date{2025, time.March, 10}) {
		t.Errorf("acquired = %s, %s", l1.Acquired, l2.Acquired)
	}
	eq(t, "lot1 qty", l1.Quantity, "1.5")
	eq(t, "lot1 cost", l1.Cost, "150")
	eq(t, "lot1 proceeds", l1.Proceeds, "240")
	eq(t, "lot2 qty", l2.Quantity, "1")
	eq(t, "lot2 cost", l2.Cost, "120")
	eq(t, "lot2 proceeds", l2.Proceeds, "160")
	if l1.Status != StatusTaxable || l2.Status != StatusTaxable {
		t.Errorf("status = %s, %s", l1.Status, l2.Status)
	}
	eq(t, "cost", s.Cost, "270")
	eq(t, "gain", s.Gain, "130")
	eq(t, "taxable gain", s.TaxableGain, "130")
	eq(t, "proceeds mkd", s.ProceedsMKD, "20000")
	eq(t, "cost mkd", s.CostMKD, "13500")
	eq(t, "taxable gain mkd", s.TaxableGainMKD, "6500")
	eq(t, "tax mkd", s.TaxMKD, "585")
	if !s.CostComplete {
		t.Error("cost should be complete")
	}

	if len(r.OpenLots) != 1 {
		t.Fatalf("open lots = %+v", r.OpenLots)
	}
	op := r.OpenLots[0]
	eq(t, "open qty", op.Quantity, "4.5")
	eq(t, "open cost", op.Cost, "540")
	if len(op.Lots) != 1 || op.EarliestAcquired != (Date{2025, time.March, 10}) || op.OldestLotTaxFreeFrom != (Date{2027, time.March, 11}) {
		t.Errorf("open = %+v", op)
	}
	eq(t, "tax-free now", op.TaxFreeQuantityNow, "0")
}

func TestExemptionBoundary(t *testing.T) {
	tests := []struct {
		name     string
		acquired string
		sold     string
		want     LotStatus
	}{
		{"exactly two years is not exempt", "2024-05-10T10:00:00Z", "2026-05-10T10:00:00Z", StatusTaxable},
		{"two years and a day is exempt", "2024-05-10T10:00:00Z", "2026-05-11T10:00:00Z", StatusExempt},
		{"local date decides (bought 00:30 local on 10 May, 9 May in UTC)", "2024-05-09T22:30:00Z", "2026-05-10T10:00:00Z", StatusTaxable},
		{"pre-2023 lot within two years is out of scope", "2022-11-15T10:00:00Z", "2024-06-03T10:00:00Z", StatusOutOfScope},
		{"pre-2023 lot held over two years is exempt", "2022-11-15T10:00:00Z", "2025-06-03T10:00:00Z", StatusExempt},
		{"first day of the regime is in scope", "2023-01-01T10:00:00Z", "2024-06-03T10:00:00Z", StatusTaxable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trades := []Trade{buy("US0000000001", tt.acquired, "1", "100", "MKD"), sell("US0000000001", tt.sold, "1", "150", "MKD")}
			m := MonthOf(DateOf(at(tt.sold))).String()
			r, err := Compute(trades, nil, flat(nil), opts(m, m, "2026-10-01T00:00:00Z"))
			if err != nil {
				t.Fatal(err)
			}
			s := onlySale(t, r)
			if s.Lots[0].Status != tt.want {
				t.Fatalf("status = %s, want %s", s.Lots[0].Status, tt.want)
			}
			tot := r.Months[0].Totals
			switch tt.want {
			case StatusTaxable:
				eq(t, "taxable", tot.TaxableGainsMKD, "50")
				eq(t, "tax", tot.CapitalGainsTaxMKD, "4.5")
			case StatusExempt:
				eq(t, "taxable", tot.TaxableGainsMKD, "0")
				eq(t, "exempt", tot.ExemptGainsMKD, "50")
			case StatusOutOfScope:
				eq(t, "taxable", tot.TaxableGainsMKD, "0")
				eq(t, "out of scope", tot.OutOfScopeGainsMKD, "50")
			}
		})
	}
}

func TestLossesNotNetted(t *testing.T) {
	trades := []Trade{
		buy("US0000000001", "2025-06-01T10:00:00Z", "10", "1000", "MKD"),
		buy("US0000000002", "2025-06-01T10:00:00Z", "5", "1000", "MKD"),
		sell("US0000000001", "2026-03-05T10:00:00Z", "10", "2000", "MKD"),
		sell("US0000000002", "2026-03-06T10:00:00Z", "5", "600", "MKD"),
	}
	r, err := Compute(trades, nil, flat(nil), opts("2026-03", "2026-03", "2026-03-20T00:00:00Z"))
	if err != nil {
		t.Fatal(err)
	}
	tot := r.Months[0].Totals
	if tot.Sales != 2 {
		t.Fatalf("sales = %d", tot.Sales)
	}
	eq(t, "taxable gains", tot.TaxableGainsMKD, "1000")
	eq(t, "losses", tot.TaxableLossesMKD, "-400")
	eq(t, "net (info)", tot.NetIfLossesOffsetMKD, "600")
	eq(t, "tax base", tot.TaxBaseMKD, "900")
	eq(t, "tax", tot.CapitalGainsTaxMKD, "90")
	eq(t, "tax if offset (info)", tot.TaxIfLossesOffsetMKD, "54")
	eq(t, "loss sale tax", r.Months[0].Sales[1].TaxMKD, "0")
	if !r.Months[0].InProgress {
		t.Error("month containing Now should be in progress")
	}
}

func TestFXMethods(t *testing.T) {
	rates := dated(map[string]string{"USD 2025-02-03": "60", "USD 2026-02-03": "50"})
	trades := []Trade{
		buy("US0000000001", "2025-02-03T10:00:00Z", "10", "1000", "USD"),
		sell("US0000000001", "2026-02-03T10:00:00Z", "10", "1100", "USD"),
	}
	r, err := Compute(trades, nil, rates, opts("2026-02", "2026-02", "2026-02-20T00:00:00Z"))
	if err != nil {
		t.Fatal(err)
	}
	s := onlySale(t, r)
	l := s.Lots[0]
	eq(t, "sale rate", s.SaleRate, "50")
	eq(t, "acq rate", l.AcquisitionRate, "60")
	eq(t, "A: gain", l.GainMKD, "-5000")
	eq(t, "B: gain", l.GainMKDB, "5000")
	eq(t, "A: taxable", s.TaxableGainMKD, "-5000")
	eq(t, "B: taxable", s.TaxableGainMKDB, "5000")
	eq(t, "taxable gain in USD", s.TaxableGain, "100")
	tot := r.Months[0].Totals
	eq(t, "A: tax", tot.CapitalGainsTaxMKD, "0")
	eq(t, "B: taxable gains", tot.TaxableGainsMKDB, "5000")
	eq(t, "B: tax", tot.CapitalGainsTaxMKDB, "450")
}

func TestDividends(t *testing.T) {
	tests := []struct {
		name                                         string
		div                                          Dividend
		grossMKD, withheld, accepted, rejected, warn string
		country                                      string
	}{
		{
			name:     "US 15% withheld: credit covers the full 10%",
			div:      Dividend{Ticker: "A", ISIN: "US0000000001", Gross: d("10"), GrossCurrency: "USD", Net: d("8.5"), NetCurrency: "USD"},
			grossMKD: "500", withheld: "75", accepted: "0", rejected: "50", country: "САД",
		},
		{
			name:     "no withholding",
			div:      Dividend{Ticker: "B", ISIN: "IE0000000002", Gross: d("10"), GrossCurrency: "USD", Net: d("10"), NetCurrency: "USD"},
			grossMKD: "500", withheld: "0", accepted: "50", rejected: "50", country: "Ирска",
		},
		{
			name:     "5% withheld: partial credit",
			div:      Dividend{Ticker: "C", ISIN: "GB0000000003", Gross: d("10"), GrossCurrency: "USD", Net: d("9.5"), NetCurrency: "USD"},
			grossMKD: "500", withheld: "25", accepted: "25", rejected: "50", country: "Велика Британија",
		},
		{
			name:     "different gross and net currencies",
			div:      Dividend{Ticker: "D", ISIN: "US0000000004", Gross: d("10"), GrossCurrency: "USD", Net: d("8"), NetCurrency: "EUR"},
			grossMKD: "500", withheld: "8", accepted: "42", rejected: "50", country: "САД",
		},
		{
			name:     "gross not reported falls back to net",
			div:      Dividend{Ticker: "E", ISIN: "US0000000005", Gross: decimal.Zero, GrossCurrency: "USD", Net: d("9"), NetCurrency: "USD"},
			grossMKD: "450", withheld: "0", accepted: "45", rejected: "45", warn: "gross", country: "САД",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.div.PaidOn = at("2026-04-15T00:00:00Z")
			r, err := Compute(nil, []Dividend{tt.div}, flat(map[string]string{"USD": "50", "EUR": "61.5"}), opts("2026-04", "2026-04", "2026-06-01T00:00:00Z"))
			if err != nil {
				t.Fatal(err)
			}
			m := r.Months[0]
			if len(m.Dividends) != 1 {
				t.Fatalf("dividends = %+v", m.Dividends)
			}
			l := m.Dividends[0]
			eq(t, "gross mkd", l.GrossMKD, tt.grossMKD)
			eq(t, "withheld", l.WithheldMKD, tt.withheld)
			eq(t, "credit accepted", l.TaxIfCreditAcceptedMKD, tt.accepted)
			eq(t, "credit rejected", l.TaxIfCreditRejectedMKD, tt.rejected)
			eq(t, "month accepted", m.Totals.TaxDueIfCreditAcceptedMKD, tt.accepted)
			eq(t, "month rejected", m.Totals.TaxDueIfCreditRejectedMKD, tt.rejected)
			if l.Country.NameMK != tt.country || l.IncomeType != IncomeTypeCapital || l.IncomeSubtype != IncomeSubtypeDividends {
				t.Errorf("codes = %+v", l)
			}
			if got := strings.Join(l.Warnings, " "); (tt.warn == "") != (got == "") || !strings.Contains(got, tt.warn) {
				t.Errorf("warnings = %q, want containing %q", got, tt.warn)
			}
			if !m.Overdue || m.FileBy != (Date{2026, time.May, 10}) || m.PayBy != (Date{2026, time.May, 15}) {
				t.Errorf("deadlines = %s %s overdue=%v", m.FileBy, m.PayBy, m.Overdue)
			}
		})
	}
}

func TestMonthBoundaryIsSkopjeLocal(t *testing.T) {
	trades := []Trade{
		buy("US0000000001", "2026-01-05T10:00:00Z", "1", "100", "MKD"),
		sell("US0000000001", "2026-05-31T22:30:00Z", "1", "120", "MKD"),
	}
	divs := []Dividend{{Ticker: "A", ISIN: "US0000000001", PaidOn: at("2026-05-31T22:30:00Z"), Gross: d("10"), GrossCurrency: "MKD", Net: d("9"), NetCurrency: "MKD"}}
	r, err := Compute(trades, divs, flat(nil), opts("2026-05", "2026-06", "2026-07-01T00:00:00Z"))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Months) != 2 {
		t.Fatalf("months = %d", len(r.Months))
	}
	may, june := r.Months[0], r.Months[1]
	if len(may.Sales) != 0 || len(may.Dividends) != 0 || may.HasEntries || may.Overdue {
		t.Errorf("May = %+v", may)
	}
	if len(june.Sales) != 1 || len(june.Dividends) != 1 || june.Sales[0].Date != (Date{2026, time.June, 1}) {
		t.Errorf("June = %+v", june)
	}
}

func TestOversellWarnsInsteadOfFailing(t *testing.T) {
	trades := []Trade{
		buy("US0000000001", "2026-01-05T10:00:00Z", "1", "100", "MKD"),
		sell("US0000000001", "2026-02-05T10:00:00Z", "3", "330", "MKD"),
	}
	r, err := Compute(trades, nil, flat(nil), opts("2026-02", "2026-02", "2026-03-01T00:00:00Z"))
	if err != nil {
		t.Fatal(err)
	}
	s := onlySale(t, r)
	if s.CostComplete || len(s.Lots) != 2 || s.Lots[1].Status != StatusUnknownCost || len(s.Warnings) == 0 {
		t.Fatalf("sale = %+v", s)
	}
	eq(t, "unmatched qty", s.Lots[1].Quantity, "2")
	eq(t, "unmatched proceeds", s.Lots[1].Proceeds, "220")
	eq(t, "known gain only", s.TaxableGainMKD, "10")
	eq(t, "proceeds mkd covers whole sale", s.ProceedsMKD, "330")
	if r.Months[0].Totals.UnknownCostSales != 1 || len(r.Months[0].Warnings) == 0 || len(r.Warnings) == 0 {
		t.Errorf("month = %+v report warnings = %v", r.Months[0], r.Warnings)
	}
	if len(r.OpenLots) != 0 {
		t.Errorf("open lots = %+v", r.OpenLots)
	}
}

func TestEmptyInputs(t *testing.T) {
	r, err := Compute(nil, nil, flat(nil), opts("2026-01", "2026-03", "2026-10-01T00:00:00Z"))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Months) != 3 || len(r.OpenLots) != 0 || len(r.Notes) == 0 {
		t.Fatalf("report = %+v", r)
	}
	for _, m := range r.Months {
		if m.HasEntries || m.Overdue || len(m.Sales) != 0 || len(m.Dividends) != 0 {
			t.Errorf("month %s = %+v", m.Month, m)
		}
		eq(t, m.Month.String()+" tax", m.Totals.TaxDueIfCreditRejectedMKD, "0")
	}
	if r.Months[2].Month != (Month{2026, time.March}) {
		t.Errorf("last month = %s", r.Months[2].Month)
	}
}

func TestFIFOUsesWholeHistory(t *testing.T) {
	trades := []Trade{
		buy("US0000000001", "2025-01-10T10:00:00Z", "2", "200", "MKD"),
		buy("US0000000001", "2025-02-10T10:00:00Z", "2", "300", "MKD"),
		sell("US0000000001", "2025-06-10T10:00:00Z", "1", "150", "MKD"),
		sell("US0000000001", "2026-02-10T10:00:00Z", "2", "400", "MKD"),
	}
	r, err := Compute(trades, nil, flat(nil), opts("2026-02", "2026-02", "2026-03-01T00:00:00Z"))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Months) != 1 {
		t.Fatalf("months = %d", len(r.Months))
	}
	s := onlySale(t, r)
	if len(s.Lots) != 2 || s.Lots[0].Acquired != (Date{2025, time.January, 10}) || s.Lots[1].Acquired != (Date{2025, time.February, 10}) {
		t.Fatalf("lots = %+v", s.Lots)
	}
	eq(t, "lot1 qty", s.Lots[0].Quantity, "1")
	eq(t, "lot1 cost", s.Lots[0].Cost, "100")
	eq(t, "lot2 cost", s.Lots[1].Cost, "150")
	eq(t, "gain", s.TaxableGainMKD, "150")
}

func TestSameInstantBuyBeforeSellAndISINBackfill(t *testing.T) {
	b := buy("US0000000001", "2026-02-10T10:00:00Z", "1", "100", "MKD")
	b.ISIN = ""
	s := sell("US0000000001", "2026-02-10T10:00:00Z", "1", "110", "MKD")
	r, err := Compute([]Trade{s, b}, nil, flat(nil), opts("2026-02", "2026-02", "2026-03-01T00:00:00Z"))
	if err != nil {
		t.Fatal(err)
	}
	got := onlySale(t, r)
	if !got.CostComplete {
		t.Fatalf("sale = %+v", got)
	}
	eq(t, "gain", got.TaxableGainMKD, "10")
}

func TestDeadlinesAndOverdue(t *testing.T) {
	trades := []Trade{buy("US0000000001", "2026-01-05T10:00:00Z", "1", "100", "MKD"), sell("US0000000001", "2026-09-15T10:00:00Z", "1", "110", "MKD")}
	tests := []struct {
		now     string
		overdue bool
	}{
		{"2026-10-10T21:00:00Z", false},
		{"2026-10-10T22:30:00Z", true},
	}
	for _, tt := range tests {
		r, err := Compute(trades, nil, flat(nil), opts("2026-09", "2026-09", tt.now))
		if err != nil {
			t.Fatal(err)
		}
		m := r.Months[0]
		if m.Overdue != tt.overdue || m.FileBy != (Date{2026, time.October, 10}) || m.PayBy != (Date{2026, time.October, 15}) || m.InProgress {
			t.Errorf("now %s: %+v", tt.now, m)
		}
	}
}

func TestOpenLotsTaxFreeNow(t *testing.T) {
	trades := []Trade{
		buy("US0000000001", "2024-01-10T10:00:00Z", "2", "200", "MKD"),
		buy("US0000000001", "2025-01-10T10:00:00Z", "3", "300", "MKD"),
		buy("IE0000000002", "2026-01-10T10:00:00Z", "1", "50", "MKD"),
	}
	r, err := Compute(trades, nil, flat(nil), opts("2026-02", "2026-02", "2026-03-01T00:00:00Z"))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.OpenLots) != 2 || r.OpenLots[0].ISIN != "IE0000000002" {
		t.Fatalf("open lots = %+v", r.OpenLots)
	}
	us := r.OpenLots[1]
	eq(t, "qty", us.Quantity, "5")
	eq(t, "tax-free now", us.TaxFreeQuantityNow, "2")
	if us.OldestLotTaxFreeFrom != (Date{2026, time.January, 11}) || us.Lots[1].TaxFreeFrom != (Date{2027, time.January, 11}) {
		t.Errorf("us = %+v", us)
	}
}

func TestComputeErrors(t *testing.T) {
	good := opts("2026-02", "2026-02", "2026-03-01T00:00:00Z")
	usd := []Trade{buy("US0000000001", "2026-01-05T10:00:00Z", "1", "100", "USD"), sell("US0000000001", "2026-02-05T10:00:00Z", "1", "110", "USD")}
	tests := []struct {
		name   string
		trades []Trade
		rates  RateFunc
		opts   Options
		want   string
	}{
		{"from after to", nil, flat(nil), opts("2026-03", "2026-02", "2026-03-01T00:00:00Z"), "after"},
		{"before the regime", nil, flat(nil), opts("2022-12", "2023-01", "2026-03-01T00:00:00Z"), "2023-01"},
		{"no now", nil, flat(nil), Options{From: month("2026-02"), To: month("2026-02")}, "Now"},
		{"no rates", nil, nil, good, "rate"},
		{"zero quantity", []Trade{buy("US0000000001", "2026-01-05T10:00:00Z", "0", "100", "USD")}, flat(nil), good, "quantity"},
		{"no currency", []Trade{buy("US0000000001", "2026-01-05T10:00:00Z", "1", "100", "")}, flat(nil), good, "currency"},
		{"no instrument", []Trade{{Time: at("2026-01-05T10:00:00Z"), Side: Buy, Quantity: d("1"), Value: d("1"), Currency: "MKD"}}, flat(nil), good, "instrument"},
		{"bad side", []Trade{{ISIN: "X", Time: at("2026-01-05T10:00:00Z"), Side: "HOLD", Quantity: d("1"), Value: d("1"), Currency: "MKD"}}, flat(nil), good, "side"},
		{"missing rate", usd, flat(nil), good, "USD on 2026-0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Compute(tt.trades, nil, tt.rates, tt.opts)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("err = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestRequiredRates(t *testing.T) {
	trades := []Trade{
		buy("US0000000001", "2025-01-10T10:00:00Z", "2", "200", "USD"),
		sell("US0000000001", "2025-06-10T10:00:00Z", "1", "150", "USD"),
		sell("US0000000001", "2026-05-15T10:00:00Z", "1", "150", "USD"),
		buy("US0000000002", "2026-05-16T10:00:00Z", "1", "100", "USD"),
		buy("MK0000000003", "2026-01-16T10:00:00Z", "1", "100", "MKD"),
		sell("MK0000000003", "2026-05-17T10:00:00Z", "1", "100", "MKD"),
	}
	divs := []Dividend{
		{Ticker: "A", ISIN: "US0000000001", PaidOn: at("2026-05-20T00:00:00Z"), Gross: d("1"), GrossCurrency: "USD", Net: d("1"), NetCurrency: "EUR"},
		{Ticker: "A", ISIN: "US0000000001", PaidOn: at("2026-04-20T00:00:00Z"), Gross: d("1"), GrossCurrency: "GBP", Net: d("1"), NetCurrency: "GBP"},
	}
	got, err := RequiredRates(trades, divs, month("2026-05"), month("2026-05"))
	if err != nil {
		t.Fatal(err)
	}
	str := func(ds []Date) string {
		var ss []string
		for _, x := range ds {
			ss = append(ss, x.String())
		}
		return strings.Join(ss, ",")
	}
	if len(got) != 2 || str(got["USD"]) != "2025-01-10,2026-05-15,2026-05-20" || str(got["EUR"]) != "2026-05-20" {
		t.Fatalf("got %v", got)
	}
}
