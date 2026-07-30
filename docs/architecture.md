# GroundedSearch Architecture

## 1. Architecture goals

GroundedSearch is a local-first, Kubernetes-native enterprise search and
grounded-answering platform with strict document-level access control and
reproducible evaluation.

The architecture is designed to:

- enforce authorisation before reranking and answer generation;
- keep lexical search available when optional ML components fail;
- support deterministic ingestion and reproducible benchmarking;
- run on a single developer laptop without paid cloud services;
- expose stable interfaces between Go, Python, storage, and Kubernetes;
- make failures, fallbacks, and reconciliation behaviour observable.

## 2. Architectural principles

### 2.1 Security before intelligence

Access-control filtering is part of the retrieval pipeline, not a presentation
layer. Unauthorised chunks must never be sent to the reranker, answer generator,
cache, citation builder, logs, metrics labels, or traces.

### 2.2 Local-first operation

Every required feature must run with Docker Compose and Kind. Cloud deployment
is optional and must not be required for development, testing, or evaluation.

### 2.3 Graceful degradation

The system must retain safe reduced functionality when optional components fail:

- vector search unavailable: use authorised lexical results;
- reranker unavailable: use authorised hybrid results;
- answer generator unavailable: return authorised search evidence only;
- connector unavailable: retain the last valid indexed state.

### 2.4 Explicit interfaces

Components communicate through versioned HTTP APIs, database schemas, event
schemas, and Kubernetes custom resources.

### 2.5 Measured claims

Quality, latency, throughput, security, and recovery claims must come from
version-controlled evaluation commands and committed reports.

## 3. System context

GroundedSearch serves four external actors:

1. A knowledge worker submits search and grounded-answer requests.
2. An administrator configures connectors and reviews indexing status.
3. A platform engineer deploys, upgrades, monitors, and troubleshoots the system.
4. An evaluator reproduces tests, benchmarks, and reported results.

Initial external systems are:

- local files;
- the GitHub API;
- a local Ollama server;
- an optional Google Cloud deployment target.

## 4. High-level component diagram

```text
                         +----------------------+
                         |  Client / Minimal UI |
                         +----------+-----------+
                                    |
                              REST / OpenAPI
                                    |
                         +----------v-----------+
                         |    Go Search API     |
                         |----------------------|
                         | Request validation   |
                         | Identity resolution  |
                         | Query orchestration  |
                         | ACL enforcement      |
                         | Rank fusion          |
                         | Events and auditing  |
                         +---+----------+-------+
                             |          |
                 +-----------+          +----------------+
                 |                                         |
        +--------v---------+                     +---------v--------+
        | Bleve lexical    |                     | Qdrant vector    |
        | index            |                     | index            |
        +------------------+                     +------------------+
                 |                                         |
                 +----------------+------------------------+
                                  |
                         authorised candidates
                                  |
                         +--------v---------+
                         | Python ML service|
                         |------------------|
                         | Embeddings       |
                         | Cross-encoder    |
                         | Reranking        |
                         +--------+---------+
                                  |
                         authorised top-k
                                  |
                         +--------v---------+
                         | Grounding adapter|
                         |------------------|
                         | Ollama provider  |
                         | Mock provider    |
                         | Citation checks  |
                         | Abstention       |
                         +------------------+
```

Supporting components:

```text
PostgreSQL       users, groups, ACLs, metadata, jobs, and interaction events
Connectors       local-file and GitHub ingestion
Redis            optional cache and background-work coordination
Prometheus       metrics collection
Grafana          dashboards
OpenTelemetry    distributed traces
Kind             local Kubernetes cluster
Operator         SearchCluster and DataConnector reconciliation
Terraform        optional GKE-ready infrastructure
```

## 5. Component responsibilities

### 5.1 Go search API

The Go service is the primary entry point and policy-enforcement boundary.

Responsibilities:

- validate requests;
- resolve users and group memberships;
- orchestrate lexical and vector retrieval;
- deduplicate candidates;
- apply Reciprocal Rank Fusion;
- enforce ACL rules;
- invoke the reranker and safe fallbacks;
- invoke the grounding adapter and safe fallbacks;
- construct citations;
- record interaction and audit events;
- expose health, readiness, metrics, and structured logs.

The service must not trust clients to provide authoritative group membership.
During local development, a request may contain a user identifier, but the server
must resolve that user's groups from PostgreSQL.

### 5.2 Python ML service

The Python service performs model inference only. It does not make access-control
decisions.

Responsibilities:

- generate document embeddings;
- generate query embeddings;
- rerank authorised candidate batches;
- expose model metadata;
- expose health and readiness endpoints;
- support deterministic test-mode responses.

Only authorised candidates may be sent to this service for reranking.

### 5.3 Bleve lexical index

Bleve stores searchable chunk text and non-sensitive retrieval metadata.

Responsibilities:

- BM25-style retrieval;
- exact-term and phrase matching;
- field filtering;
- incremental updates;
- deletion;
- index rebuilds.

Bleve is a derived store. It is not the authoritative source for identity,
document lifecycle, or final ACL decisions.

### 5.4 Qdrant vector index

Qdrant stores chunk embeddings and vector-search metadata.

Responsibilities:

- nearest-neighbour search;
- collection versioning;
- incremental upserts;
- deletion by stable chunk identifier;
- model-version metadata.

The platform must reject searches or indexing operations when vector data and
embedding-model versions are incompatible.

### 5.5 PostgreSQL

PostgreSQL is the authoritative store for:

- users;
- groups and memberships;
- document metadata;
- access-control rules;
- connector definitions;
- checkpoints and job status;
- interaction events;
- model and index version metadata;
- non-sensitive audit records.

Database migrations are version controlled. The application must not report
ready until required migrations are present.

### 5.6 Connector framework

Connectors transform external resources into the common document schema.

Every connector must support:

- full synchronisation;
- incremental synchronisation;
- stable source identifiers;
- checkpointing;
- retries with bounded backoff;
- idempotent upserts;
- deletion or tombstone handling;
- progress and error reporting.

Version 1.0 includes:

- a local-file connector;
- a GitHub connector.

### 5.7 Grounding adapter

The grounding adapter runs only after retrieval, ACL filtering, and optional
reranking.

Responsibilities:

- select authorised evidence within a context budget;
- treat retrieved text as untrusted data;
- invoke a configured answer provider;
- validate citation identifiers;
- return one explicit answer status.

Supported answer statuses:

```text
grounded
insufficient_evidence
conflicting_evidence
generation_unavailable
```

Initial providers:

- deterministic mock provider for tests;
- local Ollama provider.

A Gemini provider may be added later without changing the public API.

### 5.8 Kubernetes operator

The operator is written in Go using Kubebuilder and controller-runtime.

It reconciles:

- `SearchCluster` resources into Deployments, Services, configuration, and
  status conditions;
- `DataConnector` resources into Jobs or CronJobs and connector status.

The operator must implement:

- idempotent reconciliation;
- finalizers where required;
- status conditions;
- bounded retries;
- drift correction;
- safe deletion handling.

## 6. Core request flows

### 6.1 Search flow

```text
1. Client sends a query and user identifier.
2. Go API validates the request.
3. Go API resolves the authoritative user and group memberships.
4. Lexical and dense retrieval run in parallel when available.
5. Candidates are deduplicated and fused.
6. ACL rules are evaluated for every candidate.
7. Only authorised candidates are sent to the reranker.
8. Reranked or fallback results are returned.
9. A non-sensitive interaction event is recorded.
```

### 6.2 Grounded-answer flow

```text
1. Execute the secure search flow.
2. Select authorised evidence chunks within the context budget.
3. Escape or reject document-provided instructions.
4. Invoke the configured answer provider.
5. Validate that every cited chunk belongs to the authorised evidence set.
6. Return a grounded answer or an explicit abstention status.
```

### 6.3 Ingestion flow

```text
1. Connector reads resources after the last committed checkpoint.
2. Resources are normalised into the common document schema.
3. Documents are validated and chunked deterministically.
4. Metadata and ACL rules are persisted.
5. Lexical entries and vectors are upserted with stable identifiers.
6. Deleted resources are removed or tombstoned.
7. The checkpoint advances only after the batch reaches a consistent state.
```

### 6.4 Interaction-aware reranking flow

```text
1. Secure retrieval creates the authorised candidate set.
2. Base relevance scores are calculated.
3. Optional user, group, recency, and global interaction features are added.
4. A transparent weighted ranker is applied initially.
5. The unpersonalised baseline remains selectable.
```

## 7. Data ownership and identifiers

Stable identifiers are required for reproducibility and idempotency:

- `document_id`: connector identifier plus source identifier;
- `chunk_id`: document identifier plus deterministic chunk identifier;
- `user_id`: internal user identifier;
- `group_id`: internal group identifier;
- `connector_id`: configured connector instance;
- `index_version`: immutable logical index version;
- `model_version`: explicit embedding or reranker version;
- `trace_id`: request-correlation identifier.

PostgreSQL owns identity, ACL, and lifecycle metadata. Search indices are derived
stores and must be rebuildable.

## 8. Access-control model

Version 1.0 supports:

- public documents;
- explicitly allowed users;
- explicitly allowed groups;
- explicitly denied users;
- explicitly denied groups;
- enabled and disabled documents.

Evaluation order:

```text
1. Deleted or disabled document -> deny.
2. Explicit user deny -> deny.
3. Matching group deny -> deny.
4. Public document -> allow.
5. Explicit user allow -> allow.
6. Matching group allow -> allow.
7. Otherwise -> deny.
```

Deny rules override allow rules. Missing or malformed ACL data fails closed.

## 9. Public API boundaries

Planned versioned endpoints:

```text
POST /v1/search
POST /v1/answer
POST /v1/events
POST /v1/documents
GET  /v1/documents/{document_id}
POST /v1/connectors/{connector_id}/sync
GET  /v1/connectors/{connector_id}/status
GET  /healthz
GET  /readyz
GET  /metrics
```

The OpenAPI specification is the source of truth for external request and
response schemas. Internal service calls use separate versioned contracts.

## 10. Deployment architecture

### 10.1 Docker Compose

The default development environment runs:

- search API;
- ML service;
- PostgreSQL;
- Qdrant;
- optional Redis;
- optional Ollama;
- optional Prometheus and Grafana.

### 10.2 Kind

The local Kubernetes environment uses a single-node Kind cluster. Workloads must
use resource requests and limits suitable for a 32 GB development laptop.

### 10.3 Optional GKE

Terraform may provision a temporary GKE environment, Artifact Registry, service
accounts, networking, and supporting resources.

Cloud deployment must not introduce a required proprietary dependency into the
local architecture.

## 11. Observability architecture

All services use structured logs and correlation identifiers.

Raw document content, raw prompts, and unauthorised identifiers must not be used
as metrics labels or trace attributes.

Primary signals:

- request count;
- p50, p95, and p99 request latency;
- lexical retrieval latency;
- vector retrieval latency;
- reranking latency;
- answer-generation latency;
- fallback activation;
- ACL allow and deny counts;
- connector progress and failures;
- indexed document and chunk counts;
- model and dependency health;
- operator reconciliation outcomes.

OpenTelemetry traces cover the main search and answer stages using identifiers
and timing information rather than raw content.

## 12. Failure handling

| Failure | Required behaviour |
|---|---|
| Dense retrieval unavailable | Continue with authorised lexical retrieval |
| Lexical retrieval unavailable | Use dense retrieval only when authoritative ACL enforcement remains available |
| Reranker unavailable | Return authorised fused ranking |
| Answer provider unavailable | Return authorised search evidence without generated text |
| PostgreSQL unavailable | Fail closed and report not ready |
| Connector interrupted | Resume from the last committed checkpoint |
| Partial index update | Do not advance the connector checkpoint |
| Invalid ACL data | Deny access and emit a safe audit event |
| Model version mismatch | Reject the operation until the index is rebuilt or migrated |

## 13. Testing architecture

Testing is divided into:

- Go unit tests for validation, fusion, ACLs, orchestration, and fallbacks;
- Python unit tests for embedding and reranking interfaces;
- integration tests with PostgreSQL, Bleve, and Qdrant;
- OpenAPI and internal-contract tests;
- security tests for leakage and cache isolation;
- retrieval evaluation against labelled queries;
- Kind-based end-to-end tests;
- operator envtest and reconciliation tests;
- fault-injection tests;
- load tests.

All fixtures use deterministic seeds and must not require private credentials.

## 14. Repository architecture

GroundedSearch uses a monorepo:

```text
groundedsearch/
|-- apps/
|   `-- search-api/
|-- services/
|   `-- ml-service/
|-- connectors/
|   |-- local-files/
|   `-- github/
|-- operator/
|-- api/
|-- internal/
|-- eval/
|-- deploy/
|-- infra/
|-- data/
|-- tests/
|-- docs/
|-- scripts/
|-- Makefile
`-- README.md
```

Shared contracts, tests, deployment files, and evaluation code remain versioned
with the services they validate.

## 15. Deferred architectural decisions

The following decisions remain deferred until measurements justify them:

- whether Redis is required;
- whether weighted fusion improves on Reciprocal Rank Fusion;
- whether ONNX inference is needed for reranking;
- whether ingestion needs an asynchronous queue;
- whether a dedicated frontend is worth maintaining;
- whether learning-to-rank improves on a transparent weighted ranker.

Deferred components must not be added solely because they are common in other
architectures.

## 16. Phase 0 architecture exit criteria

The architecture is ready for implementation when:

- service boundaries are consistent with the requirements;
- the ACL evaluation order is explicit;
- every required local component has a free implementation path;
- optional-component failure behaviour is defined;
- authoritative and derived stores are identified;
- deferred decisions are documented rather than silently assumed.
