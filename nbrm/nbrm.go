package nbrm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/emilijan-koteski/mktax/internal/httpx"
	"github.com/emilijan-koteski/mktax/rates"
)

const DefaultBaseURL = "https://www.nbrm.mk"

const (
	ratesPath   = "/KLServiceNOV/GetExchangeRate"
	queryLayout = "02.01.2006"
	datumLayout = "2006-01-02T15:04:05"
	monthLayout = "2006-01"
)

type Client struct{ h *httpx.Client }

func New(baseURL string, opts ...httpx.Option) (*Client, error) {
	h, err := httpx.New(httpx.Config{
		Name:        "nbrm",
		BaseURL:     baseURL,
		Header:      map[string]string{"Accept": "application/json"},
		MinInterval: 500 * time.Millisecond,
	}, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{h: h}, nil
}

type rawRate struct {
	Datum  string      `json:"datum"`
	Oznaka string      `json:"oznaka"`
	Nomin  json.Number `json:"nomin"`
	Sreden json.Number `json:"sreden"`
}

func (c *Client) Rates(ctx context.Context, from, to time.Time) (rates.Table, error) {
	q := url.Values{"StartDate": {from.Format(queryLayout)}, "EndDate": {to.Format(queryLayout)}, "format": {"json"}}
	b, err := c.h.Get(ctx, ratesPath, q)
	if err != nil {
		return nil, err
	}
	var raw []rawRate
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, fmt.Errorf("nbrm: decode rates: %w", err)
	}
	out := rates.Table{}
	for _, r := range raw {
		d, err := time.Parse(datumLayout, r.Datum)
		if err != nil {
			return nil, fmt.Errorf("nbrm: rate date %q: %w", r.Datum, err)
		}
		mid, err := decimal.NewFromString(r.Sreden.String())
		if err != nil {
			return nil, fmt.Errorf("nbrm: %s rate %q on %s: %w", r.Oznaka, r.Sreden, r.Datum, err)
		}
		units, err := r.Nomin.Int64()
		if err != nil || units <= 0 {
			return nil, fmt.Errorf("nbrm: %s units %q on %s are not a positive integer", r.Oznaka, r.Nomin, r.Datum)
		}
		out.Set(rates.DayKey(d), strings.ToUpper(strings.TrimSpace(r.Oznaka)), rates.Rate{Units: units, Middle: mid})
	}
	return out, nil
}
