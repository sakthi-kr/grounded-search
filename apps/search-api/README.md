# Search API

The search API is the public Go service and the future policy-enforcement
boundary for GroundedSearch.

## Phase 1 functionality

Implemented endpoints:

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/healthz` | Process health |
| `GET` | `/readyz` | Phase 1 readiness |
| `GET` | `/v1/system/status` | Foundation status and dependency state |

The service currently reports the ML dependency as `not_configured`. Search,
retrieval, ACL enforcement, reranking, and answer generation are deliberately
deferred.

## Requirements

- Go 1.23 or newer

The project is currently developed with Go 1.26.5.

## Run locally

From this directory:

```bash
go run ./cmd/server
```

The default address is:

```text
http://localhost:8080
```

Test it from another terminal:

```bash
curl -i http://localhost:8080/healthz
curl -i http://localhost:8080/readyz
curl -i http://localhost:8080/v1/system/status
```

Stop the service with `Ctrl+C`. The server handles the interrupt through a
bounded graceful shutdown.

## Validate

```bash
gofmt -w .
go test ./...
go vet ./...
go build ./cmd/server
```

## Configuration

| Variable | Default |
|---|---|
| `SEARCH_API_HOST` | `0.0.0.0` |
| `SEARCH_API_PORT` | `8080` |
| `SEARCH_API_READ_HEADER_TIMEOUT` | `5s` |
| `SEARCH_API_READ_TIMEOUT` | `10s` |
| `SEARCH_API_WRITE_TIMEOUT` | `15s` |
| `SEARCH_API_IDLE_TIMEOUT` | `60s` |
| `SEARCH_API_SHUTDOWN_TIMEOUT` | `10s` |
| `LOG_LEVEL` | `info` |

Duration values use Go duration syntax such as `500ms`, `5s`, or `1m`.
