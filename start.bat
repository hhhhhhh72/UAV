@echo off
cd /d "%~dp0"

echo Stopping old API...
taskkill /f /im api.exe >nul 2>&1
timeout /t 2 /nobreak >nul

echo Starting services...
docker compose up -d db >nul 2>&1
timeout /t 3 /nobreak >nul

echo Starting API (8080)...
start "" /min run_api.bat
timeout /t 3 /nobreak >nul

echo Starting Admin (5173)...
cd frontend
if not exist node_modules\ call npm install
start "" /min cmd /c "npm run dev"
cd ..

echo.
echo Done - http://localhost:8080 / http://localhost:5173/admin
echo.
pause
