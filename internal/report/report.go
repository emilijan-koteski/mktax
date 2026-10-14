package report

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/emilijan-koteski/mktax/mktax"
	"github.com/emilijan-koteski/mktax/t212csv"
)

type Warning struct {
	Code   string            `json:"code"`
	Params map[string]string `json:"params"`
}

type Country struct {
	Code   string `json:"code"`
	NameMK string `json:"name_mk"`
	NameEN string `json:"name_en"`
}

type Lot struct {
	Acquired        string `json:"acquired"`
	AcquiredAt      string `json:"acquired_at"`
	Quantity        string `json:"quantity"`
	Cost            string `json:"cost"`
	Proceeds        string `json:"proceeds"`
	Currency        string `json:"currency"`
	AcquisitionRate string `json:"acquisition_rate"`
	CostMKD         string `json:"cost_mkd"`
	ProceedsMKD     string `json:"proceeds_mkd"`
	GainMKD         string `json:"gain_mkd"`
	GainMKDB        string `json:"gain_mkd_b"`
	Status          string `json:"status"`
	TaxFreeFrom     string `json:"tax_free_from"`
}

type Sale struct {
	Ticker            string    `json:"ticker"`
	ISIN              string    `json:"isin"`
	Name              string    `json:"name"`
	Date              string    `json:"date"`
	Time              string    `json:"time"`
	Ref               string    `json:"ref"`
	Source            string    `json:"source"`
	IncomeType        string    `json:"income_type"`
	IncomeSubtype     string    `json:"income_subtype"`
	FundUnits         bool      `json:"fund_units"`
	Country           Country   `json:"country"`
	Currency          string    `json:"currency"`
	Quantity          string    `json:"quantity"`
	Proceeds          string    `json:"proceeds"`
	Cost              string    `json:"cost"`
	Gain              string    `json:"gain"`
	TaxableGain       string    `json:"taxable_gain"`
	Fees              string    `json:"fees"`
	SaleRate          string    `json:"sale_rate"`
	ProceedsMKD       string    `json:"proceeds_mkd"`
	CostMKD           string    `json:"cost_mkd"`
	GainMKD           string    `json:"gain_mkd"`
	GainMKDB          string    `json:"gain_mkd_b"`
	TaxableGainMKD    string    `json:"taxable_gain_mkd"`
	TaxableGainMKDB   string    `json:"taxable_gain_mkd_b"`
	ExemptGainMKD     string    `json:"exempt_gain_mkd"`
	OutOfScopeGainMKD string    `json:"out_of_scope_gain_mkd"`
	TaxMKD            string    `json:"tax_mkd"`
	CostComplete      bool      `json:"cost_complete"`
	Enter             bool      `json:"enter"`
	SkipReason        string    `json:"skip_reason"`
	Lots              []Lot     `json:"lots"`
	Warnings          []Warning `json:"warnings"`
}

type Dividend struct {
	Ticker                 string    `json:"ticker"`
	ISIN                   string    `json:"isin"`
	Name                   string    `json:"name"`
	Type                   string    `json:"type"`
	Date                   string    `json:"date"`
	Time                   string    `json:"time"`
	IncomeType             string    `json:"income_type"`
	IncomeSubtype          string    `json:"income_subtype"`
	Country                Country   `json:"country"`
	Gross                  string    `json:"gross"`
	GrossCurrency          string    `json:"gross_currency"`
	Net                    string    `json:"net"`
	NetCurrency            string    `json:"net_currency"`
	ForeignTax             string    `json:"foreign_tax"`
	ForeignTaxCurrency     string    `json:"foreign_tax_currency"`
	GrossRate              string    `json:"gross_rate"`
	NetRate                string    `json:"net_rate"`
	GrossMKD               string    `json:"gross_mkd"`
	NetMKD                 string    `json:"net_mkd"`
	WithheldMKD            string    `json:"withheld_mkd"`
	TaxIfCreditAcceptedMKD string    `json:"tax_if_credit_accepted_mkd"`
	TaxIfCreditRejectedMKD string    `json:"tax_if_credit_rejected_mkd"`
	Warnings               []Warning `json:"warnings"`
}

type Totals struct {
	Sales                          int    `json:"sales"`
	UnknownCostSales               int    `json:"unknown_cost_sales"`
	ProceedsMKD                    string `json:"proceeds_mkd"`
	TaxableGainsMKD                string `json:"taxable_gains_mkd"`
	TaxableLossesMKD               string `json:"taxable_losses_mkd"`
	NetIfLossesOffsetMKD           string `json:"net_if_losses_offset_mkd"`
	TaxBaseMKD                     string `json:"tax_base_mkd"`
	CapitalGainsTaxMKD             string `json:"capital_gains_tax_mkd"`
	TaxIfLossesOffsetMKD           string `json:"tax_if_losses_offset_mkd"`
	ExemptGainsMKD                 string `json:"exempt_gains_mkd"`
	OutOfScopeGainsMKD             string `json:"out_of_scope_gains_mkd"`
	TaxableGainsMKDB               string `json:"taxable_gains_mkd_b"`
	CapitalGainsTaxMKDB            string `json:"capital_gains_tax_mkd_b"`
	Dividends                      int    `json:"dividends"`
	DividendGrossMKD               string `json:"dividend_gross_mkd"`
	DividendNetMKD                 string `json:"dividend_net_mkd"`
	DividendWithheldMKD            string `json:"dividend_withheld_mkd"`
	DividendTaxIfCreditAcceptedMKD string `json:"dividend_tax_if_credit_accepted_mkd"`
	DividendTaxIfCreditRejectedMKD string `json:"dividend_tax_if_credit_rejected_mkd"`
	TaxDueIfCreditAcceptedMKD      string `json:"tax_due_if_credit_accepted_mkd"`
	TaxDueIfCreditRejectedMKD      string `json:"tax_due_if_credit_rejected_mkd"`
}

type PortalRow struct {
	N                      int     `json:"n"`
	Kind                   string  `json:"kind"`
	Date                   string  `json:"date"`
	Ticker                 string  `json:"ticker"`
	ISIN                   string  `json:"isin"`
	Name                   string  `json:"name"`
	Ref                    string  `json:"ref"`
	IncomeType             string  `json:"income_type"`
	IncomeSubtype          string  `json:"income_subtype"`
	Country                Country `json:"country"`
	Currency               string  `json:"currency"`
	Gross                  string  `json:"gross"`
	ForeignTax             string  `json:"foreign_tax"`
	Enter                  bool    `json:"enter"`
	SkipReason             string  `json:"skip_reason"`
	GrossMKD               string  `json:"gross_mkd"`
	TaxMKD                 string  `json:"tax_mkd"`
	TaxIfCreditAcceptedMKD string  `json:"tax_if_credit_accepted_mkd"`
}

type Month struct {
	Month       string      `json:"month"`
	FileBy      string      `json:"file_by"`
	PayBy       string      `json:"pay_by"`
	HasEntries  bool        `json:"has_entries"`
	Overdue     bool        `json:"overdue"`
	InProgress  bool        `json:"in_progress"`
	Sales       []Sale      `json:"sales"`
	Dividends   []Dividend  `json:"dividends"`
	Totals      Totals      `json:"totals"`
	PortalRows  []PortalRow `json:"portal_rows"`
	Warnings    []Warning   `json:"warnings"`
	RawWarnings []string    `json:"raw_warnings"`
}

type OpenLot struct {
	Acquired    string `json:"acquired"`
	Quantity    string `json:"quantity"`
	Cost        string `json:"cost"`
	TaxFreeFrom string `json:"tax_free_from"`
}

type OpenPosition struct {
	Key                  string    `json:"key"`
	Ticker               string    `json:"ticker"`
	ISIN                 string    `json:"isin"`
	Name                 string    `json:"name"`
	Currency             string    `json:"currency"`
	Quantity             string    `json:"quantity"`
	Cost                 string    `json:"cost"`
	EarliestAcquired     string    `json:"earliest_acquired"`
	OldestLotTaxFreeFrom string    `json:"oldest_lot_tax_free_from"`
	TaxFreeQuantityNow   string    `json:"tax_free_quantity_now"`
	Lots                 []OpenLot `json:"lots"`
}

type Row struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	ID       string `json:"id"`
	Action   string `json:"action"`
	Date     string `json:"date"`
	Time     string `json:"time"`
	Ticker   string `json:"ticker"`
	ISIN     string `json:"isin"`
	Name     string `json:"name"`
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
	Notes    string `json:"notes"`
	Kind     string `json:"kind"`
}

type FileSummary struct {
	Name       string   `json:"name"`
	Rows       int      `json:"rows"`
	Trades     int      `json:"trades"`
	Dividends  int      `json:"dividends"`
	Interest   int      `json:"interest"`
	Ignored    int      `json:"ignored"`
	Duplicates int      `json:"duplicates"`
	Unknown    int      `json:"unknown"`
	First      string   `json:"first"`
	Last       string   `json:"last"`
	Columns    []string `json:"columns"`
}

type Instrument struct {
	Key              string `json:"key"`
	Ticker           string `json:"ticker"`
	ISIN             string `json:"isin"`
	Name             string `json:"name"`
	Currency         string `json:"currency"`
	FundUnits        bool   `json:"fund_units"`
	FundUnitsGuessed bool   `json:"fund_units_guessed"`
	Buys             int    `json:"buys"`
	Sells            int    `json:"sells"`
	Dividends        int    `json:"dividends"`
}

type Counts struct {
	Trades    int `json:"trades"`
	Dividends int `json:"dividends"`
}

type Import struct {
	AccountCurrency string         `json:"account_currency"`
	First           string         `json:"first"`
	Last            string         `json:"last"`
	Counts          Counts         `json:"counts"`
	Files           []FileSummary  `json:"files"`
	Instruments     []Instrument   `json:"instruments"`
	Interest        []Row          `json:"interest"`
	Excluded        []Row          `json:"excluded"`
	Ignored         map[string]int `json:"ignored"`
	Unknown         map[string]int `json:"unknown"`
	Flags           []Warning      `json:"flags"`
}

type Report struct {
	GeneratedAt string         `json:"generated_at"`
	Currency    string         `json:"currency"`
	From        string         `json:"from"`
	To          string         `json:"to"`
	FxSource    string         `json:"fx_source"`
	FxMethod    string         `json:"fx_method"`
	Months      []Month        `json:"months"`
	OpenLots    []OpenPosition `json:"open_lots"`
	Notes       []string       `json:"notes"`
	Warnings    []Warning      `json:"warnings"`
	RawWarnings []string       `json:"raw_warnings"`
	Import      Import         `json:"import"`
}

func New(r mktax.Report, imp t212csv.Import, now time.Time) Report {
	out := Report{
		GeneratedAt: ts(now), Currency: imp.AccountCurrency, From: month(r.From), To: month(r.To),
		FxSource: "NBRSM", FxMethod: "per_leg",
		Months: make([]Month, 0, len(r.Months)), OpenLots: make([]OpenPosition, 0, len(r.OpenLots)),
		Notes: strs(r.Notes), Warnings: []Warning{}, RawWarnings: strs(r.Warnings), Import: ImportSummary(imp),
	}
	for _, m := range r.Months {
		out.Months = append(out.Months, newMonth(m))
	}
	for _, p := range r.OpenLots {
		out.OpenLots = append(out.OpenLots, newOpenPosition(p))
	}
	return out
}

func newMonth(m mktax.MonthReport) Month {
	out := Month{
		Month: month(m.Month), FileBy: date(m.FileBy), PayBy: date(m.PayBy),
		HasEntries: m.HasEntries, Overdue: m.Overdue, InProgress: m.InProgress,
		Sales: make([]Sale, 0, len(m.Sales)), Dividends: make([]Dividend, 0, len(m.Dividends)),
		Totals: newTotals(m.Totals), PortalRows: []PortalRow{}, Warnings: []Warning{}, RawWarnings: strs(m.Warnings),
	}
	for _, s := range m.Sales {
		sale := newSale(s)
		out.Sales = append(out.Sales, sale)
		out.Warnings = append(out.Warnings, sale.Warnings...)
		out.PortalRows = append(out.PortalRows, saleRow(sale))
	}
	for _, l := range m.Dividends {
		dv := newDividend(l)
		out.Dividends = append(out.Dividends, dv)
		out.Warnings = append(out.Warnings, dv.Warnings...)
		out.PortalRows = append(out.PortalRows, dividendRow(dv))
	}
	slices.SortStableFunc(out.PortalRows, func(a, b PortalRow) int { return strings.Compare(a.Date, b.Date) })
	for i := range out.PortalRows {
		out.PortalRows[i].N = i + 1
	}
	return out
}

// Flattens one engine sale into the DTO, adds the portal fields the engine leaves to the caller (income codes, payer country) and derives structured warnings from the matched lots.
func newSale(s mktax.Sale) Sale {
	out := Sale{
		Ticker: s.Ticker, ISIN: s.ISIN, Name: s.Name, Date: date(s.Date), Time: ts(s.Time), Ref: s.Ref, Source: s.Source,
		IncomeType: mktax.IncomeTypeCapitalGains, IncomeSubtype: mktax.CapitalGainSubtype(s.FundUnits), FundUnits: s.FundUnits,
		Country: country(mktax.CountryFromISIN(s.ISIN)), Currency: s.Currency, Quantity: num(s.Quantity),
		Proceeds: money(s.Value), Cost: money(s.Cost), Gain: money(s.Gain), TaxableGain: money(s.TaxableGain), Fees: money(s.Fees),
		SaleRate: num(s.SaleRate), ProceedsMKD: money(s.ProceedsMKD), CostMKD: money(s.CostMKD), GainMKD: money(s.GainMKD), GainMKDB: money(s.GainMKDB),
		TaxableGainMKD: money(s.TaxableGainMKD), TaxableGainMKDB: money(s.TaxableGainMKDB), ExemptGainMKD: money(s.ExemptGainMKD),
		OutOfScopeGainMKD: money(s.OutOfScopeGainMKD), TaxMKD: money(s.TaxMKD), CostComplete: s.CostComplete,
		Lots: make([]Lot, 0, len(s.Lots)), Warnings: []Warning{},
	}
	unknown := decimal.Zero
	var taxable, exempt, outOfScope bool
	for _, l := range s.Lots {
		out.Lots = append(out.Lots, newLot(l))
		switch l.Status {
		case mktax.StatusTaxable:
			taxable = true
		case mktax.StatusExempt:
			exempt = true
		case mktax.StatusOutOfScope:
			outOfScope = true
		case mktax.StatusUnknownCost:
			unknown = unknown.Add(l.Quantity)
		}
	}
	if unknown.IsPositive() {
		out.Warnings = append(out.Warnings, warn("oversell", "ticker", s.Ticker, "date", out.Date, "quantity", num(unknown)))
	}
	for _, l := range s.Lots {
		if l.Status != mktax.StatusUnknownCost && l.Currency != s.Currency {
			out.Warnings = append(out.Warnings, warn("method_b_cross_currency", "ticker", s.Ticker, "date", out.Date, "lot_currency", l.Currency, "sale_currency", s.Currency))
		}
	}
	switch {
	case !taxable && exempt:
		out.SkipReason = "exempt"
	case !taxable && outOfScope:
		out.SkipReason = "out_of_scope"
	case !taxable:
		out.SkipReason = "unknown_cost"
	case !s.TaxableGain.IsPositive():
		out.SkipReason = "loss"
	default:
		out.Enter = true
	}
	return out
}

func newLot(l mktax.LotMatch) Lot {
	return Lot{
		Acquired: date(l.Acquired), AcquiredAt: ts(l.AcquiredAt), Quantity: num(l.Quantity), Cost: money(l.Cost), Proceeds: money(l.Proceeds),
		Currency: l.Currency, AcquisitionRate: num(l.AcquisitionRate), CostMKD: money(l.CostMKD), ProceedsMKD: money(l.ProceedsMKD),
		GainMKD: money(l.GainMKD), GainMKDB: money(l.GainMKDB), Status: string(l.Status), TaxFreeFrom: date(l.TaxFreeFrom),
	}
}

func newDividend(l mktax.DividendLine) Dividend {
	out := Dividend{
		Ticker: l.Ticker, ISIN: l.ISIN, Name: l.Name, Type: l.Type, Date: date(l.Date), Time: ts(l.PaidOn),
		IncomeType: l.IncomeType, IncomeSubtype: l.IncomeSubtype, Country: country(l.Country),
		Gross: money(l.Gross), GrossCurrency: l.GrossCurrency, Net: money(l.Net), NetCurrency: l.NetCurrency,
		ForeignTaxCurrency: l.GrossCurrency, GrossRate: num(l.GrossRate), NetRate: num(l.NetRate),
		GrossMKD: money(l.GrossMKD), NetMKD: money(l.NetMKD), WithheldMKD: money(l.WithheldMKD),
		TaxIfCreditAcceptedMKD: money(l.TaxIfCreditAcceptedMKD), TaxIfCreditRejectedMKD: money(l.TaxIfCreditRejectedMKD),
		Warnings: []Warning{},
	}
	out.ForeignTax = money(decimal.Zero)
	if l.GrossRate.IsPositive() {
		out.ForeignTax = money(l.WithheldMKD.Div(l.GrossRate))
	}
	if len(l.Warnings) > 0 {
		out.Warnings = append(out.Warnings, warn("gross_from_net", "ticker", l.Ticker, "date", out.Date))
	}
	return out
}

func saleRow(s Sale) PortalRow {
	return PortalRow{
		Kind: "sale", Date: s.Date, Ticker: s.Ticker, ISIN: s.ISIN, Name: s.Name, Ref: s.Ref,
		IncomeType: s.IncomeType, IncomeSubtype: s.IncomeSubtype, Country: s.Country, Currency: s.Currency,
		Gross: s.TaxableGain, ForeignTax: money(decimal.Zero), Enter: s.Enter, SkipReason: s.SkipReason,
		GrossMKD: s.TaxableGainMKD, TaxMKD: s.TaxMKD, TaxIfCreditAcceptedMKD: s.TaxMKD,
	}
}

func dividendRow(d Dividend) PortalRow {
	return PortalRow{
		Kind: "dividend", Date: d.Date, Ticker: d.Ticker, ISIN: d.ISIN, Name: d.Name,
		IncomeType: d.IncomeType, IncomeSubtype: d.IncomeSubtype, Country: d.Country, Currency: d.GrossCurrency,
		Gross: d.Gross, ForeignTax: d.ForeignTax, Enter: true,
		GrossMKD: d.GrossMKD, TaxMKD: d.TaxIfCreditRejectedMKD, TaxIfCreditAcceptedMKD: d.TaxIfCreditAcceptedMKD,
	}
}

func newTotals(t mktax.MonthTotals) Totals {
	return Totals{
		Sales: t.Sales, UnknownCostSales: t.UnknownCostSales, ProceedsMKD: money(t.ProceedsMKD),
		TaxableGainsMKD: money(t.TaxableGainsMKD), TaxableLossesMKD: money(t.TaxableLossesMKD), NetIfLossesOffsetMKD: money(t.NetIfLossesOffsetMKD),
		TaxBaseMKD: money(t.TaxBaseMKD), CapitalGainsTaxMKD: money(t.CapitalGainsTaxMKD), TaxIfLossesOffsetMKD: money(t.TaxIfLossesOffsetMKD),
		ExemptGainsMKD: money(t.ExemptGainsMKD), OutOfScopeGainsMKD: money(t.OutOfScopeGainsMKD),
		TaxableGainsMKDB: money(t.TaxableGainsMKDB), CapitalGainsTaxMKDB: money(t.CapitalGainsTaxMKDB),
		Dividends: t.Dividends, DividendGrossMKD: money(t.DividendGrossMKD), DividendNetMKD: money(t.DividendNetMKD),
		DividendWithheldMKD: money(t.DividendWithheldMKD), DividendTaxIfCreditAcceptedMKD: money(t.DividendTaxIfCreditAcceptedMKD),
		DividendTaxIfCreditRejectedMKD: money(t.DividendTaxIfCreditRejectedMKD),
		TaxDueIfCreditAcceptedMKD:      money(t.TaxDueIfCreditAcceptedMKD), TaxDueIfCreditRejectedMKD: money(t.TaxDueIfCreditRejectedMKD),
	}
}

func newOpenPosition(p mktax.OpenPosition) OpenPosition {
	out := OpenPosition{
		Key: p.Key, Ticker: p.Ticker, ISIN: p.ISIN, Name: p.Name, Currency: p.Currency, Quantity: num(p.Quantity), Cost: money(p.Cost),
		EarliestAcquired: date(p.EarliestAcquired), OldestLotTaxFreeFrom: date(p.OldestLotTaxFreeFrom), TaxFreeQuantityNow: num(p.TaxFreeQuantityNow),
		Lots: make([]OpenLot, 0, len(p.Lots)),
	}
	for _, l := range p.Lots {
		out.Lots = append(out.Lots, OpenLot{Acquired: date(l.Acquired), Quantity: num(l.Quantity), Cost: money(l.Cost), TaxFreeFrom: date(l.TaxFreeFrom)})
	}
	return out
}

func ImportSummary(imp t212csv.Import) Import {
	out := Import{
		AccountCurrency: imp.AccountCurrency, First: dateOf(imp.First), Last: dateOf(imp.Last),
		Counts: Counts{Trades: len(imp.Trades), Dividends: len(imp.Dividends)},
		Files:  make([]FileSummary, 0, len(imp.Files)), Instruments: make([]Instrument, 0, len(imp.Instruments)),
		Interest: rows(imp.Interest), Excluded: rows(imp.Excluded),
		Ignored: counts(imp.Ignored), Unknown: counts(imp.Unknown), Flags: make([]Warning, 0, len(imp.Flags)),
	}
	for _, f := range imp.Files {
		out.Files = append(out.Files, FileSummary{
			Name: f.Name, Rows: f.Rows, Trades: f.Trades, Dividends: f.Dividends, Interest: f.Interest, Ignored: f.Ignored,
			Duplicates: f.Duplicates, Unknown: f.Unknown, First: dateOf(f.First), Last: dateOf(f.Last), Columns: strs(f.Columns),
		})
	}
	for _, i := range imp.Instruments {
		out.Instruments = append(out.Instruments, Instrument(i))
	}
	for _, f := range imp.Flags {
		w := Warning{Code: f.Code, Params: map[string]string{}}
		for k, v := range f.Params {
			w.Params[k] = v
		}
		if f.File != "" {
			w.Params["file"] = f.File
			w.Params["line"] = fmt.Sprint(f.Line)
		}
		for k, v := range map[string]string{"id": f.ID, "action": f.Action, "ticker": f.Ticker} {
			if v != "" {
				w.Params[k] = v
			}
		}
		out.Flags = append(out.Flags, w)
	}
	return out
}

func rows(in []t212csv.Row) []Row {
	out := make([]Row, 0, len(in))
	for _, r := range in {
		out = append(out, Row{
			File: r.File, Line: r.Line, ID: r.ID, Action: r.Action, Date: dateOf(r.Time), Time: ts(r.Time),
			Ticker: r.Ticker, ISIN: r.ISIN, Name: r.Name, Amount: money(r.Amount), Currency: r.Currency, Notes: r.Notes, Kind: string(r.Kind),
		})
	}
	return out
}

func counts(in map[string]int) map[string]int {
	out := make(map[string]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func strs(in []string) []string {
	if in == nil {
		return []string{}
	}
	return slices.Clone(in)
}

func warn(code string, kv ...string) Warning {
	w := Warning{Code: code, Params: map[string]string{}}
	for i := 0; i+1 < len(kv); i += 2 {
		w.Params[kv[i]] = kv[i+1]
	}
	return w
}

func country(c mktax.Country) Country {
	return Country{Code: c.Code, NameMK: c.NameMK, NameEN: c.NameEN}
}

func money(d decimal.Decimal) string { return d.StringFixed(2) }

func num(d decimal.Decimal) string { return d.String() }

func date(d mktax.Date) string { return d.String() }

func month(m mktax.Month) string { return m.String() }

func dateOf(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return mktax.DateOf(t).String()
}

func ts(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
