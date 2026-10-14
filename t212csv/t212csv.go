package t212csv

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/emilijan-koteski/mktax/mktax"
)

type File struct {
	Name string
	Data []byte
}

type Options struct {
	FundUnits map[string]bool
}

type Kind string

const (
	KindTrade            Kind = "trade"
	KindDividend         Kind = "dividend"
	KindDividendExcluded Kind = "dividend_excluded"
	KindInterest         Kind = "interest"
	KindIgnored          Kind = "ignored"
	KindCorporateAction  Kind = "corporate_action"
	KindUnknown          Kind = "unknown"
)

type Row struct {
	File     string
	Line     int
	ID       string
	Action   string
	Time     time.Time
	Ticker   string
	ISIN     string
	Name     string
	Amount   decimal.Decimal
	Currency string
	Notes    string
	Kind     Kind
}

type Flag struct {
	Code   string
	File   string
	Line   int
	ID     string
	Action string
	Ticker string
	Params map[string]string
}

type Instrument struct {
	Key              string
	Ticker           string
	ISIN             string
	Name             string
	Currency         string
	FundUnits        bool
	FundUnitsGuessed bool
	Buys             int
	Sells            int
	Dividends        int
}

type FileSummary struct {
	Name       string
	Rows       int
	Trades     int
	Dividends  int
	Interest   int
	Ignored    int
	Duplicates int
	Unknown    int
	First      time.Time
	Last       time.Time
	Columns    []string
}

type DividendDetail struct {
	Ref                 string
	File                string
	Line                int
	Withholding         decimal.Decimal
	WithholdingCurrency string
	GrossEstimated      bool
	GrossUnknown        bool
}

type Import struct {
	Trades          []mktax.Trade
	Dividends       []mktax.Dividend
	DividendDetails map[string]DividendDetail
	Excluded        []Row
	Interest        []Row
	Ignored         map[string]int
	Unknown         map[string]int
	Flags           []Flag
	Instruments     []Instrument
	Files           []FileSummary
	AccountCurrency string
	First           time.Time
	Last            time.Time
}

const (
	colAction  = "action"
	colTime    = "time (utc)"
	colTimeAlt = "time"
	colISIN    = "isin"
	colTicker  = "ticker"
	colName    = "name"
	colNotes   = "notes"
	colID      = "id"
	colShares  = "no. of shares"
	colPrice   = "price / share"
	colPriceCy = "currency (price / share)"
	colTotal   = "total"
	colTotalCy = "currency (total)"
	colWHT     = "withholding tax"
	colWHTCy   = "currency (withholding tax)"
)

var required = []struct{ key, name string }{
	{colAction, "Action"}, {colISIN, "ISIN"}, {colTicker, "Ticker"}, {colName, "Name"},
	{colShares, "No. of shares"}, {colPrice, "Price / share"}, {colPriceCy, "Currency (Price / share)"},
	{colTotal, "Total"}, {colTotalCy, "Currency (Total)"},
}

var timeLayouts = []string{"2006-01-02 15:04:05-07:00", "2006-01-02 15:04:05", time.RFC3339, time.DateOnly}

var taxableDividendTypes = map[string]bool{
	"dividend": true, "ordinary": true,
	"dividends paid by us corporations": true, "dividends paid by foreign corporations": true,
}

var ignoredActions = map[string]bool{
	"deposit": true, "withdrawal": true, "currency conversion": true, "card debit": true, "card credit": true,
	"new card cost": true, "spending cashback": true, "cashback": true, "result adjustment": true,
}

var corporateActions = map[string]bool{
	"stock split": true, "reverse split": true, "spinoff": true, "spin-off": true, "merger": true, "demerger": true,
}

var fundTokens = map[string]bool{"ETF": true, "UCITS": true, "ETC": true, "ETN": true}

func ClassifyAction(action string) (Kind, mktax.Side, string) {
	a := strings.ToLower(strings.Join(strings.Fields(action), " "))
	switch {
	case strings.HasSuffix(a, " buy"):
		return KindTrade, mktax.Buy, ""
	case strings.HasSuffix(a, " sell"):
		return KindTrade, mktax.Sell, ""
	case strings.HasPrefix(a, "dividend"):
		typ := "Dividend"
		if i := strings.Index(action, "("); i >= 0 {
			typ = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(action[i+1:]), ")"))
		}
		if taxableDividendTypes[strings.ToLower(typ)] {
			return KindDividend, "", typ
		}
		return KindDividendExcluded, "", typ
	case a == "interest on cash" || a == "lending interest":
		return KindInterest, "", ""
	case ignoredActions[a]:
		return KindIgnored, "", ""
	case corporateActions[a]:
		return KindCorporateAction, "", ""
	}
	return KindUnknown, "", ""
}

func IsFundUnits(name string) bool {
	for _, w := range strings.FieldsFunc(name, func(r rune) bool { return (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') }) {
		if fundTokens[strings.ToUpper(w)] {
			return true
		}
	}
	return false
}

type parser struct {
	opts        Options
	imp         Import
	seen        map[string]bool
	instruments map[string]*Instrument
	currencies  map[string]int
	years       map[int]bool
}

func Parse(files []File, opts Options) (Import, error) {
	p := &parser{
		opts:        opts,
		imp:         Import{Ignored: map[string]int{}, Unknown: map[string]int{}, DividendDetails: map[string]DividendDetail{}},
		seen:        map[string]bool{},
		instruments: map[string]*Instrument{},
		currencies:  map[string]int{},
		years:       map[int]bool{},
	}
	for _, f := range files {
		if err := p.file(f); err != nil {
			return Import{}, err
		}
	}
	return p.finish(), nil
}

func (p *parser) file(f File) error {
	r := csv.NewReader(bytes.NewReader(bytes.TrimPrefix(f.Data, []byte("\uFEFF"))))
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	r.TrimLeadingSpace = true
	header, err := r.Read()
	if err != nil {
		return fmt.Errorf("%s: %w", f.Name, err)
	}
	cols := map[string]int{}
	for i, h := range header {
		k := normHeader(h)
		if _, dup := cols[k]; !dup {
			cols[k] = i
		}
	}
	if _, ok := cols[colTime]; !ok {
		if i, ok := cols[colTimeAlt]; ok {
			cols[colTime] = i
		} else {
			return fmt.Errorf("%s: missing column %q", f.Name, "Time (UTC)")
		}
	}
	var missing []string
	for _, c := range required {
		if _, ok := cols[c.key]; !ok {
			missing = append(missing, fmt.Sprintf("%q", c.name))
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("%s: missing columns %s", f.Name, strings.Join(missing, ", "))
	}
	summary := FileSummary{Name: f.Name, Columns: header}
	fees := feeColumns(cols)
	for line := 2; ; line++ {
		rec, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("%s:%d: %w", f.Name, line, err)
		}
		if blank(rec) {
			continue
		}
		summary.Rows++
		if err := p.row(record{file: f.Name, line: line, rec: rec, cols: cols, fees: fees}, &summary); err != nil {
			return err
		}
	}
	p.imp.Files = append(p.imp.Files, summary)
	return nil
}

type record struct {
	file string
	line int
	rec  []string
	cols map[string]int
	fees []feeColumn
}

type feeColumn struct{ amount, currency int }

func (r record) get(key string) string {
	i, ok := r.cols[key]
	if !ok || i >= len(r.rec) {
		return ""
	}
	return strings.TrimSpace(r.rec[i])
}

func (r record) at(i int) string {
	if i < 0 || i >= len(r.rec) {
		return ""
	}
	return strings.TrimSpace(r.rec[i])
}

func (r record) fail(col string, err error) error {
	return fmt.Errorf("%s:%d: %s: %w", r.file, r.line, col, err)
}

func (p *parser) row(r record, summary *FileSummary) error {
	action := strings.TrimSpace(r.get(colAction))
	kind, side, dtype := ClassifyAction(action)
	id := r.get(colID)
	ts := r.get(colTime)
	when, timeErr := parseTime(ts)
	if timeErr == nil {
		p.observe(summary, when)
	}
	key, composite := id, false
	if key == "" {
		key, composite = strings.Join([]string{action, ts, r.get(colISIN), r.get(colTicker), r.get(colShares), r.get(colTotal), r.get(colTotalCy)}, "|"), true
	}
	if p.seen[key] {
		summary.Duplicates++
		if composite {
			p.flag("duplicate_row", r, action, nil)
		}
		return nil
	}
	p.seen[key] = true
	if strings.Contains(strings.ToLower(r.get(colNotes)+" "+action), "free share") {
		p.flag("free_shares", r, action, nil)
	}
	switch kind {
	case KindTrade:
		if timeErr != nil {
			return r.fail("Time (UTC)", timeErr)
		}
		summary.Trades++
		return p.trade(r, action, side, when, id)
	case KindDividend:
		if timeErr != nil {
			return r.fail("Time (UTC)", timeErr)
		}
		summary.Dividends++
		return p.dividend(r, when, id, dtype)
	case KindDividendExcluded:
		p.imp.Excluded = append(p.imp.Excluded, p.listed(r, action, when, id, kind))
		p.flag("excluded_dividend_type", r, action, map[string]string{"type": dtype})
	case KindInterest:
		summary.Interest++
		p.imp.Interest = append(p.imp.Interest, p.listed(r, action, when, id, kind))
	case KindIgnored:
		summary.Ignored++
		p.imp.Ignored[action]++
	case KindCorporateAction:
		p.flag("corporate_action", r, action, nil)
	default:
		summary.Unknown++
		p.imp.Unknown[action]++
		p.flag("unrecognised_action", r, action, nil)
	}
	return nil
}

func (p *parser) trade(r record, action string, side mktax.Side, when time.Time, id string) error {
	qty, hasQty, err := number(r.get(colShares))
	if err != nil {
		return r.fail("No. of shares", err)
	}
	if !hasQty || qty.IsZero() {
		p.flag("zero_quantity", r, action, nil)
		return nil
	}
	total, _, err := number(r.get(colTotal))
	if err != nil {
		return r.fail("Total", err)
	}
	ccy := strings.ToUpper(r.get(colTotalCy))
	if ccy == "" {
		return r.fail("Currency (Total)", errors.New("empty"))
	}
	fees := decimal.Zero
	for _, fc := range r.fees {
		amt, has, err := number(r.at(fc.amount))
		if err != nil || !has || amt.IsZero() {
			continue
		}
		if strings.ToUpper(r.at(fc.currency)) != ccy {
			p.flag("fee_other_currency", r, action, map[string]string{"currency": r.at(fc.currency)})
			continue
		}
		fees = fees.Add(amt.Abs())
	}
	ins := p.instrument(r)
	if side == mktax.Buy {
		ins.Buys++
	} else {
		ins.Sells++
	}
	p.currencies[ccy]++
	p.imp.Trades = append(p.imp.Trades, mktax.Trade{
		Ticker: ins.Ticker, ISIN: ins.ISIN, Name: ins.Name, Time: when, Side: side,
		Quantity: qty.Abs(), Value: total.Abs(), Currency: ccy, Fees: fees,
		Source: r.get(colNotes), Ref: id, FundUnits: ins.FundUnits,
	})
	return nil
}

// Derives the gross dividend from the broker's net figures: Total + withholding when both are in the account currency, else net per share × shares + withholding in the instrument currency, else unknown.
func (p *parser) dividend(r record, when time.Time, id, dtype string) error {
	total, _, err := number(r.get(colTotal))
	if err != nil {
		return r.fail("Total", err)
	}
	tccy := strings.ToUpper(r.get(colTotalCy))
	if tccy == "" {
		return r.fail("Currency (Total)", errors.New("empty"))
	}
	qty, hasQty, err := number(r.get(colShares))
	if err != nil {
		return r.fail("No. of shares", err)
	}
	price, hasPrice, err := number(r.get(colPrice))
	if err != nil {
		return r.fail("Price / share", err)
	}
	price, pccy := majorUnits(price, r.get(colPriceCy))
	wht, hasWHT, err := number(r.get(colWHT))
	if err != nil {
		return r.fail("Withholding tax", err)
	}
	wht, wccy := majorUnits(wht.Abs(), r.get(colWHTCy))
	if total.IsZero() {
		p.flag("zero_amount", r, r.get(colAction), nil)
		return nil
	}
	detail := DividendDetail{Ref: id, File: r.file, Line: r.line}
	gross, gccy := total.Abs(), tccy
	switch {
	case !hasWHT || wht.IsZero():
	case wccy == tccy:
		gross = total.Abs().Add(wht)
		detail.Withholding, detail.WithholdingCurrency = wht, wccy
	case wccy == pccy && hasPrice && hasQty:
		gross, gccy = price.Mul(qty.Abs()).Add(wht), pccy
		detail.Withholding, detail.WithholdingCurrency, detail.GrossEstimated = wht, wccy, true
		p.flag("gross_estimated", r, r.get(colAction), nil)
	default:
		gross, gccy = decimal.Zero, ""
		detail.Withholding, detail.WithholdingCurrency, detail.GrossUnknown = wht, wccy, true
		p.flag("gross_unknown", r, r.get(colAction), map[string]string{"withholding_currency": wccy})
	}
	ins := p.instrument(r)
	ins.Dividends++
	p.currencies[tccy]++
	dv := mktax.Dividend{
		Ticker: ins.Ticker, ISIN: ins.ISIN, Name: ins.Name, Type: dtype, PaidOn: when,
		Gross: gross, GrossCurrency: gccy, Net: total.Abs(), NetCurrency: tccy,
	}
	p.imp.Dividends = append(p.imp.Dividends, dv)
	p.imp.DividendDetails[DividendKey(dv)] = detail
	return nil
}

func DividendKey(d mktax.Dividend) string {
	return strings.Join([]string{d.Ticker, d.ISIN, d.PaidOn.UTC().Format(time.RFC3339Nano), d.Type, d.Net.String(), d.NetCurrency}, "|")
}

func (p *parser) listed(r record, action string, when time.Time, id string, kind Kind) Row {
	amount, _, _ := number(r.get(colTotal))
	return Row{
		File: r.file, Line: r.line, ID: id, Action: action, Time: when,
		Ticker: r.get(colTicker), ISIN: strings.ToUpper(r.get(colISIN)), Name: r.get(colName),
		Amount: amount, Currency: strings.ToUpper(r.get(colTotalCy)), Notes: r.get(colNotes), Kind: kind,
	}
}

func (p *parser) instrument(r record) *Instrument {
	isin, ticker := strings.ToUpper(r.get(colISIN)), r.get(colTicker)
	key := isin
	if key == "" {
		key = ticker
	}
	ins, ok := p.instruments[key]
	if !ok {
		ins = &Instrument{Key: key, ISIN: isin, Ticker: ticker}
		p.instruments[key] = ins
	}
	if ins.Ticker == "" {
		ins.Ticker = ticker
	}
	if ins.Name == "" {
		ins.Name = r.get(colName)
	}
	if ins.Currency == "" {
		_, ins.Currency = majorUnits(decimal.Zero, r.get(colPriceCy))
	}
	if override, ok := p.opts.FundUnits[key]; ok {
		ins.FundUnits, ins.FundUnitsGuessed = override, false
	} else {
		ins.FundUnits, ins.FundUnitsGuessed = IsFundUnits(ins.Name), true
	}
	return ins
}

func (p *parser) flag(code string, r record, action string, params map[string]string) {
	p.imp.Flags = append(p.imp.Flags, Flag{Code: code, File: r.file, Line: r.line, ID: r.get(colID), Action: action, Ticker: r.get(colTicker), Params: params})
}

func (p *parser) observe(summary *FileSummary, when time.Time) {
	if summary.First.IsZero() || when.Before(summary.First) {
		summary.First = when
	}
	if when.After(summary.Last) {
		summary.Last = when
	}
	p.years[when.UTC().Year()] = true
}

func (p *parser) finish() Import {
	imp := p.imp
	for i := range imp.Trades {
		imp.Trades[i].FundUnits = p.instruments[instrumentKey(imp.Trades[i].ISIN, imp.Trades[i].Ticker)].FundUnits
	}
	slices.SortStableFunc(imp.Trades, func(a, b mktax.Trade) int {
		return cmpChain(a.Time.Compare(b.Time), strings.Compare(a.Ticker, b.Ticker), strings.Compare(a.Ref, b.Ref))
	})
	slices.SortStableFunc(imp.Dividends, func(a, b mktax.Dividend) int {
		return cmpChain(a.PaidOn.Compare(b.PaidOn), strings.Compare(a.Ticker, b.Ticker))
	})
	for _, rows := range []*[]Row{&imp.Excluded, &imp.Interest} {
		slices.SortStableFunc(*rows, func(a, b Row) int { return cmpChain(a.Time.Compare(b.Time), strings.Compare(a.ID, b.ID)) })
	}
	for _, ins := range p.instruments {
		imp.Instruments = append(imp.Instruments, *ins)
	}
	slices.SortFunc(imp.Instruments, func(a, b Instrument) int { return strings.Compare(a.Key, b.Key) })
	for _, t := range imp.Trades {
		p.span(&imp, t.Time)
	}
	for _, d := range imp.Dividends {
		p.span(&imp, d.PaidOn)
	}
	imp.AccountCurrency = mode(p.currencies)
	if !imp.First.IsZero() {
		for y := imp.First.UTC().Year(); y < imp.Last.UTC().Year(); y++ {
			if !p.years[y] {
				imp.Flags = append(imp.Flags, Flag{Code: "possible_missing_year", Params: map[string]string{"year": fmt.Sprint(y)}})
			}
		}
	}
	slices.SortStableFunc(imp.Flags, func(a, b Flag) int { return cmpChain(strings.Compare(a.File, b.File), a.Line-b.Line) })
	return imp
}

func (p *parser) span(imp *Import, when time.Time) {
	if imp.First.IsZero() || when.Before(imp.First) {
		imp.First = when
	}
	if when.After(imp.Last) {
		imp.Last = when
	}
}

func instrumentKey(isin, ticker string) string {
	if isin != "" {
		return isin
	}
	return ticker
}

func feeColumns(cols map[string]int) []feeColumn {
	var out []feeColumn
	for name, i := range cols {
		if name == colWHT || strings.HasPrefix(name, "currency (") {
			continue
		}
		if !isFeeColumn(name) {
			continue
		}
		if c, ok := cols["currency ("+name+")"]; ok {
			out = append(out, feeColumn{amount: i, currency: c})
		}
	}
	slices.SortFunc(out, func(a, b feeColumn) int { return a.amount - b.amount })
	return out
}

func isFeeColumn(name string) bool {
	return strings.Contains(name, "fee") || strings.Contains(name, "duty") || strings.Contains(name, "levy") || strings.Contains(name, "transaction tax")
}

func normHeader(h string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimPrefix(h, "\uFEFF")), " "))
}

func blank(rec []string) bool {
	for _, f := range rec {
		if strings.TrimSpace(f) != "" {
			return false
		}
	}
	return true
}

func number(s string) (decimal.Decimal, bool, error) {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "not available") {
		return decimal.Zero, false, nil
	}
	d, err := decimal.NewFromString(strings.ReplaceAll(s, ",", ""))
	if err != nil {
		return decimal.Zero, false, fmt.Errorf("not a number: %q", s)
	}
	return d, true, nil
}

func parseTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	for _, l := range timeLayouts {
		if t, err := time.Parse(l, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognised time %q", s)
}

func majorUnits(amount decimal.Decimal, currency string) (decimal.Decimal, string) {
	c := strings.ToUpper(strings.TrimSpace(currency))
	if c == "GBX" || c == "GBP" && strings.TrimSpace(currency) == "GBp" {
		return amount.Div(decimal.NewFromInt(100)), "GBP"
	}
	return amount, c
}

func mode(counts map[string]int) string {
	best, n := "", 0
	for c, k := range counts {
		if k > n || k == n && c < best {
			best, n = c, k
		}
	}
	return best
}

func cmpChain(vals ...int) int {
	for _, v := range vals {
		if v != 0 {
			return v
		}
	}
	return 0
}
