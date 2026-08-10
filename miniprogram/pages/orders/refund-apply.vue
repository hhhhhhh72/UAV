<template>
  <view class="refund-page">
    <u-nav-bar :title="submitted ? '退款申请已提交' : '申请退款'" show-back @back="goBack" />

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
      <!-- 提交成功态 -->
      <view v-if="submitted" class="success-card">
        <view class="success-mark">✓</view>
        <text class="success-title">退款申请已提交</text>
        <text class="success-desc">平台将在 1 个工作日内完成审核并同步处理进度</text>
        <view class="success-actions">
          <view class="success-btn" @tap="goAftersale">
            <text>查看售后进度</text>
          </view>
        </view>
      </view>

      <!-- 退款表单 -->
      <template v-else>
        <view class="service-card">
          <text class="svc-title">{{ order.title }}</text>
          <text class="svc-sub">{{ order.subtitle }}</text>
        </view>

        <view class="form-card">
          <text class="form-label">退款信息</text>
          <view class="data-row">
            <text class="data-label">退款类型</text>
            <text class="data-value">{{ as ? as.type : '服务未按约执行' }}</text>
          </view>
          <view class="data-row">
            <text class="data-label">退款金额</text>
            <text class="data-value accent">¥{{ as ? fmtFen(as.amount_fen) : fmtFen(order.amount_fen) }}</text>
          </view>
          <view class="data-row">
            <text class="data-label">退款方式</text>
            <text class="data-value">原路退回</text>
          </view>
        </view>

        <view class="form-card">
          <text class="form-label">问题说明</text>
          <textarea
            class="refund-textarea"
            v-model="reason"
            :placeholder="defaultReason"
            :maxlength="300"
          />
          <text class="textarea-count">{{ reason.length }}/300</text>
        </view>

        <view class="note-card">
          <text class="note-text">退款申请目前仅保存为演示草稿，平台审核后才进入真实退款流程，不会调用资金托管退款接口。</text>
        </view>

        <view class="submit-wrap">
          <view class="submit-btn" @tap="submitRefund">
            <text>提交退款申请</text>
          </view>
        </view>
      </template>
    </template>

    <view class="bottom-spacer"></view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { loadOrder, fmtFen } from '../../utils/orderAdapter'

const order = ref(null)
const loading = ref(true)
const error = ref(false)
const submitted = ref(false)
const reason = ref('')

const as = computed(() => order.value?.detail?.aftersale || null)
const defaultReason = computed(() =>
  as.value?.description ? `已提交：${as.value.description}` : '请描述需要退款的原因',
)

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

const submitRefund = async () => {
  if (submitted.value) return
  // 第一期为演示提交：真实退款需要后端售后工单契约 + 平台审核，
  // 不错误调用 escrow/refund，也不修改订单状态。
  submitted.value = true
  uni.showToast({ title: '退款申请已提交', icon: 'success' })
}

const goAftersale = () => {
  // refund-apply 由 aftersale 进入，返回即回到售后进度
  uni.navigateBack()
}

const goBack = () => {
  if (submitted.value) {
    uni.reLaunch({ url: '/pages/orders/index' })
    return
  }
  uni.navigateBack()
}
</script>

<style scoped>
.refund-page {
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

.service-card {
  margin: 20rpx 24rpx 0;
  background: #fff;
  border: 1rpx solid var(--color-border);
  border-radius: 16rpx;
  box-shadow: var(--shadow-sm);
  padding: 28rpx;
}
.svc-title {
  display: block;
  font-size: 28rpx;
  font-weight: 700;
  color: var(--color-text);
  line-height: 1.4;
}
.svc-sub {
  display: block;
  margin-top: 10rpx;
  font-size: 24rpx;
  color: var(--color-text-secondary);
}

.form-card {
  margin: 20rpx 24rpx 0;
  background: #fff;
  border: 1rpx solid var(--color-border);
  border-radius: 16rpx;
  box-shadow: var(--shadow-sm);
  padding: 28rpx;
}
.form-label {
  display: block;
  font-size: 26rpx;
  font-weight: 600;
  color: var(--color-text);
}
.data-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24rpx;
  min-height: 76rpx;
  border-top: 1rpx solid var(--color-divider);
  margin-top: 12rpx;
}
.data-row:first-of-type { border-top: none; }
.data-label { font-size: 24rpx; color: var(--color-text-secondary); flex-shrink: 0; }
.data-value { font-size: 24rpx; color: var(--color-text); text-align: right; }
.data-value.accent { color: var(--color-accent-deep); font-weight: 700; }

.refund-textarea {
  margin-top: 20rpx;
  width: 100%;
  height: 200rpx;
  box-sizing: border-box;
  background: var(--color-bg);
  border-radius: 12rpx;
  padding: 20rpx;
  font-size: 26rpx;
  line-height: 1.5;
}
.textarea-count {
  display: block;
  text-align: right;
  margin-top: 8rpx;
  font-size: 22rpx;
  color: var(--color-text-placeholder);
}

.note-card {
  margin: 20rpx 24rpx 0;
  background: #FFF4E6;
  border-radius: 12rpx;
  padding: 20rpx 24rpx;
}
.note-text {
  font-size: 22rpx;
  color: #B54708;
  line-height: 1.6;
}

.submit-wrap {
  margin: 32rpx 24rpx 0;
}
.submit-btn {
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

.success-card {
  margin: 24rpx;
  background: #fff;
  border: 1rpx solid var(--color-border);
  border-radius: 16rpx;
  box-shadow: var(--shadow-sm);
  padding: 80rpx 40rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.success-mark {
  width: 96rpx;
  height: 96rpx;
  border-radius: 50%;
  background: var(--color-success);
  color: #fff;
  font-size: 48rpx;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}
.success-title {
  margin-top: 24rpx;
  font-size: 32rpx;
  font-weight: 700;
  color: var(--color-text);
}
.success-desc {
  margin-top: 12rpx;
  font-size: 24rpx;
  color: var(--color-text-secondary);
  text-align: center;
}
.success-actions {
  margin-top: 32rpx;
  width: 100%;
}
.success-btn {
  height: 84rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12rpx;
  background: var(--color-primary);
  color: #fff;
  font-size: 28rpx;
  font-weight: 600;
}

.bottom-spacer { height: 24rpx; }
</style>
