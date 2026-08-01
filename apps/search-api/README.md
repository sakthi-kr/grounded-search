# Search API

The search API is the public Go service and the policy-enforcement boundary for
GroundedSearch.

## Phase 1 functionality

Implemented endpoints:

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/healthz` | Process health |
| `GET` | `/readyz` | Search API readiness |
| `GET` | `/v1/system/status` | Combined search API and ML service status |

The system-status endpoint calls the Python ML service through the versioned
contract in `api/internal/ml-service-v1.yaml`.

It reports:

- `healthy` when the ML service returns a valid model-info response;
- `degraded` when the dependency times out, is unavailable, or violates the
  internal response contract.

A dependency failure does not crash the Go service.

## Requirements

- Go 1.23 or newer
- The Python ML service for a healthy combined status

The project is currently developed with Go 1.26.5.

## Run locally

Start the ML service first:

```bash
cd services/ml-service
source .venv/Scripts/activate
python -m groundedsearch_ml
```

In a second terminal, start the Go service:

```bash
cd apps/search-api
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

Stop each service with `Ctrl+C`.

## Validate

```bash
gofmt -w .
go test ./...
go vet ./...
go build ./cmd/server
```

Validate the versioned contracts from the repository root with:

```bash
services/ml-service/.venv/Scripts/python.exe scripts/validate_contracts.py
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
| `ML_SERVICE_URL` | `http://localhost:8090` |
| `ML_SERVICE_TIMEOUT` | `2s` |
| `LOG_LEVEL` | `info` |

Duration values use Go duration syntax such as `500ms`, `5s`, or `1m`.
