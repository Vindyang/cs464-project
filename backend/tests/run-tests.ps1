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

function Run-UnitTestsByService() {
    $unitDirs = Get-ChildItem -Path "services" -Directory -ErrorAction SilentlyContinue |
        ForEach-Object { Join-Path $_.FullName "tests\unit" } |
        Where-Object { Test-Path $_ }

    if (-not $unitDirs) {
        Write-Host "No unit tests found."
        return
    }

    foreach ($dir in $unitDirs) {
        $tests = Get-ChildItem -Path $dir -Filter *_test.go -Recurse -File -ErrorAction SilentlyContinue
        if (-not $tests) {
            continue
        }

        $relativeDir = $dir.Replace((Get-Location).Path + "\", "").Replace("\", "/")
        Write-Host "Running unit tests in $relativeDir..."
        & go test "./$relativeDir/..." -count=1
        if ($LASTEXITCODE -ne 0) {
            throw "unit tests failed in $relativeDir"
        }
    }
}

try {
    switch ($Type) {
        "unit" {
            Run-UnitTestsByService
        }
        "integration" {
            Run-TestGroup "integration" "tests/integration"
        }
        "e2e" {
            Run-TestGroup "e2e" "tests/e2e"
        }
        "all" {
            Run-UnitTestsByService
            Run-TestGroup "integration" "tests/integration"
            Run-TestGroup "e2e" "tests/e2e"
        }
    }

    Write-Host "Test run completed successfully."
} finally {
    Pop-Location
}
