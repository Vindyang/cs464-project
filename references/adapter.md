
# Plan: Multi-Cloud Adapter with Go Backend

Build a provider-agnostic integration layer in a separate **Go API service**, acting as an **Authenticated Gateway**. Next.js handles client-side encryption and sharding (Hybrid Model), while the Go service manages OAuth token lifecycles, provider-specific API connections, and parallelized shard persistence for AWS S3, Google Drive, and OneDrive. This ensures end-to-end privacy (keys stay in-browser) while offloading complex vendor SDK management and background concurrency logic to the backend.

## Steps

1. **Define a canonical provider contract** (ids, status, capacity, capabilities, auth state) and normalize naming across current UI/store usage in `providers.ts`, `providerStore.ts`, and `ProviderMatrix.tsx`.
2. **Design the Go adapter architecture:** core provider interface, per-provider adapters (AWS S3, Google Drive, OneDrive), adapter registry, and a compatibility path for existing providers.
3. **Define Go API surface:** (providers list/connect/disconnect/health, file list/upload/download/delete, shard metadata) and response schemas that Next.js can consume without provider-specific branching.
    - **Atomic Upload Protocol:** If the required quorum of shards (e.g., 4 out of 6) fails to upload, the API must trigger an automated rollback to delete partial shards from successful providers.
4. **Implement auth/security design:** OAuth flows for Google/Microsoft (PKCE), credential strategy for AWS, encrypted token storage, and background refresh handling.
    - **Encryption Boundary:** All encryption (AES-256-GCM) and erasure coding (sharding) occurs in the browser. The Go service receives opaque, encrypted blobs (shards) and routes them to providers.
5. **Latency & Health System:** Implement a periodic background heartbeat in Go that monitors provider API latency and availability, updating the `provider_connections` table for the UI.
6. **Add a phased migration path in Next.js:** API-first store reads with mock fallback, then move onboarding/provider actions from local state to backend mutations in `app/(auth)/onboarding/page.tsx` and `app/(authenticated)/providers/page.tsx`.
7. **Unify provider identity in file/shard UX** so canonical provider ids replace display-name coupling in `uploadStore.ts` and `app/(authenticated)/files/[id]/page.tsx`.
8. **Finalize rollout and observability plan:** feature flags, error budgets, provider health checks, fallback behavior, and cutover criteria.
    - **Cleanup Worker:** Develop a background routine to identify and prune 'orphaned' shards—data segments in cloud storage that aren't linked to a completed file record in the database.
9. **Data & Reliability Contract:** Define canonical storage entities, API payloads, and constraints for adapter correctness.
    - **`providerConnections`**
        - `id` (uuid), `userId`, `workspaceId`
        - `providerId` (`awsS3` | `googleDrive` | `oneDrive` | legacy providers)
        - `providerAccountId` (remote account/tenant id), `displayName`
        - `status` (`connected` | `degraded` | `disconnected` | `error`)
        - `latencyMs` (integer, from last heartbeat)
        - `capabilitiesJson` (e.g., `{ "maxPartSize": 5242880, "supportsVersioning": true }`)
        - `authType` (`oauth2` | `accessKey`), `scopes`
        - `accessTokenEncrypted`, `refreshTokenEncrypted`, `tokenExpiresAt`
        - `credentialsRef` (secret manager key ref for S3 creds), `regionDefault`, `bucketDefault`
        - `quotaTotalBytes`, `quotaUsedBytes`, `quotaLastSyncedAt`
        - `lastHealthCheckAt`, `lastErrorCode`, `lastErrorMessage`
        - `createdAt`, `updatedAt`, `deletedAt`
    - **`files`**
        - `id` (uuid), `userId`, `workspaceId`
        - `logicalPath`, `filename`, `mimeType`, `sizeBytes`
        - `checksumSha256` (whole-file hash), `encryptionAlg`, `encryptionKeyRef`, `iv`
        - `quorumRequired`, `shardCount`, `version`
        - `status` (`pending` | `inProgress` | `committed` | `failed` | `rolledBack`)
        - `commitAt`, `failedAt`, `failureReason`
        - `createdAt`, `updatedAt`
    - **`shards`**
        - `id` (uuid), `fileId`, `index`, `sizeBytes`
        - `checksumSha256` (per-shard hash), `erasureGroup`, `parity` (bool)
        - `providerConnectionId`, `remoteObjectId`, `remotePath`
        - `region`, `bucketOrDriveId`
        - `status` (`pending` | `uploading` | `uploaded` | `verified` | `deletePending` | `deleted` | `failed`)
        - `attemptCount`, `lastAttemptAt`, `uploadedAt`, `verifiedAt`
        - `errorCode`, `errorMessage`
        - `createdAt`, `updatedAt`
    - **`uploadSessions`**
        - `id` (uuid), `fileId`, `idempotencyKey` (unique)
        - `clientRequestId`, `traceId`
        - `status`, `startedAt`, `expiresAt`, `completedAt`
        - `rollbackRequired` (bool), `rollbackCompletedAt`
        - `failureReason`, `metadataJson`
    - **`operationEvents` (audit/observability)**
        - `id`, `entityType`, `entityId`, `eventType`
        - `providerId`, `requestId`, `traceId`
        - `payloadJson` (sanitized), `createdAt`
    - **API payload fields (minimum)**
        - **Provider DTO:** `providerId`, `displayName`, `status`, `latencyMs`, `region`, `capabilities`, `quotaTotalBytes`, `quotaUsedBytes`, `lastHealthCheckAt`
        - **Create Upload:** `filename`, `sizeBytes`, `mimeType`, `checksumSha256`, `quorumRequired`, `idempotencyKey`
        - **Shard Report:** `index`, `providerId`, `remoteObjectId`, `checksumSha256`, `sizeBytes`, `status`
        - **Finalize Upload:** `uploadSessionId`, `expectedShardCount`, `commitRequestedAt`
    - **Constraints to add in the plan**
        - Unique: (`workspaceId`, `providerId`, `providerAccountId`) on `providerConnections`
        - Unique: `idempotencyKey` on `uploadSessions`
        - Unique: (`fileId`, `index`) on `shards`
        - Indexes: `status`, `providerConnectionId`, `fileId`, `updatedAt` for cleanup/health jobs
    - **Reliability Enhancements**
        - **Rate Limiting:** Implement token-bucket limiting per provider to avoid 429 errors from Google/Microsoft APIs.
        - **Backoff Strategy:** Jittered exponential backoff for all provider-bound network requests.
        - **Concurrency Control:** Atomic file status transitions using row-level locking or optimistic concurrency via `version` column.

## Verification

- Contract validation between Next.js and Go APIs (schema compatibility and error shapes)
- End-to-end provider flows: connect, list files, upload/download, disconnect
- UI parity checks for providers/files screens before removing mock fallback

## Decisions

- **Backend model:** separate Go API service with Next.js frontend/BFF
- **Integration strategy:** direct provider integrations, no Apideck dependency
- **Scope:** add AWS S3, Google Drive, OneDrive while keeping existing providers available during transition
