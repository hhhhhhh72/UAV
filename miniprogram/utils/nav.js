import { authStorage } from './request'

let navLock = false
let loginLock = false

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

// 栈感知返回：有上一页则 navigateBack；无上一页（如从预约结果 reLaunch 进来的列表页）
// 则落到 fallbackUrl，避免在首屏 navigateBack 报 "cannot navigate back at first page"。
export const safeBack = (fallbackUrl = '/pages/home/index') => {
  const pages = getCurrentPages()
  if (pages.length > 1) {
    uni.navigateBack()
    return
  }
  uni.switchTab({ url: fallbackUrl })
}

export const safeReLaunch = (url) => {
  uni.reLaunch({ url })
}

// 登录守卫：未登录时 toast 提示并跳转登录页，带防重入（1.5s 内不重复触发）
export const requireLogin = () => {
  if (authStorage.getAccessToken()) return true
  if (loginLock) return false
  loginLock = true
  uni.showToast({ title: '请先登录', icon: 'none' })
  uni.navigateTo({
    url: '/pages/login/index',
    complete: () => {
      setTimeout(() => { loginLock = false }, 1500)
    },
  })
  return false
}
