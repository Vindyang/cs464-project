# CS464 Project

## Where to run Docker Compose commands

Run all Docker Compose commands from the project root:
`cs464-project/` (this directory contains `docker-compose.yml`).

## Release flavors

The repository now supports three distinct Docker Compose use cases:

- Source-build developer workflow: `docker-compose.yml`
- Repo-local release flavor 1 reference: `deploy/compose/full-microservices.yml`
- Repo-local release flavor 2 reference: `deploy/compose/single-image-microservices.yml`

The root `docker-compose.yml` remains the contributor/developer compose and builds images from local source.
The repo-local release compose files are useful for maintainers and local release testing, but end users should prefer the official GitHub Release assets.

### Official OSS deployment

Download the latest official full-microservices release asset:

```powershell
wget -O docker-compose.yml https://github.com/Vindyang/cs464-project/releases/latest/download/docker-compose.full-microservices.yml
```

Download the latest official single-image release asset:

```powershell
wget -O docker-compose.yml https://github.com/Vindyang/cs464-project/releases/latest/download/docker-compose.single-image-microservices.yml
```

Then start the selected release:

```powershell
docker compose up -d
```

Stop it with:

```powershell
docker compose down
```

Official GitHub Release assets already point at the `nebula67` Docker Hub namespace and pin the release's semver image tag.

### Repo-local release testing

If you are maintaining the project and want to test release manifests from the repository checkout, keep using the repo-local files under `deploy/compose/`.

```powershell
$env:DOCKERHUB_NAMESPACE = "nebula67"
$env:OMNISHARD_TAG = "<release-tag-or-commit-sha>"
```

### Release flavor 1: full microservices

Runs frontend plus each backend service in its own container using published images.

```powershell
docker compose -f deploy/compose/full-microservices.yml up -d
```

Stop it with:

```powershell
docker compose -f deploy/compose/full-microservices.yml down
```

### Release flavor 2: single-image microservices

Runs frontend, gateway, and all backend services inside one container while preserving the current service boundaries internally.

```powershell
docker compose -f deploy/compose/single-image-microservices.yml up -d
```

Stop it with:

```powershell
docker compose -f deploy/compose/single-image-microservices.yml down
```

Public ports for the single-image release:

- Frontend UI: `http://localhost:3000`
- Adapter/API surface used by the current frontend: `http://localhost:8080`

## Backend only (profile: backend)

### Build backend images

Build all backend services:

```powershell
docker compose build adapter shardmap sharding orchestrator gateway
```

Build a single backend service:

```powershell
docker compose build adapter
docker compose build shardmap
docker compose build sharding
docker compose build orchestrator
docker compose build gateway
```

Or build everything assigned to backend profile:

```powershell
docker compose --profile backend build
```

### Start backend

Start all backend services:

```powershell
docker compose --profile backend up -d
```

Start one backend service (with dependencies):

```powershell
docker compose --profile backend up -d adapter
docker compose --profile backend up -d shardmap
docker compose --profile backend up -d sharding
docker compose --profile backend up -d orchestrator
docker compose --profile backend up -d gateway
```

Rebuild and start one service:

```powershell
docker compose --profile backend up -d --build adapter
```

### Stop backend

Stop one service:

```powershell
docker compose stop adapter
```

Stop/remove all backend profile containers and network:

```powershell
docker compose --profile backend down
```

Stop/remove backend profile and volumes (wipes local service data):

```powershell
docker compose --profile backend down -v
```

## Fullstack (profile: full)

The `full` profile includes backend services plus frontend.

### Build fullstack

```powershell
docker compose --profile full build
```

### Start fullstack

```powershell
docker compose --profile full up -d
```

### Stop fullstack

```powershell
docker compose --profile full down
```

### Rebuild and restart fullstack

```powershell
docker compose --profile full up -d --build
```

## Useful commands

```powershell
docker compose ps
docker compose logs -f adapter
docker compose logs -f gateway
docker compose logs -f frontend
```

## Docker Hub Continuous Deployment

Continuous image publishing from `main` is handled by:

- `.github/workflows/ci-main.yml`

Formal GitHub Releases are created manually by:

- `.github/workflows/release-github-oss.yml`

Manual image-only republishing remains available through:

- `.github/workflows/cd-dockerhub-force-deploy.yml`

Published image repositories:

- `nebula67/omnishard-adapter`
- `nebula67/omnishard-shardmap`
- `nebula67/omnishard-sharding`
- `nebula67/omnishard-orchestrator`
- `nebula67/omnishard-gateway`
- `nebula67/omnishard-frontend`
- `nebula67/omnishard-all-in-one`

### Required GitHub configuration

Set the following repository configuration in GitHub:

- Repository secret: `DOCKERHUB_USERNAME`
- Repository secret: `DOCKERHUB_TOKEN` (Docker Hub access token, not password)

### Publish behavior

- `ci-main.yml` publishes changed images on pushes to `main` using `latest` plus full commit SHA tags.
- `microservices` runs CI only and does not publish images.
- `release-github-oss.yml` is the official release path. It takes an exact commit SHA plus a semver tag, pushes release-tagged images, creates the GitHub Release, and uploads:
	- `docker-compose.full-microservices.yml`
	- `docker-compose.single-image-microservices.yml`

## Testing release builds

### Create an official release

1. Open GitHub Actions and run `.github/workflows/release-github-oss.yml`.
2. Enter the exact `commit_sha` you want to package.
3. Enter the semver `release_tag` you want to publish, such as `v1.2.3`.
4. Wait for the workflow to push the tagged images to Docker Hub and publish the GitHub Release assets.

### Force-publish test images

Use the manual workflow when you want to build and push images without waiting for a code-change trigger.

1. Open GitHub Actions and run `.github/workflows/cd-dockerhub-force-deploy.yml`.
2. Check `deploy_full_microservices` to publish flavor 1 images.
3. Check `deploy_all_in_one` to publish flavor 2.
4. Check both if you want to rebuild and push both release flavors in one run.
5. Wait for the workflow to complete and note the commit SHA for the run.

### Test flavor 1 locally

1. Set the namespace to the official Docker Hub publisher:

```powershell
$env:DOCKERHUB_NAMESPACE = "nebula67"
```

2. Set the exact image tag you want to test. Use either a release tag like `v1.2.3` or the full commit SHA pushed by the `main` publish path:

```powershell
$env:OMNISHARD_TAG = "<full-commit-sha>"
```

3. Pull and start the release stack:

```powershell
docker compose -f deploy/compose/full-microservices.yml pull
docker compose -f deploy/compose/full-microservices.yml up -d
```

4. Verify containers are running:

```powershell
docker compose -f deploy/compose/full-microservices.yml ps
```

5. Smoke-test the release:
	- Open `http://localhost:3000`
	- Visit `http://localhost:8084/api/v1/health`
	- Verify credentials and providers load in the UI
	- Upload a file and download it back

6. Check logs if needed:

```powershell
docker compose -f deploy/compose/full-microservices.yml logs -f frontend
docker compose -f deploy/compose/full-microservices.yml logs -f gateway
docker compose -f deploy/compose/full-microservices.yml logs -f orchestrator
```

7. Tear it down when finished:

```powershell
docker compose -f deploy/compose/full-microservices.yml down
```

### Test flavor 2 locally

1. Reuse the same `DOCKERHUB_NAMESPACE` and set `OMNISHARD_TAG` to a release tag like `v1.2.3` or the target full commit SHA.

2. Pull and start the all-in-one release:

```powershell
docker compose -f deploy/compose/single-image-microservices.yml pull
docker compose -f deploy/compose/single-image-microservices.yml up -d
```

3. Verify the container is running:

```powershell
docker compose -f deploy/compose/single-image-microservices.yml ps
```

4. Smoke-test the release:
	- Open `http://localhost:3000`
	- Confirm adapter-facing endpoints respond on `http://localhost:8080`
	- Upload a file and download it back through the UI
	- Restart the container once and confirm state persists

5. Inspect logs if needed:

```powershell
docker compose -f deploy/compose/single-image-microservices.yml logs -f
```

6. Tear it down when finished:

```powershell
docker compose -f deploy/compose/single-image-microservices.yml down
```
