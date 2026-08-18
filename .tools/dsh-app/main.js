// DSH 桌面版 — 独立窗口加载本机 dsh web 服务
// 启动时自动探测服务，未运行则静默拉起（服务独立于应用，关闭应用后继续后台运行）
const { app, BrowserWindow, dialog } = require('electron')
const { spawn } = require('child_process')
const net = require('net')
const path = require('path')

const PORT = 3080
const APP_URL = `http://127.0.0.1:${PORT}`
const PROBE_TIMEOUT_MS = 180 * 1000 // 等待服务就绪上限（portable 首启需解压+杀软扫描+profile 编译，约 1-2 分钟）

function probePort() {
  return new Promise((resolve) => {
    const sock = net.connect({ host: '127.0.0.1', port: PORT })
    sock.once('connect', () => {
      sock.destroy()
      resolve(true)
    })
    sock.once('error', () => {
      sock.destroy()
      resolve(false)
    })
  })
}

function startServer() {
  // 用 Electron 自带 Node（RUN_AS_NODE 模式）跑内置的 dsh CLI——对方机器无需安装 Node
  // dsh 经 extraResources 放在 resources/dsh（真文件系统，非 asar）：
  // asar 内 ESM 包解析不全，dsh 的 profile 编译产物在 ~/.dsh 下 import 依赖会失败
  const dshRoot = path.join(process.resourcesPath, 'dsh')
  const dshBin = path.join(
    dshRoot,
    'node_modules',
    '@deepseek-ai',
    'dsh',
    'lib',
    'bin.js'
  )
  const child = spawn(process.execPath, [dshBin, 'web'], {
    detached: true,
    windowsHide: true,
    stdio: 'ignore',
    cwd: dshRoot,
    env: { ...process.env, ELECTRON_RUN_AS_NODE: '1' },
  })
  child.unref()
}

async function ensureServer() {
  if (await probePort()) return true
  startServer()
  const deadline = Date.now() + PROBE_TIMEOUT_MS
  while (Date.now() < deadline) {
    await new Promise((r) => setTimeout(r, 1000))
    if (await probePort()) return true
  }
  return false
}

async function main() {
  const ok = await ensureServer()

  const win = new BrowserWindow({
    width: 1200,
    height: 800,
    minWidth: 900,
    minHeight: 600,
    title: 'DSH',
    autoHideMenuBar: true,
    backgroundColor: '#ffffff',
  })

  if (ok) {
    await win.loadURL(APP_URL)
  } else {
    dialog.showErrorBox(
      'DSH 桌面版',
      '服务启动超时（180 秒）。\n请关闭后重试；若多次失败，请手动运行：npx -y @deepseek-ai/dsh web'
    )
    app.quit()
  }
}

app.whenReady().then(main)

app.on('window-all-closed', () => {
  // 只关窗口不杀服务——服务留后台，下次启动秒开
  app.quit()
})
