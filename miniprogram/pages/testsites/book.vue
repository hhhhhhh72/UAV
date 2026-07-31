<template>
  <view class="page-container">
    <!-- Nav -->
    <van-nav-bar
      title="场地预约"
      left-arrow
      @click-left="goBack"
      custom-style="background: #06b6d4;"
    >
      <template #title>
        <text style="color: #fff;">场地预约</text>
      </template>
    </van-nav-bar>

    <!-- Site list (step 1: choose site) -->
    <template v-if="!selectedSite">
      <!-- Loading -->
      <view v-if="loadingSites" class="loading-state">
        <van-loading size="24">加载场地列表...</van-loading>
      </view>

      <!-- Error -->
      <view v-else-if="sitesError" class="error-state">
        <van-empty image="network" description="加载失败" />
        <view class="retry-btn" @tap="fetchSites">
          <text>重新加载</text>
        </view>
      </view>

      <!-- Empty -->
      <view v-else-if="sites.length === 0" class="empty-state-wrapper">
        <van-empty image="search" description="暂无可用场地" />
      </view>

      <!-- Site list -->
      <view v-else class="list-body">
        <van-cell-group inset>
          <van-cell
            v-for="site in sites"
            :key="site.id"
            :title="site.name"
            is-link
            @tap="selectSite(site)"
          >
            <template #label>
              <view class="cell-meta">
                <text v-if="site.location" class="meta-text">{{ site.location }}</text>
                <text v-if="site.facilities" class="meta-text">{{ site.facilities }}</text>
              </view>
              <view v-if="site.fee_desc" class="cell-extra">
                <text class="fee-text">{{ site.fee_desc }}</text>
              </view>
            </template>
          </van-cell>
        </van-cell-group>
      </view>
    </template>

    <!-- Site detail + booking form (step 2) -->
    <template v-else>
      <!-- Loading detail -->
      <view v-if="loadingDetail" class="loading-state">
        <van-loading size="24">加载场地详情...</van-loading>
      </view>

      <!-- Error -->
      <view v-else-if="detailError" class="error-state">
        <van-empty image="network" description="加载失败" />
        <view class="retry-btn" @tap="fetchSiteDetail">
          <text>重新加载</text>
        </view>
      </view>

      <!-- Detail + form -->
      <view v-else class="booking-body">
        <!-- Site info card -->
        <van-cell-group inset>
          <van-cell title="场地名称" :value="selectedSite.name" />
          <van-cell v-if="selectedSite.location" title="位置" :label="selectedSite.location" />
          <van-cell v-if="detail.facilities" title="设施" :value="detail.facilities" />
          <van-cell v-if="detail.fee_desc" title="费用" :value="detail.fee_desc" />
        </van-cell-group>

        <!-- Date picker -->
        <view class="section-title">选择日期</view>
        <view class="date-picker-wrap">
          <picker
            mode="date"
            :value="bookingDate"
            :start="minDate"
            @change="onDateChange"
          >
            <view class="date-display">
              <text class="date-label">预约日期</text>
              <text class="date-value">{{ bookingDate || '请选择日期' }}</text>
              <van-icon name="arrow" size="14" color="#969799" />
            </view>
          </picker>
        </view>

        <!-- Time slots -->
        <view v-if="timeSlots.length > 0" class="section">
          <view class="section-title">选择时段</view>
          <van-grid :column-num="3" gutter="8" clickable>
            <van-grid-item
              v-for="slot in timeSlots"
              :key="slot.value || slot"
              :text="slot.label || slot"
              custom-class="slot-item"
              :custom-style="getSlotStyle(slot)"
              @tap="selectSlot(slot)"
            />
          </van-grid>
        </view>

        <!-- Booking form -->
        <view class="section">
          <view class="section-title">预约信息</view>
          <van-cell-group inset>
            <van-field
              v-model="form.purpose"
              label="用途"
              placeholder="请输入预约用途"
              :border="true"
            />
            <van-field
              v-model="form.contact_name"
              label="联系人"
              placeholder="请输入联系人姓名"
              :border="true"
            />
            <van-field
              v-model="form.contact_phone"
              label="联系电话"
              type="number"
              placeholder="请输入联系电话"
              :border="false"
            />
          </van-cell-group>
        </view>

        <!-- Submit -->
        <view class="submit-section">
          <van-button
            type="primary"
            block
            round
            :loading="submitting"
            custom-style="background: #06b6d4; border-color: #06b6d4;"
            @tap="handleSubmit"
          >
            提交预约
          </van-button>
        </view>
      </view>
    </template>
  </view>
</template>

<script>
import { request } from '../../utils/request'

export default {
  data() {
    var now = new Date()
    var y = now.getFullYear()
    var m = now.getMonth() + 1
    var d = now.getDate()
    return {
      // Site list
      sites: [],
      loadingSites: false,
      sitesError: '',

      // Selected site
      selectedSite: null,
      detail: {},
      loadingDetail: false,
      detailError: '',

      // Booking
      bookingDate: y + '-' + (m < 10 ? '0' : '') + m + '-' + (d < 10 ? '0' : '') + d,
      minDate: y + '-' + (m < 10 ? '0' : '') + m + '-' + (d < 10 ? '0' : '') + d,
      timeSlots: [],
      selectedSlot: null,
      form: {
        purpose: '',
        contact_name: '',
        contact_phone: '',
      },
      submitting: false,
    }
  },
  onLoad() {
    this.fetchSites()
  },
  methods: {
    async fetchSites() {
      this.loadingSites = true
      this.sitesError = ''
      try {
        var res = await request({ url: '/api/v1/test-sites' })
        var data = Array.isArray(res) ? res : (res && res.data) || res || {}
        this.sites = Array.isArray(data) ? data : (data && data.items) || []
      } catch (e) {
        this.sitesError = '网络异常，请稍后重试'
      } finally {
        this.loadingSites = false
      }
    },

    async selectSite(site) {
      this.selectedSite = site
      this.selectedSlot = null
      this.form = { purpose: '', contact_name: '', contact_phone: '' }
      await this.fetchSiteDetail()
    },

    async fetchSiteDetail() {
      if (!this.selectedSite) return
      this.loadingDetail = true
      this.detailError = ''
      try {
        var res = await request({ url: '/api/v1/test-sites/' + encodeURIComponent(this.selectedSite.id) })
        this.detail = (res && res.data) || res || {}
        // Parse time slots from detail
        var slots = this.detail.available_slots || []
        this.timeSlots = Array.isArray(slots) ? slots : []
      } catch (e) {
        this.detailError = '网络异常，请稍后重试'
      } finally {
        this.loadingDetail = false
      }
    },

    onDateChange(e) {
      this.bookingDate = e.detail.value
      this.selectedSlot = null
      // Re-fetch detail for new date's availability
      this.fetchSiteDetail()
    },

    getSlotStyle(slot) {
      var status = slot.status || 'available'
      if (status === 'booked') {
        return 'background: #c8c9cc; color: #fff;'
      }
      if (this.selectedSlot && (this.selectedSlot.value || this.selectedSlot) === (slot.value || slot)) {
        return 'background: #06b6d4; color: #fff; border: 2px solid #06b6d4;'
      }
      return 'background: #34c759; color: #fff;'
    },

    selectSlot(slot) {
      var status = slot.status || 'available'
      if (status === 'booked') return
      this.selectedSlot = slot
    },

    backToList() {
      this.selectedSite = null
      this.detail = {}
      this.selectedSlot = null
    },

    async handleSubmit() {
      if (!this.selectedSite) {
        uni.showToast({ title: '请选择场地', icon: 'none' })
        return
      }
      if (!this.bookingDate) {
        uni.showToast({ title: '请选择日期', icon: 'none' })
        return
      }
      if (!this.selectedSlot) {
        uni.showToast({ title: '请选择时段', icon: 'none' })
        return
      }
      if (!this.form.purpose) {
        uni.showToast({ title: '请填写用途', icon: 'none' })
        return
      }
      if (!this.form.contact_name) {
        uni.showToast({ title: '请填写联系人', icon: 'none' })
        return
      }
      if (!this.form.contact_phone) {
        uni.showToast({ title: '请填写联系电话', icon: 'none' })
        return
      }

      this.submitting = true
      try {
        var payload = {
          date: this.bookingDate,
          time_slot: this.selectedSlot.value || this.selectedSlot,
          purpose: this.form.purpose,
          contact_name: this.form.contact_name,
          contact_phone: this.form.contact_phone,
        }
        await request({
          url: '/api/v1/test-sites/' + encodeURIComponent(this.selectedSite.id) + '/book',
          method: 'POST',
          data: payload,
        })
        uni.showToast({ title: '预约成功', icon: 'success' })
        var that = this
        setTimeout(function () {
          uni.navigateBack()
        }, 1200)
      } catch (e) {
        uni.showToast({ title: '预约失败，请重试', icon: 'none' })
      } finally {
        this.submitting = false
      }
    },

    goBack() {
      uni.navigateBack()
    },
  },
}
</script>

<style scoped>
.page-container {
  min-height: 100vh;
  background: #f7f8fa;
  padding-bottom: 40px;
}

.loading-state {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}

.empty-state-wrapper {
  padding-top: 60px;
}

.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 120px;
}

.retry-btn {
  margin-top: 12px;
  padding: 8px 24px;
  background: #06b6d4;
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}

.list-body {
  padding: 12px 0 24px;
}

.cell-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 4px;
}

.cell-extra {
  margin-top: 4px;
}

.meta-text {
  font-size: 12px;
  color: #969799;
}

.fee-text {
  font-size: 13px;
  color: #ee0a24;
  font-weight: 500;
}

/* Booking body */
.booking-body {
  padding: 12px 0 24px;
}

.section {
  margin-bottom: 4px;
}

.section-title {
  padding: 16px 16px 8px;
  font-size: 14px;
  font-weight: 600;
  color: #323233;
}

/* Date picker */
.date-picker-wrap {
  margin: 0 12px;
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
}

.date-display {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
}

.date-label {
  font-size: 14px;
  color: #323233;
}

.date-value {
  font-size: 14px;
  color: #0A66C2;
  margin-left: auto;
  margin-right: 8px;
}

/* Slot grid */
.slot-item {
  border-radius: 8px;
  font-size: 13px;
}

/* Submit */
.submit-section {
  padding: 20px 16px 40px;
}
</style>
