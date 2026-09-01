const fs = require('fs')
const path = require('path')
const root = path.join(__dirname, '..', 'frontend', 'src', 'views', 'admin')
const files = []
;(function walk(d) {
  for (const e of fs.readdirSync(d, { withFileTypes: true })) {
    const p = path.join(d, e.name)
    if (e.isDirectory()) walk(p)
    else if (e.name.endsWith('.vue')) files.push(p)
  }
})(root)

const allowed = ['content', 'description', 'detail', 'intro', 'summary', 'body', 'result', 'achievements']
let changed = 0
for (const f of files) {
  let s = fs.readFileSync(f, 'utf8')
  // 只匹配单个自闭合 a-input/a-textarea 标签（此前版本 [\s\S]*? 会跨标签吞噬，
  // 例如把「院校名称」到「描述」之间的整个表单替换成一条 RichEditor）。
  // [^>] 不能跨 '>'，因此匹配边界必然等于单标签自身；属性换行（v-model/type 分行）也覆盖。
  const tagRe = /<(?:a-input|a-textarea)\b[^>]*?\btype="textarea"[^>]*?\/>/g
  const ms = [...s.matchAll(tagRe)]
  if (!ms.length) continue
  const repl = [] // {start, end, html}
  for (const m of ms) {
    const vm = /v-model="(\w+)\.([a-zA-Z_]+)"/.exec(m[0])
    if (vm && allowed.includes(vm[2]) && !m[0].includes('RichEditor')) {
      repl.push({ start: m.index, end: m.index + m[0].length, html: '<RichEditor v-model="' + vm[1] + '.' + vm[2] + '" />' })
    }
  }
  if (!repl.length) continue
  let out = s
  for (const r of repl.reverse()) { // 从后往前替换，索引不漂移
    out = out.slice(0, r.start) + r.html + out.slice(r.end)
  }
  if (!out.includes("import RichEditor from '@/components/RichEditor.vue'")) {
    const lines = out.split('\n')
    let last = -1
    for (let i = 0; i < lines.length; i++) if (/^import .*from/.test(lines[i])) last = i
    if (last >= 0) lines.splice(last + 1, 0, "import RichEditor from '@/components/RichEditor.vue'")
    out = lines.join('\n')
  }
  fs.writeFileSync(f, out)
  changed++
  console.log('✅', path.relative(root, f), '->', repl.map(r => r.html.match(/v-model="([^"]+)"/)[1]).join(','))
}
console.log('共改造页面数:', changed)
