# start.ps1 — 编译并启动 drone-platform API
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptDir

Write-Host "=== Start drone-platform API ===" -ForegroundColor Cyan

# 环境变量
$env:AUTH_SECRET     = "drone-platform-dev-secret-32bytes!"
$env:APP_ENV         = "development"
$env:ADMIN_DEV_MODE  = "true"

# 编译
Write-Host "Building..." -ForegroundColor Yellow
go build -o api.exe .\cmd\api
if ($LASTEXITCODE -ne 0) {
    Write-Host "[FAIL] Build failed — check errors above" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}
Write-Host "[OK] Build done" -ForegroundColor Green

# 启动
Write-Host "Starting on :8080 ..." -ForegroundColor Yellow
Start-Process -FilePath "$scriptDir\api.exe"

Start-Sleep 2

# 验证
$code = 0
try { $code = (Invoke-WebRequest -Uri "http://localhost:8080/api/v1/admin/shops" -UseBasicParsing -TimeoutSec 3).StatusCode } catch { $code = 0 }
if ($code -eq 401 -or $code -eq 200) {
    Write-Host "[OK] Server responding (HTTP $code)" -ForegroundColor Green
    Write-Host "     Admin panel → http://localhost:5173/admin" -ForegroundColor Cyan
} else {
    Write-Host "[FAIL] Server not responding (HTTP $code)" -ForegroundColor Red
}

Write-Host ""
Read-Host "Press Enter to close"
