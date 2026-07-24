<template>
  <view class="page">
    <view class="header">消息中心</view>
    <view v-if="list.length === 0" class="empty">暂无消息</view>
    <view v-for="msg in list" :key="msg.id" class="msg-item" @click="markRead(msg.id)">
      <view class="msg-title">{{ msg.title }}</view>
      <view class="msg-content">{{ msg.content }}</view>
      <view class="msg-time">{{ msg.created_at }}</view>
      <view v-if="!msg.read" class="unread-dot" />
    </view>
  </view>
</template>

<script>
export default {
  data() { return { list: [], unread: 0 } },
  onShow() { this.fetchMessages(); this.fetchUnread() },
  methods: {
    async fetchMessages() {
      try {
        const res = await uni.request({ url: 'http://localhost:8080/api/v1/messages', header: { Authorization: 'Bearer ' + uni.getStorageSync('accessToken') } })
        this.list = res.data?.data?.items || res.data?.items || []
      } catch(e) { console.error(e) }
    },
    async fetchUnread() {
      try {
        const res = await uni.request({ url: 'http://localhost:8080/api/v1/messages/unread-count', header: { Authorization: 'Bearer ' + uni.getStorageSync('accessToken') } })
        this.unread = res.data?.data?.count || 0
      } catch(e) {}
    },
    async markRead(id) {
      try {
        await uni.request({ url: `http://localhost:8080/api/v1/messages/${id}/read`, method: 'POST', header: { Authorization: 'Bearer ' + uni.getStorageSync('accessToken') } })
        this.unread = Math.max(0, this.unread - 1)
        this.list = this.list.map(m => m.id === id ? {...m, read: true} : m)
      } catch(e) {}
    }
  }
}
</script>

<style scoped>
.page { padding: 16px; background: #f5f5f5; min-height: 100vh; }
.header { font-size: 20px; font-weight: bold; margin-bottom: 16px; }
.empty { text-align: center; color: #999; padding: 60px 0; }
.msg-item { background: #fff; padding: 14px; margin-bottom: 10px; border-radius: 8px; position: relative; }
.msg-title { font-size: 15px; font-weight: 600; }
.msg-content { font-size: 13px; color: #666; margin-top: 4px; }
.msg-time { font-size: 11px; color: #999; margin-top: 4px; }
.unread-dot { position: absolute; top: 14px; right: 14px; width: 8px; height: 8px; background: #ff3b30; border-radius: 50%; }
</style>
