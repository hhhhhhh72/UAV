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

# ---------- 1. 取提交（连同每个提交改了哪些文件） ----------
# 时间窗 = **上一班的发放时刻 → 这一班的发放时刻**（滚动 24 小时），不是自然日。
#
# 为什么改（2026-09-23 用户：「内容好少，把昨天干的事也编进来」）：
#   自然日窗口在 17:30 收口，于是 17:30 之后干的活**当班看不到、下一班也不管**
#   （下一班只看新的自然日）—— 等于每天傍晚之后的工作从来没进过日报。
#   改成滚动窗口后，每一段工作都恰好被报一次：不重复、也不漏，内容自然变多。
REPORT_AT = '17:30:00'   # 与 crontab 里「30 17 * * *」对齐


def window(day_arg):
    t = time.mktime(time.strptime(day_arg, '%Y-%m-%d'))
    prev = time.strftime('%Y-%m-%d', time.localtime(t - 86400))
    start = '%s %s' % (prev, REPORT_AT)
    if day_arg == time.strftime('%Y-%m-%d'):
        # 当班**不设上界**：git 默认就取到 HEAD。
        # 不给 --until 是故意的 —— 早先写成「取到此刻（截到秒）」，而 time.strftime 会把
        # 小数秒抹掉：若某条提交正好发生在同一秒的后半段，它就被判成"晚于上界"而漏掉
        #（变异测试时真踩到了：临时仓库里 3 条刚建的提交一条都没被算进来）。
        end = None
    else:
        end = '%s %s' % (day_arg, REPORT_AT)       # 补看历史：取到那一班的发放时刻
    return start, end


since, until = window(day_arg)
win_label = since[5:16]   # 例：09-22 17:30

git_args = ['git', '-C', repo, 'log', '--since=' + since]
if until:                      # 当班不给 --until（见 window() 里的说明）
    git_args.append('--until=' + until)
# 用 HEAD 而不是写死 master：CI 里 git init 默认分支可能是 main，
# 写死分支名会让检查器在 CI 上静默查到 0 条提交（本地却正常）。
git_args += ['--pretty=format:@@%h|%s', '--name-only', '--no-merges', 'HEAD']
raw = subprocess.run(git_args, capture_output=True, text=True, errors='replace').stdout.splitlines()

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
counts = '1. 提交 %d 个、%d 个文件（%s 起）' % (len(commits), len(files), win_label)
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
    """组装正文：**把 300 字用满**，而不是平均分给每个模块。

    为什么改（2026-09-23 用户：「内容好少」）：原实现把预算**平均**分给当天出现过的每个模块
    （通常 6–8 个），每格只剩 30 来字、只装得下 1–2 条；而有的模块本来就没几条，
    那份额度就白白空着 —— 实测 09-22 那份只用了 161/300 字、09-21 用了 167/300。
    现在改成**轮转分配**：每轮给每个模块各加一条，谁还有货谁继续拿，直到预算用满或都没货。
    轮转也保住了原来的优先级语义：预算不够时，靠后的模块连第一条都拿不到。
    """
    if not mods_present:
        return []
    head_cost = sum(len('%d. %s：' % (i, m)) for i, m in enumerate(mods_present, 2))
    free = BUDGET - fixed - head_cost
    pool, picked, used = {}, {m: [] for m in mods_present}, 0
    for m in mods_present:
        seen, keep = set(), []
        for it in items[m]:
            if it not in seen:
                seen.add(it)
                keep.append(it)
        # 「X」和「回退 X」同时在时丢掉前者：净效果就是撤回了，两条一起报纯属白占额度
        #（2026-09-20 那天就是这样：「供给大厅加卡片按压反馈」+「回退 供给大厅加卡片按压反馈」）
        revs = {it[3:] for it in keep if it.startswith('回退 ')}
        if revs:
            keep = [it for it in keep if it not in revs]
        pool[m] = keep
    while True:
        added = False
        for m in mods_present:
            while pool[m]:
                cost = len(pool[m][0]) + (1 if picked[m] else 0)   # 分隔符「；」
                if used + cost <= free:
                    picked[m].append(pool[m].pop(0))
                    used += cost
                    added = True
                    break
                # 这条装不下：预算只会越来越紧，留着也不会再被选中，直接丢
                pool[m].pop(0)
        if not added:
            break
    out = []
    for m in mods_present:
        if not picked[m]:
            continue
        line = '%d. %s：' % (len(out) + 2, m) + '；'.join(picked[m])
        out.append(line if line.endswith('…') else line + '。')
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