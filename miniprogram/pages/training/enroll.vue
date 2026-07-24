<template>
  <view class="enroll-page">
    <!-- Nav -->
    <van-nav-bar
      title="课程详情"
      fixed
      placeholder
      left-arrow
      @click-left="goBack"
    />

    <!-- Loading state -->
    <view v-if="loading" class="loading-state">
      <van-loading size="24">加载中...</van-loading>
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg" class="state-view">
      <van-empty description="加载失败" image="error" />
      <view class="retry-btn" @tap="fetchDetail">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Normal state -->
    <template v-else-if="detail">
      <!-- Header: course image & basic info -->
      <view class="course-header">
        <van-image
          v-if="detail.cover_image || detail.image"
          :src="detail.cover_image || detail.image"
          width="100%"
          height="200"
          fit="cover"
          class="course-banner"
        />
        <view class="header-info">
          <text class="course-title">{{ detail.title }}</text>
          <view class="tags-row">
            <van-tag :type="courseTypeTag(detail.course_type)" size="medium">
              {{ detail.course_type || '通用' }}
            </van-tag>
          </view>
          <view class="info-rows">
            <view class="info-row">
              <text class="info-label">价格</text>
              <text class="info-value price">{{ formatPrice(detail.price_fen) }}</text>
            </view>
            <view v-if="detail.duration" class="info-row">
              <text class="info-label">时长</text>
              <text class="info-value">{{ detail.duration }}</text>
            </view>
            <view v-if="detail.schedule" class="info-row">
              <text class="info-label">开课时间</text>
              <text class="info-value">{{ detail.schedule }}</text>
            </view>
            <view class="info-row">
              <text class="info-label">已报名</text>
              <text class="info-value">{{ detail.enrollment_count || 0 }} 人</text>
            </view>
          </view>
        </view>
      </view>

      <!-- Description -->
      <view v-if="detail.description" class="section">
        <van-cell-group inset>
          <van-cell title="课程介绍">
            <template #default>
              <view class="desc-content">{{ detail.description }}</view>
            </template>
          </van-cell>
        </van-cell-group>
      </view>

      <!-- Syllabus / additional info -->
      <view v-if="detail.syllabus && detail.syllabus.length > 0" class="section">
        <van-cell-group inset>
          <view class="section-title">课程大纲</view>
          <van-cell
            v-for="(item, idx) in detail.syllabus"
            :key="idx"
            :title="(idx + 1) + '. ' + (item.title || item)"
          />
        </van-cell-group>
      </view>

      <!-- Bottom action -->
      <view class="action-bar">
        <van-button
          v-if="!isEnrolled"
          type="primary"
          block
          round
          :loading="enrolling"
          @tap="showPayConfirm"
        >
          立即报名
        </van-button>
        <van-button
          v-else
          type="default"
          block
          round
          disabled
        >
          已报名
        </van-button>
      </view>
    </template>
  </view>
</template>

<script>
import { request } from '../../utils/request'

export default {
  data() {
    return {
      id: '',
      loading: false,
      errorMsg: '',
      detail: null,
      enrolling: false,
      isEnrolled: false,
    }
  },
  onLoad(options) {
    this.id = options.id || ''
    this.fetchDetail()
  },
  methods: {
    async fetchDetail() {
      this.loading = true
      this.errorMsg = ''

      try {
        var res = await request({ url: '/api/v1/training-courses/' + encodeURIComponent(this.id) })
        this.detail = (res && res.data) || res || null
        if (this.detail && this.detail.title) {
          uni.setNavigationBarTitle({ title: this.detail.title })
        }
      } catch (e) {
        this.errorMsg = '网络异常，请稍后重试'
      } finally {
        this.loading = false
      }
    },
    showPayConfirm() {
      var priceText = this.formatPrice(this.detail && this.detail.price_fen)
      var that = this
      uni.showModal({
        title: '确认报名',
        content: '课程价格：' + priceText + '\n确认支付并报名？',
        confirmText: '确认支付',
        cancelText: '取消',
        success: function (res) {
          if (res.confirm) {
            that.doPayAndEnroll()
          }
        },
      })
    },
    async doPayAndEnroll() {
      this.enrolling = true
      try {
        await request({
          url: '/api/v1/training-courses/' + encodeURIComponent(this.id) + '/pay-and-enroll',
          method: 'POST',
        })
        uni.showToast({ title: '报名成功', icon: 'success' })
        setTimeout(function () {
          uni.navigateBack()
        }, 1200)
      } catch (e) {
        uni.showToast({ title: '报名失败，请重试', icon: 'none' })
      } finally {
        this.enrolling = false
      }
    },
    goBack() {
      uni.navigateBack()
    },
    courseTypeTag(type) {
      var map = {
        'CAAC': 'primary',
        'UTC': 'success',
        '人社': 'warning',
        '飞手': 'danger',
      }
      return map[type] || 'default'
    },
    formatPrice(fen) {
      if (fen == null || fen === 0) return '免费'
      var yuan = (fen / 100).toFixed(2)
      return '¥' + yuan.replace(/\.00$/, '')
    },
  },
}
</script>

<style scoped>
.enroll-page {
  min-height: 100vh;
  background: #f7f8fa;
  padding-bottom: 80px;
}

.loading-state {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}

.state-view {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 120px;
}

.retry-btn {
  margin-top: 12px;
  padding: 8px 24px;
  background: #1989fa;
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}

/* Course header */
.course-header {
  background: #fff;
  margin: 0 12px 12px;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.02);
}

.course-banner {
  display: block;
}

.header-info {
  padding: 16px;
}

.course-title {
  font-size: 18px;
  font-weight: 700;
  color: #323233;
  display: block;
  margin-bottom: 10px;
  line-height: 1.4;
}

.tags-row {
  display: flex;
  gap: 8px;
  margin-bottom: 14px;
}

.info-rows {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.info-row {
  display: flex;
  align-items: center;
}

.info-label {
  width: 60px;
  font-size: 14px;
  color: #969799;
  flex-shrink: 0;
}

.info-value {
  font-size: 14px;
  color: #323233;
}

.info-value.price {
  color: #ee0a24;
  font-weight: 600;
  font-size: 16px;
}

/* Sections */
.section {
  margin-bottom: 12px;
}

.section-title {
  padding: 12px 16px 4px;
  font-size: 15px;
  font-weight: 600;
  color: #323233;
}

.desc-content {
  font-size: 14px;
  color: #333;
  line-height: 1.7;
  white-space: pre-wrap;
  padding: 8px 0;
}

/* Bottom action bar */
.action-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 12px 16px;
  background: #fff;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.04);
  padding-bottom: calc(12px + env(safe-area-inset-bottom));
  z-index: 100;
}
</style>
