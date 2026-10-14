package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"testing/fstest"
	"time"

	"github.com/shopspring/decimal"

	"github.com/emilijan-koteski/mktax/rates"
)

var testNow = func() time.Time { return time.Date(2026, time.October, 14, 12, 0, 0, 0, time.UTC) }

type fakeStore struct {
	calls    atomic.Int32
	lastCcys []string
	missing  []string
	err      error
	block    bool
}

func (f *fakeStore) Month(ctx context.Context, y int, m time.Month, ccys []string) (rates.Table, []string, error) {
	f.calls.Add(1)
	f.lastCcys = ccys
	if f.block {
		<-ctx.Done()
		return nil, nil, ctx.Err()
	}
	if f.err != nil {
		return nil, nil, f.err
	}
	tbl := rates.Table{}
	first := time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
	for d := first; d.Month() == m; d = d.AddDate(0, 0, 1) {
		k := rates.DayKey(d)
		for _, c := range ccys {
			tbl.Set(k, c, rates.Rate{Units: 1, Middle: decimal.NewFromInt(50)})
		}
		if len(ccys) == 0 {
			tbl.Set(k, "USD", rates.Rate{Units: 1, Middle: decimal.NewFromInt(50)})
		}
	}
	for _, d := range f.missing {
		delete(tbl, d)
	}
	return tbl, append([]string{}, f.missing...), nil
}

func discard() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func ratesServer(t *testing.T, store *fakeStore, timeout time.Duration) *httptest.Server {
	t.Helper()
	mux := newMux(http.NotFoundHandler(), newRatesHandler(store, testNow, timeout, discard()))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

type ratesBody struct {
	Month    string                           `json:"month"`
	From     string                           `json:"from"`
	To       string                           `json:"to"`
	Days     map[string]map[string]rates.Rate `json:"days"`
	Missing  []string                         `json:"missing"`
	Complete bool                             `json:"complete"`
}

func get(t *testing.T, url string, headers ...string) (*http.Response, []byte) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	b, _ := io.ReadAll(resp.Body)
	return resp, b
}

func TestRatesMonthHappyPath(t *testing.T) {
	store := &fakeStore{}
	srv := ratesServer(t, store, time.Second)
	resp, b := get(t, srv.URL+"/api/rates/2026-04?currencies=usd,EUR")
	if resp.StatusCode != 200 || !strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
		t.Fatalf("status %d type %s: %s", resp.StatusCode, resp.Header.Get("Content-Type"), b)
	}
	var body ratesBody
	if err := json.Unmarshal(b, &body); err != nil {
		t.Fatal(err)
	}
	if body.Month != "2026-04" || body.From != "2026-04-01" || body.To != "2026-04-30" || len(body.Days) != 30 || !body.Complete || len(body.Missing) != 0 {
		t.Fatalf("body = %+v", body)
	}
	if day := body.Days["2026-04-15"]; len(day) != 2 || day["USD"].Middle.String() != "50" {
		t.Fatalf("day = %+v", day)
	}
	if strings.Join(store.lastCcys, ",") != "USD,EUR" {
		t.Fatalf("currencies passed = %v", store.lastCcys)
	}
	if cc := resp.Header.Get("Cache-Control"); cc != "public, max-age=31536000, immutable" {
		t.Fatalf("cache-control = %q", cc)
	}
	if !bytes.Contains(b, []byte(`"missing":[]`)) {
		t.Fatalf("missing must be an empty array: %s", b)
	}
}

func TestRatesMonthIncomplete(t *testing.T) {
	store := &fakeStore{missing: []string{"2026-10-15", "2026-10-16"}}
	srv := ratesServer(t, store, time.Second)
	resp, b := get(t, srv.URL+"/api/rates/2026-10?currencies=USD")
	var body ratesBody
	_ = json.Unmarshal(b, &body)
	if resp.StatusCode != 200 || body.Complete || len(body.Missing) != 2 || len(body.Days) != 29 {
		t.Fatalf("status %d body %+v", resp.StatusCode, body)
	}
	if cc := resp.Header.Get("Cache-Control"); cc != "public, max-age=900" {
		t.Fatalf("cache-control = %q", cc)
	}
}

func TestRatesValidation(t *testing.T) {
	srv := ratesServer(t, &fakeStore{}, time.Second)
	cases := []struct {
		path string
		want int
	}{
		{"/api/rates/2026-4", 400},
		{"/api/rates/2009-12", 400},
		{"/api/rates/2026-11", 400},
		{"/api/rates/2026-10", 200},
		{"/api/rates/2010-01", 200},
		{"/api/rates/2026-04?currencies=US$", 400},
		{"/api/rates/2026-04?currencies=" + manyCurrencies(21), 400},
		{"/api/rates/2026-04?currencies=" + manyCurrencies(20), 200},
		{"/api/rates/2026-04?currencies=", 200},
		{"/api/rates/2026-04?currencies=usd,usd", 200},
	}
	for _, c := range cases {
		resp, b := get(t, srv.URL+c.path)
		if resp.StatusCode != c.want {
			t.Errorf("%s: status %d, want %d: %s", c.path, resp.StatusCode, c.want, b)
		}
		if c.want == 400 && !bytes.Contains(b, []byte(`"error":{"code":"bad_request"`)) {
			t.Errorf("%s: body %s", c.path, b)
		}
	}
}

func manyCurrencies(n int) string {
	var out []string
	for i := range n {
		out = append(out, string(rune('A'+i/26))+string(rune('A'+i%26))+"X")
	}
	return strings.Join(out, ",")
}

func TestRatesUpstreamDown(t *testing.T) {
	srv := ratesServer(t, &fakeStore{err: errors.New("secret-upstream-detail")}, time.Second)
	resp, b := get(t, srv.URL+"/api/rates/2026-04")
	if resp.StatusCode != 502 || !bytes.Contains(b, []byte(`"code":"upstream"`)) || bytes.Contains(b, []byte("secret-upstream-detail")) {
		t.Fatalf("status %d body %s", resp.StatusCode, b)
	}
}

func TestRatesTimeout(t *testing.T) {
	srv := ratesServer(t, &fakeStore{block: true}, 50*time.Millisecond)
	resp, b := get(t, srv.URL+"/api/rates/2026-04")
	if resp.StatusCode != 504 || !bytes.Contains(b, []byte(`"code":"timeout"`)) {
		t.Fatalf("status %d body %s", resp.StatusCode, b)
	}
}

func gz(t *testing.T, s string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, _ = w.Write([]byte(s))
	_ = w.Close()
	return buf.Bytes()
}

func staticServer(t *testing.T) *httptest.Server {
	t.Helper()
	fsys := fstest.MapFS{
		"index.html":          {Data: []byte("<!doctype html><title>x</title>")},
		"assets/app.abc12.js": {Data: []byte("console.log(1)")},
		"mktax.wasm":          {Data: []byte("\x00asm")},
		"mktax.wasm.gz":       {Data: gz(t, "\x00asm")},
		"wasm_exec.js":        {Data: []byte("globalThis.Go = 1")},
		".gitkeep":            {Data: []byte("")},
	}
	static, err := newStaticHandler(fsys)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(newMux(static, http.NotFoundHandler()))
	t.Cleanup(srv.Close)
	return srv
}

func TestStaticHandler(t *testing.T) {
	srv := staticServer(t)
	resp, b := get(t, srv.URL+"/")
	if resp.StatusCode != 200 || !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") || resp.Header.Get("Cache-Control") != "no-cache" || resp.Header.Get("ETag") == "" || !bytes.Contains(b, []byte("<title>x")) {
		t.Fatalf("index: %d %v %s", resp.StatusCode, resp.Header, b)
	}
	etag := resp.Header.Get("ETag")
	if resp, _ = get(t, srv.URL+"/", "If-None-Match", etag); resp.StatusCode != 304 {
		t.Fatalf("revalidate: %d", resp.StatusCode)
	}
	if resp, _ = get(t, srv.URL+"/assets/app.abc12.js"); resp.Header.Get("Cache-Control") != "public, max-age=31536000, immutable" || !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/javascript") {
		t.Fatalf("asset: %v", resp.Header)
	}
	resp, b = get(t, srv.URL+"/mktax.wasm", "Accept-Encoding", "gzip, br")
	if resp.Header.Get("Content-Encoding") != "gzip" || resp.Header.Get("Content-Type") != "application/wasm" || resp.Header.Get("Vary") != "Accept-Encoding" || resp.Header.Get("Cache-Control") != "no-cache" {
		t.Fatalf("wasm gz: %v", resp.Header)
	}
	r, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if plain, _ := io.ReadAll(r); string(plain) != "\x00asm" {
		t.Fatalf("gz body = %q", plain)
	}
	if resp, b = get(t, srv.URL+"/mktax.wasm", "Accept-Encoding", "identity"); resp.Header.Get("Content-Encoding") != "" || string(b) != "\x00asm" {
		t.Fatalf("wasm identity: %v %q", resp.Header, b)
	}
	for _, p := range []string{"/nope", "/mktax.wasm.gz", "/.gitkeep", "/assets/"} {
		if resp, _ = get(t, srv.URL+p); resp.StatusCode != 404 {
			t.Errorf("%s: %d", p, resp.StatusCode)
		}
	}
	req, _ := http.NewRequest(http.MethodHead, srv.URL+"/", nil)
	hr, err := http.DefaultClient.Do(req)
	if err != nil || hr.StatusCode != 200 {
		t.Fatalf("head: %v %v", hr, err)
	}
}

func TestSecurityHeaders(t *testing.T) {
	srv := staticServer(t)
	for _, p := range []string{"/", "/api/nope", "/healthz"} {
		resp, _ := get(t, srv.URL+p)
		csp := resp.Header.Get("Content-Security-Policy")
		if !strings.Contains(csp, "'wasm-unsafe-eval'") || !strings.Contains(csp, "frame-ancestors 'none'") || resp.Header.Get("X-Content-Type-Options") != "nosniff" || resp.Header.Get("Referrer-Policy") != "no-referrer" {
			t.Errorf("%s: headers %v", p, resp.Header)
		}
	}
}

func TestHealthzAndAPIFallbacks(t *testing.T) {
	srv := ratesServer(t, &fakeStore{}, time.Second)
	if resp, b := get(t, srv.URL+"/healthz"); resp.StatusCode != 200 || string(b) != "ok\n" {
		t.Fatalf("healthz: %d %q", resp.StatusCode, b)
	}
	if resp, b := get(t, srv.URL+"/api/nope"); resp.StatusCode != 404 || !bytes.Contains(b, []byte(`"code":"not_found"`)) {
		t.Fatalf("api 404: %d %s", resp.StatusCode, b)
	}
	resp, err := http.Post(srv.URL+"/api/rates/2026-04", "text/plain", nil)
	if err != nil || resp.StatusCode != 405 {
		t.Fatalf("post: %v %v", resp.StatusCode, err)
	}
}

func TestRequestLogHasNoQuery(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, nil))
	h := requestLog(log, newMux(http.NotFoundHandler(), newRatesHandler(&fakeStore{}, testNow, time.Second, discard())))
	srv := httptest.NewServer(h)
	defer srv.Close()
	get(t, srv.URL+"/api/rates/2026-04?currencies=USD")
	out := buf.String()
	if !strings.Contains(out, "/api/rates/{month}") || strings.Contains(out, "currencies") || strings.Contains(out, "2026-04") || !strings.Contains(out, "status=200") {
		t.Fatalf("log = %s", out)
	}
}

func TestLoadConfig(t *testing.T) {
	env := func(m map[string]string) func(string) (string, bool) {
		return func(k string) (string, bool) { v, ok := m[k]; return v, ok }
	}
	cfg, err := loadConfig(env(map[string]string{"HOME": "/home/x"}))
	if err != nil || cfg.Addr != ":8080" || cfg.CacheDir != "/home/x/.cache/mktax" || cfg.NBRMURL != "https://www.nbrm.mk" || cfg.MissRetry != time.Hour || cfg.WarmupFrom != "2019-01" || cfg.Debug {
		t.Fatalf("defaults = %+v err = %v", cfg, err)
	}
	cfg, err = loadConfig(env(map[string]string{"XDG_CACHE_HOME": "/c", "MKTAX_ADDR": ":9", "MKTAX_NBRM_URL": "http://x", "MKTAX_RATES_MISS_RETRY": "2h", "MKTAX_WARMUP_FROM": "", "MKTAX_DEBUG": "1"}))
	if err != nil || cfg.Addr != ":9" || cfg.CacheDir != "/c/mktax" || cfg.NBRMURL != "http://x" || cfg.MissRetry != 2*time.Hour || cfg.WarmupFrom != "" || !cfg.Debug {
		t.Fatalf("overrides = %+v err = %v", cfg, err)
	}
	if cfg, err = loadConfig(env(map[string]string{"MKTAX_CACHE_DIR": "/data"})); err != nil || cfg.CacheDir != "/data" {
		t.Fatalf("cache dir = %+v err = %v", cfg, err)
	}
	if _, err = loadConfig(env(map[string]string{"MKTAX_RATES_MISS_RETRY": "soon"})); err == nil {
		t.Fatal("bad duration should fail")
	}
	if _, err = loadConfig(env(map[string]string{"MKTAX_WARMUP_FROM": "2019"})); err == nil {
		t.Fatal("bad warm-up month should fail")
	}
}

func TestWarmup(t *testing.T) {
	store := &fakeStore{}
	warmup(context.Background(), store, "2026-08", testNow, discard())
	if store.calls.Load() != 3 {
		t.Fatalf("calls = %d, want 3", store.calls.Load())
	}
	store = &fakeStore{err: errors.New("down")}
	warmup(context.Background(), store, "2026-09", testNow, discard())
	if store.calls.Load() != 2 {
		t.Fatalf("calls after errors = %d, want 2", store.calls.Load())
	}
}
