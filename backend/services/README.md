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

## Running Services

From `backend/`:

```powershell
go run ./services/adapter/cmd/main.go
go run ./services/orchestrator/cmd/main.go
go run ./services/shardmap/cmd/main.go
go run ./services/sharding/cmd/main.go
```

## Running Tests

Use the central runner:

```powershell
.\tests\run-tests.ps1 -Type unit
.\tests\run-tests.ps1 -Type integration
.\tests\run-tests.ps1 -Type e2e
.\tests\run-tests.ps1 -Type all
```
