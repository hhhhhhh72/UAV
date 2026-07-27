# stop.ps1 — 停止 drone-platform API
Write-Host "=== Stop drone-platform API ===" -ForegroundColor Yellow

$killed = Get-Process -Name "api" -ErrorAction SilentlyContinue | Stop-Process -Force -PassThru
if ($killed) {
    Write-Host "[OK] Killed $($killed.Count) api.exe process(es)" -ForegroundColor Green
} else {
    Write-Host "[OK] No api.exe running" -ForegroundColor Green
}

Start-Sleep 1

$port = Get-NetTCPConnection -LocalPort 8080 -ErrorAction SilentlyContinue
if ($port) {
    Write-Host "[WARN] Port 8080 still busy (PID: $($port.OwningProcess))" -ForegroundColor Red
} else {
    Write-Host "[OK] Port 8080 freed" -ForegroundColor Green
}

Write-Host ""
Read-Host "Press Enter to close"
