<template>
  <view class="auth-page">
    <u-nav-bar title="实名认证" show-back @back="goBack" />

    <!-- 加载中 -->
    <view v-if="loading" class="state-panel">
      <u-loading size="28rpx" />
      <text class="state-desc">加载中...</text>
    </view>

    <!-- 错误态 -->
    <view v-else-if="error" class="state-panel">
      <view class="status-icon status-icon--err">!</view>
      <text class="status-title">加载失败</text>
      <text class="status-desc">实名认证信息获取失败，请检查网络后重试</text>
      <view class="retry-btn" hover-class="tap-fade" @tap="fetchMine">重新加载</view>
    </view>

    <!-- 已认证状态卡（真实飞手认证档案） -->
    <view v-else-if="record" class="status-card status-card--approved">
      <view class="status-icon">实</view>
      <text class="status-title">实名认证已完成</text>
      <text class="status-desc">身份信息核验通过，您已完成实名登记</text>

      <!-- 身份信息展示 -->
      <view class="identity-rows">
        <view class="identity-row">
          <text class="identity-label">真实姓名</text>
          <text class="identity-value">{{ record.real_name || '—' }}</text>
        </view>
        <view class="identity-row">
          <text class="identity-label">认证地区</text>
          <text class="identity-value">{{ record.region || '—' }}</text>
        </view>
        <view class="identity-row">
          <text class="identity-label">认证状态</text>
          <text class="identity-value">{{ statusText }}</text>
        </view>
      </view>
    </view>

    <!-- 空态：无飞手认证记录，引导提交申请 -->
    <view v-else class="status-card">
      <view class="status-icon">实</view>
      <text class="status-title">尚未完成实名认证</text>
      <text class="status-desc">提交飞手认证申请并通过审核后，即可完成实名认证并展示在认证飞手名录中</text>
      <view class="apply-btn" hover-class="tap-fade" @tap="goApply">去提交飞手认证</view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { request } from '../../utils/request'

// 实名认证：真实数据来自飞手认证接口（GET /api/v1/certified-pilots/mine）。
// 仅 approved 记录展示真实脱敏档案；无记录展示空态并引导申请；请求失败展示错误态。
const loading = ref(true)
const error = ref(false)
const record = ref(null) // 仅 approved 飞手认证记录

const statusText = computed(() => {
  const map = { approved: '已认证', pending: '审核中', rejected: '未通过' }
  return (record.value && map[record.value.status]) || '已认证'
})

const fetchMine = async () => {
  loading.value = true
  error.value = false
  try {
    const res = await request({ url: '/api/v1/certified-pilots/mine' })
    const mine = (res && res.data) || res || null
    record.value = mine && mine.status === 'approved' ? mine : null
  } catch (e) {
    record.value = null
    error.value = true
  } finally {
    loading.value = false
  }
}

const goApply = () => {
  uni.navigateTo({ url: '/pkg-talent/pages/pilots/apply' })
}

const goBack = () => uni.navigateBack()

onShow(() => { fetchMine() })
</script>

<style scoped>
.auth-page { min-height: 100vh; background: var(--color-bg); padding-bottom: 40rpx; }

/* ── 状态卡 ── */
.status-card {
  margin: 28rpx 24rpx 0;
  background: var(--color-bg-card);
  border-radius: 8px;
  padding: 40rpx 32rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12rpx;
  box-shadow: 0 3px 12px rgba(16,24,40,.05);
}
.status-icon {
  width: 88rpx;
  height: 88rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 36rpx;
  font-weight: 700;
  color: #fff;
  background: var(--color-success, #16A34A);
}
.status-icon--err { background: #D92D20; }
.status-title { font-size: 32rpx; font-weight: 700; color: var(--color-text); }
.status-desc { font-size: 24rpx; color: var(--color-text-secondary); text-align: center; line-height: 1.6; }

/* 身份信息展示 */
.identity-rows {
  margin-top: 16rpx;
  width: 100%;
  background: #FAFAFA;
  border-radius: 8px;
  padding: 8rpx 24rpx;
}
.identity-row { display: flex; justify-content: space-between; align-items: center; height: 72rpx; }
.identity-row + .identity-row { border-top: 1rpx solid var(--color-border, #EEF0F2); }
.identity-label { font-size: 24rpx; color: var(--color-text-secondary); }
.identity-value { font-size: 26rpx; color: var(--color-text); font-weight: 600; }

/* 加载 / 错误态 */
.state-panel {
  min-height: 480rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16rpx;
  padding: 40rpx;
}
.state-desc { font-size: 24rpx; color: var(--color-text-secondary); }
.retry-btn,
.apply-btn {
  margin-top: 20rpx;
  padding: 14rpx 56rpx;
  border-radius: 999rpx;
  background: var(--color-primary);
  color: #fff;
  font-size: 26rpx;
  font-weight: 600;
}
.tap-fade { opacity: 0.8; }
</style>
