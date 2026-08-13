#!/bin/bash
# 生产 PostgreSQL 每日备份：docker exec pg_dump → gzip，保留最近 KEEP 份
# 安装：服务器 crontab（ubuntu 用户）每日 03:00 执行本脚本
#   0 3 * * * /home/ubuntu/UAV/deploy/db-backup.sh >> /home/ubuntu/UAV-db-backups/cron.log 2>&1
set -euo pipefail

KEEP=${KEEP:-14}
DIR="$HOME/UAV-db-backups"
STAMP=$(date +%Y%m%d-%H%M%S)
LOG="$DIR/backup.log"
mkdir -p "$DIR"

sudo docker exec uav-db-1 pg_dump -U drone -d drone_platform | gzip > "$DIR/uav-db-$STAMP.sql.gz"

# 空文件视为失败（防 pg_dump 静默失败留下空档备份）
if [ ! -s "$DIR/uav-db-$STAMP.sql.gz" ]; then
  echo "[$(date '+%F %T')] FAIL empty $STAMP" >> "$LOG"
  rm -f "$DIR/uav-db-$STAMP.sql.gz"
  exit 1
fi

SIZE=$(du -h "$DIR/uav-db-$STAMP.sql.gz" | cut -f1)
echo "[$(date '+%F %T')] OK uav-db-$STAMP.sql.gz ($SIZE)" >> "$LOG"

# 只保留最近 KEEP 份
ls -1t "$DIR"/uav-db-*.sql.gz 2>/dev/null | tail -n +$((KEEP + 1)) | while read -r f; do rm -f "$f"; done
