# 每日日报生成器
#
# 汇报要求（发群前逐条核对）：
#   1) 发布时间：每日 20:00 以前（脚本超时会提示）；
#   2) 字数：全文不超过 300 字（超限会提示并给出精简方向）；
#   3) 表达：语音输入豆包 → 直接发豆包产出内容；
#      进度推进类写进度节点（已部署/已复验/待验收…），持续优化类多用量词、少形容词；
#   4) 发送样式：不发 Word/文档，群里直接发文字。
#
# 统计口径（2026-09-17 修正）：
#   ✗ 不用文件 mtime 数「今天改了哪些文件」——一次 git reset --hard 重新检出就会刷新
#     所有文件的时间戳，数字虚高好几倍（当天实测报出「后端 352 个文件」）。
#   ✓ 只用 git 的真实状态：当日提交（git log）+ 当前未提交改动（git status）。
#
# ⚠ 本脚本产出的是**机械草稿**：它只能数文件和抄提交标题，给不出「为什么改、影响是什么」。
#   发群那版仍建议人工过一遍（或让助手按当天实际工作写）。
#
# 用法: powershell -ExecutionPolicy Bypass -File scripts\daily-report.ps1
#       ... -Day "2026-09-17"   指定日期
#       ... -Detail              附逐条明细（会超 300 字，仅自用）
#       ... -NoClipboard         不写剪贴板
param(
    [string]$Day = "",
    [switch]$Detail,
    [switch]$NoClipboard
)

$ErrorActionPreference = 'Continue'
# 必须显式设成 UTF-8：PowerShell 5.1 默认用控制台代码页（简中系统是 936/GBK）去解码
# 原生命令的输出，而 git 吐的是 UTF-8 —— 结果脚本里写死的中文正常，从 git 抓来的
# 提交标题全是乱码（实测报出「鍥剧墖鎸夎嚜宸辨瘮渚嬫樉绀?」）。
# try/catch：计划任务用 -WindowStyle Hidden 运行，极端情况下可能没有控制台可设，
# 设置失败不该让整个脚本挂掉（那种情况下中文会乱，但报告仍会生成）。
try {
    [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
    $OutputEncoding            = [System.Text.Encoding]::UTF8
} catch { }
Set-Location (Split-Path -Parent $PSScriptRoot)   # → 仓库根

$date = if ($Day) { [datetime]::Parse($Day) } else { Get-Date }
$start = $date.Date
$end = $start.AddDays(1)
$week = @('日','一','二','三','四','五','六')[[int]$date.DayOfWeek]

function Get-Area([string]$path) {
    if ($path -match '^cmd/|^internal/')                  { return '后端' }
    if ($path -match '^migrations/')                      { return '迁移' }
    if ($path -match '^miniprogram/')                     { return '小程序' }
    if ($path -match '^frontend/')                        { return '管理后台' }
    if ($path -match '^deploy/|^scripts/')                { return '运维' }
    if ($path -match '^docs/|^CLAUDE\.md$|^README\.md$') { return '文档' }
    return $null
}

# ---------- 1. 当日提交 ----------
$since = $start.ToString('yyyy-MM-dd HH:mm:ss')
$until = $end.ToString('yyyy-MM-dd HH:mm:ss')
$subjects = @(git log "--since=$since" "--until=$until" --pretty=format:'%s' --no-merges 2>$null |
              Where-Object { $_ -ne '' })
$commitFiles = @(git log "--since=$since" "--until=$until" --name-only --pretty=format: --no-merges 2>$null |
                 Where-Object { $_ -ne '' } | Sort-Object -Unique)

# ---------- 2. 当前未提交的改动 ----------
$dirtyFiles = @(git status --porcelain 2>$null | ForEach-Object {
    if ($_.Length -gt 3) { $_.Substring(3).Trim('"') } })

# ---------- 3. 按区域归类 ----------
$areaCommit = [ordered]@{}
foreach ($f in $commitFiles) {
    $a = Get-Area $f
    if ($a) { if (-not $areaCommit.Contains($a)) { $areaCommit[$a] = 0 }; $areaCommit[$a]++ }
}
$areaDirty = [ordered]@{}
foreach ($f in $dirtyFiles) {
    $a = Get-Area $f
    if ($a) { if (-not $areaDirty.Contains($a)) { $areaDirty[$a] = 0 }; $areaDirty[$a]++ }
}

# 当日新增的迁移组（版本号前缀去重）
$migVersions = @($commitFiles | Where-Object { $_ -match '^migrations/\d+.*\.up\.sql$' } |
                 ForEach-Object { [regex]::Match($_, '^migrations/(\d+)').Groups[1].Value } |
                 Sort-Object -Unique)

function Format-Areas($table) {
    $parts = @()
    foreach ($k in $table.Keys) { if ($table[$k] -gt 0) { $parts += ($k + ' ' + $table[$k]) } }
    if ($parts.Count -eq 0) { return '无' }
    return ($parts -join '、')
}

# ---------- 4. 组装 ----------
$lines = @()
$lines += ('【日报 ' + $date.ToString('MM-dd') + ' 星期' + $week + '】')
$idx = 1

if ($subjects.Count -gt 0) {
    $lines += ($idx.ToString() + '. 提交 ' + $subjects.Count + ' 个、涉及 ' + $commitFiles.Count + ' 个文件（' + (Format-Areas $areaCommit) + '）。')
    $idx++
    $shown = 0
    foreach ($s in $subjects) {
        if ($shown -ge 3) { break }
        $t = $s -replace '^(fix|feat|style|chore|docs|refactor|perf|test|ops|ci)(\([\w\-/]+\))?[:：]\s*', ''
        if ($t.Length -gt 36) { $t = $t.Substring(0, 36) + '…' }
        $lines += ($idx.ToString() + '. ' + $t)
        $idx++; $shown++
    }
} else {
    $lines += ($idx.ToString() + '. 今日无提交。')
    $idx++
}
if ($migVersions.Count -gt 0) {
    # 组数多时只报区间，铺开列全会把 300 字预算吃光
    $migLabel = if ($migVersions.Count -le 3) { ($migVersions -join '、') } else { $migVersions[0] + '-' + $migVersions[-1] }
    $lines += ($idx.ToString() + '. 数据库迁移 ' + $migVersions.Count + ' 组（' + $migLabel + '）。')
    $idx++
}
if ($dirtyFiles.Count -gt 0) {
    $lines += ($idx.ToString() + '. 未提交改动 ' + $dirtyFiles.Count + ' 个文件（' + (Format-Areas $areaDirty) + '）。')
    $idx++
}
if ($Detail) {
    $lines += ''
    $lines += '—— 明细 ——'
    $lines += ('提交: ' + ($subjects -join ' | '))
    $lines += ('提交文件: ' + ($commitFiles -join ', '))
    $lines += ('未提交: ' + ($dirtyFiles -join ', '))
}

$text = $lines -join [char]10

# ---------- 5. 输出 ----------
Write-Output $text
Write-Output ''
$len = $text.Length
if ($len -gt 300) {
    Write-Output ('[提示] ' + $len + ' 字 > 300：删掉最次要的一条，或把两条合并。')
} else {
    Write-Output ('[提示] ' + $len + ' 字，可直接发群。')
}
$now = Get-Date
if ($now -ge $end) {
    Write-Output ('[提示] 补写 ' + $date.ToString('MM-dd') + ' 的日报。')
} elseif ($now.Hour -ge 20) {
    Write-Output ('[提示] 现在 ' + $now.ToString('HH:mm') + ' 已过 20:00，尽快发群。')
}
Write-Output '[提示] 群里直接发上面这段文字，不要发 Word/文档。'

$outDir = Join-Path (Get-Location).Path '日报'
$null = New-Item -ItemType Directory -Force -Path $outDir
$outFile = Join-Path $outDir ($date.ToString('yyyy-MM-dd') + '.txt')
Set-Content -Path $outFile -Value $text -Encoding UTF8
Write-Output ('[文件] ' + $outFile)
if (-not $NoClipboard) {
    try { Set-Clipboard -Value $text; Write-Output '[剪贴板] 已复制' } catch { Write-Output '[剪贴板] 复制失败（不影响文件）' }
}
