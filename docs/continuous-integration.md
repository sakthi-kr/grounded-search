# Continuous Integration

GroundedSearch uses GitHub Actions to validate every pull request into `main`
and every push to `main` or `phase-1-foundation`.

## Workflow

The workflow is defined in:

```text
.github/workflows/ci.yml
```

It contains four jobs.

### Go validation

- checks `gofmt`;
- runs Go tests with the race detector and coverage;
- runs `go vet`;
- builds the search API.

### Project-tool validation

- tests `tools/projectctl`;
- builds the project command tool.

### Python validation

- installs the ML service with development dependencies;
- checks Ruff formatting;
- runs Ruff linting;
- runs strict mypy checks;
- runs pytest with the configured coverage threshold;
- validates the public and internal OpenAPI contracts.

### Docker Compose verification

This job runs only after the Go and Python jobs succeed. It:

- validates the Compose file;
- builds both service images;
- starts both containers;
- checks healthy Go-to-Python communication;
- stops the ML service;
- confirms the Go API remains available in degraded mode;
- removes the test environment.

## Local equivalents

Before pushing, run:

```bash
go run ./tools/projectctl verify-core
python scripts/verify_compose.py
```

The local commands and CI checks must validate the same behaviour.

## Pull-request requirement

A Phase 1 pull request should not be merged until all four CI jobs pass.
