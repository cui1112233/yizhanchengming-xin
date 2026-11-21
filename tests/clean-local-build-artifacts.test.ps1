$ErrorActionPreference = 'Stop'

$repo = Join-Path ([System.IO.Path]::GetTempPath()) ("clean-artifacts-test-" + [guid]::NewGuid().ToString('N'))
$script = Join-Path (Split-Path -Parent $PSScriptRoot) 'scripts\maintenance\clean-local-build-artifacts.ps1'

try {
  New-Item -ItemType Directory -Force -Path (Join-Path $repo 'executors\doubao-relay\target\debug') | Out-Null
  New-Item -ItemType Directory -Force -Path (Join-Path $repo 'backend') | Out-Null
  Set-Content -LiteralPath (Join-Path $repo 'executors\doubao-relay\target\debug\cache.bin') -Value 'cache'
  Set-Content -LiteralPath (Join-Path $repo 'backend\qiantie-local.exe') -Value 'current'
  Set-Content -LiteralPath (Join-Path $repo 'backend\qiantie-local.previous.exe') -Value 'previous'
  Set-Content -LiteralPath (Join-Path $repo 'backend\qiantie-old-next.exe') -Value 'old'
  Set-Content -LiteralPath (Join-Path $repo 'backend\source.go') -Value 'package backend'

  $report = & $script -RepoRoot $repo | ConvertFrom-Json
  if ($report.Count -ne 2) { throw "report candidate count=$($report.Count), want 2" }
  if (-not (Test-Path -LiteralPath (Join-Path $repo 'backend\qiantie-old-next.exe'))) { throw 'report-only mode deleted old executable' }

  $applied = & $script -RepoRoot $repo -Apply | ConvertFrom-Json
  if (($applied | Where-Object action -eq 'removed').Count -ne 2) { throw 'apply did not remove both build artifacts' }
  if (Test-Path -LiteralPath (Join-Path $repo 'executors\doubao-relay\target')) { throw 'Rust target still exists' }
  if (Test-Path -LiteralPath (Join-Path $repo 'backend\qiantie-old-next.exe')) { throw 'old executable still exists' }
  foreach ($kept in @('backend\qiantie-local.exe', 'backend\qiantie-local.previous.exe', 'backend\source.go')) {
    if (-not (Test-Path -LiteralPath (Join-Path $repo $kept))) { throw "required file removed: $kept" }
  }
  Write-Output 'clean-local-build-artifacts test passed'
} finally {
  if (Test-Path -LiteralPath $repo) { Remove-Item -LiteralPath $repo -Recurse -Force }
}
