# syntax=docker/dockerfile:1
FROM golang:1.26-bookworm AS wasm
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w -X github.com/emilijan-koteski/mktax/internal/buildinfo.Version=${VERSION}" -o /out/mktax.wasm ./cmd/wasm \
 && cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" /out/wasm_exec.js

FROM node:24-bookworm-slim AS web
RUN corepack enable
WORKDIR /src/web
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ .
COPY --from=wasm /out/ public/
RUN pnpm build

FROM golang:1.26-bookworm AS server
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /src/web/dist/ cmd/server/static/
RUN gzip -9 -c cmd/server/static/mktax.wasm > cmd/server/static/mktax.wasm.gz
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X github.com/emilijan-koteski/mktax/internal/buildinfo.Version=${VERSION}" -o /out/mktax ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=server /out/mktax /mktax
COPY --from=server --chown=nonroot:nonroot /tmp /data
ENV MKTAX_ADDR=:8080 MKTAX_CACHE_DIR=/data
VOLUME /data
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/mktax"]
