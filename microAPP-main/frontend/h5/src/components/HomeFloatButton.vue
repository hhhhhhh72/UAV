<template>
  <div v-if="visible" class="home-float-btn" @click="goHome" aria-label="返回首页">
    <van-icon name="wap-home-o" size="20" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const visible = computed(() => route.path !== '/home')

const goHome = () => {
  if (route.path === '/home') return
  router.push('/home')
}
</script>

<style scoped>
.home-float-btn {
  position: fixed;
  right: 12px;
  /* 贴顶部：和回退键处于同一条导航栏行，兼容刘海/状态栏安全区 */
  top: calc(env(safe-area-inset-top) + 8px);
  z-index: 2001; /* 高于 NavBar，避免被遮挡 */
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent; /* 无轮廓/无胶囊背景 */
  border: none;
  box-shadow: none;
  color: #323233;
  user-select: none;
}

.home-float-btn:active {
  transform: scale(0.98);
  opacity: 0.85;
}

/* 点击态做轻微反馈，但不出现轮廓 */
.home-float-btn:active {
  opacity: 0.65;
}
</style>


