# Microservice Backend

## Overview

The microservice backend is the service-oriented Omnishard implementation under `backend/microservice`. It splits provider management, metadata persistence, sharding, workflow orchestration, and the public API gateway into separate deployable units.

This version is used by:

- the source-build `full` and `backend` Docker Compose profiles
- the GHCR pull-only `deploy/compose/microservices.yml` manifest
- the GHCR pull-only `deploy/compose/all-in-one-microservices.yml` manifest
- the official GitHub OSS release assets `docker-compose.microservices.yml` and `docker-compose.all-in-one-microservices.yml`

## Service Breakdown

### Adapter

- Owns provider connectivity, OAuth, credentials, settings, shard I/O, and metadata proxy routes.
- Persists credentials, provider tokens, and general settings in `Omnishard.db`.
- Exposes the frontend-facing provider, credentials, and settings routes on port `8080`.

### Shardmap

- Owns file metadata, shard placement, lifecycle history, and shard health state.
- Persists its state in `Omnishard-shardmap.db`.
- Exposes metadata APIs on port `8081`.

### Sharding

- Owns Reed-Solomon shard generation and file reconstruction.
- Exposes `/api/sharding/shard` and `/api/sharding/reconstruct` on port `8083`.

### Orchestrator

- Owns upload, download, delete, and health-refresh workflows.
- Calls adapter, shardmap, and sharding over HTTP.
- Exposes workflow routes on port `8082`.

### Gateway

- Owns the public versioned API contract and docs surface.
- Runs as NGINX and forwards requests to the owning backend services.
- Exposes the public API on port `8084`.

## Directory Structure

```text
backend/microservice/
  services/
    adapter/
    gateway/
    orchestrator/
    sharding/
    shardmap/
    shared/
  tests/
    integration/
    e2e/
    run-tests.ps1
```

Important structure notes:

- Each service owns its business logic under its own `internal/` tree.
- `services/shared/` is reserved for DTOs, HTTP helpers, reusable clients, provider abstractions, and persistence helpers that are truly cross-service.
- `tests/integration/` verifies service contracts, especially orchestrator interactions with its dependencies.

## API Shape

### Public gateway routes

- `GET /api/v1/docs`
- `GET /api/v1/openapi.yml`
- `GET /api/v1/health`
- `GET /api/v1/providers`
- `POST /api/v1/upload`
- `GET /api/v1/download/{fileId}`
- `GET /api/v1/history`
- `GET /api/v1/history/{fileId}`
- `POST /api/v1/files/health/refresh`
- `POST /api/v1/files/{fileId}/health/refresh`
- `DELETE /api/v1/files/{fileId}`

### Direct adapter routes used by the frontend

- `GET /api/providers`
- `GET /api/credentials`
- `GET /api/credentials/status`
- `PUT /api/credentials/{provider}`
- `DELETE /api/credentials/{provider}`
- `GET /api/credentials/{provider}/secret`
- `GET /api/settings`
- `PUT /api/settings`
- `POST /api/settings/reset`
- `GET /api/v1/files`
- `GET /api/v1/files/{fileId}`
- `GET /api/v1/shards/file/{fileId}`

## Runtime Modes

### Local source-build compose

From the repo root:

```powershell
docker compose --profile full up -d --build
docker compose --profile backend up -d --build
```

### Service-by-service local startup

From `backend/microservice/`:

```powershell
go run ./services/adapter/cmd/main.go
go run ./services/shardmap/cmd/main.go
go run ./services/sharding/cmd/main.go
go run ./services/orchestrator/cmd/main.go
```

Gateway from the repo root:

```powershell
docker compose --profile backend up -d --build gateway
```

### Pull-only GHCR manifests

```powershell
$env:IMAGE_NAMESPACE = "ghcr.io/vindyang"
$env:OMNISHARD_TAG = "<release-tag-or-commit-sha>"
docker compose -f deploy/compose/microservices.yml up -d
docker compose -f deploy/compose/all-in-one-microservices.yml up -d
```

## Testing

Common commands from `backend/microservice/`:

```powershell
go test ./services/...
.\tests\run-tests.ps1 -Type unit
.\tests\run-tests.ps1 -Type integration
.\tests\run-tests.ps1 -Type e2e
.\tests\run-tests.ps1 -Type all
```

Important contract coverage includes:

- positive `original_size` on shardmap register requests
- uppercase `DATA` and `PARITY` shard types on shardmap record requests
- stable orchestrator error envelopes when dependency calls fail

## References

- [../backend/microservice/README.md](../backend/microservice/README.md)
- [architecture.md](architecture.md)
- [cicd.md](cicd.md)