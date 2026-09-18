# 工作日报（开发机侧）：**只做预览与手动补发**，生成和发送都在服务器上。
#
# 为什么改成本机只是个薄壳（2026-09-18，用户问「我云服务不关机每天都有？」）：
#   第一版在开发机上生成日报（内容来自本机 git），于是「开发机不开机那天就没有日报」——
#   而服务器 24 小时开着却帮不上忙。实测服务器能直连 GitHub（git ls-remote 成功），
#   就把生成也搬到了服务器：每天 19:30 由 cron 跑 deploy/work-report.sh，
#   fetch 一份 1.4MB 的裸库 → 生成正文 → 推送企业微信。**与本机开不开机无关**。
#
# 所以本脚本只剩两个用途：
#   1) 想在本地留一份 / 复制到剪贴板（默认走服务器 --dry-run，不推送）
#   2) 需要立刻补发一次（-Send）
# 日报正文的**唯一生成器**是服务器上的 deploy/work_report.py + work-report.sh，
# 本脚本不重复实现任何生成逻辑 —— 本项目已经因为「两份实现各自漂移」吃过亏。
#
# 用法:
#   powershell -ExecutionPolicy Bypass -File scripts\daily-report.ps1         # 预览（拉服务器生成的正文，写文件+剪贴板）
#   ... -Send              立刻推送到企业微信（走服务器同一条通道）
#   ... -Day 2026-09-17    看某一天的
#   ... -NoClipboard       不写剪贴板
param(
    [switch]$Send,
    [string]$Day = "",
    [switch]$NoClipboard
)

$ErrorActionPreference = 'Continue'
# 必须显式设成 UTF-8：PowerShell 5.1 默认用控制台代码页（简中 936/GBK）解码原生命令的输出，
# 而 ssh 传回来的是 UTF-8 —— 不设的话中文全是乱码。
try {
    [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
    $OutputEncoding            = [System.Text.Encoding]::UTF8
} catch { }
Set-Location (Split-Path -Parent $PSScriptRoot)

$date = if ($Day) { [datetime]::Parse($Day) } else { Get-Date }
$remote = if ($Send) { 'bash /root/UAV/deploy/work-report.sh' }
          else       { 'bash /root/UAV/deploy/work-report.sh --dry-run' }
if ($Day) { $remote = "DAY=$Day $remote" }

# 服务器脚本在正常/干跑两种模式下都把正文打到 stdout（日志走文件），所以这里拿到的就是正文。
$text = (& ssh -o ConnectTimeout=15 root@api.cqnarc.cn $remote 2>&1) -join [char]10
if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($text)) {
    Write-Output "[失败] 取不到日报：ssh exit=$LASTEXITCODE（服务器不可达？裸库 fetch 失败？）"
    if ($text) { Write-Output $text }
    exit 1
}
$text = $text.TrimEnd()

Write-Output $text
Write-Output ''
Write-Output ('[字数] ' + $text.Length + ' / 300' + $(if ($text.Length -le 300) { '（符合要求）' } else { '（超了）' }))
Write-Output $(if ($Send) { '[发送] 已推送（服务器已记录送达心跳）' } else { '[发送] 未推送（预览模式；要发就加 -Send）' })

$outDir = Join-Path (Get-Location).Path '日报'
$null = New-Item -ItemType Directory -Force -Path $outDir
$outFile = Join-Path $outDir ($date.ToString('yyyy-MM-dd') + '.txt')
Set-Content -Path $outFile -Value $text -Encoding UTF8
Write-Output ('[文件] ' + $outFile)

$runLog = Join-Path $outDir 'run.log'
Add-Content -Path $runLog -Encoding UTF8 -Value ((Get-Date).ToString('yyyy-MM-dd HH:mm:ss') + $(if ($Send) { '  已推送' } else { '  预览' }) + " （$($text.Length) 字）")

if (-not $NoClipboard) {
    try { Set-Clipboard -Value $text; Write-Output '[剪贴板] 已复制' } catch { Write-Output '[剪贴板] 复制失败（不影响文件）' }
}