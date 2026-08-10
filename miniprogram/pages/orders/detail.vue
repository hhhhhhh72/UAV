<template>
  <view class="order-detail-page">
    <u-nav-bar :title="navTitle" show-back right-text="客服" @back="goBack" @right="openCustomerService" />

    <view v-if="loading" class="state-panel">
      <view class="loading-inline">
        <u-loading size="24rpx" />
        <text>加载中...</text>
      </view>
    </view>

    <view v-else-if="error || !order" class="state-panel">
      <u-empty description="订单加载失败" />
      <view class="retry-btn" @tap="loadData">
        <text>重新加载</text>
      </view>
    </view>

    <template v-else>
      <!-- 头部商品卡 -->
      <view class="detail-hero" v-if="order.detail && order.detail.hero">
        <text class="hero-kind" :class="kindClass(order.type)">{{ order.kind_label }}</text>
        <text class="hero-title">{{ order.detail.hero.title }}</text>
        <text class="hero-sub">{{ order.detail.hero.sub }}</text>
      </view>

      <!-- 媒体区 + 名称价格（无 hero 时） -->
      <view v-else class="detail-media-card">
        <image v-if="order.image" class="detail-media" :src="order.image" mode="aspectFill" />
        <view v-else class="detail-media detail-media-ph">
          <text>{{ kindMark(order.kind_label) }}</text>
        </view>
        <view class="media-body">
          <text class="media-kind" :class="kindClass(order.type)">{{ order.kind_label }}</text>
          <text class="media-title">{{ order.title }}</text>
          <view class="media-price-row">
            <text class="media-price">¥{{ fmtFen(order.amount_fen) }}</text>
            <text class="media-qty">{{ order.quantity_label }}</text>
          </view>
          <text class="media-sub">{{ order.subtitle }}</text>
        </view>
      </view>

      <!-- 分区明细 -->
      <view
        v-for="(section, i) in order.detail.sections || []"
        :key="i"
        class="detail-card"
      >
        <text class="detail-card-title">{{ section.title }}</text>
        <view v-for="(row, ri) in section.rows" :key="ri" class="data-row">
          <text class="data-label">{{ row.label }}</text>
          <text class="data-value" :class="rowClass(row.status)">{{ row.value }}</text>
        </view>
      </view>

      <!-- 物流入口（待收货） -->
      <view v-if="order.detail && order.detail.logistics" class="detail-list">
        <view class="detail-list-item" @tap="goLogistics">
          <view class="list-mark">物</view>
          <view class="list-copy">
            <text class="list-title">查看物流详情</text>
            <text class="list-desc">查看完整运输节点和配送进度</text>
          </view>
          <text class="list-arrow">›</text>
        </view>
      </view>

      <!-- 主操作 -->
      <view class="primary-action">
        <view class="primary-btn" @tap="handlePrimaryAction">
          <text>{{ order.action }}</text>
        </view>
        <text class="action-note" v-if="actionNote">{{ actionNote }}</text>
      </view>
    </template>

    <view class="bottom-spacer"></view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { loadOrder, fmtFen, toastCustomerService } from '../../utils/orderAdapter'

const order = ref(null)
const loading = ref(true)
const error = ref(false)

const navTitle = computed(() => {
  if (!order.value) return '订单详情'
  if (order.value.type === 'course') return '培训订单详情'
  if (order.value.type === 'service') return '服务订单详情'
  return '商品订单详情'
})

const actionNote = computed(() => {
  if (!order.value) return ''
  const o = order.value
  if (o.status === 'pending') return '支付接口接入中，暂不修改订单状态'
  if (o.status === 'paid') return '发货提醒已记录，前端不改变订单状态'
  if (o.status === 'shipped') return '确认收货需后端支持后方可生效'
  if (o.status === 'completed') return '评价后结课凭证进入「我的报名 / 证书」'
  if (o.status === 'aftersale') return '退款由平台审核，前端仅提交申请'
  return ''
})

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
    order.value = await loadOrder(id)
    if (!order.value) error.value = true
  } catch (e) {
    error.value = true
  } finally {
    loading.value = false
  }
}

onLoad(loadData)

const handlePrimaryAction = () => {
  if (!order.value) return
  const o = order.value
  if (o.status === 'pending') {
    // 去支付：无真实支付契约，仅提示入口，不把订单改为已支付
    uni.showToast({ title: '支付功能待接入，订单保持待付款', icon: 'none' })
    return
  }
  if (o.status === 'paid') {
    // 提醒发货：发送提醒成功反馈，不把状态改为待收货
    uni.showToast({ title: '已向商家发送发货提醒', icon: 'none' })
    return
  }
  if (o.status === 'shipped') {
    // 确认收货：需要后端支持后才可改变状态
    uni.showModal({
      title: '确认收货',
      content: '确认收货功能需后端支持后开放，当前订单状态不会改变。',
      showCancel: false,
      confirmText: '知道了',
    })
    return
  }
  if (o.status === 'completed') {
    // 去评价
    uni.navigateTo({
      url: `/pages/orders/review?id=${encodeURIComponent(o.id)}&type=${o.type}`,
    })
    return
  }
  if (o.status === 'aftersale') {
    // 查看售后
    uni.navigateTo({
      url: `/pages/orders/aftersale?id=${encodeURIComponent(o.id)}&type=${o.type}`,
    })
  }
}

const goLogistics = () => {
  if (!order.value) return
  uni.navigateTo({
    url: `/pages/orders/logistics?id=${encodeURIComponent(order.value.id)}&type=${order.value.type}`,
  })
}

const openCustomerService = () => {
  toastCustomerService()
}

const goBack = () => {
  uni.navigateBack()
}

// ── 展示辅助 ──
const kindMark = (label) => (label ? label.charAt(0) : '单')
const kindClass = (type) => (type === 'course' ? 'course' : type === 'service' ? 'service' : '')
const rowClass = (status) => {
  if (status === 'good') return 'good'
  if (status === 'wait') return 'wait'
  return ''
}
</script>

<style scoped>
.order-detail-page {
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

/* hero（演示数据） */
.detail-hero {
  margin: 20rpx 24rpx 0;
  padding: 32rpx;
  color: #fff;
  background: linear-gradient(135deg, var(--color-primary-deep), #126DC8);
  border-radius: 16rpx;
  position: relative;
  overflow: hidden;
}
.detail-hero::after {
  content: '';
  position: absolute;
  right: -60rpx;
  top: -80rpx;
  width: 240rpx;
  height: 240rpx;
  border: 1rpx solid rgba(255,255,255,.18);
  border-radius: 50%;
}
.hero-kind {
  display: inline-block;
  padding: 4rpx 16rpx;
  border-radius: 8rpx;
  background: rgba(255,255,255,.18);
  font-size: 22rpx;
  font-weight: 600;
  position: relative;
  z-index: 1;
}
.hero-title {
  display: block;
  margin-top: 16rpx;
  font-size: 36rpx;
  font-weight: 700;
  line-height: 1.4;
  position: relative;
  z-index: 1;
}
.hero-sub {
  display: block;
  margin-top: 10rpx;
  font-size: 24rpx;
  color: rgba(255,255,255,.76);
  position: relative;
  z-index: 1;
}

/* media（真实商品订单） */
.detail-media-card {
  margin: 20rpx 24rpx 0;
  background: #fff;
  border: 1rpx solid var(--color-border);
  border-radius: 16rpx;
  box-shadow: var(--shadow-sm);
  overflow: hidden;
}
.detail-media {
  width: 100%;
  height: 300rpx;
  background: var(--color-primary-light);
}
.detail-media-ph {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 80rpx;
  font-weight: 700;
  color: var(--color-primary);
}
.media-body { padding: 24rpx; }
.media-kind {
  display: inline-block;
  padding: 4rpx 16rpx;
  border-radius: 8rpx;
  background: var(--color-primary-light);
  color: var(--color-primary);
  font-size: 20rpx;
  font-weight: 600;
}
.media-kind.course { background: #F0EDFF; color: #7B61D1; }
.media-kind.service { background: var(--color-accent-light); color: var(--color-accent-deep); }
.media-title {
  display: block;
  margin-top: 12rpx;
  font-size: 32rpx;
  font-weight: 700;
  color: var(--color-text);
  line-height: 1.4;
}
.media-price-row {
  display: flex;
  align-items: baseline;
  gap: 12rpx;
  margin-top: 12rpx;
}
.media-price {
  font-size: 36rpx;
  font-weight: 800;
  color: var(--color-accent-deep);
}
.media-qty { font-size: 22rpx; color: var(--color-text-secondary); }
.media-sub {
  display: block;
  margin-top: 10rpx;
  font-size: 24rpx;
  color: var(--color-text-secondary);
}

/* 分区卡片 */
.detail-card {
  margin: 20rpx 24rpx 0;
  background: #fff;
  border: 1rpx solid var(--color-border);
  border-radius: 16rpx;
  box-shadow: var(--shadow-sm);
  padding: 28rpx;
}
.detail-card-title {
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
.data-value { font-size: 24rpx; color: var(--color-text); text-align: right; word-break: break-all; }
.data-value.good { color: var(--color-success); }
.data-value.wait { color: var(--color-warning); }

/* 列表入口（物流） */
.detail-list {
  margin: 20rpx 24rpx 0;
  background: #fff;
  border: 1rpx solid var(--color-border);
  border-radius: 16rpx;
  box-shadow: var(--shadow-sm);
  overflow: hidden;
}
.detail-list-item {
  display: flex;
  align-items: center;
  gap: 20rpx;
  min-height: 104rpx;
  padding: 0 28rpx;
  box-sizing: border-box;
}
.list-mark {
  width: 72rpx;
  height: 72rpx;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12rpx;
  background: var(--color-primary-light);
  color: var(--color-primary);
  font-size: 28rpx;
  font-weight: 700;
}
.list-copy { flex: 1; min-width: 0; }
.list-title {
  display: block;
  font-size: 26rpx;
  font-weight: 600;
  color: var(--color-text);
}
.list-desc {
  display: block;
  margin-top: 6rpx;
  font-size: 22rpx;
  color: var(--color-text-placeholder);
}
.list-arrow { font-size: 30rpx; color: var(--color-text-placeholder); }

/* 主操作 */
.primary-action {
  margin: 32rpx 24rpx 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12rpx;
}.primary-btn {
  width: 100%;
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
.action-note {
  font-size: 22rpx;
  color: var(--color-text-placeholder);
  text-align: center;
}

.bottom-spacer { height: 24rpx; }
</style>
