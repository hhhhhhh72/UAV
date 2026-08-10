<template>
  <view class="logistics-page">
    <u-nav-bar title="物流详情" show-back right-text="客服" @back="goBack" @right="openCustomerService" />

    <view v-if="loading" class="state-panel">
      <view class="loading-inline">
        <u-loading size="24rpx" />
        <text>加载中...</text>
      </view>
    </view>

    <view v-else-if="error || !order" class="state-panel">
      <u-empty description="物流信息加载失败" />
      <view class="retry-btn" @tap="loadData">
        <text>重新加载</text>
      </view>
    </view>

    <template v-else>
      <!-- 承运概要 -->
      <view class="logistics-card">
        <text class="lg-title">{{ lg.carrier }} · {{ lg.tracking_no }}</text>
        <text class="lg-status">{{ lg.latest }}</text>
      </view>

      <!-- 运输节点 -->
      <view class="nodes-card">
        <text class="nodes-title">运输节点</text>
        <view v-for="(node, i) in lg.nodes" :key="i" class="node-row" :class="{ current: i === 0 }">
          <view class="node-line">
            <view class="node-dot" />
            <view v-if="i < lg.nodes.length - 1" class="node-stem" />
          </view>
          <view class="node-copy">
            <text class="node-time">{{ node.time }}</text>
            <text class="node-text">{{ node.text }}</text>
          </view>
        </view>
      </view>
    </template>

    <view class="bottom-spacer"></view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { loadOrder, toastCustomerService } from '../../utils/orderAdapter'

const order = ref(null)
const loading = ref(true)
const error = ref(false)

const lg = computed(() => order.value?.detail?.logistics || null)

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
    if (!o || !o.detail || !o.detail.logistics) error.value = true
  } catch (e) {
    error.value = true
  } finally {
    loading.value = false
  }
}

onLoad(loadData)

const goBack = () => {
  uni.navigateBack()
}

const openCustomerService = () => {
  toastCustomerService()
}
</script>

<style scoped>
.logistics-page {
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

.logistics-card {
  margin: 20rpx 24rpx 0;
  background: #fff;
  border: 1rpx solid var(--color-border);
  border-radius: 16rpx;
  box-shadow: var(--shadow-sm);
  padding: 28rpx;
}
.lg-title {
  display: block;
  font-size: 28rpx;
  font-weight: 700;
  color: var(--color-text);
}
.lg-status {
  display: block;
  margin-top: 12rpx;
  font-size: 24rpx;
  color: var(--color-success);
  line-height: 1.6;
}

.nodes-card {
  margin: 20rpx 24rpx 0;
  background: #fff;
  border: 1rpx solid var(--color-border);
  border-radius: 16rpx;
  box-shadow: var(--shadow-sm);
  padding: 28rpx;
}
.nodes-title {
  display: block;
  font-size: 28rpx;
  font-weight: 700;
  color: var(--color-text);
  margin-bottom: 20rpx;
}
.node-row {
  display: flex;
  gap: 20rpx;
}
.node-line {
  width: 32rpx;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  position: relative;
}
.node-dot {
  width: 20rpx;
  height: 20rpx;
  border-radius: 50%;
  background: var(--color-text-placeholder);
  margin-top: 8rpx;
  flex-shrink: 0;
}
.node-row.current .node-dot {
  background: var(--color-primary);
  box-shadow: 0 0 0 6rpx var(--color-primary-light);
}
.node-stem {
  width: 4rpx;
  flex: 1;
  background: var(--color-divider);
  margin: 8rpx 0;
}
.node-copy {
  flex: 1;
  min-width: 0;
  padding-bottom: 28rpx;
}
.node-time {
  display: block;
  font-size: 24rpx;
  font-weight: 600;
  color: var(--color-text);
}
.node-text {
  display: block;
  margin-top: 6rpx;
  font-size: 22rpx;
  color: var(--color-text-secondary);
  line-height: 1.5;
}
.node-row.current .node-text { color: var(--color-primary); }

.bottom-spacer { height: 24rpx; }
</style>
