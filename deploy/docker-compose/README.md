# Docker Compose Deployment

The local Compose environment runs:

- PostgreSQL 18.4 as the authoritative metadata store;
- the Go search API;
- the Python ML service.

PostgreSQL is available only on the internal Compose network. It is not
published to a host port by default.

## Validate migrations

From the repository root:

```bash
python scripts/validate_migrations.py
```

The validator starts a temporary PostgreSQL container, applies all upward
migrations, verifies the expected tables and constraints, applies all downward
migrations in reverse order, verifies cleanup, and removes the container.

## Build and verify the Phase 1 services

```bash
python scripts/verify_compose.py
```

## Start manually

```bash
docker compose -f deploy/docker-compose/compose.yaml up --build
```

Endpoints:

```text
Search API: http://localhost:8080
ML service: http://localhost:8090
```

Stop the environment:

```bash
docker compose   -f deploy/docker-compose/compose.yaml   down --volumes --remove-orphans
```

The `--volumes` option deletes local PostgreSQL data. Omit it when you want the
database contents to persist between runs.
