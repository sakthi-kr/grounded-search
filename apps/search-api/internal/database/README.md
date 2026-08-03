# Database Package

The `database` package owns the PostgreSQL connection pool used by the search
API.

It provides:

- validated `pgxpool` configuration;
- bounded minimum and maximum connection counts;
- startup connectivity verification;
- timeout-bounded health checks;
- explicit shutdown;
- a small stable statistics view.

Repository packages may use `Pool.Raw()` for database operations. HTTP handlers
should depend on repositories rather than using the raw pool directly.

## Unit tests

```bash
go test ./internal/database -cover
```

## PostgreSQL integration test

Start PostgreSQL:

```bash
docker compose   -f ../../deploy/docker-compose/compose.yaml   up -d postgres
```

Run the tagged test from `apps/search-api`:

```bash
DATABASE_URL='postgres://groundedsearch:groundedsearch_dev@localhost:5432/groundedsearch?sslmode=disable' go test -tags=integration ./internal/database -run TestOpenAgainstPostgreSQL -v
```

The Compose PostgreSQL service is not published to a host port by default.
Temporarily publish it for this native integration test as described in the
project instructions.
