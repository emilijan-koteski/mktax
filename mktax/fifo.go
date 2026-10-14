package mktax

import (
	"cmp"
	"fmt"
	"slices"
	"time"

	"github.com/shopspring/decimal"
)

type lot struct {
	at       time.Time
	date     Date
	qty      decimal.Decimal
	cost     decimal.Decimal
	currency string
}

type portion struct {
	lot      lot
	proceeds decimal.Decimal
	unknown  bool
}

type matchedSale struct {
	trade    Trade
	date     Date
	portions []portion
}

func (ms matchedSale) short() decimal.Decimal {
	for _, p := range ms.portions {
		if p.unknown {
			return p.lot.qty
		}
	}
	return decimal.Zero
}

func (ms matchedSale) oversellWarning() string {
	return fmt.Sprintf("%s sale on %s exceeds the recorded buys by %s; history may be incomplete", label(ms.trade), ms.date, ms.short())
}

type position struct {
	key, ticker, isin, name string
	lots                    []lot
}

type ledger struct {
	sales []matchedSale
	open  map[string]*position
}

// Runs FIFO over every trade, oldest lot first per instrument. Partial lots keep their date and a pro-rata share of the cost.
func match(trades []Trade) (ledger, error) {
	ts, err := normalize(trades)
	if err != nil {
		return ledger{}, err
	}
	led := ledger{open: map[string]*position{}}
	for _, t := range ts {
		p := led.open[t.Key]
		if p == nil {
			p = &position{key: t.Key}
			led.open[t.Key] = p
		}
		p.ticker, p.isin, p.name = cmp.Or(t.Ticker, p.ticker), cmp.Or(t.ISIN, p.isin), cmp.Or(t.Name, p.name)
		switch t.Side {
		case Buy:
			p.lots = append(p.lots, lot{at: t.Time, date: DateOf(t.Time), qty: t.Quantity, cost: t.Value, currency: t.Currency})
		case Sell:
			led.sales = append(led.sales, p.consume(t))
		}
	}
	return led, nil
}

func normalize(trades []Trade) ([]Trade, error) {
	isinByTicker := map[string]string{}
	for _, t := range trades {
		if t.ISIN != "" && t.Ticker != "" {
			isinByTicker[t.Ticker] = t.ISIN
		}
	}
	out := make([]Trade, len(trades))
	for i, t := range trades {
		if t.ISIN == "" {
			t.ISIN = isinByTicker[t.Ticker]
		}
		if t.Key == "" {
			t.Key = cmp.Or(t.ISIN, t.Ticker)
		}
		if err := t.validate(); err != nil {
			return nil, fmt.Errorf("mktax: trade %d (%s): %w", i, label(t), err)
		}
		out[i] = t
	}
	slices.SortStableFunc(out, func(a, b Trade) int {
		return cmp.Or(a.Time.Compare(b.Time), cmp.Compare(sideRank(a.Side), sideRank(b.Side)))
	})
	return out, nil
}

func sideRank(s Side) int {
	if s == Buy {
		return 0
	}
	return 1
}

// Takes the sold quantity from the oldest lots; whatever exceeds the recorded buys becomes an unknown-cost portion.
func (p *position) consume(t Trade) matchedSale {
	ms := matchedSale{trade: t, date: DateOf(t.Time)}
	need := t.Quantity
	for need.GreaterThan(dust) && len(p.lots) > 0 {
		l := p.lots[0]
		take := decimal.Min(need, l.qty)
		part := l
		if l.qty.Sub(take).LessThanOrEqual(dust) {
			p.lots = p.lots[1:]
		} else {
			part.qty, part.cost = take, l.cost.Mul(take).Div(l.qty)
			p.lots[0].qty, p.lots[0].cost = l.qty.Sub(take), l.cost.Sub(part.cost)
		}
		need = need.Sub(part.qty)
		ms.portions = append(ms.portions, portion{lot: part})
	}
	if need.GreaterThan(dust) {
		ms.portions = append(ms.portions, portion{lot: lot{qty: need, currency: t.Currency}, unknown: true})
	}
	rest := t.Value
	for i := range ms.portions {
		if i == len(ms.portions)-1 {
			ms.portions[i].proceeds = rest
			break
		}
		share := t.Value.Mul(ms.portions[i].lot.qty).Div(t.Quantity)
		ms.portions[i].proceeds, rest = share, rest.Sub(share)
	}
	return ms
}

func (led ledger) openPositions(today Date) []OpenPosition {
	var out []OpenPosition
	for _, p := range led.open {
		if len(p.lots) == 0 {
			continue
		}
		op := OpenPosition{
			Key: p.key, Ticker: p.ticker, ISIN: p.isin, Name: p.name, Currency: p.lots[0].currency,
			EarliestAcquired: p.lots[0].date, OldestLotTaxFreeFrom: TaxFreeFrom(p.lots[0].date),
		}
		for _, l := range p.lots {
			ol := OpenLot{Acquired: l.date, Quantity: l.qty, Cost: l.cost, TaxFreeFrom: TaxFreeFrom(l.date)}
			op.Quantity, op.Cost = op.Quantity.Add(l.qty), op.Cost.Add(l.cost)
			if today.Compare(ol.TaxFreeFrom) >= 0 {
				op.TaxFreeQuantityNow = op.TaxFreeQuantityNow.Add(l.qty)
			}
			op.Lots = append(op.Lots, ol)
		}
		out = append(out, op)
	}
	slices.SortFunc(out, func(a, b OpenPosition) int { return cmp.Or(cmp.Compare(a.Ticker, b.Ticker), cmp.Compare(a.Key, b.Key)) })
	return out
}
