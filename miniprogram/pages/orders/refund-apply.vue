<template>
  <view class="pub-page" :style="{ paddingTop: topPad + 'px' }">
    <!-- 顶栏（与发布页同款） -->
    <view class="pub-nav">
      <view class="pub-back" hover-class="pub-fade" @tap="goBack">‹</view>
      <view class="pub-nav-title">{{ submitted ? '退款申请已提交' : '申请退款' }}</view>
    </view>

    <!-- 加载中 -->
    <view v-if="loading" class="loading-state">
      <u-loading size="28rpx" />
      <text>加载中...</text>
    </view>

    <!-- 加载失败 -->
    <view v-else-if="error || !order" class="pub-empty">
      <view class="pub-empty-mark">!</view>
      <view class="pub-empty-title">订单加载失败</view>
      <view class="pub-empty-desc">网络异常，请稍后重试</view>
      <view class="pub-btn pub-btn--primary retry-btn" hover-class="pub-btn--active" @tap="retryLoad">重新加载</view>
    </view>

    <template v-else>
      <!-- 提交成功态 -->
      <view v-if="submitted" class="pub-form-card">
        <view class="pub-success">
          <view class="pub-success-mark">✓</view>
          <view class="pub-success-title">退款申请已提交</view>
          <view class="pub-success-desc">平台将在 1 个工作日内完成审核并同步处理进度</view>
        </view>
      </view>

      <!-- 退款表单 -->
      <template v-else>
        <!-- 订单概要 -->
        <view class="pub-form-intro">
          <view class="pub-form-intro-h2">{{ order.title }}</view>
          <view class="pub-form-intro-p">{{ order.subtitle }}</view>
        </view>

        <!-- 退款信息（只读） -->
        <view class="pub-section">
          <view class="pub-section-title">退款信息</view>
          <view class="pub-form-card">
            <view class="pub-field">
              <view class="pub-field-label">退款类型</view>
              <view class="pub-field-value">{{ as ? as.type : '仅退款' }}</view>
            </view>
            <view class="pub-field">
              <view class="pub-field-label">退款金额</view>
              <view class="pub-field-value pub-field-value--accent">¥{{ as ? fmtFen(as.amount_fen) : fmtFen(order.amount_fen) }}</view>
            </view>
            <view class="pub-field">
              <view class="pub-field-label">退款方式</view>
              <view class="pub-field-value">原路退回</view>
            </view>
          </view>
        </view>

        <!-- 问题说明 -->
        <view class="pub-section">
          <view class="pub-section-title">问题说明</view>
          <view class="pub-form-card">
            <view class="pub-field">
              <textarea
                class="pub-input pub-input--textarea"
                v-model="reason"
                :placeholder="defaultReason"
                placeholder-class="pub-placeholder"
                :maxlength="300"
              ></textarea>
              <view class="pub-field-count">{{ reason.length }}/300</view>
            </view>
          </view>
        </view>

        <!-- 审核提示 -->
        <view class="pub-review-note">提交后由平台审核，审核通过后按原路退款；可在「售后详情」查看处理进度。</view>
      </template>
    </template>

    <!-- 固定底部操作区（与发布页同款） -->
    <view v-if="!loading && !error && order" class="pub-sticky">
      <view v-if="submitted" class="pub-btn pub-btn--primary" hover-class="pub-btn--active" @tap="goAftersale">
        查看售后进度
      </view>
      <view v-else class="pub-btn pub-btn--primary" hover-class="pub-btn--active" @tap="submitRefund">
        {{ submitting ? '提交中...' : '提交退款申请' }}
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { loadOrder, fmtFen } from '../../utils/orderAdapter'
import { request, getErrorMessage } from '../../utils/request'
import { useSafeTop } from '../../utils/safeTop'

const { topPad, initSafeTop } = useSafeTop(true)

const order = ref(null)
const loading = ref(true)
const error = ref(false)
const submitted = ref(false)
const submitting = ref(false)
const reason = ref('')
let orderId = ''

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
  orderId = id
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

const retryLoad = () => {
  loadData({ id: orderId })
}

onLoad((options) => {
  initSafeTop()
  loadData(options)
})

const submitRefund = async () => {
  if (submitted.value || submitting.value) return
  if (!order.value) return
  submitting.value = true
  try {
    // 真实提交售后单：POST /api/v1/trade-orders/{id}/aftersale
    // 仅退款（退货退款后续扩展），金额默认整单金额；后端状态机 shipped/completed → aftersale
    await request({
      url: '/api/v1/trade-orders/' + encodeURIComponent(order.value.id) + '/aftersale',
      method: 'POST',
      data: {
        aftersale_type: 'refund',
        aftersale_reason: reason.value.trim() || '商品问题申请退款',
        aftersale_desc: reason.value.trim(),
        aftersale_amount_fen: order.value.amount_fen || 0,
      },
    })
    submitted.value = true
  } catch (e) {
    uni.showToast({ title: getErrorMessage(e) || '提交失败，请稍后重试', icon: 'none' })
  } finally {
    submitting.value = false
  }
}

const goAftersale = () => {
  if (!order.value) return
  // 跳到售后详情（真实售后单展示审核状态与进度）
  uni.redirectTo({
    url: `/pages/orders/aftersale?id=${encodeURIComponent(order.value.id)}&type=${order.value.type}`,
  })
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
@import '../../pages/publish/pub-style.css';

.pub-fade { opacity: 0.6; }
.pub-form-intro-h2 {
  font-size: 20px;
  margin: 0 0 4px;
  color: #17212B;
}
.pub-form-intro-p {
  font-size: 12px;
  color: #667085;
  margin: 0;
  line-height: 1.5;
}

/* 只读字段值 */
.pub-field-value {
  font-size: 14px;
  color: #17212B;
  line-height: 1.5;
}
.pub-field-value--accent {
  color: #0A66C2;
  font-weight: 700;
}

/* 字数计数 */
.pub-field-count {
  display: block;
  text-align: right;
  margin-top: 6px;
  font-size: 11px;
  color: #98A2B3;
}

/* 成功态文案 */
.pub-success-title {
  margin: 0 0 6px;
  font-size: 16px;
  font-weight: 750;
  color: #17212B;
}
.pub-success-desc {
  margin: 0;
  font-size: 12px;
  color: #667085;
  line-height: 1.6;
}

/* 加载中 */
.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 80px 0;
  color: #667085;
  font-size: 13px;
}

/* 空态重试按钮 */
.retry-btn {
  flex: none;
  margin: 12px auto 0;
  padding: 0 22px;
}
</style>
