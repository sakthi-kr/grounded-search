# GroundedSearch Evaluation Plan

## 1. Purpose

This document defines how GroundedSearch will be evaluated before version 1.0.
No claim about search quality, ACL security, grounding, performance, reliability,
or Kubernetes behaviour is reportable unless it is produced by a reproducible,
version-controlled evaluation.

The evaluation covers:

- retrieval quality;
- access-control security;
- grounded-answer quality;
- connector correctness;
- interaction-aware ranking;
- API performance;
- reliability and degraded operation;
- Kubernetes deployment and operator behaviour;
- reproducibility.

## 2. Evaluation principles

### 2.1 Baselines before improvements

Every advanced component must be compared with a simpler baseline.

- Dense retrieval is compared with lexical retrieval.
- Hybrid retrieval is compared with lexical and dense retrieval.
- Cross-encoder reranking is compared with hybrid retrieval without reranking.
- Interaction-aware ranking is compared with the non-personalised ranking.
- Grounded answers are compared with evidence-only responses.

### 2.2 Quality and latency are separate

A method that improves relevance but adds excessive latency is not described as
unconditionally better. Reports must show both quality and resource cost.

### 2.3 Security is a hard constraint

Any configuration that returns unauthorised content fails, regardless of ranking
quality. Security failures are never averaged away.

### 2.4 Reproducible inputs

Every run records:

- dataset and version;
- random seed;
- document and query counts;
- relevance judgements;
- model names and versions;
- index version;
- Git commit;
- configuration checksum;
- hardware and software environment;
- execution timestamp.

### 2.5 Targets are not results

A target becomes reportable only after the corresponding evaluation command runs
successfully and the generated report is reviewed.

## 3. Evaluation environments

### 3.1 Unit-test environment

Used for deterministic ranking logic, ACL policy functions, citation validation,
request validation, connector transformations, fallbacks, and operator helpers.

### 3.2 Integration-test environment

Used for PostgreSQL, Bleve, Qdrant, Go-to-Python contracts, cache isolation,
connector checkpoints, and provider adapters.

### 3.3 Docker Compose environment

Used for complete local ingestion, search, grounded answers, observability, load
tests, and fault injection.

### 3.4 Kind environment

Used for Kubernetes manifests, health checks, rolling updates, operator
reconciliation, custom resources, and end-to-end ACL-protected search.

### 3.5 Optional GKE environment

A temporary GKE deployment may validate image publishing, Workload Identity,
networking, and cloud logging. It is not required for local version 1.0 acceptance.

## 4. Evaluation datasets

### 4.1 Synthetic enterprise corpus

The main corpus will contain approximately:

- 3,000 to 5,000 documents;
- 50 users;
- 10 to 20 groups;
- 8 to 12 document categories;
- public, internal, confidential, and restricted documents;
- 300 to 500 labelled queries;
- deterministic ACL assignments;
- deterministic interaction histories.

The corpus generator must accept an explicit random seed.

### 4.2 External retrieval datasets

Initial external datasets:

- SciFact;
- NFCorpus;
- one optional additional BEIR dataset.

External datasets assess retrieval behaviour. ACL evaluation remains project-specific.

### 4.3 Adversarial security corpus

The security corpus includes:

- inaccessible documents;
- conflicting allow and deny rules;
- prompt-injection instructions;
- fabricated citations;
- malformed and oversized inputs;
- path-traversal filenames;
- duplicate source identifiers;
- stale document versions.

### 4.4 Connector fixtures

Fixtures cover full sync, incremental updates, deletion, replay, interruption,
rate limits, malformed items, and checkpoint restoration.

## 5. Retrieval systems under comparison

### R1: Lexical baseline

- Bleve lexical index;
- BM25-style ranking;
- no dense retrieval;
- no reranker;
- ACL filtering enabled.

### R2: Dense baseline

- Qdrant vector retrieval;
- local query embeddings;
- no lexical retrieval;
- no reranker;
- ACL filtering enabled.

### R3: Hybrid Reciprocal Rank Fusion

- lexical candidates;
- dense candidates;
- deduplication;
- Reciprocal Rank Fusion;
- ACL filtering enabled;
- no cross-encoder reranker.

### R4: Weighted hybrid retrieval

- lexical and dense candidates;
- normalised scores;
- configurable weights;
- ACL filtering enabled.

This is evaluated only if score normalisation is defensible.

### R5: Hybrid plus cross-encoder reranking

- hybrid candidate generation;
- ACL filtering;
- reranking of authorised candidates only;
- final top-k results.

### R6: Interaction-aware reranking

- R5 base ranking;
- optional user, group, recency, and interaction features;
- non-personalised baseline retained.

## 6. Retrieval quality metrics

Required metrics:

- Precision@5;
- Precision@10;
- Recall@10;
- Recall@100;
- Mean Reciprocal Rank;
- nDCG@10;
- Success@5.

Results must also be broken down by query category:

- exact identifier;
- exact phrase;
- conceptual question;
- synonym or paraphrase;
- multi-document question;
- recent-document query;
- restricted-document query;
- no-answer query.

## 7. Retrieval experiment procedure

For each retrieval configuration:

1. Use a clean or explicitly versioned index.
2. Record corpus, query, model, and index versions.
3. Execute every query against the same relevance judgements.
4. Record ranked document identifiers and scores.
5. Apply authoritative ACL filtering.
6. Calculate quality metrics.
7. Record per-query latency.
8. Generate aggregate and category-level tables.
9. Save raw results and a human-readable report.

Required output structure:

```text
reports/retrieval/<run_id>/config.json
reports/retrieval/<run_id>/environment.json
reports/retrieval/<run_id>/metrics.json
reports/retrieval/<run_id>/per-query.jsonl
reports/retrieval/<run_id>/report.md
```

## 8. Retrieval acceptance criteria

Version 1.0 requires:

- evaluation of R1, R2, R3, and R5;
- the same labelled query set for all configurations;
- category-level results;
- latency measurements;
- zero unauthorised results;
- at least one documented category where hybrid retrieval recovers relevant
  content missed by the lexical baseline.

No minimum percentage improvement is fixed before baseline measurements.

## 9. ACL security evaluation

### 9.1 Generated property tests

Generate at least 10,000 deterministic combinations covering:

- public and private documents;
- allowed and denied users;
- allowed and denied groups;
- multiple group memberships;
- disabled users and documents;
- missing and malformed ACLs;
- changed memberships and permissions.

Required counters:

```text
authorised cases evaluated
unauthorised cases evaluated
unauthorised documents returned
unauthorised snippets returned
unauthorised citations returned
```

### 9.2 Pipeline-boundary tests

Recording mocks inspect data sent to:

- the cross-encoder reranker;
- the interaction-aware reranker;
- the answer provider;
- the citation builder;
- logs;
- traces.

A test fails if unauthorised content reaches any boundary.

### 9.3 Cache-isolation tests

Test identical queries across different users and groups, sequential and
concurrent requests, membership revocation, ACL updates, index-version changes,
and answer generation after cached search.

### 9.4 Direct-document tests

Test authorised, unauthorised, deleted, disabled, unknown, and malformed document
identifiers independently of search.

### 9.5 ACL acceptance criteria

```text
unauthorised documents returned: 0
unauthorised snippets returned: 0
unauthorised citations returned: 0
cross-user cache leaks: 0
unauthorised reranker inputs: 0
unauthorised answer-provider inputs: 0
```

Any non-zero value blocks release.

## 10. Grounded-answer evaluation

### 10.1 Answer statuses

Every request returns one of:

```text
grounded
insufficient_evidence
conflicting_evidence
generation_unavailable
```

### 10.2 Citation validity

A valid citation must:

- reference an authorised document;
- reference evidence included in the current request;
- match the current document version;
- contain a valid source location;
- support the associated statement.

### 10.3 Required grounding metrics

- citation validity rate;
- citation coverage rate;
- unauthorised citation rate;
- fabricated citation rate;
- citation offset accuracy;
- supported answer rate;
- correct abstention rate;
- false-answer rate on no-evidence queries.

### 10.4 Prompt-injection evaluation

Adversarial documents attempt to override system rules, disclose hidden content,
omit citations, fabricate citations, or treat document text as system instructions.

Measure:

- unauthorised disclosure count;
- invalid citation count;
- unsupported answer count;
- correct abstention count.

### 10.5 Grounding acceptance criteria

- every non-abstained factual answer has at least one valid citation;
- zero unauthorised citations;
- zero citations absent from current authorised evidence;
- explicit abstention when no authorised support exists;
- provider failure returns authorised evidence without generated text.

No claim of perfect factuality will be made.

## 11. Connector evaluation

Every connector must be tested for:

- initial full sync;
- unchanged repeated sync;
- content, metadata, and ACL updates;
- deletion;
- duplicate replay;
- interrupted batch;
- process restart;
- malformed item;
- rate-limit response;
- transient and permanent source errors.

Required counters:

- source items read;
- documents created, updated, unchanged, and deleted;
- duplicates created;
- failed items;
- retries;
- checkpoint advances.

Acceptance requires zero duplicates after repeated sync, correct update and
deletion behaviour, checkpoint-safe restart, and no credential leakage.

## 12. Interaction-aware ranking evaluation

Compare the interaction-aware ranker with the same retrieval configuration without
personalisation.

Synthetic histories include impressions, clicks, document opens, answer feedback,
user preferences, group preferences, and cold-start users.

Measure:

- nDCG@10 and MRR with and without interaction features;
- ranking changes by user and group;
- cold-start behaviour;
- maximum interaction-feature contribution.

The ranker must never reintroduce unauthorised candidates, override deny rules,
expose another user's history, or allow unbounded click influence.

## 13. API performance evaluation

### 13.1 Workloads

Measure lexical, dense, hybrid, hybrid-plus-reranker, grounded-answer, ingestion,
and mixed traffic separately.

### 13.2 Request rates

Initial local levels:

- 1 request per second;
- 5 requests per second;
- 10 requests per second;
- 20 requests per second;
- short burst traffic.

### 13.3 Metrics

- p50, p95, and p99 latency;
- throughput;
- error and timeout rates;
- CPU and memory use;
- indexing throughput;
- cold-start and model warm-up time.

Report total latency and component latency separately, including lexical, vector,
ACL, reranking, and generation stages.

### 13.4 Initial targets

For approximately 5,000 documents on the declared laptop:

- target p95 hybrid-search latency below 500 ms at 10 requests per second;
- target error rate below 1 percent during stable load;
- report answer-generation latency separately;
- report the measured maximum stable request rate.

These remain targets until measured.

## 14. Reliability and fault-injection evaluation

Test failures of the ML service, Qdrant, Bleve, answer provider, PostgreSQL,
connectors, indexing operations, model output, requests, and Kubernetes pods.

Expected behaviour:

| Failure | Expected result |
|---|---|
| Reranker unavailable | Return authorised hybrid results |
| Qdrant unavailable | Return authorised lexical results |
| Answer provider unavailable | Return authorised evidence only |
| Connector unavailable | Retain the last valid index |
| PostgreSQL unavailable | Fail closed and report not ready |
| Partial indexing failure | Preserve the previous checkpoint |
| Invalid model output | Reject output and abstain or return evidence |
| Pod termination | Kubernetes restarts the workload |

Measure recovery time, failed requests, fallback count, data loss, duplicate
processing, readiness transitions, and connector resume point.

## 15. Kubernetes deployment evaluation

Validate Deployments, Services, ConfigMaps, secret templates, resource limits,
health probes, storage, NetworkPolicies, autoscaling, service accounts, and RBAC.

The Kind end-to-end workflow must:

1. Create a cluster.
2. Build or load images.
3. Install dependencies.
4. Apply deployment resources.
5. Wait for readiness.
6. Seed deterministic data.
7. Execute an authorised search.
8. Execute an unauthorised search.
9. Verify the ACL difference.
10. Delete the cluster.

Acceptance requires a fresh successful deployment, working probes, resource
limits, no public exposure of internal services, an end-to-end ACL test, and
restart recovery.

## 16. Kubernetes operator evaluation

Initial custom resources:

- `SearchCluster`;
- `DataConnector`.

Test creation, repeated reconciliation, specification updates, drift correction,
child-resource failure, status conditions, finalizers, deletion, connector job
success, and connector job failure.

Required layers:

- unit tests;
- controller-runtime fake-client tests;
- envtest integration tests;
- Kind end-to-end tests.

Acceptance requires idempotent reconciliation, correct drift repair, Ready,
Progressing, and Degraded conditions, successful connector workload creation,
and deletion without orphaned managed resources.

## 17. Observability evaluation

Metrics must expose request count and latency, component latency, fallbacks, ACL
allow and deny counts, connector progress, indexed counts, and reconciliation
outcomes.

Metrics, logs, and traces must not include raw query text, document content,
prompts, credentials, or unbounded sensitive identifiers.

Log and trace tests capture output during successful, unauthorised, connector,
model, and database failure scenarios and scan for forbidden fixture strings.

## 18. Reproducibility requirements

Every report records:

- Git commit SHA and dirty-tree status;
- operating system, CPU, RAM, and GPU if used;
- Go, Python, Docker, and Kind versions;
- model and dataset versions;
- random seeds;
- configuration checksum;
- start and end times.

Evaluation commands must fail for missing inputs, invalid configurations,
incompatible model and index versions, or attempts to overwrite an immutable run.

## 19. Report structure

```text
reports/
|-- retrieval/
|-- security/
|-- grounding/
|-- connectors/
|-- personalisation/
|-- performance/
|-- reliability/
|-- kubernetes/
`-- operator/
```

Each run directory contains:

```text
config.json
environment.json
metrics.json
report.md
artifacts/
```

Large generated datasets are not committed unless explicitly justified.

## 20. Continuous integration evaluation

Pull-request CI should run:

- Go formatting and unit tests;
- Python formatting, linting, typing, and unit tests;
- OpenAPI validation;
- migration checks;
- deterministic ACL tests;
- lightweight integration tests;
- secret and dependency scanning;
- container build validation.

Long-running jobs may run nightly, manually, or before release:

- full 10,000-case ACL suite;
- complete retrieval benchmark;
- Kind end-to-end deployment;
- image scanning;
- load tests.

## 21. Version 1.0 evidence package

The release must include:

- retrieval comparison report;
- ACL security report;
- grounded-answer report;
- connector correctness report;
- performance report;
- reliability report;
- Kind end-to-end result;
- operator reconciliation result;
- exact reproduction commands;
- limitations and failed experiments.

## 22. Phase 0 evaluation-plan exit criteria

The evaluation plan is complete for Phase 0 when:

- baselines and advanced configurations are identified;
- quality, security, grounding, performance, and reliability metrics are defined;
- targets are distinguished from measured results;
- required datasets and fixtures are specified;
- report paths and reproducibility metadata are defined;
- Kubernetes and operator evaluation procedures are included.
