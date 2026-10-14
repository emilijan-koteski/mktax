# Даночко (mktax)

Даночко пресметува што треба да внесете во месечната е-ППД пресметка на УЈП од извозот (CSV) на Trading 212: капитални добивки по FIFO со двегодишно ослободување по лот, дивиденди со и без признавање на странскиот задржан данок, денарска противвредност по среден курс на НБРСМ, рокови за пријава и плаќање и преглед на отворените лотови.

Пресметката се извршува целосно во вашиот прелистувач (Go компилиран во WebAssembly). CSV датотеката никогаш не го напушта вашиот уред. Серверот само ја послужува апликацијата и ги кешира курсевите на НБРСМ. Нема сметки, нема база, нема аналитика.

**Ова не е даночен совет** и апликацијата не е поврзана со УЈП. Проверете ги сите износи пред да ги внесете во порталот.

## How it works

1. Drop one Trading 212 CSV export per calendar year (all years since the first purchase, so FIFO lots carry over).
2. The importer (`t212csv`) maps the rows to trades and dividends; the engine (`mktax`) runs FIFO, the two-year exemption, the 10% normed costs and the dividend credit; everything is converted at the NBRSM middle rate of each transaction date.
3. The browser asks the server only for exchange rates, one request per month and currency set. Nothing else is sent.
4. The result is a table per month matching the portal fields, a printable attachment and a CSV.

The rules and open interpretations are documented in [docs/rules.md](docs/rules.md); the portal walkthrough in [docs/portal.md](docs/portal.md); deployment in [docs/deploy.md](docs/deploy.md).

## Layout

```text
mktax/            tax engine: FIFO lots, exemptions, MKD conversion, e-ППД lines (stdlib + decimal)
rates/            NBRSM rate table shared by the server, the browser and the fixtures
nbrm/             NBRSM client and disk-cached store (server only)
t212csv/          Trading 212 CSV importer
internal/report/  JSON report model: portal rows, rounding, structured warnings
internal/bridge/  JSON API used by the WebAssembly module
cmd/wasm/         exposes globalThis.mktax in the browser
cmd/server/       serves the embedded app and GET /api/rates/{month}
web/              Vite + Svelte 5 + Tailwind app (Macedonian, strings in web/src/lib/i18n)
testdata/         synthetic exports, rate tables and the golden report
```

## Development

```sh
mise install
make dev
```

`make dev` builds the WebAssembly module, starts the Go server on :8080 and the Vite dev server on :5173 (which proxies `/api`).

| Target | What it does |
| --- | --- |
| `make test` | Go tests with the race detector, svelte-check and Vitest |
| `make test-wasm` | the pure packages and the bridge compiled to WebAssembly, run under Node |
| `make lint` | go vet, golangci-lint, the comment-density check, ESLint |
| `make build` | single binary `bin/mktax` with the web app and the gzipped wasm embedded |
| `make golden` | regenerate the golden report from the synthetic fixtures |
| `make docker` | the multi-stage image |

Code style: no comments except a header of at most two sentences above a complex function; `make lint` enforces it. Money is `decimal.Decimal` end to end and reaches the browser as strings.

## Privacy

The CSV is parsed and the tax is computed inside the browser. The server receives only `GET /api/rates/{month}?currencies=…` requests, keeps a cache of published exchange rates on disk, logs route patterns and status codes, and stores nothing else.

## License

MIT
