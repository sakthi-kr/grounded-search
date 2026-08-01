# Integration Tests

The integration suite starts the real Python ML service and a compiled Go search
API on dynamically selected loopback ports.

It verifies:

- healthy Go-to-Python communication;
- the versioned ML service metadata response;
- request-ID propagation;
- safe degraded operation after the ML service stops;
- cleanup of child processes and temporary binaries.

Run it from the repository root:

```bash
go run ./tools/projectctl integration
```

Or directly:

```bash
cd tests/integration
go test ./... -count=1 -v
```
