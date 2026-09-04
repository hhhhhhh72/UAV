# 每日日报生成器 —— 按日报规则自动汇总当日 git 提交
# 用法: powershell -ExecutionPolicy Bypass -File scripts\daily-report.ps1
#       powershell -ExecutionPolicy Bypass -File scripts\daily-report.ps1 -Day "2026-09-03"  # 指定日期
#       powershell ... -File scripts\daily-report.ps1 -Detail   # 输出逐条明细(超300字,仅自用)
param(
    [string]$Day = "",
    [switch]$Detail
)

Set-Location (Split-Path -Parent $PSScriptRoot)

$date = if ($Day) { [datetime]::Parse($Day) } else { Get-Date }
if ($Day) {
    $since = $date.ToString('yyyy-MM-dd') + ' 00:00:00'
    $until = $date.AddDays(1).ToString('yyyy-MM-dd') + ' 00:00:00'
} else {
    $since = 'midnight'
    $until = ''
}

$args1 = if ($until) { @("--since=$since", "--until=$until") } else { @("--since=$since") }
$log  = git log @args1 --pretty=format:%h%x7C%ad%x7C%s --date=format:%H:%M --no-merges
if (-not $log) {
    Write-Host ('# ' + $date.ToString('yyyy-MM-dd dddd') + ' 暂无提交。')
    exit 0
}

# 统计改动
$stats = git log @args1 --pretty=format: --numstat
$files = 0; $add = 0; $del = 0
foreach ($l in $stats) {
    if ($l -match '^(\d+)\t(\d+)\t') {
        $files++
        $add += [int]$Matches[1]
        $del += [int]$Matches[2]
    }
}

$rows = @()
foreach ($line in $log) {
    $p = $line -split '\|', 3
    if ($p.Count -eq 3) { $rows += , [pscustomobject]@{ h = $p[0]; t = $p[1]; s = $p[2] } }
}

# 主题聚类（小写匹配）：保持"多量词少形容词"
$groups = [ordered]@{}
foreach ($r in $rows) {
    $k = '其他'
    $s = $r.s.ToLower()
    if ($s -match 'testsite|场地') { $k = '场地预约' }
    elseif ($s -match 'competition|赛事') { $k = '赛事模块' }
    elseif ($s -match 'login|登录|守卫|gate') { $k = '登录/守卫' }
    elseif ($s -match 'apply|服务申请') { $k = '服务申请清理' }
    elseif ($s -match 'selectorquery|in\(' -or $s -match '吸顶测量') { $k = '吸顶测量' }
    if (-not $groups.Contains($k)) { $groups[$k] = @() }
    $groups[$k] += $r
}

$week = @('日', '一', '二', '三', '四', '五', '六')[[int]$date.DayOfWeek]
$lines = @()
$lines += ('【日报 ' + $date.ToString('MM-dd') + ' 星期' + $week + '】')
$lines += ('1. 今日提交 ' + $rows.Count + ' 个：' + $files + ' 个文件，新增 ' + $add + ' 行、删除 ' + $del + ' 行。')
$n = 2
foreach ($kv in $groups.GetEnumerator()) {
    $cnt = $kv.Value.Count
    $first = ($kv.Value | Select-Object -First 1).s -replace '^(fix|feat|style|chore|docs|refactor|perf)(\([\w-/]+\))?[:：]\s*', ''
    if ($first.Length -gt 18) { $first = $first.Substring(0, 18) + '…' }
    $lines += ($n.ToString() + '. ' + $kv.Key + '：' + $cnt + ' 项，' + $first)
    $n++
}
if ($Detail) {
    $lines += ''
    $lines += '—— 明细 ——'
    foreach ($r in $rows) { $lines += ($r.h + ' ' + $r.t + '  ' + $r.s) }
}

$text = $lines -join [char]10
$text
Write-Host ''
if ($text.Length -gt 300) { Write-Host ('[提示] ' + $text.Length + ' 字>300，请精简(或去掉明细)。') } else { Write-Host ('[提示] ' + $text.Length + ' 字，可直接发群。') }
