<template>
  <view class="tabbar-root">
    <van-tabbar
      :active="current"
      active-color="#1a1a1a"
      inactive-color="#969799"
      @change="handleTabChange"
      safe-area-inset-bottom
    >
      <van-tabbar-item icon="home-o" :name="0">
        <text>首页</text>
      </van-tabbar-item>
      <van-tabbar-item icon="apps-o" :name="1">
        <text>业务大厅</text>
      </van-tabbar-item>
      <van-tabbar-item icon="chat-o" :name="2" :badge="unreadBadge">
        <text>消息</text>
      </van-tabbar-item>
      <van-tabbar-item icon="user-o" :name="3">
        <text>我的</text>
      </van-tabbar-item>
    </van-tabbar>
  </view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'

const props = defineProps(['current'])

const tabPages = [
  '/pages/home/index',
  '/pages/services/index',
  '/pages/messages/index',
  '/pages/mine/index',
]

const unreadCount = ref(0)

const unreadBadge = computed(() => {
  if (unreadCount.value <= 0) return ''
  return unreadCount.value > 99 ? '99+' : String(unreadCount.value)
})

const updateUnreadCount = () => {
  try {
    const count = uni.getStorageSync('unreadCount')
    unreadCount.value = Number(count) || 0
  } catch (e) {
    unreadCount.value = 0
  }
}

onMounted(() => {
  updateUnreadCount()
  // Poll unread count from storage periodically
  setInterval(updateUnreadCount, 5000)
})

const handleTabChange = (e) => {
  const index = typeof e.detail === 'object' ? e.detail.index : e.detail
  const idx = Number(index)
  if (idx === Number(props.current)) return
  uni.switchTab({ url: tabPages[idx] })
}
</script>

<style scoped>
.tabbar-root {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 999;
}
</style>
