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
 * 修法分两层，这个检查器两层都守：
 *   · 卡片标签按角色：买家未评价「待评价」/ 买家已评价「已评价」/ 卖家「已完成」
 *   · 第 4 个入口改名「已完成」并装**我所有的已完成单**（我买的 + 我卖的）——
 *     否则卖家卖出的单既显示错、又哪儿都不在（用户紧接着就问"那我已完成的订单放哪"）
 *   · 角标仍只数"还需要我评价"的（买家侧未评价）
 *
 * 做法：把**真正的** miniprogram/utils/orderAdapter.js 复制成临时 .mjs，把它的
 * request / uni 换成桩，喂四条构造订单，跑真实的 loadOrders / loadStatusCounts 后断言。
 * **不复刻一份判断逻辑**：复刻的话源码改回去了它照样绿，等于没守。
 */
const fs = require('fs')
const os = require('os')
const path = require('path')

const ROOT = path.join(__dirname, '..')
const SRC = path.join(ROOT, 'miniprogram', 'utils', 'orderAdapter.js')

const ME = 'user-me'
const OTHER = 'user-other'

// 四条订单，覆盖四种关键情形（字段形状取自真实接口）
const ORDERS = [
  // ① 我**卖**出去的，已完成，买家已经评过 → 对我（卖家）来说就是「已完成」
  { id: 'torder-sold', product_id: 'p1', buyer_id: OTHER, seller_id: ME, amount_fen: 100, status: 'completed', reviewed: false, aftersale_status: '', created_at: '2026-09-22T14:58:42+08:00' },
  // ② 我**买**的，已完成，还没评价 → 卡片仍是「待评价」
  { id: 'torder-bought', product_id: 'p1', buyer_id: ME, seller_id: OTHER, amount_fen: 100, status: 'completed', reviewed: false, aftersale_status: '', created_at: '2026-09-22T15:58:42+08:00' },
  // ③ 我买的，已完成，服务端说我已评价 → 卡片「已评价」，但仍算"我完成的单"
  { id: 'torder-reviewed', product_id: 'p1', buyer_id: ME, seller_id: OTHER, amount_fen: 100, status: 'completed', reviewed: true, aftersale_status: '', created_at: '2026-09-22T16:58:42+08:00' },
  // ④ 我买的，售后已结案（status 回 completed）→ 只该进「退款/售后」，不能混进「已完成」
  { id: 'torder-aftersale', product_id: 'p1', buyer_id: ME, seller_id: OTHER, amount_fen: 100, status: 'completed', reviewed: false, aftersale_status: 'approved', aftersale_type: 'refund', aftersale_amount_fen: 100, aftersale_time: '2026-09-22T17:00:00+08:00', created_at: '2026-09-22T17:58:42+08:00' },
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
  const src = fs.readFileSync(SRC, 'utf8')
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

  const done = await mod.loadOrders({ status: 'completed' })
  const aftersale = await mod.loadOrders({ status: 'aftersale' })
  const counts = await mod.loadStatusCounts('all')
  const all = await mod.loadOrders({ status: 'all' })
  const inDone = (id) => done.some((o) => o.id === id)
  const label = (id) => {
    const o = all.find((x) => x.id === id)
    return o ? (o.status_text || mod.ORDER_STATUS[o.status]) : '(缺单)'
  }

  const problems = []

  // ① 卖家侧：卡片「已完成」，且**必须在「已完成」列表里**
  if (label('torder-sold') !== '已完成') {
    problems.push(`卖家卖出的 completed 单卡片标签应为「已完成」，实际「${label('torder-sold')}」——卖家没有评价入口，"待评价"对他不成立`)
  }
  if (!inDone('torder-sold')) {
    problems.push('卖家卖出的 completed 单没进「已完成」列表 —— 用户会找不到自己成交的单')
  }
  // ② 买家侧未评价：卡片「待评价」，并进列表
  if (label('torder-bought') !== '待评价' || !inDone('torder-bought')) {
    problems.push(`买家买来、还没评价的 completed 单应显示「待评价」并在「已完成」列表里，实际标签「${label('torder-bought')}」、在列表=${inDone('torder-bought')}`)
  }
  // ③ 买家侧已评价：卡片「已评价」，但**仍然**算"我完成的单"
  if (label('torder-reviewed') !== '已评价' || !inDone('torder-reviewed')) {
    problems.push(`已评价的 completed 单应显示「已评价」并留在「已完成」列表里，实际标签「${label('torder-reviewed')}」、在列表=${inDone('torder-reviewed')}`)
  }
  // ④ 售后结案单（status 也是 completed）只该进「退款/售后」
  if (inDone('torder-aftersale')) {
    problems.push('售后结案单混进了「已完成」列表 —— 它应归「退款/售后」')
  }
  if (!aftersale.some((o) => o.id === 'torder-aftersale')) {
    problems.push('售后结案单没进「退款/售后」列表')
  }
  // ⑤ 角标只数"还需要我评价"的那一条（提醒型，不是列表条数）
  if (counts.completed !== 1) {
    problems.push(`「已完成」角标应为 1（只数买家侧还没评价的），实际 ${counts.completed} —— 角标是提醒，不该等于列表条数`)
  }

  fs.rmSync(dir, { recursive: true, force: true })

  if (problems.length) {
    console.error('✗ 订单状态的角色判定有问题：')
    problems.forEach((p) => console.error('   - ' + p))
    console.error('   （数据：' + all.map((o) => `${o.id} role=${o.role} status=${o.status} 标签=${o.status_text || mod.ORDER_STATUS[o.status]}`).join(' | ') + '）')
    process.exit(1)
  }
  console.log('✓ 订单状态按角色判定正确（卖家「已完成」/ 买家未评价「待评价」/ 已评价留「已完成」/ 售后单不混入 / 角标 1）')
}

main().catch((e) => { console.error('检查器自身出错：', e); process.exit(1) })
