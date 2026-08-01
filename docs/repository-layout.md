# Repository Layout

GroundedSearch uses a monorepo so application code, service contracts,
deployment files, evaluations, tests, and documentation can change atomically.

```text
grounded-search/
|-- apps/
|   `-- search-api/
|-- services/
|   `-- ml-service/
|-- api/
|   |-- openapi/
|   `-- internal/
|-- internal/
|-- connectors/
|-- operator/
|-- deploy/
|   `-- docker-compose/
|-- eval/
|-- infra/
|-- data/
|-- scripts/
|-- tests/
|   |-- integration/
|   `-- e2e/
|-- docs/
|-- .github/
|-- .editorconfig
|-- .env.example
|-- .gitattributes
|-- CONTRIBUTING.md
|-- LICENSE
|-- SECURITY.md
`-- README.md
```

## Directory responsibilities

- `apps/search-api/`: public Go API and search orchestration.
- `services/ml-service/`: Python embedding and reranking service.
- `api/openapi/`: public HTTP API contracts.
- `api/internal/`: versioned internal service contracts.
- `internal/`: shared Go packages that are not public modules.
- `connectors/`: source-specific ingestion implementations.
- `operator/`: Kubernetes custom resources and controllers.
- `deploy/`: Docker Compose and Kubernetes deployment definitions.
- `eval/`: retrieval, security, grounding, and performance evaluation code.
- `infra/`: optional cloud infrastructure-as-code.
- `data/`: documentation and small deterministic fixtures only.
- `scripts/`: cross-platform development and automation commands.
- `tests/integration/`: tests spanning multiple components.
- `tests/e2e/`: complete user-path and deployment tests.
- `docs/`: architecture, requirements, security, evaluation, and operations.

Generated data, model files, service volumes, and private credentials are not
stored in the repository.
