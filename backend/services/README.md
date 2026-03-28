# Services Layout

This directory organizes the backend by microservice boundary.

## Services

- `adapter/`
  - Owns provider connectivity and OAuth-facing adapter endpoints.
  - Entrypoint: `services/adapter/cmd/main.go`
  - Unit tests: `services/adapter/tests/unit/`

- `orchestrator/`
  - Owns workflow orchestration across backend services.
  - Entrypoint: `services/orchestrator/cmd/main.go`
  - Unit tests: `services/orchestrator/tests/unit/`

- `shardmap/`
  - Owns shard metadata persistence and shard map APIs.
  - Entrypoint: `services/shardmap/cmd/main.go`
  - Unit tests: `services/shardmap/tests/unit/`

- `sharding/`
  - Owns sharding and reconstruction operations.
  - Entrypoint: `services/sharding/cmd/main.go`
  - Unit tests: `services/sharding/tests/unit/`

- `gateway/`
  - Owns the public API contract and versioned external interface.
  - Exposes the standardized API surface for upload, download, health, and providers.
  - Implemented as NGINX reverse proxy config: `services/gateway/nginx.conf.template`

## Service Ownership Map

- `gateway` owns public endpoint standardization and forwards to owner services.
  - `POST /api/v1/upload`
  - `GET /api/v1/download/{fileId}`
  - `GET /api/v1/health`
  - `GET /api/v1/providers`
  - `GET /api/v1/docs`
- `adapter` owns provider connectivity, OAuth, and shard-level provider I/O endpoints.
- `orchestrator` owns workflow entry/exit endpoints and orchestrates all cross-service workflow steps.
- `shardmap` owns shard-map metadata endpoints:
  - `POST /api/v1/shards/register`
  - `POST /api/v1/shards/record`
  - `GET /api/v1/shards/file/{fileId}`
  - `GET /api/v1/shards/{shardId}`
  - `PUT /api/v1/shards/{shardId}/status`
- `sharding` owns sharding endpoints:
  - `GET /api/sharding/health`
  - `POST /api/sharding/shard`
  - `POST /api/sharding/reconstruct`

Non-owning services must call owner services through client packages under `services/shared/clients/*`.

## Shared Clients

- `shared/clients/adapter`: orchestrator client for provider health and shard upload/download/delete through adapter.
- `shared/clients/sharding`: used by non-owner services when they need sharding/reconstruction capability.
- `shared/clients/shardmapworkflow`: used by orchestrator for workflow shard-map APIs.

Current usage:

- `orchestrator` calls `sharding`, `adapter`, and `shardmap` over HTTP via shared clients.
- `adapter` is no longer the workflow coordinator.

## End-to-End Ownership

Upload:
1. Request enters gateway.
2. Gateway forwards to orchestrator.
3. Orchestrator calls sharding to produce shards.
4. Orchestrator calls adapter to upload each shard.
5. Orchestrator calls shardmap to register and record metadata.

Download:
1. Request enters gateway.
2. Gateway forwards to orchestrator.
3. Orchestrator reads shard placement from shardmap.
4. Orchestrator downloads shards via adapter.
5. Orchestrator calls sharding to reconstruct file bytes.

## Running Services

From `backend/`:

```powershell
go run ./services/adapter/cmd/main.go
go run ./services/orchestrator/cmd/main.go
go run ./services/shardmap/cmd/main.go
go run ./services/sharding/cmd/main.go
```

Gateway is run via Docker Compose (NGINX):

```powershell
docker compose --profile backend up --build gateway
```

## Cross-Service URLs

Set service base URLs through env vars when running independently:

```powershell
$env:ADAPTER_URL="http://localhost:8080"
$env:SHARDMAP_URL="http://localhost:8081"
$env:SHARDING_URL="http://localhost:8083"
$env:ORCHESTRATOR_URL="http://localhost:8082"
```

## Running Tests

Use the central runner:

```powershell
.\tests\run-tests.ps1 -Type unit
.\tests\run-tests.ps1 -Type integration
.\tests\run-tests.ps1 -Type e2e
.\tests\run-tests.ps1 -Type all
```
