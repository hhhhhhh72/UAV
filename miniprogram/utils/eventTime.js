// 活动时间工具：后端 time.Time 序列化为 RFC3339（UTC），小程序端按本地时区展示
const pad = (n) => (n < 10 ? '0' + n : '' + n)

// 从任意日期/时间值取本地日历日 'YYYY-MM-DD'；非法值返回 ''
export function dateOf(v) {
  if (!v) return ''
  const s = String(v)
  if (s.includes('T')) {
    const d = new Date(s)
    if (!isNaN(d.getTime())) return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate())
    return s.slice(0, 10)
  }
  return s.slice(0, 10)
}

// 从时间值取本地 'HH:mm'；无时间部分或非法返回 ''
export function timeOf(v) {
  if (!v) return ''
  const s = String(v)
  if (!s.includes('T')) return ''
  const d = new Date(s)
  if (isNaN(d.getTime())) return ''
  return pad(d.getHours()) + ':' + pad(d.getMinutes())
}
