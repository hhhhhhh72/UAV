const sharp = require('C:/Users/21125/AppData/Local/Programs/dsh-desktop/resources/dsh/node_modules/sharp')

async function main() {
  // 原始渲染：density 300，不做任何 trim/resize
  const img = await sharp('D:/w-yao/.tools/dsh-app/src-favicon.svg', { density: 300 })
    .ensureAlpha().raw().toBuffer({ resolveWithObject: true })
  const { data: b, info } = img
  const w = info.width, h = info.height
  console.log('canvas:', w, 'x', h)
  const rows = []
  for (let y = 0; y < h; y++) {
    let n = 0
    for (let x = 0; x < w; x++) if (b[(y * w + x) * 4 + 3] > 200) n++
    rows.push(n)
  }
  const bands = []
  let start = -1
  for (let y = 0; y <= h; y++) {
    const on = y < h && rows[y] > 0
    if (on && start < 0) start = y
    if (!on && start >= 0) { bands.push([start, y - 1, Math.max(...rows.slice(start, y))]); start = -1 }
  }
  console.log('bands:', JSON.stringify(bands))
  // 顶部 12 行和底部 12 行的每行不透明数 + 颜色
  for (const y of [0, 1, 2, Math.floor(h / 2), h - 3, h - 2, h - 1]) {
    let op = 0, r = 0, g = 0, bl = 0, cnt = 0
    for (let x = 0; x < w; x++) {
      const i = (y * w + x) * 4
      if (b[i + 3] > 200) { op++; r += b[i]; g += b[i + 1]; bl += b[i + 2]; cnt++ }
    }
    if (cnt > 0) console.log('row', y, 'opaque:', op, '/', w, 'avgRGB:', Math.round(r / cnt), Math.round(g / cnt), Math.round(bl / cnt))
  }
}
main()
