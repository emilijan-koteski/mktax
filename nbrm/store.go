package nbrm

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"github.com/emilijan-koteski/mktax/mktax"
	"github.com/emilijan-koteski/mktax/rates"
)

type Store struct {
	c         *Client
	dir       string
	log       *slog.Logger
	now       func() time.Time
	missRetry time.Duration

	mu      sync.Mutex
	days    rates.Table
	loaded  map[string]bool
	missed  map[string]time.Time
	fetchMu map[string]*sync.Mutex
}

type StoreOption func(*Store)

func WithMissRetry(d time.Duration) StoreOption { return func(s *Store) { s.missRetry = d } }

func WithNow(now func() time.Time) StoreOption { return func(s *Store) { s.now = now } }

func NewStore(c *Client, dir string, log *slog.Logger, opts ...StoreOption) *Store {
	if log == nil {
		log = slog.Default()
	}
	s := &Store{c: c, dir: dir, log: log, now: time.Now, days: rates.Table{}, loaded: map[string]bool{}, missed: map[string]time.Time{}, fetchMu: map[string]*sync.Mutex{}}
	for _, o := range opts {
		o(s)
	}
	return s
}

func (s *Store) Prefetch(ctx context.Context, days []time.Time) error {
	byMonth := map[string][]time.Time{}
	for _, d := range days {
		d = civil(d)
		m := d.Format(monthLayout)
		byMonth[m] = append(byMonth[m], d)
	}
	for _, m := range slices.Sorted(maps.Keys(byMonth)) {
		if err := s.prefetchMonth(ctx, m, byMonth[m]); err != nil {
			return err
		}
	}
	return nil
}

// Fetches a month's missing days in one request while holding only that month's lock, so other months stay readable during network I/O. Days still absent afterwards are remembered and not re-requested within the miss-retry window.
func (s *Store) prefetchMonth(ctx context.Context, month string, days []time.Time) error {
	lock := s.monthLock(month)
	lock.Lock()
	defer lock.Unlock()
	missing := s.missingDays(month, days)
	if len(missing) == 0 {
		return nil
	}
	from, to := slices.MinFunc(missing, time.Time.Compare), slices.MaxFunc(missing, time.Time.Compare)
	got, err := s.c.Rates(ctx, from, to)
	if err != nil {
		return fmt.Errorf("NBRM rates %s..%s: %w", from.Format(rates.DayLayout), to.Format(rates.DayLayout), err)
	}
	s.log.Debug("fetched NBRM rates", "from", from.Format(rates.DayLayout), "to", to.Format(rates.DayLayout), "days", len(got))
	s.mu.Lock()
	s.days.Merge(got)
	now := s.now()
	for _, d := range missing {
		if _, ok := s.days[rates.DayKey(d)]; !ok {
			s.missed[rates.DayKey(d)] = now
		}
	}
	part := s.monthSlice(month)
	s.mu.Unlock()
	s.saveMonth(month, part)
	return nil
}

func (s *Store) monthLock(month string) *sync.Mutex {
	s.mu.Lock()
	defer s.mu.Unlock()
	l, ok := s.fetchMu[month]
	if !ok {
		l = &sync.Mutex{}
		s.fetchMu[month] = l
	}
	return l
}

func (s *Store) missingDays(month string, days []time.Time) []time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.loadMonth(month)
	now := s.now()
	seen := map[string]bool{}
	var out []time.Time
	for _, d := range days {
		k := rates.DayKey(d)
		if seen[k] {
			continue
		}
		seen[k] = true
		if _, ok := s.days[k]; ok {
			continue
		}
		if at, ok := s.missed[k]; ok && now.Sub(at) < s.missRetry {
			continue
		}
		out = append(out, d)
	}
	return out
}

func (s *Store) Rate(currency string, day time.Time) (decimal.Decimal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.days.Lookup(currency, day)
}

func (s *Store) Slice(from, to time.Time, currencies []string) rates.Table {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := rates.Table{}
	last := civil(to)
	for d := civil(from); !d.After(last); d = d.AddDate(0, 0, 1) {
		k := rates.DayKey(d)
		if byCcy, ok := s.days[k]; ok {
			out[k] = byCcy
		}
	}
	return out.Filter(currencies)
}

func (s *Store) Month(ctx context.Context, year int, month time.Month, currencies []string) (rates.Table, []string, error) {
	first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1)
	limit := civil(s.now().In(mktax.Location()).AddDate(0, 0, 1))
	var days []time.Time
	for d := first; !d.After(last) && !d.After(limit); d = d.AddDate(0, 0, 1) {
		days = append(days, d)
	}
	if err := s.Prefetch(ctx, days); err != nil {
		return nil, nil, err
	}
	tbl := s.Slice(first, last, currencies)
	return tbl, tbl.Missing(first, last), nil
}

func (s *Store) file(month string) string { return filepath.Join(s.dir, "rates-"+month+".json") }

func (s *Store) loadMonth(month string) {
	if s.loaded[month] {
		return
	}
	s.loaded[month] = true
	if v, _, ok := loadFile[rates.Table](s.file(month)); ok {
		s.days.Merge(v)
	}
}

func (s *Store) monthSlice(month string) rates.Table {
	part := rates.Table{}
	for k, v := range s.days {
		if strings.HasPrefix(k, month) {
			part[k] = v
		}
	}
	return part
}

func (s *Store) saveMonth(month string, part rates.Table) {
	if err := saveFile(s.file(month), part, s.now()); err != nil {
		s.log.Warn("could not write NBRM rate cache", "month", month, "err", err)
	}
}

func civil(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
