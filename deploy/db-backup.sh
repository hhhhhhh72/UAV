#!/bin/bash
# 生产 PostgreSQL 每日备份：docker exec pg_dump → gzip，保留最近 KEEP 份
# 安装：服务器 crontab（ubuntu 用户）每日 03:00 执行本脚本
#   0 3 * * * /home/ubuntu/UAV/deploy/db-backup.sh >> /home/ubuntu/UAV-db-backups/cron.log 2>&1
set -euo pipefail

# 容器名参数化：默认 uav-db-1（docker-compose.yml 中 db 服务已固定 container_name: uav-db-1，
# 不再依赖项目目录推导；多环境部署可用 CONTAINER=xxx ./db-backup.sh 覆盖）
CONTAINER=${CONTAINER:-uav-db-1}
KEEP=${KEEP:-14}
DIR="$HOME/UAV-db-backups"
STAMP=$(date +%Y%m%d-%H%M%S)
LOG="$DIR/backup.log"
# 安全加固（审计 M4）：备份含全量 PII，目录 0700、备份文件 0600——仅备份账号可读
mkdir -p "$DIR"
chmod 700 "$DIR"

sudo docker exec "$CONTAINER" pg_dump -U drone -d drone_platform | gzip > "$DIR/uav-db-$STAMP.sql.gz"
chmod 600 "$DIR/uav-db-$STAMP.sql.gz"

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

# ---- 上传文件（营业执照 / 身份证影像 / 证件照 / 案例视频）----
# 为什么必须备（2026-09-21 彻查发现）：本脚本此前**只** pg_dump，全文 0 处提及 uploads。
# 后果是「库恢复了、文件全没了」—— 而库里还留着引用这些文件的 URL，全站 404；
# 企业资质审核的凭据（营业执照）也随之丢失，这是合规层面的损失，不只是体验问题。
#
# 用**独立日志** uploads-backup.log，不混进 backup.log：ops-status 判断「最近一次
# 完成的备份」读的就是 backup.log 的最后一条 OK，混进上传包会悄悄改变那条判定的语义。
UPLOADS_LOG="$DIR/uploads-backup.log"
UPLOADS_DIR=${UPLOADS_DIR:-$(sudo docker volume inspect uav_uploads --format '{{.Mountpoint}}' 2>/dev/null || echo /var/lib/docker/volumes/uav_uploads/_data)}
uploads_ok=false
if [ -d "$UPLOADS_DIR" ]; then
  tar -czf "$DIR/uav-uploads-$STAMP.tar.gz" -C "$UPLOADS_DIR" . 2>>"$LOG" || true
  chmod 600 "$DIR/uav-uploads-$STAMP.tar.gz" 2>/dev/null || true
  if [ -s "$DIR/uav-uploads-$STAMP.tar.gz" ] && tar -tzf "$DIR/uav-uploads-$STAMP.tar.gz" >/dev/null 2>&1; then
    USIZE=$(du -h "$DIR/uav-uploads-$STAMP.tar.gz" | cut -f1)
    UFILES=$(tar -tzf "$DIR/uav-uploads-$STAMP.tar.gz" | grep -c . || true)
    echo "[$(date '+%F %T')] OK uav-uploads-$STAMP.tar.gz ($USIZE, ${UFILES:-0} 项)" >> "$UPLOADS_LOG"
    uploads_ok=true
  else
    echo "[$(date '+%F %T')] FAIL uploads $STAMP（空包或 tar 校验失败，已删除）" >> "$UPLOADS_LOG"
    rm -f "$DIR/uav-uploads-$STAMP.tar.gz"
  fi
  ls -1t "$DIR"/uav-uploads-*.tar.gz 2>/dev/null | tail -n +$((KEEP + 1)) | while read -r f; do rm -f "$f"; done
else
  echo "[$(date '+%F %T')] FAIL uploads $STAMP（找不到卷目录 $UPLOADS_DIR）" >> "$UPLOADS_LOG"
fi
# 上传包失败也要让 cron 看到非零退出（ops-status 另有一节独立判定）
[ "$uploads_ok" = true ] || exit 1
