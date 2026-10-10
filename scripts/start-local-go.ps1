param(
  [switch]$CheckOnly,
  [string]$ListenAddr = $(if ($env:QIANTIE_GO_LISTEN_ADDR) { $env:QIANTIE_GO_LISTEN_ADDR } else { ':4001' })
)

$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent $PSScriptRoot

if ($CheckOnly) {
  & (Join-Path $PSScriptRoot 'check-local-go.ps1')
  exit $LASTEXITCODE
}

$binaryPath = Join-Path $repo 'backend/qiantie-local.exe'
$goSourceRoot = Join-Path $repo 'backend'
$needsBuild = -not (Test-Path $binaryPath)
if (-not $needsBuild) {
  $binaryTime = (Get-Item $binaryPath).LastWriteTimeUtc
  $latestGoSource = Get-ChildItem $goSourceRoot -Recurse -Filter '*.go' -File | Sort-Object LastWriteTimeUtc -Descending | Select-Object -First 1
  $needsBuild = $latestGoSource -and $latestGoSource.LastWriteTimeUtc -gt $binaryTime
}
# Go embeds the prepared frontend and workbench assets. A Go-only timestamp
# check is not enough: changing frontend/dist or public/* otherwise leaves the
# local 4001 process serving an older page until someone manually rebuilds it.
if (-not $needsBuild) {
  $assetRoots = @(
    (Join-Path $repo 'frontend/dist'),
    (Join-Path $repo 'public/novel-panel'),
    (Join-Path $repo 'public/batch-rewrite'),
    (Join-Path $repo 'public/assets'),
    (Join-Path $repo 'pets')
  ) | Where-Object { Test-Path $_ }
  $latestAsset = $assetRoots |
    ForEach-Object { Get-ChildItem -LiteralPath $_ -Recurse -File -ErrorAction SilentlyContinue } |
    Sort-Object LastWriteTimeUtc -Descending |
    Select-Object -First 1
  $needsBuild = $latestAsset -and $latestAsset.LastWriteTimeUtc -gt $binaryTime
}
if ($needsBuild) {
  & (Join-Path $PSScriptRoot 'build-go-local.ps1')
}

Push-Location (Join-Path $repo 'backend')
try {
  $env:QIANTIE_GO_LISTEN_ADDR = $ListenAddr
  if (-not $env:QIANTIE_RELEASE_SHA) {
    $env:QIANTIE_RELEASE_SHA = (& git -C $repo rev-parse --short HEAD).Trim()
  }
  if (-not $env:QIANTIE_BATCH_FACTORY_V11_SLICE) {
    $env:QIANTIE_BATCH_FACTORY_V11_SLICE = '1'
  }
  if (-not $env:QIANTIE_LEGACY_ACCOUNTS_PATH) {
    $env:QIANTIE_LEGACY_ACCOUNTS_PATH = Join-Path $repo 'data/system/accounts.json'
  }
  if (-not $env:QIANTIE_GO_AUTH_ENABLED) {
    $env:QIANTIE_GO_AUTH_ENABLED = '1'
  }
  & .\qiantie-local.exe
  exit $LASTEXITCODE
} finally {
  Pop-Location
}
