[CmdletBinding()]
param(
  [string]$Output = 'docs/migration/node-api-contract-inventory.json'
)

$ErrorActionPreference = 'Stop'
$repo = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$app = Join-Path $repo 'app.js'
if (-not (Test-Path -LiteralPath $app)) { throw "app.js not found: $app" }

# This is deliberately read-only. It records Express public mount points so a
# Node route cannot disappear silently during the Go cutover. A mount is not
# treated as a complete contract; the Go handler and focused test are recorded
# separately in the manifest for human review.
$goOwned = @(
  '/api/login', '/api/account-recovery', '/api/member', '/api/team-admin',
  '/api/account-admin', '/api/client-errors', '/api/admin',
  '/api/presets', '/api/script-constraint-prompts', '/api/novel-panel',
  '/api/novel-fetch', '/api/novel-fetch-upload', '/api/novel-fetch-workshop',
  '/api/batch-rewrite', '/api/batch-factory', '/api/script-video', '/api/config',
  '/api/models', '/api/tts', '/api/prompt', '/api/history', '/api/platform-projects',
  '/api/agent', '/api/shuihuo-production', '/api/storage', '/api/web-submit',
  '/api/local-executor', '/api/giant-material-executor', '/api'
)

$routes = [System.Collections.Generic.List[object]]::new()
$lineNo = 0
foreach ($line in (Get-Content -LiteralPath $app)) {
  $lineNo++
  if ($line -notmatch 'app\.use\(\s*[''\"](?<path>/api[^''\"]*)[''\"]') { continue }
  $path = $Matches.path
  $owner = 'node-legacy'
  $status = 'blocked'
  $handler = 'none'
  $test = 'add focused Go contract test before removal'
  $match = $goOwned | Where-Object { $path -eq $_ -or $path.StartsWith($_ + '/') } | Select-Object -First 1
  if ($path -eq '/api/account-recovery/purge/:username' -or $path -eq '/api/applications') { $match = $null }
  if ($match) {
    $owner = 'go'
    $status = 'shadow'
    $handler = "backend/internal/httpapi (mount $match)"
    $test = 'backend/internal/httpapi/*_test.go'
  }
  $routes.Add([ordered]@{
    method = '*'
    path = $path
    owner = $owner
    status = $status
    goHandler = $handler
    test = $test
    source = 'app.js'
    line = $lineNo
  })
}

$manifest = [ordered]@{
  generatedAt = (Get-Date).ToUniversalTime().ToString('o')
  purpose = 'Node public API mounts retained for Go cutover audit; shadow means Go exists but Node source remains for rollback.'
  routes = @($routes)
}
$target = Join-Path $repo $Output
$parent = Split-Path -Parent $target
New-Item -ItemType Directory -Force -Path $parent | Out-Null
$json = $manifest | ConvertTo-Json -Depth 8
[System.IO.File]::WriteAllText($target, $json, (New-Object System.Text.UTF8Encoding($false)))
Write-Output ("wrote {0} route mounts to {1}" -f $routes.Count, $Output)
