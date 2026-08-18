const sharp = require('C:/Users/21125/AppData/Local/Programs/dsh-desktop/resources/dsh/node_modules/sharp')

async function main() {
  const svg = 'D:/w-yao/.tools/dsh-app/src-favicon.svg'
  const b = await sharp(svg, { density: 384 }).resize(256, 256, { fit: 'contain' }).ensureAlpha().raw().toBuffer()
  const w = 256, h = 256
  const rows = []
  for (let y = 0; y < h; y++) {
    let n = 0
    for (let x = 0; x < w; x++) {
      if (b[(y * w + x) * 4 + 3] > 16) n++
    }
    rows.push(n)
  }
  const bands = []
  let start = -1
  for (let y = 0; y <= h; y++) {
    const on = y < h && rows[y] > 0
    if (on && start < 0) start = y
    if (!on && start >= 0) {
      bands.push([start, y - 1, Math.max(...rows.slice(start, y))])
      start = -1
    }
  }
  console.log('continuous bands [y0,y1,maxWidth]:', JSON.stringify(bands))
  console.log('row sample (max width per 16 rows):')
  for (let y = 0; y < h; y += 16) {
    console.log(String(y).padStart(3) + '-' + String(Math.min(y + 15, 255)).padStart(3), Math.max(...rows.slice(y, y + 16)))
  }
  // 每 8 列的宽度分布（看杠是否全宽）
  console.log('col sample (max height per 32 cols):')
  for (let x = 0; x < w; x += 32) {
    let n = 0
    for (let yy = 0; yy < h; yy++) for (let xx = x; xx < Math.min(x + 32, w); xx++) {
      if (b[(yy * w + xx) * 4 + 3] > 16) { n++; break }
    }
    console.log('cols', x, '-', Math.min(x + 31, 255), 'rows covered:', n)
  }
}
main()
