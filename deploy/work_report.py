# 工作日报的正文生成器（服务器侧，唯一一份实现）。
#
# 为什么放在服务器、而不是开发机上（2026-09-18 用户问「我云服务不关机每天都有？」）：
#   日报的内容来自 git 提交记录。开发机不在线就没有数据，服务器 24 小时开着却无从下手。
#   实测服务器能直连 GitHub（`git ls-remote` 成功），于是每天 `git fetch` 一份**只含提交
#   与目录、不含文件内容**的裸库（1.4MB，本地 .git 是 524MB），日报就与开发机是否开机无关了。
#
# 用法：python3 work_report.py <裸库路径> <YYYY-MM-DD> <部署节点文字>
#   提交记录由本脚本自己调 git 取（不走 stdin —— 用 heredoc 喂 python 会把 stdin 抢走）。
import re
import subprocess
import sys
import time

repo, day_arg, deploy_note = sys.argv[1], sys.argv[2], sys.argv[3]

BUDGET = 300   # 汇报要求：全文不超过 300 字

# ---------- 1. 取当日提交（连同每个提交改了哪些文件） ----------
today = time.strftime('%Y-%m-%d')
if day_arg == today:
    since, until = day_arg + ' 00:00:00', time.strftime('%Y-%m-%d %H:%M:%S')
else:
    since = day_arg + ' 00:00:00'
    until = time.strftime('%Y-%m-%d', time.localtime(time.mktime(time.strptime(day_arg, '%Y-%m-%d')) + 86400)) + ' 00:00:00'

raw = subprocess.run(
    ['git', '-C', repo, 'log', '--since=' + since, '--until=' + until,
     '--pretty=format:@@%h|%s', '--name-only', '--no-merges', 'master'],
    capture_output=True, text=True, errors='replace').stdout.splitlines()

commits = []
for line in raw:
    if line.startswith('@@'):
        sha, _, subject = line[2:].partition('|')
        commits.append({'sha': sha, 'subject': subject, 'files': []})
    elif line.strip() and commits:
        commits[-1]['files'].append(line.strip())

# ---------- 2. 归类到模块 ----------
# 先看 scope（提交标题里 fix(escrow) 这种），scope 不认识再按改动文件所在区域。
# 顺序即优先级：预算不够时先牺牲靠后的，保住资金/支付/后端。
SCOPE = {
    'escrow': '资金', 'money': '资金', 'deposit': '资金', 'withdraw': '资金',
    'pay': '支付', 'payment': '支付', 'refund': '支付', 'wechatpay': '支付',
    'ops': '运维', 'deploy': '运维', 'ci': '运维', 'backup': '运维', 'alert': '运维',
    'admin': '管理后台', 'publish': '管理后台', 'web': '管理后台', 'ui': '管理后台',
    'frontend': '管理后台', 'mp': '小程序', 'miniprogram': '小程序',
    'docs': '文档',
    'errors': '后端', 'api': '后端', 'auth': '后端', 'db': '后端', 'service': '后端',
}
ORDER = ['资金', '支付', '后端', '管理后台', '小程序', '运维', '文档', '其他']
PREFIX = re.compile(r'^(fix|feat|style|chore|docs|refactor|perf|test|ops|ci)(\([^)]*\))?[:：]\s*')


def area_of(path):
    if path.startswith(('cmd/', 'internal/', 'migrations/')):
        return '后端'
    if path.startswith('frontend/'):
        return '管理后台'
    if path.startswith('miniprogram/'):
        return '小程序'
    if path.startswith(('deploy/', 'scripts/', '.github/')):
        return '运维'
    if path.startswith('docs/') or path in ('CLAUDE.md', 'README.md'):
        return '文档'
    return '其他'


def module_of(c):
    s = c['subject']
    if '(' in s and ')' in s and s[:s.index('(')].isalpha():
        scope = s[s.index('(') + 1:s.index(')')].lower()
        for k in re.split(r'[/,]', scope):
            k = k.strip()
            if k in SCOPE:
                return SCOPE[k]
    if not c['files']:
        return '其他'
    tally = {}
    for f in c['files']:
        a = area_of(f)
        tally[a] = tally.get(a, 0) + 1
    return max(tally.items(), key=lambda kv: kv[1])[0]


# 括号里的都是补充说明（「FOR UPDATE」「含存量行补加密」「migration 000120」），标题里不需要。
# 全角半角都算 —— 此前 cap() 只认半角 '('，而中文提交标题全用全角 '（'，
# 于是「别切在半个括号里」这条规则**从来没生效过**。
PARENS = re.compile(r'[（(][^（()）]*[)）]')
# 一个提交里用 ; 或 + 并列的多件事，拆成独立工作项。**必须在去掉括号之后**拆，
# 否则「密钥与备份配套性校验（keycheck + 演练）」会被括号里的 + 拆坏。
SPLIT = re.compile(r'[;；]|\s*[+＋]\s*')
WORDISH = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-'
CJK = re.compile(r'[\u4e00-\u9fff]')
# 自然分隔符：真要截，优先截在这些地方，读起来还是完整短语
SEPS = ('，', '、', '；', ' ', '/', '·', ',')


def headline(s, n=24):
    """把一条工作项压成**标题**，而不是切一半。

    为什么改（2026-09-22 用户问「日报为什么带很多省略号」）：原实现把每条硬截到 24 字，
    于是 09-21 那份报告 6 处省略号，而且断在词中间 ——
      「证书自动续期：根因是域名未备案导致 HTTP-0…」（HTTP-01 被截断）
      「日报取数通道加固（重试窗口 2.5h + SKI…」（SKIP_FETCH 被截断）
      「Revert "feat…」
    现在改成「取主干」：破折号后面是解释、冒号后面是解释，括号已经在上游去掉，
    剩下还超长才截，且只在自然分隔符处截、**不切断英文/数字词**。
    结果不仅更短，而且装得下更多条（300 字预算里塞的是完整短语，不是带省略号的残句）。
    """
    original = s.strip()
    s = original
    # Revert "feat(ui): 大屏首页改版" → 回退 大屏首页改版（引号里的才是主体）
    rv = re.match(r'^[Rr]evert\s+["“「\'](.+?)["”」\']\s*$', s)
    if rv:
        s = '回退 ' + PREFIX.sub('', rv.group(1)).strip()
    # 破折号之后是解释
    s = s.split('——')[0]
    # 全角冒号之后是解释（**不用半角 ':'** —— 那会把 19:30 → 17:30 切成 19）
    s = s.split('：')[0]
    # 整串被引号包起来时，引号里才是主体
    s = re.sub(r'^["“「\'](.+?)["”」\']$', r'\1', s)
    s = s.strip().strip('，,。;； ')
    if not s:
        s = original
    if len(s) <= n:
        return s
    cut = s[:n]
    for sep in SEPS:
        i = cut.rfind(sep)
        if i >= max(6, n // 2):
            cut = cut[:i]
            break
    else:
        # 没有自然分隔符：退到词边界（别把 HTTP-01 切成 HTTP-0、SKIP_FETCH 切成 SKI）
        cut = cut.rstrip(WORDISH)
    cut = cut.rstrip('(/、， ；;+＋.,')
    return (cut or s[:n]) + '…'


# 一个提交里常常塞了好几件事（「堵住自助充值印钞口;支付端点改 4xx;补下单限频」），
# 先按 ; 拆成**工作项**再归类 —— 拆开后每条都短，300 字里能装下更多真实内容，
# 也不会出现「一句话被切到一半」的残句。
items = {m: [] for m in ORDER}
for c in commits:
    m = module_of(c)
    # 先去括号补充说明，再按 ; / + 拆工作项 —— 顺序不能反（见 SPLIT 的注释）
    subject = PARENS.sub('', c['subject'])
    for frag in SPLIT.split(PREFIX.sub('', subject).strip()):
        frag = frag.strip()
        if not frag:
            continue
        # 按 + 拆出来的纯 ASCII 短碎片（TLS-ALPN-01、SKIP_FETCH 这种）单独成条像噪音，丢掉。
        # 带中文的短条目（「补赛事名额的库级守护测试」）照留 —— 那是真正的工作项。
        if len(frag) <= 12 and not CJK.search(frag):
            continue
        items[m].append(headline(frag))

# ---------- 3. 规模、迁移、进度节点 ----------
files = sorted({f for c in commits for f in c['files'] if f})
migs = sorted({re.match(r'^migrations/(\d+)', f).group(1)
               for f in files if re.match(r'^migrations/\d+.*\.up\.sql$', f)})

t = time.strptime(day_arg, '%Y-%m-%d')
head = '【日报 %s 周%s】' % (time.strftime('%m-%d', t), '一二三四五六日'[t.tm_wday])
counts = '1. 提交 %d 个、%d 个文件' % (len(commits), len(files))
if migs:
    counts += '，%d 组数据库迁移' % len(migs)
counts += '。'
# 空串 = 不写进度行（补看历史日期时，部署节点没有意义）
node_text = deploy_note

# ---------- 4. 组装（预算均分，装不下整条丢弃，绝不切半句） ----------
present = [m for m in ORDER if items[m]]
# 标题 + 规模行 + 换行编号的固定开销（进度行有才计入）
fixed = len(head) + len(counts) + 14 + (len(node_text) + 8 if node_text else 0)


def build(mods_present):
    out, idx = [], 2
    for m in mods_present:
        prefix = '%d. %s：' % (idx, m)
        share = max(8, (BUDGET - fixed) // max(1, len(mods_present)) - len(prefix))
        picked, used, seen = [], 0, set()
        for it in items[m]:
            if it in seen:
                continue
            seen.add(it)
            cost = len(it) + (1 if picked else 0)
            if used + cost > share:
                continue
            picked.append(it)
            used += cost
        if picked:
            line = prefix + '；'.join(picked)
            out.append(line if line.endswith('…') else line + '。')
            idx += 1
    return out


def render(mods_lines):
    lines = [head, counts] + mods_lines
    if node_text:
        lines.append('%d. 进度：%s。' % (len(mods_lines) + 2, node_text))
    return '\n'.join(lines)


mods = build(present)
# 均分只是估算，取整后仍可能超；超了就砍掉最后一个模块重排（编号会跟着变）
while len(render(mods)) > BUDGET and len(present) > 1:
    present = present[:-1]
    mods = build(present)

print(render(mods))