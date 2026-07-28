@echo off
chcp 65001 >nul 2>&1
title Drone Platform - All in One
cd /d "%~dp0"

echo   ============================================
echo     Drone Platform - 一键启动
echo   ============================================
echo.

echo  [1/4] Starting PostgreSQL...
docker compose up -d db >nul 2>&1
timeout /t 3 /nobreak >nul
echo         PostgreSQL ready (port 5433)

echo  [2/4] Starting cpolar tunnel...
start "cpolar" cmd /c "cpolar tcp 5433"
echo         cpolar starting (check http://127.0.0.1:4042 for public address)
timeout /t 5 /nobreak >nul

echo  [3/4] Starting Go API (port 8080)...
set AUTH_SECRET=drone-platform-dev-secret-32bytes!
set APP_ENV=development
set ADMIN_DEV_MODE=true
set DATABASE_URL=postgresql://drone:drone_secret@localhost:5433/drone_platform?sslmode=disable
start "Drone-API" cmd /c "cd /d "%~dp0" && set AUTH_SECRET=drone-platform-dev-secret-32bytes! && set APP_ENV=development && set ADMIN_DEV_MODE=true && set DATABASE_URL=postgresql://drone:drone_secret@localhost:5433/drone_platform?sslmode=disable && go run ./cmd/api"
timeout /t 8 /nobreak >nul

echo  [4/4] Starting Vue Frontend (port 5173)...
cd /d "%~dp0frontend"
if not exist node_modules call npm install
start "Drone-Vue" cmd /c "cd /d "%~dp0frontend" && npm run dev"
cd /d "%~dp0"

echo.
echo   ============================================
echo     All services started!
echo     Admin: http://localhost:5173/admin
echo     cpolar: http://127.0.0.1:4042
echo   ============================================
echo.
echo   Run stop.bat to quit.
pause >nul
