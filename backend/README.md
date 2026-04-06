# Omnishard Backend

Omnishard is organized as independently deployable Go services with strict ownership boundaries.
The orchestrator is the workflow owner and coordinates sharding, storage, and shard metadata.
The gateway is the public API entrypoint and enforces endpoint/versioning standards.

## Current Architecture

- `services/orchestrator`
  - Public workflow entry/exit for upload and retrieval.
  - Calls `sharding`, `adapter`, and `shardmap` through HTTP clients.

- `services/gateway`
  - Public API boundary for external clients.
  - Implemented with NGINX reverse proxy configuration.
  - Exposes the standardized v1 contract and forwards to owner services.
  - Adds standardized access logging and request-id propagation.

- `services/sharding`
  - Owner of erasure coding operations.
  - Provides `POST /api/sharding/shard` and `POST /api/sharding/reconstruct`.

- `services/adapter`
  - Owner of provider connectivity and OAuth.
  - Exposes shard-level I/O used by orchestrator.

- `services/shardmap`
  - Owner of shard metadata persistence and lookup.

- `services/shared`
  - Cross-service contracts, clients, and utilities only.
  - No service-owned business logic.

Legacy monolith code under top-level `internal/` has been removed.

## End-to-End Workflow Ownership

Upload workflow:

1. Gateway receives the request on `/api/v1/upload`.
2. Gateway forwards to orchestrator.
3. Orchestrator receives file upload.
4. Orchestrator calls sharding to split/encode shards.
5. Orchestrator calls adapter to upload shards to providers.
6. Orchestrator calls shardmap to register and record shard metadata.

Download workflow:

1. Gateway receives the request on `/api/v1/download/{fileId}`.
2. Gateway forwards to orchestrator.
3. Orchestrator requests shard locations from shardmap.
4. Orchestrator downloads shards through adapter.
5. Orchestrator calls sharding to reconstruct original data.
6. Orchestrator returns file bytes to gateway, and gateway returns them to the caller.

## Gateway API Contract (Public)

Canonical versioned endpoints:

- `POST /api/v1/upload`
- `GET /api/v1/download/{fileId}`
- `GET /api/v1/providers`
- `GET /api/v1/health`
- `GET /api/v1/docs`

Compatibility redirects exist for non-versioned forms:

- `/upload`
- `/download/{fileId}`
- `/providers`
- `/health`

## Service Ports (default local)

- Adapter: `:8080`
- Shardmap: `:8081`
- Orchestrator: `:8082`
- Sharding: `:8083`
- Gateway: `:8084`

## Environment Variables

Common service URLs:

- `ADAPTER_URL` default `http://localhost:8080`
- `SHARDMAP_URL` default `http://localhost:8081`
- `SHARDING_URL` default `http://localhost:8083`
- `ORCHESTRATOR_URL` default `http://localhost:8082`

Adapter local persistence:

- `Omnishard_DB_PATH`

Shardmap local persistence:

- `Omnishard_SHARDMAP_DB_PATH`

Testing-only environment file:

- `backend/.env.example` exists only for manual integration tests.
- Backend service code should not depend on `.env` for credentials.
- Google Drive and S3 credentials are expected to be configured through the application and stored in the local SQLite store.

## Run Locally

From `backend/`:

```powershell
go run ./services/adapter/cmd/main.go
go run ./services/shardmap/cmd/main.go
go run ./services/sharding/cmd/main.go
go run ./services/orchestrator/cmd/main.go
```

Gateway is provided by Docker (NGINX):

```powershell
docker compose --profile backend up --build gateway
```

## Run With Docker Compose

From project root:

```powershell
docker compose --profile full up --build
```

This starts the full stack:

- Frontend: `http://localhost:3000`
- Adapter API: `http://localhost:8080`
- Shardmap API: `http://localhost:8081`
- Orchestrator API: `http://localhost:8082`
- Sharding API: `http://localhost:8083`
- Gateway API: `http://localhost:8084`

For backend services only (no frontend):

```powershell
docker compose --profile backend up --build
```

Stop and remove containers:

```powershell
docker compose --profile full down
```

For backend-only mode:

```powershell
docker compose --profile backend down
```

Stop and also remove local service data volumes:

```powershell
docker compose --profile full down -v
```

If you get container name conflicts from old runs, clean orphaned containers first:

```powershell
docker compose --profile full down --remove-orphans
```

Service endpoints:

- Adapter: `http://localhost:8080`
- Shardmap: `http://localhost:8081`
- Orchestrator: `http://localhost:8082`
- Sharding/Reconstruction: `http://localhost:8083`
- Gateway: `http://localhost:8084`

Internal service DNS used in Docker network:

- `http://adapter:8080`
- `http://shardmap:8081`
- `http://orchestrator:8082`
- `http://sharding:8083`
- `http://gateway:8084`

## Release deployment flavors

Two release-focused Docker Compose manifests now sit alongside the source-build developer workflow:

- `deploy/compose/full-microservices.yml`
  - Pulls published images for adapter, shardmap, sharding, orchestrator, gateway, and frontend.
  - Best for debugging service boundaries while avoiding local builds.
- `deploy/compose/single-image-microservices.yml`
  - Pulls one published `omnishard-all-in-one` image.
  - Runs adapter, shardmap, sharding, orchestrator, gateway, and frontend as separate internal processes in one container.

For end users, the official deployment entrypoint is the GitHub Releases page rather than the repo checkout. Download the latest release asset you want and save it locally as `docker-compose.yml`:

```powershell
wget -O docker-compose.yml https://github.com/Vindyang/cs464-project/releases/latest/download/docker-compose.full-microservices.yml
wget -O docker-compose.yml https://github.com/Vindyang/cs464-project/releases/latest/download/docker-compose.single-image-microservices.yml
```

Then run:

```powershell
docker compose up -d
```

The official release assets are generated during the formal GitHub Release flow, hardcode Docker Hub namespace `nebula67`, and pin a semver image tag.

If you are testing from the repository checkout instead, use the repo-local manifests and set:

```powershell
$env:DOCKERHUB_NAMESPACE = "nebula67"
$env:OMNISHARD_TAG = "<release-tag-or-commit-sha>"
```

Run release flavor 1:

```powershell
docker compose -f deploy/compose/full-microservices.yml up -d
```

Run release flavor 2:

```powershell
docker compose -f deploy/compose/single-image-microservices.yml up -d
```

Single-image public ports:

- Frontend: `http://localhost:3000`
- Adapter/API surface: `http://localhost:8080`

The gateway remains internal to the all-in-one container on port `8084`.

## Tests

Service suite:

```powershell
go test ./services/...
```

Central test runner:

```powershell
.\tests\run-tests.ps1 -Type unit
.\tests\run-tests.ps1 -Type all
```

Integration contract suite:

- `tests/integration/orchestrator_contract_test.go`
  - Happy-path contract test: upload + download across mocked Adapter/ShardMap/Sharding.
- `tests/integration/orchestrator_upload_failure_contracts_test.go`
  - Upload failure-path contracts:
    - shard-map register failure during upload returns orchestrator `500` with stable JSON error fields (`error`, `details`).
    - malformed sharding shard response during upload returns orchestrator `500` with stable JSON error fields.
    - partial adapter shard upload failure triggers rollback delete calls for successfully uploaded shards, and shard-map record is skipped.
- `tests/integration/orchestrator_download_failure_contracts_test.go`
  - Download failure-path contracts:
    - shard-map lookup failure during download returns orchestrator `500` with stable JSON error fields (`error`, `details`).
    - sharding reconstruct returning `500` returns orchestrator `500` with stable JSON error fields (`error`, `details`).
    - sharding reconstruct returning malformed JSON payload returns orchestrator `500` with stable JSON error fields (`error`, `details`).
- `tests/integration/integration_helpers_test.go`
  - Shared integration helpers for orchestrator startup, upload request helpers, health waiting, and dynamic port allocation.

## Service Contract Notes (important)

To keep orchestrator <-> service integration stable, preserve these contracts:

- Orchestrator -> ShardMap
  - Register request must send positive `original_size`.
  - Record request shard `type` must be uppercase `DATA` / `PARITY`.

- Orchestrator -> Sharding
  - Endpoints are `/api/sharding/shard` and `/api/sharding/reconstruct`.
  - Shard request payload uses `fileId`, `fileData`, `n`, `k`.
  - Shard response parsing should support `shardData` (with legacy `shard_data` fallback).

## Key Entry Endpoints

- Gateway upload: `POST /api/v1/upload`
- Gateway download: `GET /api/v1/download/{fileId}`
- Gateway providers: `GET /api/v1/providers`
- Gateway health: `GET /api/v1/health`
- Gateway docs: `GET /api/v1/docs`

Internal service endpoints:

- Orchestrator health: `GET /health`
- Orchestrator upload: `POST /api/orchestrator/upload`
- Orchestrator download: `GET /api/orchestrator/files/{fileId}/download`
- Adapter providers: `GET /api/providers`
- Adapter shard upload: `POST /shards/upload`
- Adapter shard download/delete: `GET|DELETE /shards/{remoteId}`
- Shardmap register/record/query: `/api/v1/shards/*`
- Sharding operations: `/api/sharding/shard`, `/api/sharding/reconstruct`

## Notes

- Gateway intentionally does not add authentication yet.
- OAuth endpoints remain owned by `adapter`.

## Docker Hub CD for Individual Services

This backend is published through GitHub Actions in two distinct modes.

Continuous publishing from `main`:

- `.github/workflows/ci-main.yml`

Official GitHub Release publication:

- `.github/workflows/release-github-oss.yml`

Manual image-only republish:

- `.github/workflows/cd-dockerhub-force-deploy.yml`

Per-service repositories:

- `nebula67/omnishard-adapter`
- `nebula67/omnishard-shardmap`
- `nebula67/omnishard-sharding`
- `nebula67/omnishard-orchestrator`
- `nebula67/omnishard-gateway`

GitHub repository setup required:

- Secret: `DOCKERHUB_USERNAME`
- Secret: `DOCKERHUB_TOKEN`

How it behaves:

- `ci-main.yml` publishes changed images on push to `main` using `latest` plus full commit SHA tags.
- `microservices` runs CI only and does not publish images.
- `release-github-oss.yml` takes an exact commit SHA plus a semver tag, pushes release-tagged images to `nebula67`, creates the GitHub Release, and uploads `docker-compose.full-microservices.yml` plus `docker-compose.single-image-microservices.yml` as official release assets.
