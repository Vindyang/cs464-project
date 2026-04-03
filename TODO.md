# TODO - Local-First Migration (Service-by-Service)

## Cross-Cutting / Project Root
- [ ] Update compose flow to be one-command local startup for full app stack.
- [ ] Remove dependency on remote Supabase/PostgreSQL credentials in local mode.
- [ ] Ensure local persistent volumes are used for stateful data (tokens, shard map DB).
- [ ] Add container health checks and service_healthy startup ordering.
- [ ] Update root documentation to describe setup flow: docker compose -> setup screen -> connect Google Drive.

## Frontend Service
- [ ] Add first-run setup route/page to collect Google OAuth Client ID and Client Secret.
- [ ] Add setup-complete gate (middleware or app-level guard) so dashboard/providers are blocked until setup finishes.
- [ ] Add API client method(s) to submit OAuth credentials to adapter runtime config endpoint.
- [ ] Update provider UI flow so after setup users proceed to existing Google connect flow.
- [ ] Align file/provider API calls with single-user local mode (no manual user_id requirement).
- [ ] Add UI messaging for setup errors (invalid credentials, backend unavailable).

## Gateway Service
- [ ] Keep gateway in local architecture as public API boundary.
- [ ] Verify gateway routing remains correct for upload/download/providers/health after backend changes.
- [ ] Add health check support in compose for deterministic frontend startup dependency.

## Adapter Service
- [x] Replace PostgreSQL token storage dependency with local persistent token store (file-based or SQLite-backed abstraction).
- [x] Remove hard failure on missing DATABASE_URL in local mode.
- [x] Add runtime endpoint(s) to configure Google OAuth credentials from frontend setup screen.
- [x] Persist OAuth credential config locally and load it on startup.
- [x] Keep optional env fallback for credentials during migration period.
- [x] Preserve OAuth authorize/callback/disconnect behavior after storage migration.
- [x] Restore Google provider connection from local persisted token on service restart.
- [x] Ensure CORS/config allows setup endpoint access from frontend origin.

## Shared OAuth Handler
- [x] Refactor credential loading order: runtime-configured credentials first, env fallback second.
- [x] Ensure callback path handles missing/invalid runtime config with clear errors.
- [x] Keep state token validation and redirect flow intact.

## Shared Google Drive Adapter
- [x] Implement find-or-create behavior for folder named nebula on first successful connection.
- [x] Persist resolved folder ID for reuse (avoid searching/creating every request).
- [x] Remove hard dependency on preconfigured GDRIVE_FOLDER_ID for local mode.
- [x] Verify shard upload naming and parent folder assignment remain correct.

## Sharding Service
- [ ] Confirm no DB dependency remains and service stays stateless/local.
- [ ] Validate sharding/reconstruct APIs unchanged to preserve orchestrator contract.
- [ ] Add compose health checks and readiness dependency wiring.

## Shardmap Service
- [ ] Migrate database layer from PostgreSQL driver/config to local SQLite file.
- [ ] Convert repository SQL patterns for SQLite compatibility (placeholders, RETURNING, timestamps if needed).
- [ ] Keep shard map API response contracts unchanged for orchestrator/frontend compatibility.
- [ ] Remove strict user_id query requirement for local single-user mode (server-side default identity).
- [ ] Ensure file/shard status updates and constraints continue to work post-migration.
- [ ] Add startup behavior to initialize/create local schema if DB file is empty.

## Orchestrator Service
- [ ] Keep workflow orchestration unchanged (upload/download contract stability).
- [ ] Validate dependencies still call adapter/shardmap/sharding URLs in compose network.
- [ ] Re-test rollback behavior on partial shard upload failures after adapter/shardmap storage changes.

## Shared Database / Schema Modules
- [x] Swap PostgreSQL dependencies for SQLite dependency where needed.
- [ ] Introduce/maintain local schema definition compatible with SQLite constraints.
- [ ] Decide whether to keep existing PostgreSQL schema as legacy reference or fully replace.
- [ ] Update DB initialization path to avoid Postgres-only assumptions.

## Docker Compose / Infra
- [ ] Remove postgres service and related init-local.sql mount for local-first mode.
- [ ] Add persistent volumes for adapter local config/token data and shardmap SQLite DB.
- [ ] Add/adjust env vars for local storage paths and runtime config.
- [ ] Add health checks for adapter, shardmap, sharding, orchestrator, gateway, frontend.
- [ ] Change depends_on to service_healthy where appropriate.
- [ ] Keep full profile startup path simple and deterministic.

## Documentation
- [ ] Document user journey: start stack, complete setup screen, connect Google Drive, upload/download files.
- [ ] Document required Google Cloud OAuth redirect URI for local callback.
- [ ] Document nebula auto-folder behavior.
- [ ] Document persistence behavior across restarts (where local data is stored).
- [ ] Remove outdated references to Supabase/remote DB as required for local mode.

## Verification / QA
- [ ] Smoke test full local startup via compose with no manual backend credential files required beyond setup UI.
- [ ] Validate setup -> connect -> upload -> list -> download happy path.
- [ ] Validate restart persistence for OAuth tokens, OAuth config, and shard metadata.
- [ ] Validate failure cases: bad OAuth credentials, revoked token, upload partial failure rollback.
- [ ] Run backend tests and targeted integration contracts after migration.