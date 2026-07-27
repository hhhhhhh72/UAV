@echo off
echo === Stopping drone-platform API ===
taskkill /f /im api.exe >nul 2>&1
netstat -ano | find ":8080" | find "LISTENING" >nul
if errorlevel 1 (echo [OK] Server stopped) else (echo [WARN] Port still busy)
pause
