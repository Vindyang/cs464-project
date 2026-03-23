# Shared Service Modules

This package centralizes modules that were duplicated across individual services.

## Import path

Use:

`github.com/vindyang/cs464-project/backend/services/shared/<package>`

## Shared packages

- `adapter`: Provider interfaces, metadata, and adapter registry.
- `api/dto`: Common request/response DTOs for file operations, sharding, and shard maps.
- `api/middleware`: Reusable HTTP middleware (logging, CORS, recovery).
- `database`: SQL database bootstrap helpers.
- `db`: Token DB wrapper and persistence helpers.
- `models`: Core domain models (`File`, `Shard`, status enums).
- `oauthhandler`: OAuth flows and provider connect/disconnect handlers.
- `orchestrator`: Orchestrator service layer and HTTP clients.
- `repository`: File and shard repository interfaces/implementations.
- `service`: Core service implementations (file operations, shard map, sharding).
- `types`: Shared cross-service API payload types.

## Service-specific code stays local

Each service keeps only service-specific packages in its own `internal` directory (for example HTTP handlers and service entrypoints in `cmd`).
