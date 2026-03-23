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

## Service Ownership Map

- `adapter` owns provider and adapter-facing endpoints.
- `orchestrator` owns orchestration endpoints.
- `shardmap` owns shard-map metadata endpoints:
  - `POST /api/v1/shards/register`
  - `POST /api/v1/shards/record`
  - `GET /api/v1/shards/file/{fileId}`
  - `GET /api/v1/shards/{shardId}`
  - `PUT /api/v1/shards/{shardId}/status`
- `sharding` owns sharding endpoints:
  - `GET /health`
  - `POST /shard`
  - `POST /reconstruct`

Non-owning services must call owner services through client packages under `services/shared/clients/*`.

## Shared Clients

- `shared/clients/sharding`: used by non-owner services when they need `/shard` or `/reconstruct` capability.
- `shared/clients/shardmap`: used by non-owner services for shard-map APIs.

Current usage:

- `adapter` calls `sharding` and `shardmap` over HTTP via shared clients.
- `orchestrator` calls `adapter` and `shardmap` over HTTP via client packages.

## Running Services

From `backend/`:

```powershell
go run ./services/adapter/cmd/main.go
go run ./services/orchestrator/cmd/main.go
go run ./services/shardmap/cmd/main.go
go run ./services/sharding/cmd/main.go
```

## Cross-Service URLs

Set service base URLs through env vars when running independently:

```powershell
$env:SHARDMAP_BASE_URL="http://localhost:8081"
$env:SHARDING_BASE_URL="http://localhost:8083"
```

## Running Tests

Use the central runner:

```powershell
.\tests\run-tests.ps1 -Type unit
.\tests\run-tests.ps1 -Type integration
.\tests\run-tests.ps1 -Type e2e
.\tests\run-tests.ps1 -Type all
```
