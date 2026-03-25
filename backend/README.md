# Nebula Drive Backend

Nebula Drive is organized as independently deployable Go services with strict ownership boundaries.
The orchestrator is the workflow owner and coordinates sharding, storage, and shard metadata.

## Current Architecture

- `services/orchestrator`
  - Public workflow entry/exit for upload and retrieval.
  - Calls `sharding`, `adapter`, and `shardmap` through HTTP clients.

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
1. Orchestrator receives file upload.
2. Orchestrator calls sharding to split/encode shards.
3. Orchestrator calls adapter to upload shards to providers.
4. Orchestrator calls shardmap to register and record shard metadata.

Download workflow:
1. Orchestrator requests shard locations from shardmap.
2. Orchestrator downloads shards through adapter.
3. Orchestrator calls sharding to reconstruct original data.
4. Orchestrator returns file bytes to the caller.

## Service Ports (default local)

- Adapter: `:8080`
- Shardmap: `:8081`
- Orchestrator: `:8082`
- Sharding: `:8083`

## Environment Variables

Common service URLs:

- `ADAPTER_URL` default `http://localhost:8080`
- `SHARDMAP_URL` default `http://localhost:8081`
- `SHARDING_URL` default `http://localhost:8083`

Adapter OAuth/token storage:

- `DATABASE_URL`
- `GDRIVE_OAUTH_CREDENTIALS_FILE`
- `GDRIVE_OAUTH_REDIRECT_URI`
- `GDRIVE_FOLDER_ID`
- `FRONTEND_URL`

## Run Locally

From `backend/`:

```powershell
go run ./services/adapter/cmd/main.go
go run ./services/shardmap/cmd/main.go
go run ./services/sharding/cmd/main.go
go run ./services/orchestrator/cmd/main.go
```

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

- Orchestrator upload: `POST /api/orchestrator/upload`
- Orchestrator download: `GET /api/orchestrator/files/{fileId}/download`
- Adapter providers: `GET /api/providers`
- Adapter shard upload: `POST /shards/upload`
- Adapter shard download/delete: `GET|DELETE /shards/{remoteId}`
- Shardmap register/record/query: `/api/v1/shards/*`
- Sharding operations: `/api/sharding/shard`, `/api/sharding/reconstruct`
