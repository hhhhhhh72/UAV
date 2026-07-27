@echo off
echo === Stopping drone-platform API ===
taskkill /f /im api.exe >nul 2>&1
if errorlevel 1 (echo [OK] No api.exe running) else (echo [OK] Killed api.exe)
pause
