# Development Workflow

GroundedSearch provides a small Go command-line tool for repeatable local
validation from Windows, Linux, macOS, Git Bash, and CI.

## Primary command

Run this from the repository root:

```bash
go run ./tools/projectctl verify-core
```

It runs, in order:

1. versioned API contract validation;
2. Go unit tests with coverage;
3. Python unit tests with coverage;
4. a clean Go search-API build;
5. the real Go-to-Python integration test suite.

The command stops immediately when one stage fails.

## Individual commands

```bash
go run ./tools/projectctl help
go run ./tools/projectctl contracts
go run ./tools/projectctl test-go
go run ./tools/projectctl test-python
go run ./tools/projectctl build-go
go run ./tools/projectctl integration
```

GNU Make is optional. On systems where it is available, the equivalent commands
are:

```bash
make help
make verify-core
make integration
```

## Python interpreter selection

The tool prefers the project virtual environment:

```text
services/ml-service/.venv/Scripts/python.exe   Windows
services/ml-service/.venv/bin/python           Linux and macOS
```

If neither path exists, it falls back to `python` and then `python3` on `PATH`.
The integration suite receives the selected interpreter through the
`GROUNDEDSEARCH_PYTHON` environment variable.

## Integration test behaviour

The integration suite:

- selects unused local ports;
- starts the Python ML service;
- waits for its health endpoint;
- builds and starts the Go search API;
- verifies a healthy combined status response;
- verifies request-ID propagation;
- terminates the ML service;
- verifies that the Go API remains available in degraded mode;
- cleans up every process and temporary binary.

The test uses loopback networking only and requires no cloud account or external
API.

## Scope of `verify-core`

`verify-core` validates executable behaviour. Ruff formatting, Ruff linting, and
mypy remain separate quality gates until the full root verification command and
CI workflows are added later in Phase 1.
