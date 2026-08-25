<template>
  <view class="tsr-page" :class="{ 'no-motion': noMotion }">
    <u-nav-bar title="预约结果" left-text="" />

    <!-- 预约已提交(待审核) -->
    <view v-if="status === 'success'" class="result-body">
      <view class="icon-circle pending">
        <view class="clock-face"></view>
      </view>
      <text class="result-title">预约已提交</text>
      <text class="result-desc">预计 24 小时内，场地方将电话联系您确认费用与测试时段</text>

      <view class="result-card">
        <view class="row">
          <text class="row-label">场地</text>
          <text class="row-value">{{ siteName }}</text>
        </view>
        <view class="row">
          <text class="row-label">日期</text>
          <text class="row-value">{{ dateText(date) }}</text>
        </view>
        <view v-if="draft" class="row">
          <text class="row-label">时段</text>
          <text class="row-value">{{ draft.time_slot }}</text>
        </view>
        <view v-if="draft" class="row">
          <text class="row-label">联系人</text>
          <text class="row-value">{{ draft.contact_name }}</text>
        </view>
        <view v-if="draft" class="row">
          <text class="row-label">联系电话</text>
          <text class="row-value">{{ draft.contact_phone }}</text>
        </view>
      </view>

      <view class="notice-block">
        <text class="notice-title">费用说明</text>
        <text class="notice-line">· 定金及测试费用在线下向场地方支付，平台不参与资金流转</text>
        <text class="notice-line">· 审核通过后可在「我的预约」查看预约进度</text>
      </view>

      <view class="action-btn" @tap="goList">返回场地列表</view>
      <view class="action-btn outline" @tap="goMine">查看我的预约</view>
    </view>

    <!-- 预约失败 -->
    <view v-else class="result-body">
      <view class="icon-circle fail">
        <view class="cross-mark"></view>
      </view>
      <text class="result-title">预约失败</text>
      <text class="result-desc">预约提交未成功，请重新尝试</text>

      <view class="action-btn" @tap="goRetry">重新预约</view>
      <view class="action-btn outline" @tap="goList">返回场地列表</view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { useReduceMotion } from '@/utils/motion'

const status = ref('success')
const siteName = ref('')
const date = ref('')
const draft = ref(null)
const { noMotion, checkMotion } = useReduceMotion()

// 日期口语格式（评审 Minor：ISO "2026-08-20" → "8月20日"，与 booking/detail 日期语言一致）
function dateText(iso) {
  if (!iso) return ''
  const m = String(iso).match(/^\d{4}-(\d{2})-(\d{2})/)
  return m ? Number(m[1]) + '月' + Number(m[2]) + '日' : String(iso)
}

function goList() {
  uni.reLaunch({ url: '/pkg-service/pages/testsites/list' })
}

function goMine() {
  uni.navigateTo({ url: '/pkg-service/pages/testsites/mybookings' })
}

function goRetry() {
  // 失败重试：回到预约表单重新填写
  if (draft.value && draft.value.siteId) {
    uni.reLaunch({
      url: '/pkg-service/pages/testsites/booking?id=' + encodeURIComponent(draft.value.siteId),
    })
    return
  }
  uni.reLaunch({ url: '/pkg-service/pages/testsites/list' })
}

onLoad((options) => {
  checkMotion()
  // 评审 P1：默认渲染「预约已提交」（pending 语义）——无参数落地/重开不再误显「预约失败」；
  // fail 分支仅显式 status=fail 时出现（保留给未来 booking catch 接 result fail）
  status.value = options.status === 'fail' ? 'fail' : 'success'
  siteName.value = decodeURIComponent(options.siteName || '')
  date.value = decodeURIComponent(options.date || '')
  const stored = uni.getStorageSync('testBookingDraft')
  draft.value = stored || null
})
</script>

<style scoped>
.tsr-page {
  min-height: 100vh;
  background: #fff; /* 白上白：白底页面（对齐组） */
  padding-bottom: env(safe-area-inset-bottom);
}

.result-body {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 48rpx 32rpx;
  animation: fade-in 0.22s ease-out backwards;
}

/* 状态图标（CSS 绘制，无 emoji） */
.icon-circle {
  width: 144rpx;
  height: 144rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24rpx;
  /* focal：成功时刻圆盘轻弹入（scale+opacity，一次，无循环） */
  animation: icon-in 0.4s cubic-bezier(0.16, 1, 0.3, 1) backwards;
}

.icon-circle.pending {
  background: #0A66C2; /* 待审核：塔台蓝，非确认绿 */
}

.icon-circle.fail {
  background: #D92D20;
}

/* 待审核时钟（CSS 绘制，无 emoji）：圆盘 + 12 点分针 + 9 点时针对齐 */
.clock-face {
  position: relative;
  width: 72rpx;
  height: 72rpx;
  border: 6rpx solid #fff;
  border-radius: 50%;
  box-sizing: border-box;
}

.clock-face::before,
.clock-face::after {
  content: '';
  position: absolute;
  left: 50%;
  top: 50%;
  width: 6rpx;
  border-radius: 8rpx; /* 阶梯值；6rpx 宽笔画上 ≥3rpx 即完全圆端，8rpx 渲染等同 3rpx */
  background: #fff;
}

.clock-face::before {
  height: 24rpx; /* 分针：指向 12 点 */
  transform: translate(-50%, -100%);
  transform-origin: 50% 100%; /* 底端=钟心：旋转即绕钟心摆动（focal 指针校对） */
  animation: minute-in 0.4s cubic-bezier(0.16, 1, 0.3, 1) backwards;
}

.clock-face::after {
  height: 16rpx; /* 时针：指向 9 点 */
  transform: translate(-50%, -100%) rotate(90deg);
  transform-origin: 50% 100%;
  animation: hour-in 0.4s cubic-bezier(0.16, 1, 0.3, 1) 0.15s backwards; /* 错峰：分针先归位，时针随后摆到 9 点 */
}

.cross-mark {
  position: relative;
  width: 40rpx;
  height: 40rpx;
}

.cross-mark::before,
.cross-mark::after {
  content: '';
  position: absolute;
  left: 50%;
  top: 50%;
  width: 6rpx;
  height: 40rpx;
  border-radius: 8rpx; /* 阶梯值；6rpx 宽笔画上 ≥3rpx 即完全圆端，8rpx 渲染等同 3rpx */
  background: #fff;
}

.cross-mark::before {
  transform: translate(-50%, -50%) rotate(45deg);
}

.cross-mark::after {
  transform: translate(-50%, -50%) rotate(-45deg);
}

.result-title {
  font-size: 36rpx;
  font-weight: 700;
  color: #17212B;
  margin-bottom: 12rpx;
}

.result-desc {
  font-size: 26rpx; /* 对齐组例外 */
  color: #667085;
  text-align: center;
  line-height: 1.6;
  margin-bottom: 32rpx;
}

/* 预约信息卡（白上白卡片） */
.result-card {
  width: 100%;
  box-sizing: border-box;
  background: #fff;
  margin-bottom: 16rpx;
  border: 2rpx solid #E4E7EC;
  border-radius: 20rpx;
  padding: 28rpx 32rpx;
  box-shadow:
    0 2rpx 6rpx rgba(10, 30, 60, 0.04),
    0 12rpx 32rpx rgba(10, 30, 60, 0.05);
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

/* 费用说明 */
.notice-block {
  width: 100%;
  box-sizing: border-box;
  background: #EAF3FB; /* 蓝雾浅底（对齐组 tint） */
  border-radius: 20rpx;
  padding: 24rpx 32rpx;
  margin-bottom: 32rpx;
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

/* 操作按钮 */
.action-btn {
  width: 100%;
  box-sizing: border-box;
  height: 88rpx;
  border-radius: 16rpx; /* 对齐组主行动按钮：16rpx */
  background: #0A66C2;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 30rpx;
  font-weight: 600;
  margin-bottom: 20rpx;
  box-shadow: inset 0 2rpx 0 rgba(255,255,255,.22), inset 0 -4rpx 10rpx rgba(7,77,146,.18), 0 4rpx 14rpx rgba(10,102,194,.25);
  transition: transform 0.15s ease-out;
}
.action-btn:active { transform: scale(0.96); }

.action-btn.outline {
  background: #fff;
  color: #0A66C2;
  border: 2rpx solid #0A66C2;
  box-shadow: none;
}

/* 减弱动效（无障碍）：装饰动画与指针摆动全关，保留状态色 */
.no-motion .result-body,
.no-motion .icon-circle,
.no-motion .clock-face::before,
.no-motion .clock-face::after { animation: none; }

@keyframes fade-in {
  from { opacity: 0; transform: translateY(16rpx); }
  to { opacity: 1; transform: translateY(0); }
}

/* focal 编排：圆盘弹入 + 指针校对（预约=等待场地方确认的时钟隐喻） */
@keyframes icon-in {
  from { opacity: 0; transform: scale(0.85); }
  to { opacity: 1; transform: scale(1); }
}
@keyframes minute-in {
  from { transform: translate(-50%, -100%) rotate(-40deg); } /* 从 11 点方向摆入 */
  to { transform: translate(-50%, -100%) rotate(0); }
}
@keyframes hour-in {
  from { transform: translate(-50%, -100%) rotate(150deg); } /* 从 12 点方向摆下 */
  to { transform: translate(-50%, -100%) rotate(90deg); }
}
</style>
