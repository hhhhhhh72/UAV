<template>
  <view class="u-nav-bar" :class="{ 'u-nav-bar--fixed': fixed }" :style="{ paddingTop: statusBarHeight + 'px' }">
    <view class="u-nav-bar-inner">
      <view class="u-nav-bar-side u-nav-bar-left" @click="onBack">
        <u-icon v-if="leftText === '返回' || showBack" name="back" size="36rpx" color="#1a1a1a" />
        <text v-if="leftText && leftText !== '返回'" class="u-nav-bar-text">{{ leftText }}</text>
      </view>
      <text class="u-nav-bar-title" :style="titleStyle">{{ title }}</text>
      <view class="u-nav-bar-side u-nav-bar-right" :style="{ right: rightOffset + 'px' }" @click="onRight">
        <text v-if="rightText" class="u-nav-bar-text">{{ rightText }}</text>
        <slot name="right" />
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed, onMounted, useSlots } from 'vue'

const props = defineProps({
  title: { type: String, default: '' },
  leftText: { type: String, default: '返回' },
  rightText: { type: String, default: '' },
  fixed: { type: Boolean, default: false },
  showBack: { type: Boolean, default: false }
})
const emit = defineEmits(['back', 'right'])
const slots = useSlots()
const statusBarHeight = ref(20)
// 右侧内容右偏移：微信小程序胶囊避让（右缘内容不能被胶囊遮挡）
const rightOffset = ref(0)
onMounted(() => {
  // #ifdef MP-WEIXIN
  try {
    const sys = uni.getSystemInfoSync()
    statusBarHeight.value = sys.statusBarHeight || 20
    if (typeof uni.getMenuButtonBoundingClientRect === 'function') {
      const rect = uni.getMenuButtonBoundingClientRect()
      // 胶囊左缘到屏幕右缘的距离 + 8px 间距 = 右侧内容的右偏移
      rightOffset.value = Math.max(sys.windowWidth - rect.left + 8, 0)
    }
  } catch (e) { /* 取不到时保持默认 20 */ }
  // #endif
})

// 标题居中策略：
// - 无右侧内容（绝大多数页面）：屏幕正中央（left/right=0），保证短标题不偏移；
// - 有右侧内容（客服/筛选）：在「返回按钮 ~ 右侧内容」可用区内居中并省略过长，
//   避免与右侧文字重叠（修复「商品订单详情」撞「客服」）。
const titleStyle = computed(() => {
  const hasRight = !!props.rightText || !!slots.right
  return hasRight
    ? { left: '90px', right: (rightOffset.value + 80) + 'px' }
    : { left: '0px', right: '0px' }
})
function onBack() {
  // 冷启动直达（分享链接/扫码进入，页面栈仅 1 层）时 navigateBack 必失败：
  // 组件直接回首页，不派发页面返回逻辑，避免与页面自身 fail 兜底冲突。
  const pages = getCurrentPages()
  if (pages.length <= 1) {
    uni.switchTab({ url: '/pages/home/index' })
    return
  }
  emit('back')
}
function onRight() {
  emit('right')
}
</script>

<style scoped>
.u-nav-bar { background: #fff; }
.u-nav-bar--fixed { position: fixed; top: 0; left: 0; right: 0; z-index: 100; }
.u-nav-bar-inner { position: relative; display: flex; align-items: center; justify-content: center; height: 44px; }
.u-nav-bar-title {
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  font-weight: 600;
  color: var(--color-text, #1a1a1a);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  /* 标题绝对定位铺满导航栏但不拦截点击：返回/客服按钮必须可点（回归修复） */
  pointer-events: none;
}
.u-nav-bar-side { position: absolute; display: flex; align-items: center; padding: 0 24rpx; height: 100%; top: 0; z-index: 2; }
.u-nav-bar-left { left: 0; gap: 8rpx; }
.u-nav-bar-right { right: 0; }
.u-nav-bar-text { font-size: 28rpx; color: var(--color-text, #1a1a1a); pointer-events: none; }
</style>
