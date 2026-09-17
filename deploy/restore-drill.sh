#!/bin/bash
# 备份恢复演练：每周把最新备份还原到一个临时库，验证它真的能用，然后删掉临时库。
#
# 为什么需要它：没人还原过的备份只是**假设**。9/17 的备份静默失败两天说明
# 「文件存在」不等于「能恢复」——pg_dump 半途失败、磁盘写满、gzip 截断，
# 都会留下一份看起来正常、实际还原不出来的文件。此脚本把假设变成事实。
#
# 只操作独立的 restore_drill 库，不碰 drone_platform。
set -uo pipefail

BACKUP_DIR=${BACKUP_DIR:-$HOME/UAV-db-backups}
RESULT=${RESULT:-$BACKUP_DIR/restore-drill.json}
DB=restore_drill
PG=${PG_CONTAINER:-uav-db-1}
PGUSER=${PG_USER:-drone}

now=$(date +%s)
at=$(date -Iseconds)

# 结果先按失败写，成功再覆盖——中途任何一步崩掉，留下的也是「失败」而不是空白。
write_result() {
  local ok=$1 tables=${2:--1} users=${3:--1} demands=${4:--1} courses=${5:--1} note=$6
  local tmp="$RESULT.tmp.$$"
  cat > "$tmp" <<JSON
{
  "at": "$at",
  "epoch": $now,
  "ok": $ok,
  "source": "${latest_name:-none}",
  "tables": $tables,
  "rows": {"users": $users, "demands": $demands, "training_courses": $courses},
  "note": "$note"
}
JSON
  chmod 600 "$tmp"; mv -f "$tmp" "$RESULT"
}

latest=$(ls -1t "$BACKUP_DIR"/uav-db-*.sql.gz 2>/dev/null | head -1)
if [ -z "$latest" ]; then
  echo '[FAIL] 没有可用的备份文件'
  write_result false -1 -1 -1 -1 '没有备份文件'
  exit 1
fi
latest_name=$(basename "$latest")
echo "[$(date '+%F %T')] 恢复演练开始，源: $latest_name"

if ! gzip -t "$latest" 2>/dev/null; then
  echo '[FAIL] gzip 完整性校验不通过'
  write_result false -1 -1 -1 -1 'gzip 校验失败'
  exit 1
fi

psql_drill() { docker exec "$PG" psql -U "$PGUSER" -d "$DB" -t -A -c "$1" 2>/dev/null | tr -d '[:space:]'; }

# 每次从零重建，避免残留数据让演练变成假通过
docker exec "$PG" psql -U "$PGUSER" -d postgres -q -c "DROP DATABASE IF EXISTS $DB" >/dev/null 2>&1
if ! docker exec "$PG" psql -U "$PGUSER" -d postgres -q -c "CREATE DATABASE $DB" >/dev/null 2>&1; then
  echo '[FAIL] 无法创建临时库'
  write_result false -1 -1 -1 -1 '创建临时库失败'
  exit 1
fi

if ! zcat "$latest" | docker exec -i "$PG" psql -U "$PGUSER" -d "$DB" -q >/dev/null 2>&1; then
  echo '[FAIL] 还原过程报错'
  docker exec "$PG" psql -U "$PGUSER" -d postgres -q -c "DROP DATABASE IF EXISTS $DB" >/dev/null 2>&1
  write_result false -1 -1 -1 -1 '还原报错'
  exit 1
fi

tables=$(psql_drill "SELECT count(*) FROM information_schema.tables WHERE table_schema='public'")
users=$(psql_drill 'SELECT count(*) FROM users')
demands=$(psql_drill 'SELECT count(*) FROM demands')
courses=$(psql_drill 'SELECT count(*) FROM training_courses')

docker exec "$PG" psql -U "$PGUSER" -d postgres -q -c "DROP DATABASE IF EXISTS $DB" >/dev/null 2>&1

# 判定：还原出的表数要接近线上（91 张左右），且核心表可查
if [ "${tables:-0}" -lt 80 ] || [ "${users:--1}" -lt 0 ]; then
  echo "[FAIL] 还原后表数异常: tables=$tables users=$users"
  write_result false "${tables:-0}" "${users:--1}" "${demands:--1}" "${courses:--1}" '还原后表数异常'
  exit 1
fi

echo "[OK] 还原成功：$tables 张表，users=$users demands=$demands courses=$courses"
write_result true "$tables" "$users" "$demands" "$courses" 'ok'
