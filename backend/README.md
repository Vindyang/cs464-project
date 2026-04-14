# Backend Layout

The backend now has two explicit implementation roots:

- `backend/microservice`
  - The existing multi-service Go backend, moved intact under its own directory.
  - This is still the active backend used by the current Docker Compose, CI, and release pipelines.
- `backend/monolith`
  - The new single-process backend flavor under construction.
  - It will preserve the current application-facing contracts while removing the gateway container and inter-service deployment split.

Start here depending on which backend you need:

- Microservice backend: `backend/microservice/README.md`
- Monolith backend: `backend/monolith/README.md`

Current implementation status:

- The repository split is in place.
- Existing build, compose, and CI entrypoints now target `backend/microservice`.
- Monolith runtime and deployment assets are not wired in yet.