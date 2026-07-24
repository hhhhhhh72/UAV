<template>
  <view v-if="visible" class="home-float-btn" :style="{ top: topOffset }" @tap="goHome">
    <image class="home-icon" src="/static/icons/service.svg" mode="aspectFit" />
  </view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { onShow } from '@dcloudio/uni-app'

const currentRoute = ref('')
const statusBarHeight = ref(20)

const refreshRoute = () => {
  try {
    statusBarHeight.value = uni.getSystemInfoSync().statusBarHeight || 20
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
  right: 12px;
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
}

.home-icon {
  width: 20px;
  height: 20px;
  opacity: 0.9;
}
</style>


