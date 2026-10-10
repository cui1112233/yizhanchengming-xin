param(
  [switch]$SkipGoBuild
)

$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent $PSScriptRoot
$frontendDist = Join-Path $repo 'frontend/dist'
$goDist = Join-Path $repo 'backend/web/dist'
$petsSource = Join-Path $repo 'pets'
$petsDist = Join-Path $goDist 'pets'

if (-not (Test-Path (Join-Path $frontendDist 'index.html'))) {
  throw "Frontend build output is missing: $frontendDist"
}

New-Item -ItemType Directory -Force $goDist | Out-Null
Get-ChildItem -LiteralPath $goDist -Force | Where-Object { $_.Name -ne '.gitkeep' } | Remove-Item -Recurse -Force
Copy-Item -Path (Join-Path $frontendDist '*') -Destination $goDist -Recurse -Force

if (-not (Test-Path (Join-Path $petsSource 'stacky/spritesheet.webp'))) {
  throw "Pet assets are missing: $petsSource"
}

New-Item -ItemType Directory -Force $petsDist | Out-Null
Copy-Item -Path (Join-Path $petsSource '*') -Destination $petsDist -Recurse -Force

if (-not $SkipGoBuild) {
  Push-Location (Join-Path $repo 'backend')
  try {
    go build -o qiantie-local.exe ./cmd/qiantie
  } finally {
    Pop-Location
  }
}

Write-Output "Go static assets prepared: $goDist"
