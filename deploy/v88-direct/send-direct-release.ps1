<#
V88 代码直传上传与激活脚本（在你的 Windows 电脑上跑）。

前提：先跑过 build-direct-release.ps1 拿到成品包。
本脚本把成品包传到服务器、解压、切换软链并只重启 Go API。
全程不下载国外镜像，通常 1 分钟内完成。

用法：
  powershell -ExecutionPolicy Bypass -File deploy/v88-direct\send-direct-release.ps1 `
    -ComputerName 115.190.156.223

密码用系统弹窗输入，不会写进任何文件。
#>
[CmdletBinding()]
param(
    [string]$ComputerName = '115.190.156.223',
    [string]$UserName = 'root',
    [string]$RepoRoot = '',
    [string]$Tarball = ''
)

if (-not $RepoRoot) { $RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path }
$ErrorActionPreference = 'Stop'
function Die($m) { Write-Host "错误：$m" -ForegroundColor Red; exit 1 }

if (-not (Get-Module -ListAvailable -Name Posh-SSH)) {
    Die '没装 Posh-SSH 模块。请先以管理员身份运行：Install-Module Posh-SSH -Scope CurrentUser'
}
Import-Module Posh-SSH

# 1) 定位成品包与 SHA --------------------------------------------------------
if (-not $Tarball) {
    $latest = Get-ChildItem (Join-Path $RepoRoot 'release\v88-direct-*.tar.gz') -ErrorAction SilentlyContinue |
        Sort-Object LastWriteTime -Descending | Select-Object -First 1
    if (-not $latest) { Die '没找到成品包，请先运行 build-direct-release.ps1。' }
    $Tarball = $latest.FullName
}
if (-not (Test-Path $Tarball)) { Die "成品包不存在：$Tarball" }
if ($Tarball -match 'v88-direct-([0-9a-f]{7,40})\.tar\.gz$') { $sha = $Matches[1] } else { Die '成品包文件名里没有 SHA，请用脚本生成的包。' }
$sizeMB = [math]::Round((Get-Item $Tarball).Length / 1MB, 1)
Write-Host "待发布版本 = $sha ($sizeMB MB)，目标 = $UserName@$ComputerName" -ForegroundColor Cyan

$cred = Get-Credential -UserName $UserName -Message "输入服务器 $ComputerName 的登录密码"
if (-not $cred) { Die '已取消。' }

$directRoot = '/opt/qiantie/v88/direct'
$session = $null
try {
    $session = New-SSHSession -ComputerName $ComputerName -Credential $cred -AcceptKey -ConnectionTimeout 20 -ErrorAction Stop

    # 2) 首次启用准备：目录 + 服务器侧直接启动脚本 -----------------------------
    Write-Host '==> 准备服务器直传目录与脚本' -ForegroundColor Cyan
    $prep = @"
set -e
mkdir -p $directRoot/releases
"@
    $r = Invoke-SSHCommand -SessionId $session.SessionId -Command $prep -TimeOut 30
    if ($r.ExitStatus -ne 0) { Die "服务器目录准备失败：$($r.ErrorOutput -join ' ')" }

    Set-SCPItem -ComputerName $ComputerName -Credential $cred -AcceptKey `
        -Path (Join-Path $RepoRoot 'deploy\v88-direct\activate-direct-release.sh') -Destination $directRoot -Force
    Set-SCPItem -ComputerName $ComputerName -Credential $cred -AcceptKey `
        -Path (Join-Path $RepoRoot 'deploy\v88-direct\runtime-release-manifest.sh') -Destination $directRoot -Force
    Set-SCPItem -ComputerName $ComputerName -Credential $cred -AcceptKey `
        -Path (Join-Path $RepoRoot 'deploy\v88-direct\qiantie.service') -Destination $directRoot -Force

    # 3) 上传成品包 -----------------------------------------------------------
    Write-Host "==> 上传成品包（$sizeMB MB，走你本机到服务器的直连，不经过国外镜像站）" -ForegroundColor Cyan
    Set-SCPItem -ComputerName $ComputerName -Credential $cred -AcceptKey `
        -Path $Tarball -Destination $directRoot -Force

    # 4) 解压到 releases/<sha> 并激活 -----------------------------------------
    Write-Host '==> 解压并切换版本、重启服务（失败会自动回滚）' -ForegroundColor Cyan
    $activate = @"
set -e
cd $directRoot
rm -rf releases/$sha
mkdir -p releases/$sha
tar -xzf $(Split-Path $Tarball -Leaf) -C releases/$sha
    chmod +x releases/$sha/go/qiantie activate-direct-release.sh runtime-release-manifest.sh
    export QIANTIE_SERVICE_TEMPLATE="$directRoot/qiantie.service"
    bash $directRoot/activate-direct-release.sh $sha
"@
    $r = Invoke-SSHCommand -SessionId $session.SessionId -Command $activate -TimeOut 300
    $r.Output | ForEach-Object { Write-Host $_ }
    if ($r.ExitStatus -ne 0) {
        $r.ErrorOutput | ForEach-Object { Write-Host $_ -ForegroundColor Yellow }
        Die '服务器激活失败（已自动回滚到上一版，请检查上面日志）。'
    }
    Write-Host "直传发布成功：$sha" -ForegroundColor Green
}
finally {
    if ($session) { Remove-SSHSession -SessionId $session.SessionId | Out-Null }
}
