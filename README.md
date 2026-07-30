# GroundedSearch

GroundedSearch is a planned Kubernetes-native, access-control-aware enterprise
search and grounded-answering platform.

The project is being developed as a production-style software engineering and
machine learning portfolio project. It will combine lexical retrieval, semantic
retrieval, machine-learning reranking, document-level access control, grounded
answer generation, external data connectors, observability, and Kubernetes
operations.

## Planned capabilities

- Lexical search using a BM25-style index
- Dense semantic retrieval using document embeddings
- Hybrid retrieval using rank fusion
- Cross-encoder reranking
- User-level and group-level document access control
- Grounded answers with passage-level citations
- Abstention when evidence is insufficient
- Local-file and GitHub data connectors
- User-interaction event collection
- Interaction-aware reranking
- Docker Compose development environment
- Kubernetes deployment using Kind
- Custom Kubernetes operator written in Go
- Search-quality, security, and performance evaluation
- Optional Google Kubernetes Engine deployment

## Planned technology stack

- Go for the main search API and Kubernetes operator
- Python for embedding and reranking services
- Bleve for lexical search
- Qdrant for vector search
- PostgreSQL for metadata, identities, ACLs, and events
- Ollama with a small local model for answer generation
- Docker Compose for local development
- Kind and Kubebuilder for Kubernetes development
- Prometheus, Grafana, and OpenTelemetry for observability
- GitHub Actions for continuous integration
- Terraform for optional cloud infrastructure

## Current status

Phase 0: requirements, scope, architecture, threat model, and evaluation plan.

No application functionality has been implemented yet.

## Documentation

- `docs/requirements.md`
- `docs/architecture.md`
- `docs/threat-model.md`
- `docs/evaluation-plan.md`
- `docs/adr/`

## Project principles

- Security checks must happen before reranking and answer generation.
- Reported metrics must come from reproducible evaluations.
- The system must degrade safely when optional services are unavailable.
- Local development must remain possible without paid cloud services.
- Generated answers must be supported by authorised source passages.
