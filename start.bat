@echo off
title Drone Platform - Dev
cd /d "%~dp0"

set ADMIN_DEV_MODE=true
set AUTH_SECRET=dev-secret-key-must-be-32-bytes-long
set CORS_ORIGINS=http://localhost:5173

echo.
echo ============================================
echo   Drone Industry Platform v1.0
echo ============================================
echo.

echo [1/4] Checking backend...
if not exist drone-api.exe (
    echo   Building...
    go build -o drone-api.exe .\cmd\api
    if errorlevel 1 (
        echo   COMPILE FAILED
        pause
        exit /b 1
    )
    echo   Build OK
)
echo   drone-api.exe found

echo [2/4] Checking frontend...
if not exist "frontend\node_modules\vant" (
    echo   Installing npm packages...
    cd frontend
    call npm install
    cd ..
    if errorlevel 1 (
        echo   NPM INSTALL FAILED
        pause
        exit /b 1
    )
    echo   Install OK
)
echo   node_modules found

echo [3/4] Starting API server...
taskkill /f /im drone-api.exe 2>nul 1>nul

:: Write env to temp script so CMD passes it to the child process
echo @echo off > "%TEMP%\drone-api-start.bat"
echo cd /d %~dp0 >> "%TEMP%\drone-api-start.bat"
echo set ADMIN_DEV_MODE=true >> "%TEMP%\drone-api-start.bat"
echo set AUTH_SECRET=dev-secret-key-must-be-32-bytes-long >> "%TEMP%\drone-api-start.bat"
echo set CORS_ORIGINS=http://localhost:5173 >> "%TEMP%\drone-api-start.bat"
echo drone-api.exe >> "%TEMP%\drone-api-start.bat"

start "API-Server" "%TEMP%\drone-api-start.bat"

echo   Waiting for API...
:wait_api
timeout /t 2 /nobreak >nul
curl -s -o nul http://localhost:8080/healthz 2>nul
if errorlevel 1 goto wait_api
echo   API ready (port 8080)

echo [4/4] Starting frontend dev server...
start "Frontend" cmd /c "cd /d %~dp0frontend && npm run dev"

echo   Waiting for frontend...
:wait_fe
timeout /t 2 /nobreak >nul
curl -s -o nul http://localhost:5173 2>nul
if errorlevel 1 goto wait_fe
echo   Frontend ready (port 5173)

echo.
echo ============================================
echo   ALL DONE
echo.
echo   Frontend : http://localhost:5173
echo   Admin    : http://localhost:5173/admin
echo   Swagger  : http://localhost:8080/swagger/index.html
echo ============================================
echo.
echo   Close windows or run stop.bat to stop.
pause
