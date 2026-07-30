# ADR-001: Use a Monorepo

- Status: Accepted
- Date: 2026-07-30
- Decision owners: GroundedSearch project
- Scope: Repository organisation and cross-component development

## Context

GroundedSearch contains multiple closely related components:

- a Go search API;
- a Python embedding and reranking service;
- local-file and GitHub connectors;
- a Kubernetes operator written in Go;
- OpenAPI contracts and shared schemas;
- Docker Compose and Kubernetes deployment files;
- Terraform configuration;
- retrieval, security, performance, and reliability evaluations;
- documentation and release artefacts.

These components evolve together. A change to an API contract may require updates
to the Go client, Python server, integration tests, deployment files, and
documentation in the same pull request.

The project is maintained primarily by one developer and is intended to be easy
for reviewers to inspect and run.

## Decision

GroundedSearch will use one Git repository containing all application,
infrastructure, evaluation, test, and documentation code.

The planned top-level structure is:

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

Each component will maintain clear internal boundaries even though it shares the
same repository.

## Rules

1. Cross-component contracts must be version controlled under `api/`.
2. Shared Go packages must live under explicit Go module boundaries.
3. Python packages must not import code directly from unrelated services.
4. Generated artefacts must be reproducible and excluded from Git unless they are
   small, stable, and useful to reviewers.
5. A pull request may change multiple components when one logical change requires
   it.
6. CI must detect which checks are required, but correctness takes priority over
   minimising CI duration.
7. Secrets and local service data must never be committed.

## Consequences

### Positive

- One pull request can update an interface and all consumers atomically.
- Integration and end-to-end tests remain close to the code they validate.
- Repository setup is simpler for a single maintainer and external reviewer.
- Releases can reference one commit containing application, infrastructure, and
  evaluation state.
- Architecture documentation and implementation remain synchronised.

### Negative

- The repository will contain multiple languages and toolchains.
- CI configuration will be more complex.
- Unrelated components may appear in the same repository history.
- Care is required to avoid hidden coupling between directories.

## Alternatives considered

### Separate repository per service

Rejected for version 1.0 because it would add release coordination, duplicated
configuration, cross-repository pull requests, and more difficult reproduction.

### Separate repository for infrastructure

Rejected because the Kubernetes operator, manifests, and Terraform configuration
must remain aligned with application versions.

### Separate repository for evaluation

Rejected because evaluation code must use the exact schemas, fixtures, and
versions of the implementation under test.

## Validation

This decision is successful when:

- a new contributor can clone one repository and run documented setup commands;
- interface changes can be tested atomically;
- component boundaries remain clear;
- CI can run language-specific and integration checks;
- version 1.0 can be reproduced from one tagged commit.

## Revisit triggers

Reconsider this decision if:

- independent teams own different services;
- release cadences diverge substantially;
- repository size or CI duration becomes difficult to manage;
- access-control requirements demand separate repositories;
- a component becomes a reusable product with independent consumers.
