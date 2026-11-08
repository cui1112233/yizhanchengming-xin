param(
  [string]$EnvFile = '',
  [string]$ListenAddr = ':4001',
  [Alias('UseYcmMysql')]
  [switch]$UseLocalMysql,
  [switch]$UseReviewMysql,
  [switch]$EnableBatchProduction,
  [switch]$EnableLocalMerge,
  [switch]$SkipBuild,
  [switch]$Detached,
  [string]$BinaryPath = '',
  [string]$ManifestPath = ''
)

$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent $PSScriptRoot
if ([string]::IsNullOrWhiteSpace($EnvFile)) {
  $EnvFile = Join-Path $repo '.env.v88-review'
}
if (-not (Test-Path -LiteralPath $EnvFile)) {
  throw "找不到本地 review 配置文件：$EnvFile"
}
Write-Output "Using review config $EnvFile"

function Read-ReviewValue([string]$name) {
  $line = ([IO.File]::ReadAllText($EnvFile, [Text.Encoding]::UTF8) -split "`n") |
    Where-Object { $_.Trim().StartsWith("$name=") } | Select-Object -First 1
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

$mysqlHost = '127.0.0.1'
$mysqlPort = 3306
$mysqlUser = 'root'
$mysqlDatabase = 'qiantie_go_local'
$mysqlPassword = (& docker inspect ycm-mysql --format '{{range .Config.Env}}{{println .}}{{end}}' 2>$null | Where-Object { $_ -match '^MYSQL_ROOT_PASSWORD=' } | Select-Object -First 1) -replace '^MYSQL_ROOT_PASSWORD=', ''
if ([string]::IsNullOrWhiteSpace($mysqlPassword)) { throw '无法读取 ycm-mysql 的数据库密码' }
Write-Output 'Defaulting local runtime to the persisted YCM MySQL data store.'

if ($UseReviewMysql) {
  $mysqlHost = (& docker inspect qiantie-v88-review-mysql-1 --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' 2>$null).Trim()
  $mysqlPort = 3306
  $mysqlUser = 'qiantie'
  $mysqlDatabase = 'qiantie_v88'
  $mysqlPassword = $password
  if (Test-NetConnection -ComputerName '127.0.0.1' -Port 3307 -InformationLevel Quiet) {
    # Docker Desktop publishes the review MySQL here. Prefer it over the
    # container-only address, which can accept TCP but time out during startup.
    $mysqlHost = '127.0.0.1'
    $mysqlPort = 3307
  }
  if ([string]::IsNullOrWhiteSpace($mysqlHost)) { $mysqlHost = '127.0.0.1' }
  Write-Output 'Using explicitly requested isolated review MySQL data store.'
}
Write-Output "Using MySQL $mysqlUser@$mysqlHost`:$mysqlPort/$mysqlDatabase"
$env:QIANTIE_MYSQL_DSN = "${mysqlUser}:$mysqlPassword@tcp(${mysqlHost}:$mysqlPort)/${mysqlDatabase}?parseTime=true&multiStatements=true"
$env:QIANTIE_BRIDGE_SECRET = $bridgeSecret
$env:QIANTIE_GO_LISTEN_ADDR = $ListenAddr
# 出图并发以“统一配置”的 imageConcurrency 为准；未配置时回落到代码默认（2 路）。
# 这个环境变量只作为兜底覆盖，本地验证不要写死，否则会和页面设置不一致。
Write-Output "Shuihuo image generation concurrency fallback: $($env:QIANTIE_SHUHUO_IMAGE_CONCURRENCY)"
$relayBaseURL = Read-ReviewValue 'QIANTIE_DOUBAO_RELAY_BASE_URL'
$relayBearerToken = Read-ReviewValue 'QIANTIE_DOUBAO_RELAY_BEARER_TOKEN'
if (-not [string]::IsNullOrWhiteSpace($relayBaseURL) -and -not [string]::IsNullOrWhiteSpace($relayBearerToken)) {
  $env:QIANTIE_DOUBAO_RELAY_BASE_URL = $relayBaseURL
  $env:QIANTIE_DOUBAO_RELAY_BEARER_TOKEN = $relayBearerToken
  Remove-Item Env:QIANTIE_DOUBAO_RELAY_FIXED_ACCOUNT -ErrorAction SilentlyContinue
  Write-Output "Doubao relay image path enabled at $relayBaseURL."
} else {
  Remove-Item Env:QIANTIE_DOUBAO_RELAY_BASE_URL -ErrorAction SilentlyContinue
  Remove-Item Env:QIANTIE_DOUBAO_RELAY_BEARER_TOKEN -ErrorAction SilentlyContinue
  Remove-Item Env:QIANTIE_DOUBAO_RELAY_FIXED_ACCOUNT -ErrorAction SilentlyContinue
  Write-Output 'Doubao relay image path is not configured.'
}
$n8ComicOpenAPIKey = Read-ReviewValue 'QIANTIE_V88_N8_COMIC_OPEN_API_KEY'
$n8ComicProducerAccounts = Read-ReviewValue 'QIANTIE_V88_N8_COMIC_PRODUCER_ACCOUNTS'
$n8AdminToken = Read-ReviewValue 'QIANTIE_V88_N8_ADMIN_TOKEN'
$env:QIANTIE_N8_COMIC_PRODUCER_ACCOUNTS = $n8ComicProducerAccounts
$env:QINGYU_N8_ADMIN_TOKEN = $n8AdminToken
if ([string]::IsNullOrWhiteSpace($n8ComicOpenAPIKey)) {
  $n8ComicOpenAPIKey = $env:QIANTIE_N8_COMIC_OPEN_API_KEY
}
if ([string]::IsNullOrWhiteSpace($n8ComicOpenAPIKey)) {
  Remove-Item Env:QIANTIE_N8_COMIC_OPEN_API_KEY -ErrorAction SilentlyContinue
  Write-Output 'N8 comic publishing is not configured.'
} else {
  $env:QIANTIE_N8_COMIC_OPEN_API_KEY = $n8ComicOpenAPIKey
  Write-Output 'N8 comic publishing key loaded.'
}
# Keep the normal local review runtime read-only for video submission. When
# the caller explicitly requests it, slice 4 uses the saved account video
# model configuration (for example MiniMax H3) rather than a launcher key.
if ($EnableLocalMerge -and -not $EnableBatchProduction) {
  throw 'EnableLocalMerge requires EnableBatchProduction.'
}
if ($EnableLocalMerge) {
  $env:QIANTIE_BATCH_FACTORY_V11_SLICE = '5'
  $env:QIANTIE_BATCH_FACTORY_V11_PRODUCTION_ENABLED = '1'
  $env:QIANTIE_BATCH_FACTORY_V11_MERGE_ENABLED = '1'
  $env:QIANTIE_BATCH_FACTORY_V11_LOCAL_MERGE_ENABLED = '1'
  Write-Output 'Batch Factory local video production and merge are explicitly enabled.'
} elseif ($EnableBatchProduction) {
  $env:QIANTIE_BATCH_FACTORY_V11_SLICE = '4'
  $env:QIANTIE_BATCH_FACTORY_V11_PRODUCTION_ENABLED = '1'
  Remove-Item Env:QIANTIE_BATCH_FACTORY_V11_MERGE_ENABLED -ErrorAction SilentlyContinue
  Remove-Item Env:QIANTIE_BATCH_FACTORY_V11_LOCAL_MERGE_ENABLED -ErrorAction SilentlyContinue
  Write-Output 'Batch Factory video production is explicitly enabled.'
} else {
  # 与线上运行时对齐：默认 Slice 5（能力 13 项，仅 publish.yadi 关闭）。
  $env:QIANTIE_BATCH_FACTORY_V11_SLICE = '5'
  $env:QIANTIE_BATCH_FACTORY_V11_PRODUCTION_ENABLED = '1'
  $env:QIANTIE_BATCH_FACTORY_V11_MERGE_ENABLED = '1'
  $env:QIANTIE_BATCH_FACTORY_V11_LOCAL_MERGE_ENABLED = '1'
  Write-Output 'Batch Factory slice 5 (production + merge) enabled to match production runtime.'
}
$env:QIANTIE_BATCH_FACTORY_V11_DYNAMIC_TEXT_PROVIDER = '1'
# No Node/3001 fallback. Migrated routes are served directly by Go.
$env:QIANTIE_REDIS_URL = 'redis://127.0.0.1:6379'
# Novel Fetch and merge share the same TOS account configuration. The Go
# runtime keeps separate object-key prefixes, so credentials stay centralized
# without mixing novel text and merged video objects.
$tosReviewMapping = @{
  'QIANTIE_V88_MERGE_TOS_ENDPOINT'   = 'QIANTIE_NOVEL_FETCH_TOS_ENDPOINT'
  'QIANTIE_V88_MERGE_TOS_REGION'     = 'QIANTIE_NOVEL_FETCH_TOS_REGION'
  'QIANTIE_V88_MERGE_TOS_BUCKET'     = 'QIANTIE_NOVEL_FETCH_TOS_BUCKET'
  'QIANTIE_V88_MERGE_TOS_ACCESS_KEY' = 'QIANTIE_NOVEL_FETCH_TOS_ACCESS_KEY'
  'QIANTIE_V88_MERGE_TOS_SECRET_KEY' = 'QIANTIE_NOVEL_FETCH_TOS_SECRET_KEY'
}
foreach ($sourceName in $tosReviewMapping.Keys) {
  $value = Read-ReviewValue $sourceName
  if (-not [string]::IsNullOrWhiteSpace($value)) {
    [Environment]::SetEnvironmentVariable($tosReviewMapping[$sourceName], $value, 'Process')
    $agentAssetName = $tosReviewMapping[$sourceName] -replace '^QIANTIE_NOVEL_FETCH_', 'QIANTIE_AGENT_ASSET_'
    [Environment]::SetEnvironmentVariable($agentAssetName, $value, 'Process')
  }
}
# Prefer the local Clash proxy when it is running. If it is unavailable and
# no proxy was explicitly configured, connect to the public TTS service directly.
$ttsProxySelector = Join-Path $PSScriptRoot 'select-local-tts-proxy.ps1'
$selectedTTSProxy = & $ttsProxySelector -ConfiguredProxy $env:QIANTIE_TTS_PROXY
if ($selectedTTSProxy) {
  $env:QIANTIE_TTS_PROXY = $selectedTTSProxy
  Write-Output "Using TTS proxy $env:QIANTIE_TTS_PROXY"
} else {
  Remove-Item Env:QIANTIE_TTS_PROXY -ErrorAction SilentlyContinue
  Write-Output 'Using direct connection for TTS'
}
if ($env:QIANTIE_VIDEO_CONNECTOR_PROXY) {
  Write-Output "Using configured video management proxy $env:QIANTIE_VIDEO_CONNECTOR_PROXY"
} else {
  Write-Output 'Using direct HTTPS connection for video management'
}

$ffmpegDir = if ([string]::IsNullOrWhiteSpace($env:QIANTIE_LOCAL_FFMPEG_DIR)) { Join-Path $repo '.tools\ffmpeg' } else { $env:QIANTIE_LOCAL_FFMPEG_DIR }
$ffmpegPath = Join-Path $ffmpegDir 'ffmpeg.exe'
if (-not (Test-Path -LiteralPath $ffmpegPath)) {
  throw ('FFmpeg not found: {0}. Set QIANTIE_LOCAL_FFMPEG_DIR or install FFmpeg.' -f $ffmpegPath)
}
$env:Path = $ffmpegDir + [IO.Path]::PathSeparator + $env:Path
Write-Output ('Using FFmpeg {0}' -f $ffmpegPath)

& (Join-Path $PSScriptRoot 'start-local-go.ps1') -ListenAddr $ListenAddr -SkipBuild:$SkipBuild -Detached:$Detached -BinaryPath $BinaryPath -ManifestPath $ManifestPath
exit $LASTEXITCODE
