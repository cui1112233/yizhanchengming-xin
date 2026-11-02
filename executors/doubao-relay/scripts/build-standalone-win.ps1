param()

$ErrorActionPreference = 'Stop'

$executorRoot = Split-Path -Parent $PSScriptRoot
$frontendRoot = Join-Path $executorRoot 'frontend'
$frontendEntry = Join-Path $frontendRoot 'dist\index.html'
$cargo = Join-Path $env:USERPROFILE '.cargo\bin\cargo.exe'
$pnpm = Get-Command pnpm -ErrorAction Stop
$bundledNode = Join-Path (Split-Path -Parent $pnpm.Source) '..\..\node\bin\node.exe'
$node = if (Test-Path -LiteralPath $bundledNode) { (Resolve-Path -LiteralPath $bundledNode).Path } else { (Get-Command node -ErrorAction Stop).Source }
$releaseDir = Join-Path $executorRoot 'release'
$outputFile = Join-Path $releaseDir 'doubao-relay-standalone-installer.exe'
$hashFile = Join-Path $releaseDir 'doubao-relay-standalone-installer.sha256'

if (-not (Test-Path -LiteralPath $cargo)) { throw "Rust was not found: $cargo" }
if (-not (Test-Path -LiteralPath (Join-Path $frontendRoot 'node_modules\vite\bin\vite.js'))) {
    throw 'Frontend dependencies are missing; do not build a desktop package without them.'
}

Push-Location $frontendRoot
try {
    & $node '.\node_modules\typescript\bin\tsc' --noEmit
    if ($LASTEXITCODE -ne 0) { throw 'Frontend type check failed.' }

    & $node '.\node_modules\vite\bin\vite.js' build --debug
    if ($LASTEXITCODE -ne 0) { throw 'Frontend build failed.' }
} finally {
    Pop-Location
}

if (-not (Test-Path -LiteralPath $frontendEntry)) { throw 'Bundled frontend entry is missing.' }
$frontendHtml = Get-Content -LiteralPath $frontendEntry -Raw
$scriptReference = [regex]::Match($frontendHtml, 'src="/assets/([^"/]+\.js)"')
if (-not $scriptReference.Success) { throw 'Bundled frontend JavaScript entry is missing.' }
$frontendBundle = Join-Path $frontendRoot ('dist\assets\' + $scriptReference.Groups[1].Value)
if (-not (Test-Path -LiteralPath $frontendBundle)) { throw "Bundled frontend JavaScript is missing: $frontendBundle" }

Push-Location $executorRoot
try {
    & $cargo tauri build --bundles nsis --config 'src-tauri\tauri.build.conf.json'
    if ($LASTEXITCODE -ne 0) { throw 'Tauri NSIS build failed.' }
} finally {
    Pop-Location
}

$bundleDir = Join-Path $executorRoot 'target\release\bundle\nsis'
$bundles = @(Get-ChildItem -LiteralPath $bundleDir -Filter '*.exe' -File | Sort-Object LastWriteTime -Descending)
if ($bundles.Count -ne 1) { throw "Expected one Tauri NSIS installer, found $($bundles.Count)." }
$bundle = $bundles[0]

if ($bundle.FullName -match '\\target\\release\\doubao-relay\.exe$') {
    throw 'The raw cargo executable is not an installable standalone package.'
}
if ($bundle.DirectoryName -notmatch '\\target\\release\\bundle\\nsis$') {
    throw "Unexpected installer origin: $($bundle.FullName)"
}

New-Item -ItemType Directory -Path $releaseDir -Force | Out-Null
Copy-Item -LiteralPath $bundle.FullName -Destination $outputFile -Force
$hash = (Get-FileHash -LiteralPath $outputFile -Algorithm SHA256).Hash
$hash | Set-Content -LiteralPath $hashFile -Encoding Ascii -NoNewline
Write-Host "Standalone installer: $outputFile"
Write-Host "SHA-256: $hash"
