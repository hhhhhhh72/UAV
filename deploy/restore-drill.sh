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
# 加密字段可解密校验所需的工具与密钥来源（2026-09-21）
KEYCHECK=${KEYCHECK:-/root/UAV/tools/keycheck}
ENVFILE=${ENVFILE:-/root/UAV/.env}

now=$(date +%s)
at=$(date -Iseconds)

# 结果先按失败写，成功再覆盖——中途任何一步崩掉，留下的也是「失败」而不是空白。
write_result() {
  local ok=$1 tables=${2:--1} users=${3:--1} demands=${4:--1} courses=${5:--1} note=$6 enc=${7:-skipped}
  # 摘要里可能有引号/反斜杠，落进 JSON 前先剥掉，避免写出坏 JSON
  enc=$(printf '%s' "$enc" | tr -d '"\\' | tr '\n' ' ')
  local tmp="$RESULT.tmp.$$"
  cat > "$tmp" <<JSON
{
  "at": "$at",
  "epoch": $now,
  "ok": $ok,
  "source": "${latest_name:-none}",
  "tables": $tables,
  "rows": {"users": $users, "demands": $demands, "training_courses": $courses},
  "note": "$note",
  "encryption": "$enc"
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

# ---- 加密字段可解密校验（2026-09-21）----
# 「备份完好」不等于「恢复得出可用数据」：PII 在库里是密文，而密钥只在被备份的机器
# 之外（/root/UAV/.env）。密钥一旦丢失或被轮换，备份里的身份证/手机号就永久解不开 ——
# 此前没有任何检查会发现，因为 gzip 校验过、表数也数过，全绿。
# 这里用 **API 同一份 crypto 实现**（cmd/keycheck 链接 internal/crypto）对还原出来的
# 临时库逐行尝解密，把「密钥与这份备份是否配套」变成一条可验证的结论。
# 缺工具或缺密钥一律判失败：检查跑不起来的时候，最不该做的就是沉默放过。
enc_ok=true
enc_summary=""
# 工具新鲜度：crypto 实现改了而工具没重建，演练校验的就是**旧逻辑**——那和没校验一样。
# 部署会把 cmd/keycheck/main.go 一起带到 /root/UAV，所以源文件比二进制新 = 必须重建。
KEYCHECK_SRC=${KEYCHECK_SRC:-/root/UAV/cmd/keycheck/main.go}
if [ ! -x "$KEYCHECK" ]; then
  enc_ok=false; enc_summary="缺少 $KEYCHECK，无法校验加密字段"
elif [ -f "$KEYCHECK_SRC" ] && [ "$KEYCHECK_SRC" -nt "$KEYCHECK" ]; then
  enc_ok=false; enc_summary="$KEYCHECK_SRC 比工具新 —— 工具过期，需在开发机重建后重传"
else
  enc_key=$(grep -E '^ENCRYPTION_KEY=' "$ENVFILE" 2>/dev/null | head -1 | cut -d= -f2-)
  if [ -z "$enc_key" ]; then
    enc_ok=false; enc_summary="未能从 $ENVFILE 取到 ENCRYPTION_KEY"
  else
    docker cp "$KEYCHECK" "$PG:/tmp/keycheck" >/dev/null 2>&1
    docker exec "$PG" chmod +x /tmp/keycheck >/dev/null 2>&1
    enc_out=$(docker exec -e ENCRYPTION_KEY="$enc_key" "$PG" /tmp/keycheck "postgres://$PGUSER@/$DB?host=/var/run/postgresql" 2>&1)
    enc_rc=$?
    docker exec "$PG" rm -f /tmp/keycheck >/dev/null 2>&1
    enc_summary=$(printf '%s' "$enc_out" | tr '\n' ' ' | cut -c1-400)
    [ "$enc_rc" -eq 0 ] || enc_ok=false
  fi
fi

docker exec "$PG" psql -U "$PGUSER" -d postgres -q -c "DROP DATABASE IF EXISTS $DB" >/dev/null 2>&1

# 判定：还原出的表数要接近线上（91 张左右），且核心表可查
if [ "${tables:-0}" -lt 80 ] || [ "${users:--1}" -lt 0 ]; then
  echo "[FAIL] 还原后表数异常: tables=$tables users=$users"
  write_result false "${tables:-0}" "${users:--1}" "${demands:--1}" "${courses:--1}" '还原后表数异常' "$enc_summary"
  exit 1
fi

if [ "$enc_ok" != true ]; then
  echo "[FAIL] 加密字段校验未通过: $enc_summary"
  write_result false "$tables" "$users" "$demands" "$courses" '加密字段无法解密' "$enc_summary"
  exit 1
fi

echo "[OK] 还原成功：$tables 张表，users=$users demands=$demands courses=$courses"
echo "[OK] 加密校验：$enc_summary"
write_result true "$tables" "$users" "$demands" "$courses" 'ok' "$enc_summary"
