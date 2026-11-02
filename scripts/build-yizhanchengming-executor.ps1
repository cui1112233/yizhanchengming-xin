param()

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$executorRoot = Join-Path $repoRoot 'executors\doubao-relay'
$frontendRoot = Join-Path $executorRoot 'frontend'
$cargo = Join-Path $env:USERPROFILE '.cargo\bin\cargo.exe'
$logo = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('RzovQmFja3VwL1BpY3R1cmVzL+S4gOaImOaZn+mTreWbvuaghy5wbmc='))
$installerName = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('5LiA5oiY5pmf6ZOt5omn6KGM5ZmoLeWuieijheeJiC5leGU='))
$outputDir = Join-Path $repoRoot 'data\downloads'
$outputFile = Join-Path $outputDir $installerName

if (-not (Test-Path -LiteralPath $cargo)) { throw "Rust was not found: $cargo" }
$pnpmCommand = Get-Command pnpm -ErrorAction SilentlyContinue
if (-not $pnpmCommand) { throw 'pnpm was not found' }
if (-not (Get-Command node -ErrorAction SilentlyContinue)) { throw 'Node.js was not found' }
if (-not (Test-Path -LiteralPath $logo)) { throw "Logo was not found: $logo" }
$node = (Get-Command node).Source
$bundledNode = Join-Path (Split-Path -Parent $pnpmCommand.Source) '..\..\node\bin\node.exe'
if (Test-Path -LiteralPath $bundledNode) { $node = (Resolve-Path -LiteralPath $bundledNode).Path }

Push-Location $executorRoot
try {
    & $cargo tauri icon $logo
    if ($LASTEXITCODE -ne 0) { throw 'Icon generation failed' }

    Push-Location $frontendRoot
    & $node '.\node_modules\vitest\vitest.mjs' --run
    if ($LASTEXITCODE -ne 0) { throw 'Frontend tests failed' }

    & $node '.\node_modules\typescript\bin\tsc' --noEmit
    if ($LASTEXITCODE -ne 0) { throw 'Frontend type check failed' }
    # Vite 6 on Windows can exit early after transform when the compact reporter
    # is attached to a non-interactive PowerShell host. Debug logging avoids it.
    & $node '.\node_modules\vite\bin\vite.js' build --debug
    if ($LASTEXITCODE -ne 0) { throw 'Frontend build failed' }
    Pop-Location

    & $cargo tauri build --bundles nsis --config 'src-tauri\tauri.build.conf.json'
    if ($LASTEXITCODE -ne 0) { throw 'NSIS build failed' }
} finally {
    if ((Get-Location).Path -eq $frontendRoot) { Pop-Location }
    Pop-Location
}

$bundleDir = Join-Path $executorRoot 'target\release\bundle\nsis'
$bundles = @(Get-ChildItem -LiteralPath $bundleDir -Filter '*.exe' -File | Sort-Object LastWriteTime -Descending)
if ($bundles.Count -lt 1) { throw 'Expected an NSIS installer, but none was produced' }
if ($bundles.Count -gt 1) {
    Write-Host "Found $($bundles.Count) NSIS installers; using the newest one: $($bundles[0].Name)"
}

New-Item -ItemType Directory -Path $outputDir -Force | Out-Null
Copy-Item -LiteralPath $bundles[0].FullName -Destination $outputFile -Force
$artifact = Get-Item -LiteralPath $outputFile
$hash = Get-FileHash -LiteralPath $outputFile -Algorithm SHA256
Write-Host "Installer: $($artifact.FullName)"
Write-Host "Size: $($artifact.Length) bytes"
Write-Host "SHA-256: $($hash.Hash)"
