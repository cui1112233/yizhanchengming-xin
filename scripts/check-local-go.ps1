param(
  [string]$BaseUrl = $(if ($env:QIANTIE_GO_BASE_URL) { $env:QIANTIE_GO_BASE_URL } else { 'http://127.0.0.1:4001' }),
  [string]$ExpectedSHA = $(if ($env:QIANTIE_EXPECTED_GO_SHA) { $env:QIANTIE_EXPECTED_GO_SHA } else { (git rev-parse --short HEAD).Trim() })
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
	$status = 0
	if ($_.Exception.Response -and $_.Exception.Response.StatusCode) {
		$status = [int]$_.Exception.Response.StatusCode
	}
	if ($status -eq $ExpectedStatus) {
		Write-Output "PASS $Path ($status)"
		return
	}
    Write-Output "FAIL ${Path}: $($_.Exception.Message)"
    $script:failed = $true
  }
}

Check-Url '/health' 200
Check-Url '/' 200
Check-Url '/novel-fetch' 200
Check-Url '/novel-panel' 404
Check-Url '/shuihuo-production' 200
Check-Url '/settings' 200

try {
  $build = Invoke-RestMethod -Uri ($BaseUrl + '/api/runtime-build-info') -TimeoutSec 5
  if (-not $build.git_sha -or $build.service -ne 'go-api') {
    Write-Output 'FAIL /api/runtime-build-info returned an invalid Go build identity'
    $failed = $true
  } elseif ($ExpectedSHA -and $build.git_sha -ne $ExpectedSHA) {
    Write-Output "FAIL /api/runtime-build-info is $($build.git_sha), expected $ExpectedSHA"
    $failed = $true
  } else {
    Write-Output "PASS /api/runtime-build-info ($($build.git_sha))"
  }
} catch {
  Write-Output "FAIL /api/runtime-build-info: $($_.Exception.Message)"
  $failed = $true
}

if ($failed) { exit 1 }
Write-Output "Local Go runtime checks passed: $BaseUrl"
