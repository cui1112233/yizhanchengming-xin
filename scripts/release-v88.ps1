# V88 一体化发布脚本：存档 -> 构建 -> 数据打包 -> 上传 -> 远程校验 -> 失败自动回滚。
#
# 设计目标（对应 2026-11-04 事故）：
#   1. 画风库只写进被 .gitignore 忽略的 data/system/presets.json，旧发布只传二进制
#      -> 本脚本按白名单把 presets.json 与被引用的预览图一起打包上传，并在上传后
#      比对远程文件字节数与 SHA256。
#   2. 工作区长期存在多路未提交改动 -> 本脚本强制执行 docs/releases/<buildId>/ 存档，
#      输出“本版 = HEAD <sha> + 工作区 diff”，并在 build-info 里记录二进制 SHA256。
#   3. 4001 无法回答“跑的是哪版” -> 本脚本写 backend/current-release.json，
#      start-local-go.ps1 可用 -RequireArchivedRelease 校验“只运行有存档的构建”。
#
# 红线：
#   - 不 commit / 不 add / 不 stash；工作区始终保持原样。
#   - 不改业务代码；本脚本只写 docs/releases、.release、backend/current-release.json。
#   - data/users/**/api-config.json、upload-target.json 及任何 cookie/token/password
#     文件绝不允许进入发布包；压缩包内一旦出现这些文件名立即报错退出。
#   - 不重启本机 4001；本地二进制只构建到 .release 目录，除非显式 -UpdateLocalBinary。

[CmdletBinding()]
param(
  [switch]$Deploy,
  [switch]$Rollback,
  [switch]$ListBackups,
  [switch]$DryRun,
  [switch]$Force,
  [string]$BuildId = '',
  # 用户可见更新说明按 releaseId 索引（internal/releasenotes/notes/<releaseId>.json）。
  # 必须在构建前确定：默认取打包时刻（CST）的 qiantie-YYYYMMDDHHmm。
  [string]$ReleaseId = '',
  # 本地前置验证：候选二进制先在本地独立端口启动并验收，任一失败即中止发布。
  [int]$LocalVerifyPort = 4011,
  [int]$LocalVerifyTimeoutSeconds = 25,

  # 远端连接参数沿用 scripts/deploy-go-embedded.ps1 的默认值。
  [string]$TargetHost = '',
  [string]$TargetUser = 'deploy',
  [int]$SshPort = 22,
  [string]$TargetPath = '/opt/qiantie/releases/v88/qiantie-server',
  [string]$ServiceName = 'qiantie-v88',
  # 生产 data 根目录（presets 为 <RemoteDataDir>/system/presets.json）。仓库内找不到
  # 可信出处的默认值，因此部署数据时必须显式提供，否则拒绝上传。
  [string]$RemoteDataDir = '',
  [string]$RemoteProbeBaseUrl = 'http://127.0.0.1:4000',
  [string]$LocalApiBaseUrl = 'http://127.0.0.1:4001',
  [string]$ApiToken = $(if ($env:V88_RELEASE_API_TOKEN) { $env:V88_RELEASE_API_TOKEN } else { '' }),
  [string]$EnvFile = '',
  [string[]]$RequiredEnvVars = @(
    'QIANTIE_MYSQL_DSN',
    'QIANTIE_BRIDGE_SECRET',
    'QIANTIE_BATCH_FACTORY_V11_SLICE',
    'QINGYU_N8_ADMIN_TOKEN'
  ),
  [switch]$AllowMissingEnv,
  [switch]$SkipFrontendBuild,
  # 指定已构建好的前端产物目录（例如 frontend/dist 的临时快照）。指定后跳过 npm
  # build，并直接从该目录同步到 backend/web/dist，避免实时目录被并行会话覆盖。
  [string]$FrontendDist = '',
  [switch]$SkipBuild,
  [switch]$SkipDataUpload,
  [switch]$UpdateLocalBinary,
  [switch]$KeepOldReleases,
  [string]$BackupId = ''
)

$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent $PSScriptRoot
$frontendDir = Join-Path $repo 'frontend'
$frontendDist = Join-Path $frontendDir 'dist'
$backendDir = Join-Path $repo 'backend'
$archivesRoot = Join-Path $repo 'docs/releases'
$releaseRoot = Join-Path $repo '.release/v88'
$localBinaryPath = Join-Path $backendDir 'qiantie-local.exe'
$currentReleasePath = Join-Path $backendDir 'current-release.json'

$useSnapshotDist = $null
if (-not [string]::IsNullOrWhiteSpace($FrontendDist)) {
  $useSnapshotDist = (Resolve-Path -LiteralPath $FrontendDist).Path
  $frontendDist = $useSnapshotDist
}

function Write-Section([string]$Text) {
  Write-Host ''
  Write-Host "=== $Text ===" -ForegroundColor Cyan
}

function Write-Utf8NoBom {
  param([string]$Path, [string]$Content)
  $dir = Split-Path -Parent $Path
  if ($dir -and -not (Test-Path -LiteralPath $dir)) { New-Item -ItemType Directory -Force $dir | Out-Null }
  [System.IO.File]::WriteAllText($Path, $Content, (New-Object System.Text.UTF8Encoding($false)))
}

function Write-JsonNoBom {
  param([string]$Path, [object]$Content)
  Write-Utf8NoBom -Path $Path -Content (($Content | ConvertTo-Json -Depth 8) + "`n")
}

function Invoke-Checked {
  param([string]$Command, [string[]]$Arguments, [string]$WorkingDirectory = '')
  $previous = $null
  if ($WorkingDirectory) {
    $previous = Get-Location
    Push-Location -LiteralPath $WorkingDirectory
  }
  try {
    & $Command @Arguments
    if ($LASTEXITCODE -ne 0) {
      throw "命令执行失败（exit=$LASTEXITCODE）：$Command $($Arguments -join ' ')"
    }
  } finally {
    if ($previous) { Pop-Location }
  }
}

function Get-FileSha256([string]$Path) {
  return (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash.ToLowerInvariant()
}

function Get-RelativePath([string]$FullPath) {
  $base = [System.IO.Path]::GetFullPath($repo).TrimEnd('\', '/')
  $full = [System.IO.Path]::GetFullPath($FullPath)
  if ($full.StartsWith($base, [StringComparison]::OrdinalIgnoreCase)) {
    return ($full.Substring($base.Length).TrimStart('\', '/') -replace '\\', '/')
  }
  return $full
}

function Test-SensitiveName([string]$RelativePath) {
  $name = Split-Path -Leaf $RelativePath
  if ($name -match '(?i)^(api-config\.json|upload-target\.json)$') { return $true }
  if ($name -match '(?i)(cookie|token|password|passwd|secret|credential)') { return $true }
  if ($name -match '(?i)^\.?env(\.|$)') { return $true }
  return $false
}

function Test-BinaryLocked([string]$Path) {
  if (-not (Test-Path -LiteralPath $Path)) { return $false }
  try {
    $stream = [System.IO.File]::Open($Path, [System.IO.FileMode]::Open, [System.IO.FileAccess]::ReadWrite, [System.IO.FileShare]::None)
    $stream.Close()
    return $false
  } catch {
    return $true
  }
}

function Assert-RemoteParameters {
  if ([string]::IsNullOrWhiteSpace($TargetHost)) {
    throw '该模式必须提供 -TargetHost。'
  }
}

function Invoke-Remote {
  param([string]$Command)
  $remote = "$TargetUser@$TargetHost"
  $output = & ssh -p "$SshPort" $remote $Command 2>&1 | Out-String
  if ($LASTEXITCODE -ne 0) {
    throw "远程命令失败（exit=$LASTEXITCODE）：$Command`n$output"
  }
  return $output.Trim()
}

function Get-RemoteBackupRoot {
  $normalized = $TargetPath.TrimEnd('/')
  $index = $normalized.LastIndexOf('/')
  if ($index -lt 0) { throw "无法从 TargetPath 推导远端目录：$TargetPath" }
  return $normalized.Substring(0, $index) + '/.rollback'
}

function Get-RemoteHttpStatus([string]$Path) {
  $command = "curl -s -o /dev/null -w '%{http_code}' '$RemoteProbeBaseUrl$Path'"
  return Invoke-Remote -Command $command
}

function Get-RemoteJson([string]$Path, [string]$Token) {
  if ([string]::IsNullOrWhiteSpace($Token)) {
    return $null
  }
  $command = "curl -s -H 'Authorization: Bearer $Token' '$RemoteProbeBaseUrl$Path'"
  $raw = Invoke-Remote -Command $command
  try { return ($raw | ConvertFrom-Json) } catch { return $null }
}

function Get-LocalJson([string]$Path, [string]$Token) {
  if ([string]::IsNullOrWhiteSpace($Token)) {
    return $null
  }
  try {
    return Invoke-RestMethod -Uri "$LocalApiBaseUrl$Path" -Headers @{ Authorization = "Bearer $Token" } -TimeoutSec 10
  } catch {
    return $null
  }
}

function Get-LocalCatalogExpectation {
  # 有 token 时以本机 4001 同一接口为准；没有 token 时退化为数据文件统计，
  # 并在结果里标注来源，避免把两种口径混为一谈。
  $api = Get-LocalJson -Path '/api/presets?module=batch-factory' -Token $ApiToken
  if ($null -ne $api -and $null -ne $api.catalog) {
    return @{ count = @($api.catalog).Count; source = 'local-api' }
  }
  $presets = Get-Content -LiteralPath (Join-Path $repo 'data/system/presets.json') -Raw -Encoding UTF8 | ConvertFrom-Json
  $count = @($presets | Where-Object { $_.module -eq 'batch-factory' -and $_.status -eq 'published' }).Count
  return @{ count = $count; source = 'presets-file' }
}

function Get-LocalCapabilityAvailableCount {
  $caps = Get-LocalJson -Path '/api/batch-factory/v12/capabilities' -Token $ApiToken
  if ($null -eq $caps) { return $null }
  $available = 0
  foreach ($property in $caps.PSObject.Properties) {
    if ($property.Value -and $property.Value.available -eq $true) { $available++ }
  }
  return $available
}

function Get-EnvState {
  param([string[]]$Names, [string]$File)
  $fileKeys = @{}
  if ($File -and (Test-Path -LiteralPath $File)) {
    foreach ($line in Get-Content -LiteralPath $File -Encoding UTF8) {
      if ($line -match '^(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)$') {
        $value = $Matches[2].Trim().Trim('"').Trim("'")
        if ($value -ne '' -and $value -ne "''" -and $value -ne '""') { $fileKeys[$Matches[1]] = $true }
      }
    }
  }
  $present = @()
  $missing = @()
  foreach ($name in $Names) {
    $processValue = [Environment]::GetEnvironmentVariable($name)
    if ((-not [string]::IsNullOrWhiteSpace($processValue)) -or $fileKeys.ContainsKey($name)) {
      $present += $name
    } else {
      $missing += $name
    }
  }
  return @{ present = $present; missing = $missing; fileUsed = $File }
}

function Get-DataWhitelist {
  $presetRel = 'data/system/presets.json'
  $presetPath = Join-Path $repo $presetRel
  if (-not (Test-Path -LiteralPath $presetPath)) {
    throw "白名单数据缺失：$presetRel"
  }
  $presets = Get-Content -LiteralPath $presetPath -Raw -Encoding UTF8 | ConvertFrom-Json
  $artifactDir = Join-Path $repo 'data/local-executor-artifacts'
  $entries = @()
  $entries += [pscustomobject]@{ relative = $presetRel; absolute = $presetPath; kind = 'preset-library'; bytes = (Get-Item -LiteralPath $presetPath).Length }
  $missingPreviews = @()
  $referenced = @($presets | ForEach-Object { [string]$_.previewAssetId } | Where-Object { $_ -and $_.Trim() -ne '' } | Select-Object -Unique)
  foreach ($assetId in $referenced) {
    if ($assetId -notmatch '^[A-Za-z0-9_-]+$') {
      $missingPreviews += $assetId
      continue
    }
    $matched = $null
    if (Test-Path -LiteralPath $artifactDir) {
      $matched = Get-ChildItem -LiteralPath $artifactDir -File -ErrorAction SilentlyContinue |
        Where-Object { $_.BaseName -eq $assetId -and $_.Extension -match '(?i)^\.(png|jpg|jpeg|webp)$' } |
        Select-Object -First 1
    }
    if ($null -eq $matched) {
      $missingPreviews += $assetId
      continue
    }
    $entries += [pscustomobject]@{
      relative = "data/local-executor-artifacts/$($matched.Name)"
      absolute = $matched.FullName
      kind = 'preset-preview'
      bytes = $matched.Length
    }
  }
  return @{ entries = $entries; referenced = $referenced; missing = $missingPreviews }
}

function Sync-SnapshotDist {
  # 与 build-go-local.ps1 的静态资源准备步骤保持一致，但源目录是快照而不是实时 dist。
  param([string]$Source)
  $goDist = Join-Path $repo 'backend/web/dist'
  $petsSource = Join-Path $repo 'pets'
  $brandAssetsSource = Join-Path $repo 'frontend/public/assets'
  New-Item -ItemType Directory -Force $goDist | Out-Null
  Get-ChildItem -LiteralPath $goDist -Force | Where-Object { $_.Name -ne '.gitkeep' } | Remove-Item -Recurse -Force
  Copy-Item -Path (Join-Path $Source '*') -Destination $goDist -Recurse -Force
  $legacyEmbeddedRoute = Join-Path $goDist 'batch-rewrite'
  if (Test-Path -LiteralPath $legacyEmbeddedRoute) {
    Remove-Item -LiteralPath $legacyEmbeddedRoute -Recurse -Force
  }
  foreach ($brandAsset in @('brand-logo-black.png', 'brand-logo-white.png')) {
    $brandSource = Join-Path $brandAssetsSource $brandAsset
    if (-not (Test-Path -LiteralPath $brandSource)) { throw "Brand asset is missing: $brandSource" }
    Copy-Item -LiteralPath $brandSource -Destination (Join-Path $goDist "assets/$brandAsset") -Force
  }
  if (-not (Test-Path -LiteralPath (Join-Path $petsSource 'stacky/spritesheet.webp'))) {
    throw "Pet assets are missing: $petsSource"
  }
  New-Item -ItemType Directory -Force (Join-Path $goDist 'pets') | Out-Null
  Copy-Item -Path (Join-Path $petsSource '*') -Destination (Join-Path $goDist 'pets') -Recurse -Force
  Write-Output "Go static assets prepared from snapshot: $Source"
}

function Get-BuildIdForToday {
  $sha = (& git -C $repo rev-parse HEAD).Trim()
  if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($sha)) {
    throw '无法读取 Git HEAD，停止发布。'
  }
  $date = Get-Date -Format 'yyyyMMdd'
  $existing = @()
  if (Test-Path -LiteralPath $archivesRoot) {
    $existing = @(Get-ChildItem -LiteralPath $archivesRoot -Directory | Where-Object { $_.Name -match "^$date-\d{2}-" })
  }
  $sequence = $existing.Count + 1
  return ('{0}-{1:D2}-{2}' -f $date, $sequence, $sha.Substring(0, 8))
}

function Write-ArchiveSnapshot {
  param([string]$ArchiveDir)
  $sha = (& git -C $repo rev-parse HEAD).Trim()
  Write-Utf8NoBom -Path (Join-Path $ArchiveDir 'HEAD.txt') -Content ($sha + "`n")
  Write-Utf8NoBom -Path (Join-Path $ArchiveDir 'head-sha.txt') -Content ($sha + "`n")

  $status = (& git -C $repo status --porcelain) | Out-String
  Write-Utf8NoBom -Path (Join-Path $ArchiveDir 'git-status.txt') -Content $status
  Write-Utf8NoBom -Path (Join-Path $ArchiveDir 'git-status-porcelain.txt') -Content $status

  $untracked = @((& git -C $repo status --porcelain) | Where-Object { $_ -match '^\?\?' } | ForEach-Object { $_.Substring(3).Trim() })
  Write-Utf8NoBom -Path (Join-Path $ArchiveDir 'untracked.txt') -Content (($untracked -join "`n") + "`n")
  Write-Utf8NoBom -Path (Join-Path $ArchiveDir 'untracked-files.txt') -Content (($untracked -join "`n") + "`n")

  $diff = (& git -C $repo diff --no-color HEAD) | Out-String
  Write-Utf8NoBom -Path (Join-Path $ArchiveDir 'changes.diff') -Content $diff
  Write-Utf8NoBom -Path (Join-Path $ArchiveDir 'git-diff-worktree.patch') -Content $diff
  Write-Utf8NoBom -Path (Join-Path $ArchiveDir 'git-diff-stat.txt') -Content ((& git -C $repo diff --stat HEAD) | Out-String)

  # 未跟踪源码快照：只收 backend/frontend/src/scripts/docs 下的文本文件，
  # 跳过 docs/releases、.backups、敏感文件名与大于 2MB 的二进制。
  $snapshotRoot = Join-Path $ArchiveDir 'untracked-snapshot'
  New-Item -ItemType Directory -Force $snapshotRoot | Out-Null
  $snapshotLines = @()
  $allowedPrefixes = @('backend/', 'frontend/src/', 'scripts/', 'docs/')
  $textExtensions = @('.go', '.js', '.jsx', '.ts', '.tsx', '.json', '.md', '.ps1', '.sh', '.css', '.yml', '.yaml', '.sql')
  foreach ($item in $untracked) {
    $rel = $item.Replace('\', '/')
    if ($rel -eq 'docs/releases/' -or $rel.StartsWith('docs/releases/')) { continue }
    if ($rel -eq 'backend/current-release.json') { continue }
    if ($rel -match '(^|/)(\.backups|node_modules|\.local-review|\.release)(/|$)') { continue }
    $sourceFiles = @()
    if (Test-Path -LiteralPath (Join-Path $repo $rel) -PathType Container) {
      $sourceFiles = @(Get-ChildItem -LiteralPath (Join-Path $repo $rel) -Recurse -File -ErrorAction SilentlyContinue)
    } elseif (Test-Path -LiteralPath (Join-Path $repo $rel)) {
      $sourceFiles = @(Get-Item -LiteralPath (Join-Path $repo $rel))
    }
    foreach ($file in $sourceFiles) {
      $fileRel = Get-RelativePath $file.FullName
      $allowed = $false
      foreach ($prefix in $allowedPrefixes) {
        if ($fileRel.StartsWith($prefix)) { $allowed = $true; break }
      }
      if (-not $allowed) { continue }
      if (Test-SensitiveName $fileRel) { continue }
      if ($textExtensions -notcontains $file.Extension.ToLowerInvariant()) { continue }
      if ($file.Length -gt 2MB) { continue }
      $destination = Join-Path $snapshotRoot $fileRel
      New-Item -ItemType Directory -Force (Split-Path -Parent $destination) | Out-Null
      Copy-Item -LiteralPath $file.FullName -Destination $destination -Force
      $snapshotLines += $fileRel
    }
  }
  Write-Utf8NoBom -Path (Join-Path $ArchiveDir 'untracked-snapshot-manifest.txt') -Content (($snapshotLines -join "`n") + "`n")
}

function New-DataPackage {
  param([string]$ReleaseDir, [string]$ArchiveDir, [string[]]$ExcludedSensitiveFiles)
  $whitelist = Get-DataWhitelist
  $stage = Join-Path $ReleaseDir 'data-stage'
  if (Test-Path -LiteralPath $stage) { Remove-Item -LiteralPath $stage -Recurse -Force }
  New-Item -ItemType Directory -Force $stage | Out-Null

  $entries = @()
  foreach ($item in $whitelist.entries) {
    if (Test-SensitiveName $item.relative) {
      throw "红线拦截：白名单文件命中敏感规则，拒绝打包：$($item.relative)"
    }
    $destination = Join-Path $stage $item.relative
    New-Item -ItemType Directory -Force (Split-Path -Parent $destination) | Out-Null
    Copy-Item -LiteralPath $item.absolute -Destination $destination -Force
    $entries += [pscustomobject]@{
      path = $item.relative
      bytes = (Get-Item -LiteralPath $destination).Length
      sha256 = Get-FileSha256 $destination
      kind = $item.kind
    }
  }

  $zipPath = Join-Path $ReleaseDir "$(Split-Path -Leaf $ArchiveDir)-data.zip"
  if (Test-Path -LiteralPath $zipPath) { Remove-Item -LiteralPath $zipPath -Force }
  Compress-Archive -Path (Join-Path $stage 'data') -DestinationPath $zipPath -CompressionLevel Optimal

  Add-Type -AssemblyName System.IO.Compression.FileSystem
  $archive = [System.IO.Compression.ZipFile]::OpenRead($zipPath)
  try {
    foreach ($entry in $archive.Entries) {
      if (Test-SensitiveName $entry.FullName) {
        throw "红线拦截：发布压缩包内出现敏感文件名：$($entry.FullName)"
      }
    }
  } finally {
    $archive.Dispose()
  }

  $manifest = [ordered]@{
    buildId = (Split-Path -Leaf $ArchiveDir)
    createdAt = (Get-Date).ToUniversalTime().ToString('o')
    package = Get-RelativePath $zipPath
    packageBytes = (Get-Item -LiteralPath $zipPath).Length
    packageSha256 = Get-FileSha256 $zipPath
    files = $entries
    referencedPreviewIds = $whitelist.referenced
    missingPreviewIds = $whitelist.missing
    excludedSensitiveFiles = $ExcludedSensitiveFiles
  }
  Write-JsonNoBom -Path (Join-Path $ArchiveDir 'data-manifest.json') -Content $manifest

  return [pscustomobject]@{
    zipPath = $zipPath
    stageDir = $stage
    entries = $entries
    missing = $whitelist.missing
    excluded = $ExcludedSensitiveFiles
    sha256 = $manifest.packageSha256
    bytes = $manifest.packageBytes
    manifest = $manifest
  }
}

function Get-ExcludedSensitiveFiles {
  $dataRoot = Join-Path $repo 'data'
  if (-not (Test-Path -LiteralPath $dataRoot)) { return @() }
  $excluded = @()
  foreach ($file in Get-ChildItem -LiteralPath $dataRoot -Recurse -File -ErrorAction SilentlyContinue) {
    $rel = Get-RelativePath $file.FullName
    if (Test-SensitiveName $rel) { $excluded += $rel }
  }
  return $excluded
}

function Invoke-RemoteDeploy {
  param([object]$Context)

  $backupRoot = Get-RemoteBackupRoot
  $backupDir = "$backupRoot/$BuildId"
  $remoteTemp = "/tmp/qiantie-release-$BuildId"

  Write-Section '远端预检'
  $probe = Invoke-Remote -Command 'echo v88-ok'
  Write-Host "SSH 连通：$probe"
  Invoke-Remote -Command "mkdir -p '$backupDir' '$remoteTemp'" | Out-Null

  Write-Section '远端备份（二进制 + presets）'
  $backupCommand = "set -eu; if [ -f '$TargetPath' ]; then cp -p '$TargetPath' '$backupDir/qiantie-server'; fi;"
  if (-not [string]::IsNullOrWhiteSpace($RemoteDataDir)) {
    $backupCommand += " if [ -f '$RemoteDataDir/system/presets.json' ]; then cp -p '$RemoteDataDir/system/presets.json' '$backupDir/presets.json'; fi;"
  }
  Invoke-Remote -Command $backupCommand | Out-Null
  Write-Host "备份目录：$backupDir"

  Write-Section '上传并替换二进制'
  $remoteBinaryTemp = "$remoteTemp/qiantie-server"
  Invoke-Checked 'scp' @('-P', "$SshPort", $Context.LinuxBinaryPath, "${TargetUser}@${TargetHost}:$remoteBinaryTemp")
  Invoke-Remote -Command "set -eu; sudo install -D -m 0755 '$remoteBinaryTemp' '$TargetPath'" | Out-Null

  $dataUploaded = $false
  if (-not $SkipDataUpload -and -not [string]::IsNullOrWhiteSpace($RemoteDataDir)) {
    Write-Section '上传发布数据（presets + 被引用的预览图）'
    $localDataStage = Join-Path $Context.DataStageDir 'data'
    Invoke-Checked 'scp' @('-P', "$SshPort", '-r', $localDataStage, "${TargetUser}@${TargetHost}:$remoteTemp")
    $remoteDataStage = "$remoteTemp/data"
    $installCommand = "set -eu; sudo mkdir -p '$RemoteDataDir/system' '$RemoteDataDir/local-executor-artifacts'"
    foreach ($entry in $Context.DataEntries) {
      $installCommand += "; sudo install -D -m 0644 '$remoteDataStage/$($entry.path)' '$RemoteDataDir/$($entry.path)'"
    }
    Invoke-Remote -Command $installCommand | Out-Null
    $dataUploaded = $true
    Write-Host "数据已落地到：$RemoteDataDir"
  } elseif ($SkipDataUpload) {
    Write-Host '已按 -SkipDataUpload 跳过数据上传：线上画风库/预览图不会更新。' -ForegroundColor Yellow
  } else {
    throw '部署时必须提供 -RemoteDataDir（生产 data 根目录），否则 presets.json 不会随发布更新；如确实不想传数据请显式加 -SkipDataUpload。'
  }

  Write-Section '重启服务并执行远程校验'
  Invoke-Remote -Command "set -eu; sudo systemctl restart '$ServiceName'; sudo systemctl is-active --quiet '$ServiceName'" | Out-Null

  $checks = @()
  $failed = @()
  $health = Get-RemoteHttpStatus '/health'
  $checks += [pscustomobject]@{ name = '/health'; expected = '200'; actual = $health; pass = ($health -eq '200') }
  if ($health -ne '200') { $failed += "/health=$health" }
  $script = Get-RemoteHttpStatus '/script'
  $checks += [pscustomobject]@{ name = '/script'; expected = '200'; actual = $script; pass = ($script -eq '200') }
  if ($script -ne '200') { $failed += "/script=$script" }

  $buildInfo = Invoke-Remote -Command "curl -s '$RemoteProbeBaseUrl/api/runtime-build-info'"
  $buildMatch = $false
  try {
    $parsed = $buildInfo | ConvertFrom-Json
    $buildMatch = ($parsed.git_sha -eq $Context.ShortSha)
    $checks += [pscustomobject]@{ name = '/api/runtime-build-info git_sha'; expected = $Context.ShortSha; actual = $parsed.git_sha; pass = $buildMatch }
    if (-not $buildMatch) { $failed += "runtime-build-info git_sha=$($parsed.git_sha)" }
  } catch {
    $checks += [pscustomobject]@{ name = '/api/runtime-build-info'; expected = 'json'; actual = $buildInfo; pass = $false }
    $failed += 'runtime-build-info 无法解析'
  }

  if ($dataUploaded) {
    $remoteBytes = [int64](Invoke-Remote -Command "stat -c %s '$RemoteDataDir/system/presets.json'")
    $localBytes = [int64](Get-Item -LiteralPath (Join-Path $repo 'data/system/presets.json')).Length
    $bytesPass = ($remoteBytes -eq $localBytes)
    $checks += [pscustomobject]@{ name = 'presets.json bytes'; expected = "$localBytes"; actual = "$remoteBytes"; pass = $bytesPass }
    if (-not $bytesPass) { $failed += "presets.json 字节数不一致（本地 $localBytes / 远程 $remoteBytes）" }

    $remoteSha = (Invoke-Remote -Command "sha256sum '$RemoteDataDir/system/presets.json' | cut -d' ' -f1")
    $localSha = Get-FileSha256 (Join-Path $repo 'data/system/presets.json')
    $shaPass = ($remoteSha -eq $localSha)
    $checks += [pscustomobject]@{ name = 'presets.json sha256'; expected = $localSha; actual = $remoteSha; pass = $shaPass }
    if (-not $shaPass) { $failed += 'presets.json SHA256 不一致' }
  }

  if ([string]::IsNullOrWhiteSpace($ApiToken)) {
    $checks += [pscustomobject]@{ name = 'presets catalog 条数'; expected = $Context.ExpectedCatalog; actual = '未验证（缺少 API token）'; pass = $null }
    $checks += [pscustomobject]@{ name = 'capabilities available 数'; expected = $Context.ExpectedCapabilities; actual = '未验证（缺少 API token）'; pass = $null }
  } else {
    $remoteCatalog = Get-RemoteJson -Path '/api/presets?module=batch-factory' -Token $ApiToken
    if ($null -eq $remoteCatalog -or $null -eq $remoteCatalog.catalog) {
      $checks += [pscustomobject]@{ name = 'presets catalog 条数'; expected = $Context.ExpectedCatalog; actual = '请求失败'; pass = $false }
      $failed += '/api/presets?module=batch-factory 请求失败'
    } else {
      $count = @($remoteCatalog.catalog).Count
      $catalogPass = ($count -ge [int]$Context.ExpectedCatalog)
      $checks += [pscustomobject]@{ name = 'presets catalog 条数'; expected = ">= $($Context.ExpectedCatalog)"; actual = "$count"; pass = $catalogPass }
      if (-not $catalogPass) { $failed += "catalog 条数 $count < $($Context.ExpectedCatalog)" }
    }

    $remoteCaps = Get-RemoteJson -Path '/api/batch-factory/v12/capabilities' -Token $ApiToken
    if ($null -eq $remoteCaps -or $null -eq $Context.ExpectedCapabilities) {
      $checks += [pscustomobject]@{ name = 'capabilities available 数'; expected = $Context.ExpectedCapabilities; actual = '未验证'; pass = $null }
    } else {
      $available = 0
      foreach ($property in $remoteCaps.PSObject.Properties) {
        if ($property.Value -and $property.Value.available -eq $true) { $available++ }
      }
      $capsPass = ($available -eq [int]$Context.ExpectedCapabilities)
      $checks += [pscustomobject]@{ name = 'capabilities available 数'; expected = $Context.ExpectedCapabilities; actual = "$available"; pass = $capsPass }
      if (-not $capsPass) { $failed += "capabilities available=$available != 本地 $($Context.ExpectedCapabilities)" }
    }
  }

  $checks | Format-Table -AutoSize | Out-String | Write-Host

  $result = [ordered]@{
    buildId = $BuildId
    deployedAt = (Get-Date).ToUniversalTime().ToString('o')
    targetHost = $TargetHost
    service = $ServiceName
    backupDir = $backupDir
    dataUploaded = $dataUploaded
    remoteDataDir = $RemoteDataDir
    checks = $checks
    failed = $failed
    rolledBack = $false
  }

  if ($failed.Count -gt 0) {
    Write-Host '远程校验失败，开始自动回滚...' -ForegroundColor Red
    Invoke-RemoteRollback -BackupDirectory $backupDir
    $result.rolledBack = $true
    Write-JsonNoBom -Path (Join-Path $Context.ArchiveDir 'deploy-result.json') -Content $result
    throw "远程校验失败并已回滚：$($failed -join '; ')"
  }

  Write-Host '远程校验全部通过。' -ForegroundColor Green
  Write-JsonNoBom -Path (Join-Path $Context.ArchiveDir 'deploy-result.json') -Content $result
}

function Invoke-RemoteRollback {
  param([string]$BackupDirectory)
  $command = "set -eu;"
  $command += " if [ -f '$BackupDirectory/qiantie-server' ]; then sudo install -D -m 0755 '$BackupDirectory/qiantie-server' '$TargetPath'; fi;"
  if (-not [string]::IsNullOrWhiteSpace($RemoteDataDir)) {
    $command += " if [ -f '$BackupDirectory/presets.json' ]; then sudo install -D -m 0644 '$BackupDirectory/presets.json' '$RemoteDataDir/system/presets.json'; fi;"
  }
  $command += " sudo systemctl restart '$ServiceName'; sudo systemctl is-active --quiet '$ServiceName';"
  Invoke-Remote -Command $command | Out-Null
  $health = Get-RemoteHttpStatus '/health'
  Write-Host "回滚完成：$BackupDirectory（/health=$health）" -ForegroundColor Yellow
}

function Invoke-LocalPreflight {
  # 21 版起的硬门禁：发线上之前，先用本轮构建产物在本地独立端口起一个实例，
  # 验证“本地看到的”就是候选版本。任一项不通过就抛错中止，绝不上传。
  param(
    [string]$ReleaseId,
    [string]$ArchiveDir,
    [string]$WindowsBinary,
    [string]$FrontendEntry,
    [string]$ShortSha,
    [int]$Port,
    [int]$TimeoutSeconds
  )

  $base = "http://127.0.0.1:$Port"
  $manifestPath = Join-Path $ArchiveDir 'local-preflight-runtime-manifest.json'
  Write-JsonNoBom -Path $manifestPath -Content ([ordered]@{
    buildId = $BuildId
    releaseId = $ReleaseId
    gitSha = $ShortSha
    gitShort = $ShortSha
    frontendEntry = $FrontendEntry
    binaryPath = (Get-RelativePath $WindowsBinary)
    builtAt = (Get-Date).ToUniversalTime().ToString('o')
  })

  $checks = @()
  $failed = @()
  $instancePid = 0
  $launchOutput = ''
  $launcher = Join-Path $PSScriptRoot 'start-local-go.ps1'

  try {
    # 用 Start-Process + 文件重定向启动候选实例，再用 Process.WaitForExit 等
    # 启动器自身退出。直接管道（或 *>）捕获输出会一直读到 EOF，而被拉起的
    # 候选进程继承了写端，门禁会卡死在“已后台启动”之后。
    $launchLog = Join-Path $ArchiveDir 'local-preflight-launch.log'
    $launchErrLog = Join-Path $ArchiveDir 'local-preflight-launch.err.log'
    $launcherProcess = Start-Process -FilePath 'powershell.exe' -PassThru -WindowStyle Hidden `
      -RedirectStandardOutput $launchLog -RedirectStandardError $launchErrLog `
      -ArgumentList @(
        '-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', $launcher,
        '-SkipBuild', '-NoBuildCheck', '-Detached',
        '-ListenAddr', "127.0.0.1:$Port",
        '-BinaryPath', $WindowsBinary, '-ManifestPath', $manifestPath
      )
    [void]$launcherProcess.WaitForExit(30000)
    $launchOutput = [string](Get-Content -LiteralPath $launchLog -Raw -ErrorAction SilentlyContinue) +
      [string](Get-Content -LiteralPath $launchErrLog -Raw -ErrorAction SilentlyContinue)
    $pidMatch = [regex]::Match($launchOutput, 'pid=(\d+)')
    if ($pidMatch.Success) {
      $instancePid = [int]$pidMatch.Groups[1].Value
    } else {
      $failed += "本地实例启动失败：$($launchOutput.Trim())"
    }

    if ($failed.Count -eq 0) {
      $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
      $ready = $false
      while (-not $ready -and (Get-Date) -lt $deadline) {
        try {
          $probe = Invoke-WebRequest -Uri "$base/health" -TimeoutSec 3 -UseBasicParsing
          if ([int]$probe.StatusCode -eq 200) { $ready = $true; break }
        } catch {
          Start-Sleep -Milliseconds 500
        }
      }
      if (-not $ready) { $failed += "/health 在 $TimeoutSeconds 秒内未就绪（$base）" }
    }

    if ($failed.Count -eq 0) {
      $homeStatus = 'n/a'
      try { $homeStatus = [string](Invoke-WebRequest -Uri "$base/" -TimeoutSec 10 -UseBasicParsing).StatusCode } catch { $homeStatus = $_.Exception.Message }
      $homePass = ($homeStatus -eq '200')
      $checks += [pscustomobject]@{ name = '首页 /'; expected = '200'; actual = $homeStatus; pass = $homePass }
      if (-not $homePass) { $failed += "首页 / = $homeStatus" }

      $healthStatus = 'n/a'
      try { $healthStatus = [string](Invoke-WebRequest -Uri "$base/health" -TimeoutSec 10 -UseBasicParsing).StatusCode } catch { $healthStatus = $_.Exception.Message }
      $healthPass = ($healthStatus -eq '200')
      $checks += [pscustomobject]@{ name = '/health'; expected = '200'; actual = $healthStatus; pass = $healthPass }
      if (-not $healthPass) { $failed += "/health = $healthStatus" }

      $updatesReleaseId = ''
      $updatesStatus = 'n/a'
      try {
        $updates = Invoke-RestMethod -Uri "$base/api/product-updates/latest" -TimeoutSec 10
        $updatesReleaseId = [string]$updates.releaseId
        $updatesStatus = '200'
      } catch {
        $updatesStatus = $_.Exception.Message
      }
      $updatesPass = ($updatesStatus -eq '200' -and $updatesReleaseId -eq $ReleaseId)
      $checks += [pscustomobject]@{ name = '/api/product-updates/latest releaseId'; expected = $ReleaseId; actual = "$updatesStatus/$updatesReleaseId"; pass = $updatesPass }
      if (-not $updatesPass) { $failed += "/api/product-updates/latest releaseId=$updatesReleaseId（期望 $ReleaseId）" }

      $runtimeEntry = ''
      $runtimeSha = ''
      try {
        $runtime = Invoke-RestMethod -Uri "$base/api/runtime-build-info" -TimeoutSec 10
        $runtimeEntry = [string]$runtime.frontend_entry
        $runtimeSha = [string]$runtime.runtime_manifest_sha
      } catch {
        $runtimeEntry = $_.Exception.Message
      }
      $entryPass = ($runtimeEntry -eq $FrontendEntry)
      $checks += [pscustomobject]@{ name = '/api/runtime-build-info frontend_entry'; expected = $FrontendEntry; actual = $runtimeEntry; pass = $entryPass }
      if (-not $entryPass) { $failed += "frontend_entry=$runtimeEntry（期望 $FrontendEntry）" }

      $indexHtml = ''
      try {
        $indexHtml = (Invoke-WebRequest -Uri "$base/" -TimeoutSec 10 -UseBasicParsing).Content
      } catch {
        $indexHtml = ''
      }
      $indexPass = ($indexHtml -match [regex]::Escape($FrontendEntry))
      $checks += [pscustomobject]@{ name = '首页 index.html 入口 chunk'; expected = $FrontendEntry; actual = $(if ($indexPass) { $FrontendEntry } else { '未命中' }); pass = $indexPass }
      if (-not $indexPass) { $failed += "首页未引用候选入口 $FrontendEntry" }
    }
  } finally {
    if ($instancePid -gt 0) {
      try {
        $proc = Get-Process -Id $instancePid -ErrorAction Stop
        if ($proc.Path -and ([System.IO.Path]::GetFullPath($proc.Path) -eq [System.IO.Path]::GetFullPath($WindowsBinary))) {
          Stop-Process -Id $instancePid -Force
        }
      } catch {
        # 进程已退出或路径不匹配：不误杀其它进程。
      }
    }
  }

  $report = [ordered]@{
    releaseId = $ReleaseId
    port = $Port
    binary = (Get-RelativePath $WindowsBinary)
    binarySha256 = (Get-FileSha256 $WindowsBinary)
    runtimeManifestSha = $runtimeSha
    startedAt = (Get-Date).ToUniversalTime().ToString('o')
    checks = $checks
    failed = $failed
    pass = ($failed.Count -eq 0)
  }
  Write-JsonNoBom -Path (Join-Path $ArchiveDir 'local-preflight.json') -Content $report
  $checks | Format-Table -AutoSize | Out-String | Write-Host

  if ($failed.Count -gt 0) {
    Write-Host '本地前置验证未通过，已中止：不会上传线上。' -ForegroundColor Red
    throw "本地前置验证未通过：$($failed -join '; ')"
  }
  Write-Host "本地前置验证通过：releaseId=$ReleaseId, entry=$FrontendEntry" -ForegroundColor Green
  return $report
}

# ---------------------------------------------------------------------------
# 只读模式：列出远端回滚点
# ---------------------------------------------------------------------------
if ($ListBackups) {
  Assert-RemoteParameters
  $backupRoot = Get-RemoteBackupRoot
  $listing = Invoke-Remote -Command "ls -1dt '$backupRoot'/* 2>/dev/null || true"
  if ([string]::IsNullOrWhiteSpace($listing)) {
    Write-Host "远端没有回滚点：$backupRoot"
  } else {
    Write-Host '远端回滚点（新 -> 旧）：'
    $listing
  }
  exit 0
}

# ---------------------------------------------------------------------------
# 回滚模式
# ---------------------------------------------------------------------------
if ($Rollback) {
  Assert-RemoteParameters
  $backupRoot = Get-RemoteBackupRoot
  if ([string]::IsNullOrWhiteSpace($BackupId)) {
    $latest = Invoke-Remote -Command "ls -1dt '$backupRoot'/* 2>/dev/null | head -n 1"
    if ([string]::IsNullOrWhiteSpace($latest)) {
      throw "远端没有可用回滚点：$backupRoot"
    }
    $BackupId = Split-Path -Leaf $latest
  }
  $targetBackup = "$backupRoot/$BackupId"
  Write-Section "回滚到 $BackupId"
  Invoke-RemoteRollback -BackupDirectory $targetBackup
  exit 0
}

if ($Deploy -and [string]::IsNullOrWhiteSpace($TargetHost)) {
  throw '启用 -Deploy 时必须提供 -TargetHost。'
}

Write-Section '1/8 前置检查'
$headSha = (& git -C $repo rev-parse HEAD).Trim()
if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($headSha)) { throw '无法读取 Git HEAD。' }
$shortSha = $headSha.Substring(0, 8)
$statusLines = @((& git -C $repo status --porcelain))
$untrackedCount = @($statusLines | Where-Object { $_ -match '^\?\?' }).Count
Write-Host "HEAD：$headSha"
Write-Host "工作区改动：$($statusLines.Count) 条（其中未跟踪 $untrackedCount 条）"
if ($statusLines.Count -gt 0) {
  Write-Host "本版 = HEAD $shortSha + 工作区 diff（将完整存档，不会静默打包）。" -ForegroundColor Yellow
}

if ([string]::IsNullOrWhiteSpace($BuildId)) {
  $BuildId = Get-BuildIdForToday
} elseif ($BuildId -notmatch '^\d{8}-\d{2}-[0-9a-fA-F]{8}$') {
  throw "BuildId 格式非法：$BuildId（应为 20261104-01-6db3b5ad）"
}

# releaseId 必须先于构建确定：更新说明按 releaseId 编进二进制。
if ([string]::IsNullOrWhiteSpace($ReleaseId)) {
  $ReleaseId = 'qiantie-' + (Get-Date -Format 'yyyyMMddHHmm')
}
if ($ReleaseId -notmatch '^qiantie-\d{12}$') {
  throw "ReleaseId 格式非法：$ReleaseId（应为 qiantie-202611060214）"
}
$releaseNotesPath = Join-Path $backendDir "internal/releasenotes/notes/$ReleaseId.json"
if (-not (Test-Path -LiteralPath $releaseNotesPath)) {
  throw "缺少本次更新说明：$releaseNotesPath。notes/<releaseId>.json 缺失即视为发布失败（硬门禁）。"
}
Write-Host "releaseId：$ReleaseId（说明文件：$(Get-RelativePath $releaseNotesPath)）"

$archiveDir = Join-Path $archivesRoot $BuildId
$releaseDir = Join-Path $releaseRoot $BuildId

$excludedSensitive = Get-ExcludedSensitiveFiles

if ($DryRun) {
  Write-Section 'dry-run：数据白名单预览（不构建、不写存档、不上传）'
  $whitelist = Get-DataWhitelist
  $whitelist.entries | Format-Table kind, relative, bytes -AutoSize | Out-String | Write-Host
  if ($whitelist.missing.Count -gt 0) {
    Write-Host "缺少预览图（$($whitelist.missing.Count)）：$($whitelist.missing -join ', ')" -ForegroundColor Yellow
  }
  Write-Host "已排除敏感文件 $($excludedSensitive.Count) 个（api-config.json / upload-target.json / 含 cookie、token、password、secret、.env）："
  $excludedSensitive | Select-Object -First 20
  Write-Host 'dry-run 完成，未做任何构建或上传。'
  exit 0
}

if ((Test-Path -LiteralPath $archiveDir) -and -not $Force) {
  throw "存档已存在：$archiveDir。换一个批次或显式加 -Force。"
}
if (Test-Path -LiteralPath $archiveDir) { Remove-Item -LiteralPath $archiveDir -Recurse -Force }
if (Test-Path -LiteralPath $releaseDir) { Remove-Item -LiteralPath $releaseDir -Recurse -Force }
New-Item -ItemType Directory -Force $archiveDir | Out-Null
New-Item -ItemType Directory -Force $releaseDir | Out-Null

Write-Section '2/8 工作区存档'
Write-ArchiveSnapshot -ArchiveDir $archiveDir
Write-Host "存档目录：$archiveDir"
Write-Host "已排除敏感文件 $($excludedSensitive.Count) 个（api-config.json / upload-target.json / 含 cookie、token、password、secret、.env）"

Write-Section '3/8 构建（npm -> embed 同步 -> Go 二进制）'
$buildTimer = Measure-Command {
  if ($useSnapshotDist) {
    Write-Host "已跳过前端构建，使用产物快照：$useSnapshotDist"
  } elseif (-not $SkipFrontendBuild) {
    Invoke-Checked 'npm.cmd' @('--prefix', $frontendDir, 'run', 'build')
  } else {
    Write-Host '已跳过前端构建（-SkipFrontendBuild）。'
  }
}
Write-Host ("前端构建耗时：{0:N1}s" -f $buildTimer.TotalSeconds)

if (-not (Test-Path -LiteralPath (Join-Path $frontendDist 'index.html'))) {
  throw "前端构建结果不存在：$frontendDist/index.html"
}
# 有快照时只从快照同步，绝不读取可能被并行会话覆盖的实时 frontend/dist。
# 本地 exe 当前可能被 4001 锁定，因此二进制输出到 .release 目录，绝不覆盖运行中的文件。
if ($useSnapshotDist) {
  Sync-SnapshotDist -Source $useSnapshotDist
} else {
  Invoke-Checked 'powershell' @('-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', (Join-Path $PSScriptRoot 'build-go-local.ps1'), '-SkipGoBuild')
}

$indexHtml = Get-Content -LiteralPath (Join-Path $frontendDist 'index.html') -Raw
$entryMatch = [regex]::Match($indexHtml, '<script[^>]+type=["'']module["''][^>]+src=["'']([^"'']+\.js)["'']')
if (-not $entryMatch.Success) { throw '无法从 index.html 解析前端入口文件。' }
$frontendEntry = $entryMatch.Groups[1].Value
Write-Host "前端入口：$frontendEntry"

$windowsBinary = Join-Path $releaseDir 'qiantie-local.exe'
$linuxBinary = Join-Path $releaseDir 'qiantie-server-linux-amd64'

$goTimer = Measure-Command {
  Push-Location $backendDir
  try {
    Invoke-Checked 'go' @('build', '-o', $windowsBinary, './cmd/qiantie')

    $oldGoos = $env:GOOS; $oldGoarch = $env:GOARCH; $oldCgo = $env:CGO_ENABLED
    try {
      $env:GOOS = 'linux'; $env:GOARCH = 'amd64'; $env:CGO_ENABLED = '0'
      Invoke-Checked 'go' @('build', '-trimpath', '-ldflags=-s -w', '-o', $linuxBinary, './cmd/qiantie')
    } finally {
      $env:GOOS = $oldGoos; $env:GOARCH = $oldGoarch; $env:CGO_ENABLED = $oldCgo
    }
  } finally {
    Pop-Location
  }
}
Write-Host ("Go 构建耗时：{0:N1}s" -f $goTimer.TotalSeconds)

$windowsSha = Get-FileSha256 $windowsBinary
$linuxSha = Get-FileSha256 $linuxBinary
$builtAt = (Get-Date).ToUniversalTime().ToString('o')

Write-Section '4/8 发布数据白名单打包'
$package = New-DataPackage -ReleaseDir $releaseDir -ArchiveDir $archiveDir -ExcludedSensitiveFiles $excludedSensitive
$package.entries | Format-Table kind, path, bytes -AutoSize | Out-String | Write-Host
Write-Host "数据包：$($package.zipPath)（$($package.bytes) bytes，sha256=$($package.sha256)）"
Write-Host "已排除敏感文件 $($package.excluded.Count) 个"
if ($package.missing.Count -gt 0) {
  Write-Host "缺少被引用的预览图 $($package.missing.Count) 个：$($package.missing -join ', ')" -ForegroundColor Yellow
}

Write-Section '5/8 环境变量校验（只查是否设置，不打印明文）'
$envState = Get-EnvState -Names $RequiredEnvVars -File $EnvFile
Write-Host "已设置：$($envState.present.Count) 项；缺失：$($envState.missing.Count) 项"
if ($envState.missing.Count -gt 0) {
  Write-Host "缺失变量：$($envState.missing -join ', ')" -ForegroundColor Red
  if ($Deploy -and -not $AllowMissingEnv) {
    throw '目标环境必需变量缺失；补齐后重试，或显式加 -AllowMissingEnv 继续（默认中止上传）。'
  }
  if (-not $Deploy) {
    Write-Host '当前为 local-only 发布，环境变量缺失不阻断构建；上传前请确认目标环境已具备这些变量。' -ForegroundColor Yellow
  } else {
    Write-Host '已按 -AllowMissingEnv 继续，发布结果需人工复核。' -ForegroundColor Yellow
  }
}

$expectedCatalog = Get-LocalCatalogExpectation
$expectedCapabilities = Get-LocalCapabilityAvailableCount
Write-Host "本地 catalog 预期：$($expectedCatalog.count)（来源：$($expectedCatalog.source)）"
if ($null -eq $expectedCapabilities) {
  Write-Host '本地 capabilities available 数：未获取（4001 未运行或缺少 token）'
} else {
  Write-Host "本地 capabilities available 数：$expectedCapabilities"
}

Write-Section '6/8 本地前置验证（硬门禁：不通过不上线）'
$localPreflight = Invoke-LocalPreflight `
  -ReleaseId $ReleaseId `
  -ArchiveDir $archiveDir `
  -WindowsBinary $windowsBinary `
  -FrontendEntry $frontendEntry `
  -ShortSha $shortSha `
  -Port $LocalVerifyPort `
  -TimeoutSeconds $LocalVerifyTimeoutSeconds

$runtimeManifest = [ordered]@{
  releaseId = $ReleaseId
  buildId = $BuildId
  gitSha = $headSha
  gitShort = $shortSha
  dirty = ($statusLines.Count -gt 0)
  changedPaths = $statusLines.Count
  untrackedPaths = $untrackedCount
  frontendEntry = $frontendEntry
  builtAt = $builtAt
  binaryPath = (Get-RelativePath $windowsBinary)
}
Write-JsonNoBom -Path (Join-Path $archiveDir 'runtime-manifest.json') -Content $runtimeManifest

$buildInfo = [ordered]@{
  buildId = $BuildId
  releaseId = $ReleaseId
  builtAt = $builtAt
  gitSha = $headSha
  gitShort = $shortSha
  worktreeDirty = ($statusLines.Count -gt 0)
  changedPathCount = $statusLines.Count
  untrackedPathCount = $untrackedCount
  frontendEntry = $frontendEntry
  frontendDistSource = $(if ($useSnapshotDist) { $useSnapshotDist } else { 'frontend/dist' })
  frontendBuildSeconds = [Math]::Round($buildTimer.TotalSeconds, 1)
  goBuildSeconds = [Math]::Round($goTimer.TotalSeconds, 1)
  binaries = [ordered]@{
    windows = [ordered]@{ path = (Get-RelativePath $windowsBinary); bytes = (Get-Item -LiteralPath $windowsBinary).Length; sha256 = $windowsSha }
    linux = [ordered]@{ path = (Get-RelativePath $linuxBinary); bytes = (Get-Item -LiteralPath $linuxBinary).Length; sha256 = $linuxSha }
  }
  dataPackage = [ordered]@{ path = (Get-RelativePath $package.zipPath); bytes = $package.bytes; sha256 = $package.sha256; fileCount = $package.entries.Count }
  envRequired = $RequiredEnvVars
  envMissing = $envState.missing
  # 本地前置验证结果：21 版起的硬门禁，未通过时脚本已在上一步中止。
  localPreflight = $localPreflight
  deployTarget = ($(if ($Deploy) { "$TargetUser@$TargetHost" } else { 'local-only' }))
}
Write-JsonNoBom -Path (Join-Path $archiveDir 'build-info.json') -Content $buildInfo
# 兼容既有习惯的纯文本摘要。
$buildInfoText = @(
  "buildId=$BuildId",
  "releaseId=$ReleaseId",
  "localPreflightPass=$([bool]$localPreflight.pass)",
  "builtAt=$builtAt",
  "gitSha=$headSha",
  "worktreeDirty=$($statusLines.Count -gt 0)",
  "frontendEntry=$frontendEntry",
  "windowsBinary=$(Get-RelativePath $windowsBinary)",
  "windowsSha256=$windowsSha",
  "linuxBinary=$(Get-RelativePath $linuxBinary)",
  "linuxSha256=$linuxSha",
  "dataPackage=$(Get-RelativePath $package.zipPath)",
  "dataPackageSha256=$($package.sha256)"
) -join "`n"
Write-Utf8NoBom -Path (Join-Path $archiveDir 'build-info.txt') -Content ($buildInfoText + "`n")

$rollbackDoc = @"
# 回滚说明（$BuildId）

本版 = HEAD ``$headSha`` + 工作区 diff（工作区改动 $($statusLines.Count) 条，未跟踪 $untrackedCount 条）。

## A. 线上回滚（二进制 + presets）

``````powershell
# 查看回滚点
.\scripts\release-v88.ps1 -ListBackups -TargetHost <host> -RemoteDataDir <prod-data>

# 回滚到指定批次（远端备份目录 .rollback/$BuildId）
.\scripts\release-v88.ps1 -Rollback -BackupId $BuildId -TargetHost <host> -RemoteDataDir <prod-data>
``````

远端备份位置：``$(Get-RemoteBackupRoot)/$BuildId``（含 ``qiantie-server`` 与 ``presets.json``）。

## B. 本机 4001 回滚

``````powershell
# 不要直接覆盖运行中的 exe；先停止 4001 对应进程，再复制本批次二进制
Copy-Item "$(Get-RelativePath $windowsBinary)" "backend/qiantie-local.exe" -Force
.\scripts\start-local-go.ps1 -SkipBuild
``````

本轮构建的二进制位于 ``$(Get-RelativePath $windowsBinary)``；使用 ``-UpdateLocalBinary`` 时
旧二进制会备份为 ``$(Get-RelativePath (Join-Path $releaseDir 'previous-qiantie-local.exe'))``。

## C. 源码回滚

``````powershell
git checkout -- .
git apply --3way "docs/releases/$BuildId/changes.diff"
# 未跟踪文件内容见 docs/releases/$BuildId/untracked-snapshot/
``````

注意：``changes.diff`` 只覆盖已跟踪文件的未暂存改动；未跟踪源码在
``untracked-snapshot/`` 内，清单见 ``untracked-snapshot-manifest.txt``。

## D. 数据回滚

``````powershell
# 远端 presets 备份目录：$(Get-RemoteBackupRoot)/$BuildId
#   其中 presets.json 为上传前的线上画风库，qiantie-server 为上传前的二进制
# 本地原始数据仍在 data/system/presets.json（本脚本不修改本地数据）
``````
"@
Write-Utf8NoBom -Path (Join-Path $archiveDir 'ROLLBACK.md') -Content $rollbackDoc

$currentRelease = [ordered]@{
  buildId = $BuildId
  releaseId = $ReleaseId
  archiveDir = (Get-RelativePath $archiveDir)
  releaseDir = (Get-RelativePath $releaseDir)
  gitSha = $headSha
  frontendEntry = $frontendEntry
  localPreflightPass = [bool]$localPreflight.pass
  localBinarySha256 = $windowsSha
  linuxBinarySha256 = $linuxSha
  dataPackageSha256 = $package.sha256
  writtenAt = (Get-Date).ToUniversalTime().ToString('o')
}
Write-JsonNoBom -Path $currentReleasePath -Content $currentRelease

if ($UpdateLocalBinary) {
  Write-Section '更新本机 4001 使用的二进制'
  if (Test-BinaryLocked $localBinaryPath) {
    Write-Host 'backend/qiantie-local.exe 正被运行中的进程占用，已跳过替换（不会结束 4001）。' -ForegroundColor Yellow
    Write-Host '需要更新时请先停止 4001，再执行：'
    Write-Host "  Copy-Item '$windowsBinary' '$localBinaryPath' -Force"
  } else {
    $previousCopy = Join-Path $releaseDir 'previous-qiantie-local.exe'
    if (Test-Path -LiteralPath $localBinaryPath) { Copy-Item -LiteralPath $localBinaryPath -Destination $previousCopy -Force }
    Copy-Item -LiteralPath $windowsBinary -Destination $localBinaryPath -Force
    $manifestPath = Join-Path $repo '.local-review/runtime-manifest.json'
    $localManifest = [ordered]@{
      gitSha = $shortSha
      frontendEntry = $frontendEntry
      builtAt = $builtAt
      binaryPath = $localBinaryPath
    }
    Write-JsonNoBom -Path $manifestPath -Content $localManifest
    Write-Host "已更新 $localBinaryPath（旧版备份：$previousCopy）"
  }
}

if ($Deploy) {
  Write-Section '7/8 上传与远程校验'
  $context = [pscustomobject]@{
    BuildId = $BuildId
    ArchiveDir = $archiveDir
    LinuxBinaryPath = $linuxBinary
    DataStageDir = $package.stageDir
    DataEntries = @($package.entries | ForEach-Object { [pscustomobject]@{ path = $_.path } })
    ShortSha = $shortSha
    ExpectedCatalog = $expectedCatalog.count
    ExpectedCapabilities = $expectedCapabilities
  }
  Invoke-RemoteDeploy -Context $context
} else {
  Write-Section '7/8 未上传（local-only）'
  Write-Host '已生成本地存档与发布包。上传请显式加 -Deploy -TargetHost 与 -RemoteDataDir。'
}

if (-not $KeepOldReleases -and (Test-Path -LiteralPath $releaseRoot)) {
  Write-Section '8/8 清理旧构建目录（保留最近 3 个）'
  Get-ChildItem -LiteralPath $releaseRoot -Directory |
    Where-Object { $_.FullName -ne $releaseDir } |
    Sort-Object LastWriteTime -Descending |
    Select-Object -Skip 2 |
    ForEach-Object {
      Write-Host "删除旧构建目录：$($_.FullName)"
      Remove-Item -LiteralPath $_.FullName -Recurse -Force
    }
}

Write-Section '完成'
Write-Host "buildId：$BuildId"
Write-Host "存档：$archiveDir"
Write-Host "构建产物：$releaseDir"
Write-Host "运行清单：$currentReleasePath"
Write-Host '提示：本脚本不会 git commit；如这批改动已稳定，可自行决定是否提交。'
