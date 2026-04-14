# Monolith Backend

## Overview

The monolith backend lives under `backend/monolith` and runs the Omnishard backend as one Go process. It keeps the same provider, metadata, workflow, and docs features exposed by the frontend-facing product, but executes orchestration, sharding, and persistence in-process instead of through inter-service HTTP calls.

This version is used by:

- the source-build `monolith` Docker Compose profile
- the GHCR pull-only `deploy/compose/single-image-monolith.yml` manifest
- the official GitHub OSS release asset `docker-compose.single-image-monolith.yml`

## Internal Breakdown

```text
backend/monolith/
  cmd/
  internal/
    app/
    orchestrator/
    sharding/
    shardmap/
  shared/
    adapter/
    api/
    db/
    models/
    oauthhandler/
    onedrivehandler/
    s3handler/
    transport/
    types/
  tests/
    integration/
```

Responsibilities by area:

- `internal/app/`
  - bootstraps the store, shardmap database, provider registry, HTTP routes, docs page, and server wiring
- `internal/orchestrator/`
  - coordinates upload, download, delete, and health-refresh workflows
- `internal/sharding/`
  - owns Reed-Solomon shard and reconstruct logic
- `internal/shardmap/`
  - owns SQLite-backed file, shard, and lifecycle persistence
- `shared/`
  - contains monolith-owned copies of provider adapters, DTOs, HTTP helpers, persistence helpers, and workflow types

## HTTP Surface

The monolith exposes one HTTP server on port `8080`.

### Docs and health

- `GET /health`
- `GET /api/v1/health`
- `GET /api/v1/docs`
- `GET /api/v1/docs/openapi.yml`

### Providers, credentials, and settings

- `GET /api/providers`
- `GET /api/v1/providers`
- `GET /api/oauth/gdrive/authorize`
- `GET /api/oauth/gdrive/callback`
- `POST /api/oauth/gdrive/disconnect`
- `GET /api/oauth/onedrive/authorize`
- `GET /api/oauth/onedrive/callback`
- `POST /api/oauth/onedrive/disconnect`
- `POST /api/providers/awsS3/connect`
- `POST /api/providers/awsS3/disconnect`
- `GET /api/credentials`
- `GET /api/credentials/status`
- `GET /api/credentials/{provider}`
- `PUT /api/credentials/{provider}`
- `DELETE /api/credentials/{provider}`
- `GET /api/credentials/{provider}/secret`
- `GET /api/settings`
- `PUT /api/settings`
- `POST /api/settings/reset`

### Files, health, and workflows

- `GET /api/v1/files`
- `GET /api/v1/files/{fileId}`
- `DELETE /api/v1/files/{fileId}`
- `DELETE /api/orchestrator/files/{fileId}`
- `GET /api/v1/shards/file/{fileId}`
- `POST /api/v1/files/health/refresh`
- `POST /api/orchestrator/files/health/refresh`
- `POST /api/v1/files/{fileId}/health/refresh`
- `POST /api/orchestrator/files/{fileId}/health/refresh`
- `POST /api/v1/upload`
- `POST /api/orchestrator/upload`
- `GET /api/v1/download/{fileId}`
- `GET /api/orchestrator/files/{fileId}/download`
- `GET /api/v1/history`
- `GET /api/orchestrator/history`
- `GET /api/v1/history/{fileId}`
- `GET /api/orchestrator/files/{fileId}/history`

## Runtime Modes

### Direct local startup

From `backend/monolith/`:

```powershell
go run ./cmd/main.go
```

Environment variables:

- `PORT`, default `8080`
- `Omnishard_DB_PATH`, default `Omnishard.db`
- `Omnishard_SHARDMAP_DB_PATH`, default `Omnishard-shardmap.db`

### Local source-build compose

From the repo root:

```powershell
docker compose --profile monolith up -d --build monolith frontend-monolith
docker compose --profile monolith up -d --build monolith
```

### Pull-only GHCR manifest

```powershell
$env:IMAGE_NAMESPACE = "ghcr.io/vindyang"
$env:OMNISHARD_TAG = "<release-tag-or-commit-sha>"
docker compose -f deploy/compose/single-image-monolith.yml up -d
```

## Testing

From `backend/monolith/`:

```powershell
go test ./...
go test ./tests/integration/... -count=1
```

Current CI coverage for the monolith is a build gate in the reusable CI workflow. The integration tests exist in-repo and should be run locally before merging behavior changes.

## References

- [../backend/monolith/README.md](../backend/monolith/README.md)
- [architecture.md](architecture.md)
- [cicd.md](cicd.md)