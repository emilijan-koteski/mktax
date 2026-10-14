package nbrm

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/emilijan-koteski/mktax/internal/httpx"
)

func noPacing(c *httpx.Config) { c.MinInterval = 0; c.MaxRetries = 1; c.BaseBackoff = time.Millisecond }

func day(s string) time.Time {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestClientRates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if r.URL.Path != "/KLServiceNOV/GetExchangeRate" || q.Get("StartDate") != "01.05.2026" ||
			q.Get("EndDate") != "02.05.2026" || q.Get("format") != "json" {
			http.Error(w, "unexpected "+r.URL.String(), http.StatusBadRequest)
			return
		}
		b, _ := os.ReadFile("testdata/rates.json")
		_, _ = w.Write(b)
	}))
	defer srv.Close()
	c, err := New(srv.URL, noPacing)
	if err != nil {
		t.Fatal(err)
	}
	got, err := c.Rates(context.Background(), day("2026-05-01"), day("2026-05-02"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || len(got["2026-05-01"]) != 3 {
		t.Fatalf("got %+v", got)
	}
	usd := got["2026-05-02"]["USD"]
	if usd.Units != 1 || usd.Middle.String() != "50.1234" || usd.PerUnit().String() != "50.1234" {
		t.Errorf("usd = %+v", usd)
	}
	jpy := got["2026-05-01"]["JPY"]
	if jpy.Units != 100 || jpy.PerUnit().String() != "0.3525" {
		t.Errorf("jpy = %+v per unit %s", jpy, jpy.PerUnit())
	}
}

func TestClientRejectsBadPayload(t *testing.T) {
	for name, body := range map[string]string{
		"not json":   "<html>maintenance</html>",
		"bad date":   `[{"datum":"01.05.2026","oznaka":"USD","nomin":1,"sreden":50}]`,
		"zero units": `[{"datum":"2026-05-01T00:00:00","oznaka":"USD","nomin":0,"sreden":50}]`,
	} {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, body) }))
			defer srv.Close()
			c, err := New(srv.URL, noPacing)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := c.Rates(context.Background(), day("2026-05-01"), day("2026-05-01")); err == nil {
				t.Fatal("want error")
			}
		})
	}
}

type synth struct {
	srv    *httptest.Server
	calls  atomic.Int32
	mu     sync.Mutex
	ranges []string
	gate   func(from time.Time)
}

func synthServer(t *testing.T, lastPublished string) *synth {
	t.Helper()
	s := &synth{}
	s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.calls.Add(1)
		q := r.URL.Query()
		s.mu.Lock()
		s.ranges = append(s.ranges, q.Get("StartDate")+"-"+q.Get("EndDate"))
		s.mu.Unlock()
		from, err1 := time.Parse("02.01.2006", q.Get("StartDate"))
		to, err2 := time.Parse("02.01.2006", q.Get("EndDate"))
		if err1 != nil || err2 != nil {
			http.Error(w, "bad dates", http.StatusBadRequest)
			return
		}
		if s.gate != nil {
			s.gate(from)
		}
		limit := day(lastPublished)
		rows := []map[string]any{}
		for d := from; !d.After(to) && !d.After(limit); d = d.AddDate(0, 0, 1) {
			datum := d.Format("2006-01-02T15:04:05")
			rows = append(rows,
				map[string]any{"datum": datum, "oznaka": "USD", "nomin": 1, "sreden": json.Number("50." + d.Format("02"))},
				map[string]any{"datum": datum, "oznaka": "EUR", "nomin": 1, "sreden": json.Number("61.5")})
		}
		if len(rows) == 0 {
			http.Error(w, `{"title":"Not Found","status":404}`, http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(rows)
	}))
	t.Cleanup(s.srv.Close)
	return s
}

func (s *synth) rangeList() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return strings.Join(s.ranges, ",")
}

func newStore(t *testing.T, url, dir string, opts ...StoreOption) *Store {
	t.Helper()
	c, err := New(url, noPacing)
	if err != nil {
		t.Fatal(err)
	}
	return NewStore(c, dir, slog.New(slog.NewTextHandler(io.Discard, nil)), opts...)
}

func TestStorePrefetchAndCache(t *testing.T) {
	srv := synthServer(t, "2026-06-30")
	dir := t.TempDir()
	s := newStore(t, srv.srv.URL, dir)
	ctx := context.Background()

	need := []time.Time{day("2026-05-03"), day("2026-05-20"), day("2026-06-02"), day("2026-05-03")}
	if err := s.Prefetch(ctx, need); err != nil {
		t.Fatal(err)
	}
	if srv.calls.Load() != 2 || srv.rangeList() != "03.05.2026-20.05.2026,02.06.2026-02.06.2026" {
		t.Fatalf("calls = %d ranges = %v", srv.calls.Load(), srv.rangeList())
	}
	for _, tt := range []struct{ ccy, day, want string }{
		{"USD", "2026-05-03", "50.03"},
		{"usd", "2026-05-20", "50.2"},
		{"EUR", "2026-06-02", "61.5"},
		{"MKD", "1999-01-01", "1"},
		{"USD", "2026-05-11", "50.11"},
	} {
		got, err := s.Rate(tt.ccy, day(tt.day))
		if err != nil || got.String() != tt.want {
			t.Errorf("Rate(%s, %s) = %s, %v; want %s", tt.ccy, tt.day, got, err, tt.want)
		}
	}
	if _, err := s.Rate("GBP", day("2026-05-03")); err == nil || !strings.Contains(err.Error(), "GBP") {
		t.Errorf("unknown currency err = %v", err)
	}
	if _, err := s.Rate("USD", day("2026-04-01")); err == nil || !strings.Contains(err.Error(), "2026-04-01") {
		t.Errorf("unfetched day err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "rates-2026-05.json")); err != nil {
		t.Errorf("cache file: %v", err)
	}

	if err := s.Prefetch(ctx, need); err != nil || srv.calls.Load() != 2 {
		t.Fatalf("refetch: calls = %d err = %v", srv.calls.Load(), err)
	}
	s2 := newStore(t, srv.srv.URL, dir)
	if err := s2.Prefetch(ctx, need); err != nil || srv.calls.Load() != 2 {
		t.Fatalf("disk cache: calls = %d err = %v", srv.calls.Load(), err)
	}
	if got, err := s2.Rate("USD", day("2026-05-20")); err != nil || got.String() != "50.2" {
		t.Errorf("from disk = %s, %v", got, err)
	}
}

func TestStoreUnpublishedDay(t *testing.T) {
	srv := synthServer(t, "2026-05-10")
	s := newStore(t, srv.srv.URL, t.TempDir())
	if err := s.Prefetch(context.Background(), []time.Time{day("2026-05-09"), day("2026-05-12")}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Rate("USD", day("2026-05-09")); err != nil {
		t.Errorf("published day: %v", err)
	}
	_, err := s.Rate("USD", day("2026-05-12"))
	if err == nil || !strings.Contains(err.Error(), "not published") {
		t.Errorf("unpublished day err = %v", err)
	}
	if err := s.Prefetch(context.Background(), []time.Time{day("2026-05-12")}); err != nil {
		t.Fatalf("prefetch of only unpublished days: %v", err)
	}
	if srv.calls.Load() != 2 {
		t.Errorf("calls = %d, want 2", srv.calls.Load())
	}
}

func TestStoreMissRetryThrottle(t *testing.T) {
	srv := synthServer(t, "2026-05-10")
	now := time.Date(2026, time.May, 11, 9, 0, 0, 0, time.UTC)
	s := newStore(t, srv.srv.URL, t.TempDir(), WithMissRetry(time.Hour), WithNow(func() time.Time { return now }))
	ctx := context.Background()
	for i := range 2 {
		if err := s.Prefetch(ctx, []time.Time{day("2026-05-12")}); err != nil {
			t.Fatal(err)
		}
		if srv.calls.Load() != 1 {
			t.Fatalf("round %d: calls = %d, want 1", i, srv.calls.Load())
		}
	}
	now = now.Add(61 * time.Minute)
	if err := s.Prefetch(ctx, []time.Time{day("2026-05-12")}); err != nil {
		t.Fatal(err)
	}
	if srv.calls.Load() != 2 {
		t.Fatalf("after the window: calls = %d, want 2", srv.calls.Load())
	}
	if err := s.Prefetch(ctx, []time.Time{day("2026-05-09")}); err != nil || srv.calls.Load() != 3 {
		t.Fatalf("a different missing day is not throttled: calls = %d err = %v", srv.calls.Load(), err)
	}
}

func TestStorePrefetchError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "down", http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	s := newStore(t, srv.URL, t.TempDir())
	err := s.Prefetch(context.Background(), []time.Time{day("2026-05-09")})
	if err == nil || !strings.Contains(err.Error(), "503") {
		t.Fatalf("err = %v", err)
	}
}

func TestStoreSlice(t *testing.T) {
	srv := synthServer(t, "2026-06-30")
	s := newStore(t, srv.srv.URL, t.TempDir())
	if err := s.Prefetch(context.Background(), []time.Time{day("2026-05-03"), day("2026-05-20")}); err != nil {
		t.Fatal(err)
	}
	got := s.Slice(day("2026-05-02"), day("2026-05-04"), []string{"usd"})
	if strings.Join(got.Days(), ",") != "2026-05-03,2026-05-04" {
		t.Fatalf("days = %v", got.Days())
	}
	if len(got["2026-05-03"]) != 1 || got["2026-05-03"]["USD"].Middle.String() != "50.03" {
		t.Fatalf("slice = %v", got)
	}
	if all := s.Slice(day("2026-05-03"), day("2026-05-03"), nil); len(all["2026-05-03"]) != 2 {
		t.Fatalf("unfiltered = %v", all)
	}
}

func TestStoreMonth(t *testing.T) {
	srv := synthServer(t, "2026-05-10")
	now := time.Date(2026, time.May, 9, 10, 0, 0, 0, time.UTC)
	s := newStore(t, srv.srv.URL, t.TempDir(), WithNow(func() time.Time { return now }))
	ctx := context.Background()

	tbl, missing, err := s.Month(ctx, 2026, time.May, []string{"USD"})
	if err != nil {
		t.Fatal(err)
	}
	if srv.rangeList() != "01.05.2026-10.05.2026" {
		t.Fatalf("ranges = %s", srv.rangeList())
	}
	if len(tbl) != 10 || len(missing) != 21 || missing[0] != "2026-05-11" || missing[20] != "2026-05-31" {
		t.Fatalf("days = %d missing = %v", len(tbl), missing)
	}
	if _, ok := tbl["2026-05-01"]["EUR"]; ok {
		t.Fatal("currency filter ignored")
	}
	if _, _, err := s.Month(ctx, 2026, time.May, []string{"USD"}); err != nil || srv.calls.Load() != 1 {
		t.Fatalf("second call: calls = %d err = %v", srv.calls.Load(), err)
	}
	tbl, missing, err = s.Month(ctx, 2026, time.April, nil)
	if err != nil || len(missing) != 0 || len(tbl) != 30 || len(tbl["2026-04-01"]) != 2 {
		t.Fatalf("april: days = %d missing = %v err = %v", len(tbl), missing, err)
	}
}

func TestStoreConcurrentMonthsDoNotSerialize(t *testing.T) {
	srv := synthServer(t, "2026-06-30")
	dir := t.TempDir()
	warm := newStore(t, srv.srv.URL, dir)
	if err := warm.Prefetch(context.Background(), []time.Time{day("2026-06-02")}); err != nil {
		t.Fatal(err)
	}
	entered, release := make(chan struct{}), make(chan struct{})
	srv.gate = func(from time.Time) {
		if from.Month() == time.May {
			close(entered)
			<-release
		}
	}
	s := newStore(t, srv.srv.URL, dir)
	done := make(chan error, 1)
	go func() { done <- s.Prefetch(context.Background(), []time.Time{day("2026-05-03")}) }()
	<-entered
	finished := make(chan error, 1)
	go func() { finished <- s.Prefetch(context.Background(), []time.Time{day("2026-06-02")}) }()
	select {
	case err := <-finished:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("a cached month waited on another month's network call")
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestDiskCacheRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "x.json")
	at := time.Unix(1000, 0).UTC()
	if err := saveFile(path, []string{"a", "b"}, at); err != nil {
		t.Fatal(err)
	}
	v, got, ok := loadFile[[]string](path)
	if !ok || len(v) != 2 || !got.Equal(at) {
		t.Fatalf("v=%v at=%v ok=%v", v, got, ok)
	}
	if _, _, ok := loadFile[[]string](filepath.Join(t.TempDir(), "missing")); ok {
		t.Fatal("missing file should miss")
	}
}
