# Deploying

The app is one container: a Go binary that serves the embedded web app and proxies NBRSM exchange rates. It keeps no user data. The only state is the rate cache, a directory of small JSON files, one per month, that never changes once a month is complete.

## Docker

```sh
make docker
docker run --rm -p 8080:8080 -v mktax-data:/data mktax:latest
```

| Variable | Default | Meaning |
| --- | --- | --- |
| `MKTAX_ADDR` | `:8080` | listen address |
| `MKTAX_CACHE_DIR` | `/data` in the image, `$XDG_CACHE_HOME/mktax` otherwise | rate cache directory |
| `MKTAX_NBRM_URL` | `https://www.nbrm.mk` | upstream rate service |
| `MKTAX_RATES_MISS_RETRY` | `1h` | how long to wait before asking NBRSM again for a day it has not published yet |
| `MKTAX_WARMUP_FROM` | `2019-01` | prefetch rates from this month on start; empty disables |
| `MKTAX_DEBUG` | unset | `1` turns on debug logs |

Logs go to stderr and contain the method, route pattern, status and duration of each request. Query strings, amounts and client addresses are never logged.

## Dokploy

1. Create an application from this repository with the Dockerfile build type.
2. Mount a volume at `/data`.
3. Point the health check at `GET /healthz`.
4. Add the domain(s). The app uses only relative URLs, so moving to another domain is a DNS change plus a redirect on the old one.

## Endpoints

| Route | Purpose |
| --- | --- |
| `GET /` | the web app |
| `GET /api/rates/{YYYY-MM}?currencies=USD,EUR` | NBRSM middle rates for one month, `{"month","from","to","days":{"YYYY-MM-DD":{"USD":{"units":1,"middle":"…"}}},"missing":[…],"complete":bool}`; complete months are cached by clients as immutable |
| `GET /healthz` | liveness |

Errors are JSON: `{"error":{"code":"bad_request|upstream|timeout|not_found|method_not_allowed","message":"…"}}`.
