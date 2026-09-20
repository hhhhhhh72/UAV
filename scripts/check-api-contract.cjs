#!/usr/bin/env node
// 接口契约自检：把「客户端写死的接口路径」与「后端真正注册的路由」对一遍。
//
// 为什么要它（2026-09-20）：小程序与管理后台都是前端先行，接口路径写错、后端改名、
// 端点被删，都不会在构建期报错 —— 只有用户点下去才 404。这类问题人工 review 看不完
// （小程序 200+ 调用点、后台 90+），但机器几秒钟就能对完。
//
// 扫描范围：
//   miniprogram/  request({ url, method }) / uni.request({ url, method })
//   frontend/     axios.get|post|put|delete|patch(url) 与 axios({ url, method })
// 判定：方法 + 路径段数 + 每段匹配（后端 {id} 与客户端 ${...} 都当通配）。
//
// 用法：node scripts/check-api-contract.cjs
const fs = require('fs');
const path = require('path');

const ROOT = path.join(__dirname, '..');
const norm = (p) => p.replace(/\\/g, '/');

// ---------- 1. 后端注册的路由 ----------
const routes = [];
(function walk(dir) {
  for (const e of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) { walk(p); continue; }
    if (!e.name.endsWith('.go') || e.name.endsWith('_test.go')) continue;
    const lines = fs.readFileSync(p, 'utf8').split(/\r?\n/);
    lines.forEach((line, i) => {
      const m = line.match(/mux\.HandleFunc\(\s*"([A-Z]+)\s+(\S+?)"/);
      if (m) routes.push({ method: m[1], pattern: m[2], file: norm(p).slice(norm(ROOT).length + 1), line: i + 1 });
    });
  }
})(path.join(ROOT, 'internal', 'httpapi'));

// 匹配方式：**逐段比对**，而不是「把后端路由编译成正则去测客户端字符串」。
// 为什么（第一版踩过两次）：
//   ① 路径按 '/' 切分后开头必有一个空段（前导斜杠产生），把它当通配符会让每条正则都多要一段，
//      结果全部匹配失败 —— 自检当场拦住；
//   ② 客户端末段是变量时（`/admin/reviews/${id}/${action}`，实际取值只有 approve/reject），
//      把它当普通字符串去测后端字面量 approve 永远不相等 —— 逐段比对时只要**任一段是通配**就算匹配，
//      这类写法自然对上。
const isWild = (s) => s === '*' || /^\{.+\}$/.test(s);
const segEq = (a, b) => isWild(a) || isWild(b) || a === b;
const segsOf = (p) => p.replace(/\/+$/, '').split('/');
const matches = (method, clientPath) => {
  const cs = segsOf(clientPath);
  return routes.some((r) => {
    if (r.method !== method) return false;
    const rs = segsOf(r.pattern);
    return rs.length === cs.length && rs.every((s, i) => segEq(s, cs[i]));
  });
};

// ---------- 2. 匹配器自检（错了就别信下面的结论）----------
let selfFail = 0;
const selfTests = [
  ['路由表非空', routes.length > 400],
  ['GET  /api/v1/demands 命中', matches('GET', '/api/v1/demands')],
  ['不存在的路径不命中', !matches('GET', '/api/v1/definitely-not-a-real-route-xyz')],
  ['段数不同不命中', !matches('GET', '/api/v1/demands/extra/segments/here')],
  ['通配段命中', matches('GET', '/api/v1/demands/abc123')],
  ['POST 同路径命中', matches('POST', '/api/v1/demands')],
  ['方法不匹配则不命中', !matches('PUT', '/api/v1/demands')],
  ['多级通配命中', matches('GET', '/api/v1/demands/abc/intents')],
  // 客户端末段是变量、后端是枚举字面量：必须判为命中（frozens 的真实取值为 approve/reject）
  ['客户端通配对上后端字面量', matches('POST', '/api/v1/admin/reviews/*/*')],
  ['真实不存在的末段仍不命中', !matches('POST', '/api/v1/admin/reviews/*/definitely-not-an-action')],
];
for (const [name, ok] of selfTests) { if (!ok) { console.log('自检失败: ' + name); selfFail++; } }
console.log('匹配器自检: ' + (selfTests.length - selfFail) + '/' + selfTests.length + (selfFail ? '  ✗ 结论不可信' : '  ✓'));
if (selfFail) process.exit(2);
console.log('后端注册路由: ' + routes.length + ' 条');

// ---------- 3. 客户端调用点 ----------
function walkFiles(dir, exts, skip, out = []) {
  for (const e of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) { if (!skip.test(e.name)) walkFiles(p, exts, skip, out); }
    else if (exts.some((x) => e.name.endsWith(x))) out.push(p);
  }
  return out;
}

const calls = [];

// 从一段对象文本里取出 url 值。
// **必须取整个表达式**：`'/api/v1/demands/' + encodeURIComponent(id) + '/intents'`
// 只抓第一个字面量会得到 `/api/v1/demands/`，段数对不上后端而误报死链 ——
// 第一版就是这么一次性报出 24 条假死链的。
// 做法：取到 `,` 或换行为止（值表达式），把里面**所有**字符串字面量按顺序拼起来，
// 字面量之间补一个通配段。
function joinLits(expr) {
  const lits = [...expr.matchAll(/['"\`]([^'"\`]*)['"\`]/g)].map((x) => x[1]).filter((s) => s);
  if (!lits.length) return '';
  return lits.reduce((acc, s, i) => {
    if (i === 0) return s;
    const a = acc.endsWith('/') ? acc : acc + '/';
    const b = s.startsWith('/') ? s.slice(1) : s;
    return a + '*' + (b ? '/' + b : '');
  }, '');
}

// 拆一层三元：`cond ? A : B` → [A, B]。
// 为什么必须拆：真实写法是
//   url: editing ? '/api/v1/products/' + id : '/api/v1/products'
//   method: editing ? 'PATCH' : 'POST'
// 这是**两个不同端点**。不拆就会把两个字面量拼成
// `/api/v1/products/*/api/v1/products` 这种根本不存在的路径，误报成死链。
// 只处理一层、且按第一个 ':' 切分 —— 覆盖实际写法即可，注释在此说明局限。
function branchesOf(expr) {
  const q = expr.indexOf('?');
  if (q < 0) return null;
  const colon = expr.indexOf(':', q);
  if (colon < 0) return null;
  return [expr.slice(q + 1, colon), expr.slice(colon + 1)];
}

// 返回 [{ method, path }] —— 三元各分支各算一次调用
function extractCalls(body) {
  const uv = body.match(/url\s*:\s*([^,\n]+)/);
  if (!uv) return [];
  const ub = branchesOf(uv[1]);
  const paths = (ub || [uv[1]]).map(joinLits).filter(Boolean);
  if (!paths.length) return [];

  const mv = body.match(/method\s*:\s*([^,\n]+)/);
  let methods = ['GET'];
  if (mv) {
    const mb = branchesOf(mv[1]);
    methods = (mb || [mv[1]])
      .map((e) => (e.match(/['"\`]([A-Za-z]+)['"\`]/) || [])[1])
      .filter(Boolean);
    if (!methods.length) methods = ['GET'];
  }

  const out = [];
  if (paths.length === methods.length) {
    paths.forEach((p, i) => out.push({ method: methods[i], path: p }));
  } else {
    for (const p of paths) for (const m of methods) out.push({ method: m, path: p });
  }
  return out;
}
function addCall(file, line, method, rawPath) {
  let p = rawPath.split('?')[0].replace(/\$\{[^}]*\}/g, '*');
  // 拼接写法的字面量部分：'/api/v1/x/' + id 与 '/api/v1/x' + suffix 都按通配收尾
  if (p.endsWith('/')) p += '*';
  p = p.replace(/\/+$/, '');
  const segs = p.split('/');
  if (segs[segs.length - 1] === '') segs[segs.length - 1] = '*';
  p = segs.join('/');
  if (!/^\/api\//.test(p)) return;
  calls.push({ file, line, method: (method || 'GET').toUpperCase(), path: p });
}

// 小程序：request({...}) / uni.request({...})，对象可能跨多行，取到第一个 }) 为止
for (const f of walkFiles(path.join(ROOT, 'miniprogram'), ['.vue', '.js'], /^(unpackage|node_modules)$/)) {
  const src = fs.readFileSync(f, 'utf8');
  const rel = norm(f).slice(norm(ROOT).length + 1);
  for (const head of src.matchAll(/(?:\brequest|\buni\.request)\(\s*\{/g)) {
    const start = head.index + head[0].length;
    let end = src.indexOf('})', start);
    if (end < 0) end = Math.min(start + 600, src.length);
    const body = src.slice(start, end);
    const ln = src.slice(0, head.index).split('\n').length;
    for (const c of extractCalls(body)) addCall(rel, ln, c.method, c.path);
  }
}

// 管理后台：axios.get|post|...('...') 与 axios({ url, method })
for (const f of walkFiles(path.join(ROOT, 'frontend', 'src'), ['.js', '.vue'], /^(node_modules|dist)$/)) {
  const src = fs.readFileSync(f, 'utf8');
  const rel = norm(f).slice(norm(ROOT).length + 1);
  for (const m of src.matchAll(/axios\.(get|post|put|delete|patch)\(\s*['"\`]([^'"\`]*)/g)) {
    addCall(rel, src.slice(0, m.index).split('\n').length, m[1], m[2]);
  }
  for (const head of src.matchAll(/axios\(\s*\{/g)) {
    const start = head.index + head[0].length;
    let end = src.indexOf('})', start);
    if (end < 0) end = Math.min(start + 600, src.length);
    const body = src.slice(start, end);
    const ln = src.slice(0, head.index).split('\n').length;
    for (const c of extractCalls(body)) addCall(rel, ln, c.method, c.path);
  }
}

console.log('客户端调用点: ' + calls.length + ' 处（小程序 + 管理后台）');
console.log('');

const bad = calls.filter((c) => !matches(c.method, c.path));
const seen = new Set();
console.log('== 调了但后端没有对应路由（点下去 404）==');
if (!bad.length) console.log('  （无）');
for (const b of bad) {
  const k = b.method + ' ' + b.path;
  if (seen.has(k)) continue;
  seen.add(k);
  console.log('  ' + k + '   <- ' + b.file + ':' + b.line);
}
console.log('');
if (bad.length) { console.log('结论：' + seen.size + ' 个接口对不上，必须修'); process.exit(1); }
console.log('结论：接口契约自检通过');