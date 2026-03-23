# Shared Package Guidelines

`services/shared` is now a library area for cross-cutting code and cross-service clients.
It must not contain service-owned business logic.

## Import path

Use:

`github.com/vindyang/cs464-project/backend/services/shared/<package>`

## Shared packages (allowed)

- `adapter`: Provider interfaces, metadata, and adapter registry.
- `api/dto`: Shared DTO contracts (currently unversioned; may be versioned later).
- `api/middleware`: Reusable HTTP middleware (logging, CORS, recovery).
- `clients/adapter`: HTTP client for adapter shard and provider endpoints.
- `clients/sharding`: HTTP client for the sharding owner service.
- `clients/shardmapworkflow`: HTTP client for workflow-level shardmap operations used by orchestrator.
- `db`: Token DB wrapper and persistence helpers.
- `database`: Generic SQL bootstrap helpers.
- `models`: Shared model primitives (to be narrowed over time to truly cross-cutting types).
- `oauthhandler`: OAuth helpers used by adapter flows.
- `transport/httpx`: Shared HTTP helpers (`WriteJSON`, `WriteError`, `DecodeJSON`, `RequireMethod`).
- `types`: Shared transport payload types used by orchestrator flows.

Removed legacy package:

- `orchestrator/clients` (deleted). Use `clients/adapter` and `clients/shardmapworkflow`.

## What was migrated out

The following shared business implementations were removed and moved to owning services:

- `shared/service/*` -> moved into service-local app packages.
- `shared/repository/*` -> moved into `services/shardmap/internal/infra/repository`.
- `shared/orchestrator/service.go` and `shared/orchestrator/models.go` -> moved into `services/orchestrator/internal/app`.

Additional cleanup:

- Top-level legacy `backend/internal/*` monolith package tree was removed.

## Rule of thumb

If code enforces domain behavior for one service, it belongs in that service's `internal` tree.
If code is a reusable utility/client/contract used by multiple services, it can live in `shared`.
