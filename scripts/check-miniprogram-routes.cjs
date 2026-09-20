#!/usr/bin/env node
// 小程序路由自检：把「页面文件 ↔ pages.json ↔ 代码里的路由字符串」三者对一遍。
//
// 为什么需要（2026-09-20）：小程序**没有构建期路由校验**。删了页面忘了改 pages.json、
// 或者某份白名单里留着已删页面的路径，都要等用户点进去才发现是白屏。
// 当天就是这么发现首页横幅白名单 ALLOWED_ROUTES 里还留着 `/pkg-eco/pages/shops/index`
// —— shops 表早在迁移 000113 就被删了，目录都不存在。
//
// 检查两类「真问题」（任一即 exit 1）：
//   A. pages.json 注册了，但 .vue 文件不存在  → 构建/运行时报错
//   B. 代码里写死的路由，pages.json 没注册    → 点进去 404 / 白屏
// 另外附带打印「零引用页面」（可能是孤儿，仅提示，不影响退出码）。
//
// 用法：node scripts/check-miniprogram-routes.cjs（脚本必须放 scripts/：.tools/ 被 gitignore，CI 拿不到）
const fs = require('fs');
const path = require('path');

const ROOT = path.join(__dirname, '..');
const MP = path.join(ROOT, 'miniprogram');
const norm = (p) => p.replace(/\\/g, '/');

const pagesJson = JSON.parse(fs.readFileSync(path.join(MP, 'pages.json'), 'utf8'));
const registered = new Map();
for (const p of pagesJson.pages || []) registered.set(p.path, '主包');
for (const sp of pagesJson.subPackages || []) {
  for (const p of sp.pages || []) registered.set(sp.root + '/' + p.path, '分包 ' + sp.root);
}
const tabBar = new Set(((pagesJson.tabBar || {}).list || []).map((t) => t.pagePath));

function walk(dir, out = []) {
  for (const e of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) {
      if (e.name !== 'unpackage' && e.name !== 'node_modules') walk(p, out);
    } else if (/\.(vue|js|json|wxml|wxss)$/.test(e.name)) out.push(p);
  }
  return out;
}
const files = walk(MP).map((f) => ({ rel: norm(f).slice(norm(MP).length + 1), src: fs.readFileSync(f, 'utf8') }));

// ROUTE_MAP 这类「旧路径 → 新路径」的别名表：它的**键**本来就该是不存在的旧路径
// （/pages/demand/list → /pages/demands/list 就是这种），不能当死链报。
const aliasKeys = new Set();
for (const { src } of files) {
  for (const blk of src.matchAll(/const\s+\w*ROUTE_MAP\w*\s*=\s*\{([\s\S]*?)\n\}/g)) {
    for (const k of blk[1].matchAll(/['"\`](\/[^'"\`]+)['"\`]\s*:/g)) aliasKeys.add(k[1].replace(/^\/+/, ''));
  }
}

// ---- A. pages.json 指向的 .vue 是否存在 ----
const missingFiles = [];
for (const p of registered.keys()) {
  if (!fs.existsSync(path.join(MP, p + '.vue'))) missingFiles.push(p);
}

// ---- B. 代码里写死的路由是否都在 pages.json ----
// 只认「以 /pages/ 或 /pkg-xxx/ 开头」的字符串字面量；跳过接口路径、静态资源、含变量的模板串
const routeRe = /['"`](\/(?:pages|pkg-[a-z0-9-]+)\/[A-Za-z0-9_\/-]+)/g;

// 自检：正则捕获的字符串带前导 /，而 pages.json 的键不带。
// 第一版忘了剥前导斜杠，110 个页面全被判成死链（连 /pages/home/index 都说没有）——
// 有了这道自检，匹配器本身错了就直接中止，不会输出一屏假结论。
if (!registered.size) { console.log('pages.json 里没读到任何页面'); process.exit(2); }
const assetRe = /\.(svg|png|jpe?g|gif|webp|mp4|json|css|js|wxss)$/i;
const dead = new Map();
for (const { rel, src } of files) {
  if (rel === 'pages.json') continue;    // pages.json 自己就是注册表，不算引用
  src.split(/\r?\n/).forEach((line, i) => {
    if (line.trim().startsWith('//') || line.trim().startsWith('*')) return;
    let m;
    routeRe.lastIndex = 0;
    while ((m = routeRe.exec(line))) {
      const route = m[1].replace(/^\/+/, '').replace(/\/+$/, '');
      if (assetRe.test(route)) continue;
      if (aliasKeys.has(route)) continue;   // 旧路径别名，故意不存在
      if (registered.has(route)) continue;
      if (!dead.has(route)) dead.set(route, []);
      dead.get(route).push(rel + ':' + (i + 1));
    }
  });
}

// ---- C. 零引用（仅提示）----
const referenced = new Set();
for (const { rel, src } of files) {
  if (rel === 'pages.json') continue;
  for (const p of registered.keys()) {
    if (rel === p + '.vue') continue;
    if (src.includes('/' + p)) referenced.add(p);
  }
}
const orphan = [...registered.keys()].filter((p) => !tabBar.has(p) && !referenced.has(p));

console.log('pages.json 注册页面 ' + registered.size + ' 个（tabBar ' + tabBar.size + ' 个）');
console.log('扫描文件 ' + files.length + ' 个');
console.log('');

let bad = 0;
console.log('== A. pages.json 注册了但文件不存在 ==');
if (missingFiles.length) { bad += missingFiles.length; missingFiles.forEach((p) => console.log('  缺失  ' + p + '.vue')); }
else console.log('  （无）');
console.log('');

if (aliasKeys.size) console.log('（已忽略 ' + aliasKeys.size + ' 条旧路径别名：' + [...aliasKeys].join('、') + '）');
console.log('== B. 代码里写死的路由，pages.json 没注册（点进去就是白屏）==');
if (dead.size) {
  bad += dead.size;
  for (const [route, where] of dead) console.log('  死链  ' + route + '   <- ' + where.slice(0, 3).join(', '));
} else console.log('  （无）');
console.log('');

console.log('== C. 零引用页面（可能是孤儿，仅提示，不算失败）==');
if (orphan.length) orphan.forEach((p) => console.log('  ' + p + '   [' + registered.get(p) + ']'));
else console.log('  （无）');
console.log('');

if (bad) { console.log('结论：' + bad + ' 项必须修'); process.exit(1); }
console.log('结论：路由自检通过');