package t212csv

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/emilijan-koteski/mktax/mktax"
)

const headerA = "Action,Time (UTC),ISIN,Ticker,Name,Notes,ID,No. of shares,Price / share,Currency (Price / share),Exchange rate,Result,Currency (Result),Total,Currency (Total),Withholding tax,Currency (Withholding tax),Currency conversion from amount,Currency (Currency conversion from amount),Currency conversion to amount,Currency (Currency conversion to amount),Currency conversion fee,Currency (Currency conversion fee)"

const headerB = "Action,Time (UTC),ISIN,Ticker,Name,Notes,ID,No. of shares,Price / share,Currency (Price / share),Exchange rate,Result,Currency (Result),Total,Currency (Total),Withholding tax,Currency (Withholding tax),Charge amount,Currency (Charge amount),Deposit fee,Currency (Deposit fee),Currency conversion fee,Currency (Currency conversion fee)"

func csvA(rows ...string) File {
	return File{Name: "a.csv", Data: []byte(headerA + "\n" + strings.Join(rows, "\n") + "\n")}
}

func csvB(rows ...string) File {
	return File{Name: "b.csv", Data: []byte(headerB + "\n" + strings.Join(rows, "\n") + "\n")}
}

func buyRow(id, ts, isin, ticker, name, qty, price, pccy, total, tccy, fee, feeCcy string) string {
	return strings.Join([]string{"Market buy", ts, isin, ticker, name, "", id, qty, price, pccy, "1.0", "", "", total, tccy, "", "", "", "", "", "", fee, feeCcy}, ",")
}

func divRow(action, id, ts, isin, ticker, name, qty, price, pccy, total, tccy, wht, whtCcy string) string {
	return strings.Join([]string{action, ts, isin, ticker, name, "", id, qty, price, pccy, "1.0", "", "", total, tccy, wht, whtCcy, "", "", "", "", "", ""}, ",")
}

func mustParse(t *testing.T, files ...File) Import {
	t.Helper()
	imp, err := Parse(files, Options{})
	if err != nil {
		t.Fatal(err)
	}
	return imp
}

func TestParseHeaderVariants(t *testing.T) {
	a := mustParse(t, csvA(buyRow("1", "2026-05-28 13:39:45+00:00", "US0000000001", "TEST_US_EQ", "Test US Equity Inc", "2", "100.00", "USD", "184.00", "EUR", "0.28", "EUR")))
	b := mustParse(t, csvB(strings.Join([]string{"Market buy", "2026-05-28 13:39:45+00:00", "US0000000001", "TEST_US_EQ", "Test US Equity Inc", "", "1", "2", "100.00", "USD", "1.0", "", "", "184.00", "EUR", "", "", "", "", "", "", "0.28", "EUR"}, ",")))
	if len(a.Trades) != 1 || len(b.Trades) != 1 || !reflect.DeepEqual(a.Trades, b.Trades) {
		t.Fatalf("a = %+v\nb = %+v", a.Trades, b.Trades)
	}
	tr := a.Trades[0]
	if tr.Side != mktax.Buy || tr.Quantity.String() != "2" || tr.Value.String() != "184" || tr.Currency != "EUR" || tr.Fees.String() != "0.28" || tr.Ref != "1" || tr.ISIN != "US0000000001" || tr.Ticker != "TEST_US_EQ" {
		t.Fatalf("trade = %+v", tr)
	}
	if tr.Time.UTC().Format(time.RFC3339) != "2026-05-28T13:39:45Z" {
		t.Fatalf("time = %v", tr.Time)
	}
}

func TestParseStripsBOMAndIgnoresUnknownColumns(t *testing.T) {
	data := "\uFEFF" + headerA + ",Mystery column\n" + buyRow("1", "2026-05-28 13:39:45+00:00", "US0000000001", "TEST_US_EQ", "N", "1", "1", "USD", "1", "EUR", "", "") + ",zzz\n"
	imp := mustParse(t, File{Name: "bom.csv", Data: []byte(data)})
	if len(imp.Trades) != 1 || len(imp.Files[0].Columns) != 24 {
		t.Fatalf("imp = %+v", imp.Files)
	}
}

func TestParseMissingRequiredHeader(t *testing.T) {
	_, err := Parse([]File{{Name: "x.csv", Data: []byte("Action,Time (UTC),ISIN\nMarket buy,2026-01-01 00:00:00+00:00,US0000000001\n")}}, Options{})
	if err == nil || !strings.Contains(err.Error(), "x.csv") || !strings.Contains(err.Error(), "No. of shares") {
		t.Fatalf("err = %v", err)
	}
}

func TestParseBadNumberNamesRow(t *testing.T) {
	_, err := Parse([]File{csvA(buyRow("1", "2026-05-28 13:39:45+00:00", "US0000000001", "T", "N", "two", "1", "USD", "1", "EUR", "", ""))}, Options{})
	if err == nil || !strings.Contains(err.Error(), "a.csv:2") || !strings.Contains(err.Error(), "No. of shares") {
		t.Fatalf("err = %v", err)
	}
}

func TestClassifyAction(t *testing.T) {
	cases := []struct {
		action string
		kind   Kind
		side   mktax.Side
		typ    string
	}{
		{"Market buy", KindTrade, mktax.Buy, ""},
		{"Limit buy", KindTrade, mktax.Buy, ""},
		{"Stop sell", KindTrade, mktax.Sell, ""},
		{"Stop limit sell", KindTrade, mktax.Sell, ""},
		{" market SELL ", KindTrade, mktax.Sell, ""},
		{"Dividend (Dividend)", KindDividend, "", "Dividend"},
		{"Dividend (Ordinary)", KindDividend, "", "Ordinary"},
		{"Dividend (Dividends paid by us corporations)", KindDividend, "", "Dividends paid by us corporations"},
		{"Dividend (Dividends paid by foreign corporations)", KindDividend, "", "Dividends paid by foreign corporations"},
		{"Dividend", KindDividend, "", "Dividend"},
		{"Dividend (Return of capital)", KindDividendExcluded, "", "Return of capital"},
		{"Dividend (Property income)", KindDividendExcluded, "", "Property income"},
		{"Interest on cash", KindInterest, "", ""},
		{"Lending interest", KindInterest, "", ""},
		{"Deposit", KindIgnored, "", ""},
		{"Withdrawal", KindIgnored, "", ""},
		{"Currency conversion", KindIgnored, "", ""},
		{"Card debit", KindIgnored, "", ""},
		{"Spending cashback", KindIgnored, "", ""},
		{"Stock split", KindCorporateAction, "", ""},
		{"Spinoff", KindCorporateAction, "", ""},
		{"Mystery action", KindUnknown, "", ""},
	}
	for _, c := range cases {
		kind, side, typ := ClassifyAction(c.action)
		if kind != c.kind || side != c.side || typ != c.typ {
			t.Errorf("%q: got (%s, %q, %q), want (%s, %q, %q)", c.action, kind, side, typ, c.kind, c.side, c.typ)
		}
	}
}

func TestTimeLayouts(t *testing.T) {
	for _, ts := range []string{"2026-05-28 13:39:45+00:00", "2026-05-28 13:39:45", "2026-05-28T13:39:45Z", "2026-05-28T15:39:45+02:00"} {
		imp := mustParse(t, csvA(buyRow("1", ts, "US0000000001", "T", "N", "1", "1", "USD", "1", "EUR", "", "")))
		if got := imp.Trades[0].Time.UTC().Format(time.RFC3339); got != "2026-05-28T13:39:45Z" {
			t.Errorf("%q → %s", ts, got)
		}
	}
	imp := mustParse(t, csvA(buyRow("1", "2026-05-28", "US0000000001", "T", "N", "1", "1", "USD", "1", "EUR", "", "")))
	if got := imp.Trades[0].Time.UTC().Format(time.RFC3339); got != "2026-05-28T00:00:00Z" {
		t.Errorf("date only → %s", got)
	}
	if _, err := Parse([]File{csvA(buyRow("1", "28.05.2026", "US0000000001", "T", "N", "1", "1", "USD", "1", "EUR", "", ""))}, Options{}); err == nil {
		t.Error("bad time should fail")
	}
}

func TestDividendGrossRule(t *testing.T) {
	cases := []struct {
		name                     string
		price, pccy, total, tccy string
		wht, whtCcy              string
		gross, grossCcy, net     string
		flag                     string
	}{
		{"no withholding", "0.50", "USD", "4.60", "EUR", "", "", "4.6", "EUR", "4.6", ""},
		{"zero withholding", "0.50", "USD", "4.60", "EUR", "0", "USD", "4.6", "EUR", "4.6", ""},
		{"withholding in account currency", "0.50", "EUR", "5.00", "EUR", "1.50", "EUR", "6.5", "EUR", "5", ""},
		{"withholding in instrument currency", "0.85", "USD", "7.90", "EUR", "1.50", "USD", "10", "USD", "7.9", "gross_estimated"},
		{"withholding in a third currency", "0.85", "USD", "7.90", "EUR", "1.50", "CHF", "0", "", "7.9", "gross_unknown"},
		{"gbx price and withholding", "5.00", "GBX", "1.15", "EUR", "20.00", "GBX", "0.7", "GBP", "1.15", "gross_estimated"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			imp := mustParse(t, csvA(divRow("Dividend (Dividend)", "V1", "2026-06-20 12:00:00+00:00", "US0000000001", "T", "N", "10", c.price, c.pccy, c.total, c.tccy, c.wht, c.whtCcy)))
			if len(imp.Dividends) != 1 {
				t.Fatalf("dividends = %+v", imp.Dividends)
			}
			d := imp.Dividends[0]
			if d.Gross.String() != c.gross || d.GrossCurrency != c.grossCcy || d.Net.String() != c.net || d.NetCurrency != c.tccy || d.Type != "Dividend" {
				t.Fatalf("dividend = %+v", d)
			}
			got := ""
			if len(imp.Flags) > 0 {
				got = imp.Flags[0].Code
			}
			if got != c.flag {
				t.Fatalf("flag = %q, want %q (%+v)", got, c.flag, imp.Flags)
			}
		})
	}
}

func TestExcludedDividendListedNotImported(t *testing.T) {
	imp := mustParse(t, csvA(divRow("Dividend (Return of capital)", "V1", "2026-06-20 12:00:00+00:00", "IE0000000002", "T", "N", "5", "0.20", "EUR", "1.00", "EUR", "", "")))
	if len(imp.Dividends) != 0 || len(imp.Excluded) != 1 || imp.Excluded[0].Kind != KindDividendExcluded || imp.Excluded[0].Amount.String() != "1" || imp.Flags[0].Code != "excluded_dividend_type" {
		t.Fatalf("imp = %+v", imp)
	}
}

func TestInterestIgnoredUnknownAndCorporateActions(t *testing.T) {
	rows := []string{
		"Interest on cash,2026-07-01 06:00:00+00:00,,,,,I1,,,,,,,0.42,EUR,,,,,,,,",
		"Lending interest,2026-07-02 06:00:00+00:00,,,,,I2,,,,,,,0.10,EUR,,,,,,,,",
		"Deposit,2026-07-03 06:00:00+00:00,,,,,D1,,,,,,,100.00,EUR,,,,,,,,",
		"Deposit,2026-07-04 06:00:00+00:00,,,,Free Shares Promotion,D2,,,,,,,0.00,EUR,,,,,,,,",
		"Withdrawal,2026-07-05 06:00:00+00:00,,,,,W1,,,,,,,50.00,EUR,,,,,,,,",
		"Currency conversion,2026-07-06 06:00:00+00:00,,,,,C1,,,,,,,20.00,EUR,,,,,,,,",
		"Stock split,2026-07-07 06:00:00+00:00,US0000000001,T,N,,S1,,,,,,,,,,,,,,,,",
		"Mystery action,2026-07-08 06:00:00+00:00,,,,,M1,,,,,,,0.00,EUR,,,,,,,,",
	}
	imp := mustParse(t, csvA(rows...))
	if len(imp.Trades)+len(imp.Dividends) != 0 {
		t.Fatalf("nothing should be income: %+v %+v", imp.Trades, imp.Dividends)
	}
	if len(imp.Interest) != 2 || imp.Interest[0].Amount.String() != "0.42" || imp.Interest[1].Action != "Lending interest" {
		t.Fatalf("interest = %+v", imp.Interest)
	}
	if imp.Ignored["Deposit"] != 2 || imp.Ignored["Withdrawal"] != 1 || imp.Ignored["Currency conversion"] != 1 || imp.Unknown["Mystery action"] != 1 {
		t.Fatalf("ignored = %v unknown = %v", imp.Ignored, imp.Unknown)
	}
	codes := map[string]int{}
	for _, f := range imp.Flags {
		codes[f.Code]++
	}
	if codes["free_shares"] != 1 || codes["corporate_action"] != 1 || codes["unrecognised_action"] != 1 || len(imp.Flags) != 3 {
		t.Fatalf("flags = %+v", imp.Flags)
	}
}

func TestZeroQuantityTradeSkipped(t *testing.T) {
	imp := mustParse(t, csvA(buyRow("1", "2026-05-28 13:39:45+00:00", "US0000000001", "T", "N", "0", "1", "USD", "0", "EUR", "", "")))
	if len(imp.Trades) != 0 || len(imp.Flags) != 1 || imp.Flags[0].Code != "zero_quantity" {
		t.Fatalf("imp = %+v", imp)
	}
}

func TestFeesOnlyInAccountCurrency(t *testing.T) {
	imp := mustParse(t, csvA(buyRow("1", "2026-05-28 13:39:45+00:00", "US0000000001", "T", "N", "1", "1", "USD", "1", "EUR", "0.30", "USD")))
	if imp.Trades[0].Fees.String() != "0" || len(imp.Flags) != 1 || imp.Flags[0].Code != "fee_other_currency" {
		t.Fatalf("imp = %+v flags = %+v", imp.Trades[0], imp.Flags)
	}
}

func TestDedupeAcrossFiles(t *testing.T) {
	row := buyRow("T1", "2025-12-31 10:00:00+00:00", "US0000000001", "T", "N", "1", "1", "USD", "1", "EUR", "", "")
	a := File{Name: "2025.csv", Data: []byte(headerA + "\n" + row + "\n")}
	b := File{Name: "2026.csv", Data: []byte(headerA + "\n" + row + "\n")}
	imp := mustParse(t, a, b)
	if len(imp.Trades) != 1 || imp.Files[1].Duplicates != 1 || imp.Files[0].Duplicates != 0 || len(imp.Flags) != 0 {
		t.Fatalf("imp = %+v", imp)
	}
	noID := buyRow("", "2025-12-31 10:00:00+00:00", "US0000000001", "T", "N", "1", "1", "USD", "1", "EUR", "", "")
	imp = mustParse(t, File{Name: "x.csv", Data: []byte(headerA + "\n" + noID + "\n" + noID + "\n")})
	if len(imp.Trades) != 1 || len(imp.Flags) != 1 || imp.Flags[0].Code != "duplicate_row" {
		t.Fatalf("composite dedupe: %+v %+v", imp.Trades, imp.Flags)
	}
}

func TestFundUnitsHeuristicAndOverride(t *testing.T) {
	for name, want := range map[string]bool{"Test Global UCITS ETF": true, "iShares Core S&P 500 (Acc)": false, "Something ETC": true, "Vanguard Etf Dist": true, "Betfair plc": false, "Test US Equity Inc": false} {
		if IsFundUnits(name) != want {
			t.Errorf("IsFundUnits(%q) = %v", name, !want)
		}
	}
	etf := buyRow("1", "2026-05-28 13:39:45+00:00", "IE0000000002", "TEST_IE_ETF", "Test Global UCITS ETF", "1", "1", "EUR", "1", "EUR", "", "")
	eq := buyRow("2", "2026-05-29 13:39:45+00:00", "US0000000001", "TEST_US_EQ", "Test US Equity Inc", "1", "1", "USD", "1", "EUR", "", "")
	imp := mustParse(t, csvA(etf, eq))
	if !imp.Trades[0].FundUnits || imp.Trades[1].FundUnits || len(imp.Instruments) != 2 || !imp.Instruments[0].FundUnitsGuessed {
		t.Fatalf("heuristic: %+v %+v", imp.Trades, imp.Instruments)
	}
	imp, err := Parse([]File{csvA(etf, eq)}, Options{FundUnits: map[string]bool{"IE0000000002": false, "US0000000001": true}})
	if err != nil {
		t.Fatal(err)
	}
	if imp.Trades[0].FundUnits || !imp.Trades[1].FundUnits || imp.Instruments[0].FundUnitsGuessed || imp.Instruments[1].FundUnitsGuessed {
		t.Fatalf("override: %+v %+v", imp.Trades, imp.Instruments)
	}
	if ins := imp.Instruments[0]; ins.Key != "IE0000000002" || ins.Ticker != "TEST_IE_ETF" || ins.Currency != "EUR" || ins.Buys != 1 {
		t.Fatalf("instrument = %+v", ins)
	}
}

func TestAccountCurrencyRangeAndGap(t *testing.T) {
	a := File{Name: "2024.csv", Data: []byte(headerA + "\n" + buyRow("1", "2024-03-01 10:00:00+00:00", "US0000000001", "T", "N", "1", "1", "USD", "1", "EUR", "", "") + "\n")}
	c := File{Name: "2026.csv", Data: []byte(headerA + "\n" + buyRow("2", "2026-03-01 10:00:00+00:00", "US0000000001", "T", "N", "1", "1", "USD", "1", "EUR", "", "") + "\n")}
	imp := mustParse(t, c, a)
	if imp.AccountCurrency != "EUR" || imp.First.Year() != 2024 || imp.Last.Year() != 2026 {
		t.Fatalf("imp = %+v", imp)
	}
	if len(imp.Flags) != 1 || imp.Flags[0].Code != "possible_missing_year" || imp.Flags[0].Params["year"] != "2025" {
		t.Fatalf("flags = %+v", imp.Flags)
	}
}

func TestDeterministicOrdering(t *testing.T) {
	files := fixtureFiles(t)
	forward := mustParse(t, files...)
	backward := mustParse(t, files[2], files[1], files[0])
	if !reflect.DeepEqual(forward.Trades, backward.Trades) || !reflect.DeepEqual(forward.Dividends, backward.Dividends) || !reflect.DeepEqual(forward.Instruments, backward.Instruments) {
		t.Fatal("file order changed the import")
	}
	for i := 1; i < len(forward.Trades); i++ {
		if forward.Trades[i].Time.Before(forward.Trades[i-1].Time) {
			t.Fatal("trades are not chronological")
		}
	}
}

func fixtureFiles(t *testing.T) []File {
	t.Helper()
	var files []File
	for _, n := range []string{"2024.csv", "2025.csv", "2026.csv"} {
		b, err := os.ReadFile(filepath.Join("..", "testdata", "t212", n))
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, File{Name: n, Data: b})
	}
	return files
}

func TestFixtureFilesParse(t *testing.T) {
	imp := mustParse(t, fixtureFiles(t)...)
	if len(imp.Trades) != 10 || len(imp.Dividends) != 4 || len(imp.Excluded) != 1 || len(imp.Interest) != 3 {
		t.Fatalf("trades=%d dividends=%d excluded=%d interest=%d", len(imp.Trades), len(imp.Dividends), len(imp.Excluded), len(imp.Interest))
	}
	if imp.Ignored["Deposit"] != 3 || imp.Ignored["Withdrawal"] != 1 || imp.Ignored["Currency conversion"] != 1 || imp.Unknown["Mystery action"] != 1 {
		t.Fatalf("ignored = %v unknown = %v", imp.Ignored, imp.Unknown)
	}
	codes := map[string]int{}
	for _, f := range imp.Flags {
		codes[f.Code]++
	}
	want := map[string]int{"free_shares": 1, "gross_estimated": 3, "corporate_action": 1, "unrecognised_action": 1, "excluded_dividend_type": 1}
	if !reflect.DeepEqual(codes, want) {
		t.Fatalf("flags = %v", codes)
	}
	if imp.Files[2].Duplicates != 1 || imp.Files[2].Rows != 11 || imp.AccountCurrency != "EUR" {
		t.Fatalf("files = %+v currency = %s", imp.Files, imp.AccountCurrency)
	}
	if len(imp.Instruments) != 3 || imp.Instruments[0].Key != "GB0000000003" || imp.Instruments[0].Currency != "GBP" || !imp.Instruments[1].FundUnits {
		t.Fatalf("instruments = %+v", imp.Instruments)
	}
	sells := 0
	for _, tr := range imp.Trades {
		if tr.Side == mktax.Sell {
			sells++
		}
	}
	if sells != 5 {
		t.Fatalf("sells = %d", sells)
	}
}
