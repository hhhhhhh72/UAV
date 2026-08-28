@echo off
setlocal DisableDelayedExpansion
REM 本地开发启动脚本（安全审计加固）：密钥优先取环境变量，未设置才用 dev 占位。
REM 生产使用 docker-compose（.env 注入），本脚本仅限本地开发，勿在生产复用。
if "%AUTH_SECRET%"=="" set AUTH_SECRET=drone-platform-dev-secret-32bytes!
if "%APP_ENV%"=="" set APP_ENV=development
if "%ADMIN_DEV_MODE%"=="" set ADMIN_DEV_MODE=true
if "%DATABASE_URL%"=="" set DATABASE_URL=postgresql://drone:drone_secret@localhost:5433/drone_platform?sslmode=disable
cd /d "%~dp0"
api.exe
pause
