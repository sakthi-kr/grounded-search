# GroundedSearch Threat Model

## 1. Purpose

This document defines the main security risks, controls, and tests for
GroundedSearch.

The highest-priority security property is:

> A user must never receive, infer, cite, cache, log, trace, or send to a model
> any document content that the user is not authorised to access.

The threat model covers search, answer generation, ACL enforcement, connectors,
storage, observability, CI, Docker, and Kubernetes.

## 2. Scope

### 2.1 In scope

- Search and grounded-answer requests
- User and group resolution
- Document-level ACLs
- Lexical and vector indices
- Reranking and answer generation
- Local-file and GitHub connectors
- PostgreSQL, Bleve, Qdrant, Redis, and Ollama
- Docker Compose and Kubernetes deployments
- Logs, metrics, traces, CI, dependencies, and secrets

### 2.2 Out of scope for version 1.0

- Production-grade single sign-on
- Hardware-backed key storage
- Formal external penetration testing
- Formal compliance certification
- Multi-company tenant isolation
- Compromised host operating systems
- Physical side-channel attacks

## 3. Security objectives

### SO-001: Authorised retrieval only

Results, snippets, counts, citations, and answer evidence must contain only
documents authorised for the resolved user.

### SO-002: Fail-closed identity handling

Missing, unknown, disabled, or malformed identities must not receive protected
content.

### SO-003: Deny precedence

Explicit user or group deny rules override public, user-allow, and group-allow
rules.

### SO-004: Secure processing order

ACL filtering must occur before:

- cross-encoder reranking;
- interaction-aware reranking;
- answer prompt construction;
- citation construction;
- user-visible result counts;
- protected-response caching;
- detailed tracing.

### SO-005: No cross-user state leakage

Caches, temporary files, background jobs, and stored responses must not expose
one user's protected results to another user.

### SO-006: Grounded and authorised answers

Generated answers must use only authorised evidence selected for the current
request.

### SO-007: Safe observability

Logs, metrics, traces, and error messages must avoid raw protected content.

### SO-008: Reproducible security testing

Security claims must be backed by automated tests and committed reports.

## 4. Protected assets

| Asset | Security concern |
|---|---|
| Document content | May contain confidential technical information |
| Document metadata | Titles and existence can be sensitive |
| ACL rules | Determine who can access content |
| User identities | Used for authorisation decisions |
| Group memberships | Must be authoritative and current |
| Search queries | May reveal confidential user intent |
| Interaction history | May reveal interests or team activity |
| Citations | Can expose protected document identifiers |
| Connector credentials | May grant external-source access |
| Database credentials | Protect authoritative data |
| Model prompts | May contain retrieved evidence |
| Kubernetes secrets | Protect service credentials |
| Audit records | Must remain useful without leaking content |

## 5. Trust boundaries

### TB-001: Client to Go search API

The client is untrusted. Identity claims, query text, filters, pagination, and
event payloads must be validated.

### TB-002: Go search API to PostgreSQL

PostgreSQL is authoritative for users, groups, ACLs, document lifecycle,
connector state, and interaction events.

### TB-003: Go search API to Bleve

Bleve is a derived retrieval store. Its results are not automatically authorised.

### TB-004: Go search API to Qdrant

Qdrant is a derived retrieval store. Its payloads do not replace authoritative
ACL evaluation.

### TB-005: Go search API to Python ML service

The ML service must receive only the minimum authorised data required for its
operation.

### TB-006: Go search API to answer provider

The provider receives only authorised evidence. Retrieved text is untrusted data,
not instructions.

### TB-007: Connector to external source

GitHub and local files are untrusted inputs. Content, paths, metadata, and source
identifiers must be validated.

### TB-008: Application to observability systems

Logs, metrics, and traces have separate retention and access policies. Sensitive
payloads must be removed before export.

### TB-009: CI to dependencies and registries

Actions, packages, and images may be compromised and must be pinned, scanned, and
minimised.

### TB-010: Kubernetes control plane to workloads

RBAC, service accounts, Secrets, and NetworkPolicies define workload privileges
and isolation.

## 6. Assumptions

The version 1.0 design assumes:

- the host operating system is not already compromised;
- Kind is used for development rather than hostile multi-tenant workloads;
- local users are deterministic test identities;
- GitHub tokens are read-only and narrowly scoped;
- Ollama is local and not publicly exposed;
- security fixtures contain synthetic content only.

These assumptions must be reviewed before real organisational deployment.

## 7. Adversaries

### 7.1 Unauthorised external client

May try to retrieve protected documents, enumerate document existence, exhaust
resources, or exploit error messages.

### 7.2 Authenticated but underprivileged user

May manipulate filters, exploit stale group membership, poison interaction
history, or infer protected content through counts and timing.

### 7.3 Malicious document author

May inject instructions, manipulate ranking, crash parsers, or create oversized
documents.

### 7.4 Compromised connector source

May modify content or ACL metadata, replay stale data, or produce conflicting
identifiers.

### 7.5 Compromised dependency or CI action

May steal secrets, modify builds, or publish malicious images.

### 7.6 Misconfigured administrator

May accidentally grant broad permissions, expose services, or log raw content.

## 8. Authorisation model

The Go search API evaluates ACLs using authoritative PostgreSQL data.

Evaluation order:

```text
1. Deleted document -> deny.
2. Disabled document -> deny.
3. Disabled or unknown user -> deny.
4. Explicit user deny -> deny.
5. Matching group deny -> deny.
6. Public document -> allow.
7. Explicit user allow -> allow.
8. Matching group allow -> allow.
9. Otherwise -> deny.
```

Rules:

- deny overrides allow;
- missing or malformed ACL data fails closed;
- client-provided groups are ignored;
- groups are resolved from PostgreSQL;
- authorisation is evaluated for every request;
- protected cache keys include the security context;
- permission changes invalidate affected cached entries.

## 9. Threats and required controls

### TM-001: Forged group membership

**Risk:** A client claims membership in a privileged group.

**Controls:**

- ignore client-provided groups;
- resolve groups from PostgreSQL;
- reject unknown or disabled users;
- test forged-group requests.

### TM-002: Filtering after reranking

**Risk:** Unauthorised text reaches the reranker.

**Controls:**

- filter before reranking;
- expose an authorised-candidate type;
- test with a recording mock reranker.

### TM-003: Filtering after prompt construction

**Risk:** Unauthorised text enters the answer prompt.

**Controls:**

- build prompts only from authorised evidence;
- validate evidence IDs before provider invocation;
- test with a recording mock provider.

### TM-004: Result-count leakage

**Risk:** A response reveals the number of hidden matches.

**Controls:**

- calculate counts after ACL filtering;
- exclude raw candidate counts from public responses;
- compare responses across differently authorised users.

### TM-005: Snippet leakage

**Risk:** A hidden result is removed but its snippet remains.

**Controls:**

- build snippets after authorisation;
- avoid mutable candidate reuse;
- inspect entire serialised responses in tests.

### TM-006: Cross-user cache leakage

**Risk:** Results are cached by query alone.

**Controls:**

- include identity security context, membership version, filters, retrieval mode,
  and index version in protected cache keys;
- invalidate on ACL and membership changes;
- test sequential and concurrent users.

### TM-007: Stale group membership

**Risk:** Revoked access persists through a cache.

**Controls:**

- maintain membership versions;
- include versions in cache keys;
- invalidate affected entries;
- test immediate revocation.

### TM-008: Inconsistent ACL state across stores

**Risk:** PostgreSQL and indices disagree after updates.

**Controls:**

- treat PostgreSQL as authoritative;
- use versioned indexing operations;
- do not advance checkpoints after partial failure;
- run drift detection;
- test deletions and ACL changes.

### TM-009: Document-ID enumeration

**Risk:** A user probes known document identifiers.

**Controls:**

- authorise direct document endpoints;
- avoid distinguishable absent-versus-hidden responses where practical;
- rate-limit repeated probes.

### TM-010: Timing side channel

**Risk:** Latency reveals hidden document existence.

**Controls:**

- use a consistent secure retrieval path;
- avoid different hidden-match errors;
- measure timing distributions;
- document residual risk honestly.

### TM-011: Prompt injection

**Risk:** A document tells the model to ignore instructions or reveal data.

**Controls:**

- mark retrieved text as untrusted evidence;
- delimit evidence records;
- require citations to authorised evidence IDs;
- reject invalid citations;
- maintain an adversarial test set;
- abstain when grounding cannot be validated.

### TM-012: Fabricated citations

**Risk:** The model invents a document or chunk ID.

**Controls:**

- validate every citation against the authorised evidence set;
- reject or retry invalid output;
- abstain after repeated validation failure.

### TM-013: Corrupted citation offsets

**Risk:** Citations point to the wrong passage after updates.

**Controls:**

- use deterministic chunking;
- store source version and content hash;
- validate offsets;
- add round-trip citation tests.

### TM-014: Sensitive logs

**Risk:** Queries, snippets, prompts, or content are logged.

**Controls:**

- use allow-listed structured fields;
- log identifiers and durations instead of raw content;
- redact credentials;
- add log-capture tests.

### TM-015: Sensitive metrics labels

**Risk:** Query or document values appear as labels.

**Controls:**

- use bounded labels only;
- prohibit raw queries, users, and documents;
- inspect the metrics endpoint in tests.

### TM-016: Sensitive traces

**Risk:** Prompts or evidence are attached to spans.

**Controls:**

- trace stages, durations, counts, and safe identifiers only;
- disable body capture;
- inspect exported test spans.

### TM-017: File path traversal

**Risk:** A local connector reads outside its configured root.

**Controls:**

- normalise paths;
- require paths to remain inside the root;
- reject symbolic links by default;
- test traversal and symlink cases.

### TM-018: Oversized document

**Risk:** A document exhausts CPU, memory, or disk.

**Controls:**

- enforce size, text, and chunk limits;
- stream where possible;
- use per-document timeouts;
- isolate failures.

### TM-019: Malformed parser input

**Risk:** Invalid encodings or parser edge cases crash ingestion.

**Controls:**

- validate encodings;
- isolate document failures;
- fuzz critical normalisers;
- store safe structured errors.

### TM-020: GitHub token exposure

**Risk:** A token is committed, logged, or mounted too broadly.

**Controls:**

- use read-only least-privilege tokens;
- load from ignored environment files or Kubernetes Secrets;
- redact tokens;
- scan commits and CI output;
- rotate exposed tokens immediately.

### TM-021: Connector replay

**Risk:** A stale checkpoint creates duplicates or stale state.

**Controls:**

- use stable checkpoints and content hashes;
- make upserts idempotent;
- commit checkpoints only after consistent indexing;
- test restart and replay.

### TM-022: Interaction-event poisoning

**Risk:** Fake events manipulate personalisation.

**Controls:**

- validate ownership and result provenance;
- rate-limit events;
- bound feature influence;
- retain a non-personalised baseline.

### TM-023: Query denial of service

**Risk:** Expensive or concurrent queries exhaust resources.

**Controls:**

- limit body size, query length, top-k, and candidate depth;
- apply deadlines and cancellation;
- rate-limit endpoints;
- test burst traffic.

### TM-024: Model-service denial of service

**Risk:** Reranking or generation exhausts CPU or memory.

**Controls:**

- bound candidate counts, context, and output;
- set timeouts and concurrency limits;
- provide safe fallbacks;
- expose saturation metrics.

### TM-025: SQL injection

**Risk:** Untrusted fields are concatenated into SQL.

**Controls:**

- use parameterised queries;
- allow-list dynamic sort and filter fields;
- test injection payloads.

### TM-026: Internal service exposure

**Risk:** PostgreSQL, Qdrant, Redis, Ollama, or metrics are public.

**Controls:**

- bind local services to loopback where practical;
- use internal Kubernetes Services;
- apply NetworkPolicies;
- check deployments for unexpected exposure.

### TM-027: Excessive Kubernetes permissions

**Risk:** A compromised component has cluster-admin access.

**Controls:**

- define least-privilege RBAC;
- separate operator and application service accounts;
- scope permissions;
- test generated RBAC.

### TM-028: Kubernetes secret leakage

**Risk:** Secrets enter manifests, logs, or Git.

**Controls:**

- commit templates only;
- use runtime Secrets;
- avoid environment dumps;
- run secret scanning;
- document rotation.

### TM-029: Untrusted container image

**Risk:** Images contain malicious code or known vulnerabilities.

**Controls:**

- pin versions;
- prefer immutable digests for releases;
- scan images;
- minimise base images;
- run as non-root where supported.

### TM-030: Malicious dependency or CI action

**Risk:** A dependency steals secrets or modifies builds.

**Controls:**

- use lock files;
- pin release actions to immutable commits;
- restrict CI token permissions;
- scan dependencies;
- remove unnecessary third-party actions.

### TM-031: Unsafe error response

**Risk:** Stack traces, paths, or internal details are returned.

**Controls:**

- map failures to stable public error codes;
- keep diagnostics in safe logs;
- test malformed requests and dependency failures.

### TM-032: Insecure default configuration

**Risk:** Development defaults disable ACLs or expose services.

**Controls:**

- keep ACL enforcement enabled in every profile;
- require explicit debug options;
- fail startup when required security settings are missing;
- test default configuration.

### TM-033: Generated-data leakage

**Risk:** Dumps, prompts, reports, or local storage enter Git history.

**Controls:**

- use synthetic data;
- ignore local storage and dumps;
- inspect staged files;
- use explicit report directories.

### TM-034: Partial index update

**Risk:** One store updates while another fails.

**Controls:**

- use indexing job states;
- retain the last consistent checkpoint;
- retry idempotently;
- reconcile index versions;
- deny documents with inconsistent authoritative metadata.

## 10. Security controls by component

### 10.1 Go search API

- strict JSON decoding;
- request-size limits;
- parameterised database access;
- authoritative identity resolution;
- central ACL policy function;
- bounded top-k;
- deadlines and cancellation;
- safe errors;
- allow-listed logs;
- security-context-aware caching.

### 10.2 Python ML service

- private exposure only;
- bounded batches and input length;
- timeouts;
- no raw input logging;
- deterministic recording mocks for tests.

### 10.3 PostgreSQL

- least-privilege account;
- parameterised queries;
- migration checks;
- no public exposure;
- ACL integrity constraints.

### 10.4 Bleve and Qdrant

- internal access only;
- stable identifiers;
- model and index versions;
- deletion support;
- drift detection;
- no final authorisation based solely on index payloads.

### 10.5 Answer provider

- authorised evidence only;
- bounded context and output;
- prompt-injection delimiters;
- citation validation;
- no provider call without evidence.

### 10.6 Connectors

- least-privilege credentials;
- source allow-lists;
- input limits;
- path normalisation;
- idempotent writes;
- safe retries;
- checkpoint integrity.

### 10.7 Kubernetes operator

- least-privilege RBAC;
- validated custom resources;
- safe finalizers;
- no secrets in status;
- bounded reconciliation retries.

## 11. Security test plan

### 11.1 ACL property tests

Generate at least 10,000 combinations of:

- public and private documents;
- allowed and denied users;
- allowed and denied groups;
- disabled users and documents;
- missing and malformed ACLs;
- permission changes.

Required outcome:

```text
unauthorised documents returned: 0
unauthorised snippets returned: 0
unauthorised citations returned: 0
```

### 11.2 Cross-user isolation tests

Test:

- identical queries from different users;
- sequential and concurrent cache access;
- group revocation;
- ACL updates;
- index-version changes;
- cached search followed by answer generation.

### 11.3 Recording mock tests

Use recording mocks for:

- reranker requests;
- answer-provider prompts;
- logs;
- traces.

Tests fail if unauthorised content reaches any recording mock.

### 11.4 Prompt-injection tests

Include documents that attempt to:

- override system instructions;
- reveal hidden documents;
- fabricate citations;
- omit citations;
- encode malicious instructions indirectly.

Measure unauthorised disclosure, invalid citation, abstention, and supported
answer rates.

### 11.5 Connector security tests

Test:

- path traversal;
- symbolic links;
- oversized files;
- malformed Unicode;
- duplicate source IDs;
- interrupted sync;
- stale checkpoint replay;
- deletion;
- token redaction.

### 11.6 API abuse tests

Test:

- oversized bodies and queries;
- invalid JSON;
- unknown fields;
- extreme top-k;
- invalid filters;
- SQL injection strings;
- identifier probing;
- burst requests.

### 11.7 Kubernetes security tests

Check:

- non-root execution where practical;
- resource limits;
- separated service accounts;
- least-privilege RBAC;
- internal-only services;
- NetworkPolicies;
- no committed secrets.

### 11.8 Supply-chain tests

Run:

- dependency scanning;
- secret scanning;
- container scanning;
- licence reporting;
- software-bill-of-materials generation for releases.

## 12. Security acceptance criteria

Version 1.0 requires:

1. Zero unauthorised documents across at least 10,000 generated ACL tests.
2. Zero unauthorised snippets and citations in the same suite.
3. Zero protected-content records received by mock reranker and answer providers.
4. Zero cross-user cache leaks in sequential and concurrent tests.
5. Direct document access is independently authorised.
6. Logs, metrics, and traces pass forbidden-content checks.
7. Prompt-injection tests produce no unauthorised disclosure.
8. Connector traversal and token-redaction tests pass.
9. Default deployment exposes no internal database or model service publicly.
10. CI security scans pass or contain documented, reviewed exceptions.

These are targets, not current results.

## 13. Residual risks

Version 1.0 cannot completely eliminate:

- statistical timing signals;
- unsupported local-model output;
- ranking abuse not represented by synthetic interaction data;
- unknown malicious dependencies;
- differences between Kind and managed Kubernetes;
- host compromise;
- limitations of test identities compared with production SSO.

Residual risks must remain visible in the final documentation.

## 14. Secret-exposure response

If a credential is exposed:

1. Revoke or rotate it immediately.
2. Remove it from the working tree.
3. Check Git history, CI logs, images, and releases.
4. Invalidate affected artefacts.
5. Review access logs where available.
6. Add a preventive test or scan.
7. Document the incident without reproducing the secret.

Removing a secret from only the latest commit is insufficient if it remains in
repository history.

## 15. Security review checkpoints

Review security:

- after ACL implementation;
- after caching;
- before reranking;
- before answer generation;
- after each connector;
- before Kubernetes exposure;
- before the version 1.0 release.

A phase must not be marked complete while a known ACL leakage defect remains.

## 16. Phase 0 threat-model exit criteria

This threat model is complete for Phase 0 when:

- protected assets and trust boundaries are identified;
- ACL evaluation order is explicit;
- high-risk leakage paths are documented;
- controls map to planned components;
- security tests and acceptance targets are defined;
- residual risks and exclusions are stated honestly.
