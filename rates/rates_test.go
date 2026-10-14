package rates

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/emilijan-koteski/mktax/mktax"
)

func d(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func day(s string) time.Time {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		panic(err)
	}
	return t
}

func sample() Table {
	t := Table{}
	t.Set("2026-05-01", "USD", Rate{Units: 1, Middle: d("50.1234")})
	t.Set("2026-05-01", "JPY", Rate{Units: 100, Middle: d("35.25")})
	t.Set("2026-05-02", "USD", Rate{Units: 1, Middle: d("50.2")})
	return t
}

func TestPerUnit(t *testing.T) {
	if got := (Rate{Units: 1, Middle: d("50.1234")}).PerUnit(); got.String() != "50.1234" {
		t.Errorf("units 1: %s", got)
	}
	if got := (Rate{Units: 100, Middle: d("35.25")}).PerUnit(); got.String() != "0.3525" {
		t.Errorf("units 100: %s", got)
	}
}

func TestLookup(t *testing.T) {
	tbl := sample()
	cases := []struct {
		ccy, day, want string
		err            error
		msg            string
	}{
		{"USD", "2026-05-01", "50.1234", nil, ""},
		{"usd", "2026-05-02", "50.2", nil, ""},
		{" jpy ", "2026-05-01", "0.3525", nil, ""},
		{"MKD", "1999-01-01", "1", nil, ""},
		{"mkd", "2026-05-01", "1", nil, ""},
		{"USD", "2026-05-03", "", ErrDayMissing, "no NBRSM rate list for 2026-05-03 (not published yet, or not prefetched)"},
		{"GBP", "2026-05-01", "", ErrCurrencyMissing, "NBRSM publishes no GBP rate on 2026-05-01"},
	}
	for _, c := range cases {
		got, err := tbl.Lookup(c.ccy, day(c.day))
		if c.err != nil {
			if !errors.Is(err, c.err) || err.Error() != c.msg {
				t.Errorf("Lookup(%q, %s): err = %v, want %q", c.ccy, c.day, err, c.msg)
			}
			continue
		}
		if err != nil || got.String() != c.want {
			t.Errorf("Lookup(%q, %s) = %s, %v; want %s", c.ccy, c.day, got, err, c.want)
		}
	}
}

func TestRateFuncUsesCivilDate(t *testing.T) {
	f := sample().RateFunc()
	got, err := f("USD", mktax.Date{Year: 2026, Month: time.May, Day: 2})
	if err != nil || got.String() != "50.2" {
		t.Fatalf("got %s, %v", got, err)
	}
	if _, err := f("USD", mktax.Date{Year: 2026, Month: time.May, Day: 9}); !errors.Is(err, ErrDayMissing) {
		t.Fatalf("err = %v", err)
	}
}

func TestDayKeyIgnoresZone(t *testing.T) {
	loc := time.FixedZone("x", 2*3600)
	if got := DayKey(time.Date(2026, time.May, 1, 0, 30, 0, 0, loc)); got != "2026-05-01" {
		t.Fatalf("got %s", got)
	}
}

func TestJSONRoundTrip(t *testing.T) {
	b, err := json.Marshal(sample())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"USD":{"units":1,"middle":"50.1234"}`) {
		t.Fatalf("json = %s", b)
	}
	var back Table
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if got, _ := back.Lookup("JPY", day("2026-05-01")); got.String() != "0.3525" {
		t.Fatalf("round trip = %s", got)
	}
	var loose Table
	if err := json.Unmarshal([]byte(`{"2026-05-01":{"USD":{"currency":"USD","units":1,"middle":50.5}}}`), &loose); err != nil {
		t.Fatal(err)
	}
	if got, _ := loose.Lookup("USD", day("2026-05-01")); got.String() != "50.5" {
		t.Fatalf("loose = %s", got)
	}
}

func TestFilterMergeMissingDays(t *testing.T) {
	tbl := sample()
	f := tbl.Filter([]string{"usd"})
	if len(f["2026-05-01"]) != 1 || len(f["2026-05-02"]) != 1 {
		t.Fatalf("filter = %v", f)
	}
	if _, ok := f["2026-05-01"]["JPY"]; ok {
		t.Fatal("JPY should be filtered out")
	}
	all := tbl.Filter(nil)
	all.Set("2026-05-09", "USD", Rate{Units: 1, Middle: d("1")})
	if _, ok := tbl["2026-05-09"]; ok {
		t.Fatal("Filter(nil) must copy")
	}
	o := Table{}
	o.Set("2026-05-02", "USD", Rate{Units: 1, Middle: d("51")})
	o.Set("2026-05-03", "EUR", Rate{Units: 1, Middle: d("61.5")})
	tbl.Merge(o)
	if got, _ := tbl.Lookup("USD", day("2026-05-02")); got.String() != "51" {
		t.Fatalf("merge should overwrite, got %s", got)
	}
	if got := tbl.Days(); strings.Join(got, ",") != "2026-05-01,2026-05-02,2026-05-03" {
		t.Fatalf("days = %v", got)
	}
	if got := tbl.Missing(day("2026-04-30"), day("2026-05-05")); strings.Join(got, ",") != "2026-04-30,2026-05-04,2026-05-05" {
		t.Fatalf("missing = %v", got)
	}
}
