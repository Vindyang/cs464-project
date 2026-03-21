# Nebula Drive — Adapter Service

The backend adapter service exposes a unified JSON API over multiple cloud storage providers. The frontend never talks to cloud SDKs directly — it always goes through this service.

---

## Prerequisites

- Go 1.23+
- A Google account (personal or dedicated project account)
- A Google Cloud project with the **Google Drive API** enabled
- A [Supabase](https://supabase.com) project

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
- Application type: **Web application**
- Authorized JavaScript origins: `http://localhost:3000` and `http://localhost:8080`
- Authorized redirect URIs: `http://localhost:8080/api/oauth/gdrive/callback`
- Download the JSON file → save it **outside the repo** (e.g. `C:/secrets/gdrive-oauth-credentials.json`)

### 4. Create a Drive folder

In Google Drive, create a folder for storing shards (e.g. `nebula-shards`). Copy the folder ID from its URL:

```
https://drive.google.com/drive/folders/<FOLDER_ID>
```

---

## Supabase Setup (one-time)

### 1. Create the `provider_connections` table

In the Supabase SQL editor, run:

```sql
create table provider_connections (
  provider_id   text primary key,
  access_token  text not null,
  refresh_token text,
  token_type    text,
  expiry        timestamptz,
  updated_at    timestamptz default now()
);
```

### 2. Get the connection string

Supabase dashboard → **Project Settings** → **Database** → **Connection string** → **URI**. Use this as `DATABASE_URL`.

---

## Environment Setup

Copy the example file and fill in your values:

```bash
cp .env.example .env
```

| Variable | Description |
|---|---|
| `DATABASE_URL` | PostgreSQL connection string from Supabase |
| `GDRIVE_OAUTH_CREDENTIALS_FILE` | Path to the OAuth2 credentials JSON downloaded from GCP Console |
| `GDRIVE_OAUTH_REDIRECT_URI` | Must match the redirect URI configured in GCP (`http://localhost:8080/api/oauth/gdrive/callback`) |
| `GDRIVE_FOLDER_ID` | ID of the Google Drive folder to store shards in |
| `FRONTEND_URL` | Frontend origin for post-OAuth redirect (`http://localhost:3000`) |

> `.env` is gitignored. Never commit it. Share credentials with teammates out-of-band (e.g. a private group chat).

---

## Running the Server

```bash
go run ./cmd/server/main.go
```

Server starts on `:8080`. Check it's working:

```bash
curl http://localhost:8080/api/providers
```

---

## Connecting Google Drive

Authorization is done through the browser UI:

1. Start the backend and frontend
2. Go to `http://localhost:3000/providers`
3. Click **Connect New** → **Connect** next to Google Drive
4. Complete the Google consent screen
5. You'll be redirected back to `/providers` with Google Drive now connected

The OAuth token is stored in Supabase and loaded automatically on server restart — you only need to authorize once.

---

## Running Integration Tests

The integration tests hit the real Drive API — they upload a test shard, download it back, verify the content, then delete it.

```bash
go test ./internal/adapter/gdrive/... -v -run TestGDriveIntegration
```

Tests skipped automatically if env vars are not set, so `go test ./...` is always safe to run.

| Test | What it checks |
|---|---|
| `HealthCheck` | API connection is live |
| `GetMetadata` | Quota numbers and latency are returned |
| `UploadDownloadDelete` | Full shard round-trip against real Drive |

---

## API Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/providers` | Returns metadata for all registered providers |
| `GET` | `/api/oauth/gdrive/authorize` | Returns a Google OAuth authorization URL |
| `GET` | `/api/oauth/gdrive/callback` | Handles OAuth callback, stores token, registers adapter |
| `POST` | `/api/oauth/gdrive/disconnect` | Removes token from Supabase, unregisters adapter |

---

## Architecture

```
cmd/
  server/              ← HTTP server entry point

internal/
  adapter/
    adapter.go         ← StorageProvider interface + Registry
    gdrive/            ← Google Drive implementation
    s3/                ← AWS S3 implementation (stubbed)
  db/
    db.go              ← Supabase/PostgreSQL token storage
  oauthhandler/
    gdrive.go          ← OAuth endpoint handlers
```

All providers implement the `StorageProvider` interface (`GetMetadata`, `UploadShard`, `DownloadShard`, `DeleteShard`, `HealthCheck`). New providers are added by implementing the interface and registering in `cmd/server/main.go`.
