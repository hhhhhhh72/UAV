#!/usr/bin/env node
/**
 * 订单状态标签的**角色**回归（2026-09-22）。
 *
 * 背景：小程序把后端状态 completed 展示成「待评价」，但这个说法只对**买家**成立 ——
 * 评价是买家对卖家的行为，卖家根本没有评价入口。第一版不问角色，于是**发布方（卖家）**
 * 看着自己卖出去、买家早已评过的单，卡片写着「待评价」、按钮写着「已完成」，
 * 而且**永远出不去**「待评价」列表（reviewed 算的是"当前用户有没有评价过"，卖家恒为 false）。
 * 用户连报两次的就是这个（第一次我按买家侧修，没打中）。
 *
 * 做法：把**真正的** miniprogram/utils/orderAdapter.js 复制成临时 .mjs，把它的
 * request / uni 换成桩，喂三条构造订单，跑真实的 loadOrders / loadStatusCounts 后断言。
 * **不复刻一份判断逻辑**：复刻的话源码改回去了它照样绿，等于没守。
 */
const fs = require('fs')
const os = require('os')
const path = require('path')

const ROOT = path.join(__dirname, '..')
const SRC = path.join(ROOT, 'miniprogram', 'utils', 'orderAdapter.js')

const ME = 'user-me'
const OTHER = 'user-other'

// 三条订单，覆盖三种关键情形（字段形状取自真实接口）
const ORDERS = [
  // ① 我**卖**出去的，已完成，买家已经评过 → 对我（卖家）来说就是「已完成」
  { id: 'torder-sold', product_id: 'p1', buyer_id: OTHER, seller_id: ME, amount_fen: 100, status: 'completed', reviewed: false, aftersale_status: '', created_at: '2026-09-22T14:58:42+08:00' },
  // ② 我**买**的，已完成，还没评价 → 这才是我要处理的「待评价」
  { id: 'torder-bought', product_id: 'p1', buyer_id: ME, seller_id: OTHER, amount_fen: 100, status: 'completed', reviewed: false, aftersale_status: '', created_at: '2026-09-22T15:58:42+08:00' },
  // ③ 我买的，已完成，服务端说我已评价 → 不该再出现在「待评价」
  { id: 'torder-reviewed', product_id: 'p1', buyer_id: ME, seller_id: OTHER, amount_fen: 100, status: 'completed', reviewed: true, aftersale_status: '', created_at: '2026-09-22T16:58:42+08:00' },
]

function buildTmpModule () {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'order-role-'))
  const stub = [
    'export const BASE_URL = "https://example.invalid"',
    'export function getStoredUser () { return { id: ' + JSON.stringify(ME) + ' } }',
    'export async function request (opts = {}) {',
    '  if (String(opts.url || "").includes("/trade-orders/mine")) return { data: JSON.parse(' + JSON.stringify(JSON.stringify(ORDERS)) + ') }',
    '  return { data: [] }',
    '}',
  ].join('\n')
  fs.writeFileSync(path.join(dir, 'request.mjs'), stub)
  let src = fs.readFileSync(SRC, 'utf8')
    .replace("from './request'", "from './request.mjs'")
    .replace(/import\.meta\.env\.DEV/g, 'false')
  fs.writeFileSync(path.join(dir, 'orderAdapter.mjs'), src)
  return dir
}

async function main () {
  const dir = buildTmpModule()
  // uni 全局桩：本地存储**故意留空**，模拟换设备/清缓存 —— 顺带证明这个判定不依赖本地存储
  globalThis.uni = { getStorageSync: () => undefined, setStorageSync: () => {} }
  const mod = await import(require('url').pathToFileURL(path.join(dir, 'orderAdapter.mjs')).href)

  const list = await mod.loadOrders({ status: 'completed' })
  const counts = await mod.loadStatusCounts('all')
  const all = await mod.loadOrders({ status: 'all' })
  const label = (id) => {
    const o = all.find((x) => x.id === id)
    return o ? (o.status_text || mod.ORDER_STATUS[o.status]) : '(缺单)'
  }

  const problems = []
  const has = (id) => list.some((o) => o.id === id)

  // ① 卖家侧：标签必须是「已完成」，且不进「待评价」
  if (label('torder-sold') !== '已完成') {
    problems.push(`卖家卖出的 completed 单标签应为「已完成」，实际「${label('torder-sold')}」——卖家没有评价入口，"待评价"对他不成立`)
  }
  if (has('torder-sold')) {
    problems.push('卖家卖出的 completed 单出现在了「待评价」列表里 —— 它对卖家永远出不去（reviewed 恒为 false）')
  }
  // ② 买家侧未评价：必须在「待评价」里，标签「待评价」
  if (!has('torder-bought') || label('torder-bought') !== '待评价') {
    problems.push(`买家买来且未评价的 completed 单应显示「待评价」并进列表，实际标签「${label('torder-bought')}」、在列表=${has('torder-bought')}`)
  }
  // ③ 买家侧已评价：不该再进「待评价」
  if (has('torder-reviewed')) {
    problems.push('服务端已标记 reviewed 的订单仍出现在「待评价」列表里 —— 换了设备/清了缓存就退回「待评价」')
  }
  // ④ 角标只数买家侧未评价的那一条
  if (counts.completed !== 1) {
    problems.push(`「待评价」角标应为 1（只有买家侧未评价那条），实际 ${counts.completed}`)
  }

  fs.rmSync(dir, { recursive: true, force: true })

  if (problems.length) {
    console.error('✗ 订单状态的角色判定有问题：')
    problems.forEach((p) => console.error('   - ' + p))
    console.error('   （数据：' + all.map((o) => `${o.id} role=${o.role} status=${o.status} 标签=${o.status_text || mod.ORDER_STATUS[o.status]}`).join(' | ') + '）')
    process.exit(1)
  }
  console.log('✓ 订单状态按角色判定正确（卖家「已完成」/ 买家未评价「待评价」/ 已评价不再进列表，角标 1）')
}

main().catch((e) => { console.error('检查器自身出错：', e); process.exit(1) })
