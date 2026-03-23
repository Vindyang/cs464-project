# Test Suite

This directory contains the backend test suites, organized by test type.

## Structure

- `unit/`: Unit tests
- `integration/`: Integration tests (currently empty)
- `e2e/`: End-to-end tests (currently empty)
- `run-tests.ps1`: PowerShell test runner

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

### Run unit tests only

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
- As new integration/e2e tests are added under their folders, the runner will pick them up automatically.
