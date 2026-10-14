package main

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"errors"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
)

//go:embed all:static
var staticFS embed.FS

type asset struct {
	body  []byte
	gz    []byte
	ctype string
	etag  string
	cache string
}

func newStaticHandler(fsys fs.FS) (http.Handler, error) {
	files := map[string]*asset{}
	var gzipped []string
	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || strings.HasPrefix(path.Base(p), ".") {
			return nil
		}
		if strings.HasSuffix(p, ".gz") {
			gzipped = append(gzipped, p)
			return nil
		}
		b, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		files["/"+p] = &asset{body: b, ctype: contentType(p), etag: etagOf(b), cache: cacheControl(p)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for _, p := range gzipped {
		if a, ok := files["/"+strings.TrimSuffix(p, ".gz")]; ok {
			if a.gz, err = fs.ReadFile(fsys, p); err != nil {
				return nil, err
			}
		}
	}
	if files["/index.html"] == nil {
		return nil, errors.New("static: no index.html")
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if p == "/" {
			p = "/index.html"
		}
		a, ok := files[p]
		if !ok {
			http.NotFound(w, r)
			return
		}
		body, etag := a.body, a.etag
		if a.gz != nil {
			w.Header().Set("Vary", "Accept-Encoding")
			if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				body, etag = a.gz, strings.TrimSuffix(a.etag, `"`)+`-gz"`
				w.Header().Set("Content-Encoding", "gzip")
			}
		}
		w.Header().Set("Content-Type", a.ctype)
		w.Header().Set("Cache-Control", a.cache)
		w.Header().Set("ETag", etag)
		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}), nil
}

func contentType(p string) string {
	if t := mime.TypeByExtension(path.Ext(p)); t != "" {
		return t
	}
	return "application/octet-stream"
}

func cacheControl(p string) string {
	if strings.HasPrefix(p, "assets/") {
		return "public, max-age=31536000, immutable"
	}
	return "no-cache"
}

func etagOf(b []byte) string {
	sum := sha256.Sum256(b)
	return `"` + hex.EncodeToString(sum[:8]) + `"`
}
