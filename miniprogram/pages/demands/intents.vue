<template>
  <view class="intents-page">
    <!-- 头部 -->
    <view class="page-header">
      <view class="back-btn" @tap="goBack"><text class="back-sym">‹</text></view>
      <text class="page-title">收到的对接意向</text>
      <view class="head-spacer"></view>
    </view>

    <!-- 列表标题 -->
    <view class="list-head">
      <text class="list-title">{{ singleDemand ? '需求意向' : '全部意向' }}</text>
      <text class="list-count">共 {{ intents.length }} 条</text>
    </view>

    <!-- 空状态 -->
    <view v-if="intents.length === 0" class="state-panel">
      <view class="state-mark">⌁</view>
      <text class="state-title">{{ loadError ? '加载失败' : '暂无对接意向' }}</text>
      <text class="state-desc">{{ loadError ? '网络异常，请稍后重试' : '发布的需求收到登记后，会第一时间展示在这里' }}</text>
      <view v-if="loadError" class="state-btn" @tap="fetchIntents">重新加载</view>
    </view>

    <!-- 意向列表 -->
    <view v-else class="intent-list">
      <view v-for="intent in intents" :key="intent.id" class="intent-card">
        <view class="intent-head">
          <view class="intent-avatar"><text>{{ initialOf(intent) }}</text></view>
          <view class="intent-copy">
            <text class="intent-name">{{ intent.intentor_name }}</text>
            <text class="tag" :class="intentStatusClass(intent.status)">{{ intentStatusLabel(intent.status) }}</text>
          </view>
        </view>
        <text v-if="!singleDemand" class="intent-detail">对接项目：{{ intent.demand_title }}</text>
        <text v-if="intent.contact" class="intent-detail">联系方式：{{ intent.contact }}</text>
        <text class="intent-note">对方说明：{{ intent.remark || '未填写说明' }}</text>
        <view class="intent-actions">
          <template v-if="intent.status === 'pending'">
            <view class="intent-btn" @tap="rejectIntent(intent)">拒绝</view>
            <view class="intent-btn primary" @tap="openAccept(intent)">确认接单</view>
          </template>
          <template v-else-if="intent.status === 'contacted'">
            <view class="intent-btn primary" @tap="goOrders">查看订单</view>
          </template>
          <template v-else>
            <view class="intent-btn" @tap="toastClosed">意向已关闭</view>
          </template>
        </view>
      </view>
    </view>

    <!-- 确认接单：填写金额 -->
    <u-popup :show="acceptShow" position="bottom" round @close="acceptShow = false">
      <view class="sheet">
        <view class="sheet-head">
          <text class="sheet-title">确认接单</text>
          <view class="sheet-close" @tap="acceptShow = false"><text class="sheet-x">×</text></view>
        </view>
        <view class="sheet-body">
          <text class="sheet-desc">确认后将为「{{ acceptTarget ? acceptTarget.intentor_name : '' }}」生成作业订单，订单金额如下（可填 0 表示面议）。</text>
          <view class="amount-input">
            <text class="amount-symbol">¥</text>
            <input
              class="amount-field"
              type="digit"
              v-model="acceptAmount"
              placeholder="输入订单金额，0 为面议"
              placeholder-class="amount-ph"
            />
          </view>
          <view class="sheet-btn primary" :class="{ disabled: acceptSubmitting }" @tap="submitAccept">确认接单</view>
        </view>
      </view>
    </u-popup>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { safeNavigateTo } from '../../utils/nav'
import { request, getErrorMessage } from '../../utils/request'

const intents = ref([])
const loadError = ref(false)
const singleDemand = ref('')
const acceptShow = ref(false)
const acceptTarget = ref(null)
const acceptAmount = ref('')
const acceptSubmitting = ref(false)

const STATUS_LABEL = { pending: '待处理', contacted: '已确认', closed: '已关闭' }
const intentStatusLabel = (s) => STATUS_LABEL[s] || s || ''
const intentStatusClass = (s) => (s === 'contacted' ? 'green' : s === 'closed' ? 'gray' : 'orange')
const initialOf = (it) => (it.intentor_name || '?').slice(0, 1)

// 拉取我的需求（mine=1），再逐个取意向
const fetchIntents = async () => {
  loadError.value = false
  try {
    const demandRes = await request({ url: '/api/v1/demands?mine=1&page_size=100' })
    const demands = Array.isArray(demandRes) ? demandRes : (demandRes && demandRes.data) || []
    const all = []
    await Promise.all(
      demands.map(async (d) => {
        try {
          const res = await request({ url: '/api/v1/demands/' + encodeURIComponent(d.id) + '/intents' })
          const its = Array.isArray(res) ? res : (res && res.data) || []
          its.forEach((it) => all.push(Object.assign({}, it, { demand_title: d.title })))
        } catch { /* 单个需求失败不阻塞聚合 */ }
      })
    )
    intents.value = all
  } catch {
    loadError.value = true
    intents.value = []
  }
}

const fetchSingle = async (demandId) => {
  loadError.value = false
  try {
    const res = await request({ url: '/api/v1/demands/' + encodeURIComponent(demandId) + '/intents' })
    const its = Array.isArray(res) ? res : (res && res.data) || []
    intents.value = its
  } catch {
    loadError.value = true
    intents.value = []
  }
}

onLoad((options) => {
  if (options && options.demandId) {
    singleDemand.value = options.demandId
    fetchSingle(options.demandId)
  } else {
    fetchIntents()
  }
})

function openAccept(intent) {
  acceptTarget.value = intent
  acceptAmount.value = ''
  acceptShow.value = true
}

async function submitAccept() {
  if (acceptSubmitting.value) return
  const intent = acceptTarget.value
  if (!intent) return
  const yuan = parseFloat(acceptAmount.value)
  const amountFen = isNaN(yuan) ? 0 : Math.round(yuan * 100)
  acceptSubmitting.value = true
  try {
    await request({
      url: '/api/v1/demands/' + encodeURIComponent(intent.demand_id) + '/intents/' + encodeURIComponent(intent.id) + '/accept',
      method: 'POST',
      data: { amount_fen: amountFen },
    })
    acceptShow.value = false
    intent.status = 'contacted'
    uni.showToast({ title: '已确认接单，订单已生成', icon: 'success' })
  } catch (e) {
    uni.showToast({ title: getErrorMessage(e) || '操作失败，请重试', icon: 'none' })
  } finally {
    acceptSubmitting.value = false
  }
}

function rejectIntent(intent) {
  uni.showModal({
    title: '拒绝意向',
    content: '确定拒绝该意向？拒绝后不可恢复。',
    success: async (r) => {
      if (!r.confirm) return
      try {
        await request({
          url: '/api/v1/demands/' + encodeURIComponent(intent.demand_id) + '/intents/' + encodeURIComponent(intent.id) + '/reject',
          method: 'POST',
        })
        intent.status = 'closed'
        uni.showToast({ title: '已拒绝该意向', icon: 'none' })
      } catch (e) {
        uni.showToast({ title: getErrorMessage(e) || '操作失败，请重试', icon: 'none' })
      }
    },
  })
}

const toastClosed = () => uni.showToast({ title: '该意向已关闭', icon: 'none' })
const goOrders = () => safeNavigateTo('/pages/orders/mine')
const goBack = () => uni.navigateBack()
</script>

<style scoped>
.intents-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: 40rpx;
}

.page-header {
  height: 56px;
  padding: 0 28rpx;
  display: flex;
  align-items: center;
  gap: 8rpx;
  background: #fff;
  border-bottom: 1px solid #EEF1F4;
  position: sticky;
  top: 0;
  z-index: 10;
}
.back-btn { width: 72rpx; height: 72rpx; display: flex; align-items: center; justify-content: center; }
.back-sym { font-size: 52rpx; color: #17212B; line-height: 1; }
.page-title { flex: 1; font-size: 34rpx; font-weight: 700; color: #17212B; }
.head-spacer { width: 72rpx; }

.list-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  padding: 28rpx 32rpx 16rpx;
}
.list-title { font-size: 36rpx; font-weight: 750; color: #17212B; }
.list-count { font-size: 24rpx; color: #667085; }

.intent-list { padding: 0 32rpx 32rpx; }
.intent-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 26rpx;
  border: 1px solid #EEF1F4;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.045);
}
.intent-card + .intent-card { margin-top: 20rpx; }

.intent-head { display: flex; gap: 20rpx; align-items: center; }
.intent-avatar {
  width: 76rpx;
  height: 76rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #EAF3FB;
  color: #0A66C2;
  font-size: 28rpx;
  font-weight: 750;
  flex-shrink: 0;
}
.intent-copy { display: flex; flex-direction: column; gap: 8rpx; }
.intent-name { font-size: 26rpx; font-weight: 700; color: #17212B; }
.tag {
  align-self: flex-start;
  border-radius: 8rpx;
  padding: 6rpx 12rpx;
  font-size: 20rpx;
  line-height: 1;
}
.tag.orange { color: #DB5F0D; background: #FFF0E6; }
.tag.green { color: #168A55; background: #E9F7F0; }
.tag.gray { color: #667085; background: #F1F3F5; }

.intent-detail { display: block; font-size: 24rpx; color: #17212B; margin-top: 18rpx; }
.intent-note { display: block; font-size: 22rpx; color: #667085; line-height: 1.6; margin-top: 8rpx; }

.intent-actions {
  display: flex;
  gap: 14rpx;
  border-top: 1px solid #EEF1F4;
  margin-top: 22rpx;
  padding-top: 20rpx;
}
.intent-btn {
  flex: 1;
  height: 62rpx;
  border-radius: 10rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24rpx;
  color: #0A66C2;
  border: 1px solid #C7DEF1;
  background: #fff;
}
.intent-btn.primary { color: #fff; background: #0A66C2; border-color: #0A66C2; }

/* 确认接单弹层 */
.sheet { padding: 32rpx 32rpx calc(32rpx + env(safe-area-inset-bottom)); }
.sheet-head { display: flex; align-items: center; }
.sheet-title { flex: 1; font-size: 32rpx; font-weight: 750; color: #17212B; }
.sheet-close { padding: 8rpx; }
.sheet-x { font-size: 36rpx; color: #98A2B3; line-height: 1; }
.sheet-body { margin-top: 24rpx; }
.sheet-desc { font-size: 24rpx; color: #667085; line-height: 1.6; }
.amount-input {
  display: flex;
  align-items: center;
  gap: 12rpx;
  height: 96rpx;
  margin: 28rpx 0;
  padding: 0 24rpx;
  border: 1px solid #E4E7EC;
  border-radius: 16rpx;
  background: #FAFAFA;
}
.amount-symbol { font-size: 36rpx; font-weight: 700; color: #17212B; }
.amount-field { flex: 1; font-size: 32rpx; color: #17212B; }
.amount-ph { color: #98A2B3; }
.sheet-btn {
  height: 88rpx;
  border-radius: 50rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  font-weight: 650;
  background: #0A66C2;
  color: #fff;
  box-shadow: 0 6px 16px rgba(10, 102, 194, 0.28);
}
.sheet-btn.disabled { opacity: 0.6; }

.state-panel {
  min-height: 560rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 56rpx;
  text-align: center;
}
.state-mark {
  width: 124rpx;
  height: 124rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24rpx;
  border-radius: 50%;
  background: #EAF3FB;
  color: #0A66C2;
  font-size: 54rpx;
}
.state-title { font-size: 28rpx; font-weight: 700; color: #17212B; }
.state-desc { margin: 12rpx 0 0; font-size: 22rpx; color: #98A2B3; }
.state-btn {
  margin-top: 32rpx;
  height: 72rpx;
  padding: 0 30rpx;
  border-radius: 12rpx;
  background: #0A66C2;
  color: #fff;
  font-size: 24rpx;
  line-height: 72rpx;
}
</style>
