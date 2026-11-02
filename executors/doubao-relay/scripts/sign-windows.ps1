<#
.SYNOPSIS
  Windows 代码签名(spec 13「代码签名 / Windows」):用 signtool 给 MSI / NSIS 安装包签名并校验。

.DESCRIPTION
  证书来源(二选一):
    -CertThumbprint / $env:WINDOWS_CERT_THUMBPRINT       已安装在 Cert:\CurrentUser\My 的证书(EV 证书 / 硬件令牌)
    $env:WINDOWS_CERTIFICATE (+ WINDOWS_CERTIFICATE_PASSWORD)  base64 编码的 .pfx(CI secrets)
  时间戳服务器默认 http://timestamp.digicert.com(-TimestampUrl 覆盖)。
  只签外层安装包;安装包里的 doubao-relay.exe 已在打包时封进去,不会被本脚本签到
  (要签内层 exe 需在打包阶段用 Tauri 的 bundle.windows.signCommand,见 README「发布」)。

.PARAMETER BundleDir
  cargo tauri build 的 bundle 目录,签名其中 msi\*.msi 与 nsis\*.exe。
  例:target\x86_64-pc-windows-msvc\release\bundle

.PARAMETER Files
  直接指定要签名的文件(与 -BundleDir 二选一)。

.PARAMETER DryRun
  只打印 signtool 命令(密码打码),不执行。

.EXAMPLE
  pwsh scripts/sign-windows.ps1 -BundleDir target\x86_64-pc-windows-msvc\release\bundle
#>
[CmdletBinding()]
param(
  [string]$BundleDir,
  [string[]]$Files,
  [string]$CertThumbprint = $env:WINDOWS_CERT_THUMBPRINT,
  [string]$TimestampUrl = 'http://timestamp.digicert.com',
  [switch]$DryRun
)

$ErrorActionPreference = 'Stop'

function Fail([string]$Message, [int]$Code = 2) {
  Write-Error "sign-windows: $Message" -ErrorAction Continue
  exit $Code
}

function Find-SignTool {
  $cmd = Get-Command signtool.exe -ErrorAction SilentlyContinue
  if ($cmd) { return $cmd.Source }
  $kits = Join-Path ${env:ProgramFiles(x86)} 'Windows Kits\10\bin'
  if (Test-Path $kits) {
    $found = Get-ChildItem -Path $kits -Recurse -Filter signtool.exe -ErrorAction SilentlyContinue |
      Where-Object { $_.FullName -match '\\x64\\' } |
      Sort-Object FullName -Descending |
      Select-Object -First 1
    if ($found) { return $found.FullName }
  }
  return $null
}

# ---- 要签的文件 ---------------------------------------------------------------
$targets = @()
if ($Files) {
  $targets = $Files
} elseif ($BundleDir) {
  foreach ($pattern in @('msi\*.msi', 'nsis\*.exe')) {
    $targets += @(Get-ChildItem -Path (Join-Path $BundleDir $pattern) -ErrorAction SilentlyContinue |
      ForEach-Object { $_.FullName })
  }
} else {
  Fail '需要 -BundleDir 或 -Files'
}
if ($targets.Count -eq 0) { Fail "没有找到要签名的 .msi / .exe($BundleDir)" }
foreach ($t in $targets) {
  if (-not (Test-Path $t)) { Fail "文件不存在: $t" }
}

# ---- 证书参数 -----------------------------------------------------------------
$certArgs = @()
$shownCertArgs = @()
$tempPfx = $null
if ($CertThumbprint) {
  $certArgs = @('/sha1', $CertThumbprint)
  $shownCertArgs = $certArgs
} elseif ($env:WINDOWS_CERTIFICATE) {
  $tempPfx = Join-Path ([System.IO.Path]::GetTempPath()) ("doubao-relay-sign-{0}.pfx" -f [guid]::NewGuid())
  if (-not $DryRun) {
    [System.IO.File]::WriteAllBytes($tempPfx, [Convert]::FromBase64String($env:WINDOWS_CERTIFICATE))
  }
  $certArgs = @('/f', $tempPfx)
  $shownCertArgs = @('/f', $tempPfx)
  if ($env:WINDOWS_CERTIFICATE_PASSWORD) {
    $certArgs += @('/p', $env:WINDOWS_CERTIFICATE_PASSWORD)
    $shownCertArgs += @('/p', '***')
  }
} else {
  Fail '缺少证书:设置 WINDOWS_CERT_THUMBPRINT,或 WINDOWS_CERTIFICATE(base64 .pfx)+ WINDOWS_CERTIFICATE_PASSWORD'
}

$signtool = Find-SignTool
if (-not $signtool) {
  if ($DryRun) { $signtool = 'signtool.exe' } else { Fail '找不到 signtool.exe(安装 Windows SDK)' }
}

try {
  foreach ($t in $targets) {
    $common = @('sign', '/fd', 'sha256', '/tr', $TimestampUrl, '/td', 'sha256', '/d', 'doubao-relay')
    Write-Host ('+ "{0}" {1} {2} "{3}"' -f $signtool, ($common -join ' '), ($shownCertArgs -join ' '), $t)
    if (-not $DryRun) {
      & $signtool @common @certArgs $t
      if ($LASTEXITCODE -ne 0) { Fail "签名失败: $t" 1 }
    }
    Write-Host ('+ "{0}" verify /pa /v "{1}"' -f $signtool, $t)
    if (-not $DryRun) {
      & $signtool verify /pa /v $t
      if ($LASTEXITCODE -ne 0) { Fail "签名校验失败: $t" 1 }
    }
  }
} finally {
  if ($tempPfx -and (Test-Path $tempPfx)) { Remove-Item -Force $tempPfx }
}

Write-Host ("sign-windows: 已签名 {0} 个文件" -f $targets.Count)
