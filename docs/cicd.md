# Testing and CI/CD

## Testing

### Backend tests

From `backend/`:

```powershell
go test ./...
```

### Frontend checks

From `frontend/`:

```powershell
bun run lint
bun run build
```

### PowerShell test runner

The backend includes a grouped runner at `backend/tests/run-tests.ps1`.

Examples:

```powershell
./tests/run-tests.ps1 -Type unit
./tests/run-tests.ps1 -Type integration
./tests/run-tests.ps1 -Type e2e
./tests/run-tests.ps1 -Type all
```

### Test layers

Unit tests:

- live under `backend/services/*/tests/unit/`
- target isolated service logic

Integration tests:

- live under `backend/tests/integration/`
- validate contracts between orchestrator and the supporting services using controlled test doubles

E2E tests:

- live under `backend/tests/e2e/`
- exercise the real Dockerized stack and end-to-end workflows

### Recommended smoke test after startup

1. Open `http://localhost:3000`.
2. Configure at least one provider credential.
3. Connect a provider.
4. Upload a file.
5. Download the same file.
6. Check provider usage, file health, and history views.

## CI/CD

### Main CI pipeline

Workflow:

- `.github/workflows/ci-main.yml`

What it does:

- triggers on pushes and pull requests affecting `backend/**`, `frontend/**`, `deploy/**`, or workflow files
- resolves whether image deployment should be enabled
- runs change detection to avoid unnecessary service jobs
- runs backend quality gates with `go vet ./...` and `go build ./...`
- runs service-scoped unit tests
- runs integration contract tests
- runs E2E backend tests
- runs frontend lint and build
- publishes changed images to GHCR on pushes to `main`

Branch behavior:

- `main` push: CI plus image publishing
- `microservices` push: CI only
- pull request: CI only

### Official release workflow

Workflow:

- `.github/workflows/release-github-oss.yml`

What it does:

- accepts an exact `commit_sha`
- accepts a semver `release_tag` such as `v1.2.3`
- validates the tag and commit
- builds and pushes release-tagged images for all services to GHCR
- renders official release compose assets
- creates a GitHub Release and uploads the generated compose files

### Manual image republish workflow

Workflow:

- `.github/workflows/cd-dockerhub-force-deploy.yml`

What it does:

- manually republishes the full microservices image set and or the all-in-one image
- tags them with `latest` and the full commit SHA

### Published image names

- `ghcr.io/vindyang/omnishard-adapter`
- `ghcr.io/vindyang/omnishard-shardmap`
- `ghcr.io/vindyang/omnishard-sharding`
- `ghcr.io/vindyang/omnishard-orchestrator`
- `ghcr.io/vindyang/omnishard-gateway`
- `ghcr.io/vindyang/omnishard-frontend`
- `ghcr.io/vindyang/omnishard-all-in-one`

### Required GitHub configuration

Required for GitHub-hosted publishing workflows:

- no extra registry secret is required; workflows publish to GHCR using the built-in `GITHUB_TOKEN`

Required for local repo-based release manifests:

- `IMAGE_NAMESPACE`, for example `ghcr.io/vindyang`