package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/emilijan-koteski/mktax/mktax"
	"github.com/emilijan-koteski/mktax/rates"
)

const (
	minRatesMonth = "2010-01"
	maxCurrencies = 20
)

type monthStore interface {
	Month(ctx context.Context, year int, month time.Month, currencies []string) (rates.Table, []string, error)
}

type ratesResponse struct {
	Month    string      `json:"month"`
	From     string      `json:"from"`
	To       string      `json:"to"`
	Days     rates.Table `json:"days"`
	Missing  []string    `json:"missing"`
	Complete bool        `json:"complete"`
}

func newRatesHandler(store monthStore, now func() time.Time, timeout time.Duration, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := r.PathValue("month")
		first, err := time.Parse("2006-01", m)
		current := now().In(mktax.Location()).Format("2006-01")
		if err != nil || len(m) != 7 || m < minRatesMonth || m > current {
			writeJSONError(w, http.StatusBadRequest, "bad_request", "month must be YYYY-MM between "+minRatesMonth+" and "+current)
			return
		}
		ccys, ok := parseCurrencies(r.URL.Query().Get("currencies"))
		if !ok {
			writeJSONError(w, http.StatusBadRequest, "bad_request", "currencies must be up to 20 three-letter codes")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		tbl, missing, err := store.Month(ctx, first.Year(), first.Month(), ccys)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				writeJSONError(w, http.StatusGatewayTimeout, "timeout", "NBRSM did not answer in time")
				return
			}
			log.Warn("NBRSM rates failed", "month", m, "err", err)
			writeJSONError(w, http.StatusBadGateway, "upstream", "NBRSM did not answer")
			return
		}
		if tbl == nil {
			tbl = rates.Table{}
		}
		if missing == nil {
			missing = []string{}
		}
		last := first.AddDate(0, 1, -1)
		if len(missing) == 0 {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "public, max-age=900")
		}
		writeJSON(w, http.StatusOK, ratesResponse{
			Month: m, From: rates.DayKey(first), To: rates.DayKey(last), Days: tbl, Missing: missing, Complete: len(missing) == 0,
		})
	})
}

func parseCurrencies(q string) ([]string, bool) {
	var out []string
	seen := map[string]bool{}
	for _, c := range strings.Split(q, ",") {
		c = strings.ToUpper(strings.TrimSpace(c))
		if c == "" {
			continue
		}
		if len(c) != 3 || strings.Trim(c, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") != "" {
			return nil, false
		}
		if !seen[c] {
			seen[c] = true
			out = append(out, c)
		}
	}
	if len(out) > maxCurrencies {
		return nil, false
	}
	return out, true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": msg}})
}

func warmup(ctx context.Context, store monthStore, from string, now func() time.Time, log *slog.Logger) {
	start, err := time.Parse("2006-01", from)
	if err != nil {
		log.Warn("rate warm-up skipped", "from", from, "err", err)
		return
	}
	today := now().In(mktax.Location())
	end := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	for m := start; !m.After(end) && ctx.Err() == nil; m = m.AddDate(0, 1, 0) {
		if _, _, err := store.Month(ctx, m.Year(), m.Month(), nil); err != nil {
			log.Warn("rate warm-up", "month", m.Format("2006-01"), "err", err)
		}
	}
	log.Info("rate warm-up finished", "from", from, "to", end.Format("2006-01"))
}
