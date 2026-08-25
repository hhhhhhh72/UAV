// 联系方式信任门槛（产学研详情/转化追踪页共用）：
// 公开展示一律脱敏（maskContact），完整值经用户确认后复制（revealContactCopy）。
// 与品牌色/设计体系无关，纯工具；不引入 mock。

// maskContact 手机号脱敏：保留前 3 后 4（138****1234）；非手机号原样返回。
export function maskContact(v) {
  if (!v) return ''
  const s = String(v).trim()
  if (/^1\d{10}$/.test(s)) return s.slice(0, 3) + '****' + s.slice(-4)
  return s
}

// revealContactCopy 复制完整联系方式到剪贴板（调用方先完成确认弹窗再调用）。
export function revealContactCopy(value) {
  if (!value) return Promise.reject(new Error('empty'))
  return new Promise((resolve, reject) => {
    uni.setClipboardData({
      data: String(value),
      success: () => {
        uni.showToast({ title: '联系方式已复制', icon: 'none' })
        resolve(true)
      },
      fail: (e) => reject(e),
    })
  })
}
