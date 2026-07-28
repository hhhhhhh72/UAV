@echo off
echo === Stopping all services ===
taskkill /f /im node.exe >nul 2>&1
taskkill /f /im api.exe >nul 2>&1
taskkill /f /im cpolar.exe >nul 2>&1
docker compose down db >nul 2>&1
echo [OK] All stopped
pause
