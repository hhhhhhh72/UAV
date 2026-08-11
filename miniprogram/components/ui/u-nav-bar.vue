<template>
  <view class="u-nav-bar" :class="{ 'u-nav-bar--fixed': fixed }" :style="{ paddingTop: statusBarHeight + 'px' }">
    <view class="u-nav-bar-inner">
      <view class="u-nav-bar-side u-nav-bar-left" @click="onBack">
        <u-icon v-if="leftText === '返回' || showBack" name="back" size="36rpx" color="#1a1a1a" />
        <text v-if="leftText && leftText !== '返回'" class="u-nav-bar-text">{{ leftText }}</text>
      </view>
      <text class="u-nav-bar-title">{{ title }}</text>
      <view class="u-nav-bar-side u-nav-bar-right" :style="{ right: rightOffset + 'px' }" @click="onRight">
        <text v-if="rightText" class="u-nav-bar-text">{{ rightText }}</text>
        <slot name="right" />
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { defineEmits } from 'vue'

defineProps({
  title: { type: String, default: '' },
  leftText: { type: String, default: '返回' },
  rightText: { type: String, default: '' },
  fixed: { type: Boolean, default: false },
  showBack: { type: Boolean, default: false }
})
const emit = defineEmits(['back', 'right'])
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
function onBack() {
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
.u-nav-bar-title { font-size: 32rpx; font-weight: 600; color: var(--color-text, #1a1a1a); }
.u-nav-bar-side { position: absolute; display: flex; align-items: center; padding: 0 24rpx; height: 100%; top: 0; }
.u-nav-bar-left { left: 0; gap: 8rpx; }
.u-nav-bar-right { right: 0; }
.u-nav-bar-text { font-size: 28rpx; color: var(--color-text, #1a1a1a); }
</style>
