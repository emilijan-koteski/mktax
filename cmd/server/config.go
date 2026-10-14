package main

import (
	"cmp"
	"fmt"
	"path/filepath"
	"time"

	"github.com/emilijan-koteski/mktax/nbrm"
)

type config struct {
	Addr       string
	CacheDir   string
	NBRMURL    string
	MissRetry  time.Duration
	WarmupFrom string
	Debug      bool
}

func loadConfig(lookup func(string) (string, bool)) (config, error) {
	get := func(k string) string {
		v, _ := lookup(k)
		return v
	}
	cfg := config{Addr: ":8080", NBRMURL: nbrm.DefaultBaseURL, MissRetry: time.Hour, WarmupFrom: "2019-01"}
	cfg.CacheDir = filepath.Join(cmp.Or(get("XDG_CACHE_HOME"), filepath.Join(get("HOME"), ".cache")), "mktax")
	if v := get("MKTAX_CACHE_DIR"); v != "" {
		cfg.CacheDir = v
	}
	if v := get("MKTAX_ADDR"); v != "" {
		cfg.Addr = v
	}
	if v := get("MKTAX_NBRM_URL"); v != "" {
		cfg.NBRMURL = v
	}
	if v := get("MKTAX_RATES_MISS_RETRY"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return config{}, fmt.Errorf("MKTAX_RATES_MISS_RETRY: %w", err)
		}
		cfg.MissRetry = d
	}
	if v, ok := lookup("MKTAX_WARMUP_FROM"); ok {
		if v != "" {
			if _, err := time.Parse("2006-01", v); err != nil {
				return config{}, fmt.Errorf("MKTAX_WARMUP_FROM: want YYYY-MM, got %q", v)
			}
		}
		cfg.WarmupFrom = v
	}
	cfg.Debug = get("MKTAX_DEBUG") == "1" || get("MKTAX_DEBUG") == "true"
	return cfg, nil
}
