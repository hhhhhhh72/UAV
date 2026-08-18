// 一次性脚本：dsh favicon.svg → build/icon.ico（256×256，PNG-in-ICO）
const sharp = require('sharp')
const fs = require('fs')
const path = require('path')

async function main() {
  const svg = fs.readFileSync(path.join(__dirname, 'favicon.svg'))
  // 裁掉透明边缘后铺满 256 画布（透明底、无白框）
  const png = await sharp(svg, { density: 300 })
    .trim({ threshold: 5 })
    .resize(256, 256, { fit: 'contain' })
    .png()
    .toBuffer()

  // ICO 容器头（PNG 嵌入式，Vista+ 标准）
  const header = Buffer.alloc(6)
  header.writeUInt16LE(0, 0) // reserved
  header.writeUInt16LE(1, 2) // type: icon
  header.writeUInt16LE(1, 4) // count
  const entry = Buffer.alloc(16)
  entry[0] = 0 // width 256 以 0 表示
  entry[1] = 0 // height 256 以 0 表示
  entry[2] = 0 // palette
  entry[3] = 0 // reserved
  entry.writeUInt16LE(1, 4) // planes
  entry.writeUInt16LE(32, 6) // bpp
  entry.writeUInt32LE(png.length, 8)
  entry.writeUInt32LE(22, 12)

  fs.mkdirSync('build', { recursive: true })
  fs.writeFileSync('build/icon.ico', Buffer.concat([header, entry, png]))
  console.log('build/icon.ico written,', png.length, 'bytes PNG payload')
}
main().catch((e) => {
  console.error(e)
  process.exit(1)
})
