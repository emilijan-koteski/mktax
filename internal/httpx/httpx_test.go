package httpx

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

type sleeper struct{ got []time.Duration }

func (s *sleeper) sleep(_ context.Context, d time.Duration) error {
	s.got = append(s.got, d)
	return nil
}

func newTest(t *testing.T, h http.HandlerFunc, opts ...Option) (*Client, *sleeper) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	sl := &sleeper{}
	c, err := New(Config{
		Name: "svc", BaseURL: srv.URL, Header: map[string]string{"Accept": "application/json"},
		BaseBackoff: time.Second, MaxBackoff: 10 * time.Second, MaxRetries: 2, Sleep: sl.sleep,
	}, opts...)
	if err != nil {
		t.Fatal(err)
	}
	return c, sl
}

func TestGetSendsHeadersAndQuery(t *testing.T) {
	c, _ := newTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/json" || r.URL.Path != "/v1/x" || r.URL.Query().Get("a") != "b" {
			http.Error(w, "bad", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"n":7}`))
	})
	b, err := c.Get(context.Background(), "/v1/x", map[string][]string{"a": {"b"}})
	if err != nil || string(b) != `{"n":7}` {
		t.Fatalf("got %s, %v", b, err)
	}
}

func TestRejectsBaseURLWithoutHost(t *testing.T) {
	if _, err := New(Config{Name: "svc", BaseURL: "nbrm.mk"}); err == nil {
		t.Fatal("want error")
	}
}

func TestRetriesOn429ThenSucceeds(t *testing.T) {
	var calls atomic.Int32
	c, sl := newTest(t, func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`{}`))
	})
	if _, err := c.Get(context.Background(), "/x", nil); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 || len(sl.got) != 1 || sl.got[0] != time.Second {
		t.Fatalf("calls=%d sleeps=%v", calls.Load(), sl.got)
	}
}

func TestRetryAfterHeaderHonoured(t *testing.T) {
	var calls atomic.Int32
	c, sl := newTest(t, func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "5")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`{}`))
	})
	if _, err := c.Get(context.Background(), "/x", nil); err != nil {
		t.Fatal(err)
	}
	if len(sl.got) != 1 || sl.got[0] != 5*time.Second {
		t.Fatalf("sleeps=%v", sl.got)
	}
}

func TestGivesUpAfterMaxRetries(t *testing.T) {
	var calls atomic.Int32
	c, _ := newTest(t, func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	_, err := c.Get(context.Background(), "/x", nil)
	if !IsStatus(err, http.StatusServiceUnavailable) || calls.Load() != 3 {
		t.Fatalf("err=%v calls=%d", err, calls.Load())
	}
}

func TestNoRetryOn404(t *testing.T) {
	var calls atomic.Int32
	c, _ := newTest(t, func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		http.NotFound(w, nil)
	})
	_, err := c.Get(context.Background(), "/x", nil)
	var ae *APIError
	if !errors.As(err, &ae) || ae.StatusCode != http.StatusNotFound || calls.Load() != 1 {
		t.Fatalf("err=%v calls=%d", err, calls.Load())
	}
}

func TestErrorBodyIsTruncated(t *testing.T) {
	c, _ := newTest(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		for range 50 {
			_, _ = w.Write([]byte("0123456789"))
		}
	})
	_, err := c.Get(context.Background(), "/x", nil)
	var ae *APIError
	if !errors.As(err, &ae) || len([]rune(ae.Body)) != 301 {
		t.Fatalf("err=%v", err)
	}
}

func TestPacingSpacesRequests(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	c, sl := newTest(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{}`)) },
		func(cfg *Config) { cfg.MinInterval = 100 * time.Millisecond; cfg.Now = func() time.Time { return t0 } })
	for range 3 {
		if _, err := c.Get(context.Background(), "/x", nil); err != nil {
			t.Fatal(err)
		}
	}
	if len(sl.got) != 2 || sl.got[0] != 100*time.Millisecond || sl.got[1] != 200*time.Millisecond {
		t.Fatalf("sleeps=%v", sl.got)
	}
}

func TestContextCancellationWins(t *testing.T) {
	c, _ := newTest(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{}`)) })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.Get(ctx, "/x", nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}
