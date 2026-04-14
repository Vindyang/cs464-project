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