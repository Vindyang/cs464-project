# CS464 Project

## Where to run Docker Compose commands

Run all Docker Compose commands from the project root:
`cs464-project/` (this directory contains `docker-compose.yml`).

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

Each service is published to its own Docker Hub repository using the GitHub Actions workflow:

- `.github/workflows/cd-dockerhub-services.yml`

Published image repositories:

- `${DOCKERHUB_NAMESPACE}/omnishard-adapter`
- `${DOCKERHUB_NAMESPACE}/omnishard-shardmap`
- `${DOCKERHUB_NAMESPACE}/omnishard-sharding`
- `${DOCKERHUB_NAMESPACE}/omnishard-orchestrator`
- `${DOCKERHUB_NAMESPACE}/omnishard-gateway`

### Required GitHub configuration

Set the following repository configuration in GitHub:

- Repository variable: `DOCKERHUB_NAMESPACE` (example: `myorg`)
- Repository secret: `DOCKERHUB_USERNAME`
- Repository secret: `DOCKERHUB_TOKEN` (Docker Hub access token, not password)

### Publish behavior

- Trigger on pushes to `main` and `microservices` when service files change.
- Trigger manually with `workflow_dispatch`.
- Only services touched by the commit are rebuilt and pushed.
- If `backend/services/shared` changes, adapter/orchestrator/shardmap/sharding are all republished.
- Image tags include:
	- `latest` (default branch only)
	- branch name tag
	- commit SHA tag
