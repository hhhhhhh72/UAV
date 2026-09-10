#!/usr/bin/env node
/**
 * 路由权限矩阵扫描（只读审计）
 * 用途：枚举 internal/httpapi 全部 mux.HandleFunc 路由 → 定位 handler →
 *       检查 handler 前 30 行内是否存在鉴权标记（authenticatedActor/requireAdmin/角色判断）。
 * 输出：写操作(非GET)无鉴权、敏感GET无鉴权的疑似漏网路由。
 */
const fs = require('fs')
const path = require('path')

const DIR = path.join(__dirname, '..', 'internal', 'httpapi')
const files = fs.readdirSync(DIR).filter((f) => f.endsWith('.go') && !f.endsWith('_test.go'))

// 1) 汇聚全部 Go 源码文本（按文件）
const sources = {}
for (const f of files) {
  sources[f] = fs.readFileSync(path.join(DIR, f), 'utf8')
}
const all = Object.values(sources).join('\n')

// 2) 提取路由（兼容换行写法：HandleFunc("METHOD path", ...)）
const routeRe = /HandleFunc\(\s*"(GET|POST|PUT|PATCH|DELETE|HEAD)\s+([^"]+)",\s*s\.([A-Za-z0-9_]+)/g
const routes = []
let m
while ((m = routeRe.exec(all)) !== null) {
  routes.push({ method: m[1], path: m[2], handler: m[3] })
}

// 3) 提取每个 handler 的源码（定位 func 后取 30 行）
function handlerBody(name) {
  const re = new RegExp('func \\(s \\*Server\\) ' + name.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') + '\\(w http\\.ResponseWriter, r \\*http\\.Request\\) \\{[\\s\\S]*?\\n\\}')
  for (const text of Object.values(sources)) {
    const mm = text.match(re)
    if (mm) return mm[0]
  }
  return null
}

const AUTH_MARKERS = ['authenticatedActor(', 'requireAdmin(', 'adminGate', 'RolePlatformAdmin', 'RoleAssociationAdmin']
function hasAuth(body) {
  if (!body) return null // 找不到处理函数
  const head = body.split('\n').slice(0, 30).join('\n')
  return AUTH_MARKERS.some((k) => head.includes(k))
}

// 4) 分类输出
const issues = []
const unique = new Map() // 去重（同名路由可能注册两次）
for (const r of routes) {
  const key = r.method + ' ' + r.path
  if (unique.has(key)) continue
  unique.set(key, r)
  const isAdmin = r.path.startsWith('/api/v1/admin/') || r.path.startsWith('/api/admin/')
  if (isAdmin) continue // 中间件 adminGate 覆盖（已单独核验）
  const auth = hasAuth(r.handler)
  if (auth === false) {
    issues.push({ ...r, auth })
  }
}

console.log('总路由数:', unique.size, ' 非admin路由:', [...unique.values()].filter(r => !r.path.startsWith('/api/v1/admin/') && !r.path.startsWith('/api/admin/')).length)
console.log('')
console.log('=== ⚠ 疑似无鉴权（handler 前30行无鉴权标记） ===')
for (const i of issues) {
  const flag = i.method !== 'GET' ? 'WRITE' : (i.path.includes('mine') || i.path.includes('/me') || i.path.includes('private') ? 'SENSITIVE-GET' : 'GET')
  console.log(`${flag.padEnd(14)} ${i.method.padEnd(6)} ${i.path.padEnd(52)} -> ${i.handler}`)
}
console.log('')
console.log('=== handler 未定位（可能不在 httpapi 或命名不同，需人工看） ===')
for (const r of unique.values()) {
  if (r.path.startsWith('/api/v1/admin/') || r.path.startsWith('/api/admin/')) continue
  if (hasAuth(r.handler) === null) console.log(`${r.method.padEnd(6)} ${r.path.padEnd(52)} -> ${r.handler}`)
}
