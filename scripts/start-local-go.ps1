param(
  [switch]$CheckOnly,
  [string]$ListenAddr = $(if ($env:QIANTIE_GO_LISTEN_ADDR) { $env:QIANTIE_GO_LISTEN_ADDR } else { ':4001' })
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

$binaryPath = Join-Path $repo 'backend/qiantie-local.exe'
$goSourceRoot = Join-Path $repo 'backend'
$frontendRoot = Join-Path $repo 'frontend'
$frontendDist = Join-Path $repo 'frontend/dist'
$goDist = Join-Path $repo 'backend/web/dist'
$runtimeManifestPath = Join-Path $repo '.local-review/runtime-manifest.json'
$expectedBuildSHA = (& git -C $repo rev-parse --short HEAD).Trim()
$needsBuild = -not (Test-Path $binaryPath)
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
if ($frontendBuildRequired) {
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
if ($needsBuild) {
  & (Join-Path $PSScriptRoot 'build-go-local.ps1')
  if ($LASTEXITCODE -ne 0) {
    throw "Local Go build failed with exit code $LASTEXITCODE; refusing to start an existing binary."
  }
  if (-not (Test-Path -LiteralPath $binaryPath) -or -not (Test-Path -LiteralPath $runtimeManifestPath)) {
    throw 'Local Go build did not produce the required binary and runtime manifest; refusing to start an existing binary.'
  }
}

Push-Location (Join-Path $repo 'backend')
try {
  $env:QIANTIE_GO_LISTEN_ADDR = $ListenAddr
  # A local restart must always identify the checkout it is starting. Keeping
  # an inherited SHA here can make 4001 serve new files while reporting an old
  # release, which makes the runtime check misleading.
  $env:QIANTIE_RELEASE_SHA = (& git -C $repo rev-parse --short HEAD).Trim()
  $env:QIANTIE_RUNTIME_MANIFEST_PATH = $runtimeManifestPath
  if (-not $env:QIANTIE_BATCH_FACTORY_V11_SLICE) {
    # Slice 3 keeps H3 director/audio/compile routes on Go. Higher slices
    # still require explicit video/merge/provider settings.
    $env:QIANTIE_BATCH_FACTORY_V11_SLICE = '3'
  }
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
  & .\qiantie-local.exe
  exit $LASTEXITCODE
} finally {
  Pop-Location
}
