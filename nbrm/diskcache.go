package nbrm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type envelope[T any] struct {
	FetchedAt time.Time `json:"fetchedAt"`
	Data      T         `json:"data"`
}

func loadFile[T any](path string) (T, time.Time, bool) {
	var env envelope[T]
	b, err := os.ReadFile(path)
	if err != nil || json.Unmarshal(b, &env) != nil {
		var zero T
		return zero, time.Time{}, false
	}
	return env.Data, env.FetchedAt, true
}

func saveFile[T any](path string, v T, fetchedAt time.Time) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("cache dir: %w", err)
	}
	b, err := json.Marshal(envelope[T]{FetchedAt: fetchedAt, Data: v})
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.Write(b); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
