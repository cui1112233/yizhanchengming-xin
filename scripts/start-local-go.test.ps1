$ErrorActionPreference = 'Stop'

$launcherPath = Join-Path $PSScriptRoot 'start-local-go.ps1'
$tokens = $null
$parseErrors = $null
$ast = [System.Management.Automation.Language.Parser]::ParseFile($launcherPath, [ref]$tokens, [ref]$parseErrors)
if ($parseErrors.Count -gt 0) {
  throw "Unable to parse start-local-go.ps1: $($parseErrors[0].Message)"
}

$helper = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'ConvertTo-LocalRuntimeEntryPath' }, $true)
if ($null -eq $helper) {
  throw 'ConvertTo-LocalRuntimeEntryPath helper is missing.'
}

Invoke-Expression $helper.Extent.Text
$separator = [IO.Path]::DirectorySeparatorChar
$expected = "assets${separator}app.js"
foreach ($entry in @('/assets/app.js', '\assets\app.js')) {
  $actual = ConvertTo-LocalRuntimeEntryPath $entry
  if ($actual -ne $expected) {
    throw "Expected '$entry' to normalize to '$expected', got '$actual'."
  }
}

$manifestHelper = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Test-LocalRuntimeManifest' }, $true)
if ($null -eq $manifestHelper) {
  throw 'Test-LocalRuntimeManifest helper is missing.'
}
Invoke-Expression $manifestHelper.Extent.Text

$testRoot = Join-Path ([IO.Path]::GetTempPath()) ('qiantie-runtime-test-' + [guid]::NewGuid().ToString('N'))
try {
  $frontendDist = Join-Path $testRoot 'frontend-dist'
  $goDist = Join-Path $testRoot 'go-dist'
  $binaryPath = Join-Path $testRoot 'qiantie-local.exe'
  New-Item -ItemType Directory -Path (Join-Path $frontendDist 'assets'),(Join-Path $goDist 'assets') -Force | Out-Null
  Set-Content -LiteralPath (Join-Path $frontendDist 'assets/current.js') -Value 'frontend'
  Set-Content -LiteralPath (Join-Path $goDist 'assets/current.js') -Value 'embedded'
  Set-Content -LiteralPath $binaryPath -Value 'binary'
  $manifest = [pscustomobject]@{ gitSha = 'abc123'; frontendEntry = '/assets/current.js'; binaryPath = $binaryPath }

  if (-not (Test-LocalRuntimeManifest -Manifest $manifest -ExpectedSha 'abc123' -ExpectedBinaryPath $binaryPath -FrontendDist $frontendDist -GoDist $goDist)) {
    throw 'A matching current runtime manifest must be accepted.'
  }
  if (Test-LocalRuntimeManifest -Manifest $manifest -ExpectedSha 'stale' -ExpectedBinaryPath $binaryPath -FrontendDist $frontendDist -GoDist $goDist) {
    throw 'A stale runtime manifest SHA must be rejected.'
  }
  $manifest.binaryPath = Join-Path $testRoot 'old-qiantie.exe'
  if (Test-LocalRuntimeManifest -Manifest $manifest -ExpectedSha 'abc123' -ExpectedBinaryPath $binaryPath -FrontendDist $frontendDist -GoDist $goDist) {
    throw 'A manifest pointing at an old binary must be rejected.'
  }
} finally {
  if (Test-Path -LiteralPath $testRoot) { Remove-Item -LiteralPath $testRoot -Recurse -Force }
}

$frontendHelper = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Test-FrontendBuildRequired' }, $true)
if ($null -eq $frontendHelper) {
  throw 'Test-FrontendBuildRequired helper is missing.'
}
Invoke-Expression $frontendHelper.Extent.Text

$testRoot = Join-Path ([IO.Path]::GetTempPath()) ('qiantie-frontend-test-' + [guid]::NewGuid().ToString('N'))
try {
  $sourceRoot = Join-Path $testRoot 'src'
  $indexPath = Join-Path $testRoot 'dist/index.html'
  New-Item -ItemType Directory -Path $sourceRoot,(Split-Path $indexPath -Parent) -Force | Out-Null
  $sourcePath = Join-Path $sourceRoot 'App.jsx'
  Set-Content -LiteralPath $sourcePath -Value 'source'
  Set-Content -LiteralPath $indexPath -Value 'built'
  (Get-Item $sourcePath).LastWriteTimeUtc = [datetime]::UtcNow.AddMinutes(-2)
  (Get-Item $indexPath).LastWriteTimeUtc = [datetime]::UtcNow.AddMinutes(-1)
  if (Test-FrontendBuildRequired -SourceRoot $sourceRoot -DistIndexPath $indexPath) {
    throw 'An up-to-date frontend build must be reused.'
  }
  (Get-Item $sourcePath).LastWriteTimeUtc = [datetime]::UtcNow
  if (-not (Test-FrontendBuildRequired -SourceRoot $sourceRoot -DistIndexPath $indexPath)) {
    throw 'A frontend source change must require a rebuild.'
  }
} finally {
  if (Test-Path -LiteralPath $testRoot) { Remove-Item -LiteralPath $testRoot -Recurse -Force }
}

$legacyLauncher = Get-Content -LiteralPath (Join-Path $PSScriptRoot 'start-local-go-slice4-temp.ps1') -Raw
foreach ($forbidden in @('QIANTIE_MYSQL_DSN', 'QIANTIE_BRIDGE_SECRET', '.build-verify')) {
  if ($legacyLauncher.Contains($forbidden)) {
    throw "Legacy launcher still contains forbidden runtime input: $forbidden"
  }
}
if (-not $legacyLauncher.Contains('start-local-go-review.ps1')) {
  throw 'Legacy launcher must delegate to the canonical Go review launcher.'
}

Write-Output 'Local runtime entry path normalization passed.'
