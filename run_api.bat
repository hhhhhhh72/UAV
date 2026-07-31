@echo off
setlocal DisableDelayedExpansion
set AUTH_SECRET=drone-platform-dev-secret-32bytes!
set APP_ENV=development
set ADMIN_DEV_MODE=true
set DATABASE_URL=postgresql://drone:drone_secret@localhost:5433/drone_platform?sslmode=disable
cd /d "%~dp0"
api.exe
pause
