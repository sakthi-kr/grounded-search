# Development Scripts

This directory contains cross-platform setup, validation, build, test, and
cleanup automation.

## Contract validation

The Phase 1 API contracts are checked with:

```bash
services/ml-service/.venv/Scripts/python.exe scripts/validate_contracts.py
```

On Linux or macOS, use the virtual environment's `bin/python` path instead.

The validator checks the expected paths, operation identifiers, component
schemas, required fields, local references, and closed response objects.
