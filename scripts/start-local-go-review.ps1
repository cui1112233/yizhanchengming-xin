param(
  [string]$EnvFile = "$(Join-Path (Split-Path -Parent $PSScriptRoot) '.env.v88-review')",
  [string]$ListenAddr = ':4001',
  [switch]$UseYcmMysql
)

$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent $PSScriptRoot
if (-not (Test-Path -LiteralPath $EnvFile)) {
  throw "找不到本地 review 配置文件：$EnvFile"
}

function Read-ReviewValue([string]$name) {
  $line = Get-Content -LiteralPath $EnvFile | Where-Object { $_ -match "^\s*$name\s*=" } | Select-Object -First 1
  if (-not $line) { return '' }
  $value = ($line -split '=', 2)[1].Trim()
  if ($value.Length -ge 2 -and $value.StartsWith('"') -and $value.EndsWith('"')) {
    $value = $value.Substring(1, $value.Length - 2)
  }
  return $value
}

$password = Read-ReviewValue 'QIANTIE_V88_MYSQL_PASSWORD'
$bridgeSecret = Read-ReviewValue 'QIANTIE_V88_BRIDGE_SECRET'
if ([string]::IsNullOrWhiteSpace($password) -or [string]::IsNullOrWhiteSpace($bridgeSecret)) {
  throw 'review 配置缺少数据库密码或桥接密钥'
}
Write-Output "Review credentials loaded (password length $($password.Length), bridge length $($bridgeSecret.Length))"

$mysqlHost = (& docker inspect qiantie-v88-review-mysql-1 --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' 2>$null).Trim()
$mysqlUser = 'qiantie'
$mysqlDatabase = 'qiantie_v88'
$mysqlPassword = $password
if ($UseYcmMysql) {
  $mysqlHost = '127.0.0.1'
  $mysqlUser = 'root'
  $mysqlDatabase = 'qiantie_go_local'
  $mysqlPassword = (& docker inspect ycm-mysql --format '{{range .Config.Env}}{{println .}}{{end}}' 2>$null | Where-Object { $_ -match '^MYSQL_ROOT_PASSWORD=' } | Select-Object -First 1) -replace '^MYSQL_ROOT_PASSWORD=', ''
  if ([string]::IsNullOrWhiteSpace($mysqlPassword)) { throw '无法读取 ycm-mysql 的数据库密码' }
}
if ([string]::IsNullOrWhiteSpace($mysqlHost)) { $mysqlHost = '127.0.0.1' }
Write-Output "Using MySQL $mysqlUser@$mysqlHost/$mysqlDatabase"
$env:QIANTIE_MYSQL_DSN = "${mysqlUser}:$mysqlPassword@tcp(${mysqlHost}:3306)/${mysqlDatabase}?parseTime=true&multiStatements=true"
$env:QIANTIE_BRIDGE_SECRET = $bridgeSecret
$env:QIANTIE_GO_LISTEN_ADDR = $ListenAddr
# Review compose files may use a higher production slice. The local Go
# browser runtime intentionally starts at slice 3 until video/merge providers
# are configured and verified.
$env:QIANTIE_BATCH_FACTORY_V11_SLICE = '3'
$env:QIANTIE_BATCH_FACTORY_V11_DYNAMIC_TEXT_PROVIDER = '1'
$env:QIANTIE_NODE_AUTH_BASE_URL = 'http://127.0.0.1:3001'
$env:QIANTIE_REDIS_URL = 'redis://127.0.0.1:6379'

& (Join-Path $PSScriptRoot 'start-local-go.ps1') -ListenAddr $ListenAddr
exit $LASTEXITCODE
