package report

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/emilijan-koteski/mktax/mktax"
	"github.com/emilijan-koteski/mktax/rates"
	"github.com/emilijan-koteski/mktax/t212csv"
)

func d(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func at(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func trade(side mktax.Side, isin, ts, qty, value, ccy string) mktax.Trade {
	return mktax.Trade{Ticker: isin + "_T", ISIN: isin, Name: "Test " + isin, Time: at(ts), Side: side, Quantity: d(qty), Value: d(value), Currency: ccy, Ref: "R-" + ts}
}

func flat(perCcy map[string]string) mktax.RateFunc {
	return func(ccy string, _ mktax.Date) (decimal.Decimal, error) { return d(perCcy[ccy]), nil }
}

var now = at("2026-10-14T12:00:00Z")

func compute(t *testing.T, trades []mktax.Trade, divs []mktax.Dividend, from, to string, rf mktax.RateFunc) mktax.Report {
	t.Helper()
	f, _ := mktax.ParseMonth(from)
	tt, _ := mktax.ParseMonth(to)
	r, err := mktax.Compute(trades, divs, rf, mktax.Options{From: f, To: tt, Now: now})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestFormatters(t *testing.T) {
	cases := []struct{ in, want string }{{"1234.5", "1234.50"}, {"0", "0.00"}, {"-0.005", "-0.01"}, {"2.345", "2.35"}, {"10", "10.00"}}
	for _, c := range cases {
		if got := money(d(c.in)); got != c.want {
			t.Errorf("money(%s) = %s, want %s", c.in, got, c.want)
		}
	}
	if num(d("0.12345678")) != "0.12345678" || num(d("50.1234")) != "50.1234" {
		t.Error("num must be exact")
	}
	if date(mktax.Date{}) != "" || date(mktax.Date{Year: 2026, Month: time.May, Day: 1}) != "2026-05-01" {
		t.Error("date")
	}
	if month(mktax.Month{Year: 2026, Month: time.May}) != "2026-05" || month(mktax.Month{}) != "" {
		t.Error("month")
	}
	if ts(time.Time{}) != "" || ts(at("2026-05-01T10:00:00+02:00")) != "2026-05-01T08:00:00Z" {
		t.Error("ts")
	}
}

func TestNewReportShapeAndPortalRows(t *testing.T) {
	trades := []mktax.Trade{
		trade(mktax.Buy, "US0000000001", "2025-01-10T10:00:00Z", "10", "1000", "USD"),
		trade(mktax.Sell, "US0000000001", "2026-05-04T10:00:00Z", "4", "520", "USD"),
		trade(mktax.Buy, "IE0000000002", "2025-03-01T10:00:00Z", "5", "500", "EUR"),
		trade(mktax.Sell, "IE0000000002", "2026-05-12T10:00:00Z", "5", "480", "EUR"),
		trade(mktax.Buy, "GB0000000003", "2022-06-01T10:00:00Z", "2", "100", "GBP"),
		trade(mktax.Sell, "GB0000000003", "2026-05-20T10:00:00Z", "2", "150", "GBP"),
	}
	trades[2].FundUnits, trades[3].FundUnits = true, true
	divs := []mktax.Dividend{{Ticker: "US0000000001_T", ISIN: "US0000000001", Name: "Test", Type: "Dividend", PaidOn: at("2026-05-20T12:00:00Z"), Gross: d("10"), GrossCurrency: "USD", Net: d("8.5"), NetCurrency: "USD"}}
	r := compute(t, trades, divs, "2026-04", "2026-05", flat(map[string]string{"USD": "50", "EUR": "61.5", "GBP": "70"}))
	out := New(r, t212csv.Import{AccountCurrency: "EUR"}, now)
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), ":null") {
		t.Fatalf("null in json: %s", b)
	}
	if out.GeneratedAt != "2026-10-14T12:00:00Z" || out.From != "2026-04" || out.To != "2026-05" || out.Currency != "EUR" || out.FxSource != "NBRSM" || out.FxMethod != "per_leg" {
		t.Fatalf("header = %+v", out)
	}
	if len(out.Months) != 2 || out.Months[0].Month != "2026-04" || out.Months[0].HasEntries || len(out.Months[0].PortalRows) != 0 {
		t.Fatalf("months = %+v", out.Months)
	}
	m := out.Months[1]
	if m.FileBy != "2026-06-10" || m.PayBy != "2026-06-15" || !m.Overdue || m.InProgress {
		t.Fatalf("deadlines = %+v", m)
	}
	if len(m.Sales) != 3 || len(m.Dividends) != 1 || len(m.PortalRows) != 4 {
		t.Fatalf("counts: sales=%d dividends=%d rows=%d", len(m.Sales), len(m.Dividends), len(m.PortalRows))
	}
	gain := m.Sales[0]
	if gain.IncomeType != mktax.IncomeTypeCapitalGains || gain.IncomeSubtype != mktax.IncomeSubtypeSecuritiesGains || gain.Country.Code != "US" || gain.Country.NameMK != "САД" || gain.Quantity != "4" || gain.Proceeds != "520.00" || gain.Cost != "400.00" || gain.TaxableGain != "120.00" || gain.TaxableGainMKD != "6000.00" || gain.TaxMKD != "540.00" || !gain.CostComplete || gain.Ref != "R-2026-05-04T10:00:00Z" {
		t.Fatalf("gain sale = %+v", gain)
	}
	if gain.Lots[0].Acquired != "2025-01-10" || gain.Lots[0].Status != "taxable" || gain.Lots[0].TaxFreeFrom != "2027-01-11" || gain.Lots[0].Cost != "400.00" {
		t.Fatalf("lot = %+v", gain.Lots[0])
	}
	loss := m.Sales[1]
	if loss.IncomeSubtype != mktax.IncomeSubtypeFundUnitsGains || !loss.FundUnits || loss.TaxableGain != "-20.00" || loss.TaxMKD != "0.00" {
		t.Fatalf("loss sale = %+v", loss)
	}
	exempt := m.Sales[2]
	if exempt.ExemptGainMKD != "3500.00" || exempt.TaxableGain != "0.00" {
		t.Fatalf("exempt sale = %+v", exempt)
	}
	rows := m.PortalRows
	if rows[0].Kind != "sale" || rows[0].N != 1 || !rows[0].Enter || rows[0].Gross != "120.00" || rows[0].ForeignTax != "0.00" || rows[0].Currency != "USD" || rows[0].Date != "2026-05-04" || rows[0].TaxMKD != "540.00" || rows[0].Country.Code != "US" {
		t.Fatalf("row 0 = %+v", rows[0])
	}
	if rows[1].Enter || rows[1].SkipReason != "loss" || rows[1].Gross != "-20.00" || rows[1].IncomeSubtype != mktax.IncomeSubtypeFundUnitsGains {
		t.Fatalf("row 1 = %+v", rows[1])
	}
	if rows[2].Kind != "sale" || rows[2].Enter || rows[2].SkipReason != "exempt" {
		t.Fatalf("row 2 = %+v", rows[2])
	}
	dv := rows[3]
	if dv.Kind != "dividend" || dv.N != 4 || !dv.Enter || dv.IncomeType != mktax.IncomeTypeCapital || dv.IncomeSubtype != mktax.IncomeSubtypeDividends || dv.Gross != "10.00" || dv.ForeignTax != "1.50" || dv.Currency != "USD" || dv.GrossMKD != "500.00" || dv.TaxMKD != "50.00" || dv.TaxIfCreditAcceptedMKD != "0.00" {
		t.Fatalf("dividend row = %+v", dv)
	}
	if m.Dividends[0].ForeignTax != "1.50" || m.Dividends[0].ForeignTaxCurrency != "USD" || m.Dividends[0].WithheldMKD != "75.00" {
		t.Fatalf("dividend = %+v", m.Dividends[0])
	}
	if m.Totals.CapitalGainsTaxMKD != "540.00" || m.Totals.TaxableLossesMKD != "-1230.00" || m.Totals.ExemptGainsMKD != "3500.00" || m.Totals.TaxDueIfCreditAcceptedMKD != "540.00" || m.Totals.TaxDueIfCreditRejectedMKD != "590.00" || m.Totals.Sales != 3 || m.Totals.Dividends != 1 {
		t.Fatalf("totals = %+v", m.Totals)
	}
	if len(out.OpenLots) != 1 || out.OpenLots[0].Ticker != "US0000000001_T" || out.OpenLots[0].Quantity != "6" || out.OpenLots[0].TaxFreeQuantityNow != "0" || out.OpenLots[0].Lots[0].TaxFreeFrom != "2027-01-11" {
		t.Fatalf("open lots = %+v", out.OpenLots)
	}
	if len(out.Notes) == 0 || len(out.Warnings) != 0 || len(out.RawWarnings) != 0 {
		t.Fatalf("notes/warnings = %+v", out)
	}
}

func TestWarningsDerived(t *testing.T) {
	trades := []mktax.Trade{
		trade(mktax.Buy, "US0000000001", "2025-01-10T10:00:00Z", "2", "200", "USD"),
		trade(mktax.Sell, "US0000000001", "2026-05-04T10:00:00Z", "5", "600", "EUR"),
	}
	divs := []mktax.Dividend{{Ticker: "X", ISIN: "US0000000009", PaidOn: at("2026-05-20T12:00:00Z"), Net: d("8.5"), NetCurrency: "USD"}}
	r := compute(t, trades, divs, "2026-05", "2026-05", flat(map[string]string{"USD": "50", "EUR": "61.5"}))
	imp := t212csv.Import{Flags: []t212csv.Flag{{Code: "free_shares", File: "2024.csv", Line: 3, ID: "D1", Action: "Deposit"}, {Code: "possible_missing_year", Params: map[string]string{"year": "2025"}}}}
	out := New(r, imp, now)
	m := out.Months[0]
	codes := func(ws []Warning) string {
		var cs []string
		for _, w := range ws {
			cs = append(cs, w.Code)
		}
		return strings.Join(cs, ",")
	}
	if codes(m.Sales[0].Warnings) != "oversell,method_b_cross_currency" {
		t.Fatalf("sale warnings = %+v", m.Sales[0].Warnings)
	}
	if w := m.Sales[0].Warnings[0]; w.Params["quantity"] != "3" || w.Params["ticker"] != "US0000000001_T" || w.Params["date"] != "2026-05-04" {
		t.Fatalf("oversell params = %+v", w.Params)
	}
	if w := m.Sales[0].Warnings[1]; w.Params["lot_currency"] != "USD" || w.Params["sale_currency"] != "EUR" {
		t.Fatalf("cross currency params = %+v", w.Params)
	}
	if codes(m.Dividends[0].Warnings) != "gross_from_net" || codes(m.Warnings) != "oversell,method_b_cross_currency,gross_from_net" {
		t.Fatalf("month warnings = %+v %+v", m.Dividends[0].Warnings, m.Warnings)
	}
	if len(m.RawWarnings) != 3 || len(out.RawWarnings) != 1 || !strings.Contains(out.RawWarnings[0], "exceeds the recorded buys") {
		t.Fatalf("raw = %+v / %+v", m.RawWarnings, out.RawWarnings)
	}
	if m.PortalRows[0].Enter != true || m.PortalRows[0].Gross != "40.00" {
		t.Fatalf("known part still enters: %+v", m.PortalRows[0])
	}
	if codes(out.Import.Flags) != "free_shares,possible_missing_year" || out.Import.Flags[0].Params["file"] != "2024.csv" || out.Import.Flags[0].Params["line"] != "3" || out.Import.Flags[1].Params["year"] != "2025" {
		t.Fatalf("import flags = %+v", out.Import.Flags)
	}
}

func TestImportSummaryFromFixture(t *testing.T) {
	var files []t212csv.File
	for _, n := range []string{"2024.csv", "2025.csv", "2026.csv"} {
		b, err := os.ReadFile(filepath.Join("..", "..", "testdata", "t212", n))
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, t212csv.File{Name: n, Data: b})
	}
	imp, err := t212csv.Parse(files, t212csv.Options{})
	if err != nil {
		t.Fatal(err)
	}
	tbl := rates.Table{}
	for _, dd := range []string{"2026-05-05", "2026-05-20"} {
		tbl.Set(dd, "USD", rates.Rate{Units: 1, Middle: d("50")})
		tbl.Set(dd, "EUR", rates.Rate{Units: 1, Middle: d("61.5")})
	}
	r := compute(t, imp.Trades, imp.Dividends, "2026-05", "2026-05", tbl.RateFunc())
	out := New(r, imp, now)
	s := out.Import
	if s.AccountCurrency != "EUR" || s.Counts.Trades != 10 || s.Counts.Dividends != 4 || len(s.Files) != 3 || s.Files[2].Duplicates != 1 || len(s.Instruments) != 3 || len(s.Interest) != 3 || len(s.Excluded) != 1 || s.Ignored["Deposit"] != 3 || s.Unknown["Mystery action"] != 1 {
		t.Fatalf("summary = %+v", s)
	}
	if s.Interest[0].Amount != "0.42" || s.Interest[0].Currency != "EUR" || s.Interest[0].Date != "2024-07-01" || s.Interest[0].Kind != "interest" {
		t.Fatalf("interest = %+v", s.Interest[0])
	}
	if s.Instruments[1].Key != "IE0000000002" || !s.Instruments[1].FundUnits || !s.Instruments[1].FundUnitsGuessed || s.Instruments[1].Buys != 2 {
		t.Fatalf("instrument = %+v", s.Instruments[1])
	}
	if s.First != "2024-01-10" || s.Last != "2026-05-20" {
		t.Fatalf("range = %s..%s", s.First, s.Last)
	}
}

func TestDividendForeignTaxFromImport(t *testing.T) {
	div := mktax.Dividend{Ticker: "TEST_US_EQ", ISIN: "US0000000001", Name: "Test", Type: "Dividend", PaidOn: at("2026-05-20T12:00:00Z"), Gross: d("10"), GrossCurrency: "USD", Net: d("7.9"), NetCurrency: "EUR"}
	r := compute(t, nil, []mktax.Dividend{div}, "2026-05", "2026-05", flat(map[string]string{"USD": "50", "EUR": "61.5"}))
	imp := t212csv.Import{DividendDetails: map[string]t212csv.DividendDetail{
		t212csv.DividendKey(div): {Ref: "V1", File: "2026.csv", Line: 12, Withholding: d("1.50"), WithholdingCurrency: "USD", GrossEstimated: true},
	}}
	out := New(r, imp, now)
	m := out.Months[0]
	dv := m.Dividends[0]
	if dv.ForeignTax != "1.50" || dv.ForeignTaxCurrency != "USD" || m.PortalRows[0].ForeignTax != "1.50" {
		t.Fatalf("foreign tax = %s %s / row %s", dv.ForeignTax, dv.ForeignTaxCurrency, m.PortalRows[0].ForeignTax)
	}
	if len(dv.Warnings) != 1 || dv.Warnings[0].Code != "gross_estimated" || dv.Warnings[0].Params["line"] != "12" || dv.Warnings[0].Params["file"] != "2026.csv" {
		t.Fatalf("warnings = %+v", dv.Warnings)
	}
	if len(m.Warnings) != 0 {
		t.Fatalf("gross_estimated must not be a month-level alarm: %+v", m.Warnings)
	}
	fallback := New(r, t212csv.Import{}, now).Months[0].Dividends[0]
	if fallback.ForeignTax != "0.28" {
		t.Fatalf("fallback = %s", fallback.ForeignTax)
	}
}
