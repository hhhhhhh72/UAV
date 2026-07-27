@echo off
echo Starting drone-platform API...

set AUTH_SECRET=drone-platform-dev-secret-32bytes!
set APP_ENV=development
set ADMIN_DEV_MODE=true

echo Building...
go build -o api.exe .\cmd\api
if %errorlevel% neq 0 (
    echo [FAIL] Build failed
    pause
    exit /b 1
)

echo Starting on :8080 ...
start "drone-api" api.exe
echo [OK] Server started
echo        Vite dev server → http://localhost:5173
echo        Admin panel    → http://localhost:5173/admin
pause
