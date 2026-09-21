#!/bin/bash
# 运维快照的回归演练：在**临时目录**里造出各种备份现场，验证 ops-status.sh 判得对。
#
# 为什么单独有这个脚本（2026-09-20）：03:02 那条「运维快照异常：backup」是假阳性，
# 根因是快照按 mtime 取「最新文件」，而 03:00 的备份与 03:00 的快照同秒启动，
# 快照读到了 pg_dump 正在写、还没写完的半个文件 → gzip -t 失败 → 判损坏。
# 这不是偶发：crontab 里备份条目排在前面，顺序固定，所以**每天**都会发一对
# 「异常 + 10 分钟后恢复」。这类噪音会把真故障淹掉，必须有回归测试钉住。
#
# 全程 OUT / BACKUP_DIR / BACKUP_LOG / DRILL 都指向临时文件，碰不到生产快照。
set -uo pipefail

T=$(mktemp -d /tmp/ops-drill.XXXXXX)
trap 'rm -rf "$T"' EXIT
PROD_SNAPSHOT=/var/www/ops-status.json
[ -f "$PROD_SNAPSHOT" ] || { echo "找不到生产快照 $PROD_SNAPSHOT"; exit 1; }
prod_before=$(stat -c %Y "$PROD_SNAPSHOT")

pass=0; fail=0
ok()   { echo "  ok   $1"; pass=$((pass+1)); }
bad()  { echo "  FAIL $1"; fail=$((fail+1)); }

D="$T/backups"; mkdir -p "$D"

# 造一份**能通过 gzip -t** 的完整备份
mk_ok() { printf 'payload-%s' "$2" | gzip > "$D/$1"; }
# 造一份**截断**的备份：去掉 gzip 尾部，gzip -t 必失败（模拟「正在写」）
mk_partial() { printf 'payload-%s' "$2" | gzip > "$D/$1"; truncate -s -6 "$D/$1"; }

# 取快照里的字段
field() { python3 -c "import json,sys;d=json.load(open('$T/out.json'));print(d['backup'].get('$1'))"; }
run() { OUT="$T/out.json" BACKUP_DIR="$D" BACKUP_LOG="${1:-$D/backup.log}" UPLOADS_LOG="$D/uploads-backup.log" DRILL="$T/none.json" \
        bash /root/UAV/deploy/ops-status.sh >/dev/null 2>&1; }

# 上传文件备份的"健康"现场（2026-09-21 起 backup 一节要求库备份**和**上传包都在）
mkdir -p "$T/up"; printf 'license-image' > "$T/up/idcard.jpg"
mk_uploads_ok() {
  tar -czf "$D/uav-uploads-$1.tar.gz" -C "$T/up" idcard.jpg 2>/dev/null
  printf '[2026-01-01 03:00:02] OK uav-uploads-%s.tar.gz (1K, 2 项)\n' "$1" > "$D/uploads-backup.log"
}

echo "演练目录：$T"
echo "生产快照：$PROD_SNAPSHOT（演练期间不会被改动）"
echo
echo "== 备份判定演练 =="

# ① 今天的 bug：目录里最新的文件是「正在写」的半个，backup.log 指向上一次完整的
rm -f "$D"/*.sql.gz "$D"/backup.log
mk_ok uav-db-20260101-030000.sql.gz a
touch -d '24 hours ago' "$D/uav-db-20260101-030000.sql.gz"
printf '[2026-01-01 03:00:02] OK uav-db-20260101-030000.sql.gz (92K)\n' > "$D/backup.log"
mk_partial uav-db-20260102-030000.sql.gz b      # 刚写完一半，mtime = 现在
mk_uploads_ok 20260101-030000
run
if [ "$(field file)" = "uav-db-20260101-030000.sql.gz" ] && [ "$(field ok)" = "True" ]; then
  ok "写入中的新文件被跳过，选了最近一次**完成**的备份"
else
  bad "选中了 $(field file) / ok=$(field ok) —— 又读了正在写的半个文件"
fi

# ② 备份真的过期（上次完成在 40 小时前）仍必须报出来
rm -f "$D"/*.sql.gz "$D"/backup.log
mk_ok uav-db-20260101-030000.sql.gz c
touch -d '40 hours ago' "$D/uav-db-20260101-030000.sql.gz"
printf '[2026-01-01 03:00:02] OK uav-db-20260101-030000.sql.gz (92K)\n' > "$D/backup.log"
run
if [ "$(field ok)" = "False" ]; then ok "真过期（40h）仍判不达标"; else bad "真过期却没报：ok=$(field ok)"; fi

# ③ 完整的那份虽然存在，但内容损坏 → 完整性检查仍要抓住
rm -f "$D"/*.sql.gz "$D"/backup.log
mk_partial uav-db-20260101-030000.sql.gz d
touch -d '2 hours ago' "$D/uav-db-20260101-030000.sql.gz"
printf '[2026-01-01 03:00:02] OK uav-db-20260101-030000.sql.gz (92K)\n' > "$D/backup.log"
run
if [ "$(field integrity)" = "corrupt" ] && [ "$(field ok)" = "False" ]; then
  ok "损坏的备份仍被 gzip -t 抓住"
else
  bad "损坏备份没被抓住：integrity=$(field integrity) ok=$(field ok)"
fi

# ④ 没有 backup.log 的老环境：兜底逻辑也不能挑中「还在写」的那个
rm -f "$D"/*.sql.gz "$D"/backup.log
mk_ok uav-db-20260101-030000.sql.gz e
touch -d '5 minutes ago' "$D/uav-db-20260101-030000.sql.gz"
mk_partial uav-db-20260102-030000.sql.gz f      # 最新，但还在写
mk_uploads_ok 20260101-030000
run "$T/does-not-exist.log"
if [ "$(field file)" = "uav-db-20260101-030000.sql.gz" ] && [ "$(field ok)" = "True" ]; then
  ok "无 backup.log 时按 settle 时间跳过写入中的文件"
else
  bad "兜底逻辑选中了 $(field file) / ok=$(field ok)"
fi

# ⑤ 只备了数据库、没有上传包 → backup 一节必须失守
#    （上传的营业执照/身份证影像是不可再生的原始凭据，只备库等于它们全丢）
rm -f "$D"/*.sql.gz "$D"/backup.log "$D"/uav-uploads-*.tar.gz "$D"/uploads-backup.log
mk_ok uav-db-20260101-030000.sql.gz g
touch -d '2 hours ago' "$D/uav-db-20260101-030000.sql.gz"
printf '[2026-01-01 03:00:02] OK uav-db-20260101-030000.sql.gz (92K)\n' > "$D/backup.log"
run
if [ "$(field integrity)" = "ok" ] && [ "$(field ok)" = "False" ] && [ "$(field uploads_integrity)" = "missing" ]; then
  ok "缺上传包时判失守（库备份本身完好：integrity=$(field integrity)）"
else
  bad "缺上传包却没报：库 integrity=$(field integrity) backup.ok=$(field ok) uploads=$(field uploads_integrity)"
fi

echo
echo "== 实例数判定演练（进程内键锁的前提：只允许一个 API 实例）=="
ifield() { python3 -c "import json;d=json.load(open('$T/out.json'));print((d.get('instances') or {}).get('$1'))"; }
toplevel() { python3 -c "import json;d=json.load(open('$T/out.json'));print(d.get('ok'))"; }
run_env() { OUT="$T/out.json" BACKUP_DIR="$D" BACKUP_LOG="$D/backup.log" DRILL="$T/none.json" \
        env "$@" bash /root/UAV/deploy/ops-status.sh >/dev/null 2>&1; }

# ⑤ 阈值收紧到 0：真实的单实例也必须判失守 —— 证明这一节真的接进了顶层 ok，不是摆设
run_env INSTANCE_MAX=0
if [ "$(ifield ok)" = "False" ] && [ "$(toplevel)" = "False" ]; then
  ok "阈值 0 时 instances 与顶层 ok 一起变红"
else
  bad "instances 没接进顶层判定：instances.ok=$(ifield ok) top=$(toplevel)"
fi

# ⑥ 查不到连接来源时必须报（宁可报，不可假绿 —— 与备份/告警那两次同一个口径）
run_env ESCROW_DB=definitely_not_a_db
if [ "$(ifield ok)" = "False" ]; then
  ok "取不到数据库连接来源时判失守"
else
  bad "取不到数据却判绿：instances.ok=$(ifield ok)"
fi

# ⑦ 正常环境必须为绿，且报出真实计数
run_env
if [ "$(ifield ok)" = "True" ]; then
  ok "真实单实例判绿（api 容器 $(ifield api_containers) 个、连接来源 $(ifield db_client_addrs) 个）"
else
  bad "正常环境却判失守：$(ifield detail)"
fi

echo
echo "== 计数列一致性判定演练 =="
cfield() { python3 -c "import json;d=json.load(open('$T/out.json'));print((d.get('counters') or {}).get('$1'))"; }

# ⑧ 查不到数据时必须报（宁可报，不可假绿）
run_env ESCROW_DB=definitely_not_a_db
if [ "$(cfield ok)" = "False" ]; then
  ok "计数列查不到数据时判失守"
else
  bad "计数列取不到数据却判绿：ok=$(cfield ok)"
fi

# ⑨ 正常环境必须为绿且四项都为 0
run_env
if [ "$(cfield ok)" = "True" ] && [ "$(cfield drift_total)" = "0" ]; then
  ok "真实环境计数列无漂移（赛事 $(cfield competition) / 活动 $(cfield event) / 课程余位 $(cfield course_remain) / 课程少算 $(cfield course_undercount)）"
else
  bad "计数列判定异常：ok=$(cfield ok) drift=$(cfield drift_total)"
fi

echo
echo "== 结论：$pass 项通过，$fail 项失败 =="
prod_after=$(stat -c %Y "$PROD_SNAPSHOT")
if [ "$prod_before" = "$prod_after" ]; then
  echo "生产快照未被改动（mtime 一致）ok"
else
  echo "FAIL 生产快照被改动了，演练脚本有 bug"; fail=$((fail+1))
fi
[ "$fail" = 0 ] || exit 1