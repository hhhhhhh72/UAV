#!/bin/bash
# 日报正文生成器的回归演练（2026-09-22）。
#
# 为什么要有它：负责人问「日报为什么带很多省略号」—— 真跑出来 4 天里 16 处，而且断在词中间：
#   「证书自动续期：根因是域名未备案导致 HTTP-0…」（HTTP-01 被切断）
#   「日报取数通道加固（重试窗口 2.5h + SKI…」（SKIP_FETCH 被切断）
#   「研学容量的库级兜底（FOR UPDATE）+ 场…」（括号规则只认半角 '('，中文全用全角 '（'，从未生效）
# 这些都是"看着像正常、实际读不成句"的假绿 —— 日报没人会逐字核对，所以必须由机器守。
#
# 做法：造一个**临时 git 仓库**，把上面这些真实踩过坑的提交标题原样提交进去，
# 跑真正的 work_report.py，断言输出里没有"被切断的残句"、字数仍在预算内、
# 且关键短语被完整保留。全程在 mktemp 目录里，不碰生产裸库与日报产物。
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
mk() { echo "$RANDOM$1" > "f$RANDOM.txt"; git add -A; git commit -q -m "$1"; }

# ① 冒号后面是解释 → 应留「证书自动续期」，且不能把 HTTP-01 切断
mk 'feat(cert): 证书自动续期：根因是域名未备案导致 HTTP-01 不可能，改用 acme.sh + TLS-ALPN-01'
# ② 括号补充要整段去掉（全角括号！），且 + 并列要拆成两条
mk 'feat(ops): 研学容量的库级兜底（FOR UPDATE）+ 场地/场馆时段重叠补上库级兜底（migration 000120）'
mk 'chore(ops): 日报取数通道加固（重试窗口 2.5h + SKIP_FETCH）'
# ③ Revert "..." 要还原成「回退 xxx」
mk 'Revert "feat(ui): 供给大厅卡片加入场动画"'
# ④ 半角冒号在时间/比例里**不能**当解释分隔符（19:30、1:1）
mk 'fix(ui): 日报发送时间 19:30 → 17:30（负责人要求提前）'
mk 'fix(ui): 商品主图 1:1/3:4 引导'
# ⑤ 破折号后面是解释
mk 'feat(ops): 新增 instances 一节 —— 把「只允许一个 API 实例」变成可验证的绊线'
# ⑥ 正常条目不能被改坏
mk 'fix(escrow): 售后重复退款也下沉到数据库'

DAY=$(date +%F)
out=$(python3 "$GEN" "$R" "$DAY" "" 2>&1) || { echo "生成器自身报错：$out"; exit 1; }

echo "演练仓库：$R"
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

# 断言 3：冒号前的主干保留，解释部分丢掉
printf '%s' "$out" | grep -q '证书自动续期' && ! printf '%s' "$out" | grep -q '根因是域名未备案' \
  && ok "冒号后是解释：留「证书自动续期」，丢掉「根因是域名未备案…」" \
  || bad "冒号主干提取不对：$(printf '%s' "$out" | grep '证书' || echo '(整条丢了)')"

# 断言 4：全角括号补充被去掉，且 + 并列拆成两条独立工作项
if printf '%s' "$out" | grep -q 'FOR UPDATE'; then
  bad "全角括号里的补充说明没去掉（FOR UPDATE 还在）"
elif printf '%s' "$out" | grep -q '研学容量的库级兜底' && printf '%s' "$out" | grep -q '场地/场馆时段重叠补上库级兜底'; then
  ok "全角括号被去掉，且 + 并列拆成了两条独立工作项"
else
  bad "括号/+ 拆分结果不对"
fi

# 断言 5：Revert 还原成「回退 xxx」
printf '%s' "$out" | grep -q '回退 供给大厅卡片加入场动画' \
  && ok 'Revert "…" 还原成「回退 供给大厅卡片加入场动画」' \
  || bad "Revert 没被还原：$(printf '%s' "$out" | grep '回退' || echo '(没有回退条目)')"

# 断言 6：半角冒号在时间里不当分隔符（19:30 → 17:30 必须完整）
printf '%s' "$out" | grep -q '19:30 → 17:30' \
  && ok "半角冒号没被当成解释分隔符（19:30 → 17:30 完整保留）" \
  || bad "时间被冒号规则切坏了：$(printf '%s' "$out" | grep '日报发送时间' || echo '(整条丢了)')"

# 断言 7：破折号后是解释
if printf '%s' "$out" | grep -q '可验证的绊线'; then
  bad "破折号后面的解释没去掉"
else
  printf '%s' "$out" | grep -q '新增 instances 一节' && ok "破折号后是解释：留「新增 instances 一节」" || bad "破折号规则不对"
fi

echo
echo "== 结论：$pass 项通过，$fail 项失败 =="
[ "$fail" = 0 ] || exit 1
