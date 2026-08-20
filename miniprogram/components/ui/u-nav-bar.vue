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
import { defineEmits } from 'vue'

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

// 标题居中但必须避让左右侧内容（返回/客服/筛选），过长省略——
// 修复「商品订单详情」等长标题与右侧「客服」重叠挤压。
const titleStyle = computed(() => {
  const hasLeft = props.showBack || !!props.leftText
  const hasRight = !!props.rightText || !!slots.right
  return {
    left: (hasLeft ? 96 : 12) + 'px',
    right: (hasRight ? rightOffset.value + 72 : 12) + 'px',
  }
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
.u-nav-bar-title {
  position: absolute;
  top: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  font-weight: 600;
  color: var(--color-text, #1a1a1a);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.u-nav-bar-side { position: absolute; display: flex; align-items: center; padding: 0 24rpx; height: 100%; top: 0; }
.u-nav-bar-left { left: 0; gap: 8rpx; }
.u-nav-bar-right { right: 0; }
.u-nav-bar-text { font-size: 28rpx; color: var(--color-text, #1a1a1a); }
</style>
