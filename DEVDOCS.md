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

## Two Ways to Run Omnishard

Before you start: there are **2 supported ways** to run the app.

1. **Clone + build from source (developer workflow)**  
   This is the default workflow in this guide (the Quick Start section below).

2. **Pull the latest published images from GitHub Packages (GHCR)**  
   This skips local image builds and runs published container images.

### Run from latest GHCR images

Download the latest full-microservices release compose file:

```bash
curl -L -o [docker-compose.yml](http://_vscodecontentref_/0) https://github.com/Vindyang/cs464-project/releases/latest/download/docker-compose.full-microservices.yml
```

Start the stack:
```
docker compose up -d
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

#### Step 1: Create an IAM User

1. Open [IAM Console](https://console.aws.amazon.com/iam/)
2. Click **Users** → **Create user**
3. Enter a username (e.g., `omnishard-app`)
4. Click **Next**

#### Step 2: Attach S3 Permissions

1. Click **Attach policies directly**
2. Click **Create policy**
3. Paste this JSON:
```json
   {
   "Version": "2012-10-17",
   "Statement": [
      {
         "Effect": "Allow",
         "Action": [
         "s3:ListAllMyBuckets"
         ],
         "Resource": "arn:aws:s3:::*"
      },
      {
         "Effect": "Allow",
         "Action": [
         "s3:ListBucket",
         "s3:GetObject",
         "s3:PutObject",
         "s3:DeleteObject"
         ],
         "Resource": [
         "arn:aws:s3:::omnishard-*",
         "arn:aws:s3:::omnishard-*/*"
         ]
      }
   ]
   }
```
4. Click **Next** → Enter a policy name (e.g., `omnishard-policy`)
5. Click **Create policy** → attach it to your user
6. Search for your policy name (`omnishard-policy`) in the policy search
7. Click **Entities attached** → **Attach**
8. Select your IAM user and confirm

#### Step 3: Create Access Keys

1. Go to **Users** → Select your user
2. Click **Security credentials** tab
3. Scroll to **Access keys** → **Create access key**
4. Select **Application running outside AWS**
5. Click **Next** → **Create access key**
6. **⚠️ Copy both keys immediately** (AWS only shows them once):
   - Access Key ID
   - Secret Access Key

#### Step 4: Fill in Your Credentials

Enter the values in your application:
- **Access Key ID:** Your Access Key ID from Step 3
- **Secret Access Key:** Your Secret Access Key from Step 3
- **Region:** Your bucket's region (example: `ap-southeast-1`)

#### Step 5: Enable Console Access (Optional)

To log in to AWS Console as this IAM user:

1. Go to **Users** → Select your user
2. Click **Security credentials** tab
3. Scroll to **Console sign-in credentials**
4. Click **Create login password**
5. Set a password and note it down
6. You can now log in at https://console.aws.amazon.com with:
   - **IAM user name:** `omnishard-app` (or your username)
   - **Password:** The password you just created
7. After login, navigate to [S3 Console](https://s3.console.aws.amazon.com/s3/home) to view your buckets

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