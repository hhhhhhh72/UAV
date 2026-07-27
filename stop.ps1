Write-Host "Stopping drone-platform API..." -ForegroundColor Yellow
Get-Process -Name "api" -ErrorAction SilentlyContinue | Stop-Process -Force
$p = Get-NetTCPConnection -LocalPort 8080 -ErrorAction SilentlyContinue
if ($p) {
    Write-Host "Port 8080 still busy, force killing..." -ForegroundColor Red
    $p | ForEach-Object { Stop-Process -Id $_.OwningProcess -Force -ErrorAction SilentlyContinue }
}
Start-Sleep 2
if (Get-NetTCPConnection -LocalPort 8080 -ErrorAction SilentlyContinue) {
    Write-Host "[WARN] Port 8080 still occupied" -ForegroundColor Red
} else {
    Write-Host "[OK] Server stopped" -ForegroundColor Green
}
Read-Host "Press Enter to close"
