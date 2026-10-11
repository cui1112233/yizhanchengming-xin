param(
  [int]$Port = $(if ($env:QIANTIE_NODE_PORT) { [int]$env:QIANTIE_NODE_PORT } else { 3001 })
)

$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent $PSScriptRoot

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
