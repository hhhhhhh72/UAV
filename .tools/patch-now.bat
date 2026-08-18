@echo off
set LOG=D:\w-yao\.tools\dsh-app\patch-log.txt
echo %date% %time% [task2] start >> "%LOG%"
taskkill /IM DSH-Desktop.exe /F >> "%LOG%" 2>&1
timeout /t 10 /nobreak >> "%LOG%" 2>&1
cd /d D:\w-yao\.tools
"D:\Node\node.exe" patch-icons.js >> "%LOG%" 2>&1
echo %date% %time% [task2] relaunch >> "%LOG%"
start "" "C:\Users\21125\AppData\Local\Programs\dsh-desktop\DSH-Desktop.exe"
timeout /t 5 /nobreak >nul
ie4uinit.exe -show
echo %date% %time% [task2] done >> "%LOG%"
