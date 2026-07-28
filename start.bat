@echo off
chcp 65001 >nul 2>&1
setlocal enabledelayedexpansion
title Drone Platform
cd /d "%~dp0"

echo   ============================================
echo     Drone Platform - Dev Start
echo   ============================================
echo.

:: ---- Step 1: Check node/go ----
echo  [1/3] Checking tools...
where go >nul 2>&1 || (echo [ERROR] Go not found && pause && exit /b 1)
where node >nul 2>&1 || (echo [ERROR] Node.js not found && pause && exit /b 1)

:: ---- Step 2: Database ----
echo  [2/3] Database...

if defined DATABASE_URL (
    echo         Using DATABASE_URL: %DATABASE_URL%
    goto :skip_docker
)

echo         Using local JSON storage (dev mode)
echo         For shared DB: set DATABASE_URL=postgres://...
echo         ^> neon.tech (free PostgreSQL)

:skip_docker

:: ---- Step 3: Start API & Frontend ----
echo  [3/3] Starting services...
echo.

set AUTH_SECRET=drone-platform-dev-secret-32bytes!
set APP_ENV=development
set ADMIN_DEV_MODE=true

start "Drone-API" cmd /c "cd /d "%~dp0" && set AUTH_SECRET=drone-platform-dev-secret-32bytes! && set APP_ENV=development && set ADMIN_DEV_MODE=true && if defined DATABASE_URL set DATABASE_URL=%DATABASE_URL% && go run ./cmd/api"
timeout /t 5 /nobreak >nul

start "Drone-Vue" cmd /c "cd /d "%~dp0frontend" && npm run dev"
timeout /t 5 /nobreak >nul

echo   ============================================
echo     Admin:  http://localhost:5173/admin
echo     API:    http://localhost:8080
echo   ============================================
echo   Run stop.bat to quit.
pause >nul
