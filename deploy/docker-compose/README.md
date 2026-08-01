# Docker Compose Deployment

The Phase 1 Compose environment runs the Go search API and Python ML service as
separate containers.

## Build and verify

From the repository root:

```bash
python scripts/verify_compose.py
```

The script:

1. validates the Compose file;
2. builds both images;
3. starts the services;
4. waits for both health checks;
5. confirms the Go API reports the ML service as healthy;
6. stops the ML container;
7. confirms the Go API remains available in degraded mode;
8. removes containers, networks, and temporary resources.

## Start manually

```bash
docker compose -f deploy/docker-compose/compose.yaml up --build
```

Endpoints:

```text
Search API: http://localhost:8080
ML service: http://localhost:8090
```

Stop the environment:

```bash
docker compose -f deploy/docker-compose/compose.yaml down --remove-orphans
```

## Port overrides

Use environment variables when the default host ports are occupied:

```bash
SEARCH_API_HOST_PORT=18080 ML_SERVICE_HOST_PORT=18090 docker compose -f deploy/docker-compose/compose.yaml up --build
```

Container-to-container communication continues to use ports `8080` and `8090`.
