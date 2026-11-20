param(
  [switch]$NoStart
)

$ErrorActionPreference = 'Stop'
$workbenchRoot = Split-Path -Parent $PSScriptRoot
$frontendRoot = Join-Path $workbenchRoot '前台'
$apiRoot = Join-Path $workbenchRoot 'api'
$apiDist = Join-Path $apiRoot 'web\dist'
$repositoryRoot = Split-Path -Parent $workbenchRoot
$outputRoot = Join-Path $repositoryRoot '_local-archive-20261028\magic-cut'
$binaryPath = Join-Path $outputRoot 'magic-cut-4040.exe'

pnpm --dir $frontendRoot install --registry 'https://registry.npmjs.org'
if ($LASTEXITCODE -ne 0) { throw '前端依赖安装失败' }

$nodePath = (Get-Command node -ErrorAction Stop).Source
$nodeMajor = [int](& $nodePath -p "process.versions.node.split('.')[0]")
if ($nodeMajor -ge 24) {
  $node22 = Get-ChildItem -Path (Join-Path $env:USERPROFILE '.workbuddy\binaries\node\versions\22*\node.exe') -File -ErrorAction SilentlyContinue |
    Sort-Object FullName -Descending |
    Select-Object -First 1
  if (-not $node22) {
    throw '当前 Node 24 与 Windows 下的 Rollup 原生模块不兼容，请安装 Node 22 LTS 后重试'
  }
  $nodePath = $node22.FullName
}
Push-Location $frontendRoot
try {
  & $nodePath (Join-Path $frontendRoot 'node_modules\vite\bin\vite.js') build
  if ($LASTEXITCODE -ne 0) { throw '前端构建失败' }
} finally {
  Pop-Location
}

$resolvedWorkbench = [IO.Path]::GetFullPath($workbenchRoot)
$resolvedApiDist = [IO.Path]::GetFullPath($apiDist)
if (-not $resolvedApiDist.StartsWith($resolvedWorkbench, [StringComparison]::OrdinalIgnoreCase)) {
  throw "拒绝清理工作台之外的目录：$resolvedApiDist"
}
New-Item -ItemType Directory -Force -Path $apiDist | Out-Null
Get-ChildItem -LiteralPath $apiDist -Force | Remove-Item -Recurse -Force
Get-ChildItem -LiteralPath (Join-Path $frontendRoot 'dist') -Force |
  Copy-Item -Destination $apiDist -Recurse -Force

New-Item -ItemType Directory -Force -Path $outputRoot | Out-Null
Push-Location $apiRoot
try {
  go test ./...
  if ($LASTEXITCODE -ne 0) { throw 'Go 测试失败' }
  go build -o $binaryPath ./cmd/magic-cut-4040
  if ($LASTEXITCODE -ne 0) { throw 'Go 构建失败' }
} finally {
  Pop-Location
}

if (-not $NoStart) {
  Start-Process -FilePath $binaryPath -WorkingDirectory $workbenchRoot -WindowStyle Hidden
  Write-Host "魔剪已启动：http://127.0.0.1:4040"
}
