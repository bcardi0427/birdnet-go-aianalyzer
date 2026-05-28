@echo off
setlocal

REM Start BirdNET-Go from this repository directory.
REM Adds MSYS2 UCRT64 to PATH so required native DLL dependencies can be found.
set "APP_DIR=%~dp0"
set "PATH=C:\msys64\ucrt64\bin;%APP_DIR%;%PATH%"

REM Attempt to auto-detect FFmpeg on Windows and expose it to the BirdNET-Go process
for /f "tokens=*" %%p in ('where ffmpeg 2^>nul') do (
    if exist "%%p" (
        set "BIRDNET_FFMPEG_PATH=%%p"
        echo Detected FFmpeg at %%p
        goto :BREAK_DETECT
    )
)
:BREAK_DETECT
if not defined BIRDNET_FFMPEG_PATH (
    echo FFmpeg not found in PATH. You may set BIRDNET_FFMPEG_PATH to point to ffmpeg.exe.
)

REM Enable verbose health logging for debugging RTSP health issues when requested
if not defined BIRDNET_HEALTH_VERBOSE (
    set "BIRDNET_HEALTH_VERBOSE=0"
)
if "%BIRDNET_HEALTH_VERBOSE%" == "1" (
    echo Enabling verbose RTSP/health logs
    set "HEALTH_VERBOSE_OPT=--verbose-health"
)

cd /d "%APP_DIR%"

if not exist "birdnet-go.exe" (
    echo ERROR: birdnet-go.exe was not found in %APP_DIR%
    echo Build it first, then run this script again.
    pause
    exit /b 1
)

set "FIRST_ARG=%~1"
if not defined FIRST_ARG (
    set "RUN_ARGS=serve"
    set "IS_SERVE=1"
) else (
    set "FIRST_CHAR=%FIRST_ARG:~0,1%"
    if "%FIRST_CHAR%" == "-" (
        set "RUN_ARGS=serve %*"
        set "IS_SERVE=1"
    ) else (
        set "RUN_ARGS=%*"
        if /i "%FIRST_ARG%" == "serve" set "IS_SERVE=1"
        if /i "%FIRST_ARG%" == "realtime" set "IS_SERVE=1"
    )
)

if defined IS_SERVE (
    if defined HEALTH_VERBOSE_OPT (
        set "RUN_ARGS=%RUN_ARGS% %HEALTH_VERBOSE_OPT%"
    )
)

REM Check if birdnet-go.local resolves
ping -n 1 -w 100 birdnet-go.local >nul 2>&1
if errorlevel 1 (
    echo [TIP] To access the dashboard via http://birdnet-go.local:8080
    echo       run 'setup_hosts.ps1' as Administrator once to configure local DNS.
    echo.
)

echo Starting BirdNET-Go server...
echo Press Ctrl+C to stop.
"%APP_DIR%birdnet-go.exe" %RUN_ARGS%


set "EXIT_CODE=%ERRORLEVEL%"
echo.
echo BirdNET-Go exited with code %EXIT_CODE%.
pause
exit /b %EXIT_CODE%
