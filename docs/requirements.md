# GroundedSearch Requirements

## 1. Purpose

GroundedSearch is a Kubernetes-native enterprise search and grounded-answering
platform designed to demonstrate production-oriented software engineering,
information retrieval, machine learning, access control, observability, and
cloud-native operations.

The system will ingest structured and unstructured documents, retrieve relevant
content using lexical and semantic search, enforce document-level permissions,
rerank authorised candidates, and generate answers supported by exact source
citations.

## 2. Primary users

### 2.1 Knowledge worker

A knowledge worker searches internal technical content and requests grounded
answers. The worker must only receive documents authorised for that user or for
one of that user's groups.

### 2.2 System administrator

A system administrator configures data sources, observes indexing status,
reviews health information, and manages Kubernetes resources.

### 2.3 Platform engineer

A platform engineer deploys, upgrades, scales, monitors, and troubleshoots the
system through Docker Compose, Kubernetes manifests, and a custom operator.

### 2.4 Evaluator or recruiter

An evaluator reviews the repository, runs the system locally, reproduces the
reported measurements, and inspects the architecture and test coverage.

## 3. Core use cases

The completed system must support the following workflows:

1. Ingest documents from a local-file connector.
2. Ingest issues, pull requests, comments, and Markdown content from GitHub.
3. Store document metadata, chunks, source information, and access rules.
4. Search using lexical retrieval.
5. Search using dense semantic retrieval.
6. Combine lexical and dense candidates using hybrid rank fusion.
7. Remove unauthorised candidates before reranking or answer generation.
8. Rerank authorised candidates using a cross-encoder model.
9. Generate an answer only from authorised retrieved evidence.
10. Return document-level and passage-level citations.
11. Abstain when the available evidence is insufficient or contradictory.
12. Record user interactions such as impressions, clicks, and accepted answers.
13. Use interaction history as an optional reranking signal.
14. Run locally with Docker Compose.
15. Run on a local Kubernetes cluster created with Kind.
16. Deploy and manage the platform through custom Kubernetes resources.
17. Expose metrics, logs, and traces for operational inspection.
18. Produce reproducible quality, security, and performance reports.

## 4. Functional requirements

### FR-001: Document ingestion

The system must ingest documents with stable identifiers, source metadata,
timestamps, content type, and access-control metadata.

### FR-002: Deterministic chunking

The system must divide documents into deterministic chunks while preserving the
information required to map citations back to exact source passages.

### FR-003: Incremental synchronisation

Connectors must support initial ingestion and incremental updates without
creating duplicate documents.

### FR-004: Deletion handling

When a source document is deleted or becomes unavailable, the corresponding
searchable content must be removed or marked as unavailable.

### FR-005: Lexical retrieval

The search API must support BM25-style lexical retrieval and metadata filters.

### FR-006: Dense retrieval

The search API must support semantic retrieval using locally generated document
and query embeddings.

### FR-007: Hybrid retrieval

The system must combine lexical and dense result sets using Reciprocal Rank
Fusion. Weighted fusion may be added as an evaluated alternative.

### FR-008: Access-control enforcement

The system must support public access, explicit user access, group-based access,
and explicit deny rules.

### FR-009: Secure processing order

Access-control filtering must occur before cross-encoder reranking, grounded
answer generation, citation construction, and user-visible result counting.

### FR-010: Cross-encoder reranking

The system must optionally rerank the authorised candidate set using a local
cross-encoder model.

### FR-011: Safe degradation

If dense retrieval or reranking is unavailable, the system must return a safe
fallback result where possible instead of exposing data or failing silently.

### FR-012: Grounded answers

The answer API must generate responses only from authorised retrieved passages.

### FR-013: Citations

Every grounded answer must include citations containing at least the document
identifier, chunk identifier, and source-passage location.

### FR-014: Abstention

The answer API must return an explicit insufficient-evidence or
conflicting-evidence status when a supported answer cannot be produced.

### FR-015: Interaction events

The system must record search submissions, result impressions, result clicks,
document openings, and answer feedback.

### FR-016: Interaction-aware ranking

The system must support an optional ranking stage that uses previous interaction
signals. The non-personalised baseline must remain available.

### FR-017: Search API

The main API must expose versioned endpoints for search, answers, documents,
interaction events, health, and readiness.

### FR-018: Connector status

The system must expose connector progress, last successful checkpoint, failures,
and retry status.

### FR-019: Kubernetes deployment

The project must include Kubernetes resources for application workloads,
configuration, persistence, networking, health checks, and autoscaling.

### FR-020: Kubernetes operator

A custom operator written in Go must reconcile at least a SearchCluster custom
resource and a DataConnector custom resource.

### FR-021: Observability

The platform must expose metrics and structured logs. Distributed tracing must
cover the main search and answer request paths.

### FR-022: Reproducible evaluation

Quality, security, and performance evaluations must be runnable through
version-controlled commands and configuration files.

## 5. Non-functional requirements

### NFR-001: Security

The system must fail closed when user identity or access-control data is missing,
invalid, or inconsistent.

### NFR-002: Privacy

Unauthorised document content must not appear in responses, snippets, citations,
caches, logs, metrics labels, traces, or model prompts.

### NFR-003: Reliability

The system must define and test degraded behaviour for unavailable optional
components.

### NFR-004: Testability

Core ranking, access-control, connector, citation, and reconciliation behaviour
must be covered by automated tests.

### NFR-005: Reproducibility

A new contributor must be able to build, test, seed, and evaluate the project
using documented commands.

### NFR-006: Local-first operation

The complete development and evaluation workflow must be possible without a
paid cloud service.

### NFR-007: Resource awareness

The default local configuration must be suitable for a Windows laptop with
32 GB RAM and CPU-based model inference.

### NFR-008: Maintainability

Go and Python code must use clear package boundaries, static checks, formatting,
type checking, and documented interfaces.

### NFR-009: Explainability

Search responses must expose enough metadata to identify the retrieval mode,
ranking stages, and source documents used.

### NFR-010: Versioning

APIs, schemas, model configurations, database migrations, and Kubernetes custom
resources must use explicit versioning.

## 6. Initial success criteria

The following are project targets, not claims. They become reportable results
only after the corresponding evaluation has been executed.

### Search quality

- Compare lexical, dense, hybrid, and hybrid-plus-reranker pipelines.
- Report nDCG@10, MRR, Precision@5, Recall@10, and Recall@100.
- Demonstrate at least one documented query category where hybrid retrieval
  recovers relevant content missed by the lexical baseline.

### Access-control security

- Return zero unauthorised documents across at least 10,000 generated access
  tests.
- Return zero unauthorised citations.
- Return zero cross-user cache leaks in the automated test suite.

### Grounding

- Every non-abstained factual answer must include at least one valid authorised
  citation.
- Unsupported questions must produce an explicit abstention status.

### Performance

For the initial local benchmark using approximately 5,000 documents:

- Measure p50, p95, and p99 latency.
- Measure throughput and error rate at several request rates.
- Target p95 hybrid-search latency below 500 ms at 10 requests per second.
- Measure answer-generation latency separately from retrieval latency.

### Reliability

- Reranker failure must fall back to authorised hybrid results.
- Vector-search failure must fall back to authorised lexical results.
- Answer-generation failure must return search evidence without a generated
  answer.
- Connector interruption must resume from a saved checkpoint.

### Kubernetes

- A Kind-based end-to-end test must deploy the platform and execute an
  ACL-protected search request.
- Creating a SearchCluster custom resource must reconcile the required
  workloads.
- Creating a DataConnector custom resource must start and report an ingestion
  operation.

## 7. Constraints

- The project must not require paid cloud infrastructure.
- Large-language-model fine-tuning is outside the project scope.
- The default implementation must not depend on a discrete GPU.
- The system will initially target thousands, not millions, of documents.
- The project will use one substantial external connector: GitHub.
- The web interface, if added, will remain minimal and secondary to backend
  engineering and evaluation.

## 8. Explicit non-goals

The following are deliberately excluded from version 1.0:

- Training a foundation model or embedding model from scratch
- Fine-tuning a large language model
- Multi-region or multi-cloud deployment
- Production-grade single sign-on integration
- More than two document connectors
- A custom vector database
- A complex frontend application
- Kubernetes service mesh adoption
- Production-scale multi-node database clustering
- Formal compliance certification

## 9. Planned version 1.0 deliverables

- Go search API
- Python embedding and reranking service
- PostgreSQL schema and migrations
- Bleve lexical index
- Qdrant vector index
- Hybrid retrieval and cross-encoder reranking
- User-level and group-level ACL enforcement
- Grounded-answer API with citations and abstention
- Local-file and GitHub connectors
- Interaction-event storage and optional interaction-aware reranking
- Docker Compose environment
- Kubernetes manifests
- Go-based Kubernetes operator and CRDs
- Metrics, logs, traces, and operational dashboards
- Automated unit, integration, security, and end-to-end tests
- Reproducible search-quality and performance reports
- Optional GKE-ready Terraform configuration
- Complete setup, architecture, security, and evaluation documentation

## 10. Phase 0 completion condition

Phase 0 is complete when the following documents have been reviewed and committed:

- `docs/requirements.md`
- `docs/architecture.md`
- `docs/threat-model.md`
- `docs/evaluation-plan.md`
- Architecture decision records under `docs/adr/`

No production functionality is required during Phase 0.
