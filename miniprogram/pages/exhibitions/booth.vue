<template>
  <view class="page-container">
    <!-- Nav -->
    <van-nav-bar
      title="展位申请"
      left-arrow
      @click-left="goBack"
    />

    <!-- Loading state -->
    <view v-if="loading" class="loading-state">
      <van-loading size="24">加载中...</van-loading>
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg" class="error-state">
      <van-empty image="network" description="加载失败" />
      <view class="retry-btn" @tap="fetchDetail">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Normal state -->
    <template v-else-if="detail">
      <!-- Exhibition summary card -->
      <view class="summary-card">
        <text class="expo-title">{{ detail.title }}</text>
        <view class="expo-info">
          <view v-if="detail.dates || detail.start_date" class="info-row">
            <text class="info-label">时间</text>
            <text class="info-value">{{ formatDateRange(detail) }}</text>
          </view>
          <view v-if="detail.location" class="info-row">
            <text class="info-label">地点</text>
            <text class="info-value">{{ detail.location }}</text>
          </view>
        </view>
      </view>

      <!-- Floor plan -->
      <view v-if="detail.floor_plan_image" class="section">
        <view class="section-title">展位平面图</view>
        <view class="floor-plan-wrap">
          <van-image
            :src="detail.floor_plan_image"
            width="100%"
            fit="widthFix"
            class="floor-plan-img"
            @tap="previewFloorPlan"
          />
        </view>
      </view>

      <!-- Booth selection grid -->
      <view v-if="booths.length > 0" class="section">
        <view class="section-title">
          选择展位
          <text class="legend">
            <text class="legend-dot available-dot"></text>可选
            <text class="legend-dot booked-dot"></text>已订
            <text class="legend-dot selected-dot"></text>已选
          </text>
        </view>
        <view class="booth-grid">
          <view
            v-for="booth in booths"
            :key="booth.id || booth.number || booth"
            class="booth-item"
            :class="getBoothClass(booth)"
            @tap="selectBooth(booth)"
          >
            <text class="booth-num">{{ booth.number || booth.name || booth }}</text>
          </view>
        </view>
      </view>

      <!-- Application form -->
      <view class="section">
        <view class="section-title">申请信息</view>
        <van-cell-group inset>
          <van-field
            v-model="form.company_name"
            label="企业名称"
            placeholder="请输入企业名称"
            :border="true"
          />
          <van-field
            v-model="form.contact"
            label="联系人"
            placeholder="请输入联系人"
            :border="true"
          />
          <van-field
            v-model="form.contact_phone"
            label="联系电话"
            type="number"
            placeholder="请输入联系电话"
            :border="true"
          />
          <van-field
            v-model="form.booth_type"
            label="展位类型"
            placeholder="请选择展位类型"
            :border="true"
            readonly
            right-icon="arrow"
            @tap="showBoothTypePicker"
          />
          <van-field
            v-model="form.requirements"
            label="特殊需求"
            type="textarea"
            placeholder="请输入特殊需求（选填）"
            :border="false"
            autosize
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
          @tap="handleSubmit"
        >
          提交申请
        </van-button>
      </view>
    </template>

    <!-- Booth type picker popup -->
    <van-popup
      :show="boothTypePickerVisible"
      position="bottom"
      round
      @close="boothTypePickerVisible = false"
    >
      <view class="picker-popup">
        <view
          v-for="type in boothTypes"
          :key="type"
          class="picker-option"
          :class="{ active: form.booth_type === type }"
          @tap="selectBoothType(type)"
        >
          <text>{{ type }}</text>
          <text v-if="form.booth_type === type" class="check-icon">&#10003;</text>
        </view>
        <view class="picker-cancel" @tap="boothTypePickerVisible = false">
          <text>取消</text>
        </view>
      </view>
    </van-popup>
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
      booths: [],
      selectedBooth: null,

      form: {
        company_name: '',
        contact: '',
        contact_phone: '',
        booth_type: '',
        requirements: '',
      },
      submitting: false,

      boothTypePickerVisible: false,
      boothTypes: ['标准', '豪华', '特装'],
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
        var res = await request({ url: '/api/v1/exhibitions/' + encodeURIComponent(this.id) })
        this.detail = (res && res.data) || res || null
        if (this.detail) {
          var boothList = this.detail.booths || this.detail.available_booths || []
          this.booths = Array.isArray(boothList) ? boothList : []
          if (this.detail.title) {
            uni.setNavigationBarTitle({ title: this.detail.title + ' - 展位申请' })
          }
        }
      } catch (e) {
        this.errorMsg = '网络异常，请稍后重试'
      } finally {
        this.loading = false
      }
    },

    getBoothClass(booth) {
      var status = booth.status || 'available'
      if (status === 'booked' || status === 'reserved') return 'booth-booked'
      if (this.selectedBooth && this.selectedBooth.id === booth.id) return 'booth-selected'
      if (this.selectedBooth && (this.selectedBooth.number || this.selectedBooth.name) === (booth.number || booth.name)) return 'booth-selected'
      return 'booth-available'
    },

    selectBooth(booth) {
      var status = booth.status || 'available'
      if (status === 'booked' || status === 'reserved') return
      this.selectedBooth = booth
    },

    showBoothTypePicker() {
      this.boothTypePickerVisible = true
    },

    selectBoothType(type) {
      this.form.booth_type = type
      this.boothTypePickerVisible = false
    },

    async handleSubmit() {
      if (!this.selectedBooth) {
        uni.showToast({ title: '请选择展位', icon: 'none' })
        return
      }
      if (!this.form.company_name) {
        uni.showToast({ title: '请填写企业名称', icon: 'none' })
        return
      }
      if (!this.form.contact) {
        uni.showToast({ title: '请填写联系人', icon: 'none' })
        return
      }
      if (!this.form.contact_phone) {
        uni.showToast({ title: '请填写联系电话', icon: 'none' })
        return
      }
      if (!this.form.booth_type) {
        uni.showToast({ title: '请选择展位类型', icon: 'none' })
        return
      }

      this.submitting = true
      try {
        var payload = {
          booth_id: this.selectedBooth.id || this.selectedBooth.number || this.selectedBooth.name,
          booth_number: this.selectedBooth.number || this.selectedBooth.name,
          company_name: this.form.company_name,
          contact: this.form.contact,
          contact_phone: this.form.contact_phone,
          booth_type: this.form.booth_type,
          requirements: this.form.requirements,
        }
        await request({
          url: '/api/v1/exhibitions/' + encodeURIComponent(this.id) + '/booths',
          method: 'POST',
          data: payload,
        })
        uni.showToast({ title: '申请提交成功', icon: 'success' })
        var that = this
        setTimeout(function () {
          uni.navigateBack()
        }, 1200)
      } catch (e) {
        uni.showToast({ title: '提交失败，请重试', icon: 'none' })
      } finally {
        this.submitting = false
      }
    },

    previewFloorPlan() {
      if (this.detail && this.detail.floor_plan_image) {
        uni.previewImage({
          urls: [this.detail.floor_plan_image],
          current: this.detail.floor_plan_image,
        })
      }
    },

    formatDateRange(item) {
      if (item.dates) return item.dates
      var start = item.start_date || ''
      var end = item.end_date || ''
      if (start && end) return start + ' ~ ' + end
      return start || end || ''
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

.error-state {
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

/* Summary card */
.summary-card {
  background: #fff;
  margin: 12px;
  border-radius: 12px;
  padding: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.03);
}

.expo-title {
  font-size: 18px;
  font-weight: 700;
  color: #323233;
  display: block;
  margin-bottom: 12px;
  line-height: 1.4;
}

.expo-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.info-row {
  display: flex;
  align-items: center;
}

.info-label {
  width: 48px;
  font-size: 13px;
  color: #969799;
  flex-shrink: 0;
}

.info-value {
  font-size: 13px;
  color: #323233;
}

/* Section */
.section {
  margin-bottom: 12px;
}

.section-title {
  padding: 12px 16px 4px;
  font-size: 14px;
  font-weight: 600;
  color: #323233;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

/* Floor plan */
.floor-plan-wrap {
  margin: 0 12px;
  border-radius: 8px;
  overflow: hidden;
}

.floor-plan-img {
  display: block;
}

/* Booth grid */
.booth-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  padding: 0 12px;
  margin-top: 8px;
}

.booth-item {
  width: calc((100% - 30px) / 4);
  aspect-ratio: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  border: 2px solid transparent;
  transition: all 0.2s;
}

.booth-available {
  background: #34c759;
  color: #fff;
}

.booth-booked {
  background: #c8c9cc;
  color: #fff;
}

.booth-selected {
  background: #34c759;
  color: #fff;
  border-color: #1989fa;
  box-shadow: 0 0 0 2px rgba(25, 137, 250, 0.3);
}

.booth-num {
  font-size: 12px;
  font-weight: 500;
}

/* Legend */
.legend {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  font-weight: 400;
  color: #969799;
}

.legend-dot {
  display: inline-block;
  width: 10px;
  height: 10px;
  border-radius: 2px;
  margin-right: 2px;
}

.available-dot {
  background: #34c759;
}

.booked-dot {
  background: #c8c9cc;
}

.selected-dot {
  background: #1989fa;
}

/* Submit */
.submit-section {
  padding: 20px 16px 40px;
}

/* Picker popup */
.picker-popup {
  padding: 0 0 20px;
}

.picker-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  font-size: 16px;
  color: #323233;
  border-bottom: 1px solid #f2f3f5;
}

.picker-option.active {
  color: #1989fa;
  font-weight: 600;
}

.check-icon {
  color: #1989fa;
  font-weight: 700;
}

.picker-cancel {
  text-align: center;
  padding: 16px 20px;
  font-size: 16px;
  color: #969799;
  margin-top: 8px;
}
</style>
