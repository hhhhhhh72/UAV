/**
 * 减弱动效检测（无障碍规范）
 *
 * 系统开启"减弱动态效果"（reduceMotion）时 noMotion 变为 true：
 * 调用方在根节点挂 no-motion class，CSS 中关闭装饰动画与位移/缩放，保留淡入与颜色反馈。
 * 低端安卓机型降级共用同一条路径（页面自行决定何时触发降级）。
 *
 * 用法：
 *   const { noMotion, checkMotion } = useReduceMotion()
 *   onLoad(checkMotion)
 */
import { ref } from 'vue'

export function useReduceMotion() {
  const noMotion = ref(false)
  const checkMotion = () => {
    try {
      const sys = uni.getSystemInfoSync()
      if (sys && sys.reduceMotion) noMotion.value = true // 快照检测（基础库 2.11.0+，旧版 undefined）
    } catch (e) { /* 忽略 */ }
    try {
      if (typeof uni.onAccessibilityInfoChange === 'function') {
        uni.onAccessibilityInfoChange((res) => { noMotion.value = !!(res && res.reduceMotion) }) // 开关变化实时响应
      }
    } catch (e) { /* 旧基础库无此 API */ }
  }
  return { noMotion, checkMotion }
}
