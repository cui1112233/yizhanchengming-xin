param(
    [switch]$Launch
)

$ErrorActionPreference = 'Stop'

$executorRoot = Split-Path -Parent $PSScriptRoot
$installer = Join-Path $executorRoot 'release\doubao-relay-standalone-installer.exe'
$hashFile = Join-Path $executorRoot 'release\doubao-relay-standalone-installer.sha256'
$appFolder = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('5LiA5oiY5pmf6ZOt5omn6KGM5Zmo'))
$installed = Join-Path $env:LOCALAPPDATA "$appFolder\doubao-relay.exe"

if (-not (Test-Path -LiteralPath $installer)) { throw "Standalone installer not found: $installer" }
if (-not (Test-Path -LiteralPath $hashFile)) { throw "Standalone installer checksum not found: $hashFile" }
if ($installer -match '\\target\\release\\doubao-relay\.exe$') {
    throw 'The raw cargo executable is not an installable standalone package.'
}
$expectedHash = (Get-Content -LiteralPath $hashFile -Raw).Trim()
$actualHash = (Get-FileHash -LiteralPath $installer -Algorithm SHA256).Hash
if ($actualHash -ne $expectedHash) { throw 'Standalone installer checksum mismatch.' }

Get-Process -Name 'doubao-relay' -ErrorAction SilentlyContinue | Stop-Process -Force
if (Test-Path -LiteralPath $installed) {
    $backup = "$installed.pre-standalone-$(Get-Date -Format 'yyyyMMddHHmmss').exe"
    Copy-Item -LiteralPath $installed -Destination $backup -ErrorAction Stop
    Write-Host "Backup: $backup"
}
Start-Process -FilePath $installer -ArgumentList '/S' -Wait

if (-not (Test-Path -LiteralPath $installed)) { throw "Installation did not produce: $installed" }
Write-Host "Installed standalone relay: $installed"
if ($Launch) {
    Start-Process -FilePath $installed
    Write-Host 'Started standalone relay.'
}
