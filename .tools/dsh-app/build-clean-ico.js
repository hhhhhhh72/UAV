// 从干净 favicon.svg 重建 icon.ico（256×256 PNG-in-ICO，透明底，无黑杠）
// 流程与 app.asar 内 convert-icon.js 一致，但保留透明背景（不产生信箱式黑边）
const sharp = require('C:/Users/21125/AppData/Local/Programs/dsh-desktop/resources/dsh/node_modules/sharp')
const fs = require('fs')

async function main() {
  const svg = fs.readFileSync('D:/w-yao/.tools/dsh-app/src-favicon.svg')
  // 直接以透明画布渲染（不放背景色，不产生黑杠）。
  // 关键修复：resize fit:'contain' 时必须显式 background alpha:0，
  // 否则 sharp 默认用不透明黑色填充留白区 → 图标上下出现黑杠（原 convert-icon.js 的 bug）。
  const png = await sharp(svg, { density: 300 })
    .trim({ threshold: 5 })
    .resize(256, 256, { fit: 'contain', background: { r: 0, g: 0, b: 0, alpha: 0 } })
    .png()
    .toBuffer()

  // 验证：渲染出的 PNG 不得有满宽不透明行（黑杠检测）
  const probe = await sharp(png).raw().toBuffer()
  const w = 256
  let bars = 0
  for (let y = 0; y < w; y++) {
    let opaque = 0
    for (let x = 0; x < w; x++) if (probe[(y * w + x) * 4 + 3] > 200) opaque++
    if (opaque === w) bars++
  }
  if (bars > 0) {
    console.error('FAIL: 重建结果仍含', bars, '条满宽不透明行（黑杠未消除）')
    process.exit(1)
  }
  console.log('PNG 无满宽行（无黑杠），opaque check ok,', png.length, 'bytes')

  // ICO 容器（PNG 嵌入式，Vista+ 标准）
  const header = Buffer.alloc(6)
  header.writeUInt16LE(0, 0)
  header.writeUInt16LE(1, 2)
  header.writeUInt16LE(1, 4)
  const entry = Buffer.alloc(16)
  entry[0] = 0
  entry[1] = 0
  entry[2] = 0
  entry[3] = 0
  entry.writeUInt16LE(1, 4)
  entry.writeUInt16LE(32, 6)
  entry.writeUInt32LE(png.length, 8)
  entry.writeUInt32LE(22, 12)
  fs.writeFileSync('D:/w-yao/.tools/dsh-app/icon-clean.ico', Buffer.concat([header, entry, png]))
  console.log('icon-clean.ico written:', 22 + png.length, 'bytes')
}
main().catch((e) => {
  console.error(e)
  process.exit(1)
})
