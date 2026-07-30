# ADR-002: Use Bleve, Qdrant, and PostgreSQL

- Status: Accepted
- Date: 2026-07-30
- Decision owners: GroundedSearch project
- Scope: Search indices and authoritative metadata storage

## Context

GroundedSearch requires:

- lexical retrieval for exact terms, identifiers, phrases, and rare tokens;
- dense retrieval for semantic similarity;
- document-level access-control decisions;
- stable document and chunk identifiers;
- connector checkpoints and lifecycle state;
- interaction-event storage;
- reproducible local operation without paid services.

A single search product could provide lexical, vector, and metadata storage, but
that would hide important orchestration and rank-fusion logic that the project is
intended to demonstrate.

## Decision

GroundedSearch will use:

- Bleve for lexical indexing and BM25-style retrieval;
- Qdrant for vector indexing and nearest-neighbour retrieval;
- PostgreSQL as the authoritative store for identities, groups, ACLs, document
  metadata, connector state, and interaction events.

The Go search API will coordinate lexical and dense retrieval and combine
candidate sets using Reciprocal Rank Fusion.

## Data ownership

### PostgreSQL owns

- users;
- groups;
- memberships;
- document lifecycle state;
- ACL rules;
- connector definitions;
- connector checkpoints;
- interaction events;
- model and index versions;
- audit metadata.

### Bleve owns derived data

- searchable chunk text;
- lexical fields;
- lexical scores;
- non-sensitive retrieval metadata.

### Qdrant owns derived data

- chunk embeddings;
- vector payload identifiers;
- collection version metadata;
- nearest-neighbour scores.

Bleve and Qdrant must be rebuildable from authoritative source and metadata.

## Security rule

Neither Bleve nor Qdrant is authoritative for final access decisions.

Candidate documents returned by either index must be checked against the
authoritative ACL state before:

- reranking;
- answer generation;
- citation construction;
- result counting;
- protected-response caching.

## Indexing rules

1. Every document has a stable `document_id`.
2. Every chunk has a stable `chunk_id`.
3. Embedding model versions are stored explicitly.
4. Index versions are stored explicitly.
5. Incompatible vector and model versions must not be mixed.
6. Connector checkpoints advance only after a consistent indexing batch.
7. Deletions and ACL changes must propagate to derived stores.
8. Drift-detection jobs must compare authoritative and derived state.

## Failure behaviour

- If Qdrant is unavailable, the system falls back to authorised lexical search.
- If the reranker is unavailable, the system returns authorised fused results.
- If Bleve is unavailable, dense-only retrieval is allowed only when final ACL
  enforcement remains authoritative.
- If PostgreSQL is unavailable, the system fails closed and reports not ready.
- Partial index updates do not advance the connector checkpoint.

## Consequences

### Positive

- The project demonstrates explicit hybrid-retrieval orchestration.
- Lexical and vector behaviour can be evaluated independently.
- ACL state remains centralised in a relational authoritative store.
- Each technology can run locally without paid services.
- Derived indices can be rebuilt and versioned.

### Negative

- Three storage systems increase integration complexity.
- Cross-store consistency must be designed and tested.
- Local resource use is higher than a single-engine design.
- Operational monitoring must cover multiple dependencies.

## Alternatives considered

### OpenSearch for lexical and vector search

Not selected for version 1.0 because it would reduce the amount of explicit
retrieval orchestration demonstrated and would require a heavier local runtime.

### PostgreSQL with pgvector for all storage

Not selected because it would simplify deployment but provide less direct
experience with a dedicated vector service and separate retrieval paths.

### Qdrant only

Rejected because vector search alone is weak for exact identifiers, codes, and
rare technical terms.

### Elasticsearch

Not selected because the project needs a lighter local-first configuration and
does not require the full operational complexity of Elasticsearch.

### Custom vector database

Rejected because building a vector database is outside the project scope and
would distract from search orchestration, security, and Kubernetes engineering.

## Validation

This decision is successful when:

- lexical, dense, and hybrid retrieval can be benchmarked separately;
- all derived data can be rebuilt;
- ACL changes take effect without relying on stale index payloads;
- Qdrant failure safely falls back to lexical search;
- local development remains workable on the target laptop;
- index and model versions appear in evaluation reports.

## Revisit triggers

Reconsider this decision if:

- local resource usage becomes unacceptable;
- measured cross-store consistency complexity outweighs portfolio value;
- one engine provides a clearly better evaluated design;
- document volume grows beyond the planned thousands;
- operational requirements justify consolidation.
