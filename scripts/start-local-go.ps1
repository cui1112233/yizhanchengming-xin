param(
  [switch]$CheckOnly,
  [string]$ListenAddr = $(if ($env:QIANTIE_GO_LISTEN_ADDR) { $env:QIANTIE_GO_LISTEN_ADDR } else { ':4001' })
)

$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent $PSScriptRoot

if ($CheckOnly) {
  & (Join-Path $PSScriptRoot 'check-local-go.ps1')
  exit $LASTEXITCODE
}

$binaryPath = Join-Path $repo 'backend/qiantie-local.exe'
$goSourceRoot = Join-Path $repo 'backend'
$needsBuild = -not (Test-Path $binaryPath)
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
    (Join-Path $repo 'public/novel-panel'),
    (Join-Path $repo 'public/batch-rewrite'),
    (Join-Path $repo 'public/assets'),
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
}

Push-Location (Join-Path $repo 'backend')
try {
  $env:QIANTIE_GO_LISTEN_ADDR = $ListenAddr
  # A local restart must always identify the checkout it is starting. Keeping
  # an inherited SHA here can make 4001 serve new files while reporting an old
  # release, which makes the runtime check misleading.
  $env:QIANTIE_RELEASE_SHA = (& git -C $repo rev-parse --short HEAD).Trim()
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
