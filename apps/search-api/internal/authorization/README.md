# Authorisation Policy

This package contains the authoritative in-memory ACL decision function.

The policy is deliberately independent of HTTP, PostgreSQL, Bleve, Qdrant,
reranking, and answer generation. Repository code will load the required
identity, membership, lifecycle, and rule data, then call `Evaluate`.

The returned reason makes every decision explainable and testable.

## Validate

```bash
go test ./internal/authorization -cover
go test ./internal/authorization -run=Fuzz -fuzz=FuzzDenyOverrides -fuzztime=5s
```
