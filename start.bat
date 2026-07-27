@echo off
echo === Starting drone-platform API ===
set AUTH_SECRET=drone-platform-dev-secret-32bytes!
set APP_ENV=development
set ADMIN_DEV_MODE=true
echo Building...
go build -o api.exe .\cmdpi
if errorlevel 1 (echo [FAIL] Build failed & pause & exit /b 1)
echo Starting on port 8080...
start "" api.exe
echo [OK] Server running
echo Admin panel: http://localhost:5173/admin
pause
