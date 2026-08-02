# Development Scripts

This directory contains cross-platform project automation that is not tied to a
single application service.

## Contract validation

```bash
python scripts/validate_contracts.py
```

## Core service verification

```bash
go run ./tools/projectctl verify-core
```

## Docker Compose verification

```bash
python scripts/verify_compose.py
```

## PostgreSQL migration validation

```bash
python scripts/validate_migrations.py
```

The migration validator uses a temporary PostgreSQL 18.4 container, applies all
upward migrations, verifies the schema and critical constraints, rolls all
migrations back, verifies cleanup, and removes the container.
