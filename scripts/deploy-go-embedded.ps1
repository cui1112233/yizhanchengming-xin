param(
  [string]$OutputDir = '',
  [switch]$SkipFrontendBuild,
  [switch]$RequireCleanFrontend,
  [switch]$Deploy,
  [string]$TargetHost = '',
  [string]$TargetUser = 'deploy',
  [int]$SshPort = 22,
  [string]$TargetPath = '/opt/qiantie/releases/v88/qiantie-server',
  [string]$ServiceName = 'qiantie-v88',
  [switch]$KeepBuildDirectory
)

$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent $PSScriptRoot
$frontendDir = Join-Path $repo 'frontend'
$frontendDist = Join-Path $frontendDir 'dist'
$buildScript = Join-Path $PSScriptRoot 'build-go-local.ps1'
$backendDir = Join-Path $repo 'backend'
$releaseRoot = if ([string]::IsNullOrWhiteSpace($OutputDir)) { Join-Path $repo '.release/go-embedded' } else { $OutputDir }
$releaseRoot = [System.IO.Path]::GetFullPath($releaseRoot)

if ($Deploy -and [string]::IsNullOrWhiteSpace($TargetHost)) {
  throw '启用 -Deploy 时必须同时提供 -TargetHost；不提供目标主机时只生成本地发布包。'
}
if (-not $Deploy -and -not [string]::IsNullOrWhiteSpace($TargetHost)) {
  throw '已提供 -TargetHost，但没有启用 -Deploy。为避免误发布，请同时显式传入 -Deploy。'
}
if (-not (Test-Path -LiteralPath $buildScript)) {
  throw "构建脚本不存在：$buildScript"
}

function Invoke-Checked {
  param(
    [string]$Command,
    [string[]]$Arguments
  )
  & $Command @Arguments
  if ($LASTEXITCODE -ne 0) {
    throw "命令执行失败（$LASTEXITCODE）：$Command $($Arguments -join ' ')"
  }
}

# 前端产物来源校验：发布包必须能说明前端源码是否干净、嵌入的是哪个构建。
# 默认只记录状态；显式传入 -RequireCleanFrontend 时，源码有未提交改动或
# dist 落后于源码都会直接拒绝构建，避免把其他会话的半成品页面带上线。
$frontendSourceChanges = @(& git -C $repo status --porcelain -- frontend/src)
if ($LASTEXITCODE -ne 0) {
  throw '无法读取 frontend/src 的 Git 状态，停止生成来源不明的发布包。'
}
$frontendSourceState = if ($frontendSourceChanges.Count -eq 0) { 'clean' } else { 'dirty' }
if ($RequireCleanFrontend -and $frontendSourceState -ne 'clean') {
  $preview = ($frontendSourceChanges | Select-Object -First 8) -join '; '
  throw "frontend/src 存在未提交改动（$($frontendSourceChanges.Count) 项）：$preview。请先提交前端源码，或去掉 -RequireCleanFrontend。"
}

if (-not $SkipFrontendBuild) {
  Write-Host '1/5 构建 React 前端源码...'
  Invoke-Checked 'npm' @('--prefix', $frontendDir, 'run', 'build')
}

if (-not (Test-Path -LiteralPath (Join-Path $frontendDist 'index.html'))) {
  throw "前端构建结果不存在：$frontendDist/index.html。请先构建 frontend，或去掉 -SkipFrontendBuild。"
}

$indexPath = Join-Path $frontendDist 'index.html'
$indexHtml = Get-Content -LiteralPath $indexPath -Raw
$entryMatch = [regex]::Match($indexHtml, 'src="(/assets/[^"]+\.js)"')
$frontendEntry = if ($entryMatch.Success) { $entryMatch.Groups[1].Value } else { '' }
if ($RequireCleanFrontend -and [string]::IsNullOrWhiteSpace($frontendEntry)) {
  throw "无法从 $indexPath 识别入口脚本，停止发布无法校验的前端产物。"
}
if ($RequireCleanFrontend -and $SkipFrontendBuild) {
  $newestSource = Get-ChildItem -LiteralPath (Join-Path $frontendDir 'src') -Recurse -File -Include *.ts,*.tsx,*.css,*.json |
    Sort-Object LastWriteTimeUtc -Descending |
    Select-Object -First 1
  $indexTime = (Get-Item -LiteralPath $indexPath).LastWriteTimeUtc
  if ($null -ne $newestSource -and $newestSource.LastWriteTimeUtc -gt $indexTime) {
    throw "frontend/dist 落后于源码（$($newestSource.FullName) 晚于 index.html），请重新构建前端或去掉 -SkipFrontendBuild。"
  }
}

Write-Host '2/5 同步前端产物到 Go embed 目录...'
Invoke-Checked 'powershell' @('-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', $buildScript, '-SkipGoBuild')

$gitSha = (& git -C $repo rev-parse HEAD).Trim()
if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($gitSha)) {
  throw '无法读取当前 Git SHA，停止生成不可追踪的发布包。'
}
$shortSha = $gitSha.Substring(0, [Math]::Min(12, $gitSha.Length))
$timestamp = Get-Date -Format 'yyyyMMdd-HHmmss'
$releaseDir = Join-Path $releaseRoot "$timestamp-$shortSha"
New-Item -ItemType Directory -Force $releaseDir | Out-Null
$binaryPath = Join-Path $releaseDir 'qiantie-server-linux-amd64'

Write-Host '3/5 编译 Linux amd64 Go 二进制...'
$oldGoos = $env:GOOS
$oldGoarch = $env:GOARCH
$oldCgo = $env:CGO_ENABLED
try {
  $env:GOOS = 'linux'
  $env:GOARCH = 'amd64'
  $env:CGO_ENABLED = '0'
  Push-Location $backendDir
  try {
    Invoke-Checked 'go' @('build', '-trimpath', '-ldflags=-s -w', '-o', $binaryPath, './cmd/qiantie')
  } finally {
    Pop-Location
  }
} finally {
  $env:GOOS = $oldGoos
  $env:GOARCH = $oldGoarch
  $env:CGO_ENABLED = $oldCgo
}

$hash = (Get-FileHash -LiteralPath $binaryPath -Algorithm SHA256).Hash.ToLowerInvariant()
$metadata = [ordered]@{
  git_sha = $gitSha
  binary = 'qiantie-server-linux-amd64'
  sha256 = $hash
  built_at = (Get-Date).ToUniversalTime().ToString('o')
  frontend_source = 'frontend/src'
  frontend_source_state = $frontendSourceState
  frontend_entry = $frontendEntry
  frontend_index_sha256 = (Get-FileHash -LiteralPath $indexPath -Algorithm SHA256).Hash.ToLowerInvariant()
  frontend_built_at = (Get-Item -LiteralPath $indexPath).LastWriteTimeUtc.ToString('o')
  embedded_assets = 'backend/web/dist'
  service = $ServiceName
}
$metadata | ConvertTo-Json | Set-Content -LiteralPath (Join-Path $releaseDir 'release.json') -Encoding utf8
Copy-Item -LiteralPath $binaryPath -Destination (Join-Path $releaseRoot 'qiantie-server-linux-amd64') -Force
Copy-Item -LiteralPath (Join-Path $releaseDir 'release.json') -Destination (Join-Path $releaseRoot 'release.json') -Force

Write-Host "4/5 本地发布包已生成：$releaseDir"
Write-Host "Git SHA：$gitSha"
Write-Host "SHA-256：$hash"
Write-Host "前端源码状态：$frontendSourceState；入口：$frontendEntry"

if ($Deploy) {
  Write-Host '5/5 上传并原子替换远端二进制...'
  $remote = "$TargetUser@$TargetHost"
  $remoteTemp = "/tmp/qiantie-v88-$timestamp-$shortSha"
  Invoke-Checked 'scp' @('-P', "$SshPort", $binaryPath, "${remote}:$remoteTemp")
  $remoteCommand = "set -eu; sudo install -D -m 0755 '$remoteTemp' '$TargetPath'; rm -f '$remoteTemp'; sudo systemctl restart '$ServiceName'; sudo systemctl is-active --quiet '$ServiceName'"
  Invoke-Checked 'ssh' @('-p', "$SshPort", $remote, $remoteCommand)
  Write-Host "远端服务已重启并通过 systemd 健康检查：$ServiceName"
} else {
  Write-Host '5/5 未执行远端部署：如需上传，请明确传入 -Deploy -TargetHost。'
}

if (-not $KeepBuildDirectory) {
  Get-ChildItem -LiteralPath $releaseRoot -Directory |
    Where-Object { $_.FullName -ne $releaseDir } |
    Sort-Object LastWriteTime -Descending |
    Select-Object -Skip 2 |
    Remove-Item -Recurse -Force
}
