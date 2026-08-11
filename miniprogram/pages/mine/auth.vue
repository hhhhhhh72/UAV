<template>
  <view class="auth-page">
    <u-nav-bar title="实名认证" show-back @back="goBack" />

    <!-- 已认证状态卡（写死数据，仅演示闭环） -->
    <view class="status-card status-card--approved">
      <view class="status-icon">实</view>
      <text class="status-title">实名认证已完成</text>
      <text class="status-desc">身份信息核验通过，您已完成实名登记</text>

      <!-- 身份信息展示 -->
      <view class="identity-rows">
        <view class="identity-row">
          <text class="identity-label">真实姓名</text>
          <text class="identity-value">{{ record.real_name }}</text>
        </view>
        <view class="identity-row">
          <text class="identity-label">身份证号</text>
          <text class="identity-value">{{ record.id_card }}</text>
        </view>
      </view>
    </view>

    <!-- 演示说明 -->
    <view class="privacy">
      <text class="privacy-tag">说明</text>
      <text class="privacy-text">本页为演示数据：真实身份核验需对接权威渠道，当前版本实名认证状态固定展示，供流程闭环验收使用。</text>
    </view>
  </view>
</template>

<script setup>
// 实名认证：写死已认证数据（不接外部核验渠道、无真实接口）
// 让「认证飞手」流程闭环：mine 页实名状态显示已认证，点击进入本页查看身份信息。
const record = {
  real_name: '张航',
  id_card: '500***********123X',
}

const goBack = () => uni.navigateBack()
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

/* 说明 */
.privacy { display: flex; align-items: flex-start; gap: 12rpx; padding: 24rpx 40rpx 0; }
.privacy-tag {
  font-size: 20rpx;
  padding: 2rpx 12rpx;
  border-radius: 4px;
  background: var(--color-primary-light);
  color: var(--color-primary);
  font-weight: 600;
  flex-shrink: 0;
}
.privacy-text { font-size: 22rpx; color: var(--color-text-placeholder); line-height: 1.6; }
</style>
