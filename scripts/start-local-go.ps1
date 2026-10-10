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
