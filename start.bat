@echo off
cd /d "%~dp0"
title Drone Platform

rem Load .env
if exist .env (
    for /f "tokens=1,2 delims==" %%a in (.env) do (
        set %%a=%%b 2>nul
    )
)

rem Defaults
if "%AUTH_SECRET%"=="" set AUTH_SECRET=drone-platform-dev-secret-32bytes!
if "%ADMIN_DEV_MODE%"=="" set ADMIN_DEV_MODE=true
if "%ENV%"=="" set ENV=development

if "%1"=="" goto menu
if "%1"=="--api" goto api
if "%1"=="--mp" goto mp
if "%1"=="--all" goto all
goto api

:menu
echo.
echo   [1] Start API server
echo   [2] Setup miniprogram dev
echo   [3] Start all
echo.
set /p choice="Choose (1/2/3): "
if "%choice%"=="1" goto api
if "%choice%"=="2" goto mp
if "%choice%"=="3" goto all
goto api

:api
echo.
echo === Build API ===
go build -o drone-api.exe .\cmd\api
if %ERRORLEVEL% NEQ 0 (
    echo Build FAILED!
    pause
    exit /b 1
)
echo Build OK
echo.
echo === Starting Server ===
echo API:        http://localhost:8080
echo Admin UI:   http://localhost:8080/admin
echo Health:     http://localhost:8080/healthz
echo OpenAPI:    docs/openapi.yaml
echo Press Ctrl+C to stop
echo.
start "" http://localhost:8080/admin
drone-api.exe
goto end

:mp
echo.
echo === Miniprogram Setup ===
if not exist "miniprogram\node_modules" (
    echo Installing npm deps...
    cd miniprogram
    call npm install
    cd ..
)
echo npm deps ready
echo.
echo Next steps:
echo   1. Open WeChat DevTools
echo   2. Import project - select "miniprogram" folder
echo   3. Enter AppID - click Compile
echo.
echo Backend API: http://localhost:8080
echo (run "start --api" or "start --all" first)
pause
goto end

:all
echo.
echo === Start All ===
if not exist "miniprogram\node_modules" (
    echo Installing npm deps...
    cd miniprogram
    call npm install
    cd ..
)
echo Building API...
go build -o drone-api.exe .\cmd\api
if %ERRORLEVEL% NEQ 0 (
    echo Build FAILED!
    pause
    exit /b 1
)
echo.
echo API:        http://localhost:8080
echo Admin UI:   http://localhost:8080/admin
echo Miniprogram: WeChat DevTools - import miniprogram folder
echo Press Ctrl+C to stop
echo.
start "" http://localhost:8080/admin
drone-api.exe

:end
