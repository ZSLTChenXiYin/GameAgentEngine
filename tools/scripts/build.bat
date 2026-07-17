@echo off
setlocal enabledelayedexpansion

REM Change working directory to project root (2 levels up from tools/scripts/)
pushd "%~dp0..\.." || (echo Failed to locate project root & exit /b 1)

REM ============ CONFIG ============
set ALL_PLATFORMS=windows/amd64 windows/arm64 linux/amd64 linux/arm64 darwin/amd64 darwin/arm64
set SOURCE_DIR=tools\source
set OUTPUT_DIR=dist
set VERSION=v0.5.0
REM ================================

REM Detect current platform
set CURRENT_OS=windows
set CURRENT_ARCH=amd64
if "%PROCESSOR_ARCHITECTURE%"=="x86" set CURRENT_ARCH=386
if "%PROCESSOR_ARCHITECTURE%"=="ARM64" set CURRENT_ARCH=arm64
set CURRENT_PLATFORM=%CURRENT_OS%/%CURRENT_ARCH%

REM Parse command line
if "%~1"=="" (
  set TARGETS=%CURRENT_PLATFORM%
) else (
  set TARGETS=%*
  echo %* | findstr /i "all/all" >nul
  if not errorlevel 1 set TARGETS=%ALL_PLATFORMS%
)

set LDFLAGS=-s -w -X "github.com/ZSLTChenXiYin/GameAgentEngine/internal/version.Version=%VERSION%" -X "github.com/ZSLTChenXiYin/GameAgentEngine/cmd/gameagentdevcli.devCliVersion=%VERSION%"

echo =========================================
echo  GameAgentEngine Build Script
echo  Version: %VERSION%
echo  Current: %CURRENT_PLATFORM%
echo  Targets: %TARGETS%
echo  Output:  %OUTPUT_DIR%\
echo =========================================
echo.

echo Generating Creator component metadata...
go run .\tools\scripts\generate_component_meta.go
if errorlevel 1 popd & exit /b 1
echo.

for %%p in (%TARGETS%) do (
  for /f "tokens=1,2 delims=/" %%a in ("%%p") do (
    set OUT_DIR=%OUTPUT_DIR%\GameAgentEngine-%%a-%%b-%VERSION%
    set EXT=
    if "%%a"=="windows" set EXT=.exe

    if not exist "!OUT_DIR!" mkdir "!OUT_DIR!"

    echo [%%a/%%b] Building GameAgentEngine...
    set GOOS=%%a
    set GOARCH=%%b
    set CGO_ENABLED=0
    go build -trimpath -ldflags="%LDFLAGS%" -o "!OUT_DIR!\GameAgentEngine!EXT!" .\cmd\gameagentengine\
    if errorlevel 1 popd & exit /b 1

    echo [%%a/%%b] Building GameAgentDevCli...
    go build -trimpath -ldflags="%LDFLAGS%" -o "!OUT_DIR!\GameAgentDevCli!EXT!" .\cmd\gameagentdevcli\
    if errorlevel 1 popd & exit /b 1

    echo [%%a/%%b] Building GameAgentWorker...
    go build -trimpath -ldflags="%LDFLAGS%" -o "!OUT_DIR!\GameAgentWorker!EXT!" .\cmd\gameagentworker\
    if errorlevel 1 popd & exit /b 1

    if exist gameagentengine.conf.yaml copy gameagentengine.conf.yaml "!OUT_DIR!" >nul
    if exist "%SOURCE_DIR%" xcopy /E /I /Y "%SOURCE_DIR%\*" "!OUT_DIR!" >nul

    REM Inject version into packaged Creator asset without mutating source files
    if exist "!OUT_DIR!\web\GameAgentCreator\js\version.js" (
        powershell -Command "(Get-Content '!OUT_DIR!\web\GameAgentCreator\js\version.js') -replace 'CREATOR_MIN_COMPATIBLE = \"v[0-9]+\.[0-9]+\.[0-9]+\"', 'CREATOR_MIN_COMPATIBLE = \"%VERSION%\"' | Set-Content '!OUT_DIR!\web\GameAgentCreator\js\version.js'"
    )

    echo [%%a/%%b] -^> !OUT_DIR!\
    dir /a-d /b "!OUT_DIR!"

    echo [%%a/%%b] Packaging to zip...
    powershell -Command "Compress-Archive -Path '!OUT_DIR!\*' -DestinationPath '!OUT_DIR!.zip' -Force"
    echo.
  )
)

echo =========================================
echo  Build complete.
for %%p in (%TARGETS%) do (
  for /f "tokens=1,2 delims=/" %%a in ("%%p") do (
    echo   %OUTPUT_DIR%\GameAgentEngine-%%a-%%b-%VERSION%\
  )
)
echo =========================================
popd
endlocal



