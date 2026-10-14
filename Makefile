GO ?= go
GOROOT := $(shell $(GO) env GOROOT)
MODULE := github.com/emilijan-koteski/mktax
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X $(MODULE)/internal/buildinfo.Version=$(VERSION)
STATIC := cmd/server/static
WASM_PKGS := ./mktax/... ./rates/... ./t212csv/... ./internal/report/... ./internal/bridge/... ./cmd/wasm/...

.PHONY: dev dev-server dev-web wasm web build test test-wasm golden lint docker clean

wasm:
	GOOS=js GOARCH=wasm $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o web/public/mktax.wasm ./cmd/wasm
	cp "$(GOROOT)/lib/wasm/wasm_exec.js" web/public/wasm_exec.js

web: wasm
	pnpm -C web install --frozen-lockfile
	pnpm -C web build

build: web
	rm -rf $(STATIC)
	mkdir -p $(STATIC)
	touch $(STATIC)/.gitkeep
	cp -R web/dist/. $(STATIC)/
	gzip -9 -c $(STATIC)/mktax.wasm > $(STATIC)/mktax.wasm.gz
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o bin/mktax ./cmd/server

dev: wasm
	$(MAKE) -j2 dev-server dev-web

dev-server:
	MKTAX_DEBUG=1 $(GO) run ./cmd/server

dev-web:
	pnpm -C web install
	pnpm -C web dev

test:
	$(GO) test -race ./...
	pnpm -C web check
	pnpm -C web test

test-wasm:
	PATH="$(GOROOT)/lib/wasm:$$PATH" GOOS=js GOARCH=wasm $(GO) test $(WASM_PKGS)

golden:
	$(GO) test ./internal/bridge -run TestCalculateGolden -update
	cp testdata/golden/*.json web/tests/fixtures/

lint:
	$(GO) vet ./...
	GOOS=js GOARCH=wasm $(GO) vet ./cmd/wasm/...
	golangci-lint run ./...
	$(GO) run ./tools/checkcomments ./...
	pnpm -C web lint

docker:
	docker build --build-arg VERSION=$(VERSION) -t mktax:$(VERSION) -t mktax:latest .

clean:
	rm -rf bin web/dist web/public/mktax.wasm web/public/mktax.wasm.gz web/public/wasm_exec.js $(STATIC)/*
	touch $(STATIC)/.gitkeep
