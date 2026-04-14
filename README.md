# CS464 Project

Omnishard is a distributed cloud storage platform that shards files with Reed-Solomon encoding, stores shard data across external cloud providers, tracks placement and lifecycle metadata, and reconstructs files from any viable shard subset during download.

The stack consists of a Next.js frontend, Go backend services, SQLite metadata stores, and an NGINX gateway.

## Run Omnishard

### Recommended: official OSS release assets

This is the recommended way to start the project if you want a ready-to-run deployment without building images locally.

Full microservices deployment:

```powershell
wget -O docker-compose.yml https://github.com/Vindyang/cs464-project/releases/latest/download/docker-compose.full-microservices.yml
docker compose up -d
```

Single-image deployment:

```powershell
wget -O docker-compose.yml https://github.com/Vindyang/cs464-project/releases/latest/download/docker-compose.single-image-microservices.yml
docker compose up -d
```

Default endpoints:

- Frontend UI: `http://localhost:3000`
- Adapter API: `http://localhost:8080`

Stop the deployment:

```powershell
docker compose down
```

### Local source-build workflow

Use this when developing from the repository checkout.

Prerequisites:

- Go `1.25.0`
- Bun `1.x`
- Docker Desktop with the `docker compose` plugin

Start the full stack from the project root:

```powershell
docker compose --profile full up -d --build
```

Default endpoints:

- Frontend UI: `http://localhost:3000`
- Adapter: `http://localhost:8080`
- Shardmap: `http://localhost:8081`
- Orchestrator: `http://localhost:8082`
- Sharding: `http://localhost:8083`
- Gateway: `http://localhost:8084`

Stop the local stack:

```powershell
docker compose --profile full down
```

Reset local persisted data:

```powershell
docker compose --profile full down -v
```

### Repo-local release manifests

These pull published images without using the official downloaded release asset.

Full microservices manifest:

```powershell
$env:IMAGE_NAMESPACE = "ghcr.io/vindyang"
$env:OMNISHARD_TAG = "<release-tag-or-commit-sha>"
docker compose -f deploy/compose/full-microservices.yml up -d
```

Single-image manifest:

```powershell
docker compose -f deploy/compose/single-image-microservices.yml up -d
```

## Documentation

Further reference:

- [docs/architecture.md](docs/architecture.md) for architecture, service boundaries, request flows, and database schema.
- [docs/cicd.md](docs/cicd.md) for testing workflows, CI/CD pipelines, and release process details.
- [DEVDOCS.md](DEVDOCS.md) for contributor notes and operator setup details.
- [TODO.md](TODO.md) for the current project backlog.
- [references/CS464 Proposal.md](references/CS464%20Proposal.md) for the original project proposal.
- [references/service-breakdown.md](references/service-breakdown.md) for an additional service-level reference.
