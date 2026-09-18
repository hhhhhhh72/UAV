# 每日工作日报：生成 + 投递
#
# 汇报要求（发群前逐条核对，2026-09-18 用户明确）：
#   1) 发布时间：每日 20:00 以前（计划任务 17:30 触发本脚本）
#   2) 字数：全文不超过 300 字（脚本会硬性裁剪到预算内，并在结尾报告实际字数）
#   3) 表达：进度推进类写**进度节点**（已推送 / 已部署 / 待部署）；
#            持续优化类**多用量词、少形容词**（N 个提交、N 个文件、N 组迁移、34%→24%）
#   4) 发送样式：不发 Word/文档，群里直接发**纯文字**
#
# 与旧版的区别（旧版被否掉，原话「日报内容错了」）：
#   旧版只数文件 + 抄前 3 条提交标题，读起来像构建日志，看不出「今天到底推进了什么」。
#   新版**按模块归类**：每个提交按 scope（escrow/pay/ops/…，不认就按改动文件所在区域）
#   归到 资金/支付/后端/管理后台/小程序/运维/文档，逐模块给出「做了哪几件事」——
#   提交标题本身就是真实工作内容，比文件计数有意义得多。
#
# 进度节点不是猜的：
#   已推送 / 有 N 个未推送 → git rev-list --count origin/master..HEAD
#   今日是否部署过生产      → ssh 查 uav-api-1 容器的 StartedAt
#   未提交改动              → git status --porcelain
#
# 发送路径：本机生成 → scp 到服务器 → 服务器用群机器人发出去（deploy/post-work-report.sh）。
#   为什么不本机直接 POST：本机 schannel 拿不到 TLS 凭据，curl 与 Invoke-WebRequest
#   一律报 (35) SEC_E_NO_CREDENTIALS，任何 https 都发不出去。详见那个脚本的注释。
#   发送成功后服务器才写心跳，运维告警据此判断「日报有没有真的发出去」。
#
# 用法:
#   powershell -ExecutionPolicy Bypass -File scriptsdaily-report.ps1           # 生成并发送（计划任务用的就是这条）
#   ... -NoSend        只生成不发送（本地预览，仍写文件与剪贴板）
#   ... -Day 2026-09-17    指定日期
#   ... -Detail            附逐条明细（会超 300 字，仅自用）
#   ... -NoClipboard       不写剪贴板
param(
    [string]$Day = "",
    [switch]$Detail,
    [switch]$NoClipboard,
    [switch]$NoSend
)

$ErrorActionPreference = 'Continue'
# 必须显式设成 UTF-8：PowerShell 5.1 默认用控制台代码页（简中系统是 936/GBK）去解码
# 原生命令的输出，而 git 吐的是 UTF-8 —— 结果脚本里写死的中文正常，从 git 抓来的
# 提交标题全是乱码（实测报出「鍥剧墖鎸夎嚜宸辨瘮渚嬫樉绀?」）。
# 计划任务用 -WindowStyle Hidden 运行，极端情况下可能没有控制台可设，
# 设置失败不该让整个脚本挂掉（那种情况下中文会乱，但报告仍会生成）。
try {
    [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
    $OutputEncoding            = [System.Text.Encoding]::UTF8
} catch { }
Set-Location (Split-Path -Parent $PSScriptRoot)   # → 仓库根

$BUDGET = 300                                      # 汇报要求：不超过 300 字
$date = if ($Day) { [datetime]::Parse($Day) } else { Get-Date }
$start = $date.Date
$end = $start.AddDays(1)
$week = @('日','一','二','三','四','五','六')[[int]$date.DayOfWeek]

# ---------- 1. 取当日提交（连同每个提交改了哪些文件） ----------
$since = $start.ToString('yyyy-MM-dd HH:mm:ss')
$until = $end.ToString('yyyy-MM-dd HH:mm:ss')
$raw = @(git log "--since=$since" "--until=$until" --pretty=format:'@@%h|%s' --name-only --no-merges 2>$null)

$commits = @()
foreach ($line in $raw) {
    if ($line -like '@@*') {
        $body = $line.Substring(2)
        $p = $body.Split('|', 2)
        $commits += [pscustomobject]@{ Sha = $p[0]; Subject = $p[1]; Files = @() }
    } elseif ($line.Trim() -ne '' -and $commits.Count -gt 0) {
        $commits[-1].Files += $line.Trim()
    }
}

# ---------- 2. 归类到模块 ----------
# 先看 scope（提交标题里 fix(escrow): 这种），scope 不认识再按改动文件所在区域。
$scopeMap = @{
    'escrow' = '资金'; 'money' = '资金'; 'deposit' = '资金'; 'withdraw' = '资金'
    'pay' = '支付'; 'payment' = '支付'; 'refund' = '支付'; 'wechatpay' = '支付'
    'ops' = '运维'; 'deploy' = '运维'; 'ci' = '运维'; 'backup' = '运维'; 'alert' = '运维'
    'admin' = '管理后台'; 'publish' = '管理后台'; 'web' = '管理后台'; 'ui' = '管理后台'
    'frontend' = '管理后台'; 'mp' = '小程序'; 'miniprogram' = '小程序'
    'docs' = '文档'
    'errors' = '后端'; 'api' = '后端'; 'auth' = '后端'; 'db' = '后端'; 'service' = '后端'
}
$order = @('资金','支付','后端','管理后台','小程序','运维','文档','其他')

function Get-AreaOfFile([string]$path) {
    if ($path -match '^cmd/|^internal/|^migrations/') { return '后端' }
    if ($path -match '^frontend/')                    { return '管理后台' }
    if ($path -match '^miniprogram/')                 { return '小程序' }
    if ($path -match '^deploy/|^scripts/|^\.github/')  { return '运维' }
    if ($path -match '^docs/|^CLAUDE\.md$|^README\.md$') { return '文档' }
    return '其他'
}

function Get-ModuleOf($commit) {
    $scope = ''
    if ($commit.Subject -match '^[a-zA-Z]+\(([^)]+)\)') { $scope = $Matches[1].ToLower() }
    foreach ($k in ($scope -split '[/,]')) {
        $k = $k.Trim()
        if ($k -ne '' -and $scopeMap.ContainsKey($k)) { return $scopeMap[$k] }
    }
    if ($commit.Files.Count -eq 0) { return '其他' }
    $tally = @{}
    foreach ($f in $commit.Files) {
        $a = Get-AreaOfFile $f
        if (-not $tally.ContainsKey($a)) { $tally[$a] = 0 }
        $tally[$a]++
    }
    return ($tally.GetEnumerator() | Sort-Object Value -Descending | Select-Object -First 1).Key
}

# 提交标题去掉 fix(xxx): 前缀
function Get-Title([string]$subject) {
    return ($subject -replace '^(fix|feat|style|chore|docs|refactor|perf|test|ops|ci)(\([^)]*\))?[:：]\s*', '').Trim()
}

# 一个提交里常常塞了好几件事（例如「堵住自助充值印钞口;支付端点改 4xx;补下单限频」），
# 所以先按 ; 拆成**工作项**再归类 —— 拆开后每条都短，300 字里能装下更多真实内容，
# 也不会出现「一句话被切到一半」的残句。
$itemsByModule = @{}
foreach ($c in $commits) {
    $m = Get-ModuleOf $c
    if (-not $itemsByModule.ContainsKey($m)) { $itemsByModule[$m] = @() }
    foreach ($frag in ((Get-Title $c.Subject) -split '[;；]')) {
        $f = $frag.Trim()
        if ($f -eq '') { continue }
        # 超长就截，但**不切在半个括号里**：退到左括号之前，看起来才是完整短语
        # （「微信退款链路打通(服务层/handler/回调/…」→「微信退款链路打通…」）
        if ($f.Length -gt 24) {
            $cut = $f.Substring(0, 24)
            $open = $cut.LastIndexOf('(')
            if ($open -ge 0 -and $cut.IndexOf(')', $open) -lt 0) { $cut = $cut.Substring(0, $open) }
            $f = $cut.TrimEnd('(', '/', '、', '，', ' ', '；', ';') + '…'
        }
        $itemsByModule[$m] += $f
    }
}

# ---------- 3. 规模与迁移 ----------
$allFiles = @($commits | ForEach-Object { $_.Files } | Where-Object { $_ -ne '' } | Sort-Object -Unique)
$migVersions = @($allFiles | Where-Object { $_ -match '^migrations/\d+.*\.up\.sql$' } |
                 ForEach-Object { [regex]::Match($_, '^migrations/(\d+)').Groups[1].Value } |
                 Sort-Object -Unique)
$dirty = @(git status --porcelain 2>$null | ForEach-Object { if ($_.Length -gt 3) { $_.Substring(3).Trim('"') } })

# ---------- 4. 进度节点 ----------
$unpushed = -1
try {
    $n = (git rev-list --count 'origin/master..HEAD' 2>$null | Select-Object -First 1)
    if ($n -match '^\d+$') { $unpushed = [int]$n }
} catch { }

$deployNote = ''
try {
    $started = (& ssh -o ConnectTimeout=8 -o BatchMode=yes root@api.cqnarc.cn "docker inspect -f '{{.State.StartedAt}}' uav-api-1" 2>$null | Select-Object -First 1)
    if ($started -and $started.Trim() -ne '') {
        $t = [datetime]::Parse($started.Trim(), [Globalization.CultureInfo]::InvariantCulture,
                                [Globalization.DateTimeStyles]::RoundtripKind).ToLocalTime()
        if ($t.Date -eq (Get-Date).Date) { $deployNote = '今日已部署生产 ' + $t.ToString('HH:mm') }
        else { $deployNote = '生产上次部署 ' + $t.ToString('MM-dd HH:mm') }
    }
} catch { }

$nodes = @()
if ($unpushed -eq 0)      { $nodes += '已推送' }
elseif ($unpushed -gt 0)  { $nodes += "有 $unpushed 个提交未推送" }
if ($deployNote -ne '')   { $nodes += $deployNote }
if ($dirty.Count -gt 0)   { $nodes += "$($dirty.Count) 个文件未提交" }
if ($nodes.Count -eq 0)   { $nodes += '状态未取到' }

# ---------- 5. 组装（超预算就从最后一个模块起往前裁） ----------
$head = '【日报 ' + $date.ToString('MM-dd') + ' 周' + $week + '】'
$counts = '1. 提交 ' + $commits.Count + ' 个、' + $allFiles.Count + ' 个文件'
if ($migVersions.Count -gt 0) { $counts += '，' + $migVersions.Count + ' 组数据库迁移' }
$counts += '。'

function Render($lines) { return ($lines -join [char]10) }

# 预算分配：先扣掉标题/规模行/进度行的固定开销，剩下的**均分**给各模块；
# 每个模块在自己的份额内尽量多装**完整**的工作项 —— 装不下就整条跳过，绝不切半句。
function Build-Modules($presentModules) {
    $out = @()
    $idx = 2
    foreach ($m in $presentModules) {
        $prefix = $idx.ToString() + '. ' + $m + '：'
        $share = [Math]::Floor(($BUDGET - $fixed) / [Math]::Max(1, $presentModules.Count)) - $prefix.Length
        if ($share -lt 8) { $share = 8 }
        $picked = @()
        $used = 0
        foreach ($item in ($itemsByModule[$m] | Select-Object -Unique)) {
            $cost = $item.Length + $(if ($picked.Count -gt 0) { 1 } else { 0 })
            if ($used + $cost -gt $share) { continue }   # 装不下就试下一条（更短的），不切半句
            $picked += $item
            $used += $cost
        }
        if ($picked.Count -gt 0) {
            $line = $prefix + ($picked -join '；')
            if ($line.EndsWith('…')) { $out += $line } else { $out += ($line + '。') }   # 「…。」读着别扭
            $idx++
        }
    }
    return $out
}

$present = @($order | Where-Object { $itemsByModule.ContainsKey($_) -and $itemsByModule[$_].Count -gt 0 })
# fixed 是标题+规模行+进度行的固定开销（含换行与编号余量）
$fixed = $head.Length + $counts.Length + $nodes.Length + 26
$mods = Build-Modules $present
$nodeLine = ($mods.Count + 2).ToString() + '. 进度：' + ($nodes -join '，') + '。'
$body = @($head, $counts) + $mods + @($nodeLine)

# 均分只是估算，取整后仍可能超；超了就砍掉最后一个模块重排 —— 编号会跟着变，
# 所以要整段重建而不是删一行。保住靠前的模块（资金/支付/后端…），先牺牲靠后的。
while ((Render $body).Length -gt $BUDGET -and $present.Count -gt 1) {
    $present = @($present[0..($present.Count - 2)])
    $mods = Build-Modules $present
    $nodeLine = ($mods.Count + 2).ToString() + '. 进度：' + ($nodes -join '，') + '。'
    $body = @($head, $counts) + $mods + @($nodeLine)
}

if ($Detail) {
    $body += ''
    $body += '—— 明细 ——'
    foreach ($c in $commits) { $body += ($c.Sha + ' ' + $c.Subject) }
    $body += ('文件: ' + ($allFiles -join ', '))
}

$text = Render $body

# ---------- 6. 输出 ----------
Write-Output $text
Write-Output ''
Write-Output ('[字数] ' + $text.Length + ' / ' + $BUDGET + $(if ($text.Length -le $BUDGET) { '（符合要求）' } else { '（超了，需手工精简）' }))
$now = Get-Date
if ($now.Hour -ge 20) { Write-Output '[提示] 现在已过 20:00，尽快发群。' }

$outDir = Join-Path (Get-Location).Path '日报'
$null = New-Item -ItemType Directory -Force -Path $outDir
$outFile = Join-Path $outDir ($date.ToString('yyyy-MM-dd') + '.txt')
Set-Content -Path $outFile -Value $text -Encoding UTF8
Write-Output ('[文件] ' + $outFile)

if (-not $NoClipboard) {
    try { Set-Clipboard -Value $text; Write-Output '[剪贴板] 已复制' } catch { Write-Output '[剪贴板] 复制失败（不影响文件）' }
}

# ---------- 7. 投递 ----------
if ($NoSend) {
    Write-Output '[发送] 已跳过（-NoSend）'
    exit 0
}

$tmp = Join-Path $env:TEMP 'work-report.txt'
# 不写 BOM（服务器侧也会剥，双保险）；scp 传文件而不是管道，绕开 PS 5.1 向原生程序
# 写 stdin 时的编码问题（$OutputEncoding 默认 ASCII，中文会变问号）
[IO.File]::WriteAllText($tmp, $text, (New-Object Text.UTF8Encoding($false)))
& scp -o ConnectTimeout=10 $tmp root@api.cqnarc.cn:/tmp/work-report.txt 2>&1 | Out-Null
if ($LASTEXITCODE -ne 0) {
    Write-Output '[发送] 失败：scp 传文件不成功（网络或免密登录的问题）'
    exit 1
}
$out = & ssh -o ConnectTimeout=10 root@api.cqnarc.cn 'bash /root/UAV/deploy/post-work-report.sh /tmp/work-report.txt' 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Output '[发送] 已推送（服务器已记录送达心跳）'
    exit 0
}
Write-Output ('[发送] 失败：' + ($out -join ' '))
exit 1