param(
  [string]$ReleaseJson = '',
  [string]$BaseUrl = 'https://yzcm.site',
  [int]$TimeoutSec = 20
)

$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent $PSScriptRoot
if ([string]::IsNullOrWhiteSpace($ReleaseJson)) {
  $ReleaseJson = Join-Path $repo '.release/go-embedded/release.json'
}
if (-not (Test-Path -LiteralPath $ReleaseJson)) {
  throw "发布记录不存在：$ReleaseJson"
}

$release = Get-Content -LiteralPath $ReleaseJson -Raw | ConvertFrom-Json
$base = $BaseUrl.TrimEnd('/')
$remote = Invoke-RestMethod -Uri "$base/api/runtime-build-info" -TimeoutSec $TimeoutSec

$failures = New-Object System.Collections.Generic.List[string]

if ($release.git_sha -ne $remote.git_sha) {
  $failures.Add("git_sha 不一致：release=$($release.git_sha) production=$($remote.git_sha)")
}
if (-not [string]::IsNullOrWhiteSpace($remote.runtime_manifest_sha)) {
  $a = [string]$release.git_sha
  $b = [string]$remote.runtime_manifest_sha
  if (-not ($a.StartsWith($b) -or $b.StartsWith($a))) {
    $failures.Add("runtime_manifest_sha 不一致：release=$a production=$b")
  }
}
$remoteEntry = [string]$remote.frontend_entry
$releaseEntry = [string]$release.frontend_entry
if ($remoteEntry -eq '') {
  Write-Warning '生产未上报 frontend_entry（runtime-manifest 未包含入口），改为校验首页引用。'
} elseif ($releaseEntry -ne $remoteEntry) {
  $failures.Add("frontend_entry 不一致：release=$releaseEntry production=$remoteEntry")
}
if ($releaseEntry -ne '') {
  $homePage = Invoke-WebRequest -Uri "$base/" -UseBasicParsing -TimeoutSec $TimeoutSec
  if ($homePage.Content -notmatch [regex]::Escape($releaseEntry)) {
    $failures.Add("首页未引用发布包入口：release=$releaseEntry")
  }
}

Write-Host "release.json：$ReleaseJson"
Write-Host "前端源码状态：$($release.frontend_source_state)；入口：$($release.frontend_entry)"
Write-Host "生产：git_sha=$($remote.git_sha) manifest_sha=$($remote.runtime_manifest_sha) frontend_entry=$($remote.frontend_entry)"

if ($failures.Count -gt 0) {
  foreach ($item in $failures) { Write-Host "FAIL: $item" }
  exit 1
}
Write-Host 'PASS: 发布记录与生产构建信息一致。'
