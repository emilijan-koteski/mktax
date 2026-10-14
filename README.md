# Даночко (mktax)

Даночко пресметува што треба да внесете во месечната е-ППД пресметка на УЈП од извозот (CSV) на Trading 212: капитални добивки по FIFO со двегодишно ослободување по лот, дивиденди со и без признавање на странскиот задржан данок, денарска противвредност по среден курс на НБРСМ, рокови за пријава и плаќање и преглед на отворените лотови.

Пресметката се извршува целосно во вашиот прелистувач (Go компилиран во WebAssembly). CSV датотеката никогаш не го напушта вашиот уред. Серверот само ја послужува апликацијата и ги кешира курсевите на НБРСМ. Нема сметки, нема база, нема аналитика.

**Ова не е даночен совет** и апликацијата не е поврзана со УЈП. Проверете ги сите износи пред да ги внесете во порталот.

## Development

```sh
mise install
make dev
```

`make test`, `make lint` and `make build` must be green before every commit. `make build` produces a single binary `bin/mktax` with the web app embedded.

## Deployment

`make docker` builds a multi-stage image. Run it with a volume mounted at `/data` for the exchange-rate cache and point a health check at `GET /healthz`. The app uses only relative URLs, so no domain is configured anywhere.

## License

MIT
