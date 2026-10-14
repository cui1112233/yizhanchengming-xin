<#
V88 代码直传打包脚本（在你的 Windows 电脑上跑）。

做的事：
  1. 取当前 Git 版本号（SHA）；
  2. 把 Go 后端交叉编译成 1 个 Linux 单文件；
  3. 构建前端并同步到 Go embed 目录；
  4. 组装 Go 单文件成品目录并压缩成一个 tar.gz。

全程只在你电脑本地操作，不碰服务器、不碰 Docker、不碰国外镜像。
产物：release/v88-direct-<sha>.tar.gz

用法（在仓库根目录）：
  powershell -ExecutionPolicy Bypass -File deploy/v88-direct/build-direct-release.ps1
#>
[CmdletBinding()]
param(
    [string]$RepoRoot = ''
)

if (-not $RepoRoot) { $RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path }
$ErrorActionPreference = 'Stop'
function Step($msg) { Write-Host "==> $msg" -ForegroundColor Cyan }
function Die($msg) { Write-Host "错误：$msg" -ForegroundColor Red; exit 1 }

Set-Location $RepoRoot

# 1) 工具与版本检查 ----------------------------------------------------------
foreach ($c in 'git', 'go', 'npm', 'tar') {
    if (-not (Get-Command $c -ErrorAction SilentlyContinue)) { Die "没找到 $c，请先安装后再跑。" }
}
$sha = (git rev-parse HEAD 2>$null)
if (-not $sha) { Die '当前目录不是 Git 仓库。' }
$sha = $sha.Trim()
if (-not (git rev-parse --git-dir 2>$null)) { Die '必须在 v88 仓库里运行。' }
$dirty = (git status --porcelain 2>$null | Measure-Object -Line).Lines
if ($dirty -gt 0) {
    Die "工作区有 $dirty 处未提交改动。请先把改动提交并推送到 v88，再打包（保证线上版本能追到 Git 编号）。"
}
Write-Host "打包版本 SHA = $sha" -ForegroundColor Green

$staging = Join-Path $RepoRoot ".direct-staging"
$goOut   = Join-Path $staging 'go\qiantie'
if (Test-Path $staging) { Remove-Item $staging -Recurse -Force }
New-Item -ItemType Directory -Force -Path (Join-Path $staging 'go') | Out-Null

# 2) 交叉编译 Go（Linux amd64 单文件，与线上镜像同架构）---------------------
Step '交叉编译 Go 后端（linux/amd64 单文件）'
Push-Location (Join-Path $RepoRoot 'backend')
try {
    $env:GOOS = 'linux'; $env:GOARCH = 'amd64'; $env:CGO_ENABLED = '0'
    go build -trimpath -o $goOut ./cmd/qiantie
    if ($LASTEXITCODE -ne 0) { Die 'Go 编译失败。' }
} finally {
    Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED -ErrorAction SilentlyContinue
    Pop-Location
}

# 3) 前端依赖与构建 ----------------------------------------------------------
$fe = Join-Path $RepoRoot 'frontend'
if (-not (Test-Path (Join-Path $fe 'node_modules'))) {
    Step '安装前端依赖（仅首次较慢）'
    Push-Location $fe; try { npm ci; if ($LASTEXITCODE -ne 0) { Die '前端依赖安装失败。' } } finally { Pop-Location }
}
Step '构建前端'
Push-Location $fe; try { npm run build; if ($LASTEXITCODE -ne 0) { Die '前端构建失败。' } } finally { Pop-Location }

# 4) 把前端和兼容工作台同步到 Go embed 目录，再编译 Go 单文件。
Step '准备 Go embed 静态资源'
& (Join-Path $RepoRoot 'scripts\build-go-local.ps1') -SkipGoBuild
if ($LASTEXITCODE -ne 0) { Die 'Go embed 静态资源准备失败。' }

# 6) 压缩成品 ----------------------------------------------------------------
Step '压缩成品 tar.gz'
$releaseDir = Join-Path $RepoRoot 'release'
New-Item -ItemType Directory -Force -Path $releaseDir | Out-Null
$tarball = Join-Path $releaseDir "v88-direct-$sha.tar.gz"
if (Test-Path $tarball) { Remove-Item $tarball -Force }
Push-Location $staging
try {
    tar -czf $tarball go
    if ($LASTEXITCODE -ne 0) { Die '打包失败。' }
} finally { Pop-Location }

$sizeMB = [math]::Round((Get-Item $tarball).Length / 1MB, 1)
Write-Host ""
Write-Host "完成！成品包：$tarball ($sizeMB MB)" -ForegroundColor Green
Write-Host "下一步：上传并在服务器激活（由部署流程通过 SSH 执行）。" -ForegroundColor Green
