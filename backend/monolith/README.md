# Omnishard Monolith Backend

This directory contains the standalone monolith backend implementation. It is intentionally separate from `backend/microservice`, owns its own shared packages, and has no compile-time dependency on the microservice module.

The monolith collapses adapter, shardmap, sharding, orchestrator, and docs serving into one Go process while preserving the frontend-facing routes on `http://localhost:8080`.

## Directory Layout

- `cmd/`
  - Monolith service entrypoint.
- `internal/app/`
  - HTTP routing, request handlers, provider restoration, docs rendering, and runtime wiring.
- `internal/orchestrator/`
  - In-process upload, download, delete, and health-refresh workflow logic.
- `internal/shardmap/`
  - SQLite-backed file, shard, and lifecycle persistence logic.
- `internal/sharding/`
  - Reed-Solomon shard and reconstruct operations.
- `shared/adapter/`
  - Provider registry, provider abstractions, and concrete Google Drive, OneDrive, and S3 adapters.
- `shared/api/`
  - Shared DTOs used by the monolith HTTP handlers and workflow logic.
- `shared/db/`, `shared/models/`, `shared/transport/`, `shared/types/`
  - Storage helpers, data models, HTTP helpers, and workflow payload types.
- `shared/oauthhandler/`, `shared/onedrivehandler/`, `shared/s3handler/`
  - Provider-specific connection flows.
- `tests/integration/`
  - Monolith integration tests.

## Exposed API Surface

The live endpoint index is available at `GET /api/v1/docs`, and the raw OpenAPI YAML is available at `GET /api/v1/docs/openapi.yml`.

### System and docs

- `GET /health`
- `GET /api/v1/health`
- `GET /api/v1/docs`
- `GET /api/v1/docs/openapi.yml`

### Provider management

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

### Credentials and settings

- `GET /api/credentials`
- `GET /api/credentials/status`
- `GET /api/credentials/{provider}`
- `PUT /api/credentials/{provider}`
- `DELETE /api/credentials/{provider}`
- `GET /api/credentials/{provider}/secret`
- `GET /api/settings`
- `PUT /api/settings`
- `POST /api/settings/reset`

### File metadata and health

- `GET /api/v1/files`
- `GET /api/v1/files/{fileId}`
- `DELETE /api/v1/files/{fileId}`
- `DELETE /api/orchestrator/files/{fileId}`
- `GET /api/v1/shards/file/{fileId}`
- `POST /api/v1/files/health/refresh`
- `POST /api/orchestrator/files/health/refresh`
- `POST /api/v1/files/{fileId}/health/refresh`
- `POST /api/orchestrator/files/{fileId}/health/refresh`

### Workflow routes

- `POST /api/v1/upload`
- `POST /api/orchestrator/upload`
- `GET /api/v1/download/{fileId}`
- `GET /api/orchestrator/files/{fileId}/download`
- `GET /api/v1/history`
- `GET /api/orchestrator/history`
- `GET /api/v1/history/{fileId}`
- `GET /api/orchestrator/files/{fileId}/history`

## How It Differs From The Microservice Backend

- `backend/microservice` uses separate adapter, shardmap, sharding, orchestrator, and gateway services.
- `backend/monolith` executes those responsibilities in-process inside one HTTP server.
- The monolith serves its own docs endpoint directly and does not require an NGINX gateway container.
- The frontend points both `API_INTERNAL_URL` and `GATEWAY_URL` at the monolith service in the monolith deployment profile.

## Run This Version

### Direct local process startup

From `backend/monolith/`:

```powershell
go run ./cmd/main.go
```

Default environment variables:

- `PORT`, default `8080`
- `Omnishard_DB_PATH`, default `Omnishard.db`
- `Omnishard_SHARDMAP_DB_PATH`, default `Omnishard-shardmap.db`

### Local Docker Compose from source

From the repository root:

```powershell
docker compose --profile monolith up -d --build monolith frontend-monolith
```

For backend-only monolith startup:

```powershell
docker compose --profile monolith up -d --build monolith
```

Default endpoints:

- Monolith API: `http://localhost:8080`
- Frontend when `frontend-monolith` is running: `http://localhost:3000`

Stop the monolith profile:

```powershell
docker compose --profile monolith down
docker compose --profile monolith down -v
```

### Official GitHub OSS release asset

Download the published monolith compose asset and run it:

```powershell
Invoke-WebRequest https://github.com/Vindyang/cs464-project/releases/latest/download/docker-compose.single-image-monolith.yml -OutFile docker-compose.yml
docker compose up -d
```

```bash
curl -L -o docker-compose.yml https://github.com/Vindyang/cs464-project/releases/latest/download/docker-compose.single-image-monolith.yml
docker compose up -d
```

### Repo-local GHCR pull manifest

From the repository root:

```powershell
$env:IMAGE_NAMESPACE = "ghcr.io/vindyang"
$env:OMNISHARD_TAG = "<release-tag-or-commit-sha>"
docker compose -f deploy/compose/single-image-monolith.yml up -d
```

## Tests

Run the monolith module test suite:

```powershell
go test ./...
```

Run the monolith integration tests explicitly:

```powershell
go test ./tests/integration/... -count=1
```

Current CI coverage for the monolith is a build gate on `go build ./...`. The integration tests exist and should still be run locally before merging changes that affect monolith behavior.

## References

- `../../docs/backend-monolith.md` for the deeper monolith architecture walkthrough.
- `../../docs/architecture.md` for the comparison between the monolith and microservice variants.
- `../../docs/cicd.md` for CI/CD and release workflow details.