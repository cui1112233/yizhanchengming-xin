$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent $PSScriptRoot
$script = Get-Content -LiteralPath (Join-Path $PSScriptRoot 'start-local-go-review.ps1') -Raw
$ffmpegDir = Join-Path $repo '.tools\ffmpeg'
$ttsProxySelector = Join-Path $PSScriptRoot 'select-local-tts-proxy.ps1'

if (-not (Test-Path -LiteralPath $ttsProxySelector)) { throw 'missing local TTS proxy selector' }
$listener = [Net.Sockets.TcpListener]::new([Net.IPAddress]::Loopback, 0)
try {
  $listener.Start()
  $proxyPort = ([Net.IPEndPoint]$listener.LocalEndpoint).Port
  $fallbackProxy = "http://127.0.0.1:$proxyPort"
  $selectedProxy = & $ttsProxySelector -ConfiguredProxy '' -FallbackProxy $fallbackProxy
  if ($selectedProxy -ne $fallbackProxy) { throw 'available local TTS proxy must be selected' }
} finally {
  $listener.Stop()
}
$selectedProxy = & $ttsProxySelector -ConfiguredProxy '' -FallbackProxy $fallbackProxy
if ($selectedProxy) { throw 'unavailable local TTS proxy must fall back to a direct connection' }
$configuredProxy = 'http://proxy.example:8181'
$selectedProxy = & $ttsProxySelector -ConfiguredProxy $configuredProxy -FallbackProxy $fallbackProxy
if ($selectedProxy -ne $configuredProxy) { throw 'explicit TTS proxy configuration must be preserved' }

if (-not $script.Contains('QIANTIE_LOCAL_FFMPEG_DIR')) { throw 'missing QIANTIE_LOCAL_FFMPEG_DIR support' }
if (-not $script.Contains('ffmpeg.exe')) { throw 'missing ffmpeg executable check' }
if (-not $script.Contains('$env:Path')) { throw 'missing PATH update' }
if ($script.Contains("QIANTIE_MYSQL_DSN = 'root:")) { throw 'must not hardcode MySQL DSN' }
if (-not $script.Contains("`$mysqlPort = 3307")) { throw 'missing published review MySQL port support' }
if (-not $script.Contains("Test-NetConnection -ComputerName '127.0.0.1' -Port 3307")) { throw 'review launcher must prefer the published local MySQL port' }
if (-not $script.Contains('[switch]$UseReviewMysql')) { throw 'review database must require an explicit switch' }
if (-not $script.Contains('Defaulting local runtime to the persisted YCM MySQL data store.')) { throw 'local runtime must default to the persisted YCM MySQL data store' }
if (-not $script.Contains('[switch]$EnableBatchProduction')) { throw 'missing explicit batch production switch' }
if (-not $script.Contains('[switch]$EnableLocalMerge')) { throw 'missing explicit local merge switch' }
if (-not $script.Contains('[string]$BinaryPath')) { throw 'review launcher must accept a candidate binary path' }
if (-not $script.Contains('[string]$ManifestPath')) { throw 'review launcher must accept a candidate runtime manifest path' }
if (-not $script.Contains('-BinaryPath $BinaryPath -ManifestPath $ManifestPath')) { throw 'review launcher must forward candidate paths to the base launcher' }
if (-not $script.Contains("QIANTIE_BATCH_FACTORY_V11_PRODUCTION_ENABLED = '1'")) { throw 'missing production enablement contract' }
if (-not $script.Contains("QIANTIE_BATCH_FACTORY_V11_MERGE_ENABLED = '1'")) { throw 'missing merge enablement contract' }
if (-not $script.Contains("QIANTIE_BATCH_FACTORY_V11_LOCAL_MERGE_ENABLED = '1'")) { throw 'missing local merge enablement contract' }
if (-not $script.Contains("QIANTIE_BATCH_FACTORY_V11_SLICE = '5'")) { throw 'missing local merge slice contract' }
if (-not $script.Contains('QIANTIE_V88_MERGE_TOS_ENDPOINT')) { throw 'missing shared TOS source config' }
if (-not $script.Contains('QIANTIE_NOVEL_FETCH_TOS_ENDPOINT')) { throw 'missing Novel Fetch TOS runtime config' }
if (-not $script.Contains('QIANTIE_AGENT_ASSET_')) { throw 'missing Agent asset TOS runtime config' }
if (-not $script.Contains("Read-ReviewValue 'QIANTIE_V88_N8_COMIC_OPEN_API_KEY'")) { throw 'missing local N8 comic key source config' }
if (-not $script.Contains('QIANTIE_N8_COMIC_OPEN_API_KEY')) { throw 'missing N8 comic runtime config' }
if (-not $script.Contains("Read-ReviewValue 'QIANTIE_V88_N8_COMIC_PRODUCER_ACCOUNTS'")) { throw 'missing local N8 producer account source config' }
if (-not $script.Contains('QIANTIE_N8_COMIC_PRODUCER_ACCOUNTS')) { throw 'missing N8 producer account runtime config' }
if (-not (Test-Path -LiteralPath (Join-Path $ffmpegDir 'ffmpeg.exe'))) { throw 'bundled ffmpeg missing' }
Write-Output 'Local review launcher FFmpeg and local merge contract passed.'
