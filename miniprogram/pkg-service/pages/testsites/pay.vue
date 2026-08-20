<template>
  <view class="tsp-page">
    <u-nav-bar title="确认预约" show-back @back="goBack" />

    <template v-if="draft">
      <!-- 预约摘要 -->
      <view class="summary-card">
        <text class="card-title">预约信息</text>
        <view class="row">
          <text class="row-label">场地</text>
          <text class="row-value">{{ draft.siteName }}</text>
        </view>
        <view class="row">
          <text class="row-label">日期</text>
          <text class="row-value">{{ draft.date }}</text>
        </view>
        <view class="row">
          <text class="row-label">时段</text>
          <text class="row-value">{{ draft.time_slot }}</text>
        </view>
        <view class="row">
          <text class="row-label">联系人</text>
          <text class="row-value">{{ draft.contact_name }}</text>
        </view>
        <view class="row">
          <text class="row-label">联系电话</text>
          <text class="row-value">{{ draft.contact_phone }}</text>
        </view>
        <view class="row">
          <text class="row-label">用途</text>
          <text class="row-value">{{ draft.purpose }}</text>
        </view>
        <view class="divider"></view>
        <view class="row">
          <text class="row-label">参考价格</text>
          <text class="row-value price">{{ formatPrice(draft.price_fen) }}</text>
        </view>
        <text class="price-note">以场地方实际报价为准</text>
      </view>

      <!-- 费用说明 -->
      <view class="notice-block">
        <text class="notice-title">费用说明</text>
        <text class="notice-line">· 定金及测试费用在线下向场地方缴纳，平台不参与资金流转</text>
        <text class="notice-line">· 预约提交后，场地方将与您联系确认费用与缴纳方式</text>
        <text class="notice-line warn">· 请勿向任何线上渠道缴纳费用，谨防上当受骗</text>
      </view>
    </template>

    <!-- 无预约草稿 -->
    <view v-else class="state-inline">
      <u-empty description="暂无预约信息" />
      <view class="retry-btn" @tap="goList">返回场地列表</view>
    </view>

    <!-- 底部确认栏 -->
    <view v-if="draft" class="bottom-bar">
      <view class="confirm-btn" @tap="confirmBooking">确认预约</view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'

const draft = ref(null)

function formatPrice(fen) {
  if (fen == null || fen <= 0) return '面议'
  const yuan = fen / 100
  const text = Number.isInteger(yuan) ? String(yuan) : yuan.toFixed(2)
  return '¥' + text
}

function confirmBooking() {
  // 平台不参与资金流转，此处不做任何支付，直接跳转预约结果
  const siteName = encodeURIComponent(draft.value.siteName || '')
  const date = encodeURIComponent(draft.value.date || '')
  uni.redirectTo({
    url: '/pkg-service/pages/testsites/result?status=success&siteName=' + siteName + '&date=' + date,
  })
}

function goBack() {
  uni.navigateBack()
}

function goList() {
  uni.reLaunch({ url: '/pkg-service/pages/testsites/list' })
}

onLoad(() => {
  const stored = uni.getStorageSync('testBookingDraft')
  draft.value = stored || null
})
</script>

<style scoped>
.tsp-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: 88px;
}

.state-inline {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 80px 20px;
  gap: 12px;
}

.retry-btn {
  padding: 8px 24px;
  background: #0A66C2;
  color: #fff;
  border-radius: 8px;
  font-size: 13px;
}

/* 预约摘要 */
.summary-card {
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  padding: 16px;
  margin: 12px;
}

.card-title {
  font-size: 15px;
  font-weight: 700;
  color: #17212B;
  display: block;
  margin-bottom: 12px;
}

.row {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  color: #344054;
  padding: 7px 0;
  line-height: 1.5;
}

.row-label {
  color: #667085;
  flex-shrink: 0;
  margin-right: 12px;
}

.row-value {
  text-align: right;
  word-break: break-all;
}

.row-value.price {
  color: #E96012;
  font-weight: 700;
  font-size: 15px;
}

.price-note {
  display: block;
  text-align: right;
  font-size: 11px;
  color: #98A2B3;
  margin-top: 4px;
}

.divider {
  border-top: 1px dashed #E4E7EC;
  margin: 10px 0;
}

/* 费用说明 */
.notice-block {
  background: #F4F8FC;
  border-radius: 8px;
  padding: 12px 16px;
  margin: 0 12px 8px;
}

.notice-title {
  font-size: 13px;
  font-weight: 600;
  color: #0A66C2;
  display: block;
  margin-bottom: 6px;
}

.notice-line {
  display: block;
  font-size: 12px;
  color: #344054;
  line-height: 1.7;
}

.notice-line.warn {
  color: #B54708;
}

/* 底部确认栏 */
.bottom-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 12px;
  padding-bottom: calc(12px + env(safe-area-inset-bottom));
  background: #fff;
  border-top: 1px solid #EEF1F4;
}

.confirm-btn {
  height: 46px;
  border-radius: 8px;
  background: #0A66C2;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 15px;
  font-weight: 600;
}
</style>
