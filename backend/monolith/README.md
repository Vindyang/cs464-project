# Omnishard Monolith Backend

This directory will hold the new monolith backend implementation.

## Intended Role

The monolith backend is a separate implementation from the microservice backend in `backend/microservice`.
Code duplication is allowed by design so both backends can evolve independently.
The monolith now carries its own shared packages and has no compile-time dependency on the microservice module.

The monolith will:

- run as a single Go backend process
- expose the same user-facing API contracts currently surfaced through the gateway
- preserve the adapter-facing endpoints the frontend already calls directly
- serve API documentation without requiring a separate gateway container

## Planned Structure

- `cmd/`
  - Monolith entrypoints.
- `internal/`
  - Monolith-specific application wiring, handlers, persistence, and workflow logic.
- `shared/`
  - Monolith-owned copies of the contracts, provider integrations, storage helpers, and HTTP utilities it needs to build independently.
- `tests/`
  - Monolith-specific integration and end-to-end verification.

## How It Differs From The Microservice Backend

- `backend/microservice` deploys separate services for adapter, shardmap, sharding, orchestrator, and gateway.
- `backend/monolith` will collapse those responsibilities into one process while keeping the external HTTP behavior stable.
- `backend/microservice` uses HTTP between backend services.
- `backend/monolith` will use in-process calls internally and reserve HTTP for the external surface.

## Current Status

This directory now contains a standalone monolith backend:

- `go.mod` for an independent monolith module
- `shared/` packages owned by the monolith module itself
- `cmd/main.go` and `internal/` runtime code for the monolith service
- `Dockerfile` for a standalone monolith image build

The remaining follow-up work is around test coverage and broader endpoint parity, not module separation.