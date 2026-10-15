param(
  [string]$Version = "0.1.0",
  [string]$ControlURL = "http://115.190.156.223",
  [string]$UpdatePublicKey = "",
  [string]$OutputRoot = "$(Join-Path $PSScriptRoot '..\dist\windows-x64')"
)

$ErrorActionPreference = "Stop"
$moduleRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$stage = Join-Path $OutputRoot "stage"
$binary = Join-Path $stage "GiantMaterialExecutor.exe"

if (Test-Path $stage) { Remove-Item -Recurse -Force $stage }
New-Item -ItemType Directory -Force -Path $stage, (Join-Path $stage 'worker') | Out-Null

Push-Location $moduleRoot
try {
  # 国内代理，避免 rsrc/go get 下载模块超时
  $env:GOPROXY = "https://goproxy.cn,direct"
  # Embed the icon into the EXE so Explorer/taskbar show a real logo
  # (generates cmd/giant-material-executor/rsrc_windows_amd64.syso).
  & go run github.com/akavel/rsrc@v0.10.2 -ico (Join-Path $moduleRoot 'assets\icon.ico') -arch amd64 -o (Join-Path $moduleRoot 'cmd\giant-material-executor\rsrc_windows_amd64.syso')
  if ($LASTEXITCODE -ne 0) { throw "rsrc failed" }
  $env:GOOS = "windows"
  $env:GOARCH = "amd64"
  $env:CGO_ENABLED = "0"
  # windowsgui makes double-clicking the EXE open the native pairing window
  # instead of briefly showing a console window and then disappearing.
  # -X main.bakedPublicAPIURL 烧入控制服务地址，用户无需手填。
  & go build -trimpath -ldflags "-s -w -H=windowsgui -X main.version=$Version -X main.bakedPublicAPIURL=$ControlURL -X main.bakedUpdatePublicKey=$UpdatePublicKey" -o $binary ./cmd/giant-material-executor
  if ($LASTEXITCODE -ne 0) { throw "go build failed" }
  & go build -trimpath -ldflags "-s -w -H=windowsgui" -o (Join-Path $stage 'GiantMaterialExecutorUpdater.exe') ./cmd/giant-material-executor-updater
  if ($LASTEXITCODE -ne 0) { throw "updater build failed" }
  Copy-Item (Join-Path $moduleRoot 'worker\ocr_worker.py') (Join-Path $stage 'worker\ocr_worker.py')
  Copy-Item (Join-Path $moduleRoot 'worker\requirements-lock.txt') (Join-Path $stage 'worker\requirements-lock.txt')
  Copy-Item (Join-Path $moduleRoot 'portable\README-Windows.txt') (Join-Path $stage 'README-Windows.txt')
  Copy-Item (Join-Path $moduleRoot 'portable\start-giant-material-executor.cmd') (Join-Path $stage 'start-giant-material-executor.cmd')
} finally {
  Pop-Location
}

$forbidden = Get-ChildItem -Recurse -File $stage | Where-Object {
  $_.Name -match '(?i)(model|paddle|onnx|.pt$|.pdmodel$|.pdiparams$|.bin$)'
}
if ($forbidden) {
  $names = ($forbidden | ForEach-Object FullName) -join ", "
  throw "OCR runtime/model files must not be included in the installer stage: $names"
}

$portableZip = Join-Path $OutputRoot "GiantMaterialExecutor-windows-x64.zip"
if (Test-Path $portableZip) { Remove-Item -Force $portableZip }
Compress-Archive -Path (Join-Path $stage '*') -DestinationPath $portableZip -CompressionLevel Optimal
Write-Output $portableZip

# Generate latest.json (version manifest for executor self-update) and sync it with
# the zip into frontend/public so every Node deployment ships the current release.
$zipHash = (Get-FileHash -Algorithm SHA256 $portableZip).Hash.ToLowerInvariant()
$latestJson = @{ version = $Version; sha256 = $zipHash } | ConvertTo-Json -Compress
$latestPath = Join-Path $OutputRoot "latest.json"
[System.IO.File]::WriteAllText($latestPath, $latestJson, (New-Object System.Text.UTF8Encoding($false)))
$publicDir = Join-Path $moduleRoot '..\frontend\public\downloads\giant-material-executor'
if (Test-Path $publicDir) {
  Copy-Item $portableZip $publicDir -Force
  Copy-Item $latestPath $publicDir -Force
  Write-Output "synced zip+latest.json to $publicDir"
}
Write-Output $latestPath

$nsis = Get-Command makensis -ErrorAction SilentlyContinue
if ($null -eq $nsis) {
  Write-Warning "makensis was not found; executable stage is ready but installer was not built."
  exit 0
}

$installer = Join-Path $OutputRoot "GiantMaterialExecutor-$Version.exe"
& $nsis.Source "/DPRODUCT_VERSION=$Version" "/DINPUT_STAGE=$stage" "/DOUTPUT_FILE=$installer" (Join-Path $moduleRoot 'installer\giant-material-executor.nsi')
if ($LASTEXITCODE -ne 0) { throw "makensis failed" }
Write-Output $installer
