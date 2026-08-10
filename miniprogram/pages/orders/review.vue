<template>
  <view class="review-page">
    <u-nav-bar :title="submitted ? '评价成功' : '发表评价'" show-back @back="goBack" />

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
        <text class="success-title">评价提交成功</text>
        <text class="success-desc">结课凭证已存入「我的报名 / 证书」</text>
        <view class="success-actions">
          <view class="success-btn" @tap="goOrders">
            <text>返回我的订单</text>
          </view>
          <view class="success-btn ghost" @tap="goCertificates">
            <text>查看我的证书</text>
          </view>
        </view>
      </view>

      <!-- 评价表单 -->
      <template v-else>
        <!-- 课程信息 -->
        <view class="course-card">
          <text class="course-title">{{ order.title }}</text>
          <text class="course-sub">{{ order.subtitle }}</text>
        </view>

        <!-- 星级 -->
        <view class="form-card">
          <text class="form-label">课程评价</text>
          <view class="star-row">
            <text
              v-for="i in 5"
              :key="i"
              class="star"
              :class="{ on: rating >= i }"
              @tap="rating = i"
            >★</text>
          </view>
          <text class="star-hint">{{ rating }} 星 · {{ ratingText }}</text>
        </view>

        <!-- 文字评价 -->
        <view class="form-card">
          <text class="form-label">文字评价</text>
          <textarea
            class="review-textarea"
            v-model="content"
            placeholder="说说课程内容、讲师讲解与实操安排..."
            :maxlength="200"
          />
          <text class="textarea-count">{{ content.length }}/200</text>
        </view>

        <!-- 提示 -->
        <view class="note-card">
          <text class="note-text">当前评价接口不校验订单归属与重复评价，属于演示提交，不改变订单状态。</text>
        </view>

        <view class="submit-wrap">
          <view class="submit-btn" @tap="submitReview">
            <text>提交评价</text>
          </view>
        </view>
      </template>
    </template>

    <view class="bottom-spacer"></view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { loadOrder } from '../../utils/orderAdapter'

const order = ref(null)
const loading = ref(true)
const error = ref(false)
const rating = ref(5)
const content = ref('')
const submitted = ref(false)

const ratingText = ['极差', '较差', '一般', '满意', '非常满意']

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
    else if (order.value.detail?.review?.rating) rating.value = order.value.detail.review.rating
    else if (order.value.detail?.review?.default_text) content.value = order.value.detail.review.default_text
  } catch (e) {
    error.value = true
  } finally {
    loading.value = false
  }
}

onLoad(loadData)

const submitReview = async () => {
  if (submitted.value) return
  if (!rating.value) {
    uni.showToast({ title: '请选择星级', icon: 'none' })
    return
  }
  // 第一期为演示提交：评价接口（POST /reviews）不校验订单归属/是否已评价，
  // 不把订单状态改为已评价。这里仅记录本地演示提交结果。
  submitted.value = true
  uni.showToast({ title: '评价提交成功', icon: 'success' })
}

const goOrders = () => {
  uni.reLaunch({ url: '/pages/orders/index' })
}

const goCertificates = () => {
  uni.navigateTo({ url: '/pkg-talent/pages/training/certificates' })
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
.review-page {
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

.course-card {
  margin: 20rpx 24rpx 0;
  background: #fff;
  border: 1rpx solid var(--color-border);
  border-radius: 16rpx;
  box-shadow: var(--shadow-sm);
  padding: 28rpx;
}
.course-title {
  display: block;
  font-size: 28rpx;
  font-weight: 700;
  color: var(--color-text);
  line-height: 1.4;
}
.course-sub {
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
.star-row {
  display: flex;
  gap: 16rpx;
  margin-top: 24rpx;
}
.star {
  font-size: 56rpx;
  color: var(--color-divider);
  line-height: 1;
}
.star.on {
  color: var(--color-accent);
}
.star-hint {
  display: block;
  margin-top: 16rpx;
  font-size: 22rpx;
  color: var(--color-accent-deep);
}

.review-textarea {
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

/* 成功态 */
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
}
.success-actions {
  margin-top: 32rpx;
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 16rpx;
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
.success-btn.ghost {
  background: var(--color-primary-light);
  color: var(--color-primary);
}

.bottom-spacer { height: 24rpx; }
</style>
