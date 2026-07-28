@echo off
chcp 65001 >nul 2>&1
setlocal enabledelayedexpansion

title Drone Platform - Starting...

cd /d "%~dp0"

echo.
echo   ============================================
echo     Drone Platform - Dev Environment Start
echo   ============================================
echo.

:: ---- Check Tools ----
echo  [1/5] Checking tools...

where go >nul 2>&1
if %errorlevel% neq 0 (
    echo  [ERROR] Go not found
    pause && exit /b 1
)

where node >nul 2>&1
if %errorlevel% neq 0 (
    echo  [ERROR] Node.js not found
    pause && exit /b 1
)

echo         Go:       OK
echo         Node.js:  OK
echo.

:: ---- Start PostgreSQL ----
echo  [2/5] Starting PostgreSQL...

docker info >nul 2>&1
if %errorlevel% neq 0 (
    echo  [WARN] Docker not running, skipping DB
    goto :skip_db
)

docker compose up -d db >nul 2>&1
if %errorlevel% neq 0 (
    docker-compose up -d db >nul 2>&1
)

echo         Waiting for PostgreSQL...
set retry=0
:wait_db
timeout /t 2 /nobreak >nul
docker compose exec db pg_isready -U drone -d drone_platform >nul 2>&1
if !errorlevel! equ 0 (
    echo         PostgreSQL ready
    goto :db_ok
)
set /a retry+=1
if !retry! lss 15 goto :wait_db
echo  [WARN] PostgreSQL startup timeout
:db_ok
:skip_db
echo.

:: ---- Start Go API ----
echo  [3/5] Starting Go API (port 8080)...

start "Drone-API" cmd /c "cd /d "%~dp0" && title Drone-API-8080 && go run ./cmd/api"

echo         Waiting for Go API...
set retry=0
:wait_api
timeout /t 1 /nobreak >nul
curl -s http://localhost:8080/api/v1/health >nul 2>&1
if !errorlevel! equ 0 (
    echo         Go API ready
    goto :api_ok
)
set /a retry+=1
if !retry! lss 15 goto :wait_api
echo  [WARN] Go API may still be compiling...
:api_ok
echo.

:: ---- Start Vue Frontend ----
echo  [4/5] Starting Vue Frontend (port 5173)...

if not exist "%~dp0frontend\node_modules" (
    echo         Installing frontend deps...
    cd /d "%~dp0frontend"
    call npm install
    cd /d "%~dp0"
)

start "Drone-Vue" cmd /c "cd /d "%~dp0frontend" && title Drone-Vue-5173 && npm run dev"

echo         Waiting for Vite...
set retry=0
:wait_vite
timeout /t 1 /nobreak >nul
curl -s http://localhost:5173 >nul 2>&1
if !errorlevel! equ 0 (
    echo         Vue Frontend ready
    goto :vite_ok
)
set /a retry+=1
if !retry! lss 20 goto :wait_vite
echo  [WARN] Vite may still be compiling...
:vite_ok
echo.

:: ---- Done ----
echo  [5/5] All services started!
echo.
echo   ============================================
echo     Service                   URL
echo   --------------------------------------------
echo     Admin (recommended)       http://localhost:5173/admin
echo     Vue Frontend              http://localhost:5173
echo     Go API                    http://localhost:8080
echo     PostgreSQL                localhost:5433
echo   ============================================
echo.
echo   Run stop.bat to stop all services.
echo.

pause >nul
endlocal
