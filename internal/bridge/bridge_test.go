package bridge

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/emilijan-koteski/mktax/internal/report"
	"github.com/emilijan-koteski/mktax/mktax"
	"github.com/emilijan-koteski/mktax/rates"
	"github.com/emilijan-koteski/mktax/t212csv"
)

var update = flag.Bool("update", false, "rewrite the golden files")

const fixedNow = "2026-10-14T12:00:00Z"

var goldenDir = filepath.Join("..", "..", "testdata", "golden")

func fixtureFiles(t *testing.T) []FileIn {
	t.Helper()
	var files []FileIn
	for _, n := range []string{"2024.csv", "2025.csv", "2026.csv"} {
		b, err := os.ReadFile(filepath.Join("..", "..", "testdata", "t212", n))
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, FileIn{Name: n, Data: string(b)})
	}
	return files
}

func syntheticRates(months []string) rates.Table {
	tbl := rates.Table{}
	for _, m := range months {
		first, _ := time.Parse("2006-01", m)
		for d := first; d.Month() == first.Month(); d = d.AddDate(0, 0, 1) {
			k := rates.DayKey(d)
			tbl.Set(k, "USD", rates.Rate{Units: 1, Middle: decimal.RequireFromString(fmt.Sprintf("50.%02d", d.Day()))})
			tbl.Set(k, "EUR", rates.Rate{Units: 1, Middle: decimal.RequireFromString("61.5")})
			tbl.Set(k, "GBP", rates.Rate{Units: 1, Middle: decimal.RequireFromString("70.25")})
		}
	}
	return tbl
}

func parsed(t *testing.T) ParseOut {
	t.Helper()
	out, perr := Parse(ParseIn{Files: fixtureFiles(t), Now: fixedNow})
	if perr != nil {
		t.Fatalf("parse: %+v", perr)
	}
	return out
}

func calculated(t *testing.T, mutate func(*CalculateIn)) (CalculateOut, *Error) {
	t.Helper()
	p := parsed(t)
	in := CalculateIn{Files: fixtureFiles(t), Rates: syntheticRates(p.RatesNeeded.Months), Now: fixedNow}
	if mutate != nil {
		mutate(&in)
	}
	return Calculate(in)
}

func TestParseFixture(t *testing.T) {
	out := parsed(t)
	if out.Import.Counts.Trades != 10 || out.Import.Counts.Dividends != 4 || len(out.Import.Files) != 3 {
		t.Fatalf("import = %+v", out.Import.Counts)
	}
	wantMonths := "2024-01,2024-03,2024-06,2024-11,2025-02,2025-05,2025-06,2025-09,2026-01,2026-02,2026-03,2026-05"
	if got := strings.Join(out.RatesNeeded.Months, ","); got != wantMonths {
		t.Fatalf("months = %s", got)
	}
	if got := strings.Join(out.RatesNeeded.Currencies, ","); got != "EUR,USD" {
		t.Fatalf("currencies = %s", got)
	}
	if out.SuggestedPeriod.From != "2024-01" || out.SuggestedPeriod.To != "2026-10" {
		t.Fatalf("period = %+v", out.SuggestedPeriod)
	}
}

func TestParseWithoutFiles(t *testing.T) {
	out, perr := Parse(ParseIn{Now: fixedNow})
	if perr != nil || len(out.RatesNeeded.Months) != 0 || out.SuggestedPeriod.From != "2026-10" {
		t.Fatalf("out = %+v err = %+v", out, perr)
	}
}

func TestParseCSVError(t *testing.T) {
	_, perr := Parse(ParseIn{Files: []FileIn{{Name: "x.csv", Data: "Action,Time (UTC)\nMarket buy,2026-01-01 00:00:00+00:00\n"}}})
	if perr == nil || perr.Code != "csv_error" || !strings.Contains(perr.Message, "x.csv") {
		t.Fatalf("err = %+v", perr)
	}
}

func TestCalculateGolden(t *testing.T) {
	p := parsed(t)
	tbl := syntheticRates(p.RatesNeeded.Months)
	out, cerr := Calculate(CalculateIn{Files: fixtureFiles(t), Rates: tbl, Now: fixedNow})
	if cerr != nil {
		t.Fatalf("calculate: %+v", cerr)
	}
	got := indent(t, out)
	if *update {
		for name, v := range map[string]any{"parse.json": p, "rates.json": tbl, "report.json": out} {
			if err := os.WriteFile(filepath.Join(goldenDir, name), indent(t, v), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	want, err := os.ReadFile(filepath.Join(goldenDir, "report.json"))
	if err != nil {
		t.Fatalf("%v (run with -update)", err)
	}
	if string(want) != string(got) {
		t.Fatal("report differs from testdata/golden/report.json (run with -update after reviewing)")
	}
	r := out.Report
	if r.From != "2024-01" || r.To != "2026-10" || len(r.Months) != 34 || !r.Months[33].InProgress || !r.Months[28].Overdue || r.Months[28].Month != "2026-05" {
		t.Fatalf("months = %d from %s to %s", len(r.Months), r.From, r.To)
	}
	jan := monthOf(t, r, "2026-01").Sales[0]
	if len(jan.Lots) != 2 || jan.Lots[0].Status != "exempt_held_over_2y" || jan.Lots[0].Quantity != "6" || jan.Lots[1].Status != "taxable" || jan.Lots[1].Quantity != "2" {
		t.Fatalf("jan lots = %+v", jan.Lots)
	}
	mar := monthOf(t, r, "2026-03")
	if len(mar.Sales) != 1 || mar.Sales[0].CostComplete || mar.Sales[0].Warnings[0].Code != "oversell" || mar.Sales[0].Warnings[0].Params["quantity"] != "6" || len(r.RawWarnings) != 1 {
		t.Fatalf("mar = %+v", mar.Sales[0].Warnings)
	}
	sep := monthOf(t, r, "2025-09")
	if sep.Sales[0].IncomeSubtype != mktax.IncomeSubtypeFundUnitsGains || !sep.PortalRows[0].Enter {
		t.Fatalf("sep = %+v", sep.Sales[0])
	}
	if len(r.OpenLots) != 1 || r.OpenLots[0].Ticker != "TEST_IE_ETF" || r.OpenLots[0].Quantity != "3" {
		t.Fatalf("open lots = %+v", r.OpenLots)
	}
	if out.Import.Counts.Trades != 10 || r.Import.Counts.Trades != 10 {
		t.Fatalf("import summary missing")
	}
}

func monthOf(t *testing.T, r report.Report, m string) report.Month {
	t.Helper()
	for _, x := range r.Months {
		if x.Month == m {
			return x
		}
	}
	t.Fatalf("no month %s", m)
	return report.Month{}
}

func TestCalculateMatchesHandBuiltInputs(t *testing.T) {
	p := parsed(t)
	tbl := syntheticRates(p.RatesNeeded.Months)
	out, cerr := Calculate(CalculateIn{Files: fixtureFiles(t), Rates: tbl, Now: fixedNow})
	if cerr != nil {
		t.Fatal(cerr)
	}
	d := decimal.RequireFromString
	at := func(s string) time.Time {
		v, err := time.Parse("2006-01-02 15:04", s)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	us := func(side mktax.Side, ts, qty, total, fee, ref string) mktax.Trade {
		return mktax.Trade{Ticker: "TEST_US_EQ", ISIN: "US0000000001", Name: "Test US Equity Inc", Time: at(ts), Side: side, Quantity: d(qty), Value: d(total), Currency: "EUR", Fees: d(fee), Ref: ref}
	}
	ie := func(side mktax.Side, ts, qty, total, ref string) mktax.Trade {
		return mktax.Trade{Ticker: "TEST_IE_ETF", ISIN: "IE0000000002", Name: "Test Global UCITS ETF", Time: at(ts), Side: side, Quantity: d(qty), Value: d(total), Currency: "EUR", Fees: d("0"), Ref: ref, FundUnits: true}
	}
	gb := func(side mktax.Side, ts, qty, total, fee, ref string) mktax.Trade {
		return mktax.Trade{Ticker: "TEST_GB_EQ", ISIN: "GB0000000003", Name: "Test British plc", Time: at(ts), Side: side, Quantity: d(qty), Value: d(total), Currency: "EUR", Fees: d(fee), Ref: ref}
	}
	trades := []mktax.Trade{
		us(mktax.Buy, "2024-01-10 14:30", "10", "920.00", "1.38", "T2024-0001"),
		ie(mktax.Buy, "2024-03-15 10:15", "5", "400.00", "T2024-0002"),
		gb(mktax.Buy, "2024-11-02 15:45", "20", "58.00", "0.09", "T2024-0003"),
		us(mktax.Sell, "2025-02-14 15:00", "4", "480.00", "0.72", "T2025-0001"),
		us(mktax.Buy, "2025-05-10 13:20", "6", "660.00", "0.99", "T2025-0002"),
		ie(mktax.Sell, "2025-09-09 09:30", "5", "450.00", "T2025-0003"),
		us(mktax.Sell, "2026-01-15 14:00", "8", "1050.00", "1.58", "T2026-0001"),
		gb(mktax.Sell, "2026-02-10 11:00", "20", "70.00", "0.11", "T2026-0002"),
		us(mktax.Sell, "2026-03-20 16:00", "10", "1380.00", "2.07", "T2026-0003"),
		ie(mktax.Buy, "2026-05-05 10:00", "3", "285.00", "T2026-0004"),
	}
	trades[4].Source = "AutoInvest"
	divs := []mktax.Dividend{
		{Ticker: "TEST_US_EQ", ISIN: "US0000000001", Name: "Test US Equity Inc", Type: "Dividend", PaidOn: at("2024-06-20 12:00"), Gross: d("10"), GrossCurrency: "USD", Net: d("7.90"), NetCurrency: "EUR"},
		{Ticker: "TEST_US_EQ", ISIN: "US0000000001", Name: "Test US Equity Inc", Type: "Ordinary", PaidOn: at("2025-06-20 12:00"), Gross: d("12.6"), GrossCurrency: "USD", Net: d("9.60"), NetCurrency: "EUR"},
		{Ticker: "TEST_GB_EQ", ISIN: "GB0000000003", Name: "Test British plc", Type: "Dividend", PaidOn: at("2026-03-03 12:00"), Gross: d("1.15"), GrossCurrency: "EUR", Net: d("1.15"), NetCurrency: "EUR"},
		{Ticker: "TEST_US_EQ", ISIN: "US0000000001", Name: "Test US Equity Inc", Type: "Dividends paid by us corporations", PaidOn: at("2026-05-20 12:00"), Gross: d("4.6"), GrossCurrency: "USD", Net: d("3.40"), NetCurrency: "EUR"},
	}
	from, _ := mktax.ParseMonth("2024-01")
	to, _ := mktax.ParseMonth("2026-10")
	now, _ := time.Parse(time.RFC3339, fixedNow)
	r, err := mktax.Compute(trades, divs, tbl.RateFunc(), mktax.Options{From: from, To: to, Now: now})
	if err != nil {
		t.Fatal(err)
	}
	var files []t212csv.File
	for _, f := range fixtureFiles(t) {
		files = append(files, t212csv.File{Name: f.Name, Data: []byte(f.Data)})
	}
	imp, err := t212csv.Parse(files, t212csv.Options{})
	if err != nil {
		t.Fatal(err)
	}
	want := report.New(r, t212csv.Import{AccountCurrency: "EUR", DividendDetails: imp.DividendDetails}, now)
	if !reflect.DeepEqual(want.Months, out.Report.Months) || !reflect.DeepEqual(want.OpenLots, out.Report.OpenLots) || !reflect.DeepEqual(want.RawWarnings, out.Report.RawWarnings) {
		t.Fatal("the CSV path and hand-built engine inputs produce different reports")
	}
}

func TestCalculateMissingRates(t *testing.T) {
	_, cerr := calculated(t, func(in *CalculateIn) { delete(in.Rates, "2026-01-15") })
	if cerr == nil || cerr.Code != "rates_missing" || len(cerr.Missing) != 1 || cerr.Missing[0].Currency != "EUR" || cerr.Missing[0].Date != "2026-01-15" {
		t.Fatalf("err = %+v", cerr)
	}
	_, cerr = calculated(t, func(in *CalculateIn) { in.Rates = nil; in.From, in.To = "2026-03", "2026-03" })
	if cerr == nil || cerr.Code != "rates_missing" || len(cerr.Missing) != 3 {
		t.Fatalf("err = %+v", cerr)
	}
}

func TestCalculatePeriodSubset(t *testing.T) {
	out, cerr := calculated(t, func(in *CalculateIn) { in.From, in.To = "2026-01", "2026-03" })
	if cerr != nil || len(out.Report.Months) != 3 || out.Report.From != "2026-01" {
		t.Fatalf("out = %+v err = %+v", out.Report.From, cerr)
	}
}

func TestCalculateValidation(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*CalculateIn)
		code string
	}{
		{"from after to", func(in *CalculateIn) { in.From, in.To = "2026-05", "2026-01" }, "invalid_period"},
		{"before the regime", func(in *CalculateIn) { in.From, in.To = "2022-12", "2026-01" }, "invalid_period"},
		{"to in the future", func(in *CalculateIn) { in.From, in.To = "2026-01", "2026-11" }, "invalid_period"},
		{"bad month", func(in *CalculateIn) { in.From = "2026-1" }, "invalid_period"},
		{"too long", func(in *CalculateIn) { in.Now = "2040-01-01T00:00:00Z"; in.From, in.To = "2023-01", "2040-01" }, "invalid_period"},
		{"bad now", func(in *CalculateIn) { in.Now = "yesterday" }, "bad_input"},
		{"csv error", func(in *CalculateIn) { in.Files = []FileIn{{Name: "x.csv", Data: "Action\nMarket buy\n"}} }, "csv_error"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, cerr := calculated(t, c.mut)
			if cerr == nil || cerr.Code != c.code {
				t.Fatalf("err = %+v, want %s", cerr, c.code)
			}
		})
	}
}

func TestEnvelopeNeverPanics(t *testing.T) {
	for _, in := range []string{"", "garbage", "[]", `{"files":5}`, `{"files":[{"name":"x.csv","data":"\u0000"}]}`} {
		for _, fn := range []func([]byte) []byte{ParseJSON, CalculateJSON} {
			var env struct {
				OK    bool            `json:"ok"`
				Error *Error          `json:"error"`
				Raw   json.RawMessage `json:"result"`
			}
			if err := json.Unmarshal(fn([]byte(in)), &env); err != nil {
				t.Fatalf("%q: invalid envelope: %v", in, err)
			}
			if !env.OK && env.Error == nil {
				t.Fatalf("%q: error envelope without error", in)
			}
		}
	}
	out := CalculateJSON([]byte(`{"now":"2026-10-14T12:00:00Z"}`))
	if !strings.HasPrefix(string(out), `{"ok":true,"result":`) {
		t.Fatalf("empty calculation = %s", out)
	}
}

func TestFundUnitsOverrideChangesSubtype(t *testing.T) {
	out, cerr := calculated(t, func(in *CalculateIn) { in.FundUnits = map[string]bool{"IE0000000002": false} })
	if cerr != nil {
		t.Fatal(cerr)
	}
	if got := monthOf(t, out.Report, "2025-09").Sales[0].IncomeSubtype; got != mktax.IncomeSubtypeSecuritiesGains {
		t.Fatalf("subtype = %s", got)
	}
	if ins := out.Import.Instruments[1]; ins.FundUnits || ins.FundUnitsGuessed {
		t.Fatalf("instrument = %+v", ins)
	}
}

func TestVersion(t *testing.T) {
	if Version() == "" {
		t.Fatal("empty version")
	}
}

func indent(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return append(b, '\n')
}
