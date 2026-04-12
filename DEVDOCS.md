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

**Official setup links**

- IAM Console: https://console.aws.amazon.com/iam/
- S3 Console: https://s3.console.aws.amazon.com/s3/home

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

Enter the values in the credentials page:
- **Access Key ID:** Your Access Key ID from Step 3
- **Secret Access Key:** Your Secret Access Key from Step 3
- **Region:** Your bucket's region (example: `ap-southeast-1`)

#### Step 5: Connect Provider

1. Navigate to the **Providers** page
2. Find **AWS S3** and click **Connect**
3. You're all set!

#### Step 6: Enable Console Access (Optional)

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

#### Step 1: Create or Select a Project

1. Open [Google Cloud Console](https://console.cloud.google.com/)
2. Click the **Project selector** (top left)
3. Click **New Project**
4. Enter a project name (e.g., `omnishard-gdrive`)
5. Click **Create**
6. Wait for the project to be created, then select it

#### Step 2: Enable Google Drive API

1. Go to your project → Click **APIs & Services** (left sidebar)
2. Click **Enabled APIs & services**
3. Click **Enable APIs and Services** (top button)
4. Search for **Google Drive API**
5. Click on it and click **Enable**

#### Step 3: Configure OAuth Consent Screen

1. Go to **APIs & Services** → **OAuth consent screen** (left sidebar)
2. Select **External** (or **Internal** if within an organization)
3. Click **Create**
4. Fill in the required fields:
   - **App name:** `omnishard-gdrive`
   - **User support email:** Your email
   - **Developer contact information:** Your email
5. Click **Save and Continue**
6. Click **Save and Continue** → **Back to Dashboard**

#### Step 4: Create OAuth Client ID

1. Go to **APIs & Services** → **Credentials** (left sidebar)
2. Click **Create Credentials** (top button)
3. Select **OAuth Client ID**
4. Choose **Web application**
5. Enter a name (e.g., `omnishard-gdrive-client`)
6. Under **Authorized redirect URIs**, click **Add URI**
7. Enter the redirect URI:
http://localhost:8080/api/oauth/gdrive/callback
8. Click **Create**

#### Step 5: Copy Your Credentials

A popup will appear with your credentials. **⚠️ Copy both immediately:**
- **Client ID**
- **Client Secret**

#### Step 6: Fill in Your Credentials

Enter the values in the credentials page:
- **Client ID:** Your Client ID from Step 5
- **Client Secret:** Your Client Secret from Step 5
- **Redirect URI:** `http://localhost:8080/api/oauth/gdrive/callback`

#### Step 7: Connect Provider

1. Navigate to the **Providers** page
2. Find **Google Drive** and click **Connect**
3. Authorize the application when prompted
4. You're all set!

---

### Microsoft OneDrive OAuth

**Official setup links**

- Azure Portal: https://portal.azure.com/
- App registrations: https://portal.azure.com/#view/Microsoft_AAD_RegisteredApps/ApplicationsListBlade

#### Step 1: Register Application in Azure

1. Open [App registrations](https://portal.azure.com/#view/Microsoft_AAD_RegisteredApps/ApplicationsListBlade)
2. Click **New registration**
3. Enter an application name (e.g., `omnishard-onedrive`)
4. Click **Register**

#### Step 2: Configure Redirect URI

1. Go to your app registration → **Authentication** (left sidebar)
2. Click **Add a platform** → **Web**
3. Enter the redirect URI:
   `http://localhost:8080/api/oauth/onedrive/callback`
4. Check **Access tokens** and **ID tokens**
5. Click **Configure**

#### Step 3: Create Client Secret

1. Go to your app registration → **Certificates & secrets** (left sidebar)
2. Click the **Client secrets** tab
3. Click **New client secret**
4. Click **Add**
5. **Copy the secret value immediately** from the **Value** column (the long string)

#### Step 4: Gather Credentials

1. Go to your app registration → **Overview** (left sidebar)
2. Copy the **Application (client) ID**
3. From Step 3, you already have the **Client Secret**
4. Your **Redirect URI** is: `http://localhost:8080/api/oauth/onedrive/callback`

Enter these values in the credentials page:
- **Client ID:** Your Application (client) ID
- **Client Secret:** Your Secret Access Key from Step 3
- **Redirect URI:** `http://localhost:8080/api/oauth/onedrive/callback`

#### Step 5: Connect Provider

1. Navigate to the **Providers** page
2. Find **OneDrive** and click **Connect**
3. Authorize the application when prompted
4. You're all set!

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