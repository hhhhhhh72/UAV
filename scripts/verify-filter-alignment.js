#!/usr/bin/env node
// 筛选栏改造批量验证 v2：残留旧类名 + 标签配平 + 关键函数（区分单维/多维）
const fs = require('fs')
const path = require('path')

const PAGES = [
  // [相对路径, 是否多维（需面板+蒙层）]
  ['pages/demands/index.vue', true],
  ['pages/demands/list.vue', true],
  ['pkg-eco/pages/cases/index.vue', false],
  ['pkg-eco/pages/competitions/list.vue', false],
  ['pkg-eco/pages/enterprise/list.vue', false],
  ['pkg-emergency/pages/emergency/cases.vue', false],
  ['pkg-service/pages/compliance/knowledge.vue', false],
  ['pkg-service/pages/compliance/news.vue', false],
  ['pkg-service/pages/resources/list.vue', false],
  ['pkg-service/pages/testsites/list.vue', false],
  ['pkg-talent/pages/experts/list.vue', false],
  ['pkg-talent/pages/jobs/list.vue', false],
  ['pkg-talent/pages/study/index.vue', false],
  ['pkg-eco/pages/challenges/list.vue', true],
  ['pkg-eco/pages/projects/list.vue', true],
  ['pkg-app/pages/applications/index.vue', false],
  ['pkg-emergency/pages/emergency/dispatches.vue', false],
  ['pkg-demand/pages/demands/favorites.vue', false],
  ['pkg-demand/pages/demands/mine.vue', false],
  ['pkg-demand/pages/orders/mine.vue', false],
]

const STALE = [
  /\bclass="fbar\b/,
  /\bclass="fpill\b/,
  /\bclass="filter-scroll\b/,
  /\bclass="filter-bar\b/,
  /\bclass="filter-row\b/,
  /\.fpill\s*\{/,
  /\.fbar\s*\{/,
  /\.filter-area\s*\{/,
  /\.filter-pill\s*\{/,
  /class="filter-chip\b/,
  /class="mine-filters\b/,
  /class="orders-filters\b/,
  /class="filter-inner\b/,
]

// 模板 view 配对统计：属性引号内可含 '>'（如 :class="{x >= 1}"），
// 老实现 [^>]* 会在引号内 '>' 处提前截断导致自闭合漏数（历史上 dispatches.vue 误报过）。
// 与 scripts/check-wxml.js 同口径：逐标签解析（引号内字符原样跳过），只统计 template 区。
function templateViews(src) {
  const i = src.indexOf('<template>')
  const j = src.lastIndexOf('</template>')
  if (i < 0 || j < 0) return { open: 0, close: 0 }
  const t = src.slice(i, j + 11)
  const re = /<(\/?)(view)\b((?:\"[^\"]*\"|'[^']*'|[^>"'])*)>/g
  let open = 0
  let close = 0
  let m
  while ((m = re.exec(t)) !== null) {
    if (m[1] === '/') close++
    else if (!/\/>\s*$/.test(m[0])) open++
  }
  return { open, close }
}

let fail = 0
const results = []
for (const [rel, multi] of PAGES) {
  const fp = path.join('D:/w-yao/miniprogram', rel)
  if (!fs.existsSync(fp)) { console.log('[MISSING] ' + rel); fail++; continue }
  const src = fs.readFileSync(fp, 'utf8')
  const staleHits = STALE.filter(r => r.test(src))
  const { open, close } = templateViews(src)
  const balanced = open === close
  const hasStage = /\bstage-wrap\b|\bstages\b/.test(src)
  const hasStg = /\.stg\b/.test(src)
  const hasMask = /\bpanel-mask\b/.test(src)
  const hasToggle = /togglePanel|pickStageTab|pickX|pickTab/.test(src)
  const hasChip = /\.p-chip\b/.test(src)
  // 多维须有：panel-mask + togglePanel；单维须有：stg（纯 tab 即可）
  const multiOk = hasStage && hasStg && hasMask && hasToggle && hasChip
  const singleOk = hasStage && hasStg
  const ok = staleHits.length === 0 && balanced && (multi ? multiOk : singleOk)
  if (!ok) fail++
  const status = ok ? 'OK ' : 'FAIL'
  results.push(`[${status}] ${rel}  stale=[${staleHits.map(r => r.source).join(',')}]  balance=${balanced ? 'ok' : `view:${open}/close:${close}`}  stage=${hasStage} stg=${hasStg} mask=${hasMask} toggle=${hasToggle} chip=${hasChip}`)
}
console.log(results.join('\n'))
console.log('\n====== ' + (fail === 0 ? 'ALL PASS' : fail + ' FAILED') + ' ======')
