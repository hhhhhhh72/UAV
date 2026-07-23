let navLock = false

export const safeNavigateTo = (url) => {
  if (navLock) return
  navLock = true
  
  uni.navigateTo({
    url,
    success: () => {
      // 成功后快速释放，300ms 跨过点击间隔即可
      setTimeout(() => { navLock = false }, 300)
    },
    fail: (err) => {
      console.error('[safeNavigateTo] fail:', url, err)
      // 如果超时或栈满，强制 redirectTo 挽救
      if (err.errMsg && (err.errMsg.includes('timeout') || err.errMsg.includes('full'))) {
        uni.redirectTo({ 
          url,
          complete: () => { navLock = false }
        })
      } else {
        navLock = false
      }
    }
  })
}

export const safeSwitchTab = (url) => {
  uni.switchTab({ url })
}

export const safeReLaunch = (url) => {
  uni.reLaunch({ url })
}
