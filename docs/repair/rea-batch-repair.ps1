param(
  [string]$ProjectRoot = 'F:\脚本测试\一战晟铭-v88',
  [string]$PatchPath = (Join-Path $PSScriptRoot 'rea-batch-repair.patch')
)
$ErrorActionPreference = 'Stop'
function Run-Git {
  param([string[]]$GitArgs)
  $savedPreference = $ErrorActionPreference
  try {
    $ErrorActionPreference = 'Continue'
    $result = & git -C $ProjectRoot @GitArgs 2>&1
    $gitExitCode = $LASTEXITCODE
  } finally { $ErrorActionPreference = $savedPreference }
  if ($gitExitCode -ne 0) { throw ($result -join "`n") }
  return $result
}
try {
  if (-not (Test-Path -LiteralPath (Join-Path $ProjectRoot '.git'))) { throw 'Project repository was not found.' }
  if (-not (Test-Path -LiteralPath $PatchPath)) { throw 'Keep rea-batch-repair.patch beside this script.' }
  if ((Run-Git -GitArgs @('branch','--show-current')).Trim() -ne 'v88') { throw 'Expected v88. No files were changed.' }
  & git -C $ProjectRoot diff --cached --quiet
  if ($LASTEXITCODE -ne 0) { throw 'There are staged changes. Preserve them before running this repair.' }
  Run-Git -GitArgs @('apply','--check','--index',$PatchPath) | Out-Host
  $baseline = (Run-Git -GitArgs @('rev-parse','HEAD')).Trim()
  $stamp = Get-Date -Format 'yyyyMMdd-HHmmss'
  $repairBranch = "fix/rea-batch-repair-$stamp"
  Run-Git -GitArgs @('branch',"backup/before-rea-batch-repair-$stamp",$baseline) | Out-Host
  Run-Git -GitArgs @('switch','-c',$repairBranch) | Out-Host
  Run-Git -GitArgs @('apply','--index',$PatchPath) | Out-Host
  Run-Git -GitArgs @('diff','--cached','--check') | Out-Host
  # Keep validation independent of the business database; never import .env here.
  $oldTestDsn = $env:QIANTIE_TEST_MYSQL84_DSN
  $env:QIANTIE_TEST_MYSQL84_DSN = ''
  try {
    Push-Location (Join-Path $ProjectRoot 'backend')
    try {
      & go test ./internal/videomanagementconnector ./internal/novelfetchworkshop ./internal/storage -count=1
      if ($LASTEXITCODE -ne 0) { throw 'Backend package tests failed; changes remain on the repair branch.' }
      & go test ./internal/httpapi -run 'Test(Contract|WorkerContract|ExternalSubmit|NovelFetch|Workshop|WebSubmit|ExpiredUpload|WorkerLogin|LoginRecovery|PausedAccountProbe)' -count=1
      if ($LASTEXITCODE -ne 0) { throw 'Workflow tests failed; changes remain on the repair branch.' }
    } finally { Pop-Location }
    Push-Location (Join-Path $ProjectRoot 'frontend')
    try {
      $testFiles = @(Get-ChildItem -LiteralPath 'src/user/pages' -Filter 'novel-fetch*.test.js' | ForEach-Object FullName)
      & node --test @testFiles
      if ($LASTEXITCODE -ne 0) { throw 'Frontend tests failed; changes remain on the repair branch.' }
    } finally { Pop-Location }
  } finally { $env:QIANTIE_TEST_MYSQL84_DSN = $oldTestDsn }
  Run-Git -GitArgs @('commit','-m','fix(novel-fetch): apply tested REA recovery and upload guards') | Out-Host
  Run-Git -GitArgs @('switch','v88') | Out-Host
  Run-Git -GitArgs @('merge','--ff-only',$repairBranch) | Out-Host
  & (Join-Path $ProjectRoot 'scripts/start-local-go-review.ps1') -ListenAddr ':4001' -EnableBatchProduction -EnableLocalMerge -Detached
  if ($LASTEXITCODE -ne 0) { throw 'Local startup failed. Read startup logs; do not clear task data.' }
  $health = Invoke-RestMethod 'http://127.0.0.1:4001/health' -TimeoutSec 10
  if ($health.ok -ne $true) { throw '4001 health check failed.' }
  $runtime = Invoke-RestMethod 'http://127.0.0.1:4001/api/runtime-build-info' -TimeoutSec 10
  $expected = (Run-Git -GitArgs @('rev-parse','HEAD')).Trim()
  if (-not $runtime.git_sha -or -not $expected.StartsWith([string]$runtime.git_sha)) { throw '4001 still runs another version. Do not report repair complete.' }
  Write-Host "4001 loaded $($runtime.git_sha). Single-book fetch/AI/upload/readback still requires acceptance."
} catch {
  Write-Error $_ -ErrorAction Continue
  Write-Host 'Repair stopped. No task data was deleted; inspect the current Git branch before continuing.'
  exit 1
}
