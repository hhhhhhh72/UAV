#!/bin/bash
# 日报正文生成器的回归演练（2026-09-22 建，2026-09-23 扩）。
#
# 守两类已经真实发生过的问题：
#  ① 「日报为什么带很多省略号」：真跑 4 天 16 处，而且断在词中间（HTTP-01 → HTTP-0、
#     SKIP_FETCH → SKI）。根因是硬截 24 字，且「别切在半个括号里」只认半角 '('，
#     而中文标题全用全角 '（' —— 那条规则从未生效。
#  ② 「内容好少，把昨天干的事也编进来」：自然日窗口在 17:30 收口，17:30 之后干的活
#     当班看不到、下一班也不管；预算又是**平均分**给每个模块，有的模块没几条，
#     额度白白空着（实测 09-22 只用了 161/300）。
# 日报没人会逐字核对，所以这两种「看着正常、实际不对」的假绿必须由机器守。
#
# 做法：造一个**临时 git 仓库**，先铺一批"日常"提交，再放真实踩过坑的提交标题，
# 最后补一条**昨天 18:00** 的提交用来验窗口；跑真正的 work_report.py 后逐条断言。
# 全程在 mktemp 目录里，不碰生产裸库与日报产物。
set -uo pipefail

T=$(mktemp -d /tmp/wr-drill.XXXXXX)
trap 'rm -rf "$T"' EXIT
# 默认取脚本自己旁边的生成器 —— 服务器上是 /root/UAV/deploy/work_report.py，
# CI 里是仓库根的 deploy/work_report.py，同一份脚本两边都能跑
GEN=${GEN:-$(dirname "$(readlink -f "$0")")/work_report.py}
[ -f "$GEN" ] || { echo "找不到 $GEN"; exit 1; }

pass=0; fail=0
ok()  { echo "  ok   $1"; pass=$((pass+1)); }
bad() { echo "  FAIL $1"; fail=$((fail+1)); }

R="$T/repo"
git init -q "$R" && cd "$R"
git config user.email drill@local && git config user.name drill
# 用真实目录结构提交，模块归类才会走 area_of（全用 .txt 会统统落到「其他」）
mk() { local p="$1" s="$2"; mkdir -p "$(dirname "$p")"; echo "$RANDOM$s" >> "$p"; git add -A; git commit -q -m "$s"; }

# ---- 昨天 18:00 的提交：必须**最先建**，让历史保持时间递增 ----
# 为什么强调顺序（2026-09-23 踩到）：git log --since 为了性能会**边走边停** ——
# 一旦走到一条比 --since 还旧的提交就停止遍历。把一条回溯日期的提交放在 HEAD
#（或任何比它父提交旧的位置），整段历史会在它那里被截断：实测同一条命令
#「--since=今天 00:00」返回 0 条、「--since=昨天 17:30」返回 29 条。
# 生产上如果哪天有人 rebase/改提交日期，日报就可能静默少报 —— 这里把顺序钉住。
YDAY=$(date -d 'yesterday' +%F)
mkdir -p internal/escrow && echo "$RANDOM" >> internal/escrow/night.go && git add -A
GIT_AUTHOR_DATE="$YDAY 18:00:00" GIT_COMMITTER_DATE="$YDAY 18:00:00" \
  git commit -q -m 'fix(escrow): 昨天下班后补的资金对账脚本'

# ---- 再铺一批"日常"提交（较旧，先占位）----
for i in 1 2 3 4 5; do
  mk internal/service/x$i.go        "fix(escrow): 资金对账第 $i 项"
  mk internal/httpapi/y$i.go        "feat(api): 后端接口收敛第 $i 项"
  mk deploy/z$i.sh                  "chore(ops): 运维脚本加固第 $i 项"
  mk docs/d$i.md                    "docs: 文档补充第 $i 项"
done

# ---- 真实踩过坑的提交标题（较新，预算充足时优先被选中）----
mk internal/crypto/c.go   'feat(cert): 证书自动续期：根因是域名未备案导致 HTTP-01 不可能，改用 acme.sh + TLS-ALPN-01'
mk internal/study/s.go    'feat(ops): 研学容量的库级兜底（FOR UPDATE）+ 场地/场馆时段重叠补上库级兜底（migration 000120）'
mk deploy/work_report.py  'chore(ops): 日报取数通道加固（重试窗口 2.5h + SKIP_FETCH）'
mk miniprogram/ui/a.vue   'Revert "feat(ui): 供给大厅卡片加入场动画"'
mk miniprogram/ui/b.vue   'fix(ui): 日报发送时间 19:30 → 17:30（负责人要求提前）'
mk frontend/src/c.vue     'fix(ui): 商品主图 1:1/3:4 引导'
mk deploy/ops-status.sh   'feat(ops): 新增 instances 一节 —— 把「只允许一个 API 实例」变成可验证的绊线'
mk internal/escrow/e.go   'fix(escrow): 售后重复退款也下沉到数据库'

TOTAL=$(git rev-list --count HEAD)

DAY=$(date +%F)
out=$(python3 "$GEN" "$R" "$DAY" "" 2>&1) || { echo "生成器自身报错：$out"; exit 1; }

echo "演练仓库：$R（共 $TOTAL 个提交，其中 1 个是昨天的）"
echo "---- 生成结果 ----"
printf '%s\n' "$out"
echo "------------------"

# 断言 1：省略号不能紧跟在 ASCII 字母/数字后面（那就是切断了 HTTP-01 / SKIP_FETCH 这类词）
if printf '%s' "$out" | grep -qE '[A-Za-z0-9]…'; then
  bad "省略号跟在英文/数字后面 —— 又把词切断了：$(printf '%s' "$out" | grep -oE '[A-Za-z0-9]+…' | tr '\n' ' ')"
else
  ok "没有把英文/数字词切断（HTTP-01 / SKIP_FETCH 这类）"
fi

# 断言 2：整份仍在 300 字预算内
n=${#out}
if [ "$n" -le 300 ]; then ok "字数 $n ≤ 300（预算内）"; else bad "字数 $n 超预算 300"; fi

# 断言 3：条目够的时候预算要**用起来**（均分时代每天只用到一半）
if [ "$n" -ge 250 ]; then
  ok "预算用起来了（字数 $n ≥ 250，不再大面积空着）"
else
  bad "字数只有 $n —— 条目明明够，预算又被平均分掉了"
fi

# 断言 4：窗口必须覆盖**昨天 17:30 之后**的提交（用规模行的提交数对账，不受预算挤占影响）
reported=$(printf '%s' "$out" | sed -n 's/.*提交 \([0-9]*\) 个.*/\1/p' | head -1)
if [ "$reported" = "$TOTAL" ]; then
  ok "规模行说提交 $reported 个 = 仓库里的 $TOTAL 个 —— 昨天傍晚那笔也算进来了（滚动窗口）"
else
  bad "规模行说提交 $reported 个，实际有 $TOTAL 个 —— 昨天 17:30 之后的活没进窗口"
fi

# 断言 5：规模行要写明窗口起点，别让人以为只统计了自然日
printf '%s' "$out" | grep -qE '（[0-9]{2}-[0-9]{2} 17:30 起）' \
  && ok "规模行写明了窗口起点（xx-xx 17:30 起）" \
  || bad "规模行没写窗口起点：$(printf '%s' "$out" | sed -n '2p')"

# 断言 6：冒号前的主干保留，解释部分丢掉
printf '%s' "$out" | grep -q '证书自动续期' && ! printf '%s' "$out" | grep -q '根因是域名未备案' \
  && ok "冒号后是解释：留「证书自动续期」，丢掉「根因是域名未备案…」" \
  || bad "冒号主干提取不对：$(printf '%s' "$out" | grep '证书' || echo '(整条丢了)')"

# 断言 7：全角括号补充被去掉。
# 「+ 拆成两条」不在这个仓库里断言 —— 预算有限时后一半可能被挤掉，那是正常的取舍；
# 拆没拆要在一个只有这一条提交的仓库里验（断言 7b），才不受预算影响。
if printf '%s' "$out" | grep -q 'FOR UPDATE'; then
  bad "全角括号里的补充说明没去掉（FOR UPDATE 还在）"
elif printf '%s' "$out" | grep -q '研学容量的库级兜底'; then
  ok "全角括号里的补充说明被去掉（FOR UPDATE / migration 000120 都不出现）"
else
  bad "研学容量那条整个丢了"
fi

# 断言 7b：单独一个只有该提交的仓库 —— 两条必须都出现（证明 + 真被拆开了）
R2="$T/repo2"
git init -q "$R2"
(
  cd "$R2" && git config user.email d@l && git config user.name d
  echo x > a.go && git add -A
  git commit -q -m 'feat(ops): 研学容量的库级兜底（FOR UPDATE）+ 场地/场馆时段重叠补上库级兜底（migration 000120）'
)
out2=$(python3 "$GEN" "$R2" "$DAY" "" 2>&1)
if printf '%s' "$out2" | grep -q '研学容量的库级兜底' && printf '%s' "$out2" | grep -q '场地/场馆时段重叠补上库级兜底' \
   && ! printf '%s' "$out2" | grep -q 'FOR UPDATE'; then
  ok "单条提交里 + 并列被拆成两条独立工作项（且括号补充已去掉）"
else
  bad "十/加号拆分不对：$(printf '%s' "$out2" | sed -n '2,3p' | tr '\n' ' ')"
fi

# 断言 8：Revert 还原成「回退 xxx」
printf '%s' "$out" | grep -q '回退 供给大厅卡片加入场动画' \
  && ok 'Revert "…" 还原成「回退 供给大厅卡片加入场动画」' \
  || bad "Revert 没被还原：$(printf '%s' "$out" | grep '回退' || echo '(没有回退条目)')"

# 断言 9：半角冒号在时间里不当分隔符（19:30 → 17:30 必须完整）
printf '%s' "$out" | grep -q '19:30 → 17:30' \
  && ok "半角冒号没被当成解释分隔符（19:30 → 17:30 完整保留）" \
  || bad "时间被冒号规则切坏了：$(printf '%s' "$out" | grep '日报发送时间' || echo '(整条丢了)')"

# 断言 10：破折号后是解释
if printf '%s' "$out" | grep -q '可验证的绊线'; then
  bad "破折号后面的解释没去掉"
else
  printf '%s' "$out" | grep -q '新增 instances 一节' && ok "破折号后是解释：留「新增 instances 一节」" || bad "破折号规则不对"
fi

echo
echo "== 结论：$pass 项通过，$fail 项失败 =="
[ "$fail" = 0 ] || exit 1
