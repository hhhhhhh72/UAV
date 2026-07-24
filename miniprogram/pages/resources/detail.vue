<template>
  <view class="resource-detail-page">
    <van-nav-bar
      title="资源详情"
      fixed
      placeholder
      left-arrow
      @click-left="goBack"
    />

    <!-- Loading state -->
    <view v-if="loading" class="state-view">
      <van-loading size="24">加载中...</van-loading>
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg" class="state-view">
      <van-empty description="加载失败" image="error" />
      <view class="retry-btn" @tap="fetchDetail">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Empty state -->
    <view v-else-if="!detail" class="state-view">
      <van-empty description="资源不存在" image="search" />
    </view>

    <!-- Normal state -->
    <template v-else>
      <!-- Hero image -->
      <view class="hero-section">
        <van-image
          v-if="detail.photo_url"
          :src="detail.photo_url"
          width="100%"
          height="220"
          fit="cover"
          round="0"
        />
        <view v-else class="hero-placeholder">
          <text class="placeholder-emoji">{{ resourceEmoji(detail.res_type) }}</text>
        </view>
      </view>

      <!-- Info cards -->
      <van-cell-group inset>
        <van-cell title="资源名称" :value="detail.name || '--'" />
        <van-cell v-if="detail.model" title="型号" :value="detail.model" />
        <van-cell v-if="detail.location" title="所在地">
          <template #value>
            <view class="cell-with-icon">
              <van-icon name="location-o" size="14" color="#1989fa" />
              <text>{{ detail.location }}</text>
            </view>
          </template>
        </van-cell>
        <van-cell title="日租费用">
          <template #value>
            <text class="fee-text">¥{{ detail.daily_fee || 0 }}/天</text>
          </template>
        </van-cell>
        <van-cell v-if="detail.contact" title="联系方式" :value="detail.contact" />
        <van-cell v-if="detail.booking_method" title="预约方式" :value="detail.booking_method" />
      </van-cell-group>

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
      <van-button
        type="primary"
        block
        round
        @tap="showBookingPopup"
      >
        立即预约
      </van-button>
    </view>

    <!-- Booking popup -->
    <van-popup
      :show="bookingPopupVisible"
      position="bottom"
      round
      closeable
      @close="bookingPopupVisible = false"
      custom-style="padding: 24px 16px 32px; max-height: 80vh; overflow-y: auto;"
    >
      <view class="popup-title">预约资源</view>

      <view class="popup-field">
        <text class="popup-label">预约日期 <text class="required">*</text></text>
        <view class="date-picker-wrapper" @tap="showDatePicker = true">
          <text :class="{ placeholder: !bookingForm.date }">
            {{ bookingForm.date || '请选择日期' }}
          </text>
          <van-icon name="arrow" size="12" color="#969799" />
        </view>
      </view>

      <van-field
        v-model="bookingForm.purpose"
        label="用途说明"
        placeholder="请填写用途说明"
        :border="true"
      />

      <van-field
        v-model="bookingForm.contact_name"
        label="联系人"
        placeholder="请填写联系人姓名"
        :border="true"
      />

      <van-field
        v-model="bookingForm.contact_phone"
        label="联系电话"
        type="number"
        placeholder="请填写11位手机号"
        :border="true"
        maxlength="11"
      />

      <view class="popup-actions">
        <van-button type="default" round @tap="bookingPopupVisible = false">取消</van-button>
        <van-button
          type="primary"
          round
          :loading="bookingSubmitting"
          @tap="submitBooking"
        >
          确认预约
        </van-button>
      </view>
    </van-popup>

    <!-- Date picker popup (van-datetime-picker cannot be used directly in uni-app, use native picker) -->
    <van-popup
      :show="showDatePicker"
      position="bottom"
      round
      @close="showDatePicker = false"
    >
      <van-datetime-picker
        type="date"
        :min-date="minDate"
        @confirm="onDateConfirm"
        @cancel="showDatePicker = false"
      />
    </van-popup>
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
      showDatePicker: false,
      minDate: Date.now(),
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
      var timestamp = e.detail
      var d = new Date(timestamp)
      var m = d.getMonth() + 1
      var day = d.getDate()
      this.bookingForm.date = d.getFullYear() + '-' + (m < 10 ? '0' : '') + m + '-' + (day < 10 ? '0' : '') + day
      this.showDatePicker = false
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
        drone: '🚁',
        airport: '🏪',
        test_site: '🏞',
        test_base: '🏗',
      }
      return map[type] || '📋'
    },
  },
}
</script>

<style scoped>
.resource-detail-page {
  min-height: 100vh;
  background: #f7f8fa;
  padding-bottom: 80px;
}

/* State views */
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

/* Hero section */
.hero-section {
  margin-bottom: 12px;
}

.hero-placeholder {
  height: 220px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
}

.placeholder-emoji {
  font-size: 64px;
}

/* Info */
.cell-with-icon {
  display: flex;
  align-items: center;
  gap: 4px;
}

.fee-text {
  color: #ee0a24;
  font-weight: 600;
}

/* Description */
.desc-card {
  margin: 12px;
  padding: 16px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.02);
}

.desc-label {
  font-size: 14px;
  font-weight: 600;
  color: #323233;
  display: block;
  margin-bottom: 8px;
}

.desc-text {
  font-size: 14px;
  color: #333;
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
  background: #fff;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.04);
  padding-bottom: calc(12px + env(safe-area-inset-bottom));
  z-index: 100;
}

/* Popup */
.popup-title {
  font-size: 18px;
  font-weight: 700;
  color: #323233;
  text-align: center;
  margin-bottom: 20px;
}

.popup-field {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: #f7f8fa;
  border-radius: 8px;
  margin-bottom: 12px;
  font-size: 14px;
}

.popup-label {
  color: #323233;
  flex-shrink: 0;
  margin-right: 12px;
}

.required {
  color: #ee0a24;
}

.date-picker-wrapper {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #323233;
}

.date-picker-wrapper .placeholder {
  color: #c8c9cc;
}

.popup-actions {
  display: flex;
  gap: 12px;
  margin-top: 20px;
}

.popup-actions .van-button {
  flex: 1;
}
</style>
