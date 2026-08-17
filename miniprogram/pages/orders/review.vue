<template>
  <view class="pub-page" :style="{ paddingTop: topPad + 'px' }">
    <!-- 顶栏（与发布页同款） -->
    <view class="pub-nav">
      <view class="pub-back" hover-class="pub-fade" @tap="goBack">‹</view>
      <view class="pub-nav-title">{{ navTitle }}</view>
    </view>

    <!-- 加载中 -->
    <view v-if="loading" class="loading-state">
      <u-loading size="24rpx" />
      <text>加载中...</text>
    </view>

    <!-- 加载失败 -->
    <view v-else-if="error || !order" class="pub-empty">
      <view class="pub-empty-mark">!</view>
      <view class="pub-empty-title">订单加载失败</view>
      <view class="pub-btn pub-btn--primary retry-btn" hover-class="pub-btn--active" @tap="retryLoad">重新加载</view>
    </view>

    <template v-else>
      <!-- 提交成功态 -->
      <view v-if="submitted" class="success-card">
        <view class="pub-success-mark">✓</view>
        <text class="success-title">评价提交成功</text>
        <text class="success-desc">结课凭证已存入「我的报名 / 证书」</text>
        <view class="success-actions">
          <view class="pub-btn pub-btn--primary" hover-class="pub-btn--active" @tap="goOrders">返回我的订单</view>
          <view class="pub-btn pub-btn--ghost" hover-class="pub-btn--active" @tap="goCertificates">查看我的证书</view>
        </view>
      </view>

      <!-- 评价表单 -->
      <template v-else>
        <!-- 表单头部：所评课程信息 -->
        <view class="pub-form-intro">
          <view class="pub-form-intro-h2">{{ order.title }}</view>
          <view class="pub-form-intro-p">{{ order.subtitle }}</view>
        </view>

        <!-- 星级 -->
        <view class="pub-section">
          <view class="pub-section-title">课程评价</view>
          <view class="pub-form-card">
            <view class="pub-field">
              <view class="star-row">
                <text
                  v-for="i in 5"
                  :key="i"
                  class="star"
                  :class="{ on: rating >= i }"
                  @tap="rating = i"
                >★</text>
              </view>
              <!-- 1~5 星解释横排，选中星级高亮 -->
              <view class="star-legend">
                <text
                  v-for="(t, i) in ratingText"
                  :key="i"
                  class="star-legend-item"
                  :class="{ on: rating === i + 1 }"
                >{{ i + 1 }} 星 · {{ t }}</text>
              </view>
            </view>
          </view>
        </view>

        <!-- 文字评价 -->
        <view class="pub-section">
          <view class="pub-section-title">文字评价</view>
          <view class="pub-form-card">
            <view class="pub-field">
              <textarea
                class="pub-input pub-input--textarea review-textarea"
                v-model="content"
                placeholder="说说课程内容、讲师讲解与实操安排..."
                placeholder-class="pub-placeholder"
                :maxlength="200"
              ></textarea>
              <text class="textarea-count">{{ content.length }}/200</text>
            </view>
          </view>
        </view>

        <!-- 提示 -->
        <view class="pub-review-note">{{ reviewed ? '该订单已完成评价，可修改后重新提交（覆盖原评价）。' : '评价提交后经平台审核展示，感谢您的反馈。' }}</view>

        <!-- 固定底部操作区（与发布页同款） -->
        <view class="pub-sticky">
          <view class="pub-btn pub-btn--primary" hover-class="pub-btn--active" @tap="submitReview">提交评价</view>
        </view>
      </template>
    </template>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { loadOrder, getReview, submitReview as submitOrderReview } from '../../utils/orderAdapter'
import { useSafeTop } from '../../utils/safeTop'

const { topPad, initSafeTop } = useSafeTop(true)

const order = ref(null)
const loading = ref(true)
const error = ref(false)
const rating = ref(5)
const content = ref('')
const submitted = ref(false)
const reviewed = ref(false)
let orderId = ''

const ratingText = ['极差', '较差', '一般', '满意', '非常满意']

const navTitle = computed(() => {
  if (submitted.value) return '评价成功'
  return reviewed.value ? '修改评价' : '发表评价'
})

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
    else {
      // 已提交过的评价优先回显（本地存储），否则用订单自带的引导默认值
      const saved = getReview(id)
      if (saved) {
        reviewed.value = true
        rating.value = saved.rating
        content.value = saved.content
      } else if (order.value.detail?.review?.rating) {
        rating.value = order.value.detail.review.rating
      } else if (order.value.detail?.review?.default_text) {
        content.value = order.value.detail.review.default_text
      }
    }
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

const submitReview = async () => {
  if (submitted.value) return
  if (!rating.value) {
    uni.showToast({ title: '请选择星级', icon: 'none' })
    return
  }
  try {
    // 真实评价接口 POST /api/v1/reviews（orderAdapter 内处理开发环境本地回退）
    await submitOrderReview(orderId, { rating: rating.value, content: content.value })
    reviewed.value = true
    submitted.value = true
    uni.showToast({ title: '评价成功', icon: 'success' })
  } catch (e) {
    uni.showToast({ title: '评价提交失败，请重试', icon: 'none' })
  }
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

/* 错误重试按钮 */
.retry-btn {
  flex: none;
  margin: 12px auto 0;
  padding: 0 22px;
}

/* 星级（选中橙 #F97316，未选灰 #D7DEE6） */
.star-row {
  display: flex;
  gap: 8px;
  padding: 2px 0 0;
}
.star {
  font-size: 28px;
  color: #D7DEE6;
  line-height: 1;
}
.star.on { color: #F97316; }
.star-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 16px;
  margin-top: 10px;
}
.star-legend-item {
  font-size: 11px;
  color: #98A2B3;
  white-space: nowrap;
}
.star-legend-item.on {
  color: #E96012;
  font-weight: 700;
}

/* 文字评价 */
.review-textarea { min-height: 110px; }
.textarea-count {
  display: block;
  text-align: right;
  margin-top: 6px;
  font-size: 11px;
  color: #98A2B3;
}

/* 提交成功态 */
.success-card {
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 10px;
  padding: 44px 30px;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.success-title {
  margin-top: 16px;
  font-size: 16px;
  font-weight: 750;
  color: #17212B;
}
.success-desc {
  margin-top: 8px;
  font-size: 12px;
  color: #667085;
}
.success-actions {
  margin-top: 26px;
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 9px;
}
.success-actions .pub-btn { flex: none; }
</style>
