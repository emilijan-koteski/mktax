package mktax

import "strings"

type Country struct {
	Code   string
	NameMK string
	NameEN string
}

var countries = map[string][2]string{
	"US": {"САД", "United States"},
	"IE": {"Ирска", "Ireland"},
	"GB": {"Велика Британија", "United Kingdom"},
	"DE": {"Германија", "Germany"},
	"FR": {"Франција", "France"},
	"NL": {"Холандија", "Netherlands"},
	"LU": {"Луксембург", "Luxembourg"},
	"CA": {"Канада", "Canada"},
	"CH": {"Швајцарија", "Switzerland"},
}

func CountryFromISIN(isin string) Country {
	isin = strings.ToUpper(strings.TrimSpace(isin))
	if len(isin) < 2 {
		return Country{}
	}
	code := isin[:2]
	if n, ok := countries[code]; ok {
		return Country{Code: code, NameMK: n[0], NameEN: n[1]}
	}
	return Country{Code: code, NameMK: code, NameEN: code}
}
