@echo off
echo =========================================
echo   低空综合服务平台 - 快速配置
echo =========================================
echo.

:: 检查 .env 文件
if not exist .env (
    echo 创建 .env 配置文件...
    copy .env.example .env
    echo [OK] 已创建 .env 文件
    echo.
    echo [!] 请编辑 .env 文件设置以下关键配置:
    echo     - JWT_SECRET (必须)
    echo     - DSL_ADMIN_PASSWORD (必须)
    echo     - STUDY_ADMIN_PASSWORD (必须)
    echo.
    pause
) else (
    echo [OK] .env 文件已存在
    echo.
)

:: 安装依赖
echo 安装依赖...
call npm install
echo [OK] 依赖安装完成
echo.

:: 创建日志目录
if not exist logs (
    mkdir logs
    echo [OK] 创建日志目录
    echo.
)

:: 创建上传目录
if not exist uploads (
    mkdir uploads
    echo [OK] 创建上传目录
    echo.
)

echo =========================================
echo   配置完成!
echo =========================================
echo.
echo 启动服务:
echo   开发模式: npm run dev
echo   生产模式: npm start
echo.
echo 查看日志:
echo   type logs\%date:~0,4%-%date:~5,2%-%date:~8,2%.log
echo.
echo =========================================
pause
