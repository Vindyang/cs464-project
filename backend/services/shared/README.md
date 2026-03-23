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
- `clients/sharding`: HTTP client for the sharding owner service.
- `clients/shardmap`: HTTP client for the shardmap owner service.
- `db`: Token DB wrapper and persistence helpers.
- `database`: Generic SQL bootstrap helpers.
- `models`: Shared model primitives (to be narrowed over time to truly cross-cutting types).
- `oauthhandler`: OAuth helpers used by adapter flows.
- `transport/httpx`: Shared HTTP helpers (`WriteJSON`, `WriteError`, `DecodeJSON`, `RequireMethod`).
- `types`: Shared transport payload types used by orchestrator flows.

## Legacy path kept temporarily

- `orchestrator/clients`: existing orchestrator HTTP clients still used by orchestrator.
	- This path is transport-only client code, not business logic.
	- It can be moved to `shared/clients/*` in a follow-up cleanup.

## What was migrated out

The following shared business implementations were removed and moved to owning services:

- `shared/service/*` -> moved into service-local app packages.
- `shared/repository/*` -> moved into `services/shardmap/internal/infra/repository`.
- `shared/orchestrator/service.go` and `shared/orchestrator/models.go` -> moved into `services/orchestrator/internal/app`.

## Rule of thumb

If code enforces domain behavior for one service, it belongs in that service's `internal` tree.
If code is a reusable utility/client/contract used by multiple services, it can live in `shared`.
