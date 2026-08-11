import { ref } from 'vue'

// 自定义导航页顶部安全区：避开状态栏、微信胶囊与手机前摄/灵动岛区域。
//
// 两种模式：
//  - useSafeTop(true)  页面有自定义顶栏：顶栏与胶囊同排（topPad=胶囊顶），
//    右侧操作按钮通过 capsuleGap 水平避让胶囊；标题保持屏幕居中。
//  - useSafeTop()      页面无顶栏（如 tabbar 首页）：仅避开状态栏（topPad=状态栏高）。
export function useSafeTop(withNavBar = false) {
  const topPad = ref(24)
  const capsuleGap = ref(0)

  function initSafeTop() {
    try {
      const info = uni.getSystemInfoSync()
      const sb = info.statusBarHeight || 20
      let mr = null
      if (typeof uni.getMenuButtonBoundingClientRect === 'function') {
        mr = uni.getMenuButtonBoundingClientRect()
      }
      if (withNavBar && mr) {
        // 顶栏与胶囊同排；右侧按钮让出胶囊水平宽度
        topPad.value = mr.top
        capsuleGap.value = info.windowWidth - mr.left + 4
      } else {
        // 无顶栏页面只需避开状态栏，标题区自带内边距
        topPad.value = sb
      }
    } catch (e) { /* keep default */ }
  }

  return { topPad, capsuleGap, initSafeTop }
}
