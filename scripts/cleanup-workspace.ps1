param(
    [switch]$DryRun
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$removedPaths = New-Object System.Collections.Generic.List[string]
$failedPaths = New-Object System.Collections.Generic.List[string]

function Get-RelativePath {
    param(
        [string]$FullPath
    )

    if ($FullPath.StartsWith($repoRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
        $relative = $FullPath.Substring($repoRoot.Length).TrimStart("\", "/")
        if ($relative) {
            return $relative
        }
    }

    return $FullPath
}

function Remove-Matches {
    param(
        [Parameter(Mandatory = $true)]
        [System.IO.FileSystemInfo[]]$Items
    )

    foreach ($item in $Items) {
        $relativePath = Get-RelativePath -FullPath $item.FullName

        if ($DryRun) {
            Write-Host "[dry-run] remove $relativePath"
            continue
        }

        $removeError = $null
        for ($attempt = 1; $attempt -le 5; $attempt++) {
            try {
                Remove-Item -LiteralPath $item.FullName -Recurse -Force
                $removeError = $null
                break
            }
            catch {
                $removeError = $_
                if ($attempt -lt 5) {
                    Start-Sleep -Seconds 2
                }
            }
        }

        if ($removeError) {
            $failedPaths.Add($relativePath) | Out-Null
            Write-Warning "Could not remove $relativePath after multiple attempts: $($removeError.Exception.Message)"
            continue
        }

        $removedPaths.Add($relativePath) | Out-Null
        Write-Host "[removed] $relativePath"
    }
}

function Remove-PathIfExists {
    param(
        [Parameter(Mandatory = $true)]
        [string]$RelativePath
    )

    $fullPath = Join-Path $repoRoot $RelativePath
    if (-not (Test-Path -LiteralPath $fullPath)) {
        return
    }

    $item = Get-Item -Force -LiteralPath $fullPath
    Remove-Matches -Items @($item)
}

function Remove-GlobIfExists {
    param(
        [Parameter(Mandatory = $true)]
        [string]$RelativePattern
    )

    $pattern = Join-Path $repoRoot $RelativePattern
    $items = @(Get-ChildItem -Force -Path $pattern -ErrorAction SilentlyContinue)
    if ($items.Count -eq 0) {
        return
    }

    Remove-Matches -Items $items
}

Remove-PathIfExists -RelativePath "node_modules"
Remove-PathIfExists -RelativePath "apps/server-web/node_modules"
Remove-PathIfExists -RelativePath "apps/server-web/dist"
Remove-PathIfExists -RelativePath "apps/server-web/coverage"
Remove-PathIfExists -RelativePath "apps/server-api/.venv"
Remove-PathIfExists -RelativePath "apps/server-api/.pytest_cache"
Remove-PathIfExists -RelativePath "apps/server-api/.coverage"
Remove-PathIfExists -RelativePath "apps/server-api/.data"
Remove-GlobIfExists -RelativePattern "tmp-*"
Remove-GlobIfExists -RelativePattern "docs/assets/screenshots/*.png"

if ($DryRun) {
    Write-Host "Dry run finished."
    exit 0
}

if ($failedPaths.Count -gt 0) {
    if ($removedPaths.Count -gt 0) {
        Write-Host "Removed $($removedPaths.Count) generated workspace artifact(s)."
    }
    Write-Warning "Still locked or unavailable: $($failedPaths -join ', ')"
    exit 1
}

if ($removedPaths.Count -eq 0) {
    Write-Host "No generated workspace artifacts were found."
    exit 0
}

Write-Host "Removed $($removedPaths.Count) generated workspace artifact(s)."
