param(
  [switch]$CheckOnly,
  [string]$ListenAddr = $(if ($env:QIANTIE_GO_LISTEN_ADDR) { $env:QIANTIE_GO_LISTEN_ADDR } else { ':4000' })
)

$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent $PSScriptRoot

if ($CheckOnly) {
  & (Join-Path $PSScriptRoot 'check-local-go.ps1')
  exit $LASTEXITCODE
}

if (-not (Test-Path (Join-Path $repo 'backend/qiantie-local.exe'))) {
  & (Join-Path $PSScriptRoot 'build-go-local.ps1')
}

Push-Location (Join-Path $repo 'backend')
try {
  $env:QIANTIE_GO_LISTEN_ADDR = $ListenAddr
  & .\qiantie-local.exe
  exit $LASTEXITCODE
} finally {
  Pop-Location
}
