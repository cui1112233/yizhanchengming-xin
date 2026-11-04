# 打包媒体引擎（ffmpeg/ffprobe）目录，供部署时整体放到 <tools>/ffmpeg/。
# 只复制本地已有文件，不联网下载任何二进制。
param(
  [Parameter(Mandatory = $true)][string]$SourceDir,
  [string]$OutputDir = ''
)

$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent $PSScriptRoot
$outputRoot = $OutputDir
if ([string]::IsNullOrWhiteSpace($outputRoot)) {
  $outputRoot = Join-Path $repo '.release/media-engine'
}
$outputRoot = [System.IO.Path]::GetFullPath($outputRoot)
$target = Join-Path $outputRoot 'ffmpeg'

$sourceRoot = (Resolve-Path -LiteralPath $SourceDir).Path
$searchRoots = @($sourceRoot)
foreach ($dir in (Get-ChildItem -LiteralPath $sourceRoot -Directory -ErrorAction SilentlyContinue)) {
  $searchRoots += $dir.FullName
}

$binaries = @{}
foreach ($name in @('ffmpeg', 'ffprobe')) {
  $found = $null
  foreach ($root in $searchRoots) {
    foreach ($sub in @('', 'bin')) {
      foreach ($ext in @('', '.exe')) {
        $candidate = Join-Path (Join-Path $root $sub) ($name + $ext)
        if (-not $found -and (Test-Path -LiteralPath $candidate)) {
          $found = (Resolve-Path -LiteralPath $candidate).Path
        }
      }
    }
  }
  $binaries[$name] = $found
}

if (-not $binaries['ffmpeg'] -or -not $binaries['ffprobe']) {
  throw "源目录中缺少 ffmpeg 或 ffprobe：$sourceRoot。请确认目标平台二进制（线上为 linux-amd64）。"
}

New-Item -ItemType Directory -Force $target | Out-Null
foreach ($name in @('ffmpeg', 'ffprobe')) {
  $source = $binaries[$name]
  Copy-Item -LiteralPath $source -Destination (Join-Path $target (Split-Path -Leaf $source)) -Force
}

$licenseFiles = @()
foreach ($root in $searchRoots) {
  $licenseFiles += @(Get-ChildItem -LiteralPath $root -File -ErrorAction SilentlyContinue |
    Where-Object { $_.Name -match '^(LICENSE|COPYING|README)' })
}
$licenseSeen = @{}
foreach ($file in $licenseFiles) {
  if (-not $licenseSeen.ContainsKey($file.Name)) {
    $licenseSeen[$file.Name] = $true
    Copy-Item -LiteralPath $file.FullName -Destination (Join-Path $target $file.Name) -Force
  }
}
if ($licenseFiles.Count -eq 0) {
  Write-Warning '未找到 LICENSE/README，随包分发 ffmpeg 前必须补齐许可证文件与源码获取说明。'
}

Write-Host "媒体引擎目录已生成：$target"
Get-ChildItem -LiteralPath $target | Select-Object Name, Length
Write-Host ''
Write-Host '部署时把整个 ffmpeg 目录放到服务端 <tools>/ffmpeg/，例如：'
Write-Host ('  scp -r ' + $target + ' root@<host>:/opt/www/qiantie/tools/')
Write-Host '启动自检会打印识别到的路径与版本；可用 QIANTIE_TOOLS_DIR 覆盖 tools 根目录。'
