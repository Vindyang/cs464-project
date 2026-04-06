# CS464 Project

Omnishard is a distributed cloud storage system built as a small microservice platform. It takes a user file, splits it into Reed-Solomon data and parity shards, stores those shards across cloud providers, records shard placement and lifecycle metadata, and reconstructs the file from any viable subset of shards during download.

This README is the primary operator and contributor guide. It covers installation, local development, service responsibilities, exposed APIs, database schemas, testing, CI/CD, and deployment options.

## Table of Contents

1. System Overview
2. Repository Layout
3. Architecture and End-to-End Flows
4. Services
5. Persistence and Database Schemas
6. Installation and Local Development
7. Docker and Deployment Options
8. Testing
9. CI/CD
10. Useful Commands

## System Overview

At a high level, Omnishard is composed of:

- A Next.js frontend for file management, credentials, providers, history, and settings.
- A Go adapter service that manages provider connectivity, credentials, tokens, shard I/O, and application settings.
- A Go shardmap service that owns file metadata, shard metadata, lifecycle history, and health state.
- A Go sharding service that performs Reed-Solomon shard and reconstruct operations.
- A Go orchestrator service that coordinates upload, download, delete, and health-refresh workflows.
- An NGINX gateway that exposes a stable public API surface over the backend workflow services.

### What the system stores

- File metadata and shard placement live in the shardmap SQLite database.
- Provider credentials, OAuth tokens, and app settings live in the adapter SQLite database.
- Actual shard bytes live in external storage providers such as Google Drive, OneDrive, and AWS S3.
- Upload, download, and delete lifecycle logs live in the shardmap SQLite database.

## Repository Layout

```text
cs464-project/
├── backend/
│   ├── services/
│   │   ├── adapter/
│   │   ├── gateway/
│   │   ├── orchestrator/
│   │   ├── shardmap/
│   │   ├── sharding/
│   │   └── shared/
│   └── tests/
├── frontend/
├── deploy/
│   ├── compose/
│   ├── images/all-in-one/
│   └── release-assets/
├── docker-compose.yml
└── .github/workflows/
```

## Architecture and End-to-End Flows

### Service Topology

```text
Browser
  |
  |  UI + Next app routes
  v
Frontend (3000)
  |                    \
  | direct metadata      \ upload/download/history
  v                       v
Adapter (8080)         Gateway (8084)
  |                        |
  | settings/creds/files   v
  |                     Orchestrator (8082)
  |                      /      |       \
  |                     /       |        \
  v                    v        v         v
SQLite               Adapter  Shardmap  Sharding
Omnishard.db         (8080)   (8081)    (8083)
                                |
                                v
                         SQLite Omnishard-shardmap.db
```

### Upload Flow

1. The user uploads a file from the frontend.
2. The frontend calls its internal upload route, which proxies to the gateway.
3. The gateway forwards the request to the orchestrator.
4. The orchestrator calls the sharding service to split the file into `n` shards with recovery threshold `k`.
5. The orchestrator registers the file with the shardmap service.
6. The orchestrator uploads shards to the adapter service, which writes them to the selected cloud providers.
7. The orchestrator records shard placement metadata in shardmap.
8. The orchestrator logs an upload lifecycle event in shardmap.

### Download Flow

1. The user requests a download from the frontend.
2. The frontend calls its internal download route, which proxies to the gateway.
3. The gateway forwards the request to the orchestrator.
4. The orchestrator queries shardmap for the file's shard map.
5. The orchestrator downloads shards from the adapter service.
6. Once at least `k` valid shards are available, the orchestrator reconstructs the original file through the sharding service.
7. The orchestrator returns the file bytes and logs a download lifecycle event.

### Delete Flow

1. The user deletes a file from the UI.
2. The request goes to the adapter service.
3. The adapter optionally deletes provider-side shards.
4. The adapter deletes file and shard metadata through shardmap.
5. The orchestrator or shardmap logs the delete lifecycle event.

### Health Refresh Flow

1. The user requests a health refresh for one file or all files.
2. The frontend calls its internal API route or directly triggers backend refresh paths.
3. The orchestrator probes shard existence through the adapter.
4. The shardmap service updates shard status fields and file health summaries.

### Reset Flow

The settings danger-zone actions support:

- deleting all file data,
- deleting all credentials and provider tokens,
- deleting all data, including lifecycle history.

The `all_data` reset now clears:

- file metadata,
- shard metadata,
- provider-side shards when requested,
- stored credentials,
- provider tokens,
- lifecycle history logs.

## Services

### Service Matrix

| Service | Port | Primary Role | Persistent Storage | Main Dependencies |
| --- | --- | --- | --- | --- |
| Frontend | `3000` | UI, SSR, internal proxy routes | none | Adapter, Gateway |
| Adapter | `8080` | Provider connectivity, credentials, settings, shard I/O, metadata proxy | `Omnishard.db` | Shardmap, cloud providers |
| Shardmap | `8081` | File metadata, shard metadata, lifecycle history, health state | `Omnishard-shardmap.db` | none |
| Orchestrator | `8082` | Upload/download/delete/health workflows | none | Adapter, Shardmap, Sharding |
| Sharding | `8083` | Reed-Solomon shard/reconstruct operations | none | none |
| Gateway | `8084` | Stable public API contract and reverse proxy | none | Orchestrator, Adapter |

### Frontend Service

The frontend is a Next.js 16 app using React 19, TypeScript, Tailwind CSS, `next-themes`, `sonner`, and a small number of custom UI primitives.

#### Main user-facing routes

- `/dashboard` — system overview, provider usage, recent files, degraded files, credential banner.
- `/files` — file listing and file-health view.
- `/files/[fileId]` — file metadata, recent lifecycle activity, shard map.
- `/history` — global lifecycle history.
- `/providers` — provider management and upload entrypoint.
- `/settings` — redundancy defaults, storage behaviors, destructive resets.
- `/credentials` — provider credential management.

#### Frontend internal API routes

These routes exist so the frontend can proxy requests to the gateway or backend services while keeping browser code simpler:

- `POST /api/upload`
- `GET /api/download/{fileId}`
- `GET /api/history/{fileId}`
- `DELETE /api/files/{fileId}`
- `POST /api/files/health/refresh`
- `POST /api/files/{fileId}/health/refresh`

#### How the frontend talks to the backend

The frontend uses two patterns:

- It talks directly to the adapter service for settings, credentials, provider metadata, file metadata, and shard-map lookups.
- It uses the gateway for upload, download, and lifecycle history routes that belong to workflow orchestration.

Important environment variables:

- `NEXT_PUBLIC_API_URL` — browser-visible adapter URL, usually `http://localhost:8080`
- `API_INTERNAL_URL` — server-side adapter URL, usually `http://adapter:8080` in Docker
- `GATEWAY_URL` — server-side gateway URL, usually `http://gateway:8084`

### Adapter Service

The adapter service is the provider integration boundary. It owns credential storage, token storage, provider restoration on startup, shard upload/download/delete calls, settings, and the frontend-facing metadata proxy for files and shards.

#### Default port

- `8080`

#### Responsibilities

- Persist provider credentials and OAuth tokens.
- Restore Google Drive, OneDrive, and S3 provider adapters on startup.
- Expose shard upload/download/delete APIs used by orchestrator.
- Proxy file and shard metadata reads to shardmap.
- Own app settings and destructive reset actions.

#### Exposed APIs

Core service:

- `GET /health`
- `GET /api/providers`

Google Drive OAuth:

- `GET /api/oauth/gdrive/authorize`
- `GET /api/oauth/gdrive/callback`
- `POST /api/oauth/gdrive/disconnect`

OneDrive OAuth:

- `GET /api/oauth/onedrive/authorize`
- `GET /api/oauth/onedrive/callback`
- `POST /api/oauth/onedrive/disconnect`

AWS S3 provider management:

- `POST /api/providers/awsS3/connect`
- `POST /api/providers/awsS3/disconnect`

Credentials and status:

- `GET /api/credentials/status`
- `GET /api/credentials`
- `PUT /api/credentials/{providerId}`
- `DELETE /api/credentials/{providerId}`
- `GET /api/credentials/{providerId}/secret`

Settings and reset:

- `GET /api/settings`
- `PUT /api/settings`
- `POST /api/settings/reset`

Shard I/O for orchestrator:

- `POST /shards/upload`
- `GET /shards/{remoteId}?provider={providerId}`
- `DELETE /shards/{remoteId}?provider={providerId}`

Metadata proxying to shardmap:

- `GET /api/v1/files`
- `GET /api/v1/files/{fileId}`
- `DELETE /api/v1/files/{fileId}`
- `GET /api/v1/shards/file/{fileId}`

### Shardmap Service

The shardmap service owns the canonical metadata model for files, shards, lifecycle history, and shard health. It is the authoritative persistence layer for workflow state.

#### Default port

- `8081`

#### Responsibilities

- Register uploaded files before shards are persisted.
- Record shard placement after provider upload succeeds.
- Return file and shard metadata for download and UI workflows.
- Track shard health and file-level health summaries.
- Persist lifecycle logs for upload, download, and delete events.

#### Exposed APIs

- `GET /health`
- `POST /api/v1/shards/register`
- `POST /api/v1/shards/record`
- `GET /api/v1/shards/file/{fileId}`
- `GET /api/v1/shards/{shardId}`
- `PUT /api/v1/shards/{shardId}/status`
- `GET /api/v1/files`
- `GET /api/v1/files/{fileId}`
- `DELETE /api/v1/files/{fileId}`
- `POST /api/v1/files/{fileId}/health-refresh`
- `POST /api/v1/lifecycle`
- `GET /api/v1/lifecycle`
- `DELETE /api/v1/lifecycle`
- `GET /api/v1/lifecycle/{fileId}`

### Orchestrator Service

The orchestrator is the workflow owner. It is stateless and coordinates upload, download, history, delete, and health-refresh flows across the other services.

#### Default port

- `8082`

#### Responsibilities

- Accept uploads and route them through sharding, provider I/O, and metadata recording.
- Retrieve shard maps, fetch shards, and reconstruct files during download.
- Refresh shard health across stored files.
- Expose lifecycle history through the gateway-facing contract.

#### Exposed APIs

- `GET /health`
- `POST /api/orchestrator/upload`
- `GET /api/orchestrator/history`
- `POST /api/orchestrator/files/health/refresh`
- `POST /api/orchestrator/files/{fileId}/health/refresh`
- `GET /api/orchestrator/files/{fileId}/download`
- `GET /api/orchestrator/files/{fileId}/history`
- `DELETE /api/orchestrator/files/{fileId}`

### Sharding Service

The sharding service is a stateless Reed-Solomon worker.

#### Default port

- `8083`

#### Responsibilities

- Shard an input file into `n` total shards with threshold `k`.
- Reconstruct the original file from any viable set of shards.

#### Exposed APIs

- `GET /api/sharding/health`
- `POST /api/sharding/shard`
- `POST /api/sharding/reconstruct`

### Gateway Service

The gateway is an NGINX reverse proxy that exposes the stable public contract for clients and release deployments.

#### Default port

- `8084`

#### Responsibilities

- Route public API paths to the correct service.
- Enforce request method boundaries.
- Attach request IDs and structured logs.
- Serve gateway documentation assets.

#### Public APIs

- `GET /`
- `GET /api/v1/docs`
- `GET /api/v1/openapi.yml`
- `POST /api/v1/upload`
- `GET /api/v1/download/{fileId}`
- `GET /api/v1/history`
- `GET /api/v1/history/{fileId}`
- `POST /api/v1/files/health/refresh`
- `POST /api/v1/files/{fileId}/health/refresh`
- `DELETE /api/v1/files/{fileId}`
- `GET /api/v1/providers`
- `GET /api/v1/health`

Legacy redirects also exist for:

- `/upload`
- `/download/{fileId}`
- `/history`
- `/history/{fileId}`
- `/files/{fileId}`
- `/providers`
- `/health`

## Persistence and Database Schemas

Omnishard uses two SQLite databases plus cloud-provider object storage.

### Adapter database

Default path:

- `Omnishard.db`

Configured by:

- `Omnishard_DB_PATH`

Tables:

#### `provider_tokens`

Stores OAuth or provider session tokens.

| Column | Type | Notes |
| --- | --- | --- |
| `provider_id` | `TEXT PRIMARY KEY` | logical provider identifier |
| `access_token` | `TEXT NOT NULL` | active access token |
| `refresh_token` | `TEXT NOT NULL DEFAULT ''` | refresh token when present |
| `token_type` | `TEXT NOT NULL DEFAULT 'Bearer'` | token type |
| `expiry` | `DATETIME` | token expiry |
| `updated_at` | `DATETIME DEFAULT CURRENT_TIMESTAMP` | audit timestamp |

#### `credentials`

Stores provider client credentials or S3 connection details.

| Column | Type | Notes |
| --- | --- | --- |
| `provider_id` | `TEXT PRIMARY KEY` | provider identifier |
| `client_id` | `TEXT NOT NULL` | OAuth client id or access key id |
| `client_secret` | `TEXT NOT NULL` | OAuth secret or S3 secret access key |
| `redirect_uri` | `TEXT NOT NULL` | redirect URI or region depending on provider |
| `updated_at` | `DATETIME DEFAULT CURRENT_TIMESTAMP` | audit timestamp |

#### `provider_config`

Stores general settings and provider-specific key-value configuration.

| Column | Type | Notes |
| --- | --- | --- |
| `key` | `TEXT PRIMARY KEY` | config key |
| `value` | `TEXT NOT NULL` | config value |
| `updated_at` | `DATETIME DEFAULT CURRENT_TIMESTAMP` | audit timestamp |

Common settings keys include:

- `settings_redundancy`
- `settings_encrypt_default`
- `settings_auto_delete`

### Shardmap database

Default path:

- `Omnishard-shardmap.db`

Configured by:

- `Omnishard_SHARDMAP_DB_PATH`

Tables:

#### `files`

Stores canonical file metadata.

| Column | Type | Notes |
| --- | --- | --- |
| `id` | `TEXT PRIMARY KEY` | file UUID |
| `original_name` | `TEXT` | original filename |
| `original_size` | `INTEGER NOT NULL` | original byte length |
| `total_chunks` | `INTEGER NOT NULL` | total chunk count |
| `n` | `INTEGER NOT NULL` | total shards per chunk |
| `k` | `INTEGER NOT NULL` | minimum shards required to reconstruct |
| `shard_size` | `INTEGER NOT NULL` | shard payload size |
| `status` | `TEXT NOT NULL` | file health / workflow status |
| `created_at` | `DATETIME NOT NULL` | creation timestamp |
| `updated_at` | `DATETIME NOT NULL` | update timestamp |
| `last_health_refresh_at` | `DATETIME` | most recent refresh timestamp |

#### `shards`

Stores shard placement and health state.

| Column | Type | Notes |
| --- | --- | --- |
| `id` | `TEXT PRIMARY KEY` | shard UUID |
| `file_id` | `TEXT NOT NULL` | foreign key to `files(id)` |
| `chunk_index` | `INTEGER NOT NULL` | chunk number |
| `shard_index` | `INTEGER NOT NULL` | shard number within chunk |
| `shard_type` | `TEXT NOT NULL` | data or parity |
| `remote_id` | `TEXT NOT NULL` | provider-side object identifier |
| `provider` | `TEXT NOT NULL` | owning provider |
| `checksum_sha256` | `TEXT NOT NULL` | shard checksum |
| `status` | `TEXT NOT NULL` | shard health state |
| `created_at` | `DATETIME NOT NULL` | creation timestamp |
| `updated_at` | `DATETIME NOT NULL` | update timestamp |

Indexes:

- `idx_shards_file_id`
- `idx_shards_file_chunk`

#### `file_lifecycle_log`

Stores workflow history for upload, download, and delete events.

| Column | Type | Notes |
| --- | --- | --- |
| `id` | `INTEGER PRIMARY KEY AUTOINCREMENT` | event row id |
| `file_id` | `TEXT NOT NULL` | file UUID |
| `event_type` | `TEXT NOT NULL` | upload, download, or delete |
| `file_name` | `TEXT` | filename snapshot |
| `file_size` | `INTEGER` | byte length snapshot |
| `shard_count` | `INTEGER` | shard count snapshot |
| `providers` | `TEXT` | comma-separated providers |
| `started_at` | `DATETIME NOT NULL` | workflow start |
| `ended_at` | `DATETIME NOT NULL` | workflow end |
| `duration_ms` | `INTEGER NOT NULL` | duration in milliseconds |
| `status` | `TEXT NOT NULL` | success or failed |
| `error_msg` | `TEXT` | failure details when present |
| `created_at` | `DATETIME DEFAULT CURRENT_TIMESTAMP` | insertion time |

Indexes:

- `idx_lifecycle_file_id`
- `idx_lifecycle_event_created`

### External provider storage

Shard bytes are persisted outside the application databases in whichever cloud provider owns the shard:

- Google Drive
- Microsoft OneDrive
- AWS S3

The application databases only store metadata and provider object identifiers, not the actual user file payloads.

## Installation and Local Development

### Prerequisites

For source-based development:

- Go `1.25.0`
- Bun `1.x` for the frontend and CI-aligned local builds
- Docker Desktop if you want the full stack, gateway, or E2E workflows

Notes:

- The repo is built in CI and Docker using Bun for the frontend.
- `npm run dev` can still be used locally, but Bun is the canonical package manager for reproducible installs and builds in this repo.

### Where to run Docker Compose commands

Run all Docker Compose commands from the project root:

```text
cs464-project/
```

That directory contains the source-build `docker-compose.yml`.

### Quick start: full stack with local source builds

```powershell
docker compose --profile full up -d --build
```

Default public endpoints:

- Frontend UI: `http://localhost:3000`
- Adapter: `http://localhost:8080`
- Shardmap: `http://localhost:8081`
- Orchestrator: `http://localhost:8082`
- Sharding: `http://localhost:8083`
- Gateway: `http://localhost:8084`

Stop the stack:

```powershell
docker compose --profile full down
```

Wipe local service data volumes:

```powershell
docker compose --profile full down -v
```

### Backend-only Docker workflow

Build backend services:

```powershell
docker compose --profile backend build
```

Start backend services only:

```powershell
docker compose --profile backend up -d
```

Start a single service with dependencies:

```powershell
docker compose --profile backend up -d adapter
docker compose --profile backend up -d shardmap
docker compose --profile backend up -d sharding
docker compose --profile backend up -d orchestrator
docker compose --profile backend up -d gateway
```

Stop backend services:

```powershell
docker compose --profile backend down
```

### Run backend services directly from source

From `backend/`:

```powershell
go run ./services/adapter/cmd/main.go
go run ./services/shardmap/cmd/main.go
go run ./services/sharding/cmd/main.go
go run ./services/orchestrator/cmd/main.go
```

The gateway is NGINX-based and is usually easiest to run through Docker Compose:

```powershell
docker compose --profile backend up --build gateway
```

### Run the frontend from source

From `frontend/`:

```powershell
bun install --frozen-lockfile
bun run dev
```

Alternative local dev command if you already have `node_modules` installed through npm:

```powershell
npm run dev
```

Recommended environment assumptions for local source dev:

- `NEXT_PUBLIC_API_URL=http://localhost:8080`
- `GATEWAY_URL=http://localhost:8084`

## Docker and Deployment Options

The repository supports three deployment styles.

### 1. Source-build developer workflow

File:

- `docker-compose.yml`

Use this when you are developing locally from source and want Docker to build images from the repo checkout.

### 2. Repo-local full-microservices release reference

File:

- `deploy/compose/full-microservices.yml`

This pulls published images for:

- adapter
- shardmap
- sharding
- orchestrator
- gateway
- frontend

Use this when you want to test a release manifest without building from source.

Prepare the image source:

```powershell
$env:DOCKERHUB_NAMESPACE = "nebula67"
$env:OMNISHARD_TAG = "<release-tag-or-commit-sha>"
```

Run it:

```powershell
docker compose -f deploy/compose/full-microservices.yml up -d
```

Stop it:

```powershell
docker compose -f deploy/compose/full-microservices.yml down
```

### 3. Repo-local all-in-one release reference

File:

- `deploy/compose/single-image-microservices.yml`

This pulls one `omnishard-all-in-one` image that internally runs frontend, gateway, adapter, shardmap, sharding, and orchestrator as separate processes.

Run it:

```powershell
docker compose -f deploy/compose/single-image-microservices.yml up -d
```

Stop it:

```powershell
docker compose -f deploy/compose/single-image-microservices.yml down
```

Single-image public ports:

- Frontend UI: `http://localhost:3000`
- Adapter/API surface: `http://localhost:8080`

### Official OSS deployment

End users should normally use the official GitHub Release assets instead of the repo-local compose files.

Download the latest official full-microservices release asset:

```powershell
wget -O docker-compose.yml https://github.com/Vindyang/cs464-project/releases/latest/download/docker-compose.full-microservices.yml
```

Download the latest official single-image release asset:

```powershell
wget -O docker-compose.yml https://github.com/Vindyang/cs464-project/releases/latest/download/docker-compose.single-image-microservices.yml
```

Start the selected release:

```powershell
docker compose up -d
```

Stop it:

```powershell
docker compose down
```

Official release assets already pin the `nebula67` Docker Hub namespace and a semver image tag.

### Release packaging assets

Release-support files live under `deploy/`:

- `deploy/compose/` — pull-only repo-local release manifests
- `deploy/images/all-in-one/` — all-in-one image runtime assets such as `Dockerfile`, `entrypoint.sh`, `nginx.conf`, and `supervisord.conf`
- `deploy/release-assets/` — templates and rendering scripts for the official GitHub Release compose assets

## Testing

### Backend tests

From `backend/`:

```powershell
go test ./...
```

### Frontend checks

From `frontend/`:

```powershell
bun run lint
bun run build
```

### PowerShell test runner

The backend includes a grouped runner at `backend/tests/run-tests.ps1`.

Examples:

```powershell
./tests/run-tests.ps1 -Type unit
./tests/run-tests.ps1 -Type integration
./tests/run-tests.ps1 -Type e2e
./tests/run-tests.ps1 -Type all
```

### Test layers

Unit tests:

- live under `backend/services/*/tests/unit/`
- target isolated service logic

Integration tests:

- live under `backend/tests/integration/`
- validate contracts between orchestrator and the supporting services using controlled test doubles

E2E tests:

- live under `backend/tests/e2e/`
- exercise the real Dockerized stack and end-to-end workflows

### Recommended smoke tests after deployment

1. Open `http://localhost:3000`.
2. Configure at least one provider credential.
3. Connect a provider.
4. Upload a file.
5. Download the same file.
6. Check provider usage, file health, and history views.

## CI/CD

### Main CI pipeline

Workflow:

- `.github/workflows/ci-main.yml`

What it does:

- triggers on pushes and pull requests affecting `backend/**`, `frontend/**`, `deploy/**`, or workflow files
- resolves whether image deployment should be enabled
- runs change detection to avoid unnecessary service jobs
- runs backend quality gates with `go vet ./...` and `go build ./...`
- runs service-scoped unit tests
- runs integration contract tests
- runs E2E backend tests
- runs frontend lint and build
- publishes changed images on pushes to `main`

Branch behavior:

- `main` push: CI plus image publishing
- `microservices` push: CI only
- pull request: CI only

### Official release workflow

Workflow:

- `.github/workflows/release-github-oss.yml`

What it does:

- accepts an exact `commit_sha`
- accepts a semver `release_tag` such as `v1.2.3`
- validates the tag and commit
- builds and pushes release-tagged images for all services
- renders official release compose assets
- creates a GitHub Release and uploads the generated compose files

### Manual image republish workflow

Workflow:

- `.github/workflows/cd-dockerhub-force-deploy.yml`

What it does:

- manually republishes the full microservices image set and/or the all-in-one image
- tags them with `latest` and the full commit SHA

### Published image names

- `nebula67/omnishard-adapter`
- `nebula67/omnishard-shardmap`
- `nebula67/omnishard-sharding`
- `nebula67/omnishard-orchestrator`
- `nebula67/omnishard-gateway`
- `nebula67/omnishard-frontend`
- `nebula67/omnishard-all-in-one`

### Required GitHub secrets and variables

Required repository secrets:

- `DOCKERHUB_USERNAME`
- `DOCKERHUB_TOKEN`

Required repository variable:

- `DOCKERHUB_NAMESPACE`

## Useful Commands

Inspect running containers:

```powershell
docker compose ps
```

Tail key service logs:

```powershell
docker compose logs -f frontend
docker compose logs -f gateway
docker compose logs -f adapter
docker compose logs -f orchestrator
```

Rebuild and restart the full stack:

```powershell
docker compose --profile full up -d --build
```

Remove orphaned containers if a previous topology changed:

```powershell
docker compose --profile full down --remove-orphans
```

## Notes for Contributors

- The frontend is the user-facing control plane, but it does not own system persistence.
- The adapter and shardmap services each own their own SQLite database and should remain the sole writers for their domains.
- The orchestrator should remain stateless and workflow-focused.
- Gateway remains the public API boundary and should preserve versioned route compatibility.
- Release deployment assets should continue to live under `deploy/` and should not replace the source-build root `docker-compose.yml`.
