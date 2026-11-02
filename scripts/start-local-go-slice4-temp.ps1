[CmdletBinding()]
param(
  [string]$EnvFile = '',
  [string]$ListenAddr = ':4001'
)

$ErrorActionPreference = 'Stop'
Write-Warning 'This compatibility launcher now delegates to the canonical Go review runtime.'
& (Join-Path $PSScriptRoot 'start-local-go-review.ps1') `
  -EnvFile $EnvFile `
  -ListenAddr $ListenAddr `
  -UseYcmMysql `
  -EnableBatchProduction `
  -EnableLocalMerge
exit $LASTEXITCODE
