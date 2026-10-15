@echo off
setlocal
cd /d "%~dp0"
set "LOG=%APPDATA%\YizhanShengming\GiantMaterialExecutor\executor.log"
if not exist "%APPDATA%\YizhanShengming\GiantMaterialExecutor" mkdir "%APPDATA%\YizhanShengming\GiantMaterialExecutor" >nul 2>&1
echo [%date% %time%] starting GiantMaterialExecutor.exe >> "%LOG%"
set "PAIRING_CODE="
if not exist "%APPDATA%\YizhanShengming\GiantMaterialExecutor\credential.bin" (
  echo 首次启动：请先在网页设置中生成巨量素材配对码。
  set /p "PAIRING_CODE=请输入配对码（直接回车可跳过）： "
)
if defined PAIRING_CODE set "GIANT_MATERIAL_EXECUTOR_PAIRING_CODE=%PAIRING_CODE%"
start "Giant Material Executor" /min "%~dp0GiantMaterialExecutor.exe"
timeout /t 2 /nobreak >nul
powershell -NoProfile -ExecutionPolicy Bypass -Command "$ok = Test-NetConnection -ComputerName 127.0.0.1 -Port 17861 -InformationLevel Quiet; if (-not $ok) { exit 1 }"
if errorlevel 1 (
  echo.
  echo 巨量素材执行器没有监听 127.0.0.1:17861。
  echo 这是启动日志：
  echo ----------------------------------------
  type "%LOG%"
  echo ----------------------------------------
  echo 请把上面的错误信息发给技术支持。
  pause
  exit /b 1
)
echo 巨量素材执行器已在后台启动。
echo 日志文件：%LOG%
endlocal
