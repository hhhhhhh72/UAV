@echo off
chcp 65001 >nul 2>&1
setlocal enabledelayedexpansion
cd /d "%~dp0"

echo.
echo   ============================================
echo     Drone Platform - Stop All Services
echo   ============================================
echo.

:: ---- helper: kill by port using temp file (no pipe hang) ----
call :KillPort 8080 "Go API"
call :KillPort 5173 "Vue Frontend"

:: ---- Stop PostgreSQL ----
echo  [3/4] Stopping PostgreSQL...
docker compose down 2>nul || docker-compose down 2>nul
echo         Done
echo.

:: ---- Final cleanup ----
echo  [4/4] Cleaning up...
taskkill /f /im drone-platform.exe >nul 2>&1
taskkill /f /im go.exe >nul 2>&1
echo         Done
echo.

echo   ============================================
echo     All services stopped.
echo   ============================================
echo.
timeout /t 2 /nobreak >nul
exit /b

:: ============================================
::  KillPort - find PID on port and kill it
::  Uses temp file to avoid pipe hang
:: ============================================
:KillPort
set port=%~1
set label=%~2
echo  Stopping %label% (port %port%)...

:: Write netstat output to temp file (avoids pipe hang)
set tmpfile=%TEMP%\stop-port-%port%.txt
netstat -ano > "%tmpfile%" 2>nul

:: Read from temp file
set found=0
for /f "tokens=5" %%a in ('type "%tmpfile%" 2^>nul ^| findstr ":%port% "') do (
    set pid=%%a
    if not "!pid!"=="" (
        taskkill /f /pid !pid! >nul 2>&1
        set found=1
    )
)

:: Fallback: kill by window title
if %found%==0 (
    taskkill /fi "WINDOWTITLE eq *%label%*" /f >nul 2>&1
)

del "%tmpfile%" 2>nul
echo         Done
echo.
exit /b
