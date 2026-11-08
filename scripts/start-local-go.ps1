param(
  [switch]$CheckOnly,
  # 临时验证/排障用：复用现有二进制直接启动，不重建前端与 Go。
  [switch]$SkipBuild,
  # 版本纪律：要求本次启动的二进制必须对应 docs/releases 下已存档的构建。
  [switch]$RequireArchivedRelease,
  # 后台常驻启动：用 Start-Process 独立进程 + 文件日志，避免父 PowerShell 退出后把服务带走。
  [switch]$Detached,
  [string]$ListenAddr = $(if ($env:QIANTIE_GO_LISTEN_ADDR) { $env:QIANTIE_GO_LISTEN_ADDR } else { ':4001' }),
  # 发布前置验证用：直接运行指定二进制和清单，不覆盖 backend/qiantie-local.exe。
  [string]$BinaryPath = '',
  [string]$ManifestPath = '',
  # 发布前置验证用：跳过“是否需要重建”的源码/资源时间戳推断，避免误重建候选二进制。
  [switch]$NoBuildCheck
)

$ErrorActionPreference = 'Stop'

function ConvertTo-LocalRuntimeEntryPath {
  param([string]$Entry)

  return ($Entry.TrimStart([char[]]@('/', '\')) -replace '/', [IO.Path]::DirectorySeparatorChar)
}

function Test-LocalRuntimeManifest {
  param(
    [object]$Manifest,
    [string]$ExpectedSha,
    [string]$ExpectedBinaryPath,
    [string]$FrontendDist,
    [string]$GoDist
  )

  if ($null -eq $Manifest -or [string]$Manifest.gitSha -ne $ExpectedSha) { return $false }
  $entry = [string]$Manifest.frontendEntry
  if ([string]::IsNullOrWhiteSpace($entry)) { return $false }
  $entryPath = ConvertTo-LocalRuntimeEntryPath $entry
  if ([IO.Path]::IsPathRooted($entryPath) -or $entryPath -match '(^|[\\/])\.\.([\\/]|$)') { return $false }
  if (-not (Test-Path -LiteralPath $ExpectedBinaryPath)) { return $false }
  try {
    $manifestBinary = [IO.Path]::GetFullPath([string]$Manifest.binaryPath)
    $expectedBinary = [IO.Path]::GetFullPath($ExpectedBinaryPath)
  } catch {
    return $false
  }
  if (-not $manifestBinary.Equals($expectedBinary, [StringComparison]::OrdinalIgnoreCase)) { return $false }
  return (Test-Path -LiteralPath (Join-Path $FrontendDist $entryPath)) -and
    (Test-Path -LiteralPath (Join-Path $GoDist $entryPath))
}

function Test-FrontendBuildRequired {
  param(
    [string]$SourceRoot,
    [string]$DistIndexPath,
    [string[]]$AdditionalInputs = @()
  )

  if (-not (Test-Path -LiteralPath $DistIndexPath)) { return $true }
  $builtAt = (Get-Item -LiteralPath $DistIndexPath).LastWriteTimeUtc
  $inputs = @()
  foreach ($inputPath in @($SourceRoot) + $AdditionalInputs) {
    if (-not (Test-Path -LiteralPath $inputPath)) { continue }
    $item = Get-Item -LiteralPath $inputPath
    if ($item.PSIsContainer) {
      $inputs += Get-ChildItem -LiteralPath $item.FullName -Recurse -File -ErrorAction SilentlyContinue
    } else {
      $inputs += $item
    }
  }
  return [bool]($inputs | Where-Object { $_.LastWriteTimeUtc -gt $builtAt } | Select-Object -First 1)
}

$repo = Split-Path -Parent $PSScriptRoot

# Slice 是 V12 批次路由的注册门槛：代码默认值是 3（config.FromEnv），
# Slice<1 时 registerV12IntakeRoutes 不会执行，/api/batch-factory/v12/batches/summary
# 会变成裸 404，前端表现就是“作品全没了”。注意 PowerShell 里字符串 "0"
# 是非空值，`if (-not $env:...)` 不会把它当成“未设置”，所以必须显式校验。
$sliceValue = 0
$rawSlice = [string]$env:QIANTIE_BATCH_FACTORY_V11_SLICE
if (-not [int]::TryParse($rawSlice, [ref]$sliceValue) -or $sliceValue -lt 1) {
  if (-not [string]::IsNullOrWhiteSpace($rawSlice)) {
    Write-Warning "QIANTIE_BATCH_FACTORY_V11_SLICE='$rawSlice' 会关闭 V12 批次路由（summary 返回 404）；本地启动已强制提升为 5。"
  }
  $sliceValue = 5
  $env:QIANTIE_BATCH_FACTORY_V11_SLICE = '5'
  # config.FromEnv 对 Slice>=4/5 强制要求生产与合并开关，这里一并补齐，
  # 否则强制提升后二进制会因缺少开关直接启动失败。
  $env:QIANTIE_BATCH_FACTORY_V11_PRODUCTION_ENABLED = '1'
  $env:QIANTIE_BATCH_FACTORY_V11_MERGE_ENABLED = '1'
  $env:QIANTIE_BATCH_FACTORY_V11_LOCAL_MERGE_ENABLED = '1'
}
# Slice 不足 4 时视频生产路由不会注册，此时 PRODUCTION_ENABLED=1 只是空转。
if ($sliceValue -lt 4 -and $env:QIANTIE_BATCH_FACTORY_V11_PRODUCTION_ENABLED -eq '1') {
  Write-Warning "当前 Slice=$sliceValue，视频生产/提交路由不会注册；如需视频请改用 scripts/start-local-go-review.ps1 -EnableBatchProduction（Slice 4）。"
}

# Go is the only supported application runtime. Keep historical Node files on
# disk for migration/rollback audit, but refuse to start Go alongside an
# accidentally enabled application Node service. The checker ignores Node
# processes owned by unrelated tools such as Codex or Photoshop.
$runtimeGuard = Join-Path $PSScriptRoot 'verify-no-node-runtime.ps1'
if (Test-Path -LiteralPath $runtimeGuard) {
  & $runtimeGuard -Strict | Out-Host
  if ($LASTEXITCODE -ne 0) {
    throw 'Legacy application Node process detected; Go startup was blocked to prevent two runtimes from serving the same application.'
  }
}

if ($CheckOnly) {
  & (Join-Path $PSScriptRoot 'check-local-go.ps1')
  exit $LASTEXITCODE
}

$binaryPath = if ([string]::IsNullOrWhiteSpace($BinaryPath)) {
  Join-Path $repo 'backend/qiantie-local.exe'
} else {
  [System.IO.Path]::GetFullPath($BinaryPath)
}
$goSourceRoot = Join-Path $repo 'backend'
$frontendRoot = Join-Path $repo 'frontend'
$frontendDist = Join-Path $repo 'frontend/dist'
$goDist = Join-Path $repo 'backend/web/dist'
$runtimeManifestPath = if ([string]::IsNullOrWhiteSpace($ManifestPath)) {
  Join-Path $repo '.local-review/runtime-manifest.json'
} else {
  [System.IO.Path]::GetFullPath($ManifestPath)
}
$expectedBuildSHA = (& git -C $repo rev-parse --short HEAD).Trim()
$needsBuild = -not (Test-Path $binaryPath)
$frontendBuildRequired = $false
if (-not $NoBuildCheck) {
$frontendBuildRequired = Test-FrontendBuildRequired `
  -SourceRoot (Join-Path $frontendRoot 'src') `
  -DistIndexPath (Join-Path $frontendDist 'index.html') `
  -AdditionalInputs @(
    (Join-Path $frontendRoot 'public'),
    (Join-Path $frontendRoot 'package.json'),
    (Join-Path $frontendRoot 'package-lock.json'),
    (Join-Path $frontendRoot 'vite.config.js'),
    (Join-Path $frontendRoot 'scripts/clean-dist.mjs')
  )
if ($frontendBuildRequired -and -not $SkipBuild) {
  & npm.cmd --prefix $frontendRoot run build
  if ($LASTEXITCODE -ne 0) {
    throw "Frontend build failed with exit code $LASTEXITCODE; refusing to start an old embedded page."
  }
  $needsBuild = $true
}
$runtimeManifest = $null
if (-not $needsBuild) {
  if (-not (Test-Path -LiteralPath $runtimeManifestPath)) {
    $needsBuild = $true
  } else {
    try {
      $runtimeManifest = Get-Content -LiteralPath $runtimeManifestPath -Raw | ConvertFrom-Json
    } catch {
      $needsBuild = $true
    }
  }
}
if (-not $needsBuild -and -not (Test-LocalRuntimeManifest -Manifest $runtimeManifest -ExpectedSha $expectedBuildSHA -ExpectedBinaryPath $binaryPath -FrontendDist $frontendDist -GoDist $goDist)) {
  $needsBuild = $true
}
if (-not $needsBuild) {
  $binaryTime = (Get-Item $binaryPath).LastWriteTimeUtc
  $latestGoSource = Get-ChildItem $goSourceRoot -Recurse -Filter '*.go' -File | Sort-Object LastWriteTimeUtc -Descending | Select-Object -First 1
  $needsBuild = $latestGoSource -and $latestGoSource.LastWriteTimeUtc -gt $binaryTime
}
# Go embeds the prepared frontend and workbench assets. A Go-only timestamp
# check is not enough: changing frontend/dist or public/* otherwise leaves the
# local 4001 process serving an older page until someone manually rebuilds it.
if (-not $needsBuild) {
  $assetRoots = @(
    (Join-Path $repo 'frontend/dist'),
    (Join-Path $repo 'frontend/public/assets'),
    (Join-Path $repo 'pets')
  ) | Where-Object { Test-Path $_ }
  $latestAsset = $assetRoots |
    ForEach-Object { Get-ChildItem -LiteralPath $_ -Recurse -File -ErrorAction SilentlyContinue } |
    Sort-Object LastWriteTimeUtc -Descending |
    Select-Object -First 1
  $needsBuild = $latestAsset -and $latestAsset.LastWriteTimeUtc -gt $binaryTime
}
}  # end of -NoBuildCheck guard
if ($SkipBuild) {
  if (-not (Test-Path -LiteralPath $binaryPath)) {
    throw "SkipBuild 已指定，但本地二进制不存在：$binaryPath；请先运行 scripts/build-go-local.ps1。"
  }
  Write-Output "SkipBuild：直接复用现有二进制 $binaryPath，不重建前端/Go。"
}
if ($needsBuild -and -not $SkipBuild) {
  & (Join-Path $PSScriptRoot 'build-go-local.ps1')
  if ($LASTEXITCODE -ne 0) {
    throw "Local Go build failed with exit code $LASTEXITCODE; refusing to start an existing binary."
  }
  if (-not (Test-Path -LiteralPath $binaryPath) -or -not (Test-Path -LiteralPath $runtimeManifestPath)) {
    throw 'Local Go build did not produce the required binary and runtime manifest; refusing to start an existing binary.'
  }
}

# 版本纪律：可选地要求当前二进制对应 docs/releases 下已存档的构建。
# 默认不校验，避免打断日常本地开发；-RequireArchivedRelease 用于正式验证环境。
$currentReleasePath = Join-Path $repo 'backend/current-release.json'
if (Test-Path -LiteralPath $currentReleasePath) {
  try {
    $currentRelease = Get-Content -LiteralPath $currentReleasePath -Raw -Encoding UTF8 | ConvertFrom-Json
  } catch {
    throw "无法解析 backend/current-release.json：$($_.Exception.Message)"
  }
  $releaseBuildId = [string]$currentRelease.buildId
  if ($releaseBuildId -notmatch '^[0-9]{8}-[0-9]{2}-[0-9a-fA-F]{8}$') {
    throw "backend/current-release.json 的 buildId 非法：$releaseBuildId"
  }
  $releaseArchive = Join-Path $repo (Join-Path 'docs/releases' $releaseBuildId)
  # 日常本地启动只要求 current-release.json 可解析；存档完整性仅在
  # -RequireArchivedRelease 的正式验证场景强制校验，避免本地被未归档构建挡住。
  if ($RequireArchivedRelease -and -not (Test-Path -LiteralPath (Join-Path $releaseArchive 'build-info.json'))) {
    throw "backend/current-release.json 指向的存档不存在或不完整：$releaseArchive"
  }
  if ($RequireArchivedRelease) {
    if ($needsBuild) {
      # needsBuild=true 表示本次启动会重新构建，产物不属于任何已存档构建。
      throw '当前源码需要重新构建，但重建产物不是任何已存档构建；请先运行 scripts/release-v88.ps1 生成存档。'
    }
    if ([string]::IsNullOrWhiteSpace([string]$currentRelease.localBinarySha256)) {
      throw 'backend/current-release.json 缺少 localBinarySha256，无法校验二进制。'
    }
    $actualHash = (Get-FileHash -LiteralPath $binaryPath -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($actualHash -ne ([string]$currentRelease.localBinarySha256).ToLowerInvariant()) {
      throw "二进制 SHA256 与存档不一致（当前 $actualHash，存档 $($currentRelease.localBinarySha256)）。"
    }
    Write-Output "发布批次校验通过：buildId=$releaseBuildId"
  } else {
    Write-Output "当前发布批次：$releaseBuildId（未启用 -RequireArchivedRelease 校验）"
  }
} elseif ($RequireArchivedRelease) {
  throw '未找到 backend/current-release.json；请先运行 scripts/release-v88.ps1。'
}

Push-Location (Join-Path $repo 'backend')
try {
  $env:QIANTIE_GO_LISTEN_ADDR = $ListenAddr
  # A local restart must always identify the checkout it is starting. Keeping
  # an inherited SHA here can make 4001 serve new files while reporting an old
  # release, which makes the runtime check misleading.
  $env:QIANTIE_RELEASE_SHA = (& git -C $repo rev-parse --short HEAD).Trim()
  $env:QIANTIE_RUNTIME_MANIFEST_PATH = $runtimeManifestPath
  # Slice 已在脚本入口统一校验并兜底为 3（>=1 才会注册 V12 批次路由）。
  if (-not $env:QIANTIE_BATCH_FACTORY_V11_DYNAMIC_TEXT_PROVIDER) {
    # Text credentials are selected from the account model catalog at request
    # time; do not put a key in this launcher.
    $env:QIANTIE_BATCH_FACTORY_V11_DYNAMIC_TEXT_PROVIDER = '1'
  }
  if (-not $env:QIANTIE_LEGACY_ACCOUNTS_PATH) {
    $env:QIANTIE_LEGACY_ACCOUNTS_PATH = Join-Path $repo 'data/system/accounts.json'
  }
  # The process runs from backend/, while the shared data files live at the
  # repository root. Keep local Go pointed at the same data as the Node app.
  $localData = Join-Path $repo 'data'
  if (-not $env:QIANTIE_PRESET_PATH) {
    $env:QIANTIE_PRESET_PATH = Join-Path $localData 'system/presets.json'
  }
  if (-not $env:QIANTIE_PERSONAL_PROMPT_PATH) {
    $env:QIANTIE_PERSONAL_PROMPT_PATH = Join-Path $localData 'system/script-constraint-prompts.json'
  }
  if (-not $env:QIANTIE_BATCH_FACTORY_AUTOMATION_PRESETS_PATH) {
    $env:QIANTIE_BATCH_FACTORY_AUTOMATION_PRESETS_PATH = Join-Path $localData 'batch-factory-v11-automation-presets.json'
  }
  if (-not $env:QIANTIE_BATCH_FACTORY_AUTOMATION_STATE_PATH) {
    $env:QIANTIE_BATCH_FACTORY_AUTOMATION_STATE_PATH = Join-Path $localData 'batch-factory-v11-automation.json'
  }
  if (-not $env:QIANTIE_SCRIPT_HISTORY_DIR) {
    $env:QIANTIE_SCRIPT_HISTORY_DIR = Join-Path $localData 'users'
  }
  if (-not $env:QIANTIE_MEMBER_SYSTEM_DIR) {
    $env:QIANTIE_MEMBER_SYSTEM_DIR = Join-Path $localData 'system'
  }
  if (-not $env:QIANTIE_MEMBER_AVATAR_DIR) {
    $env:QIANTIE_MEMBER_AVATAR_DIR = Join-Path $localData 'avatars'
  }
  if (-not $env:QIANTIE_LOCAL_EXECUTOR_ARTIFACT_DIR) {
    $env:QIANTIE_LOCAL_EXECUTOR_ARTIFACT_DIR = Join-Path $localData 'local-executor-artifacts'
  }
  if (-not $env:QIANTIE_AGENT_ASSET_LOCAL_DIR) {
    $env:QIANTIE_AGENT_ASSET_LOCAL_DIR = Join-Path $localData 'agent-assets'
  }
  if (-not $env:QIANTIE_LOCAL_EXECUTOR_DOWNLOADS_DIR) {
    $env:QIANTIE_LOCAL_EXECUTOR_DOWNLOADS_DIR = Join-Path $localData 'downloads'
  }
  if (-not $env:QIANTIE_GO_AUTH_ENABLED) {
    $env:QIANTIE_GO_AUTH_ENABLED = '1'
  }
  Write-Output ("本地启动配置：slice={0} production={1} merge={2} addr={3}" -f `
    $env:QIANTIE_BATCH_FACTORY_V11_SLICE, `
    $env:QIANTIE_BATCH_FACTORY_V11_PRODUCTION_ENABLED, `
    $env:QIANTIE_BATCH_FACTORY_V11_MERGE_ENABLED, `
    $ListenAddr)
  if ($Detached) {
    # 直接以前台方式调用 .exe 时，Go 写 stderr 会在这个脚本的
    # $ErrorActionPreference='Stop' 下变成终止性错误，父 PowerShell 一退出
    # 服务也跟着消失。后台启动改用独立进程，日志固定落到 .local-review/logs。
    $logDir = Join-Path $repo '.local-review/logs'
    New-Item -ItemType Directory -Force $logDir | Out-Null
    $portLabel = (($ListenAddr -replace '^\s*:', '').Trim() -replace '[^0-9A-Za-z._-]', '-')
    if ([string]::IsNullOrWhiteSpace($portLabel)) { $portLabel = '4001' }
    $stamp = Get-Date -Format 'yyyyMMdd-HHmmss'
    $stdoutLog = Join-Path $logDir ("qiantie-$portLabel-$stamp.out.log")
    $stderrLog = Join-Path $logDir ("qiantie-$portLabel-$stamp.err.log")
    $proc = Start-Process -FilePath $binaryPath -WorkingDirectory (Join-Path $repo 'backend') `
      -RedirectStandardOutput $stdoutLog -RedirectStandardError $stderrLog -WindowStyle Hidden -PassThru
    Write-Output ("已后台启动 pid={0} stdout={1} stderr={2}" -f $proc.Id, $stdoutLog, $stderrLog)
    exit 0
  }
  & .\qiantie-local.exe
  exit $LASTEXITCODE
} finally {
  Pop-Location
}
