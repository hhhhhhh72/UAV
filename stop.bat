@echo off
echo Stopping Drone Platform...
taskkill /F /IM drone-api.exe >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo Server stopped.
) else (
    echo Server is not running.
)
