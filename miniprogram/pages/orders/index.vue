<template>
  <view class="orders-page">
    <u-nav-bar title="我的订单" show-back @back="goBack" />

    <!-- 筛选行：订单类型 + 筛选按钮 -->
    <view class="filter-line">
      <view class="filter-line-label">
        <text class="filter-name">订单类型</text>
        <text class="filter-value">：{{ orderTypeLabel }}</text>
      </view>
      <view class="filter-btn" @tap="openFilter">
        <text class="filter-btn-text">筛选</text>
        <text class="filter-btn-arrow">›</text>
      </view>
    </view>

    <!-- 五个状态入口（固定顺序与文案） -->
    <view class="status-summary">
      <view
        v-for="s in statusEntries"
        :key="s.key"
        class="status-cell"
        @tap="goStatusList(s.key)"
      >
        <view class="status-icon-wrap">
          <text class="status-icon">{{ s.icon }}</text>
          <text v-if="statusCounts[s.key]" class="status-badge">{{ statusCounts[s.key] }}</text>
        </view>
        <text class="status-label">{{ s.label }}</text>
      </view>
    </view>

    <!-- 最近订单 -->
    <view class="section-title-row">
      <text class="section-title">最近订单</text>
      <view class="section-all" @tap="goAllOrders">
        <text class="section-all-text">全部订单</text>
        <text class="section-all-arrow">›</text>
      </view>
    </view>

    <!-- 加载态 -->
    <view v-if="loading" class="state-panel">
      <view class="skel-list">
        <view v-for="i in 3" :key="i" class="skel-card">
          <view class="skel-line w40" />
          <view class="skel-body">
            <view class="skel-line w70" />
            <view class="skel-line" />
            <view class="skel-line w50" />
          </view>
        </view>
      </view>
    </view>

    <!-- 失败态 -->
    <view v-else-if="error" class="state-panel">
      <view class="state-empty-wrap">
        <u-empty description="订单加载失败" />
        <view class="retry-btn" @tap="loadData">
          <text>重新加载</text>
        </view>
      </view>
    </view>

    <!-- 空态：当前筛选下暂无订单 -->
    <view v-else-if="orders.length === 0" class="state-panel">
      <u-empty description="当前筛选下暂无订单" />
    </view>

    <!-- 订单列表 -->
    <view v-else class="order-list">
      <view
        v-for="o in orders"
        :key="o.id"
        class="order-card"
        @tap="goDetail(o)"
      >
        <view class="order-head">
          <view class="order-origin">
            <text class="order-origin-mark">{{ kindMark(o.kind_label) }}</text>
            <text class="order-origin-text">{{ o.origin }}</text>
          </view>
          <text class="order-state" :class="stateClass(o.status)">{{ o.status_text || ORDER_STATUS[o.status] || o.status }}</text>
        </view>

        <view class="order-body" @tap.stop="goDetail(o)">
          <image v-if="o.image" class="order-thumb" :src="o.image" mode="aspectFill" />
          <view v-else class="order-thumb order-thumb-ph">
            <text>{{ kindMark(o.kind_label) }}</text>
          </view>
          <view class="order-info">
            <view class="order-info-head">
              <text class="order-kind-badge" :class="kindClass(o.type)">{{ o.kind_label }}</text>
            </view>
            <text class="order-title">{{ o.title }}</text>
            <text class="order-subtitle">{{ o.subtitle }}</text>
            <view class="order-price-row">
              <text class="order-price">¥{{ fmtFen(o.amount_fen) }}</text>
              <text class="order-qty">{{ o.quantity_label }}</text>
            </view>
          </view>
        </view>

        <view class="order-foot">
          <text class="order-due">{{ o.due_text }}</text>
          <view class="order-actions">
            <view class="action-btn" @tap.stop="goDetail(o)">{{ o.action }}</view>
          </view>
        </view>
      </view>
    </view>

    <view class="bottom-spacer"></view>

    <!-- 订单类型筛选抽屉：草稿态 / 已应用态 分离 -->
    <u-popup :show="showFilter" position="bottom" round @close="closeFilter">
      <view class="filter-sheet">
        <view class="sheet-grip" />
        <view class="sheet-head">
          <text class="sheet-title">筛选订单</text>
          <view class="sheet-close" @tap="closeFilter">
            <text class="sheet-x">×</text>
          </view>
        </view>
        <text class="sheet-desc">按订单类型筛选，不会改变订单状态分类。</text>
        <view class="filter-options">
          <view
            v-for="t in typeOptions"
            :key="t.value"
            class="filter-option"
            :class="{ active: draftType === t.value }"
            @tap="pickType(t.value)"
          >
            <text>{{ t.label }}</text>
          </view>
        </view>
        <view class="filter-confirm" @tap="applyFilter">
          <text>应用筛选</text>
        </view>
      </view>
    </u-popup>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { getStoredUser } from '../../utils/request'
import {
  ORDER_STATUS,
  ORDER_TYPE_LABEL,
  loadOrders,
  loadStatusCounts,
  fmtFen,
} from '../../utils/orderAdapter'

const loading = ref(false)
const error = ref(false)
const orders = ref([])
const statusCounts = ref({ pending: 0, paid: 0, shipped: 0, completed: 0, aftersale: 0 })

// 已应用的筛选（决定列表与状态页带入）
const appliedType = ref('all')
// 抽屉草稿筛选
const showFilter = ref(false)
const draftType = ref('all')

const typeOptions = [
  { label: '全部类型', value: 'all' },
  { label: '商品', value: 'product' },
  { label: '培训课程', value: 'course' },
  { label: '无人机服务', value: 'service' },
]

const statusEntries = [
  { key: 'pending', icon: '付', label: '待付款' },
  { key: 'paid', icon: '发', label: '待发货' },
  { key: 'shipped', icon: '收', label: '待收货' },
  { key: 'completed', icon: '评', label: '待评价' },
  { key: 'aftersale', icon: '售', label: '退款/售后' },
]

const orderTypeLabel = computed(() => ORDER_TYPE_LABEL[appliedType.value] || '全部类型')

// 未登录跳登录（受保护业务页）
const ensureLogin = () => {
  if (getStoredUser()) return true
  uni.navigateTo({ url: '/pages/login/index' })
  return false
}

const loadData = async () => {
  if (!ensureLogin()) return
  loading.value = true
  error.value = false
  try {
    const [list, counts] = await Promise.all([
      loadOrders({ status: 'all', order_type: appliedType.value }),
      loadStatusCounts(appliedType.value),
    ])
    orders.value = list
    statusCounts.value = counts
  } catch (e) {
    error.value = true
  } finally {
    loading.value = false
  }
}

onLoad((query) => {
  // 支持从其他入口带 order_type 进入（如评价成功返回）
  if (query && query.order_type && ORDER_TYPE_LABEL[query.order_type]) {
    appliedType.value = query.order_type
  }
})

onShow(() => {
  // 返回本页时刷新并保持已应用筛选
  loadData()
})

const openFilter = () => {
  draftType.value = appliedType.value
  showFilter.value = true
}

const closeFilter = () => {
  showFilter.value = false
}

const pickType = (v) => {
  draftType.value = v
}

const applyFilter = () => {
  appliedType.value = draftType.value
  showFilter.value = false
  loadData()
}

const goStatusList = (status) => {
  if (!ensureLogin()) return
  uni.navigateTo({
    url: `/pages/orders/list?status=${status}&order_type=${appliedType.value}`,
  })
}

const goAllOrders = () => {
  if (!ensureLogin()) return
  uni.navigateTo({
    url: `/pages/orders/list?status=all&order_type=${appliedType.value}`,
  })
}

const goDetail = (o) => {
  if (!ensureLogin()) return
  uni.navigateTo({
    url: `/pages/orders/detail?id=${encodeURIComponent(o.id)}&type=${o.type}`,
  })
}

const goBack = () => {
  uni.navigateBack()
}

// ── 展示辅助 ──
const kindMark = (label) => (label ? label.charAt(0) : '单')
const kindClass = (type) => (type === 'course' ? 'course' : type === 'service' ? 'service' : '')
const stateClass = (status) => (status === 'completed' || status === 'paid' ? 'done' : status === 'aftersale' ? 'service' : status === 'cancelled' ? 'cancelled' : '')
</script>

<style scoped>
.orders-page {
  min-height: 100vh;
  background: var(--color-bg);
  padding-bottom: calc(32rpx + env(safe-area-inset-bottom));
}

/* 筛选行 */
.filter-line {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 24rpx 24rpx 20rpx;
  padding: 0 24rpx;
  height: 76rpx;
  background: #fff;
  border: 1rpx solid var(--color-border);
  border-radius: 12rpx;
  box-shadow: var(--shadow-sm);
  box-sizing: border-box;
}
.filter-line-label {
  display: flex;
  align-items: center;
  gap: 4rpx;
  font-size: 26rpx;
  color: var(--color-text-secondary);
}
.filter-value {
  color: var(--color-primary);
  font-weight: 600;
}
.filter-btn {
  display: flex;
  align-items: center;
  gap: 6rpx;
  padding: 8rpx 0 8rpx 16rpx;
}
.filter-btn-text {
  font-size: 26rpx;
  color: var(--color-primary);
  font-weight: 600;
}
.filter-btn-arrow {
  font-size: 28rpx;
  color: var(--color-text-placeholder);
}

/* 五状态入口：一张白卡 5 等分 */
.status-summary {
  display: flex;
  margin: 0 24rpx 24rpx;
  background: #fff;
  border: 1rpx solid var(--color-border);
  border-radius: 16rpx;
  box-shadow: var(--shadow-sm);
  padding: 16rpx 0;
}
.status-cell {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10rpx;
  min-height: 140rpx;
  position: relative;
}
.status-icon-wrap {
  position: relative;
  width: 60rpx;
  height: 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12rpx;
  background: var(--color-primary-light);
}
.status-icon {
  font-size: 24rpx;
  font-weight: 700;
  color: var(--color-primary);
}
.status-cell:nth-child(4) .status-icon-wrap { background: #F0EDFF; }
.status-cell:nth-child(4) .status-icon { color: #7B61D1; }
.status-cell:nth-child(5) .status-icon-wrap { background: var(--color-accent-light); }
.status-cell:nth-child(5) .status-icon { color: var(--color-accent-deep); }
.status-badge {
  position: absolute;
  top: -8rpx;
  right: -12rpx;
  min-width: 32rpx;
  height: 32rpx;
  padding: 0 8rpx;
  border-radius: 16rpx;
  background: var(--color-accent);
  color: #fff;
  font-size: 20rpx;
  font-weight: 700;
  line-height: 32rpx;
  text-align: center;
  box-sizing: border-box;
}
.status-label {
  font-size: 20rpx;
  color: var(--color-text);
  font-weight: 500;
}

/* 最近订单标题 */
.section-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 28rpx 16rpx;
}
.section-title {
  font-size: 32rpx;
  font-weight: 700;
  color: var(--color-text);
}
.section-all {
  display: flex;
  align-items: center;
  gap: 4rpx;
}
.section-all-text {
  font-size: 24rpx;
  color: var(--color-text-secondary);
}
.section-all-arrow {
  font-size: 26rpx;
  color: var(--color-text-placeholder);
}

/* 订单卡 */
.order-list {
  padding: 0 24rpx;
}
.order-card {
  background: #fff;
  border: 1rpx solid var(--color-border);
  border-radius: 16rpx;
  box-shadow: var(--shadow-sm);
  overflow: hidden;
  margin-bottom: 20rpx;
}
.order-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 76rpx;
  padding: 0 24rpx;
  border-bottom: 1rpx solid var(--color-divider);
  box-sizing: border-box;
}
.order-origin {
  display: flex;
  align-items: center;
  gap: 12rpx;
  min-width: 0;
}
.order-origin-mark {
  width: 36rpx;
  height: 36rpx;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8rpx;
  background: var(--color-primary-light);
  color: var(--color-primary);
  font-size: 22rpx;
  font-weight: 700;
}
.order-origin-text {
  font-size: 24rpx;
  color: var(--color-text);
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.order-state {
  font-size: 24rpx;
  font-weight: 600;
  color: var(--color-accent-deep);
  flex-shrink: 0;
  margin-left: 12rpx;
}
.order-state.done { color: var(--color-success); }
.order-state.service { color: #6B5CC7; }
.order-state.cancelled { color: var(--color-text-placeholder); }

.order-body {
  display: flex;
  gap: 20rpx;
  padding: 20rpx 24rpx;
}
.order-thumb {
  width: 112rpx;
  height: 112rpx;
  flex-shrink: 0;
  border-radius: 12rpx;
  background: var(--color-primary-light);
  overflow: hidden;
}
.order-thumb-ph {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  font-weight: 700;
  color: var(--color-primary);
}
.order-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}
.order-info-head {
  display: flex;
  margin-bottom: 8rpx;
}
.order-kind-badge {
  display: inline-flex;
  align-items: center;
  align-self: flex-start;
  padding: 4rpx 14rpx;
  border-radius: 8rpx;
  background: var(--color-primary-light);
  color: var(--color-primary);
  font-size: 20rpx;
  font-weight: 600;
}
.order-kind-badge.course { background: #F0EDFF; color: #7B61D1; }
.order-kind-badge.service { background: var(--color-accent-light); color: var(--color-accent-deep); }
.order-title {
  font-size: 28rpx;
  font-weight: 700;
  color: var(--color-text);
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.order-subtitle {
  margin-top: 8rpx;
  font-size: 22rpx;
  color: var(--color-text-secondary);
  line-height: 1.45;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.order-price-row {
  display: flex;
  align-items: baseline;
  gap: 12rpx;
  margin-top: auto;
  padding-top: 10rpx;
}
.order-price {
  font-size: 28rpx;
  font-weight: 700;
  color: var(--color-accent-deep);
}
.order-qty {
  font-size: 22rpx;
  color: var(--color-text-secondary);
}

.order-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12rpx;
  min-height: 86rpx;
  padding: 12rpx 24rpx;
  border-top: 1rpx solid var(--color-divider);
  box-sizing: border-box;
}
.order-due {
  font-size: 22rpx;
  color: var(--color-text-placeholder);
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.order-actions {
  flex-shrink: 0;
}
.action-btn {
  min-height: 56rpx;
  padding: 0 28rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1rpx solid #B9D6F3;
  border-radius: 12rpx;
  background: #F7FBFF;
  color: var(--color-primary);
  font-size: 24rpx;
  font-weight: 600;
  box-sizing: border-box;
}

/* 状态面板 */
.state-panel {
  padding: 24rpx;
}
.state-empty-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
}
.retry-btn {
  margin-top: 12rpx;
  padding: 16rpx 48rpx;
  background: var(--color-primary);
  color: #fff;
  border-radius: 12rpx;
  font-size: 26rpx;
}

/* 骨架 */
.skel-list { padding: 0 24rpx; }
.skel-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 20rpx 24rpx;
  margin-bottom: 24rpx;
}
.skel-line {
  height: 28rpx;
  border-radius: 8rpx;
  background: var(--color-divider);
  margin-bottom: 16rpx;
}
.skel-line.w40 { width: 40%; }
.skel-line.w50 { width: 50%; }
.skel-line.w70 { width: 70%; }

/* 筛选抽屉 */
.filter-sheet {
  padding: 24rpx 32rpx calc(32rpx + env(safe-area-inset-bottom));
}
.sheet-grip {
  width: 72rpx;
  height: 8rpx;
  border-radius: 4rpx;
  background: #D0D5DD;
  margin: 0 auto 24rpx;
}
.sheet-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.sheet-title {
  font-size: 34rpx;
  font-weight: 700;
  color: var(--color-text);
}
.sheet-close {
  width: 56rpx;
  height: 56rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}
.sheet-x {
  font-size: 40rpx;
  color: var(--color-text-secondary);
  line-height: 1;
}
.sheet-desc {
  display: block;
  margin: 16rpx 0 24rpx;
  font-size: 24rpx;
  color: var(--color-text-secondary);
}
.filter-options {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20rpx;
  margin-bottom: 32rpx;
}
.filter-option {
  height: 80rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12rpx;
  background: #F7F9FB;
  border: 1rpx solid var(--color-border);
  font-size: 26rpx;
  color: var(--color-text-secondary);
}
.filter-option.active {
  background: var(--color-primary-light);
  border-color: #B9D6F3;
  color: var(--color-primary);
  font-weight: 600;
}
.filter-confirm {
  height: 88rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12rpx;
  background: var(--color-primary);
  color: #fff;
  font-size: 30rpx;
  font-weight: 700;
  box-shadow: 0 6rpx 16rpx rgba(10, 102, 194, 0.22);
}

.bottom-spacer {
  height: 24rpx;
}
</style>
