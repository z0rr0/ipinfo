# AGENTS.md

Guidance for AI coding agents working in this repository.

## Overview

IPINFO is a small Go web service that reports geolocation/network info about the
caller's IP as plain text, JSON, XML or HTML. It uses the MaxMind GeoLite2 City
database and the standard-library HTTP server (no framework). User docs: `README.md`,
API: `api.md`.

## Commands

```bash
make prepare    # test fixtures: copy config.example.json to /tmp/ipinfo_test.json, download /tmp/GeoLite2-City.mmdb
make test       # lint + prepare + go test -race -cover ./...
make gh         # CI target: check_fmt + prepare + go vet + tests
make lint       # gofmt + go vet (blocking); golangci-lint, govulncheck, staticcheck, gosec (non-blocking)
make build      # lint, then build ./ipinfo with version ldflags (GO_LDFLAGS)
make tools      # install linters as Go tool dependencies
make start / stop / restart   # run a local instance with config.example.json
make docker / docker-push     # host-arch image / multi-arch push to Docker Hub
```

- Tests need the `/tmp` fixtures from `make prepare`; a bare `go test ./...` fails without them.
- Single test: `make prepare && go test -race -run TestName ./conf/...`
- `make lint` succeeds even if the external linters report issues — read their output and keep it clean.

## Architecture

`ipinfo.go` → `conf.Cfg.Info()` (client IP → GeoLite2 lookup → `IPInfo`) → formatter in `handle/`.

- `ipinfo.go` — HTTP server, graceful shutdown, routing via a `map[string]handler` keyed by the
  trimmed path (`/short`, `/compact`, `/json`, `/xml`, `/html`, `/full`, `/version`); anything else
  goes to `handle.TextHandler`. `Version`/`Revision`/`BuildDate` are injected with `-ldflags -X`.
- `conf/` — JSON config, client IP detection, GeoLite2 reader behind an optional LRU cache,
  and the `IPInfo` response model (struct tags drive JSON/XML output).
- `handle/` — one handler per format; `index.html`/`full.html` are `//go:embed`-ed `html/template`s.

## Gotchas

- Config is file-only: flags `-config <path>` (default `config.json`) and `-version`; no env vars.
- `readConfig` only accepts `/tmp/ipinfo_test.json`, paths under `/data/conf` (Docker) or under
  the working dir — keep this restriction.
- Client IP is taken from the `ip_header` request header (falls back to `RemoteAddr` when empty).
  If `ip_itself` is set, private/loopback IPs are replaced with it (e.g. requests from other
  containers on the same host).
- `cache_size <= 0` disables the LRU cache.
- Time zones are embedded via the `time/tzdata` import; the Docker image has no tzdata.

## Conventions

- Go 1.27 (toolchain go1.27.1). Lint config `golangci.yml`: `default: all` with a disable list,
  cyclomatic complexity ≤ 20, line length ≤ 160, snake_case JSON tags.
- New `.go` files start with the BSD-3-Clause copyright header.
- New config fields go into `config.example.json` and are documented in `README.md`.
