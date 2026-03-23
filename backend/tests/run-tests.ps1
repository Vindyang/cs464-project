param(
    [ValidateSet("all", "unit", "integration", "e2e")]
    [string]$Type = "all"
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
Push-Location $repoRoot

function Run-TestGroup([string]$label, [string]$folderPath) {
    $tests = Get-ChildItem -Path $folderPath -Filter *_test.go -Recurse -File -ErrorAction SilentlyContinue
    if (-not $tests) {
        Write-Host "No $label tests found."
        return
    }

    Write-Host "Running $label tests..."
    & go test "./$folderPath/..." -count=1
    if ($LASTEXITCODE -ne 0) {
        throw "$label tests failed"
    }
}

try {
    switch ($Type) {
        "unit" {
            Run-TestGroup "unit" "tests/unit"
        }
        "integration" {
            Run-TestGroup "integration" "tests/integration"
        }
        "e2e" {
            Run-TestGroup "e2e" "tests/e2e"
        }
        "all" {
            Run-TestGroup "unit" "tests/unit"
            Run-TestGroup "integration" "tests/integration"
            Run-TestGroup "e2e" "tests/e2e"
        }
    }

    Write-Host "Test run completed successfully."
} finally {
    Pop-Location
}
