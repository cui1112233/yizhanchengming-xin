$ErrorActionPreference = 'Stop'

$repoRoot = Split-Path -Parent $PSScriptRoot
$cargoTomlPath = Join-Path $repoRoot 'executors\doubao-relay\src-tauri\Cargo.toml'
$buildScriptPath = Join-Path $repoRoot 'scripts\build-yizhanchengming-executor.ps1'
$cargoToml = Get-Content -LiteralPath $cargoTomlPath -Raw
$buildScript = Get-Content -LiteralPath $buildScriptPath -Raw

if ($cargoToml -notmatch '(?ms)^\[features\]\s+default\s*=\s*\[\s*"custom-protocol"\s*\]\s+custom-protocol\s*=\s*\[\s*"tauri/custom-protocol"\s*\]') {
    throw 'doubao-relay release builds must enable the embedded frontend through the default custom-protocol feature'
}

if ($buildScript -notmatch '\$cargo\s+tauri\s+build\s+--bundles\s+nsis') {
    throw 'the distributable executor must be built through cargo tauri build'
}

Write-Output 'Executor release build contract passed.'
