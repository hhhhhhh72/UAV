<template>
  <div class="messages-page">
    <van-nav-bar title="消息中心" fixed placeholder />

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-state">
      <van-loading size="24">加载中...</van-loading>
    </div>

    <!-- 空状态 -->
    <div v-else-if="!messageStore.list.length" class="empty-state">
      <van-icon name="chat-o" size="48" color="#c8c9cc" />
      <p>暂无消息</p>
    </div>

    <!-- 消息列表 -->
    <div v-else class="message-list">
      <van-cell-group inset>
        <van-cell
          v-for="msg in messageStore.list"
          :key="msg.id"
          :title="msg.title"
          :label="msg.content"
          :value="formatTime(msg.created_at)"
          is-link
          :class="{ 'is-unread': !msg.is_read }"
          @click="onMessageClick(msg)"
        >
          <template #icon>
            <van-badge :dot="!msg.is_read">
              <div class="msg-icon-wrapper">
                <van-icon name="chat" size="20" color="#1989fa" />
              </div>
            </van-badge>
          </template>
        </van-cell>
      </van-cell-group>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { showToast } from 'vant'
import { useMessageStore } from '@/stores/message'

const messageStore = useMessageStore()
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    await messageStore.fetchMessages()
  } catch {
    showToast('加载消息失败')
  } finally {
    loading.value = false
  }
})

async function onMessageClick(msg) {
  if (msg.is_read) return
  try {
    await messageStore.markRead(msg.id)
    msg.is_read = true
  } catch {
    showToast('操作失败')
  }
}

function formatTime(timeStr) {
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
</script>

<style scoped>
.messages-page {
  background: #f5f5f7;
  min-height: 100vh;
  padding-bottom: 24px;
}

.messages-page :deep(.van-nav-bar--fixed) {
  left: 50% !important;
  transform: translateX(-50%) !important;
  width: 100% !important;
  max-width: var(--page-max-width);
}

.loading-state {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}

.empty-state {
  text-align: center;
  padding: 80px 0;
  color: #969799;
}

.empty-state p {
  margin: 12px 0 0;
  font-size: 14px;
}

.message-list {
  margin: 12px 0;
}

.message-list :deep(.van-cell) {
  align-items: flex-start;
}

.message-list :deep(.van-cell__title) {
  font-size: 14px;
}

.message-list :deep(.van-cell__label) {
  font-size: 13px;
  color: #969799;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.message-list :deep(.van-cell__value) {
  font-size: 12px;
  color: #c8c9cc;
  flex: 0 0 auto;
  white-space: nowrap;
  margin-left: 8px;
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

.is-unread :deep(.van-cell__title) {
  font-weight: 600;
  color: #1a1a1a;
}
</style>
