# ML Service

The ML service is the Python inference boundary for GroundedSearch.

## Phase 1 functionality

Implemented endpoints:

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/healthz` | Process health |
| `GET` | `/readyz` | Phase 1 readiness |
| `GET` | `/v1/model/info` | Deterministic service and model metadata |

No embedding or reranking model is loaded during Phase 1. Both model fields are
returned as `null`, and the mode is reported as `foundation`.

## Requirements

- Python 3.10 through Python 3.14

The project is currently developed locally with Python 3.10.5.

## Create the virtual environment

From this directory in Git Bash:

```bash
python -m venv .venv
source .venv/Scripts/activate
python -m pip install --upgrade pip
python -m pip install -e ".[dev]"
```

On Linux or WSL, activate with:

```bash
source .venv/bin/activate
```

## Run locally

With the virtual environment active:

```bash
python -m groundedsearch_ml
```

The default address is:

```text
http://localhost:8090
```

Test it from another terminal:

```bash
curl -i http://localhost:8090/healthz
curl -i http://localhost:8090/readyz
curl -i http://localhost:8090/v1/model/info
```

Stop the service with `Ctrl+C`.

## Validate

```bash
python -m ruff format --check .
python -m ruff check .
python -m mypy src
python -m pytest
```

The test configuration enforces a minimum total coverage of 90 percent.

## Configuration

| Variable | Default |
|---|---|
| `ML_SERVICE_HOST` | `0.0.0.0` |
| `ML_SERVICE_PORT` | `8090` |
| `LOG_LEVEL` | `INFO` |
| `GROUNDEDSEARCH_ENV` | `development` |
