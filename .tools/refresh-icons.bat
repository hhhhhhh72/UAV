@echo off
set LOG=D:\w-yao\.tools\dsh-app\patch-log.txt
echo %date% %time% [iconcache] start >> "%LOG%"
taskkill /f /im explorer.exe >> "%LOG%" 2>&1
timeout /t 2 /nobreak >nul
del /f /q "%LOCALAPPDATA%\IconCache.db" >> "%LOG%" 2>&1
del /f /q "%LOCALAPPDATA%\Microsoft\Windows\Explorer\iconcache_*.db" >> "%LOG%" 2>&1
del /f /q "%LOCALAPPDATA%\Microsoft\Windows\Explorer\thumbcache_*.db" >> "%LOG%" 2>&1
start explorer.exe
timeout /t 3 /nobreak >nul
ie4uinit.exe -show
echo %date% %time% [iconcache] done >> "%LOG%"
