# Test Suite

This directory contains the backend test runner and shared test categories.

## Structure

- `run-tests.ps1`: PowerShell test runner
- `integration/`: Integration tests (currently empty)
- `e2e/`: End-to-end tests (currently empty)

Unit tests are now colocated with each service under:

- `services/adapter/tests/unit/`
- `services/orchestrator/tests/unit/`
- `services/shardmap/tests/unit/`
- `services/sharding/tests/unit/`

## Prerequisites

- Go installed and available in PATH
- PowerShell (Windows PowerShell or PowerShell 7)
- Run commands from the backend folder

## Run Tests

From `backend/`:

### Run all test suites

```powershell
.\tests\run-tests.ps1 -Type all
```

### Run unit tests only (all services)

```powershell
.\tests\run-tests.ps1 -Type unit
```

### Run integration tests only

```powershell
.\tests\run-tests.ps1 -Type integration
```

### Run e2e tests only

```powershell
.\tests\run-tests.ps1 -Type e2e
```

## Notes

- If a suite has no test files, the runner prints `No <type> tests found.` and continues.
- By default, each suite runs with `go test -count=1` to avoid cached test results.
- For unit tests, the runner auto-discovers `services/*/tests/unit` folders and runs each service's unit suite.
- As new integration/e2e tests are added under their folders, the runner will pick them up automatically.
