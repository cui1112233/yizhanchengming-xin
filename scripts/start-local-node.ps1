param(
  [int]$Port = $(if ($env:QIANTIE_NODE_PORT) { [int]$env:QIANTIE_NODE_PORT } else { 3001 }),
  [string]$EnvFile = ''
)

$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent $PSScriptRoot

if ([string]::IsNullOrWhiteSpace($EnvFile)) { $EnvFile = Join-Path $repo '.env.v88-review' }
if (Test-Path -LiteralPath $EnvFile) {
  $envText = [IO.File]::ReadAllText($EnvFile, [Text.Encoding]::UTF8)
  function Read-ReviewValue([string]$name) {
    $line = ($script:envText -split "`n") | Where-Object { $_.Trim().StartsWith("$name=") } | Select-Object -First 1
    if (-not $line) { return '' }
    return (($line -split '=', 2)[1]).Trim().Trim('"')
  }
  if (-not $env:QIANTIE_BRIDGE_SECRET) { $env:QIANTIE_BRIDGE_SECRET = Read-ReviewValue 'QIANTIE_V88_BRIDGE_SECRET' }
  if (-not $env:QIANTIE_121_CREDENTIAL_SECRET) { $env:QIANTIE_121_CREDENTIAL_SECRET = Read-ReviewValue 'QIANTIE_V88_121_CREDENTIAL_SECRET' }
  Write-Output "Using local Node config $EnvFile"
}

Push-Location $repo
try {
  $env:QIANTIE_NODE_PORT = [string]$Port
  if (-not $env:QIANTIE_GO_BASE_URL) {
    $env:QIANTIE_GO_BASE_URL = 'http://127.0.0.1:4001'
  }
  if (-not $env:QIANTIE_GO_AUTH_BASE_URL) {
    $env:QIANTIE_GO_AUTH_BASE_URL = $env:QIANTIE_GO_BASE_URL
  }
  Write-Output "Node 本地服务启动在 $Port，Go 鉴权地址：${env:QIANTIE_GO_AUTH_BASE_URL}"
  & node server.js
  exit $LASTEXITCODE
} finally {
  Pop-Location
}
