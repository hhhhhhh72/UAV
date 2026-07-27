@echo off
chcp 65001 >nul
echo 停止所有服务...
taskkill /f /IM drone-api.exe 2>nul && echo  ✓ 后端已停止
taskkill /f /IM node.exe /FI "WINDOWTITLE eq Drone-Frontend*" 2>nul
:: 杀掉占用 5173 端口的 node 进程
for /f "tokens=5" %%a in ('netstat -ano 2^>nul ^| findstr ":5173" ^| findstr "LISTENING"') do (
    taskkill /f /PID %%a 2>nul
)
echo  ✓ 前端已停止
echo 全部服务已停止。
pause
