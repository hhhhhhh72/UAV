// 给 DSH 两个 exe 打图标补丁（重试 + 兜底替换）
// rcedit 5.x 具名导出 { rcedit }
const { rcedit } = require('rcedit')
const fs = require('fs')

const icon = 'D:/w-yao/.tools/dsh-app/icon-clean.ico'
const targets = [
  'C:/Users/21125/AppData/Local/Programs/dsh-desktop/DSH-Desktop.exe',
  'C:/Users/21125/AppData/Local/Programs/dsh-desktop/Uninstall DSH-Desktop.exe',
]

const sleep = (ms) => new Promise((r) => setTimeout(r, ms))

async function patchInPlace(exe) {
  for (let attempt = 1; attempt <= 3; attempt++) {
    try {
      await rcedit(exe, { icon })
      console.log('patched (attempt ' + attempt + '):', exe)
      return true
    } catch (e) {
      console.log('attempt ' + attempt + ' failed:', exe, '->', e && e.message)
      await sleep(4000)
    }
  }
  return false
}

async function patchViaReplace(exe) {
  const tmp = exe + '.iconpatch.tmp'
  console.log('fallback: copy -> patch copy -> replace for', exe)
  fs.copyFileSync(exe, tmp)
  await rcedit(tmp, { icon })
  for (let i = 0; i < 10; i++) {
    try {
      fs.renameSync(tmp, exe)
      console.log('replaced:', exe)
      return true
    } catch (e) {
      console.log('replace retry', i + 1, e && e.message)
      await sleep(2000)
    }
  }
  try { fs.unlinkSync(tmp) } catch (_) {}
  return false
}

async function main() {
  let exitCode = 0
  for (const exe of targets) {
    let ok = await patchInPlace(exe)
    if (!ok) ok = await patchViaReplace(exe)
    if (!ok) {
      console.error('FINAL FAIL:', exe)
      exitCode = 1
    }
  }
  process.exit(exitCode)
}
main()
