#!/bin/bash
# 磁盘保洁：回收 Docker 构建缓存、历史 SPA 备份、过期部署包、journal。
#
# 为什么需要它（2026-09-17）：磁盘涨到 74%，其中 20.46 GB 是 docker build 的构建缓存
# （每次 docker compose build api 都留一层，累积了几十次部署），另有 972 MB 是
# 27 个历次 SPA 部署备份、128 MB 历史部署包。系统里没有任何自动清理任务，只增不减。
#
# 保守策略：只删可再生/已入库的东西——代码在 git，镜像可由 Dockerfile 重建，
# 数据库有独立备份。每类都保留最近几份以备回滚。
set -uo pipefail

KEEP_SPA_BAK=$(printf %s "${KEEP_SPA_BAK:-3}")
KEEP_TARBALL_DAYS=$(printf %s "${KEEP_TARBALL_DAYS:-7}")
BUILD_CACHE_KEEP=$(printf %s "${BUILD_CACHE_KEEP:-2GB}")
JOURNAL_MAX=$(printf %s "${JOURNAL_MAX:-50M}")

before=$(df -P / | awk 'NR==2 {print $4}')
echo "[$(date '+%F %T')] 保洁开始，可用空间 $((before/1024)) MB"

echo '--- 1 Docker 构建缓存 ---'
docker system df 2>/dev/null | grep -i 'build cache' | sed 's/^/    /'
docker builder prune -f --keep-storage "$BUILD_CACHE_KEEP" 2>&1 | tail -2 | sed 's/^/    /'

echo "--- 2 历史 SPA 备份（保留最近 $KEEP_SPA_BAK 份）---"
kept=0
for d in $(ls -dt /var/www/admin.bak.* 2>/dev/null); do
  case "$d" in /var/www/admin.bak.*) ;; *) continue ;; esac
  kept=$((kept+1))
  if [ "$kept" -gt "$KEEP_SPA_BAK" ]; then rm -rf "$d"; fi
done
echo "    保留 $KEEP_SPA_BAK 份，其余已删"

echo "--- 3 过期部署包（超过 $KEEP_TARBALL_DAYS 天）---"
n=0
for f in /root/*.tar.gz /root/UAV/*.tar.gz; do
  [ -f "$f" ] || continue
  if [ -n "$(find "$f" -mtime +"$KEEP_TARBALL_DAYS" 2>/dev/null)" ]; then rm -f "$f"; n=$((n+1)); fi
done
echo "    删除 $n 个"

echo "--- 4 journal 压缩到 $JOURNAL_MAX ---"
journalctl --vacuum-size="$JOURNAL_MAX" 2>&1 | tail -1 | sed 's/^/    /'

after=$(df -P / | awk 'NR==2 {print $4}')
echo "[$(date '+%F %T')] 保洁完成，释放 $(( (after-before)/1024 )) MB，现可用 $((after/1024)) MB"
