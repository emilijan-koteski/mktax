package mktax

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const (
	IncomeTypeCapital            = "T7 Доход од капитал"
	IncomeSubtypeDividends       = "S7.1 Дивиденди и друг доход остварени со учество во добивката"
	IncomeTypeCapitalGains       = "T8 Капитални добивки"
	IncomeSubtypeSecuritiesGains = "S8.1 Капитална добивка од продажба на хартии од вредност"
	IncomeSubtypeFundUnitsGains  = "S.8.7 Капитални добивки од продажба на удели издадени од инвестициски фондови"
)

func CapitalGainSubtype(fundUnits bool) string {
	if fundUnits {
		return IncomeSubtypeFundUnitsGains
	}
	return IncomeSubtypeSecuritiesGains
}

const exemptYears = 2

var (
	regimeStart  = Date{2023, time.January, 1}
	taxRate      = decimal.New(1, -1)
	taxBaseShare = decimal.New(9, -1)
	dust         = decimal.New(1, -8)
	one          = decimal.NewFromInt(1)
)

func FirstMonth() Month { return MonthOf(regimeStart) }

type Side string

const (
	Buy  Side = "BUY"
	Sell Side = "SELL"
)

type Trade struct {
	Key       string
	Ticker    string
	ISIN      string
	Name      string
	Time      time.Time
	Side      Side
	Quantity  decimal.Decimal
	Value     decimal.Decimal
	Currency  string
	Fees      decimal.Decimal
	Source    string
	Ref       string
	FundUnits bool
}

type Dividend struct {
	Ticker        string
	ISIN          string
	Name          string
	Type          string
	PaidOn        time.Time
	Gross         decimal.Decimal
	GrossCurrency string
	Net           decimal.Decimal
	NetCurrency   string
}

type RateFunc func(currency string, day Date) (decimal.Decimal, error)

type Options struct {
	From, To Month
	Now      time.Time
}

type LotStatus string

const (
	StatusTaxable     LotStatus = "taxable"
	StatusExempt      LotStatus = "exempt_held_over_2y"
	StatusOutOfScope  LotStatus = "out_of_scope_pre_2023"
	StatusUnknownCost LotStatus = "unknown_cost"
)

type LotMatch struct {
	Acquired        Date
	AcquiredAt      time.Time
	Quantity        decimal.Decimal
	Cost            decimal.Decimal
	Proceeds        decimal.Decimal
	Currency        string
	AcquisitionRate decimal.Decimal
	CostMKD         decimal.Decimal
	ProceedsMKD     decimal.Decimal
	GainMKD         decimal.Decimal
	GainMKDB        decimal.Decimal
	Status          LotStatus
	TaxFreeFrom     Date
}

type Sale struct {
	Trade
	Date              Date
	SaleRate          decimal.Decimal
	ProceedsMKD       decimal.Decimal
	Cost              decimal.Decimal
	Gain              decimal.Decimal
	TaxableGain       decimal.Decimal
	CostMKD           decimal.Decimal
	GainMKD           decimal.Decimal
	GainMKDB          decimal.Decimal
	TaxableGainMKD    decimal.Decimal
	TaxableGainMKDB   decimal.Decimal
	ExemptGainMKD     decimal.Decimal
	OutOfScopeGainMKD decimal.Decimal
	TaxMKD            decimal.Decimal
	CostComplete      bool
	Lots              []LotMatch
	Warnings          []string
}

type DividendLine struct {
	Dividend
	Date                   Date
	Country                Country
	IncomeType             string
	IncomeSubtype          string
	GrossRate              decimal.Decimal
	NetRate                decimal.Decimal
	GrossMKD               decimal.Decimal
	NetMKD                 decimal.Decimal
	WithheldMKD            decimal.Decimal
	TaxIfCreditAcceptedMKD decimal.Decimal
	TaxIfCreditRejectedMKD decimal.Decimal
	Warnings               []string
}

type MonthTotals struct {
	Sales                int
	UnknownCostSales     int
	ProceedsMKD          decimal.Decimal
	TaxableGainsMKD      decimal.Decimal
	TaxableLossesMKD     decimal.Decimal
	NetIfLossesOffsetMKD decimal.Decimal
	TaxBaseMKD           decimal.Decimal
	CapitalGainsTaxMKD   decimal.Decimal
	TaxIfLossesOffsetMKD decimal.Decimal
	ExemptGainsMKD       decimal.Decimal
	OutOfScopeGainsMKD   decimal.Decimal
	TaxableGainsMKDB     decimal.Decimal
	CapitalGainsTaxMKDB  decimal.Decimal

	Dividends                      int
	DividendGrossMKD               decimal.Decimal
	DividendNetMKD                 decimal.Decimal
	DividendWithheldMKD            decimal.Decimal
	DividendTaxIfCreditAcceptedMKD decimal.Decimal
	DividendTaxIfCreditRejectedMKD decimal.Decimal

	TaxDueIfCreditAcceptedMKD decimal.Decimal
	TaxDueIfCreditRejectedMKD decimal.Decimal
}

type MonthReport struct {
	Month      Month
	FileBy     Date
	PayBy      Date
	HasEntries bool
	Overdue    bool
	InProgress bool
	Sales      []Sale
	Dividends  []DividendLine
	Totals     MonthTotals
	Warnings   []string
}

type OpenLot struct {
	Acquired    Date
	Quantity    decimal.Decimal
	Cost        decimal.Decimal
	TaxFreeFrom Date
}

type OpenPosition struct {
	Key                  string
	Ticker               string
	ISIN                 string
	Name                 string
	Currency             string
	Quantity             decimal.Decimal
	Cost                 decimal.Decimal
	EarliestAcquired     Date
	OldestLotTaxFreeFrom Date
	TaxFreeQuantityNow   decimal.Decimal
	Lots                 []OpenLot
}

type Report struct {
	From, To Month
	Months   []MonthReport
	OpenLots []OpenPosition
	Notes    []string
	Warnings []string
}

func RuleNotes() []string {
	return []string{
		"Capital gain per sale = proceeds − FIFO cost of the matched lots (Art. 58 ЗДЛД; FIFO oldest-first per instrument, Art. 22 of Sl. vesnik 274/22). Tax = 10% of 90% of the taxable gain (10% normed costs, Art. 60(2)), i.e. 9% effective.",
		"A matched lot is exempt when the sale date is more than two calendar years after its acquisition date (Art. 12(1) item 39-а), tested per lot on Europe/Skopje dates.",
		"Lots acquired before 2023-01-01 are consumed by FIFO like any lot, but their gain is treated as out of scope under transitional Art. 21 of Sl. vesnik 274/22. This is an interpretation; confirm it with UJP.",
		"Foreign losses are not offset against gains: Art. 62 allows netting only for sales on a domestic exchange. net_if_losses_offset and tax_if_losses_offset are information only.",
		"MKD conversion uses the NBRSM middle rate valid on each transaction's date. Tax figures use method A (per_leg: proceeds at the sale-date rate, each lot's cost at its acquisition-date rate); method B (gain_at_sale_rate) is shown for comparison. UJP has not specified the method.",
		"e-ППД codes (read from the e-ППД portal calculator, 2026-10-14): a share sale is T8/S8.1, a fund-unit (ETF) sale is T8/S.8.7, a dividend is T7/S7.1. Choose the payer's country (not Macedonia) to open the foreign-currency fields. Enter the gain itself: the portal applies the 10% normed costs (1,000 gain → 90 tax).",
		"Dividends are taxed at 10% of the gross amount (Art. 55–57). The foreign tax credit (Art. 8(4), up to 10%) is currently accepted by UJP only with proof from the foreign tax authority, so tax is shown both with and without the credit.",
		"Deadlines: file the e-ППД by the 10th and pay by the 15th of the month after the income; these dates are not shifted for weekends or public holidays.",
		"Payer country comes from the ISIN prefix, which can differ from the issuer's tax residence (for example ADRs carry a US ISIN).",
	}
}

// Builds the monthly report: FIFO runs over every trade and only months in [From, To] are emitted. A sale exceeding the recorded buys yields an unknown-cost portion and a warning instead of an error.
func Compute(trades []Trade, divs []Dividend, rates RateFunc, opts Options) (Report, error) {
	if err := opts.validate(); err != nil {
		return Report{}, err
	}
	if rates == nil {
		return Report{}, errors.New("mktax: a rate function is required")
	}
	led, err := match(trades)
	if err != nil {
		return Report{}, err
	}
	rate := mkdRate(rates)
	today := DateOf(opts.Now)
	r := Report{From: opts.From, To: opts.To, Notes: RuleNotes()}
	idx := map[Month]int{}
	for m := opts.From; m.Compare(opts.To) <= 0; m = m.AddMonths(1) {
		idx[m] = len(r.Months)
		r.Months = append(r.Months, MonthReport{Month: m, FileBy: fileBy(m), PayBy: payBy(m), InProgress: m == MonthOf(today)})
	}
	for _, ms := range led.sales {
		if ms.short().IsPositive() {
			r.Warnings = append(r.Warnings, ms.oversellWarning())
		}
		i, ok := idx[MonthOf(ms.date)]
		if !ok {
			continue
		}
		s, err := priceSale(ms, rate)
		if err != nil {
			return Report{}, err
		}
		r.Months[i].Sales = append(r.Months[i].Sales, s)
	}
	for n, dv := range divs {
		if err := dv.validate(); err != nil {
			return Report{}, fmt.Errorf("mktax: dividend %d (%s): %w", n, dv.Ticker, err)
		}
		date := DateOf(dv.PaidOn)
		i, ok := idx[MonthOf(date)]
		if !ok {
			continue
		}
		l, err := priceDividend(dv, date, rate)
		if err != nil {
			return Report{}, err
		}
		r.Months[i].Dividends = append(r.Months[i].Dividends, l)
	}
	for i := range r.Months {
		r.Months[i].finish(today)
	}
	r.OpenLots = led.openPositions(today)
	return r, nil
}

// Lists, per currency, the dates whose MKD rate Compute will request for the same inputs, so callers can prefetch them. MKD is omitted.
func RequiredRates(trades []Trade, divs []Dividend, from, to Month) (map[string][]Date, error) {
	led, err := match(trades)
	if err != nil {
		return nil, err
	}
	set := map[string]map[Date]bool{}
	add := func(ccy string, d Date) {
		if ccy == "" || strings.EqualFold(ccy, "MKD") {
			return
		}
		if set[ccy] == nil {
			set[ccy] = map[Date]bool{}
		}
		set[ccy][d] = true
	}
	in := func(d Date) bool {
		m := MonthOf(d)
		return m.Compare(from) >= 0 && m.Compare(to) <= 0
	}
	for _, ms := range led.sales {
		if !in(ms.date) {
			continue
		}
		add(ms.trade.Currency, ms.date)
		for _, p := range ms.portions {
			if !p.unknown {
				add(p.lot.currency, p.lot.date)
			}
		}
	}
	for _, dv := range divs {
		date := DateOf(dv.PaidOn)
		if dv.PaidOn.IsZero() || !in(date) {
			continue
		}
		_, gc, _ := dv.gross()
		add(gc, date)
		add(dv.NetCurrency, date)
	}
	out := make(map[string][]Date, len(set))
	for ccy, days := range set {
		ds := make([]Date, 0, len(days))
		for d := range days {
			ds = append(ds, d)
		}
		slices.SortFunc(ds, Date.Compare)
		out[ccy] = ds
	}
	return out, nil
}

func (o Options) validate() error {
	switch {
	case o.From.IsZero() || o.To.IsZero():
		return errors.New("mktax: From and To months are required")
	case o.From.Compare(o.To) > 0:
		return fmt.Errorf("mktax: From %s is after To %s", o.From, o.To)
	case o.From.Compare(FirstMonth()) < 0:
		return fmt.Errorf("mktax: the rules encoded here apply from %s; got From %s", FirstMonth(), o.From)
	case o.Now.IsZero():
		return errors.New("mktax: Options.Now is required (the package has no clock)")
	}
	return nil
}

func (t Trade) validate() error {
	switch {
	case t.Key == "":
		return errors.New("no instrument: set Key, ISIN or Ticker")
	case t.Side != Buy && t.Side != Sell:
		return fmt.Errorf("invalid side %q", t.Side)
	case t.Time.IsZero():
		return errors.New("no time")
	case !t.Quantity.IsPositive():
		return errors.New("quantity must be positive")
	case t.Value.IsNegative():
		return errors.New("value must not be negative")
	case t.Currency == "":
		return errors.New("no currency")
	}
	return nil
}

func (d Dividend) validate() error {
	switch {
	case d.PaidOn.IsZero():
		return errors.New("no payment date")
	case d.NetCurrency == "":
		return errors.New("no net currency")
	}
	return nil
}

func (d Dividend) gross() (decimal.Decimal, string, bool) {
	if !d.Gross.IsPositive() || d.GrossCurrency == "" {
		return d.Net, d.NetCurrency, true
	}
	return d.Gross, d.GrossCurrency, false
}

type rateFn func(ccy string, day Date) (decimal.Decimal, error)

func mkdRate(f RateFunc) rateFn {
	return func(ccy string, day Date) (decimal.Decimal, error) {
		if strings.EqualFold(ccy, "MKD") {
			return one, nil
		}
		r, err := f(ccy, day)
		if err != nil {
			return decimal.Zero, fmt.Errorf("mktax: MKD rate for %s on %s: %w", ccy, day, err)
		}
		if !r.IsPositive() {
			return decimal.Zero, fmt.Errorf("mktax: MKD rate for %s on %s is not positive: %s", ccy, day, r)
		}
		return r, nil
	}
}

func taxOnGain(g decimal.Decimal) decimal.Decimal {
	if !g.IsPositive() {
		return decimal.Zero
	}
	return g.Mul(taxBaseShare).Mul(taxRate)
}

func classify(acquired, sold Date) LotStatus {
	switch {
	case sold.Compare(TaxFreeFrom(acquired)) >= 0:
		return StatusExempt
	case acquired.Before(regimeStart):
		return StatusOutOfScope
	default:
		return StatusTaxable
	}
}

// Prices one sale in MKD: proceeds at the sale-date rate and each lot's cost at its acquisition-date rate (method A). Method B converts the gain at the sale-date rate and repeats A when the currencies differ.
func priceSale(ms matchedSale, rate rateFn) (Sale, error) {
	t := ms.trade
	s := Sale{Trade: t, Date: ms.date, CostComplete: true}
	sr, err := rate(t.Currency, ms.date)
	if err != nil {
		return Sale{}, err
	}
	s.SaleRate, s.ProceedsMKD = sr, t.Value.Mul(sr)
	known := decimal.Zero
	for _, p := range ms.portions {
		lm := LotMatch{Quantity: p.lot.qty, Proceeds: p.proceeds, ProceedsMKD: p.proceeds.Mul(sr), Currency: t.Currency}
		if p.unknown {
			lm.Status, s.CostComplete = StatusUnknownCost, false
			s.Warnings = append(s.Warnings, fmt.Sprintf("%s: sold %s more than the recorded buys on %s; that part's cost is unknown and it is left out of the tax figures (tax may be understated)",
				label(t), p.lot.qty, ms.date))
			s.Lots = append(s.Lots, lm)
			continue
		}
		ar, err := rate(p.lot.currency, p.lot.date)
		if err != nil {
			return Sale{}, err
		}
		lm.Acquired, lm.AcquiredAt, lm.Cost, lm.Currency = p.lot.date, p.lot.at, p.lot.cost, p.lot.currency
		lm.AcquisitionRate, lm.CostMKD = ar, p.lot.cost.Mul(ar)
		lm.GainMKD = lm.ProceedsMKD.Sub(lm.CostMKD)
		if p.lot.currency == t.Currency {
			lm.GainMKDB = p.proceeds.Sub(p.lot.cost).Mul(sr)
		} else {
			lm.GainMKDB = lm.GainMKD
			s.Warnings = append(s.Warnings, fmt.Sprintf("%s: lot bought in %s and sold in %s; method B is undefined across currencies, so it repeats method A", label(t), p.lot.currency, t.Currency))
		}
		lm.TaxFreeFrom, lm.Status = TaxFreeFrom(lm.Acquired), classify(lm.Acquired, ms.date)

		known = known.Add(p.proceeds)
		s.Cost, s.CostMKD = s.Cost.Add(lm.Cost), s.CostMKD.Add(lm.CostMKD)
		s.GainMKD, s.GainMKDB = s.GainMKD.Add(lm.GainMKD), s.GainMKDB.Add(lm.GainMKDB)
		switch lm.Status {
		case StatusTaxable:
			s.TaxableGain = s.TaxableGain.Add(p.proceeds.Sub(p.lot.cost))
			s.TaxableGainMKD, s.TaxableGainMKDB = s.TaxableGainMKD.Add(lm.GainMKD), s.TaxableGainMKDB.Add(lm.GainMKDB)
		case StatusExempt:
			s.ExemptGainMKD = s.ExemptGainMKD.Add(lm.GainMKD)
		case StatusOutOfScope:
			s.OutOfScopeGainMKD = s.OutOfScopeGainMKD.Add(lm.GainMKD)
		}
		s.Lots = append(s.Lots, lm)
	}
	s.Gain = known.Sub(s.Cost)
	s.TaxMKD = taxOnGain(s.TaxableGainMKD)
	return s, nil
}

// Converts gross and net at the payment-date rate and caps the foreign tax credit at the 10% tax. Net stands in for an unknown gross.
func priceDividend(dv Dividend, date Date, rate rateFn) (DividendLine, error) {
	l := DividendLine{Dividend: dv, Date: date, Country: CountryFromISIN(dv.ISIN), IncomeType: IncomeTypeCapital, IncomeSubtype: IncomeSubtypeDividends}
	var estimated bool
	l.Gross, l.GrossCurrency, estimated = dv.gross()
	if estimated {
		l.Warnings = append(l.Warnings, "gross amount not reported: the net amount is used as gross, so tax may be understated and withheld tax is unknown")
	}
	gr, err := rate(l.GrossCurrency, date)
	if err != nil {
		return DividendLine{}, err
	}
	nr, err := rate(l.NetCurrency, date)
	if err != nil {
		return DividendLine{}, err
	}
	l.GrossRate, l.NetRate = gr, nr
	l.GrossMKD, l.NetMKD = l.Gross.Mul(gr), l.Net.Mul(nr)
	l.WithheldMKD = decimal.Max(decimal.Zero, l.GrossMKD.Sub(l.NetMKD))
	l.TaxIfCreditRejectedMKD = l.GrossMKD.Mul(taxRate)
	l.TaxIfCreditAcceptedMKD = l.TaxIfCreditRejectedMKD.Sub(decimal.Min(l.WithheldMKD, l.TaxIfCreditRejectedMKD))
	return l, nil
}

func (m *MonthReport) finish(today Date) {
	slices.SortStableFunc(m.Dividends, func(a, b DividendLine) int {
		return cmp.Or(a.PaidOn.Compare(b.PaidOn), cmp.Compare(a.Ticker, b.Ticker))
	})
	t := &m.Totals
	for _, s := range m.Sales {
		t.Sales++
		if !s.CostComplete {
			t.UnknownCostSales++
		}
		m.Warnings = append(m.Warnings, s.Warnings...)
		t.ProceedsMKD = t.ProceedsMKD.Add(s.ProceedsMKD)
		if s.TaxableGainMKD.IsPositive() {
			t.TaxableGainsMKD = t.TaxableGainsMKD.Add(s.TaxableGainMKD)
		} else {
			t.TaxableLossesMKD = t.TaxableLossesMKD.Add(s.TaxableGainMKD)
		}
		if s.TaxableGainMKDB.IsPositive() {
			t.TaxableGainsMKDB = t.TaxableGainsMKDB.Add(s.TaxableGainMKDB)
		}
		t.NetIfLossesOffsetMKD = t.NetIfLossesOffsetMKD.Add(s.TaxableGainMKD)
		t.ExemptGainsMKD = t.ExemptGainsMKD.Add(s.ExemptGainMKD)
		t.OutOfScopeGainsMKD = t.OutOfScopeGainsMKD.Add(s.OutOfScopeGainMKD)
	}
	t.TaxBaseMKD = t.TaxableGainsMKD.Mul(taxBaseShare)
	t.CapitalGainsTaxMKD = t.TaxBaseMKD.Mul(taxRate)
	t.TaxIfLossesOffsetMKD = taxOnGain(t.NetIfLossesOffsetMKD)
	t.CapitalGainsTaxMKDB = taxOnGain(t.TaxableGainsMKDB)

	for _, l := range m.Dividends {
		t.Dividends++
		m.Warnings = append(m.Warnings, l.Warnings...)
		t.DividendGrossMKD = t.DividendGrossMKD.Add(l.GrossMKD)
		t.DividendNetMKD = t.DividendNetMKD.Add(l.NetMKD)
		t.DividendWithheldMKD = t.DividendWithheldMKD.Add(l.WithheldMKD)
		t.DividendTaxIfCreditAcceptedMKD = t.DividendTaxIfCreditAcceptedMKD.Add(l.TaxIfCreditAcceptedMKD)
		t.DividendTaxIfCreditRejectedMKD = t.DividendTaxIfCreditRejectedMKD.Add(l.TaxIfCreditRejectedMKD)
	}
	t.TaxDueIfCreditAcceptedMKD = t.CapitalGainsTaxMKD.Add(t.DividendTaxIfCreditAcceptedMKD)
	t.TaxDueIfCreditRejectedMKD = t.CapitalGainsTaxMKD.Add(t.DividendTaxIfCreditRejectedMKD)

	m.HasEntries = len(m.Sales)+len(m.Dividends) > 0
	m.Overdue = m.HasEntries && today.After(m.FileBy)
}

func label(t Trade) string { return cmp.Or(t.Ticker, t.ISIN, t.Key) }
