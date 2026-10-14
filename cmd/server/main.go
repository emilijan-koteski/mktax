package main

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/emilijan-koteski/mktax/internal/buildinfo"
	"github.com/emilijan-koteski/mktax/nbrm"
)

func main() {
	if err := run(); err != nil {
		slog.Error("mktax", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := loadConfig(os.LookupEnv)
	if err != nil {
		return err
	}
	level := slog.LevelInfo
	if cfg.Debug {
		level = slog.LevelDebug
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	client, err := nbrm.New(cfg.NBRMURL)
	if err != nil {
		return err
	}
	store := nbrm.NewStore(client, cfg.CacheDir, log, nbrm.WithMissRetry(cfg.MissRetry))
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		return err
	}
	static, err := newStaticHandler(sub)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           requestLog(log, newMux(static, newRatesHandler(store, time.Now, 25*time.Second, log))),
		ReadHeaderTimeout: 10 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if cfg.WarmupFrom != "" {
		go warmup(ctx, store, cfg.WarmupFrom, time.Now, log)
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	log.Info("listening", "addr", cfg.Addr, "version", buildinfo.Version, "cache", cfg.CacheDir)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
