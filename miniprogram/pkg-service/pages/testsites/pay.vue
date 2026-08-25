<template>
  <view class="tsp-page" :class="{ 'no-motion': noMotion }">
    <u-nav-bar title="预约已提交" show-back @back="goBack" />

    <template v-if="draft">
      <!-- 预约摘要 -->
      <view class="summary-card">
        <text class="submit-status">预约已提交，等待场地方确认（预计 24 小时内电话联系）</text>
        <text class="card-title">预约信息</text>
        <view class="row">
          <text class="row-label">场地</text>
          <text class="row-value">{{ draft.siteName }}</text>
        </view>
        <view v-if="draft.booking_type" class="row">
          <text class="row-label">预约类型</text>
          <text class="row-value">{{ typeLabel(draft.booking_type) }}</text>
        </view>
        <view class="row">
          <text class="row-label">日期</text>
          <text class="row-value">{{ dateText(draft.date) }}</text>
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
          <text class="row-value price" :class="{ face: isFace(draft.price_fen) }">{{ formatPrice(draft.price_fen) }}</text>
        </view>
        <text v-if="isFace(draft.price_fen)" class="price-face-note">价格面议，场地方确认后报价</text>
        <text v-else class="price-note">以场地方实际报价为准</text>
        <text class="price-note">如确认预约，场地方可能收取 50% 定金（线下支付）</text>
      </view>

      <!-- 费用说明 -->
      <view class="notice-block">
        <text class="notice-title">费用说明</text>
        <text class="notice-line">· 定金及测试费用在线下向场地方支付，平台不参与资金流转</text>
        <text class="notice-line">· 预约提交后，场地方将与您联系确认费用与支付方式</text>
        <text class="notice-line warn">· 请勿向任何线上渠道支付定金，谨防上当受骗</text>
      </view>
    </template>

    <!-- 无预约草稿 -->
    <view v-else class="state-inline">
      <u-empty description="暂无预约信息" />
      <view class="retry-btn" @tap="goList">返回场地列表</view>
    </view>

    <!-- 底部确认栏 -->
    <view v-if="draft" class="bottom-bar">
      <!-- 决策点回显（评审 P2）：日期/时段/无支付三要素在 CTA 上方可见 -->
      <text class="decision-line">{{ draft.date }} · {{ draft.time_slot }} · 无需在线支付</text>
      <view class="confirm-btn" @tap="goResult">查看预约结果</view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { useReduceMotion } from '@/utils/motion'

const draft = ref(null)
const { noMotion, checkMotion } = useReduceMotion()

function formatPrice(fen) {
  if (fen == null || fen <= 0) return '面议'
  const yuan = fen / 100
  const text = Number.isInteger(yuan) ? String(yuan) : yuan.toFixed(2)
  return '¥' + text
}
function isFace(fen) { return fen == null || fen <= 0 }
// 预约类型文案（booking.vue 提交的 booking_type）
function typeLabel(t) { return t === 'group' ? '团体预约' : '个人预约' }
// 日期口语格式（评审 Minor：ISO "2026-08-20" → "8月20日"，与 booking/detail 日期语言一致）
function dateText(iso) {
  if (!iso) return ''
  const m = String(iso).match(/^\d{4}-(\d{2})-(\d{2})/)
  return m ? Number(m[1]) + '月' + Number(m[2]) + '日' : String(iso)
}

let resultTapped = false
function goResult() {
  // 预约在 booking 页已服务端落库；本页不做任何支付，「查看预约结果」仅跳转（评审 P1 杀掉假高潮）
  if (resultTapped) return // 重入守卫（评审 P3）：双击只放行一次 redirectTo
  resultTapped = true
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
  checkMotion()
  // 评审 Minor：getStorageSync 更新后可返回垃圾，读取兜底为空
  let stored = null
  try {
    stored = uni.getStorageSync('testBookingDraft')
  } catch (e) {
    stored = null
  }
  draft.value = stored || null
})
</script>

<style scoped>
.tsp-page {
  min-height: 100vh;
  background: #fff; /* 白上白：白底页面 + 描边软角卡片（对齐组） */
  padding-bottom: 240rpx; /* 固定底栏之上留呼吸 */
}

.state-inline {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 120rpx 40rpx;
  gap: 24rpx;
}

.retry-btn {
  min-height: 88rpx;
  padding: 0 48rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #0A66C2;
  color: #fff;
  border-radius: 16rpx; /* 对齐组按钮：16rpx，非全圆 */
  font-size: 28rpx;
  font-weight: 600;
  box-shadow: inset 0 2rpx 0 rgba(255,255,255,.22), inset 0 -4rpx 10rpx rgba(7,77,146,.18), 0 4rpx 14rpx rgba(10,102,194,.25);
}
.retry-btn:active { opacity: 0.85; }

/* 预约摘要（白上白卡片） */
.summary-card {
  background: #fff;
  margin: 16rpx 24rpx 0;
  padding: 32rpx;
  border: 2rpx solid #E4E7EC;
  border-radius: 20rpx;
  box-shadow:
    0 2rpx 6rpx rgba(10, 30, 60, 0.04),
    0 12rpx 32rpx rgba(10, 30, 60, 0.05);
  animation: fade-in 0.22s ease-out backwards;
}

.submit-status {
  display: block;
  font-size: 24rpx;
  color: #667085;
  margin-bottom: 12rpx;
}

.card-title {
  font-size: 30rpx; /* 对齐组例外：卡片标题 */
  font-weight: 700;
  color: #17212B;
  display: block;
  margin-bottom: 16rpx;
}

.row {
  display: flex;
  justify-content: space-between;
  font-size: 26rpx;
  color: #344054;
  padding: 14rpx 0;
  line-height: 1.5;
}

.row-label {
  color: #667085;
  flex-shrink: 0;
  margin-right: 24rpx;
}

.row-value {
  text-align: right;
  word-break: break-all;
}

.row-value.price {
  color: #C2410C; /* 价格深色令牌：烬橙 #E96012 白底 ≈3.4:1 → 提深过 AA（评审 P3） */
  font-weight: 700;
  font-size: 30rpx; /* 对齐组例外：金额强调 */
}
.row-value.price.face {
  font-size: 26rpx;
  font-weight: 600;
  color: #667085; /* 面议降级：不是数字，不用金额重量渲染 */
}

.price-face-note {
  display: block;
  text-align: right;
  font-size: 24rpx;
  color: #667085;
  margin-top: 8rpx;
}

.price-note {
  display: block;
  text-align: right;
  font-size: 24rpx;
  color: #667085;
  margin-top: 4rpx;
}

.divider {
  border-top: 2rpx solid #f0f1f3; /* 实线（Solid-Line：正常态禁虚线） */
  margin: 20rpx 0;
}

/* 费用说明 */
.notice-block {
  background: #EAF3FB; /* 蓝雾浅底（对齐组 tint） */
  border-radius: 20rpx;
  padding: 24rpx 32rpx;
  margin: 16rpx 24rpx 0;
  animation: fade-in 0.22s ease-out backwards;
  animation-delay: 20ms;
}

.notice-title {
  font-size: 26rpx;
  font-weight: 600;
  color: #0A66C2;
  display: block;
  margin-bottom: 8rpx;
}

.notice-line {
  display: block;
  font-size: 24rpx;
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
  padding: 24rpx 32rpx calc(24rpx + env(safe-area-inset-bottom));
  background: #fff;
  border-top: 2rpx solid #f0f1f3;
  box-shadow: 0 -4rpx 20rpx rgba(10, 30, 60, 0.05);
  z-index: 50;
}

.decision-line {
  display: block;
  text-align: center;
  font-size: 24rpx;
  color: #667085;
  margin-bottom: 12rpx;
}

.confirm-btn {
  height: 88rpx;
  border-radius: 16rpx; /* 对齐组主行动按钮：16rpx */
  background: #0A66C2;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 30rpx;
  font-weight: 600;
  box-shadow: inset 0 2rpx 0 rgba(255,255,255,.22), inset 0 -4rpx 10rpx rgba(7,77,146,.18), 0 4rpx 14rpx rgba(10,102,194,.25);
  transition: transform 0.15s ease-out;
}
.confirm-btn:active { transform: scale(0.96); }

/* 减弱动效（无障碍）：装饰动画全关 */
.no-motion .summary-card,
.no-motion .notice-block { animation: none; }

@keyframes fade-in {
  from { opacity: 0; transform: translateY(16rpx); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
