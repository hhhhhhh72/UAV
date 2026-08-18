# ============================================================================
# C 盘清理脚本（安全版）
# 用法：
#   PowerShell（管理员）:  .\clean-c.ps1                    # 标准清理（默认安全项）
#                          .\clean-c.ps1 -DryRun             # 只统计预览，不删除
#                          .\clean-c.ps1 -IncludeRisky       # 加上 Windows.old / 休眠文件等大项
#   或直接双击 clean-c.bat（会自动申请管理员权限）
#
# 安全设计：
#   - 默认只清"系统/应用缓存与临时文件"，绝不碰 文档/桌面/下载/图片 等用户数据
#   - 每个环节独立 try/catch，占用中的文件自动跳过
#   - 删除前统计大小，结束后报告释放了多少空间
# ============================================================================
[CmdletBinding()]
param(
    [switch]$DryRun,
    [switch]$IncludeRisky,
    [switch]$SkipBrowserCache
)

$ErrorActionPreference = 'SilentlyContinue'
$Host.UI.RawUI.WindowTitle = 'C盘清理'
$log = "$env:LOCALAPPDATA\c-clean.log"
$freedMB = 0
$script:started = Get-Date

# ---------- 管理员权限自提升 ----------
$isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Host '需要管理员权限，正在重新以管理员身份启动...' -ForegroundColor Yellow
    Start-Process powershell -Verb RunAs -ArgumentList "-NoProfile -ExecutionPolicy Bypass -File `"$PSCommandPath`" $(if($DryRun){'-DryRun'}) $(if($IncludeRisky){'-IncludeRisky'})"
    exit
}

function Get-DirSizeMB($path) {
    if (-not (Test-Path $path)) { return 0 }
    $sum = 0
    Get-ChildItem $path -Recurse -Force -ErrorAction SilentlyContinue |
        ForEach-Object { try { $sum += $_.Length } catch {} }
    return [math]::Round($sum / 1MB, 1)
}

function Clean-Dir($path, $name) {
    if (-not (Test-Path $path)) { Write-Host "  跳过 $name（不存在）" -ForegroundColor DarkGray; return }
    $size = Get-DirSizeMB $path
    if ($DryRun) {
        Write-Host ("  [预览] {0}: {1} MB（仅显示，不删除）" -f $name, $size) -ForegroundColor Cyan
        return
    }
    try {
        Remove-Item $path\* -Recurse -Force -ErrorAction SilentlyContinue
        $global:freedMB += $size
        Write-Host ("  ✅ {0}: 释放约 {1} MB" -f $name, $size) -ForegroundColor Green
    } catch {
        Write-Host "  ⚠️ $name 部分文件被占用，已跳过" -ForegroundColor Yellow
    }
}

Write-Host "`n========== C 盘清理 ==========" -ForegroundColor Cyan
Write-Host "模式: $(if($DryRun){'预览(不删除)'}else{'清理'}) | 风险项: $(if($IncludeRisky){'包含'}else{'不包含'})"
Write-Host "日志: $log`n"

# ---------- 1. 系统临时目录 ----------
Write-Host '[1/9] 临时文件' -ForegroundColor White
Clean-Dir $env:TEMP '用户临时目录'
Clean-Dir 'C:\Windows\Temp' '系统临时目录'
Clean-Dir 'C:\Windows\Prefetch' '预读取缓存'
Clean-Dir "$env:LOCALAPPDATA\Microsoft\Windows\INetCache" 'IE/系统网络缓存'

# ---------- 2. 回收站 ----------
Write-Host '[2/9] 回收站' -ForegroundColor White
if (-not $DryRun) {
    try {
        $rb = New-Object -ComObject Shell.Application
        $rb.Namespace(10).Items() | Out-Null
        Clear-RecycleBin -Force -ErrorAction SilentlyContinue
        Write-Host '  ✅ 回收站已清空' -ForegroundColor Green
    } catch { Write-Host '  ⚠️ 回收站清空失败' -ForegroundColor Yellow }
} else {
    Write-Host '  [预览] 回收站将被清空' -ForegroundColor Cyan
}

# ---------- 3. Windows 更新缓存 ----------
Write-Host '[3/9] Windows 更新缓存' -ForegroundColor White
$wuSize = Get-DirSizeMB 'C:\Windows\SoftwareDistribution\Download'
if ($DryRun) { Write-Host "  [预览] Windows更新下载缓存: $wuSize MB" -ForegroundColor Cyan }
elseif (-not (Test-Path 'C:\Windows\SoftwareDistribution\Download')) { Write-Host '  跳过（不存在）' -ForegroundColor DarkGray }
else {
    try {
        Stop-Service wuauserv, bits -Force
        Remove-Item 'C:\Windows\SoftwareDistribution\Download\*' -Recurse -Force
        Start-Service wuauserv, bits
        $global:freedMB += $wuSize
        Write-Host "  ✅ Windows更新缓存: 释放约 $wuSize MB" -ForegroundColor Green
    } catch { Write-Host '  ⚠️ 更新服务处理失败，已跳过' -ForegroundColor Yellow }
}

# ---------- 4. 错误报告 / 崩溃转储 ----------
Write-Host '[4/9] 错误报告与转储' -ForegroundColor White
Clean-Dir 'C:\ProgramData\Microsoft\Windows\WER\ReportQueue' '错误报告队列'
Clean-Dir 'C:\ProgramData\Microsoft\Windows\WER\ReportArchive' '错误报告归档'
Clean-Dir "$env:LOCALAPPDATA\CrashDumps" '应用崩溃转储'
Clean-Dir 'C:\Windows\Minidump' '系统蓝屏转储'

# ---------- 5. 传递优化缓存 ----------
Write-Host '[5/9] 传递优化缓存' -ForegroundColor White
if (-not $DryRun) {
    try { Delete-DeliveryOptimizationCache -Force -ErrorAction SilentlyContinue | Out-Null
          Write-Host '  ✅ 传递优化缓存已清除' -ForegroundColor Green } catch { Write-Host '  ⚠️ 清除失败' -ForegroundColor Yellow }
} else { Write-Host '  [预览] 传递优化缓存将被清除' -ForegroundColor Cyan }

# ---------- 6. 浏览器缓存 ----------
Write-Host '[6/9] 浏览器缓存' -ForegroundColor White
if ($SkipBrowserCache) { Write-Host '  已跳过（-SkipBrowserCache）' -ForegroundColor DarkGray }
else {
    Clean-Dir "$env:LOCALAPPDATA\Google\Chrome\User Data\Default\Cache" 'Chrome 缓存'
    Clean-Dir "$env:LOCALAPPDATA\Google\Chrome\User Data\Default\Code Cache" 'Chrome 代码缓存'
    Clean-Dir "$env:LOCALAPPDATA\Microsoft\Edge\User Data\Default\Cache" 'Edge 缓存'
    Clean-Dir "$env:LOCALAPPDATA\Microsoft\Edge\User Data\Default\Code Cache" 'Edge 代码缓存'
}
# Firefox：只清各 profile 的缓存子目录，绝不碰书签/密码等配置
if (-not $SkipBrowserCache) {
    if (Test-Path "$env:APPDATA\Mozilla\Firefox\Profiles") {
        if ($DryRun) { Write-Host '  [预览] Firefox 各配置的 cache2/startupCache 将被清除' -ForegroundColor Cyan }
        else {
            Get-ChildItem "$env:APPDATA\Mozilla\Firefox\Profiles\*.default*" -Directory -ErrorAction SilentlyContinue | ForEach-Object {
                Remove-Item "$($_.FullName)\cache2\*" -Recurse -Force -ErrorAction SilentlyContinue
                Remove-Item "$($_.FullName)\startupCache\*" -Recurse -Force -ErrorAction SilentlyContinue
            }
            Write-Host '  ✅ Firefox 缓存已清除' -ForegroundColor Green
        }
    } else { Write-Host '  跳过 Firefox（未安装）' -ForegroundColor DarkGray }
}

# ---------- 7. 开发工具缓存 ----------
Write-Host '[7/9] 开发工具缓存' -ForegroundColor White
Clean-Dir "$env:LOCALAPPDATA\npm-cache" 'npm 缓存'
Clean-Dir "$env:LOCALAPPDATA\pip\cache" 'pip 缓存'
Clean-Dir "$env:LOCALAPPDATA\pnpm-cache" 'pnpm 缓存'
Clean-Dir "$env:LOCALAPPDATA\Yarn\Cache" 'yarn 缓存'
Clean-Dir "$env:LOCALAPPDATA\go-build" 'Go 构建缓存'
Clean-Dir "$env:LOCALAPPDATA\NuGet\v3-cache" 'NuGet 缓存'
# JetBrains：只清各产品的 caches 子目录，保留配置/插件
if (Test-Path "$env:LOCALAPPDATA\JetBrains") {
    if ($DryRun) { Write-Host '  [预览] JetBrains 各产品 caches 子目录将被清除' -ForegroundColor Cyan }
    else {
        Get-ChildItem "$env:LOCALAPPDATA\JetBrains\*\caches" -Directory -ErrorAction SilentlyContinue | ForEach-Object {
            Remove-Item "$($_.FullName)\*" -Recurse -Force -ErrorAction SilentlyContinue
        }
        Write-Host '  ✅ JetBrains caches 已清除' -ForegroundColor Green
    }
} else { Write-Host '  跳过 JetBrains（未安装）' -ForegroundColor DarkGray }

# ---------- 8. 大文件报告（只统计用户目录，不删除；全盘扫描太慢故不做） ----------
Write-Host '[8/9] 大文件扫描（仅统计用户目录，不删除）' -ForegroundColor White
$big = Get-ChildItem $env:USERPROFILE -Recurse -Force -File -ErrorAction SilentlyContinue |
    Where-Object Length -gt 500MB |
    Sort-Object Length -Descending | Select-Object -First 10 FullName, @{N='GB';E={[math]::Round($_.Length/1GB,2)}}
if ($big) { $big | Format-Table -AutoSize } else { Write-Host '  用户目录下未发现 >500MB 的文件' -ForegroundColor DarkGray }

# ---------- 9. 风险大项（需 -IncludeRisky） ----------
Write-Host '[9/9] 可选大项' -ForegroundColor White
if (-not $IncludeRisky) {
    Write-Host '  未启用。以下项目需要 -IncludeRisky：' -ForegroundColor DarkGray
    Write-Host '    - Windows.old 旧系统备份（若存在，通常数 GB）' -ForegroundColor DarkGray
    Write-Host '    - hiberfil.sys 休眠文件（关闭休眠可释放约内存大小的空间）' -ForegroundColor DarkGray
    Write-Host '    - DISM 组件库清理（释放 WinSxS，耗时较长）' -ForegroundColor DarkGray
} else {
    if (Test-Path 'C:\Windows.old') {
        $s = Get-DirSizeMB 'C:\Windows.old'
        if ($DryRun) { Write-Host "  [预览] Windows.old: $s MB" -ForegroundColor Cyan }
        else {
            Remove-Item 'C:\Windows.old' -Recurse -Force -ErrorAction SilentlyContinue
            $global:freedMB += $s
            Write-Host "  ✅ Windows.old: 释放约 $s MB" -ForegroundColor Green
        }
    } else { Write-Host '  Windows.old 不存在' -ForegroundColor DarkGray }

    if (Test-Path 'C:\hiberfil.sys') {
        $s = [math]::Round((Get-Item 'C:\hiberfil.sys').Length / 1MB, 1)
        if ($DryRun) { Write-Host "  [预览] 休眠文件: $s MB（将执行 powercfg /h off）" -ForegroundColor Cyan }
        else { powercfg /h off; $global:freedMB += $s; Write-Host "  ✅ 休眠已关闭: 释放约 $s MB" -ForegroundColor Green }
    }

    if (-not $DryRun) {
        Write-Host '  正在运行 DISM 组件清理（可能需要几分钟）...' -ForegroundColor Yellow
        Dism.exe /Online /Cleanup-Image /StartComponentCleanup /Quiet | Out-Null
    } else { Write-Host '  [预览] DISM 组件库将被清理' -ForegroundColor Cyan }
}

# ---------- 汇总 ----------
$drive = Get-PSDrive C
$totalGB = [math]::Round(($drive.Used) / 1GB, 1)
$freeGB = [math]::Round(($drive.Free) / 1GB, 1)
Write-Host "`n========== 完成 ==========" -ForegroundColor Cyan
if ($DryRun) { Write-Host '（预览模式：以上均未删除）' -ForegroundColor Yellow }
else {
    Write-Host "本次释放约: $freedMB MB" -ForegroundColor Green
    "$(Get-Date -Format 'yyyy-MM-dd HH:mm') 释放约 ${freedMB}MB" | Out-File $log -Append -Encoding utf8
}
Write-Host "C盘当前: 已用 $totalGB GB / 可用 $freeGB GB"
Write-Host "`n按任意键退出..." -ForegroundColor DarkGray
$null = $Host.UI.RawUI.ReadKey('NoEcho,IncludeKeyDown')
