<template>
  <view class="tsr-page">
    <u-nav-bar title="预约结果" left-text="" />

    <!-- 预约成功 -->
    <view v-if="status === 'success'" class="result-body">
      <view class="icon-circle success">
        <view class="check-mark"></view>
      </view>
      <text class="result-title">预约提交成功</text>
      <text class="result-desc">预约已提交，场地方审核后将与您联系确认</text>

      <view class="result-card">
        <view class="row">
          <text class="row-label">场地</text>
          <text class="row-value">{{ siteName }}</text>
        </view>
        <view class="row">
          <text class="row-label">日期</text>
          <text class="row-value">{{ date }}</text>
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
      </view>

      <view class="action-btn" @tap="goList">返回场地列表</view>
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

const status = ref('fail')
const siteName = ref('')
const date = ref('')
const draft = ref(null)

function goList() {
  uni.reLaunch({ url: '/pkg-service/pages/testsites/list' })
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
  status.value = options.status === 'success' ? 'success' : 'fail'
  siteName.value = decodeURIComponent(options.siteName || '')
  date.value = decodeURIComponent(options.date || '')
  const stored = uni.getStorageSync('testBookingDraft')
  draft.value = stored || null
})
</script>

<style scoped>
.tsr-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: env(safe-area-inset-bottom);
}

.result-body {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px 16px;
}

/* 状态图标（CSS 绘制，无 emoji） */
.icon-circle {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 16px;
}

.icon-circle.success {
  background: #168A55;
}

.icon-circle.fail {
  background: #D92D20;
}

.check-mark {
  width: 20px;
  height: 10px;
  border-left: 3px solid #fff;
  border-bottom: 3px solid #fff;
  transform: rotate(-45deg);
  margin-top: -4px;
}

.cross-mark {
  position: relative;
  width: 20px;
  height: 20px;
}

.cross-mark::before,
.cross-mark::after {
  content: '';
  position: absolute;
  left: 50%;
  top: 50%;
  width: 3px;
  height: 20px;
  border-radius: 2px;
  background: #fff;
}

.cross-mark::before {
  transform: translate(-50%, -50%) rotate(45deg);
}

.cross-mark::after {
  transform: translate(-50%, -50%) rotate(-45deg);
}

.result-title {
  font-size: 19px;
  font-weight: 700;
  color: #17212B;
  margin-bottom: 6px;
}

.result-desc {
  font-size: 13px;
  color: #667085;
  text-align: center;
  line-height: 1.6;
  margin-bottom: 20px;
}

/* 预约信息卡 */
.result-card {
  width: 100%;
  box-sizing: border-box;
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  padding: 14px 16px;
  margin-bottom: 8px;
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

/* 费用说明 */
.notice-block {
  width: 100%;
  box-sizing: border-box;
  background: #F4F8FC;
  border-radius: 8px;
  padding: 12px 16px;
  margin-bottom: 20px;
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

/* 操作按钮 */
.action-btn {
  width: 100%;
  box-sizing: border-box;
  height: 46px;
  border-radius: 8px;
  background: #0A66C2;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 15px;
  font-weight: 600;
  margin-bottom: 10px;
}

.action-btn.outline {
  background: #fff;
  color: #0A66C2;
  border: 1px solid #0A66C2;
}
</style>
