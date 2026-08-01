# Development Scripts

This directory contains cross-platform project automation that is not tied to a
single application service.

## Contract validation

```bash
python scripts/validate_contracts.py
```

## Docker Compose verification

```bash
python scripts/verify_compose.py
```

The Compose verifier builds both Phase 1 images, tests healthy communication,
tests degraded operation, and removes the temporary Compose environment.
