const fs = require('fs')
const path = require('path')
const files = process.argv.slice(2)
for (const f of files) {
  const s = fs.readFileSync(f, 'utf8')
  const i = s.indexOf('<template>')
  const j = s.lastIndexOf('</template>')
  const t = s.slice(i, j + 11)
  const re = /<\s*\/?\s*([a-zA-Z][\w-]*)((?:"[^"]*"|'[^']*'|[^>"'])*)>/g
  const stack = [] // {tag, line}
  let m
  let line = 1
  let bad = null
  while ((m = re.exec(t)) !== null) {
    const before = t.slice(0, m.index)
    line = before.split('\n').length
    const tag = m[1]
    const self = /\/\s*>$/.test(m[0])
    if (m[0][1] !== '/') {
      if (!self) stack.push({ tag, line })
    } else {
      if (!stack.length || stack[stack.length - 1].tag !== tag) {
        bad = `line ${line}: 闭合 </${tag}> 与栈顶 <${stack.length ? stack[stack.length - 1].tag : '?'}> (期望闭合行 ${stack.length ? stack[stack.length - 1].line : '-'}) 不匹配`
        break
      }
      stack.pop()
    }
  }
  if (!bad && stack.length) {
    bad = `文件末尾仍有 ${stack.length} 个未闭合: ${stack.map(x => `<${x.tag}>@${x.line}`).join(', ')}`
  }
  console.log(f + (bad ? ' ❌ ' + bad : ' ✅ 标签配对正常'))
}
