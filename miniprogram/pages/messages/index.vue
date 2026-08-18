<template>
  <view class="messages-page">
    <!-- 加载状态 -->
    <view v-if="loading" class="loading-state">
      <view class="loading-inline">
        <u-loading size="24rpx" />
        <text>加载中...</text>
      </view>
    </view>

    <!-- 空状态 -->
    <view v-else-if="messages.length === 0" class="empty-state-wrapper">
      <u-empty description="暂无消息" />
    </view>

    <!-- 消息列表 -->
    <view v-else class="message-list">
      <u-cell-group inset>
        <u-cell
          v-for="msg in messages"
          :key="msg.id"
          :label="msg.content"
          :value="formatTime(msg.created_at || msg.createdAt)"
          is-link
          @click="onMessageClick(msg)"
        >
          <template #icon>
            <view class="msg-icon-wrapper">
              <text class="msg-icon-text">信</text>
              <view v-if="!(msg.is_read || msg.isRead)" class="unread-dot" />
            </view>
          </template>
          <template #title>
            <text class="cell-title-text">{{ msg.title }}</text>
          </template>
        </u-cell>
      </u-cell-group>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow, onPullDownRefresh } from '@dcloudio/uni-app'
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
    // silent fail
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
  background: var(--color-bg);
  min-height: 100vh;
  padding-bottom: 24px;
}

.loading-state {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}

.loading-inline {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: var(--color-text-secondary);
}

.empty-state-wrapper {
  padding-top: 60px;
}

.message-list {
  margin: 12px 0;
}

.msg-icon-wrapper {
  position: relative;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--color-primary-light);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 12px;
}

.msg-icon-text {
  font-size: 16px;
  color: var(--color-primary);
  font-weight: 600;
}

.unread-dot {
  position: absolute;
  top: -2px;
  right: -2px;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: var(--color-danger);
  border: 2px solid #fff;
}

.cell-title-text {
  font-size: 15px;
  color: var(--color-text);
}

.is-unread .cell-title-text {
  font-weight: 600;
  color: var(--color-text);
}
</style>
