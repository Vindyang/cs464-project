# Nebula Drive — Adapter Service

The backend adapter service exposes a unified JSON API over multiple cloud storage providers. The frontend never talks to cloud SDKs directly — it always goes through this service.

---

## Prerequisites

- Go 1.23+
- A Google account (personal or dedicated project account)
- A Google Cloud project with the **Google Drive API** enabled

---

## Google Cloud Setup (one-time)

### 1. Enable the Drive API

GCP Console → **APIs & Services** → **Enable APIs and Services** → search **Google Drive API** → Enable.

### 2. Configure the OAuth consent screen

GCP Console → **APIs & Services** → **OAuth consent screen**:
- User type: **External**
- Fill in app name (e.g. `Nebula Drive`), support email, developer email
- Scopes: skip (handled in code)
- Test users: **add your Google account email**

> The app does not need to be verified by Google. Testing mode with added test users is sufficient.

### 3. Create OAuth 2.0 credentials

GCP Console → **APIs & Services** → **Credentials** → **Create Credentials** → **OAuth 2.0 Client ID**:
- Application type: **Desktop app**
- Download the JSON file → save it **outside the repo** (e.g. `C:/secrets/gdrive-oauth-credentials.json`)

### 4. Create a Drive folder

In Google Drive, create a folder for storing shards (e.g. `nebula-shards`). Copy the folder ID from its URL:

```
https://drive.google.com/drive/folders/<FOLDER_ID>
```

---

## Environment Setup

Copy the example file and fill in your values:

```bash
cp .env.example .env
```

| Variable | Description |
|---|---|
| `GDRIVE_OAUTH_CREDENTIALS_FILE` | Path to the OAuth2 credentials JSON downloaded from GCP Console |
| `GDRIVE_TOKEN_FILE` | Path where the OAuth token will be saved after running `gdrive-auth` |
| `GDRIVE_FOLDER_ID` | ID of the Google Drive folder to store shards in (from the folder URL) |

> `.env` is gitignored. Never commit it. Share credentials with teammates out-of-band (e.g. a private group chat).

---

## First-Time Authorization

Run this once per machine to authorize your Google account and save a token:

```bash
go run ./cmd/gdrive-auth/main.go
```

It will:
1. Print an authorization URL
2. You open the URL in your browser and grant access
3. Paste the authorization code back into the terminal
4. Save the token to the path in `GDRIVE_TOKEN_FILE`

The token auto-refreshes on subsequent runs — you only need to do this once.

---

## Running the Server

```bash
go run ./cmd/server/main.go
```

Server starts on `:8080`. Check it's working:

```bash
curl http://localhost:8080/api/providers
```

Example response:

```json
[
  {
    "providerId": "googleDrive",
    "displayName": "Google Drive",
    "status": "connected",
    "latencyMs": 142,
    "region": "global",
    "capabilities": { "supportsVersioning": true },
    "quotaTotal": 16106127360,
    "quotaUsed": 1234567,
    "lastCheck": "2026-03-16T10:00:00Z"
  }
]
```

---

## Running Integration Tests

The integration tests hit the real Drive API — they upload a test shard, download it back, verify the content, then delete it.

```bash
go test ./internal/adapter/gdrive/... -v -run TestGDriveIntegration
```

Tests skipped automatically if env vars are not set, so `go test ./...` is always safe to run.

Sub-tests:

| Test | What it checks |
|---|---|
| `HealthCheck` | API connection is live |
| `GetMetadata` | Quota numbers and latency are returned |
| `UploadDownloadDelete` | Full shard round-trip against real Drive |

---

## API Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/providers` | Returns metadata (status, quota, latency) for all registered providers |

Upload/download/delete endpoints are pending — currently exercised via integration tests only.

---

## Architecture

```
cmd/
  server/        ← HTTP server entry point
  gdrive-auth/   ← One-time OAuth2 authorization tool

internal/adapter/
  adapter.go     ← StorageProvider interface + Registry
  gdrive/        ← Google Drive implementation
  s3/            ← AWS S3 implementation (stubbed)
```

All providers implement the `StorageProvider` interface (`GetMetadata`, `UploadShard`, `DownloadShard`, `DeleteShard`, `HealthCheck`). New providers are added by implementing the interface and registering in `cmd/server/main.go`.
