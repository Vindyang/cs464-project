# Omnishard Microservice Backend

This directory contains the split-service Omnishard backend. It is the service-oriented implementation used by the `full`, `backend`, `full-microservices`, and `single-image-microservices` deployment flavors.

The microservice version keeps strict ownership boundaries between workflow orchestration, provider I/O, metadata persistence, erasure coding, and the public gateway.

## Directory Layout

- `services/adapter/`
  - Provider connectivity, OAuth, credentials, settings, file metadata proxying, and shard I/O.
- `services/shardmap/`
  - File metadata, shard metadata, lifecycle history, and health state.
- `services/sharding/`
  - Reed-Solomon shard and reconstruct operations.
- `services/orchestrator/`
  - Upload, download, delete, and health-refresh workflow coordination.
- `services/gateway/`
  - NGINX public API boundary and OpenAPI docs surface.
- `services/shared/`
  - Cross-service DTOs, clients, transport helpers, persistence helpers, and provider abstractions.
  - This area must remain utility-only. Service-owned business logic belongs in the owning service's `internal/` tree.
- `tests/integration/`
  - Contract tests across service boundaries.
- `tests/e2e/`
  - End-to-end backend tests.
- `tests/run-tests.ps1`
  - Grouped PowerShell test runner for unit, integration, and e2e suites.

## Service Ownership

| Service | Port | Responsibility |
| --- | --- | --- |
| Adapter | `8080` | Provider connectivity, credentials, settings, shard I/O, metadata proxy endpoints |
| Shardmap | `8081` | File metadata, shard metadata, lifecycle logs, shard health |
| Orchestrator | `8082` | Workflow coordination for upload, download, delete, and health refresh |
| Sharding | `8083` | Reed-Solomon shard and reconstruct operations |
| Gateway | `8084` | Public API contract, route normalization, docs, and request logging |

## Public And Internal APIs

### Gateway public contract

Canonical public endpoints:

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

Compatibility redirects also exist for non-versioned paths such as `/upload`, `/download/{fileId}`, `/history`, `/providers`, and `/health`.

### Adapter service endpoints

Health and provider status:

- `GET /health`
- `GET /api/providers`

Google Drive OAuth:

- `GET /api/oauth/gdrive/authorize`
- `GET /api/oauth/gdrive/callback`
- `POST /api/oauth/gdrive/disconnect`

OneDrive OAuth:

- `GET /api/oauth/onedrive/authorize`
- `GET /api/oauth/onedrive/callback`
- `POST /api/oauth/onedrive/disconnect`

AWS S3 management:

- `POST /api/providers/awsS3/connect`
- `POST /api/providers/awsS3/disconnect`

Credentials and settings:

- `GET /api/credentials/status`
- `GET /api/credentials`
- `PUT /api/credentials/{provider}`
- `DELETE /api/credentials/{provider}`
- `GET /api/credentials/{provider}/secret`
- `GET /api/settings`
- `PUT /api/settings`
- `POST /api/settings/reset`

Shard I/O and metadata proxy routes:

- `POST /shards/upload`
- `GET /shards/{remoteId}?provider={providerId}`
- `DELETE /shards/{remoteId}?provider={providerId}`
- `GET /api/v1/files`
- `GET /api/v1/files/{fileId}`
- `DELETE /api/v1/files/{fileId}`
- `GET /api/v1/shards/file/{fileId}`

### Shardmap service endpoints

- `GET /health`
- `POST /api/v1/shards/register`
- `POST /api/v1/shards/record`
- `GET /api/v1/shards/file/{fileId}`
- `GET /api/v1/shards/{shardId}`
- `PUT /api/v1/shards/{shardId}/status`
- `GET /api/v1/files`
- `GET /api/v1/files/{fileId}`
- `DELETE /api/v1/files/{fileId}`
- `POST /api/v1/files/{fileId}/health-refresh`
- `POST /api/v1/lifecycle`
- `GET /api/v1/lifecycle`
- `DELETE /api/v1/lifecycle`
- `GET /api/v1/lifecycle/{fileId}`

### Orchestrator service endpoints

- `GET /health`
- `POST /api/orchestrator/upload`
- `GET /api/orchestrator/files/{fileId}/download`
- `GET /api/orchestrator/history`
- `GET /api/orchestrator/files/{fileId}/history`
- `POST /api/orchestrator/files/health/refresh`
- `POST /api/orchestrator/files/{fileId}/health/refresh`
- `DELETE /api/orchestrator/files/{fileId}`

### Sharding service endpoints

- `GET /api/sharding/health`
- `POST /api/sharding/shard`
- `POST /api/sharding/reconstruct`

## Run This Version

### Local process-by-process startup

From `backend/microservice/`:

```powershell
go run ./services/adapter/cmd/main.go
go run ./services/shardmap/cmd/main.go
go run ./services/sharding/cmd/main.go
go run ./services/orchestrator/cmd/main.go
```

Run the gateway from the repo root with Docker Compose:

```powershell
docker compose --profile backend up -d --build gateway
```

### Local Docker Compose from source

From the repository root:

```powershell
docker compose --profile full up -d --build
docker compose --profile backend up -d --build
```

The `full` profile starts the frontend and all backend services. The `backend` profile starts the backend services only.

### Official GitHub OSS release assets

Download one of the published release assets, save it locally as `docker-compose.yml`, and start it:

```powershell
Invoke-WebRequest https://github.com/Vindyang/cs464-project/releases/latest/download/docker-compose.full-microservices.yml -OutFile docker-compose.yml
docker compose up -d
```

```bash
curl -L -o docker-compose.yml https://github.com/Vindyang/cs464-project/releases/latest/download/docker-compose.full-microservices.yml
docker compose up -d
```

or

```powershell
Invoke-WebRequest https://github.com/Vindyang/cs464-project/releases/latest/download/docker-compose.single-image-microservices.yml -OutFile docker-compose.yml
docker compose up -d
```

```bash
curl -L -o docker-compose.yml https://github.com/Vindyang/cs464-project/releases/latest/download/docker-compose.single-image-microservices.yml
docker compose up -d
```

### Repo-local GHCR pull manifests

From the repository root:

```powershell
$env:IMAGE_NAMESPACE = "ghcr.io/vindyang"
$env:OMNISHARD_TAG = "<release-tag-or-commit-sha>"
docker compose -f deploy/compose/full-microservices.yml up -d
docker compose -f deploy/compose/single-image-microservices.yml up -d
```

## Environment Variables

Common service URLs:

- `ADAPTER_URL`, default `http://localhost:8080`
- `SHARDMAP_URL`, default `http://localhost:8081`
- `SHARDING_URL`, default `http://localhost:8083`
- `ORCHESTRATOR_URL`, default `http://localhost:8082`

Persistence:

- `Omnishard_DB_PATH` for adapter credentials, tokens, and settings
- `Omnishard_SHARDMAP_DB_PATH` for shardmap metadata and lifecycle history

Testing-only note:

- `backend/microservice/.env.example` exists for manual integration flows only.
- Backend services should not rely on `.env` for production credentials.

## Tests

Run all service package tests:

```powershell
go test ./services/...
```

Use the grouped test runner:

```powershell
.\tests\run-tests.ps1 -Type unit
.\tests\run-tests.ps1 -Type integration
.\tests\run-tests.ps1 -Type e2e
.\tests\run-tests.ps1 -Type all
```

Current test layout:

- Unit tests live under `services/*/tests/unit/`.
- Integration contract tests live under `tests/integration/`.
- E2E tests live under `tests/e2e/`.

Important contract checks preserved by the integration suite:

- Shardmap register requests must send positive `original_size`.
- Recorded shard `type` values must be uppercase `DATA` or `PARITY`.
- Sharding endpoints remain `/api/sharding/shard` and `/api/sharding/reconstruct`.

## References

- [../../docs/backend-microservice.md](../../docs/backend-microservice.md) for the deeper microservice architecture walkthrough.
- [../../docs/architecture.md](../../docs/architecture.md) for the cross-repo backend comparison.
- [../../docs/cicd.md](../../docs/cicd.md) for CI/CD and release details.

Published GHCR images for this backend flavor:

- `ghcr.io/vindyang/omnishard-adapter`
- `ghcr.io/vindyang/omnishard-shardmap`
- `ghcr.io/vindyang/omnishard-sharding`
- `ghcr.io/vindyang/omnishard-orchestrator`
- `ghcr.io/vindyang/omnishard-gateway`
- `ghcr.io/vindyang/omnishard-frontend`
- `ghcr.io/vindyang/omnishard-all-in-one`
