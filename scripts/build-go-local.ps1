param(
  [switch]$SkipGoBuild
)

$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent $PSScriptRoot
$frontendDist = Join-Path $repo 'frontend/dist'
$goDist = Join-Path $repo 'backend/web/dist'
$binaryPath = Join-Path $repo 'backend/qiantie-local.exe'
$petsSource = Join-Path $repo 'pets'
$petsDist = Join-Path $goDist 'pets'
$brandAssetsSource = Join-Path $repo 'frontend/public/assets'
$brandAssetsDist = Join-Path $goDist 'assets'

if (-not (Test-Path (Join-Path $frontendDist 'index.html'))) {
  throw "Frontend build output is missing: $frontendDist"
}

New-Item -ItemType Directory -Force $goDist | Out-Null
Get-ChildItem -LiteralPath $goDist -Force | Where-Object { $_.Name -ne '.gitkeep' } | Remove-Item -Recurse -Force
Copy-Item -Path (Join-Path $frontendDist '*') -Destination $goDist -Recurse -Force
# The reference V78 files remain under frontend/public for comparison, and Vite
# copies public/ verbatim. They must not enter the Go embed bundle: the only
# supported Novel Fetch runtime is the React page built from frontend/src.
$legacyEmbeddedRoute = Join-Path $goDist 'batch-rewrite'
if (Test-Path -LiteralPath $legacyEmbeddedRoute) {
  Remove-Item -LiteralPath $legacyEmbeddedRoute -Recurse -Force
}
# Vite copies frontend/public verbatim. The old V78 workbench remains there as
# a reference source, but it must not survive in the generated dist either:
# otherwise a stale browser route can load the old page and call retired APIs.
$legacyFrontendDist = Join-Path $frontendDist 'batch-rewrite'
if (Test-Path -LiteralPath $legacyFrontendDist) {
  Remove-Item -LiteralPath $legacyFrontendDist -Recurse -Force
}

# The liquid-metal brand component reads these stable, non-hashed images from
# /assets. Keep them beside the hashed Vite assets in the embedded Go bundle.
foreach ($brandAsset in @('brand-logo-black.png', 'brand-logo-white.png')) {
  $source = Join-Path $brandAssetsSource $brandAsset
  if (-not (Test-Path $source)) {
    throw "Brand asset is missing: $source"
  }
  Copy-Item -LiteralPath $source -Destination (Join-Path $brandAssetsDist $brandAsset) -Force
}

if (-not (Test-Path (Join-Path $petsSource 'stacky/spritesheet.webp'))) {
  throw "Pet assets are missing: $petsSource"
}

New-Item -ItemType Directory -Force $petsDist | Out-Null
Copy-Item -Path (Join-Path $petsSource '*') -Destination $petsDist -Recurse -Force

if (-not $SkipGoBuild) {
  Push-Location (Join-Path $repo 'backend')
  try {
    go build -o $binaryPath ./cmd/qiantie
    if ($LASTEXITCODE -ne 0) {
      throw "Go API build failed with exit code $LASTEXITCODE"
    }
    $executorDownloads = Join-Path $goDist 'downloads/local-executor'
    New-Item -ItemType Directory -Force $executorDownloads | Out-Null
    $executorFile = Join-Path $executorDownloads 'qiantie-executor-v88-0.1.0-win-x64.exe'
    go build -o $executorFile ./cmd/qiantie-executor
    if ($LASTEXITCODE -ne 0) {
      throw "Local executor build failed with exit code $LASTEXITCODE"
    }
    # The API serves operator-provided downloads from data/downloads, while
    # backend/web/dist is the embedded fallback. Keep both copies in sync so
    # a freshly built Go runtime never advertises a 503 installer URL.
    $runtimeDownloads = Join-Path $repo 'data/downloads'
    New-Item -ItemType Directory -Force $runtimeDownloads | Out-Null
    Copy-Item -LiteralPath $executorFile -Destination (Join-Path $runtimeDownloads (Split-Path $executorFile -Leaf)) -Force

    # Keep a small ignored build marker beside the executable. The launcher
    # uses it to detect a commit made after the last binary build, including
    # commits that only changed embedded assets or release scripts.
    $buildMetaDir = Join-Path $repo '.direct-staging'
    New-Item -ItemType Directory -Force $buildMetaDir | Out-Null
    $buildSHA = (& git -C $repo rev-parse --short HEAD).Trim()
    Set-Content -LiteralPath (Join-Path $buildMetaDir 'qiantie-local.build-meta') -Value "git_sha=$buildSHA" -Encoding utf8

    $indexHTML = Get-Content -LiteralPath (Join-Path $frontendDist 'index.html') -Raw
    $entryMatch = [regex]::Match($indexHTML, '<script[^>]+type=["'']module["''][^>]+src=["'']([^"'']+\.js)["'']')
    if (-not $entryMatch.Success) {
      throw "Frontend module entry was not found in: $frontendDist/index.html"
    }
    $runtimeManifestDir = Join-Path $repo '.local-review'
    New-Item -ItemType Directory -Force $runtimeManifestDir | Out-Null
    @{ gitSha = $buildSHA; frontendEntry = $entryMatch.Groups[1].Value; builtAt = (Get-Date).ToUniversalTime().ToString('o'); binaryPath = $binaryPath } |
      ConvertTo-Json -Compress |
      Set-Content -LiteralPath (Join-Path $runtimeManifestDir 'runtime-manifest.json') -Encoding utf8
  } finally {
    Pop-Location
  }
}

Write-Output "Go static assets prepared: $goDist"
