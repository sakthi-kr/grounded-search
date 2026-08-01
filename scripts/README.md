# Development Commands

The primary cross-platform project command is implemented in Go:

```bash
go run ./tools/projectctl help
```

The most useful command is:

```bash
go run ./tools/projectctl verify-core
```

It validates contracts, runs Go and Python tests, builds the Go API, and executes
the real service integration suite.

`validate_contracts.py` remains the contract-specific validator and is invoked by
`projectctl`.
