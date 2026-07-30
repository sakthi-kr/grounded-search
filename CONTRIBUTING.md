# Contributing to GroundedSearch

## Development workflow

1. Create a focused branch from the latest `main`.
2. Keep each commit limited to one logical change.
3. Add or update tests with behaviour changes.
4. Update documentation when commands, contracts, or architecture change.
5. Run the relevant local checks before pushing.
6. Open a pull request into `main`.
7. Merge only after required checks pass.

## Branch naming

```text
phase-1-foundation
feature/search-api-health
feature/ml-service-contract
fix/request-timeout
docs/development-setup
```

## Commit messages

```text
feat: add Go health endpoint
fix: handle ML service timeout
test: cover degraded system status
docs: document local service startup
build: add container build workflow
```

## Code-quality expectations

- Go code must be formatted and pass configured static checks.
- Python code must be formatted, linted, typed, and tested.
- External API changes require contract updates.
- Generated artefacts must be reproducible.
- Secrets, local databases, models, and private datasets must not be committed.
- Metrics in documentation must come from committed evaluation outputs.

## Pull requests

A pull request should explain what changed, why it changed, how it was tested,
security or compatibility implications, and known limitations.
