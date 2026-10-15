param(
  [switch]$SkipGoBuild
)

$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent $PSScriptRoot
$frontendDist = Join-Path $repo 'frontend/dist'
$goDist = Join-Path $repo 'backend/web/dist'
$petsSource = Join-Path $repo 'pets'
$petsDist = Join-Path $goDist 'pets'
$novelPanelSource = Join-Path $repo 'public/novel-panel'
$novelPanelDist = Join-Path $goDist 'novel-panel'
$novelFetchSource = Join-Path $repo 'frontend/public/batch-rewrite'
$novelFetchPatchSource = Join-Path $repo 'public/batch-rewrite'
$novelFetchDist = Join-Path $goDist 'batch-rewrite'
$brandAssetsSource = Join-Path $repo 'public/assets'
$brandAssetsDist = Join-Path $goDist 'assets'

if (-not (Test-Path (Join-Path $frontendDist 'index.html'))) {
  throw "Frontend build output is missing: $frontendDist"
}

New-Item -ItemType Directory -Force $goDist | Out-Null
Get-ChildItem -LiteralPath $goDist -Force | Where-Object { $_.Name -ne '.gitkeep' } | Remove-Item -Recurse -Force
Copy-Item -Path (Join-Path $frontendDist '*') -Destination $goDist -Recurse -Force

# The liquid-metal brand component reads these stable, non-hashed images from
# /assets. Keep them beside the hashed Vite assets in the embedded Go bundle.
foreach ($brandAsset in @('brand-logo-black.png', 'brand-logo-white.png')) {
  $source = Join-Path $brandAssetsSource $brandAsset
  if (-not (Test-Path $source)) {
    throw "Brand asset is missing: $source"
  }
  Copy-Item -LiteralPath $source -Destination (Join-Path $brandAssetsDist $brandAsset) -Force
}

# The Go composer assembles the novel-fetch page from the tracked public
# workbench and injects the current retry/status/date compatibility scripts.
# Go 4001 embeds that composed page instead of the older copied dist snapshot.
if (-not (Test-Path (Join-Path $novelFetchSource 'index.html'))) {
  throw "Novel fetch workbench assets are missing: $novelFetchSource"
}
New-Item -ItemType Directory -Force $novelFetchDist | Out-Null
Copy-Item -Path (Join-Path $novelFetchSource '*') -Destination $novelFetchDist -Recurse -Force
if (-not (Test-Path (Join-Path $novelFetchPatchSource 'v78-novel-fetch-v2-runtime.js'))) {
  throw "Novel fetch V78 compatibility scripts are missing: $novelFetchPatchSource"
}
Copy-Item -Path (Join-Path $novelFetchPatchSource '*') -Destination $novelFetchDist -Recurse -Force
Push-Location (Join-Path $repo 'backend')
try {
  go run ./cmd/compose-novel-fetch (Join-Path $novelFetchSource 'index.html') (Join-Path $novelFetchDist 'index.html')
} finally {
  Pop-Location
}

# Give the embedded workbench runtime a content-derived cache key. This keeps
# an already-open browser from reusing an older compatibility script after a
# Go rebuild, without modifying the tracked source page on every build.
$runtimeAsset = Join-Path $novelFetchDist 'v78-novel-fetch-v2-runtime.js'
$composedIndex = Join-Path $novelFetchDist 'index.html'
if ((Test-Path -LiteralPath $runtimeAsset) -and (Test-Path -LiteralPath $composedIndex)) {
  $runtimeVersion = (Get-FileHash -LiteralPath $runtimeAsset -Algorithm SHA256).Hash.Substring(0, 12).ToLowerInvariant()
  $indexText = [IO.File]::ReadAllText($composedIndex, [Text.Encoding]::UTF8)
  $indexText = [Text.RegularExpressions.Regex]::Replace($indexText, 'v78-novel-fetch-v2-runtime\.js\?v=[^"'' ]+', "v78-novel-fetch-v2-runtime.js?v=$runtimeVersion")
  [IO.File]::WriteAllText($composedIndex, $indexText, (New-Object Text.UTF8Encoding($false)))
}

if (-not (Test-Path (Join-Path $petsSource 'stacky/spritesheet.webp'))) {
  throw "Pet assets are missing: $petsSource"
}

New-Item -ItemType Directory -Force $petsDist | Out-Null
Copy-Item -Path (Join-Path $petsSource '*') -Destination $petsDist -Recurse -Force

if (-not (Test-Path (Join-Path $novelPanelSource 'workbench/index.html'))) {
  throw "Novel panel workbench assets are missing: $novelPanelSource"
}

New-Item -ItemType Directory -Force $novelPanelDist | Out-Null
Copy-Item -Path (Join-Path $novelPanelSource '*') -Destination $novelPanelDist -Recurse -Force

if (-not $SkipGoBuild) {
  Push-Location (Join-Path $repo 'backend')
  try {
    go build -o qiantie-local.exe ./cmd/qiantie
    $executorDownloads = Join-Path $goDist 'downloads/local-executor'
    New-Item -ItemType Directory -Force $executorDownloads | Out-Null
    $executorFile = Join-Path $executorDownloads 'qiantie-executor-v88-0.1.0-win-x64.exe'
    go build -o $executorFile ./cmd/qiantie-executor
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
    Set-Content -LiteralPath (Join-Path $buildMetaDir 'qiantie-local.build-meta') -Value @("git_sha=$buildSHA", "runtime_sha=$runtimeVersion") -Encoding utf8
  } finally {
    Pop-Location
  }
}

Write-Output "Go static assets prepared: $goDist"
