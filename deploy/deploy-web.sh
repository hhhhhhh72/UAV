#!/usr/bin/env bash
# 管理后台前端发布：把 Vite 构建产物同步到 nginx 站点根目录。
#
# 为什么不是 `rm -rf "$WEB_ROOT"/*`：
# 站点根下除了构建产物（index.html、assets/）还放着历史媒体
# （static/home/*.jpg 被商品封面引用）。整目录清空会连带删掉它们 →
# 线上商品图直接 404。这里改成「覆盖写入 + 只清理过期的 assets/*」。
#
# 用法（服务器上，构建产物已解包到 $SRC）：
#   bash deploy/deploy-web.sh /tmp/admin-dist
set -euo pipefail

SRC="${1:-/tmp/admin-dist}"
WEB_ROOT="${WEB_ROOT:-/var/www/admin}"
KEEP_BACKUPS="${KEEP_BACKUPS:-3}"

[ -f "$SRC/index.html" ] || { echo "✗ $SRC/index.html 不存在，构建产物不完整"; exit 1; }
[ -d "$WEB_ROOT" ] || { echo "✗ 站点根 $WEB_ROOT 不存在"; exit 1; }

ts=$(date +%Y%m%d-%H%M%S)
echo "==> 备份 $WEB_ROOT → ${WEB_ROOT}.bak.$ts"
cp -a "$WEB_ROOT" "${WEB_ROOT}.bak.$ts"

echo "==> 覆盖写入构建产物"
cp -a "$SRC/." "$WEB_ROOT/"

echo "==> 清理过期 assets（新构建里已不存在的）"
pruned=0
for f in "$WEB_ROOT"/assets/*; do
  [ -e "$f" ] || continue
  rel="assets/$(basename "$f")"
  if [ ! -e "$SRC/$rel" ]; then rm -f "$f"; pruned=$((pruned+1)); fi
done
echo "    清理 $pruned 个文件"

echo "==> 只保留最近 $KEEP_BACKUPS 份备份"
if ls -1dt "${WEB_ROOT}".bak.* >/dev/null 2>&1; then
  ls -1dt "${WEB_ROOT}".bak.* | tail -n +$((KEEP_BACKUPS + 1)) | xargs -r rm -rf
fi

echo "==> 校验关键文件"
for f in index.html assets static/home/home-bg.jpg images/training/practice-field.svg video/lift1.mp4; do
  if [ -e "$WEB_ROOT/$f" ]; then echo "    ok   $f"; else echo "    ✗ 缺失 $f"; exit 1; fi
done
echo "完成：$(find "$WEB_ROOT" -type f | wc -l) 个文件，$(du -sh "$WEB_ROOT" | cut -f1)"
