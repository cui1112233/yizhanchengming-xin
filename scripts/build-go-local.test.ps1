$ErrorActionPreference = 'Stop'

$script = Get-Content -LiteralPath (Join-Path $PSScriptRoot 'build-go-local.ps1') -Raw
foreach ($required in @(
  "Join-Path `$repo 'frontend/dist'",
  "Join-Path `$repo 'backend/web/dist'",
  "Join-Path `$repo 'backend/qiantie-local.exe'",
  "frontendEntry = `$entryMatch.Groups[1].Value",
  'binaryPath = $binaryPath',
  "Join-Path `$goDist 'batch-rewrite'",
  "Join-Path `$frontendDist 'batch-rewrite'"
)) {
  if (-not $script.Contains($required)) {
    throw "build-go-local.ps1 is missing required current-runtime contract: $required"
  }
}

Write-Output 'Local Go build contract passed.'
