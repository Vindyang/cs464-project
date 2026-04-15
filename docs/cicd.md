# Testing and CI/CD

## Local Validation

### Frontend

From `frontend/`:

```powershell
bun install --frozen-lockfile
bun run lint
bun run build
```

### Microservice backend

From `backend/microservice/`:

```powershell
go test ./services/...
.\tests\run-tests.ps1 -Type unit
.\tests\run-tests.ps1 -Type integration
.\tests\run-tests.ps1 -Type e2e
.\tests\run-tests.ps1 -Type all
```

The grouped runner discovers:

- unit tests under `services/*/tests/unit/`
- integration tests under `tests/integration/`
- e2e tests under `tests/e2e/`

### Monolith backend

From `backend/monolith/`:

```powershell
go test ./...
go test ./tests/integration/... -count=1
```

### Source-build smoke test

Recommended manual smoke test after startup:

1. Open `http://localhost:3000`.
2. Configure at least one provider credential.
3. Connect a provider.
4. Upload a file.
5. Download the same file.
6. Check provider usage, file health, history, and docs views.

## CI Pipelines

### Main CI and publish workflow

Workflow:

- `.github/workflows/ci-main.yml`

This workflow triggers on pushes to `main` and pull requests targeting `main` when files under `backend/**`, `frontend/**`, `deploy/**`, or `.github/workflows/**` change.

It delegates most validation work to `.github/workflows/reusable-ci.yml`, which performs:

- change detection so unaffected jobs can be skipped
- microservice backend quality gates with `go vet ./...` and `go build ./...`
- service-scoped unit tests for adapter, orchestrator, shardmap, and sharding
- microservice integration contract tests
- microservice e2e backend tests
- monolith quality gate with `go build ./...` in `backend/monolith`
- frontend lint and production build

On pushes to `main`, successful CI also publishes changed images to GHCR.

Published image targets from this pipeline:

- `omnishard-adapter`
- `omnishard-shardmap`
- `omnishard-sharding`
- `omnishard-orchestrator`
- `omnishard-gateway`
- `omnishard-monolith`
- `omnishard-frontend`
- `omnishard-all-in-one`
- `omnishard-all-in-one-monolith`

### Official GitHub OSS release workflow

Workflow:

- `.github/workflows/release-github-oss.yml`

This manually triggered workflow:

- accepts an exact `commit_sha`
- accepts a semver `release_tag` such as `v1.2.3`
- validates the commit and tag inputs
- builds and pushes release-tagged images to GHCR
- renders the official release compose assets
- creates the GitHub Release and uploads the generated compose files

Official release assets produced by this workflow:

- `docker-compose.full-microservices.yml`
- `docker-compose.single-image-microservices.yml`
- `docker-compose.single-image-monolith.yml`
- `docker-compose.all-in-one-monolith.yml`

### Manual image republish workflow

Workflow:

- `.github/workflows/cd-dockerhub-force-deploy.yml`

Despite the historical name, this workflow now targets GHCR. It can republish:

- the full microservices image set
- the all-in-one microservices image
- the monolith image
- the all-in-one monolith image

Those images are tagged with `latest` and the full commit SHA.

## GHCR Naming And Pull-Only Manifests

Published image names follow the pattern `ghcr.io/<owner>/<image>`.

Current images:

- `ghcr.io/vindyang/omnishard-adapter`
- `ghcr.io/vindyang/omnishard-shardmap`
- `ghcr.io/vindyang/omnishard-sharding`
- `ghcr.io/vindyang/omnishard-orchestrator`
- `ghcr.io/vindyang/omnishard-gateway`
- `ghcr.io/vindyang/omnishard-monolith`
- `ghcr.io/vindyang/omnishard-frontend`
- `ghcr.io/vindyang/omnishard-all-in-one`
- `ghcr.io/vindyang/omnishard-all-in-one-monolith`

Repo-local pull-only compose manifests live under `deploy/compose/` and require:

- `IMAGE_NAMESPACE`, for example `ghcr.io/vindyang`
- `OMNISHARD_TAG`, for example `v1.2.3` or a full commit SHA

## Notes

- The root `docker-compose.yml` is the build-from-source developer compose file and should not be overwritten for normal repo development.
- Monolith integration tests exist under `backend/monolith/tests/integration`, but CI currently enforces a monolith build gate rather than a dedicated monolith test job.
- No extra registry secret is required for the GitHub-hosted publishing workflows; GHCR publishing uses the built-in `GITHUB_TOKEN`.