#!/usr/bin/env node
// 小程序演示数据守卫自检：凡是用了假数据的地方，必须带生产环境守卫。
//
// 为什么需要（2026-09-20 契约体检发现）：
//   miniprogram/utils 下的 6 个 mock 模块（mockProjects / mockChallenges / mockBrands /
//   mockAchievements / mockReports / mockExhibitions）**自身没有任何守卫**，导出的是完整真实数组；
//   生产是否安全完全取决于调用点。逐点核对后查出 3 个页面漏了守卫：
//     pkg-eco/pages/projects/detail.vue   整个文件零守卫，且失败分支没置 err → 生产上是空白页
//     pkg-eco/pages/projects/list.vue     接口一挂就把编造的课题攻关列表当真实数据展示
//     pkg-eco/pages/challenges/list.vue   同上（研发难题）
//   而隔壁 pkg-eco/pages/challenges/detail.vue:137 明明写着「数字诚实铁律」并做了守卫，
//   说明这三处是漏改而非有意为之。列表页更隐蔽：横幅是 v-if="mockMode && isDev"，
//   生产构建里 mockMode 是个死标志 —— 假数据结构照渲染，却没有任何「演示数据」提示。
//
// 判定规则（任一命中该文件即需带守卫）：
//   ① 从 utils/mock* 导入（MOCK_/SEED_ 数组与标签常量同住一个模块，已核对 14 个导入点全部真的取了假数组）
//   ② 文件自己 export const MOCK_* / SEED_* 定义假数据（如 utils/hallData.js）
// 豁免：utils/mock*.js 数据模块本身。
//
// 一个「假数据引用行」算安全，须在它自身或 ±25 行内出现**已应用**的守卫：
//   守卫写法：process.env.NODE_ENV === 'production' / import.meta.env.DEV / isDev / isProduction
//   已应用：所在行是 if(...) / && / ?:（即真的拿它做了判断）
// 两条不算数：
//   · 声明行不算已应用 —— `const isDev = process.env.NODE_ENV === 'development'` 只说明文件里有这个标志，
//     不说明它罩住了这一行；把守卫从 if 里摘掉后，检查器必须变红（已用变异验证）。
//   · 注释里的守卫不算 —— 会先剥注释，否则 utils/config.js 里那句提到 NODE_ENV 的说明文字就能让任何文件蒙混过关。
// import 行与 export const MOCK_* 定义行不算「使用」（导入本身不是渲染假数据）。
//
// 这是**页面级绊线**，不是证明。已实测的能力边界：
//   ✅ 抓得住：新增零守卫的回退；守卫只写在注释里；只在文件顶部声明却没应用（const isDev = ... 之后直接用假数据）。
//   ❌ 抓不住：同一函数里有两个相邻回退，只摘掉其中一个的守卫 —— 另一个的守卫落在 ±25 行窗口内会掩盖它
//      （projects/detail.vue 实测确认）。这一类只能靠「守卫紧贴使用处」的人工约定，检查器给不了保证。
//   为什么不按缩进/括号配对做块级判定：仓库里列表页的主导写法是「守卫在调用点（深层缩进）+ 假数据在
//   顶层 useMock() 助手函数里」，按缩进判定会把 portfolios/list、reports/list、exhibitions/list 三处
//   正确的守卫全判成违规（已实测）。宁可漏报一类，不可误报一片 —— 误报的检查器会被直接绕过。
//
// 用法：node scripts/check-mock-guards.cjs        （加 --selftest 跑判定逻辑自检）
const fs = require('fs');
const path = require('path');

const ROOT = path.join(__dirname, '..');
const MP = path.join(ROOT, 'miniprogram');
const WINDOW = 25;
const GUARD_RE = /process\.env\.NODE_ENV\s*===\s*['"](?:production|development)['"]|import\.meta\.env\.DEV|!\s*isProduction|\b(isDev|isProduction)\b/;
const APPLIED_RE = /if\s*\(|&&|\?/;
const DECL_RE = /^\s*(?:const|let|var)\s+(?:isDev|isProduction)\s*=/;
const MOCK_USE_RE = /\b(MOCK|SEED)_[A-Z0-9_]+\b/;
const MOCK_DEF_RE = /^\s*export\s+const\s+(MOCK|SEED)_[A-Z0-9_]+/;

// 剥注释但保行号（块注释折成等量空行，否则行号整体前移、报错指错地方）
function stripComments(src) {
  return src
    .replace(/\/\*[\s\S]*?\*\//g, (m) => m.replace(/[^\n]/g, ' '))
    .replace(/<!--[\s\S]*?-->/g, (m) => m.replace(/[^\n]/g, ' '))
    .split('\n')
    .map((l) => l.replace(/(^|\s)\/\/.*$/, '$1'))
    .join('\n');
}

const applied = (line) => !!line && !DECL_RE.test(line) && GUARD_RE.test(line) && APPLIED_RE.test(line);

function analyze(relPath, text) {
  const p = relPath.replace(/\\/g, '/');
  if (/^miniprogram\/utils\/mock[A-Za-z0-9]*\.js$/.test(p)) return { exempt: true };
  const stripped = stripComments(text);
  const lines = stripped.split('\n');
  const importsMock = /from\s+['"][^'"]*utils\/mock[^'"]*['"]/.test(stripped);
  const definesMock = lines.some((l) => MOCK_DEF_RE.test(l));
  if (!importsMock && !definesMock) return { exempt: false, inScope: false };
  const src = text.split('\n');
  const violations = [];
  for (let i = 0; i < lines.length; i++) {
    const l = lines[i];
    if (MOCK_DEF_RE.test(l) || /^\s*import\b/.test(l)) continue; // 定义/导入不算使用
    if (!MOCK_USE_RE.test(l)) continue;
    let ok = applied(l);
    for (let k = Math.max(0, i - WINDOW); !ok && k <= Math.min(lines.length - 1, i + WINDOW); k++) {
      if (k !== i && applied(lines[k])) ok = true;
    }
    if (!ok) violations.push({ line: i + 1, text: (src[i] || '').trim() });
  }
  return { exempt: false, inScope: true, via: importsMock ? 'import' : 'define', violations };
}

function walk(dir, out = []) {
  for (const e of fs.readdirSync(dir, { withFileTypes: true })) {
    if (['node_modules', 'unpackage', 'dist', '.git', '.hbuilderx'].includes(e.name)) continue;
    const full = path.join(dir, e.name);
    if (e.isDirectory()) walk(full, out);
    else if (/\.(vue|js)$/.test(e.name)) out.push(full);
  }
  return out;
}

function selftest() {
  const IMP = "import { MOCK_PROJECTS } from '@/utils/mockProjects'\n";
  const cases = [
    ['已应用的守卫（if + isDev）', 'pkg-eco/pages/x/a.vue', IMP + "const isDev = process.env.NODE_ENV === 'development'\nif (isDev && MOCK_PROJECTS.length) { d = MOCK_PROJECTS }\n", 0],
    ['内联守卫（isDev &&）', 'pkg-eco/pages/x/a2.vue', IMP + 'if (isDev && MOCK_PROJECTS.length) { d = MOCK_PROJECTS }\n', 0],
    ['零守卫', 'pkg-eco/pages/x/b.vue', IMP + 'if (MOCK_PROJECTS.length) { d = MOCK_PROJECTS }\n', 1],
    ['只在文件顶部声明、未应用', 'pkg-eco/pages/x/b2.vue', IMP + "const isDev = process.env.NODE_ENV === 'development'\nif (MOCK_PROJECTS.length) { d = MOCK_PROJECTS }\n", 1],
    ['守卫只写在注释里', 'pkg-eco/pages/x/c.vue', IMP + "// 生产：process.env.NODE_ENV === 'production' 时绝不回退\nif (MOCK_PROJECTS.length) { d = MOCK_PROJECTS }\n", 1],
    ['守卫只在 25 行外', 'pkg-eco/pages/x/d.vue', IMP + 'if (isDev) { ok() }\n' + '\n'.repeat(30) + 'if (MOCK_PROJECTS.length) { d = MOCK_PROJECTS }\n', 1],
    ['自带假数据 + import.meta.env.DEV', 'miniprogram/utils/hall.js', 'export const MOCK_DEMANDS = [1]\nexport function get() {\n  if (!import.meta.env.DEV) return []\n  return MOCK_DEMANDS\n}\n', 0],
    ['自带假数据 + 无守卫', 'miniprogram/utils/hall2.js', 'export const MOCK_DEMANDS = [1]\nexport function get() { return MOCK_DEMANDS }\n', 1],
    ['数据模块自身豁免', 'miniprogram/utils/mockProjects.js', 'export const MOCK_PROJECTS = [1]\n', 0],
    ['无关文件（FORCE_MOCK_REPORTS）', 'miniprogram/utils/config.js', '// 演示数据开关（NODE_ENV）\nexport const FORCE_MOCK_REPORTS = false\n', 0],
    ['只导入不使用 → 不算违规', 'pkg-eco/pages/x/f.vue', IMP + 'const a = 1\n', 0],
    ['URL 里的 // 不算注释', 'pkg-eco/pages/x/e.vue', "import { MOCK_REPORTS } from '@/utils/mockReports'\nconst u = 'https://a.cn/x'\nif (import.meta.env.DEV) { d = MOCK_REPORTS }\n", 0],
    ['块注释跨行不打乱行号', 'pkg-eco/pages/x/g.vue', IMP + '/* 一段\n   多行注释\n   说明 */\nif (MOCK_PROJECTS.length) { d = MOCK_PROJECTS }\n', 1],
  ];
  let bad = 0;
  for (const [name, p, text, want] of cases) {
    const r = analyze(p, text);
    const got = r.exempt || !r.inScope ? 0 : r.violations.length;
    const pass = got === want;
    if (!pass) bad++;
    console.log(`  ${pass ? 'PASS' : 'FAIL'}  ${name}（期望 ${want}，实得 ${got}）`);
  }
  console.log(bad === 0 ? `\n自检 ${cases.length}/${cases.length} 通过` : `\n自检失败 ${bad}/${cases.length}`);
  process.exit(bad === 0 ? 0 : 1);
}

if (process.argv.includes('--selftest')) selftest();

const files = walk(MP);
const offenders = [];
let scoped = 0;
for (const f of files) {
  const rel = path.relative(ROOT, f).replace(/\\/g, '/');
  const r = analyze(rel, fs.readFileSync(f, 'utf8'));
  if (r.exempt || !r.inScope) continue;
  scoped++;
  if (r.violations.length) offenders.push({ rel, via: r.via, violations: r.violations });
}

console.log(`扫描 ${files.length} 个文件，其中 ${scoped} 个引用了演示数据（守卫窗口 ±${WINDOW} 行，须为已应用的判断）`);
if (offenders.length === 0) {
  console.log('✓ 全部带生产环境守卫');
  process.exit(0);
}
console.log(`\n✗ ${offenders.length} 个文件用了演示数据却没有守卫：`);
for (const o of offenders) {
  console.log(`  ${o.rel}（${o.via}）`);
  for (const v of o.violations) console.log(`    ${o.rel}:${v.line}  ${v.text.slice(0, 88)}`);
}
process.exit(1);
