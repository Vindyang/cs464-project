# Quick Start

This guide helps contributors and operators run Omnishard with Docker and set up provider credentials through the frontend UI.

---

## Prerequisites

- 64-bit system (macOS/Linux recommended)
- **Minimum:** 2 CPU cores, 6 GB RAM  
  **Recommended:** 4 CPU cores, 8 GB RAM
- Docker Desktop (macOS/Windows) or Docker Engine (Linux)
- Docker Compose plugin (`docker compose` command)
- Internet access for cloud provider OAuth/API calls
- Open local ports: `3000`, `8080`, `8081`, `8082`, `8083`, `8084`

Verify Docker:

```bash
docker --version
docker compose version
```

---

## Quick Start (Docker)

From repo root (`cs464-project/`):

```bash
docker compose --profile full up -d --build
```

Endpoints:

- Frontend: `http://localhost:3000`
- Adapter: `http://localhost:8080`
- Shardmap: `http://localhost:8081`
- Orchestrator: `http://localhost:8082`
- Sharding: `http://localhost:8083`
- Gateway: `http://localhost:8084`

Stop stack:

```bash
docker compose --profile full down
```

Reset all persisted Docker volumes (clean slate, deletes local DB data):

```bash
docker compose --profile full down -v
```

---

## Credential Configuration

Provider credentials/tokens are configured directly in the frontend:

1. Open `http://localhost:3000`
2. Go to **Credentials**
3. Choose provider in **Add or Update**
4. Fill fields shown in the form
5. Click **Save Credentials**
6. Go to **Providers** page and connect/authorize provider if required

---

## Cloud Provider Secrets/Tokens: How to Obtain

### AWS S3

**Official setup links**

- IAM Console: https://console.aws.amazon.com/iam/
- S3 Console: https://s3.console.aws.amazon.com/s3/home

#### Fields in UI

- **Access Key ID**
- **Secret Access Key**
- **Region**

#### How to obtain

1. Open IAM Console.
2. Create/select an IAM user (or role for programmatic access).
3. Grant least-privilege bucket/object permissions:
   - `s3:ListBucket`
   - `s3:GetObject`
   - `s3:PutObject`
   - `s3:DeleteObject` (if delete is enabled)
4. Create access keys.
5. Copy Access Key ID + Secret Access Key.
6. Use your bucket region in **Region** (example: `ap-southeast-1`).

---

### Google Drive OAuth

**Official setup links**

- Google Cloud Console: https://console.cloud.google.com/
- OAuth Credentials page: https://console.cloud.google.com/apis/credentials

#### Fields in UI

- **Client ID**
- **Client Secret**
- **Redirect URI** (must match exactly):  
  `http://localhost:8080/api/oauth/gdrive/callback`

#### How to obtain

1. Create/select a project in Google Cloud.
2. Enable **Google Drive API**.
3. Configure **OAuth consent screen**.
4. Create **OAuth Client ID** (Web application).
5. Add redirect URI exactly:  
   `http://localhost:8080/api/oauth/gdrive/callback`
6. Copy Client ID + Client Secret into Omnishard Credentials UI.
7. Connect provider from **Providers** page.

---

### Microsoft OneDrive OAuth

**Official setup links**

- Azure Portal: https://portal.azure.com/
- App registrations: https://portal.azure.com/#view/Microsoft_AAD_RegisteredApps/ApplicationsListBlade

#### Fields in UI

- **Client ID**
- **Client Secret**
- **Redirect URI** (must match exactly):  
  `http://localhost:8080/api/oauth/onedrive/callback`

#### How to obtain

1. Open App registrations and create/select an app.
2. Add redirect URI exactly:  
   `http://localhost:8080/api/oauth/onedrive/callback`
3. Create a client secret in **Certificates & secrets**.
4. Copy:
   - Application (client) ID
   - Client secret **Value** (not Secret ID)
5. Save in Omnishard Credentials UI.
6. Connect provider from **Providers** page.

---

## Troubleshooting

Check container status:

```bash
docker compose ps
```

Tail logs:

```bash
docker compose logs -f frontend
docker compose logs -f gateway
docker compose logs -f adapter
docker compose logs -f orchestrator
docker compose logs -f shardmap
docker compose logs -f sharding
```