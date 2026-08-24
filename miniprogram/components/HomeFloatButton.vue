<template>
  <view v-if="visible" class="home-float-btn" :style="{ top: topOffset, right: rightOffset + 'px' }" @tap="goHome">
    <image class="home-icon" src="/static/icons/service.svg" mode="aspectFit" />
  </view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { onShow } from '@dcloudio/uni-app'

const currentRoute = ref('')
const statusBarHeight = ref(20)
const rightOffset = ref(12)

const refreshRoute = () => {
  try {
    const info = uni.getSystemInfoSync()
    statusBarHeight.value = info.statusBarHeight || 20
    // 避让右上角微信胶囊：按钮停在胶囊左缘左侧，否则被系统胶囊遮挡
    if (typeof uni.getMenuButtonBoundingClientRect === 'function') {
      const mr = uni.getMenuButtonBoundingClientRect()
      rightOffset.value = Math.max(info.windowWidth - mr.left + 6, 12)
    }
  } catch (e) {
    statusBarHeight.value = 20
  }
  const pages = getCurrentPages()
  const last = pages[pages.length - 1]
  currentRoute.value = last?.route || ''
}

onMounted(refreshRoute)
onShow(refreshRoute)

const visible = computed(() => currentRoute.value !== 'pages/home/index')

// 贴顶部：按状态栏高度自适应，确保和系统回退同一行视觉层级
const topOffset = computed(() => `${statusBarHeight.value + 6}px`)

const goHome = () => {
  uni.switchTab({ url: '/pages/home/index' })
}
</script>

<style scoped>
.home-float-btn {
  position: fixed;
  top: 0;
  z-index: 2001; /* 高于页面内容，避免被遮挡 */
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent; /* 无轮廓/无胶囊背景 */
  border: none;
  box-shadow: none;
  animation: floatBtnIn 0.4s ease both;
}

/* 悬浮按钮入场：淡入 + 上移 */
@keyframes floatBtnIn {
  from {
    transform: translateY(-8px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

.home-icon {
  width: 20px;
  height: 20px;
  opacity: 0.9;
}
</style>


