# whatismyip

Single-binary Go service: HTTP/TLS/QUIC server that returns the client's IP, geolocation, ASN, headers, TCP port scan, and DNS discovery. Uses [gin](https://github.com/gin-gonic/gin) + [httprouter](https://github.com/julienschmidt/httprouter). Requires Go >= 1.25.

## Commands

```sh
make build                                           # CGO_ENABLED=0 -> ./whatismyip (GOOS/GOARCH for cross-compile)
make build-all                                       # cross-compile all 5 platforms into dist/
make unit-test                                       # go test -count=1 -race -short -cover ./...
make integration-test                                # requires Docker; runs integration-tests/ module
make test                                            # unit-test + integration-test
make lint                                            # gofmt -l + golangci-lint + shadow (auto-installs tools)
make docker-build                                    # local single-arch build (root Dockerfile)
make docker-run                                      # builds dev image + runs with test DBs, ports :8080 :8081 :9100
```

**Order and quirks**: `lint -> unit-test -> integration-test`. Integration tests use testcontainers-go (needs Docker). They are a separate Go module with its own `go.mod` — run from the `integration-tests/` directory. They use `test/Dockerfile` (distinct from the root `Dockerfile`).

## Architecture

```
cmd/whatismyip.go          # entrypoint: parses flags, wires servers, runs
├── server/server.go       # Manager: Start/Stop/SIGHUP-reload for all server types
├── server/tcp.go          # plain HTTP
├── server/tls.go          # TLS/HTTP2
├── server/quic.go         # HTTP/3 (requires TLS server)
├── server/dns.go          # micro-DNS server for discovery
├── server/prometheus.go   # separate metrics endpoint
├── router/setup.go        # Gin route registration
│   ├── router/generic.go  # /, /client-port, /all, /json
│   ├── router/geo.go      # /geo/*, /asn/*
│   ├── router/headers.go  # /headers, /:header
│   ├── router/dns.go      # DNS discovery HTTP handler
│   ├── router/port_scanner.go  # /scan/tcp/:port
│   └── router/templates.go     # embedded HTML template
├── service/geo.go         # GeoIP lookup service (MaxMind MMDB)
├── service/port_scanner.go
├── resolver/setup.go      # authoritative micro-DNS server (miekg/dns)
├── internal/setting/      # flag parsing + resolver YAML config
├── internal/httputils/    # header filtering and formatting
├── internal/metrics/      # Prometheus counters/histograms (opt-in)
├── internal/validator/uuid/
├── internal/core/version.go
└── models/geo.go          # GeoDB wrapper around oschwald/maxminddb-golang
```

## Key facts

- **DNS discovery**: enabled via `-resolver test/resolver.yml`. The resolver is an authoritative DNS server answering for a subdomain. Clients prove their DNS provider by resolving `<uuid>.dns.<domain>`.
- **Geo**: both `-geoip2-city` and `-geoip2-asn` required together (or omitted). Uses test DBs in `test/`.
- **Trusted headers**: `-trusted-header X-Real-IP` for proxy mode. When set without `-trusted-port-header`, client port shows "unknown".
- **SIGHUP**: reloads GeoIP databases and restarts all servers without full stop.
- **HTTP/3**: requires `-tls-bind`; reuses its port for UDP.
- **Prometheus**: separate process/port via `-metrics-bind`. Metrics prefixed `whatismyip_*`.
- **Port scan**: enabled by default; disable with `-disable-scan`.
- **Embedded template**: `router/templates.go` has the default `home` template. `-template` overrides.
- **Version**: injected at build via `-ldflags="-X 'github.com/dcarrillo/whatismyip/internal/core.Version=${VERSION}'"`.

## Testing quirks

- Integration tests need Docker (testcontainers-go) and will take longer. They run against a real containerized build.
- Unit tests in `internal/setting/`, `internal/httputils/`, `internal/metrics/`, `internal/validator/uuid/`, `models/`, `router/`, and `service/` use test DBs from `test/`.
- The router test reads `test/` resources by relative path — run from repo root.

## Lint

Uses `golangci-lint` v2 config (`.golangci.yaml`) with strict `revive` rules and `goimports` as formatter. One exception: `internal/metrics/` gets a pass on `var-naming` (package name conflicts with stdlib `metrics`). Also runs `shadow` and `gofmt -l -d`.
