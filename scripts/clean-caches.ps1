<#
  清理开发机上堆积的缓存与产物（2026-09-24 建）。

  为什么要有它：D:\w-yao 一度涨到 13.29 GB，其中 10.1 GB 是 **Go 编译缓存** ——
  而且是**四份**：.tools\gocache2（09-04）、.cache\go-build（07-12）、
  .impeccable\gocache（08-19）、.gocache（在用）。不同工具/时期各挑了一个路径，
  谁也不认得谁，就一直堆着。剩下 3 GB 是往期发布包（.tools 下 15 份 admin-dist-*.tgz
  共 505MB、根目录十几个 uav-deploy-*.tar.gz）、源码快照、改历史残留、本地编译的 .exe。
  这些东西**没有一个属于项目本身**：仓库 + 源码 + 依赖合计只有 1.1 GB 左右。

  用法：
    powershell -File scripts\clean-caches.ps1                    # 干跑：只列，不删
    powershell -File scripts\clean-caches.ps1 -Apply             # 真删（保守：保留在用缓存）
    powershell -File scripts\clean-caches.ps1 -Apply -IncludeGoCaches
                                                                  # 连在用 Go 缓存一起清
                                                                  # （下次编译慢几分钟，会自己长回来）
  注意：本文件带 UTF-8 BOM —— PS 5.1 读无 BOM 的 .ps1 会按 GBK 解码，中文注释会把语法搞坏。
        编辑后务必补回 BOM，否则满屏「意外的标记」。
#>
param(
  [switch]$Apply,
  [switch]$IncludeGoCaches
)

$ErrorActionPreference = 'SilentlyContinue'
$root = Split-Path -Parent $PSScriptRoot

function Get-SizeMB($path) {
  if (-not (Test-Path -LiteralPath $path)) { return 0 }
  $item = Get-Item -LiteralPath $path -Force
  if ($item.PSIsContainer) {
    $sum = (Get-ChildItem -LiteralPath $path -Recurse -File -Force | Measure-Object Length -Sum).Sum
    if ($null -eq $sum) { return 0 }
    return [math]::Round($sum / 1MB, 1)
  }
  return [math]::Round($item.Length / 1MB, 1)
}

# ---- 清单：先收集"要删什么"，再决定动不动手 ----
$targets = New-Object System.Collections.ArrayList
function Add-Target($path, $why, $optional) {
  if (-not (Test-Path -LiteralPath $path)) { return }
  [void]$targets.Add([pscustomobject]@{
    Path = $path; Why = $why; Optional = $optional; MB = (Get-SizeMB $path)
  })
}

# 别的工具/时期留下的 Go 编译缓存（一律可删）
Add-Target (Join-Path $root '.tools\gocache2')     '旧 Go 编译缓存' $false
Add-Target (Join-Path $root '.cache\go-build')     '旧 Go 编译缓存' $false
Add-Target (Join-Path $root '.impeccable\gocache') '旧 Go 编译缓存' $false

# 往期发布包 / 源码快照 / 编译产物
foreach ($pat in @('*.tgz', '*.tar.gz', '*.tar')) {
  foreach ($f in Get-ChildItem -Path (Join-Path $root '.tools') -File -Force -Filter $pat) {
    Add-Target $f.FullName '往期发布包' $false
  }
}
foreach ($f in Get-ChildItem -Path $root -File -Force -Filter '*.tar.gz') { Add-Target $f.FullName '往期部署包' $false }
foreach ($f in Get-ChildItem -Path $root -File -Force -Filter '*.exe')   { Add-Target $f.FullName '本地编译产物' $false }
foreach ($f in Get-ChildItem -Path (Join-Path $root '.migrate') -File -Force -Filter 'source_*.tar.gz') {
  Add-Target $f.FullName '源码快照（git 里有全量历史）' $false
}

# 杂项缓存
Add-Target (Join-Path $root '.git-rewrite')       '改历史残留' $false
Add-Target (Join-Path $root '.npm-cache-audit')   'npm 审计缓存' $false
Add-Target (Join-Path $root '.cache\npm-cache')  'npm 缓存' $false

# 在用的 Go 编译缓存：默认**不动**（删了下次编译要重来），要清得显式加开关
Add-Target (Join-Path $root '.gocache') '在用 Go 编译缓存（下次编译会重建）' $true

$doomed = $targets | Where-Object { -not $_.Optional -or $IncludeGoCaches }

Write-Host ""
Write-Host ("D:\w-yao 缓存/产物清单（{0}）" -f $(if ($Apply) { '执行删除' } else { '干跑，加 -Apply 才真删' }))
Write-Host ("{0,-52} {1,10}  {2}" -f '路径', '大小(MB)', '说明')
Write-Host ('-' * 100)
foreach ($t in ($doomed | Sort-Object MB -Descending)) {
  Write-Host ("{0,-52} {1,10:N1}  {2}" -f $t.Path.Replace($root + '\', ''), $t.MB, $t.Why)
}
$totalMB = ($doomed | Measure-Object MB -Sum).Sum
if ($null -eq $totalMB) { $totalMB = 0 }
Write-Host ('-' * 100)
Write-Host ("可释放合计：{0:N2} GB" -f ($totalMB / 1024))

if (-not $IncludeGoCaches) {
  $goMB = (Get-SizeMB (Join-Path $root '.gocache'))
  Write-Host ("（另有在用 Go 缓存 {0:N0} MB 未计入 —— 加 -IncludeGoCaches 可一并清）" -f $goMB)
}

if (-not $Apply) {
  Write-Host ""
  Write-Host "干跑结束，什么都没删。确认无误后加 -Apply。"
  return
}

Write-Host ""
foreach ($t in $doomed) {
  Remove-Item -LiteralPath $t.Path -Recurse -Force
  if (Test-Path -LiteralPath $t.Path) { Write-Host ("  失败  {0}" -f $t.Path) }
  else { Write-Host ("  已删  {0}" -f $t.Path.Replace($root + '\', '')) }
}
$left = (Get-ChildItem -Path $root -Recurse -File -Force | Measure-Object Length -Sum).Sum
Write-Host ""
Write-Host ("完成。D:\w-yao 现在 {0:N2} GB。" -f ($left / 1GB))
