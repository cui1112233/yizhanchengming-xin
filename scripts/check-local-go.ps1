param(
  [string]$BaseUrl = $(if ($env:QIANTIE_GO_BASE_URL) { $env:QIANTIE_GO_BASE_URL } else { 'http://127.0.0.1:4001' })
)

$ErrorActionPreference = 'Stop'
$BaseUrl = $BaseUrl.TrimEnd('/')
$failed = $false

function Check-Url([string]$Path, [int]$ExpectedStatus) {
  try {
    $response = Invoke-WebRequest -Uri ($BaseUrl + $Path) -UseBasicParsing -TimeoutSec 5
    if ([int]$response.StatusCode -ne $ExpectedStatus) {
      Write-Output "FAIL $Path returned $($response.StatusCode), expected $ExpectedStatus"
      $script:failed = $true
      return
    }
    Write-Output "PASS $Path ($($response.StatusCode))"
  } catch {
    Write-Output "FAIL ${Path}: $($_.Exception.Message)"
    $script:failed = $true
  }
}

Check-Url '/health' 200
Check-Url '/' 200

if ($failed) { exit 1 }
Write-Output "Local Go runtime checks passed: $BaseUrl"
