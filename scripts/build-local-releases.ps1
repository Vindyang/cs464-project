param(
    [string]$ImageTag = "local",
    [string]$ImageNamespace = "ghcr.io/vindyang",
    [string]$Platform = "linux/amd64",
    [string]$ReleaseAssetsOutputDir = "deploy/release-assets/dist-local",
    [string[]]$Include = @(),
    [switch]$SkipRenderAssets
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Format-SizeMB {
    param([long]$Bytes)

    return [math]::Round($Bytes / 1e6, 2)
}

function Convert-SizeStringToBytes {
    param([string]$Size)

    if ([string]::IsNullOrWhiteSpace($Size) -or $Size -eq "N/A") {
        return $null
    }

    $match = [regex]::Match($Size.Trim(), '^(?<value>[0-9]+(?:\.[0-9]+)?)(?<unit>B|kB|MB|GB|TB)$')
    if (-not $match.Success) {
        throw "Unsupported size format: $Size"
    }

    $value = [double]$match.Groups['value'].Value
    $unit = $match.Groups['unit'].Value

    $multiplier = switch ($unit) {
        'B' { 1 }
        'kB' { 1e3 }
        'MB' { 1e6 }
        'GB' { 1e9 }
        'TB' { 1e12 }
        default { throw "Unsupported size unit: $unit" }
    }

    return [int64][math]::Round($value * $multiplier)
}

function Get-ImageDiskUsageMap {
    $lines = @(& docker system df -v)
    $imageRows = @{}
    $inImagesSection = $false

    foreach ($line in $lines) {
        if ($line -eq 'Images space usage:') {
            $inImagesSection = $true
            continue
        }

        if (-not $inImagesSection) {
            continue
        }

        if ($line -eq 'Containers space usage:') {
            break
        }

        if ([string]::IsNullOrWhiteSpace($line) -or $line -like 'REPOSITORY*') {
            continue
        }

        $parts = $line -split '\s{2,}'
        if ($parts.Count -lt 8) {
            continue
        }

        $repository = $parts[0]
        $tag = $parts[1]
        $size = $parts[4]
        $sharedSize = $parts[5]
        $uniqueSize = $parts[6]
        $key = '{0}:{1}' -f $repository, $tag

        $imageRows[$key] = [pscustomobject]@{
            Repository = $repository
            Tag = $tag
            LocalUnpackedBytes = Convert-SizeStringToBytes -Size $size
            LocalUnpackedDisplay = $size
            SharedLocalBytes = Convert-SizeStringToBytes -Size $sharedSize
            SharedLocalDisplay = $sharedSize
            UniqueLocalBytes = Convert-SizeStringToBytes -Size $uniqueSize
            UniqueLocalDisplay = $uniqueSize
        }
    }

    return $imageRows
}

function Get-CompressedArchiveSizeBytes {
    param([string]$ImageRef)

    $tempTar = Join-Path ([System.IO.Path]::GetTempPath()) (([System.IO.Path]::GetRandomFileName()) + '.tar')
    $tempGz = Join-Path ([System.IO.Path]::GetTempPath()) (([System.IO.Path]::GetRandomFileName()) + '.tar.gz')

    try {
        & docker image save -o $tempTar $ImageRef
        if ($LASTEXITCODE -ne 0) {
            throw "docker image save failed for $ImageRef"
        }

        $inputStream = [System.IO.File]::OpenRead($tempTar)
        try {
            $outputStream = [System.IO.File]::Create($tempGz)
            try {
                $gzipStream = New-Object System.IO.Compression.GZipStream($outputStream, [System.IO.Compression.CompressionLevel]::Optimal)
                try {
                    $inputStream.CopyTo($gzipStream)
                }
                finally {
                    $gzipStream.Dispose()
                }
            }
            finally {
                $outputStream.Dispose()
            }
        }
        finally {
            $inputStream.Dispose()
        }

        return (Get-Item -Path $tempGz).Length
    }
    finally {
        if (Test-Path -Path $tempTar) {
            Remove-Item -Path $tempTar -Force
        }

        if (Test-Path -Path $tempGz) {
            Remove-Item -Path $tempGz -Force
        }
    }
}

function Render-Template {
    param(
        [string]$TemplatePath,
        [string]$OutputPath,
        [string]$ReleaseTag,
        [string]$Namespace
    )

    $content = Get-Content -Path $TemplatePath -Raw
    $content = $content.Replace("__OMNISHARD_TAG__", $ReleaseTag)
    $content = $content.Replace("__IMAGE_NAMESPACE__", $Namespace)

    $outputDir = Split-Path -Parent $OutputPath
    if (-not (Test-Path -Path $outputDir)) {
        New-Item -ItemType Directory -Path $outputDir | Out-Null
    }

    Set-Content -Path $OutputPath -Value $content -NoNewline
}

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$outputRoot = Join-Path $repoRoot "scripts/out/$ImageTag"

$buildMatrix = @(
    [pscustomobject]@{ Name = "omnishard-adapter"; Context = "backend/microservice"; Dockerfile = "backend/microservice/services/adapter/Dockerfile"; BuildArgs = @() }
    [pscustomobject]@{ Name = "omnishard-shardmap"; Context = "backend/microservice"; Dockerfile = "backend/microservice/services/shardmap/Dockerfile"; BuildArgs = @() }
    [pscustomobject]@{ Name = "omnishard-sharding"; Context = "backend/microservice"; Dockerfile = "backend/microservice/services/sharding/Dockerfile"; BuildArgs = @() }
    [pscustomobject]@{ Name = "omnishard-orchestrator"; Context = "backend/microservice"; Dockerfile = "backend/microservice/services/orchestrator/Dockerfile"; BuildArgs = @() }
    [pscustomobject]@{ Name = "omnishard-gateway"; Context = "backend/microservice"; Dockerfile = "backend/microservice/services/gateway/Dockerfile"; BuildArgs = @() }
    [pscustomobject]@{ Name = "omnishard-monolith"; Context = "backend/monolith"; Dockerfile = "backend/monolith/Dockerfile"; BuildArgs = @() }
    [pscustomobject]@{ Name = "omnishard-frontend"; Context = "frontend"; Dockerfile = "frontend/Dockerfile"; BuildArgs = @("--build-arg", "NEXT_PUBLIC_API_URL=http://localhost:8080") }
    [pscustomobject]@{ Name = "omnishard-all-in-one"; Context = "."; Dockerfile = "deploy/images/all-in-one/Dockerfile"; BuildArgs = @("--build-arg", "NEXT_PUBLIC_API_URL=http://localhost:8080") }
    [pscustomobject]@{ Name = "omnishard-all-in-one-monolith"; Context = "."; Dockerfile = "deploy/images/all-in-one-monolith/Dockerfile"; BuildArgs = @("--build-arg", "NEXT_PUBLIC_API_URL=http://localhost:8080") }
)

if ($Include.Count -gt 0) {
    $buildMatrix = $buildMatrix | Where-Object { $Include -contains $_.Name }
}

if ($buildMatrix.Count -eq 0) {
    throw "No release images selected. Pass -Include with one or more valid image names."
}

Push-Location $repoRoot
try {
    foreach ($build in $buildMatrix) {
        $imageRef = "{0}:{1}" -f $build.Name, $ImageTag
        $dockerArgs = @(
            "build",
            "--platform", $Platform,
            "-f", $build.Dockerfile,
            "-t", $imageRef
        ) + $build.BuildArgs + @($build.Context)

        Write-Host "Building $imageRef" -ForegroundColor Cyan
        & docker @dockerArgs
        if ($LASTEXITCODE -ne 0) {
            throw "docker build failed for $imageRef"
        }
    }

    if (-not $SkipRenderAssets) {
        $assetDir = Join-Path $repoRoot $ReleaseAssetsOutputDir
        $legacyOutputs = @(
            "docker-compose.full-microservices.yml"
            "docker-compose.single-image-microservices.yml"
            "docker-compose.single-image-monolith.yml"
        )
        $templates = @(
            [pscustomobject]@{ Template = "deploy/release-assets/templates/docker-compose.microservices.yml.tpl"; Output = "docker-compose.microservices.yml" }
            [pscustomobject]@{ Template = "deploy/release-assets/templates/docker-compose.all-in-one-microservices.yml.tpl"; Output = "docker-compose.all-in-one-microservices.yml" }
            [pscustomobject]@{ Template = "deploy/release-assets/templates/docker-compose.monolith.yml.tpl"; Output = "docker-compose.monolith.yml" }
            [pscustomobject]@{ Template = "deploy/release-assets/templates/docker-compose.all-in-one-monolith.yml.tpl"; Output = "docker-compose.all-in-one-monolith.yml" }
        )

        foreach ($legacyOutput in $legacyOutputs) {
            $legacyPath = Join-Path $assetDir $legacyOutput
            if (Test-Path -Path $legacyPath) {
                Remove-Item -Path $legacyPath -Force
            }
        }

        foreach ($asset in $templates) {
            Render-Template -TemplatePath (Join-Path $repoRoot $asset.Template) -OutputPath (Join-Path $assetDir $asset.Output) -ReleaseTag $ImageTag -Namespace $ImageNamespace
        }
    }

    if (-not (Test-Path -Path $outputRoot)) {
        New-Item -ItemType Directory -Path $outputRoot | Out-Null
    }

    $imageDiskUsage = Get-ImageDiskUsageMap

    $imageStats = foreach ($build in $buildMatrix) {
        $imageRef = "{0}:{1}" -f $build.Name, $ImageTag
        $diskUsage = $imageDiskUsage[$imageRef]
        if ($null -eq $diskUsage) {
            throw "Unable to find docker system df usage data for $imageRef"
        }

        $compressedArchiveBytes = Get-CompressedArchiveSizeBytes -ImageRef $imageRef

        [pscustomobject]@{
            Image = $build.Name
            Tag = $ImageTag
            CompressedArchiveBytes = $compressedArchiveBytes
            CompressedArchiveMB = Format-SizeMB -Bytes $compressedArchiveBytes
            LocalUnpackedBytes = [int64]$diskUsage.LocalUnpackedBytes
            LocalUnpackedMB = Format-SizeMB -Bytes ([int64]$diskUsage.LocalUnpackedBytes)
            LocalUnpackedDisplay = $diskUsage.LocalUnpackedDisplay
            UniqueLocalBytes = [int64]$diskUsage.UniqueLocalBytes
            UniqueLocalMB = Format-SizeMB -Bytes ([int64]$diskUsage.UniqueLocalBytes)
            UniqueLocalDisplay = $diskUsage.UniqueLocalDisplay
        }
    }

    $imageStats | Sort-Object Image | Export-Csv -Path (Join-Path $outputRoot "image-sizes.csv") -NoTypeInformation

    $imageSizeByName = @{}
    foreach ($image in $imageStats) {
        $imageSizeByName[$image.Image] = $image
    }

    $releaseMatrix = @(
        [pscustomobject]@{ Release = "microservices"; Images = @("omnishard-frontend", "omnishard-adapter", "omnishard-shardmap", "omnishard-sharding", "omnishard-orchestrator", "omnishard-gateway") }
        [pscustomobject]@{ Release = "all-in-one-microservices"; Images = @("omnishard-all-in-one") }
        [pscustomobject]@{ Release = "monolith"; Images = @("omnishard-frontend", "omnishard-monolith") }
        [pscustomobject]@{ Release = "all-in-one-monolith"; Images = @("omnishard-all-in-one-monolith") }
    )

    $releaseStats = foreach ($release in $releaseMatrix) {
        $available = @($release.Images | Where-Object { $imageSizeByName.ContainsKey($_) })
        if ($available.Count -ne $release.Images.Count) {
            continue
        }

        $totalCompressedArchiveBytes = 0L
        $totalLocalUnpackedBytes = 0L
        $totalUniqueLocalBytes = 0L
        foreach ($imageName in $release.Images) {
            $totalCompressedArchiveBytes += $imageSizeByName[$imageName].CompressedArchiveBytes
            $totalLocalUnpackedBytes += $imageSizeByName[$imageName].LocalUnpackedBytes
            $totalUniqueLocalBytes += $imageSizeByName[$imageName].UniqueLocalBytes
        }

        [pscustomobject]@{
            Release = $release.Release
            ImageCount = $release.Images.Count
            Images = ($release.Images -join ", ")
            CompressedArchiveBytes = $totalCompressedArchiveBytes
            CompressedArchiveMB = Format-SizeMB -Bytes $totalCompressedArchiveBytes
            LocalUnpackedBytes = $totalLocalUnpackedBytes
            LocalUnpackedMB = Format-SizeMB -Bytes $totalLocalUnpackedBytes
            UniqueLocalBytes = $totalUniqueLocalBytes
            UniqueLocalMB = Format-SizeMB -Bytes $totalUniqueLocalBytes
        }
    }

    $releaseStats | Sort-Object Release | Export-Csv -Path (Join-Path $outputRoot "release-sizes.csv") -NoTypeInformation

    Write-Host ""
    Write-Host "Image Sizes" -ForegroundColor Green
    $imageStats | Sort-Object Image | Format-Table `
        Image,
        Tag,
        @{ Label = 'ArchiveMB'; Expression = { $_.CompressedArchiveMB } },
        @{ Label = 'UnpackedMB'; Expression = { $_.LocalUnpackedMB } },
        @{ Label = 'UniqueMB'; Expression = { $_.UniqueLocalMB } } -AutoSize

    Write-Host ""
    Write-Host "Release Totals" -ForegroundColor Green
    $releaseStats | Sort-Object Release | Format-Table `
        Release,
        ImageCount,
        @{ Label = 'ArchiveMB'; Expression = { $_.CompressedArchiveMB } },
        @{ Label = 'UnpackedMB'; Expression = { $_.LocalUnpackedMB } },
        @{ Label = 'UniqueMB'; Expression = { $_.UniqueLocalMB } } -AutoSize

    Write-Host ""
    Write-Host "CSV reports written to $outputRoot" -ForegroundColor Green
    if (-not $SkipRenderAssets) {
        Write-Host "Rendered release assets to $(Join-Path $repoRoot $ReleaseAssetsOutputDir)" -ForegroundColor Green
    }
}
finally {
    Pop-Location
}