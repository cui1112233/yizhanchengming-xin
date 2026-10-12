param(
  [int]$GoPort = 4002,
  [int]$PluginPort = 18181,
  [switch]$UseYcmMysql
)

$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent $PSScriptRoot
$runtimeDir = Join-Path $repo ("backend\.build-verify\giant-material-plugin-http\sandbox-{0}" -f $GoPort)
New-Item -ItemType Directory -Force -Path $runtimeDir | Out-Null

$pluginExe = Join-Path $runtimeDir 'giant-material-plugin.exe'
Push-Location (Join-Path $repo 'backend')
try {
  go build -o $pluginExe ./cmd/giant-material-plugin
} finally {
  Pop-Location
}

$tokenBytes = New-Object byte[] 32
[System.Security.Cryptography.RandomNumberGenerator]::Fill($tokenBytes)
$sharedToken = [Convert]::ToBase64String($tokenBytes)
$env:QIANTIE_GIANT_MATERIAL_PLUGIN_LISTEN_ADDR = "127.0.0.1:$PluginPort"
$env:QIANTIE_GIANT_MATERIAL_PLUGIN_TOKEN = $sharedToken
$env:QIANTIE_GIANT_MATERIAL_PLUGIN_URL = "http://127.0.0.1:$PluginPort"

Start-Process -FilePath $pluginExe -WindowStyle Hidden `
  -RedirectStandardOutput (Join-Path $runtimeDir 'plugin.out.log') `
  -RedirectStandardError (Join-Path $runtimeDir 'plugin.err.log') | Out-Null

& (Join-Path $PSScriptRoot 'start-local-go-review.ps1') -ListenAddr (":{0}" -f $GoPort) -UseYcmMysql:$UseYcmMysql
exit $LASTEXITCODE
