package httpx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Name        string
	BaseURL     string
	Header      map[string]string
	MinInterval time.Duration
	MaxRetries  int
	BaseBackoff time.Duration
	MaxBackoff  time.Duration
	HTTP        *http.Client
	Now         func() time.Time
	Sleep       func(context.Context, time.Duration) error
}

type Option func(*Config)

type Client struct {
	cfg  Config
	base *url.URL

	mu   sync.Mutex
	next time.Time
}

func New(cfg Config, opts ...Option) (*Client, error) {
	for _, o := range opts {
		o(&cfg)
	}
	u, err := url.Parse(cfg.BaseURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("%s: invalid base URL %q", cfg.Name, cfg.BaseURL)
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.BaseBackoff == 0 {
		cfg.BaseBackoff = time.Second
	}
	if cfg.MaxBackoff == 0 {
		cfg.MaxBackoff = 30 * time.Second
	}
	if cfg.HTTP == nil {
		cfg.HTTP = &http.Client{Timeout: 30 * time.Second}
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.Sleep == nil {
		cfg.Sleep = sleepCtx
	}
	return &Client{cfg: cfg, base: u}, nil
}

type APIError struct {
	Service    string
	Path       string
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	msg := fmt.Sprintf("%s GET %s: HTTP %d", e.Service, e.Path, e.StatusCode)
	if e.Body != "" {
		msg += ": " + e.Body
	}
	return msg
}

func IsStatus(err error, code int) bool {
	var ae *APIError
	return errors.As(err, &ae) && ae.StatusCode == code
}

// Performs a paced GET and retries 429 and 5xx responses with exponential backoff, honouring Retry-After. Other non-2xx statuses fail immediately with an APIError.
func (c *Client) Get(ctx context.Context, path string, q url.Values) ([]byte, error) {
	u := c.base.ResolveReference(&url.URL{Path: path, RawQuery: q.Encode()})
	for attempt := 0; ; attempt++ {
		if err := c.pace(ctx); err != nil {
			return nil, err
		}
		status, body, hdr, err := c.roundTrip(ctx, u.String(), path)
		if err != nil {
			return nil, err
		}
		if status >= 200 && status <= 299 {
			return body, nil
		}
		retryable := status == http.StatusTooManyRequests || status >= 500
		if !retryable || attempt >= c.cfg.MaxRetries {
			return nil, &APIError{Service: c.cfg.Name, Path: path, StatusCode: status, Body: truncate(string(body), 300)}
		}
		if err := c.cfg.Sleep(ctx, c.backoff(attempt, hdr)); err != nil {
			return nil, err
		}
	}
}

func (c *Client) roundTrip(ctx context.Context, rawURL, path string) (int, []byte, http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("%s: build request: %w", c.cfg.Name, err)
	}
	for k, v := range c.cfg.Header {
		req.Header.Set(k, v)
	}
	resp, err := c.cfg.HTTP.Do(req)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return 0, nil, nil, ctxErr
		}
		return 0, nil, nil, fmt.Errorf("%s GET %s: %w", c.cfg.Name, path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return 0, nil, nil, fmt.Errorf("%s: read response: %w", c.cfg.Name, err)
	}
	return resp.StatusCode, b, resp.Header, nil
}

func (c *Client) pace(ctx context.Context) error {
	if c.cfg.MinInterval <= 0 {
		return nil
	}
	c.mu.Lock()
	now := c.cfg.Now()
	at := c.next
	if at.Before(now) {
		at = now
	}
	c.next = at.Add(c.cfg.MinInterval)
	c.mu.Unlock()
	if wait := at.Sub(now); wait > 0 {
		return c.cfg.Sleep(ctx, wait)
	}
	return nil
}

func (c *Client) backoff(attempt int, hdr http.Header) time.Duration {
	d := c.cfg.BaseBackoff << attempt
	if s, err := strconv.Atoi(strings.TrimSpace(hdr.Get("Retry-After"))); err == nil {
		if ra := time.Duration(s) * time.Second; ra > d {
			d = ra
		}
	}
	return min(d, c.cfg.MaxBackoff)
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
