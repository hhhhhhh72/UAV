const sharp = require('C:/Users/21125/AppData/Local/Programs/dsh-desktop/resources/dsh/node_modules/sharp')

async function main() {
  const png = 'D:/w-yao/.tools/dsh-app/exe-icon.png'
  const img = sharp(png).ensureAlpha().resize(64, 64, { fit: 'contain' }).raw().toBuffer({ resolveWithObject: true })
  const { data: b, info } = await img
  const w = info.width, h = info.height
  const rows = []
  for (let y = 0; y < h; y++) {
    let n = 0
    for (let x = 0; x < w; x++) if (b[(y * w + x) * 4 + 3] > 16) n++
    rows.push(n)
  }
  const bands = []
  let start = -1
  for (let y = 0; y <= h; y++) {
    const on = y < h && rows[y] > 0
    if (on && start < 0) start = y
    if (!on && start >= 0) { bands.push([start, y - 1, Math.max(...rows.slice(start, y))]); start = -1 }
  }
  console.log('exe icon bands [y0,y1,maxWidth]:', JSON.stringify(bands))
  // 顶部/底部 8 行是否有不透明像素（杠检测）
  console.log('top rows alpha counts:', rows.slice(0, 8).join(','))
  console.log('bottom rows alpha counts:', rows.slice(h - 8).join(','))
  // 杠 = 行几乎满宽且位于顶部/底部；鱼身 = 中间
  console.log('row sample:')
  for (let y = 0; y < h; y += 8) console.log(String(y).padStart(2), Math.max(...rows.slice(y, y + 8)))
}
main()
