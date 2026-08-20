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

      <!-- 申请售后（买家：已发货/已完成且无售后记录时） -->
      <view v-if="showAftersaleEntry" class="aftersale-entry" hover-class="aftersale-entry--active" @tap="goRefundApply">
        <view class="aftersale-entry-copy">
          <text class="aftersale-entry-title">申请售后</text>
          <text class="aftersale-entry-hint">商品有问题可申请退款，由平台审核处理</text>
        </view>
        <text class="aftersale-entry-arrow">›</text>
      </view>
    </template>

    <view class="bottom-spacer"></view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { loadOrder, fmtFen, toastCustomerService } from '../../utils/orderAdapter'

const order = ref(null)
const loading = ref(true)
const error = ref(false)
let orderId = ''

const navTitle = computed(() => {
  if (!order.value) return '订单详情'
  if (order.value.type === 'course') return '培训订单详情'
  if (order.value.type === 'service') return '服务订单详情'
  return '商品订单详情'
})

// 申请售后入口：仅买家、已付款（未发货退款）/已发货/已完成且从未申请过售后（aftersale 记录存在时不再显示）
const showAftersaleEntry = computed(() => {
  const o = order.value
  if (!o || o.role === 'seller') return false
  if (o.aftersale) return false
  return o.status === 'paid' || o.status === 'shipped' || o.status === 'completed'
})

const actionNote = computed(() => {
  if (!order.value) return ''
  const o = order.value
  // 卖家视角：发货方文案
  if (o.role === 'seller') {
    if (o.status === 'pending') return '等待买家完成付款'
    if (o.status === 'paid') return '买家已付款，发货功能即将上线'
    if (o.status === 'shipped') return '等待买家确认收货'
    if (o.status === 'completed') return '交易已完成，感谢你的销售'
    return ''
  }
  // 买家视角：付款方文案
  if (o.aftersale) return '退款由平台审核，可在售后详情查看处理进度'
  if (o.status === 'pending') return '支付功能即将上线，敬请期待'
  if (o.status === 'paid') return '订单已支付，请留意卖家发货安排'
  if (o.status === 'shipped') return '订单已发货，确认收货功能即将上线'
  if (o.status === 'completed') return o.action === '已评价' ? '感谢你的评价，结课凭证已存入「我的报名 / 证书」' : '评价后结课凭证进入「我的报名 / 证书」'
  return ''
})

const loadData = async (query = {}) => {
  const id = query.id || orderId
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

onLoad((options) => {
  orderId = (options && options.id) || ''
  loadData(options)
})

// onShow 刷新：从支付/发货/收货/评价等子页返回后重载订单状态（订单详情为低频页，
// 重复请求可接受；onLoad 先执行完成首屏加载，onShow 只负责后续返回时的刷新）
onShow(() => {
  if (orderId) loadData({ id: orderId })
})

const handlePrimaryAction = () => {
  if (!order.value) return
  const o = order.value
  // 有售后记录的订单（含结案单：状态已回 completed）一律看售后详情
  if (o.aftersale) {
    uni.navigateTo({
      url: `/pages/orders/aftersale?id=${encodeURIComponent(o.id)}&type=${o.type}`,
    })
    return
  }
  // 已取消订单：无任何可操作项
  if (o.status === 'cancelled') {
    uni.showToast({ title: '订单已取消', icon: 'none' })
    return
  }
  // ── 卖家视角：交易闭环下线，发货状态机不开放，仅提示 ──
  if (o.role === 'seller') {
    const sellerTips = {
      pending: '等待买家完成付款',
      paid: '发货功能即将上线',
      shipped: '等待买家确认收货',
      completed: '交易已完成',
      aftersale: '售后处理中',
    }
    uni.showToast({ title: sellerTips[o.status] || '请等待买家操作', icon: 'none' })
    return
  }
  // ── 买家视角：交易闭环下线，支付/收货状态机不开放，仅提示 ──
  if (o.status === 'pending') {
    // 支付功能未接入：不调用后端支付接口
    uni.showToast({ title: '支付功能即将上线', icon: 'none' })
    return
  }
  if (o.status === 'paid') {
    // 提醒发货：无后端调用，仅中性提示，不制造"已发送提醒"的假动作
    uni.showToast({ title: '请耐心等待卖家发货', icon: 'none' })
    return
  }
  if (o.status === 'shipped') {
    // 确认收货未开放：不调用后端状态机
    uni.showToast({ title: '等待确认收货', icon: 'none' })
    return
  }
  if (o.status === 'completed') {
    // 去评价（结案单已在上方 o.aftersale 分支处理，不会走到这里）
    uni.navigateTo({
      url: `/pages/orders/review?id=${encodeURIComponent(o.id)}&type=${o.type}`,
    })
    return
  }
}

// 申请售后：进入退款申请页（提交后回这里/售后详情看进度）
const goRefundApply = () => {
  if (!order.value) return
  uni.navigateTo({
    url: `/pages/orders/refund-apply?id=${encodeURIComponent(order.value.id)}&type=${order.value.type}`,
  })
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

/* 申请售后入口 */
.aftersale-entry {
  margin: 20rpx 24rpx 0;
  background: #fff;
  border: 1rpx solid var(--color-border);
  border-radius: 16rpx;
  box-shadow: var(--shadow-sm);
  min-height: 104rpx;
  padding: 0 28rpx;
  display: flex;
  align-items: center;
  gap: 20rpx;
  box-sizing: border-box;
}
.aftersale-entry--active { opacity: 0.7; }
.aftersale-entry-copy { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 6rpx; }
.aftersale-entry-title { font-size: 26rpx; font-weight: 600; color: var(--color-text); }
.aftersale-entry-hint { font-size: 22rpx; color: var(--color-text-placeholder); }
.aftersale-entry-arrow { font-size: 30rpx; color: var(--color-text-placeholder); }

.bottom-spacer { height: 24rpx; }
</style>
