<template>
  <view class="messages-page">
    <!-- ① 分类栏：全部 / 未读 / 公告 / 供需 / 求职 / 认证
         分类按 resource_type 归组（消息表已有该字段，无需新表字段）；未读取服务端口径 -->
    <scroll-view scroll-x class="tabs" :show-scrollbar="false">
      <view class="tabs-inner">
        <view
          v-for="t in tabList"
          :key="t.key"
          class="tab"
          :class="{ act: tab === t.key }"
          @tap="tab = t.key"
        >
          <text class="tab-label">{{ t.label }}</text>
          <text v-if="t.n > 0" class="tab-num" :class="{ alert: t.alert }">{{ t.n > 99 ? '99+' : t.n }}</text>
        </view>
      </view>
    </scroll-view>

    <!-- ② 加载态 -->
    <view v-if="loading" class="loading-state">
      <view class="loading-inline">
        <u-loading size="24rpx" />
        <text>加载中...</text>
      </view>
    </view>

    <!-- ③ 空态（分口径：全部空 vs 该分类空） -->
    <view v-else-if="shown.length === 0" class="empty-state-wrapper">
      <u-empty :description="emptyText" />
      <view v-if="tab !== 'all' && messages.length > 0" class="empty-back" @tap="tab = 'all'">看全部消息</view>
    </view>

    <!-- ④ 消息列表 -->
    <view v-else class="message-list">
      <u-cell-group inset>
        <u-cell
          v-for="msg in shown"
          :key="msg.id"
          :label="stripHtml(msg.content)"
          :value="formatTime(msg.created_at || msg.createdAt)"
          :is-link="!!targetOf(msg)"
          @click="onMessageClick(msg)"
        >
          <template #icon>
            <view class="msg-icon-wrapper" :class="'cat-' + categoryOf(msg)">
              <text class="msg-icon-text">{{ catChar(msg) }}</text>
              <view v-if="!isRead(msg)" class="unread-dot" />
            </view>
          </template>
          <template #title>
            <text class="cell-title-text" :class="{ 'is-unread': !isRead(msg) }">{{ msg.title }}</text>
          </template>
        </u-cell>
      </u-cell-group>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onShow, onPullDownRefresh } from '@dcloudio/uni-app'
import { request, getStoredUser } from '../../utils/request'
import { stripHtml } from '../../utils/html'

const messages = ref([])
const loading = ref(false)
const tab = ref('all')

// ── 分类：resource_type → 分组（'notice' 收空值 = 管理端公告广播）──
const TABS = [
  { key: 'all', label: '全部' },
  { key: 'unread', label: '未读', alert: true },
  { key: 'notice', label: '公告' },
  { key: 'supply', label: '供需' },
  { key: 'talent', label: '求职' },
  { key: 'cert', label: '认证' },
]
const CATEGORY_KEYS = {
  notice: ['', 'notice', 'system'],
  supply: ['demand', 'work_order'],
  talent: ['application'],
  cert: ['pilot', 'enterprise', 'certificate'],
}
const categoryOf = (msg) => {
  const rt = String((msg && msg.resource_type) || '')
  for (const key of Object.keys(CATEGORY_KEYS)) {
    if (CATEGORY_KEYS[key].includes(rt)) return key
  }
  return 'notice'
}
// 每个分类一个字符标识 + 全站四色板底色（不引新资源，沿用页面既有"字符图标"口径）
const CAT_CHAR = { notice: '公', supply: '需', talent: '职', cert: '证' }
const CAT_CHAR_BY_TYPE = { work_order: '单', enterprise: '企', pilot: '飞', certificate: '证', application: '职' }
const catChar = (msg) => CAT_CHAR_BY_TYPE[String((msg && msg.resource_type) || '')] || CAT_CHAR[categoryOf(msg)]
const isRead = (msg) => !!(msg && (msg.is_read || msg.isRead))

const matchesTab = (msg, key) => {
  if (key === 'all') return true
  if (key === 'unread') return !isRead(msg)
  return categoryOf(msg) === key
}
const countFor = (key) => messages.value.filter((m) => matchesTab(m, key)).length
const tabList = computed(() => TABS.map((t) => ({ ...t, n: countFor(t.key) })))
const shown = computed(() => messages.value.filter((m) => matchesTab(m, tab.value)))
const emptyText = computed(() => {
  if (messages.value.length === 0) return '暂无消息'
  if (tab.value === 'unread') return '没有未读消息'
  return '该分类下暂无消息'
})

// ── 点击跳转：消息自带的 resource_type/resource_id 正好是业务页入参 ──
const TARGETS = {
  demand: (id) => '/pages/demands/detail?id=' + encodeURIComponent(id),
  work_order: (id) => '/pages/work-orders/detail?id=' + encodeURIComponent(id),
  enterprise: (id) => '/pkg-eco/pages/enterprise/detail?id=' + encodeURIComponent(id),
  pilot: (id) => '/pkg-talent/pages/pilots/detail?id=' + encodeURIComponent(id),
  certificate: () => '/pkg-talent/pages/training/certificates',
  application: () => '/pkg-talent/pages/jobs/applications',
}
const targetOf = (msg) => {
  const rt = String((msg && msg.resource_type) || '')
  const build = TARGETS[rt]
  if (!build) return ''
  const id = String((msg && msg.resource_id) || '')
  const needsId = rt === 'demand' || rt === 'work_order' || rt === 'enterprise' || rt === 'pilot'
  if (needsId && !id) return ''
  return build(id)
}

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

const markRead = async (msg) => {
  if (isRead(msg)) return
  try {
    await request({ url: '/api/v1/messages/' + msg.id + '/read', method: 'POST' })
    msg.is_read = true
    msg.isRead = true
    fetchUnreadCount()
  } catch (e) {
    uni.showToast({ title: '操作失败', icon: 'none' })
  }
}

const onMessageClick = (msg) => {
  const url = targetOf(msg)
  // 标已读不挡住跳转：失败也只是红点晚一步消失
  markRead(msg)
  if (!url) return
  uni.navigateTo({ url, fail: () => uni.showToast({ title: '打开失败，请稍后重试', icon: 'none' }) })
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

/* ===== 分类栏 ===== */
.tabs {
  background: #ffffff;
  border-bottom: 1rpx solid #EBEDF0;
  white-space: nowrap;
}
.tabs-inner {
  display: inline-flex;
  align-items: center;
  gap: 8rpx;
  padding: 16rpx 24rpx;
}
.tab {
  display: flex;
  align-items: center;
  gap: 6rpx;
  padding: 10rpx 24rpx;
  border-radius: 999rpx;
  background: #F5F6F8;
}
.tab.act { background: #EAF3FB; }
.tab-label {
  font-size: 26rpx;
  color: #667085;
  font-weight: 500;
}
.tab.act .tab-label {
  color: #0A66C2;
  font-weight: 700;
}
.tab-num {
  font-size: 22rpx;
  color: #667085;
  font-variant-numeric: tabular-nums;
}
.tab.act .tab-num { color: #0A66C2; }
.tab-num.alert { color: #D92D20; font-weight: 700; }

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
  display: flex;
  flex-direction: column;
  align-items: center;
}
.empty-back {
  margin-top: 24rpx;
  padding: 14rpx 40rpx;
  border-radius: 999rpx;
  border: 2rpx solid #0A66C2;
  color: #0A66C2;
  font-size: 26rpx;
}

.message-list {
  margin: 12px 0;
}

/* 分类字符图标：四色板同源（蓝/橙/绿/紫 + 中性） */
.msg-icon-wrapper {
  position: relative;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 12px;
  background: #EEF1F4; /* 公告：中性 */
}
.msg-icon-text {
  font-size: 16px;
  font-weight: 700;
  color: #667085;
}
.cat-supply.msg-icon-wrapper { background: #EAF3FB; }
.cat-supply .msg-icon-text { color: #0A66C2; }
.cat-talent.msg-icon-wrapper { background: #F0EDFF; }
.cat-talent .msg-icon-text { color: #7056D6; }
.cat-cert.msg-icon-wrapper { background: #E9F7F0; }
.cat-cert .msg-icon-text { color: #25915A; }

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
.cell-title-text.is-unread {
  font-weight: 600;
  color: var(--color-text);
}
</style>
