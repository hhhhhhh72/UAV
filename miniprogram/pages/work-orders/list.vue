<template>
  <view class="wo-page">
    <!-- 视角 tab -->
    <view class="wo-tabs">
      <view class="wo-tab" :class="{ on: role === 'pub' }" @tap="role = 'pub'">我发出的</view>
      <view class="wo-tab" :class="{ on: role === 'worker' }" @tap="role = 'worker'">我接到的</view>
    </view>

    <view v-if="loading" class="wo-state">加载中...</view>

    <view v-else-if="roleList.length === 0" class="wo-empty">
      <view class="wo-empty-mark">!</view>
      <text class="wo-empty-title">{{ role === 'pub' ? '暂无发出的工单' : '暂无接到的工单' }}</text>
      <text class="wo-empty-desc">{{ role === 'pub' ? '在需求详情「确认接单」后生成工单' : '申请接单被确认后，工单会出现在这里' }}</text>
    </view>

    <view v-else class="wo-list">
      <view v-for="w in roleList" :key="w.id" class="wo-card" hover-class="wo-card-hover" @tap="goDetail(w)">
        <view class="wo-head">
          <text class="wo-no">{{ w.order_no || shortId(w.id) }}</text>
          <text class="wo-status" :class="'wo-status--' + w.status">{{ statusLabel(w.status) }}</text>
        </view>
        <text class="wo-title">{{ w.demand_title || w.demand_id || '作业需求' }}</text>
        <view class="wo-meta">
          <text class="wo-side">{{ role === 'pub' ? '承接方' : '需求方' }}：{{ role === 'pub' ? w.worker_name : w.publisher_name }}</text>
          <text class="wo-amount">{{ amountText(w) }}</text>
        </view>
        <text class="wo-time">{{ formatTime(w.created_at) }}</text>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { request, getStoredUser, requireLogin } from '../../utils/request'

// 默认视角按身份：企业→我发出的；个人/飞手→我接到的（避免默认空 tab）
const u0 = getStoredUser()
const role = ref((u0 && (u0.role === 'enterprise' || u0.user_type === 'enterprise')) ? 'pub' : 'worker')
const loading = ref(true)
const orders = ref([])

const shortId = (id) => (id || '').length > 10 ? (id || '').slice(-8) : (id || '-')
const statusLabel = (s) => ({ pending: '待开始', ongoing: '进行中', awaiting_accept: '待验收', completed: '已完成', cancelled: '已取消' }[s] || s || '')
const amountText = (w) => (w.amount_fen && w.amount_fen > 0 ? '¥' + (w.amount_fen / 100).toLocaleString() : '金额面议')
const formatTime = (iso) => (iso || '').slice(0, 16).replace('T', ' ')

const user = getStoredUser()
const myId = user && (user.id || user.user_id)

const roleList = computed(() => {
  if (!myId) return []
  return orders.value.filter((w) => (role.value === 'pub' ? w.publisher_id === myId : w.worker_id === myId))
})

const fetchAll = async () => {
  loading.value = true
  try {
    const res = await request({ url: '/api/v1/work-orders/mine' })
    orders.value = Array.isArray(res) ? res : ((res && res.data) || [])
  } catch (e) {
    orders.value = []
  } finally {
    loading.value = false
  }
}

const goDetail = (w) => uni.navigateTo({ url: '/pages/work-orders/detail?id=' + encodeURIComponent(w.id) })

onShow(() => {
  if (requireLogin('请先登录后查看工单', '/pages/home/index')) fetchAll()
})
</script>

<style>
page { background: var(--color-bg); }
.wo-page { min-height: 100vh; }
.wo-tabs { display: flex; background: #fff; border-bottom: 1rpx solid #F0F2F5; position: sticky; top: 0; z-index: 5; }
.wo-tab { flex: 1; text-align: center; font-size: 28rpx; color: #667085; padding: 26rpx 0; position: relative; }
.wo-tab.on { color: #0A66C2; font-weight: 600; }
.wo-tab.on::after { content: ''; position: absolute; left: 50%; bottom: 0; transform: translateX(-50%); width: 48rpx; height: 6rpx; border-radius: 3rpx; background: #074D92; }
.wo-state { text-align: center; padding: 120rpx 0; color: #98A2B3; font-size: 26rpx; }
.wo-empty { text-align: center; padding: 100rpx 40rpx 0; }
.wo-empty-mark { width: 96rpx; height: 96rpx; border-radius: 50%; background: #F4F6F8; color: #98A2B3; font-size: 44rpx; display: flex; align-items: center; justify-content: center; margin: 0 auto 24rpx; }
.wo-empty-title { display: block; font-size: 30rpx; font-weight: 600; color: #17212B; }
.wo-empty-desc { display: block; font-size: 24rpx; color: #98A2B3; margin-top: 10rpx; }
.wo-list { padding: 20rpx 24rpx; }
.wo-card { background: #fff; border: 1rpx solid #E4E7EC; border-radius: 10px; padding: 24rpx; margin-bottom: 20rpx; box-shadow: 0 4px 20px rgba(16,24,40,.06); }
.wo-card-hover { opacity: .9; }
.wo-head { display: flex; justify-content: space-between; align-items: center; }
.wo-no { font-size: 22rpx; color: #98A2B3; }
.wo-status { font-size: 22rpx; font-weight: 600; padding: 6rpx 16rpx; border-radius: 6rpx; }
.wo-status--pending { background: #FFF4E5; color: #B54708; }
.wo-status--ongoing { background: #EAF3FB; color: #0A66C2; }
.wo-status--awaiting_accept { background: #FFF4E5; color: #B54708; }
.wo-status--completed { background: #E9F7F0; color: #0B6B41; }
.wo-status--cancelled { background: #FEF3F2; color: #D92D20; }
.wo-title { display: block; font-size: 30rpx; font-weight: 600; color: #17212B; margin-top: 12rpx; }
.wo-meta { display: flex; align-items: center; justify-content: space-between; margin-top: 12rpx; }
.wo-side { font-size: 24rpx; color: #475467; }
.wo-amount { font-size: 28rpx; font-weight: 700; color: #E96012; }
.wo-time { display: block; font-size: 22rpx; color: #98A2B3; margin-top: 8rpx; }
</style>
