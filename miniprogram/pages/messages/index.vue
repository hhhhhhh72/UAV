<template>
  <Layout :current="2">
    <view class="messages-page">
      <!-- 加载状态 -->
      <view v-if="loading" class="loading-state">
        <van-loading size="24">加载中...</van-loading>
      </view>

      <!-- 空状态 -->
      <view v-else-if="messages.length === 0" class="empty-state-wrapper">
        <van-empty description="暂无消息" image="search" />
      </view>

      <!-- 消息列表 -->
      <view v-else class="message-list">
        <van-cell-group inset>
          <van-cell
            v-for="msg in messages"
            :key="msg.id"
            :title="msg.title"
            :label="msg.content"
            :value="formatTime(msg.created_at || msg.createdAt)"
            is-link
            @tap="onMessageClick(msg)"
          >
            <template #icon>
              <view class="msg-icon-wrapper">
                <van-icon name="chat" size="20" color="#1989fa" />
                <view v-if="!(msg.is_read || msg.isRead)" class="unread-dot" />
              </view>
            </template>
          </van-cell>
        </van-cell-group>
      </view>
    </view>
  </Layout>
</template>

<script setup>
import { ref } from 'vue'
import { onShow, onPullDownRefresh } from '@dcloudio/uni-app'
import Layout from '@/components/Layout.vue'
import { request, getStoredUser } from '../../utils/request'

const messages = ref([])
const loading = ref(false)

const fetchMessages = async () => {
  const user = getStoredUser()
  if (!user) {
    messages.value = []
    return
  }

  loading.value = true
  try {
    const res = await request({ url: '/api/v1/messages' })
    const list = res?.data || res || []
    messages.value = Array.isArray(list) ? list : []
  } catch (e) {
    console.warn('Failed to load messages:', e)
    messages.value = []
  } finally {
    loading.value = false
  }
}

const fetchUnreadCount = async () => {
  try {
    const res = await request({ url: '/api/v1/messages/unread-count' })
    const count = res?.data?.count ?? res?.count ?? 0
    uni.setStorageSync('unreadCount', count)
  } catch (e) {
    // ignore
  }
}

const onMessageClick = async (msg) => {
  const read = msg.is_read || msg.isRead
  if (read) return

  try {
    await request({
      url: `/api/v1/messages/${msg.id}/read`,
      method: 'POST',
    })
    msg.is_read = true
    msg.isRead = true
    // Update unread count
    fetchUnreadCount()
  } catch (e) {
    uni.showToast({ title: '操作失败', icon: 'none' })
  }
}

const formatTime = (timeStr) => {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  const now = new Date()
  const diffMs = now - date
  const diffMin = Math.floor(diffMs / 60000)
  const diffHour = Math.floor(diffMs / 3600000)
  const diffDay = Math.floor(diffMs / 86400000)

  if (diffMin < 1) return '刚刚'
  if (diffMin < 60) return `${diffMin}分钟前`
  if (diffHour < 24) return `${diffHour}小时前`
  if (diffDay < 7) return `${diffDay}天前`
  return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}

onShow(() => {
  fetchMessages()
  fetchUnreadCount()
})

onPullDownRefresh(() => {
  fetchMessages().then(() => {
    uni.stopPullDownRefresh()
  })
})
</script>

<style scoped>
.messages-page {
  background: #f7f8fa;
  min-height: 100vh;
  padding-bottom: 24px;
}

.loading-state {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}

.empty-state-wrapper {
  padding-top: 60px;
}

.message-list {
  margin: 12px 0;
}

.msg-icon-wrapper {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: #e8f4fd;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 12px;
}

.is-unread .cell-title-text {
  font-weight: 600;
  color: #1a1a1a;
}
</style>
