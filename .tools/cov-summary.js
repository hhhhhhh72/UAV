// 汇总 coverprofile 的总体语句覆盖率（node 版，与 PowerShell 循环解耦）
const fs = require('fs')
const path = require('path')

const dir = process.argv[2]
if (!dir) { console.error('usage: node cov-summary.js <profdir>'); process.exit(1) }
const files = fs.readdirSync(dir).filter((f) => f.endsWith('.out'))
let totalStmts = 0
let coveredStmts = 0
const per = []

for (const f of files) {
  const lines = fs.readFileSync(path.join(dir, f), 'utf8').split('\n')
  let pkgTotal = 0
  let pkgCovered = 0
  for (const line of lines) {
    if (line.startsWith('mode:') || !line.trim()) continue
    const parts = line.split(' ')
    const numStmt = parseInt(parts[1], 10)
    const count = parseInt(parts[2], 10)
    pkgTotal += numStmt
    if (count > 0) pkgCovered += numStmt
  }
  totalStmts += pkgTotal
  coveredStmts += pkgCovered
  per.push({ p: f.replace('.out', '').replace('internal_', ''), pkgTotal, pkgCovered, pct: pkgTotal ? (pkgCovered / pkgTotal * 100) : 0 })
}

per.sort((a, b) => b.pct - a.pct)
for (const { p, pkgTotal, pkgCovered, pct } of per) {
  console.log(p.padEnd(24), String(pkgCovered).padStart(6), '/', String(pkgTotal).padStart(6), '=', pct.toFixed(1).padStart(6) + '%')
}
console.log('--------------------------------------------------')
console.log('TOTAL'.padEnd(24), String(coveredStmts).padStart(6), '/', String(totalStmts).padStart(6), '=', (coveredStmts / totalStmts * 100).toFixed(2) + '%')
