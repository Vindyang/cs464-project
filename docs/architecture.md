# Architecture

## Overview

Omnishard ships one frontend and two backend implementations.

- The frontend is a Next.js application for file management, provider onboarding, settings, lifecycle history, and download flows.
- The microservice backend splits responsibilities across adapter, shardmap, sharding, orchestrator, and gateway services.
- The monolith backend collapses those responsibilities into one Go process while preserving the same frontend-facing workflows.

Both backend variants use the same high-level storage model:

- provider credentials, OAuth tokens, and app settings are stored in `Omnishard.db`
- file metadata, shard placement, lifecycle history, and health state are stored in `Omnishard-shardmap.db`
- shard bytes are stored in external providers such as Google Drive, OneDrive, and AWS S3

## Database Schema

Both backend variants use the same two-file SQLite layout. The microservice deployment splits ownership across the adapter and shardmap services, while the monolith opens both databases inside one process.

### `Omnishard.db`

This database stores provider onboarding state and runtime configuration.

| Table | Primary key | Purpose | Key columns |
| --- | --- | --- | --- |
| `provider_tokens` | `provider_id` | Stores the active OAuth token set for a provider. | `access_token`, `refresh_token`, `token_type`, `expiry`, `updated_at` |
| `credentials` | `provider_id` | Stores provider OAuth client credentials configured through the UI or environment-backed setup. | `client_id`, `client_secret`, `redirect_uri`, `updated_at` |
| `provider_config` | `key` | Stores app-level provider configuration and settings values. | `value`, `updated_at` |

Operational notes:

- `provider_id` identifies a concrete storage provider integration such as Google Drive or OneDrive.
- Token and credential rows are upserted, so reconnecting or reconfiguring a provider replaces the existing record.
- This database does not store shard bytes or file placement metadata.

### `Omnishard-shardmap.db`

This database stores file metadata, shard placement, lifecycle history, and health refresh timestamps.

| Table | Primary key | Purpose | Key columns |
| --- | --- | --- | --- |
| `files` | `id` | One row per uploaded logical file. | `original_name`, `original_size`, `total_chunks`, `n`, `k`, `shard_size`, `status`, `created_at`, `updated_at`, `last_health_refresh_at` |
| `shards` | `id` | One row per persisted shard fragment. | `file_id`, `chunk_index`, `shard_index`, `shard_type`, `remote_id`, `provider`, `checksum_sha256`, `status`, `created_at`, `updated_at` |
| `file_lifecycle_log` | `id` | Append-only history of upload, download, refresh, and related workflow events. | `file_id`, `event_type`, `file_name`, `file_size`, `shard_count`, `providers`, `started_at`, `ended_at`, `duration_ms`, `status`, `error_msg`, `created_at` |

Relationship notes:

- `files.id` is the parent identifier for logical file state.
- `shards.file_id` references `files.id` with `ON DELETE CASCADE`, so deleting a file record also removes its shard placement rows.
- `file_lifecycle_log.file_id` links lifecycle events back to the logical file, but serves as an audit/history stream rather than placement state.
- `last_health_refresh_at` on `files` captures the most recent health probe time, while shard-level `status` captures fragment availability.

### Logical ER View

```text
Omnishard.db
  provider_tokens (provider_id PK)
  credentials     (provider_id PK)
  provider_config (key PK)

Omnishard-shardmap.db
  files               (id PK)
      |
      +----< shards              (id PK, file_id FK -> files.id)
      |
      +----< file_lifecycle_log  (id PK, file_id -> files.id)
```

The result is a clean split: provider connection state lives in `Omnishard.db`, durable file and shard metadata lives in `Omnishard-shardmap.db`, and the actual shard payloads remain external in the configured cloud providers.

## Shared System Shape

```text
Browser
  |
  v
Frontend (3000)
  |
  +--> direct metadata/settings/provider calls
  |
  +--> workflow calls (upload, download, history, refresh)
```

The difference between the two backend variants is where those frontend calls terminate.

## Backend Variant 1: Microservice Topology

```text
Frontend (3000)
  |                    \
  | direct metadata      \ workflow routes
  v                       v
Adapter (8080)         Gateway (8084)
  |                        |
  |                        v
  |                   Orchestrator (8082)
  |                     /          \
  v                    v            v
Omnishard.db       Shardmap (8081) Sharding (8083)
                        |
                        v
                Omnishard-shardmap.db
```

Characteristics:

- Strict service ownership boundaries.
- HTTP is used between backend services.
- Gateway owns the public versioned API contract and docs surface.
- Best fit when you need explicit service seams for debugging, testing, or deployment experimentation.

Detailed microservice reference:

- [../backend/microservice/README.md](../backend/microservice/README.md)
- [backend-microservice.md](backend-microservice.md)

## Backend Variant 2: Monolith Topology

```text
Frontend (3000)
  |
  v
Monolith (8080)
  |
  +--> provider connectivity and credentials
  +--> metadata persistence
  +--> workflow orchestration
  +--> sharding and reconstruction
  +--> docs and health endpoints
```

Internally, the monolith still preserves package-level separation for app wiring, shardmap persistence, sharding logic, and workflow orchestration, but those calls are in-process instead of HTTP hops.

Characteristics:

- Single Go backend process with one public HTTP surface.
- No NGINX gateway container is required.
- The frontend points both metadata and workflow URLs at the same backend.
- Best fit when you want a simpler deployment model or want to iterate on a standalone backend implementation.

Detailed monolith reference:

- [../backend/monolith/README.md](../backend/monolith/README.md)
- [backend-monolith.md](backend-monolith.md)

## Shared Functional Flows

### Upload

1. The frontend receives a file from the user.
2. The backend shards the file into `n` fragments with recovery threshold `k`.
3. The backend registers file metadata before persisting shard placement.
4. Shards are uploaded to one or more configured providers.
5. Placement metadata and lifecycle history are recorded.

### Download

1. The frontend requests file download by `fileId`.
2. The backend resolves shard placement and provider ownership.
3. The backend downloads enough viable shards to satisfy `k`.
4. The original file is reconstructed and streamed back to the caller.
5. A download lifecycle event is recorded.

### Health refresh

1. The frontend requests a refresh for one file or all files.
2. The backend probes shard availability through the active provider adapters.
3. Shard health and file-level health summaries are updated in the metadata store.

### Reset actions

Settings reset flows can clear:

- file metadata and shard metadata
- provider-side shards when requested
- stored credentials and provider tokens
- lifecycle history

## Frontend Integration Model

The frontend always runs on `http://localhost:3000`, but the backend URL map changes by deployment flavor:

- Microservice profiles:
  - `NEXT_PUBLIC_API_URL=http://localhost:8080`
  - `API_INTERNAL_URL=http://adapter:8080`
  - `GATEWAY_URL=http://gateway:8084`
- Monolith profile:
  - `NEXT_PUBLIC_API_URL=http://localhost:8080`
  - `API_INTERNAL_URL=http://monolith:8080`
  - `GATEWAY_URL=http://monolith:8080`

## Further Reading

- [backend-microservice.md](backend-microservice.md)
- [backend-monolith.md](backend-monolith.md)
- [cicd.md](cicd.md)