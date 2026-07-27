$env:AUTH_SECRET = "drone-platform-dev-secret-32bytes!"
$env:APP_ENV = "development"
$env:ADMIN_DEV_MODE = "true"

Write-Host "Building..." -ForegroundColor Cyan
go build -o api.exe .\cmd\api
if ($LASTEXITCODE -ne 0) {
    Write-Host "[FAIL] Build failed" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}

Write-Host "[OK] Build done, starting server..." -ForegroundColor Green
Start-Process -FilePath ".\api.exe" -WindowStyle Hidden

Write-Host "Server running on :8080" -ForegroundColor Yellow
Write-Host "Admin panel → http://localhost:5173/admin"
Read-Host "Press Enter to stop server"
taskkill /F /IM api.exe 2>$null
