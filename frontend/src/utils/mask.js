/**
 * 敏感数据脱敏工具（安全审计统一实现）
 * 详情默认脱敏 + 显式展开（revealPII 模式由页面自行控制），列表一律脱敏。
 * 例：maskPhone('13800138000') -> '138****8000'；maskIdCard('500101199001011234') -> '500101********1234'
 */
export function maskPhone(p) {
  if (!p) return ''
  if (p.length < 7) return p
  return p.slice(0, 3) + '****' + p.slice(-4)
}

export function maskIdCard(c) {
  if (!c) return ''
  if (c.length <= 10) return c
  return c.slice(0, 6) + '********' + c.slice(-4)
}

export function maskEmail(e) {
  if (!e) return ''
  const at = e.indexOf('@')
  if (at <= 0) return e
  const name = e.slice(0, at)
  const domain = e.slice(at)
  if (name.length <= 1) return '*' + domain
  return name[0] + '***' + name[name.length - 1] + domain
}

export function maskAddress(a) {
  if (!a) return ''
  if (a.length <= 6) return a
  return a.slice(0, 4) + '****' + a.slice(-2)
}

/**
 * 复合联系字段（"姓名 13800138000" / "电话：010-12345678"）：
 * 提取手机号/座机号码截尾脱敏，无号码文本原样返回（无 PII）。
 */
export function maskContact(c) {
  if (!c) return ''
  const s = String(c).trim()
  const m = s.match(/(1[3-9]\d{9}|0\d{2,3}-?\d{7,8})/)
  if (!m) return s
  return s.replace(m[0], maskPhone(m[0]))
}
