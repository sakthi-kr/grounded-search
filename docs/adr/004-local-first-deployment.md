# ADR-004: Use a Local-First Deployment Model

- Status: Accepted
- Date: 2026-07-30
- Decision owners: GroundedSearch project
- Scope: Development, testing, Kubernetes, and optional cloud deployment

## Context

GroundedSearch must be developable on a Windows laptop with:

- 32 GB RAM;
- CPU-based inference;
- approximately 2 GB GPU memory;
- limited local storage;
- no mandatory paid cloud services.

The project must still demonstrate:

- containerised services;
- Kubernetes deployment;
- a custom Kubernetes operator;
- observability;
- failure recovery;
- infrastructure-as-code;
- an optional Google Cloud path.

A cloud-only architecture would increase cost, slow iteration, and make the
repository harder for reviewers to reproduce.

## Decision

GroundedSearch will use a local-first deployment model.

Required environments:

1. native development processes for fast unit-level work;
2. Docker Compose for integrated local development;
3. Kind for Kubernetes and operator testing;
4. optional temporary GKE deployment after local completion.

Cloud services are optional deployment targets, not required dependencies.

## Local component mapping

| Capability | Local implementation |
|---|---|
| Search API | Go process or container |
| ML service | Python process or container |
| Lexical index | Bleve |
| Vector index | Qdrant container |
| Metadata database | PostgreSQL container |
| Cache or coordination | Optional Redis container |
| Answer generation | Local Ollama or deterministic mock |
| Metrics | Prometheus |
| Dashboards | Grafana |
| Tracing | OpenTelemetry-compatible local collector |
| Kubernetes | Kind |
| Operator development | Kubebuilder and controller-runtime |
| Infrastructure validation | Terraform CLI |

## Development profiles

### Profile A: Core development

Run only:

- Go search API;
- Python ML service;
- PostgreSQL;
- Qdrant;
- Bleve within the Go service.

Optional components remain stopped.

### Profile B: Integrated Docker Compose

Run:

- search API;
- ML service;
- PostgreSQL;
- Qdrant;
- optional Redis;
- optional Ollama.

Observability services are started only when required.

### Profile C: Kubernetes validation

Run:

- one-node Kind cluster;
- operator;
- application workloads;
- minimal persistence;
- minimal observability.

### Profile D: Performance benchmark

Stop non-essential components and run only the services required by the selected
benchmark.

## Resource rules

1. All containers must define sensible resource expectations.
2. The local default must use CPU inference.
3. Document and candidate limits must prevent unbounded memory use.
4. Models must be small enough for local execution.
5. Observability services must be optional.
6. Generated data and local volumes must remain outside Git.
7. Cleanup commands must remove unused local resources safely.
8. Benchmark reports must record the actual hardware and active services.

## Kubernetes rules

The local Kubernetes environment will use Kind.

The project must support:

- Deployments;
- Services;
- ConfigMaps;
- Secret templates;
- PersistentVolumeClaims;
- liveness probes;
- readiness probes;
- resource requests and limits;
- service accounts;
- RBAC;
- NetworkPolicies;
- HorizontalPodAutoscaler configuration;
- custom resources managed by the operator.

A single-node local cluster is acceptable for functional and reconciliation
testing. It is not evidence of multi-node availability.

## Optional GKE deployment

A temporary GKE deployment may be added after local completion.

The cloud workflow should:

1. provision resources with Terraform;
2. build and publish images;
3. deploy the application;
4. run smoke tests;
5. collect evidence;
6. destroy all resources.

The repository must not claim GKE deployment unless it was actually performed.

## Secrets

Local secrets must be supplied through:

- ignored `.env` files;
- local secret files;
- Kubernetes Secrets created outside committed manifests.

Committed files may contain templates and documented variable names, but never
real credentials.

## Consequences

### Positive

- The project can be completed without paying for cloud services.
- Development feedback is fast.
- Reviewers can reproduce the system locally.
- Kubernetes and operator behaviour can be tested before cloud deployment.
- Cloud costs and accidental resource leakage are reduced.
- CPU-only operation matches the available laptop.

### Negative

- Kind does not reproduce every managed-cluster property.
- Local performance results are hardware-specific.
- Running many services may consume significant RAM and storage.
- Local networking and persistence differ from production GKE.
- A real cloud demonstration requires a separate optional step.

## Alternatives considered

### GKE-first development

Rejected because it introduces cost, slower iteration, and unnecessary
dependency on cloud access.

### Docker Compose only

Rejected because the project must demonstrate Kubernetes resources and a custom
operator.

### Minikube instead of Kind

Not selected for the initial plan because Kind integrates cleanly with
container-based CI and disposable test clusters.

### Large local LLM

Rejected because answer generation is not the central engineering challenge and
large models do not fit the available GPU memory.

### Paid hosted vector database

Rejected because Qdrant can run locally and the project must remain free to
develop.

## Validation

This decision is successful when:

- core development runs without cloud credentials;
- Docker Compose starts the integrated application;
- Kind deploys the platform and executes an ACL-protected search;
- operator tests run locally and in CI;
- optional services can be stopped without breaking secure search fallbacks;
- setup and cleanup commands are documented;
- no paid cloud service is required for version 1.0 acceptance.

## Revisit triggers

Reconsider this decision if:

- local resource usage prevents reliable development;
- cloud-specific behaviour becomes a core project requirement;
- CI cannot support required Kind tests;
- managed services provide a clearly justified capability;
- the project moves from portfolio demonstration to real organisational use.
