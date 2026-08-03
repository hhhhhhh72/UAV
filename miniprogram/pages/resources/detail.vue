<template>
  <view class="resource-detail-page">
    <u-nav-bar
      title="资源详情"
      show-back
      @back="goBack"
    />

    <!-- Loading state -->
    <view v-if="loading" class="state-view">
      <view class="loading-inline">
        <u-loading size="24rpx" />
        <text>加载中...</text>
      </view>
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg" class="state-view">
      <u-empty description="加载失败" />
      <view class="retry-btn" @tap="fetchDetail">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Empty state -->
    <view v-else-if="!detail" class="state-view">
      <u-empty description="资源不存在" />
    </view>

    <!-- Normal state -->
    <template v-else>
      <!-- Hero image -->
      <view class="hero-section">
        <image
          v-if="detail.photo_url"
          :src="detail.photo_url"
          mode="aspectFill"
          class="hero-img"
        />
        <view v-else class="hero-placeholder">
          <text class="placeholder-emoji">{{ resourceEmoji(detail.res_type) }}</text>
        </view>
      </view>

      <!-- Info cards -->
      <u-cell-group inset>
        <u-cell title="资源名称" :value="detail.name || '--'" />
        <u-cell v-if="detail.model" title="型号" :value="detail.model" />
        <u-cell v-if="detail.location" title="所在地">
          <template #value>
            <view class="cell-with-icon">
              <u-icon name="location" size="26rpx" color="var(--color-primary)" />
              <text>{{ detail.location }}</text>
            </view>
          </template>
        </u-cell>
        <u-cell title="日租费用">
          <template #value>
            <text class="fee-text">¥{{ detail.daily_fee || 0 }}/天</text>
          </template>
        </u-cell>
        <u-cell v-if="detail.contact" title="联系方式" :value="detail.contact" />
        <u-cell v-if="detail.booking_method" title="预约方式" :value="detail.booking_method" />
      </u-cell-group>

      <!-- Description -->
      <view v-if="detail.description" class="desc-card">
        <text class="desc-label">资源描述</text>
        <text class="desc-text">{{ detail.description }}</text>
      </view>

      <!-- Required margin for bottom button -->
      <view class="bottom-spacer" />
    </template>

    <!-- Fixed bottom button -->
    <view v-if="detail" class="action-bar">
      <u-button
        type="primary"
        block
        round
        @click="showBookingPopup"
      >
        立即预约
      </u-button>
    </view>

    <!-- Booking popup -->
    <u-popup
      :show="bookingPopupVisible"
      position="bottom"
      round
      show-close
      @close="bookingPopupVisible = false"
    >
      <view class="popup-body">
        <view class="popup-title">预约资源</view>

        <view class="popup-field">
          <text class="popup-label">预约日期 <text class="required">*</text></text>
          <picker mode="date" :value="bookingForm.date" :start="todayStr" @change="onDateConfirm">
            <view class="date-picker-wrapper">
              <text :class="{ placeholder: !bookingForm.date }">
                {{ bookingForm.date || '请选择日期' }}
              </text>
              <text class="field-arrow">›</text>
            </view>
          </picker>
        </view>

        <u-field
          v-model="bookingForm.purpose"
          label="用途说明"
          placeholder="请填写用途说明"
        />

        <u-field
          v-model="bookingForm.contact_name"
          label="联系人"
          placeholder="请填写联系人姓名"
        />

        <u-field
          v-model="bookingForm.contact_phone"
          label="联系电话"
          type="number"
          placeholder="请填写11位手机号"
        />

        <view class="popup-actions">
          <u-button type="default" round class="action-btn" @click="bookingPopupVisible = false">取消</u-button>
          <u-button
            type="primary"
            round
            class="action-btn"
            :loading="bookingSubmitting"
            @click="submitBooking"
          >
            确认预约
          </u-button>
        </view>
      </view>
    </u-popup>
  </view>
</template>

<script>
import { request, getStoredUser } from '../../utils/request'

export default {
  data() {
    return {
      id: '',
      loading: false,
      errorMsg: '',
      detail: null,
      bookingPopupVisible: false,
      bookingSubmitting: false,
      // 预约日期可选择的最小日期（今天），限制不可预约过去日期
      todayStr: (function () {
        var d = new Date()
        return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0')
      })(),
      bookingForm: {
        date: '',
        purpose: '',
        contact_name: '',
        contact_phone: '',
      },
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
        var res = await request({
          url: '/api/v1/industry-resources/' + encodeURIComponent(this.id),
        })
        this.detail = (res && res.data) || res || null
      } catch (e) {
        this.errorMsg = '网络异常，请稍后重试'
      } finally {
        this.loading = false
      }
    },
    showBookingPopup() {
      var user = getStoredUser()
      if (!user) {
        uni.showModal({
          title: '提示',
          content: '请先登录后预约',
          confirmText: '去登录',
          success: function (modalRes) {
            if (modalRes.confirm) {
              uni.navigateTo({ url: '/pages/login/index' })
            }
          },
        })
        return
      }
      // Pre-fill user info
      this.bookingForm.contact_name = user.name || ''
      this.bookingForm.contact_phone = user.phone || ''
      this.bookingForm.purpose = ''
      this.bookingPopupVisible = true
    },
    onDateConfirm(e) {
      // 原生 picker mode=date 直接返回 YYYY-MM-DD 字符串
      this.bookingForm.date = e.detail.value
    },
    async submitBooking() {
      // Validation
      if (!this.bookingForm.date) {
        uni.showToast({ title: '请选择预约日期', icon: 'none' })
        return
      }
      if (!this.bookingForm.contact_name) {
        uni.showToast({ title: '请填写联系人', icon: 'none' })
        return
      }
      if (!this.bookingForm.contact_phone) {
        uni.showToast({ title: '请填写联系电话', icon: 'none' })
        return
      }
      if (!/^\d{11}$/.test(this.bookingForm.contact_phone)) {
        uni.showToast({ title: '请输入11位手机号', icon: 'none' })
        return
      }

      this.bookingSubmitting = true
      try {
        await request({
          url: '/api/v1/industry-resources/' + encodeURIComponent(this.id) + '/book',
          method: 'POST',
          data: {
            date: this.bookingForm.date,
            purpose: this.bookingForm.purpose,
            contact_name: this.bookingForm.contact_name,
            contact_phone: this.bookingForm.contact_phone,
            resource_id: this.id,
          },
        })
        uni.showToast({ title: '预约成功', icon: 'success' })
        this.bookingPopupVisible = false
      } catch (e) {
        uni.showToast({ title: '预约失败，请稍后重试', icon: 'none' })
      } finally {
        this.bookingSubmitting = false
      }
    },
    goBack() {
      uni.navigateBack()
    },
    resourceEmoji(type) {
      var map = {
        drone: '机',
        airport: '场',
        test_site: '地',
        test_base: '基',
      }
      return map[type] || '源'
    },
  },
}
</script>

<style scoped>
.resource-detail-page {
  min-height: 100vh;
  background: var(--color-bg);
  padding-bottom: 80px;
}

/* State views */
.state-view {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 120px;
}

.loading-inline {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: var(--color-text-secondary);
}

.retry-btn {
  margin-top: 12px;
  padding: 8px 24px;
  background: var(--color-primary);
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}

/* Hero section */
.hero-section {
  margin-bottom: 12px;
}

.hero-img {
  width: 100%;
  height: 440rpx;
  display: block;
}

.hero-placeholder {
  height: 440rpx;
  background: linear-gradient(135deg, var(--color-primary), #1565c0);
  display: flex;
  align-items: center;
  justify-content: center;
}

.placeholder-emoji {
  font-size: 80rpx;
  font-weight: 600;
  color: #ffffff;
}

/* Info */
.cell-with-icon {
  display: flex;
  align-items: center;
  gap: 4px;
}

.fee-text {
  color: var(--color-danger);
  font-weight: 600;
}

/* Description */
.desc-card {
  margin: 12px;
  padding: 16px;
  background: var(--color-bg-card);
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.02);
}

.desc-label {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
  display: block;
  margin-bottom: 8px;
}

.desc-text {
  font-size: 14px;
  color: var(--color-text);
  line-height: 1.6;
  display: block;
  white-space: pre-wrap;
}

/* Spacer */
.bottom-spacer {
  height: 80px;
}

/* Action bar */
.action-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 12px 16px;
  background: var(--color-bg-card);
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.04);
  padding-bottom: calc(12px + env(safe-area-inset-bottom));
  z-index: 100;
}

/* Popup */
.popup-body {
  padding: 24px 16px 32px;
  max-height: 80vh;
  overflow-y: auto;
}

.popup-title {
  font-size: 18px;
  font-weight: 700;
  color: var(--color-text);
  text-align: center;
  margin-bottom: 20px;
}

.popup-field {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--color-bg);
  border-radius: 8px;
  margin-bottom: 12px;
  font-size: 14px;
}

.popup-label {
  color: var(--color-text);
  flex-shrink: 0;
  margin-right: 12px;
}

.required {
  color: var(--color-danger);
}

.date-picker-wrapper {
  display: flex;
  align-items: center;
  gap: 4px;
  color: var(--color-text);
}

.date-picker-wrapper .placeholder {
  color: var(--color-text-placeholder);
}

.field-arrow {
  color: var(--color-text-placeholder);
  font-size: 20px;
}

.popup-actions {
  display: flex;
  gap: 12px;
  margin-top: 20px;
}

.popup-actions .action-btn {
  flex: 1;
}
</style>
