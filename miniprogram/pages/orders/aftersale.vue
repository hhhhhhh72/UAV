<template>
  <view class="aftersale-page">
    <u-nav-bar title="售后详情" show-back right-text="客服" @back="goBack" @right="openCustomerService" />

    <view v-if="loading" class="state-panel">
      <view class="loading-inline">
        <u-loading size="24rpx" />
        <text>加载中...</text>
      </view>
    </view>

    <view v-else-if="error || !order" class="state-panel">
      <u-empty description="售后信息加载失败" />
      <view class="retry-btn" @tap="loadData">
        <text>重新加载</text>
      </view>
    </view>

    <template v-else>
      <!-- 服务概要 -->
      <view class="service-card">
        <text class="svc-kind" :class="kindClass(order.type)">{{ order.kind_label }}</text>
        <text class="svc-title">{{ order.title }}</text>
        <text class="svc-sub">{{ order.subtitle }}</text>
      </view>

      <!-- 售后信息 -->
      <view v-if="as" class="info-card">
        <text class="info-title">售后信息</text>
        <view class="data-row">
          <text class="data-label">售后类型</text>
          <text class="data-value">{{ as.type }}</text>
        </view>
        <view class="data-row">
          <text class="data-label">申请状态</text>
          <text class="data-value wait">{{ as.status }}</text>
        </view>
        <view class="data-row">
          <text class="data-label">退款金额</text>
          <text class="data-value accent">¥{{ fmtFen(as.amount_fen) }}</text>
        </view>
        <view class="data-row">
          <text class="data-label">申请时间</text>
          <text class="data-value">{{ as.created_at }}</text>
        </view>
      </view>

      <!-- 售后进度 -->
      <view v-if="as" class="progress-card">
        <text class="info-title">售后进度</text>
        <view v-for="(p, i) in as.progress" :key="i" class="progress-row">
          <view class="progress-line">
            <view class="progress-dot" :class="{ current: i === 0 }" />
            <view v-if="i < as.progress.length - 1" class="progress-stem" />
          </view>
          <view class="progress-copy">
            <text class="progress-time">{{ p.time }}</text>
            <text class="progress-text">{{ p.text }}</text>
          </view>
        </view>
      </view>

      <!-- 问题说明 -->
      <view v-if="as" class="desc-card">
        <text class="info-title">问题说明</text>
        <text class="desc-text">{{ as.description }}</text>
      </view>

      <!-- 操作 -->
      <!-- 卖家视角：售后待审核时显示审核按钮（同意退款 / 驳回） -->
      <view v-if="isSeller && as && as.status === '待审核'" class="action-wrap">
        <view class="action-btn action-btn--danger" @tap="review(false)">
          <text>驳回申请</text>
        </view>
        <view class="action-btn action-btn--primary" @tap="review(true)">
          <text>同意退款</text>
        </view>
      </view>
      <!-- 买家视角：售后未结案时可补充说明（结案单 approved/rejected 后端拒绝重复申请，不展示入口） -->
      <view v-else-if="!isSeller && order && as && as.status === '待审核'" class="action-wrap">
        <view class="action-btn" @tap="goRefundApply">
          <text>补充说明</text>
        </view>
      </view>
    </template>

    <view class="bottom-spacer"></view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { loadOrder, fmtFen, toastCustomerService } from '../../utils/orderAdapter'
import { request } from '../../utils/request'

const order = ref(null)
const loading = ref(true)
const error = ref(false)

const as = computed(() => order.value?.aftersale || null)

// 卖家视角：订单角色为 seller（orderAdapter 按 seller_id 与当前用户匹配）
const isSeller = computed(() => !!(order.value && order.value.role === 'seller'))

// 卖家审核售后：同意退款 / 驳回
const reviewing = ref(false)
const review = async (approve) => {
  const o = order.value
  if (!o || reviewing.value) return
  const action = approve ? 'approve' : 'reject'
  const tip = approve ? '确认同意退款？' : '确认驳回该售后申请？'
  const confirmed = await new Promise((resolve) => {
    uni.showModal({
      title: tip,
      content: approve ? '同意后售后单结案，订单完成' : '驳回后订单结案，买家可联系客服',
      success: (r) => resolve(!!r.confirm),
      fail: () => resolve(false),
    })
  })
  if (!confirmed) return
  reviewing.value = true
  uni.showLoading({ title: '处理中...' })
  try {
    await request({
      url: '/api/v1/trade-orders/' + encodeURIComponent(o.id) + '/aftersale/review',
      method: 'POST',
      data: { action },
    })
    uni.hideLoading()
    uni.showToast({ title: approve ? '已同意退款' : '已驳回申请', icon: 'success' })
    setTimeout(() => { loadData({ id: o.id }) }, 600)
  } catch (e) {
    uni.hideLoading()
    const msg = (e && e.data && e.data.error && e.data.error.message) || '操作失败，请重试'
    uni.showToast({ title: msg, icon: 'none' })
  } finally {
    reviewing.value = false
  }
}

const loadData = async (query = {}) => {
  const id = query.id
  if (!id) {
    error.value = true
    loading.value = false
    return
  }
  loading.value = true
  error.value = false
  try {
    const o = await loadOrder(id)
    order.value = o
    if (!o || !o.aftersale) error.value = true
  } catch (e) {
    error.value = true
  } finally {
    loading.value = false
  }
}

onLoad(loadData)

const goRefundApply = () => {
  if (!order.value) return
  uni.navigateTo({
    url: `/pages/orders/refund-apply?id=${encodeURIComponent(order.value.id)}&type=${order.value.type}`,
  })
}

const goBack = () => {
  uni.navigateBack()
}

const openCustomerService = () => {
  toastCustomerService()
}

const kindClass = (type) => (type === 'service' ? 'service' : type === 'course' ? 'course' : '')
</script>

<style scoped>
.aftersale-page {
  min-height: 100vh;
  background: var(--color-bg);
  padding-bottom: calc(40rpx + env(safe-area-inset-bottom));
}
.state-panel {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 120rpx;
}
.loading-inline {
  display: flex;
  align-items: center;
  gap: 12rpx;
  font-size: 26rpx;
  color: var(--color-text-secondary);
}
.retry-btn {
  margin-top: 12rpx;
  padding: 16rpx 48rpx;
  background: var(--color-primary);
  color: #fff;
  border-radius: 12rpx;
  font-size: 26rpx;
}

.service-card {
  margin: 20rpx 24rpx 0;
  background: #fff;
  border: 1rpx solid var(--color-border);
  border-radius: 16rpx;
  box-shadow: var(--shadow-sm);
  padding: 28rpx;
}
.svc-kind {
  display: inline-block;
  padding: 4rpx 16rpx;
  border-radius: 8rpx;
  background: var(--color-accent-light);
  color: var(--color-accent-deep);
  font-size: 20rpx;
  font-weight: 600;
}
.svc-kind.course { background: #F0EDFF; color: #7B61D1; }
.svc-title {
  display: block;
  margin-top: 14rpx;
  font-size: 28rpx;
  font-weight: 700;
  color: var(--color-text);
  line-height: 1.4;
}
.svc-sub {
  display: block;
  margin-top: 10rpx;
  font-size: 24rpx;
  color: var(--color-text-secondary);
}

.info-card,
.progress-card,
.desc-card {
  margin: 20rpx 24rpx 0;
  background: #fff;
  border: 1rpx solid var(--color-border);
  border-radius: 16rpx;
  box-shadow: var(--shadow-sm);
  padding: 28rpx;
}
.info-title {
  display: block;
  font-size: 28rpx;
  font-weight: 700;
  color: var(--color-text);
  margin-bottom: 16rpx;
}
.data-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24rpx;
  min-height: 76rpx;
  border-top: 1rpx solid var(--color-divider);
}
.data-row:first-of-type { border-top: none; }
.data-label { font-size: 24rpx; color: var(--color-text-secondary); flex-shrink: 0; }
.data-value { font-size: 24rpx; color: var(--color-text); text-align: right; }
.data-value.wait { color: var(--color-warning); }
.data-value.accent { color: var(--color-accent-deep); font-weight: 700; }

.progress-row { display: flex; gap: 20rpx; }
.progress-line {
  width: 32rpx;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  position: relative;
}
.progress-dot {
  width: 20rpx;
  height: 20rpx;
  border-radius: 50%;
  background: var(--color-text-placeholder);
  margin-top: 8rpx;
  flex-shrink: 0;
}
.progress-dot.current {
  background: var(--color-primary);
  box-shadow: 0 0 0 6rpx var(--color-primary-light);
}
.progress-stem {
  width: 4rpx;
  flex: 1;
  background: var(--color-divider);
  margin: 8rpx 0;
}
.progress-copy {
  flex: 1;
  min-width: 0;
  padding-bottom: 28rpx;
}
.progress-time {
  display: block;
  font-size: 24rpx;
  font-weight: 600;
  color: var(--color-text);
}
.progress-text {
  display: block;
  margin-top: 6rpx;
  font-size: 22rpx;
  color: var(--color-text-secondary);
  line-height: 1.5;
}

.desc-text {
  display: block;
  font-size: 24rpx;
  color: var(--color-text-secondary);
  line-height: 1.7;
}

.action-wrap {
  margin: 32rpx 24rpx 0;
  display: flex;
  gap: 16rpx;
}
.action-btn {
  flex: 1;
  height: 92rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12rpx;
  background: var(--color-primary);
  color: #fff;
  font-size: 30rpx;
  font-weight: 700;
  box-shadow: 0 8rpx 20rpx rgba(10, 102, 194, 0.22);
}
.action-btn--primary {
  background: #0A66C2;
  box-shadow: 0 8rpx 20rpx rgba(10, 102, 194, 0.22);
}
.action-btn--danger {
  background: #ffffff;
  color: #B42318;
  border: 1rpx solid #FDECEC;
  box-shadow: none;
}

.bottom-spacer { height: 24rpx; }
</style>
