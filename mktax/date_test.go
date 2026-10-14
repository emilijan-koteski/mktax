package mktax

import (
	"testing"
	"time"
)

func TestDateOf(t *testing.T) {
	tests := []struct {
		name string
		in   time.Time
		want Date
	}{
		{"summer UTC+2 crosses midnight", time.Date(2026, 5, 31, 22, 30, 0, 0, time.UTC), Date{2026, time.June, 1}},
		{"summer just before local midnight", time.Date(2026, 5, 31, 21, 59, 0, 0, time.UTC), Date{2026, time.May, 31}},
		{"winter UTC+1 crosses midnight", time.Date(2026, 1, 31, 23, 0, 0, 0, time.UTC), Date{2026, time.February, 1}},
		{"winter before local midnight", time.Date(2026, 1, 31, 22, 59, 0, 0, time.UTC), Date{2026, time.January, 31}},
		{"year end", time.Date(2025, 12, 31, 23, 30, 0, 0, time.UTC), Date{2026, time.January, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DateOf(tt.in); got != tt.want {
				t.Errorf("DateOf(%s) = %s, want %s", tt.in, got, tt.want)
			}
		})
	}
}

func TestTaxFreeFrom(t *testing.T) {
	tests := []struct {
		acquired, want Date
	}{
		{Date{2024, time.May, 10}, Date{2026, time.May, 11}},
		{Date{2024, time.February, 29}, Date{2026, time.March, 2}},
		{Date{2024, time.December, 31}, Date{2027, time.January, 1}},
	}
	for _, tt := range tests {
		if got := TaxFreeFrom(tt.acquired); got != tt.want {
			t.Errorf("TaxFreeFrom(%s) = %s, want %s", tt.acquired, got, tt.want)
		}
	}
}

func TestParseMonth(t *testing.T) {
	m, err := ParseMonth("2026-05")
	if err != nil || m != (Month{2026, time.May}) || m.String() != "2026-05" {
		t.Fatalf("ParseMonth = %v, %v", m, err)
	}
	if m.First() != (Date{2026, time.May, 1}) || m.Last() != (Date{2026, time.May, 31}) {
		t.Errorf("first/last = %s %s", m.First(), m.Last())
	}
	if n := (Month{2026, time.December}).AddMonths(1); n != (Month{2027, time.January}) {
		t.Errorf("AddMonths = %s", n)
	}
	if p := (Month{2026, time.January}).AddMonths(-1); p != (Month{2025, time.December}) {
		t.Errorf("AddMonths(-1) = %s", p)
	}
	for _, bad := range []string{"", "2026", "2026-13", "26-05", "2026-5-1"} {
		if _, err := ParseMonth(bad); err == nil {
			t.Errorf("ParseMonth(%q) should fail", bad)
		}
	}
}

func TestCountryFromISIN(t *testing.T) {
	tests := []struct {
		isin, code, mk, en string
	}{
		{"US0000000001", "US", "САД", "United States"},
		{"IE0000000002", "IE", "Ирска", "Ireland"},
		{"GB0000000003", "GB", "Велика Британија", "United Kingdom"},
		{"ZZ0000000004", "ZZ", "ZZ", "ZZ"},
		{"", "", "", ""},
	}
	for _, tt := range tests {
		c := CountryFromISIN(tt.isin)
		if c.Code != tt.code || c.NameMK != tt.mk || c.NameEN != tt.en {
			t.Errorf("CountryFromISIN(%q) = %+v", tt.isin, c)
		}
	}
}

func TestCapitalGainSubtype(t *testing.T) {
	if got := CapitalGainSubtype(false); got != IncomeSubtypeSecuritiesGains || got[:5] != "S8.1 " {
		t.Errorf("shares = %q", got)
	}
	if got := CapitalGainSubtype(true); got != IncomeSubtypeFundUnitsGains || got[:6] != "S.8.7 " {
		t.Errorf("fund units = %q", got)
	}
}
