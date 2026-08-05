<template>
  <view class="tsb-page">
    <u-nav-bar title="预约测试" show-back @back="goBack" />

    <!-- 场地摘要 -->
    <view v-if="site" class="summary-card">
      <view class="sum-top">
        <text class="sum-name">{{ site.name }}</text>
        <text class="sum-type">{{ typeLabel(site.site_type) }}</text>
      </view>
      <view class="sum-row">
        <text class="sum-label">参考价格</text>
        <text class="sum-price">{{ formatPrice(site.price_fen) }}</text>
      </view>
      <text v-if="site.booking_rule" class="sum-rule">{{ site.booking_rule }}</text>
    </view>

    <view v-if="site" class="form-wrap">
      <!-- 预约用途 -->
      <view class="form-group">
        <view class="form-label">预约用途 <text class="required">*</text></view>
        <input
          class="form-input"
          v-model="form.purpose"
          placeholder="如：无人机性能测试 / 产品研发验证"
          placeholder-class="ph"
        />
      </view>

      <!-- 预约日期 -->
      <view class="form-group">
        <view class="form-label">预约日期 <text class="required">*</text></view>
        <picker mode="date" :value="form.date" :start="minDate" @change="onDateChange">
          <view class="picker-box">
            <text class="picker-text" :class="{ placeholder: !form.date }">{{ form.date || '请选择日期' }}</text>
            <text class="picker-arrow">›</text>
          </view>
        </picker>
      </view>

      <!-- 预约时段（单时段，格式 HH:MM-HH:MM） -->
      <view class="form-group">
        <view class="form-label">预约时段 <text class="required">*</text></view>
        <view class="slot-grid">
          <view
            v-for="slot in slots"
            :key="slot"
            class="slot-item"
            :class="{ selected: form.time_slot === slot }"
            @tap="selectSlot(slot)"
          >{{ slot }}</view>
        </view>
        <text class="form-hint">具体时段以场地方确认为准</text>
      </view>

      <!-- 联系人 -->
      <view class="form-group">
        <view class="form-label">联系人 <text class="required">*</text></view>
        <input
          class="form-input"
          v-model="form.contact_name"
          placeholder="请输入联系人姓名"
          placeholder-class="ph"
        />
      </view>

      <!-- 联系电话 -->
      <view class="form-group">
        <view class="form-label">联系电话 <text class="required">*</text></view>
        <input
          class="form-input"
          v-model="form.contact_phone"
          type="number"
          maxlength="11"
          placeholder="请输入联系电话"
          placeholder-class="ph"
        />
        <text v-if="phoneError" class="form-error">请输入正确的手机号</text>
      </view>

      <!-- 资金说明 -->
      <view class="notice-block">
        <text class="notice-title">费用说明</text>
        <text class="notice-line">· 定金及测试费用在线下向场地方支付，平台不参与资金流转</text>
      </view>
    </view>

    <!-- 底部提交栏 -->
    <view v-if="site" class="submit-bar">
      <view class="submit-btn" :class="{ disabled: submitting }" @tap="handleSubmit">
        {{ submitting ? '提交中...' : '提交预约' }}
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request, authStorage, getStoredUser } from '../../utils/request'

const SITE_TYPE_MAP = {
  flying_field: '飞行场地',
  lab: '实验室',
  anechoic_chamber: '暗室',
  wind_tunnel: '风洞',
}

// 预约时段选项（单时段，提交格式 HH:MM-HH:MM）
const SLOTS = [
  '08:00-10:00',
  '10:00-12:00',
  '13:00-15:00',
  '15:00-17:00',
  '17:00-19:00',
]

const site = ref(null)
const submitting = ref(false)
const minDate = ref('')
const slots = SLOTS
const form = ref({
  purpose: '',
  date: '',
  time_slot: '',
  contact_name: '',
  contact_phone: '',
})
let siteId = ''

const phoneError = computed(() => {
  const p = form.value.contact_phone
  return !!p && !/^1[3-9]\d{9}$/.test(p)
})

function typeLabel(t) { return SITE_TYPE_MAP[t] || t || '测试场地' }
function formatPrice(fen) {
  if (fen == null || fen <= 0) return '面议'
  const yuan = fen / 100
  const text = Number.isInteger(yuan) ? String(yuan) : yuan.toFixed(2)
  return '¥' + text
}
function today() {
  const now = new Date()
  const pad = (n) => (n < 10 ? '0' + n : '' + n)
  return now.getFullYear() + '-' + pad(now.getMonth() + 1) + '-' + pad(now.getDate())
}

async function fetchSite() {
  if (!siteId) return
  try {
    const res = await request({ url: '/api/v1/test-sites/' + encodeURIComponent(siteId) })
    const d = (res && res.data) || res
    if (d && d.id) site.value = d
  } catch (e) { /* 保持空态 */ }
}

function onDateChange(e) {
  form.value.date = e.detail.value
  form.value.time_slot = ''
}

function selectSlot(slot) {
  if (!form.value.date) {
    uni.showToast({ title: '请先选择日期', icon: 'none' })
    return
  }
  form.value.time_slot = form.value.time_slot === slot ? '' : slot
}

async function handleSubmit() {
  var token = authStorage.getAccessToken()
  if (!token) {
    uni.showToast({ title: '请先登录', icon: 'none' })
    uni.navigateTo({ url: '/pages/login/index' })
    return
  }
  if (!siteId) {
    uni.showToast({ title: '场地信息缺失', icon: 'none' })
    return
  }
  if (!form.value.purpose) {
    uni.showToast({ title: '请填写预约用途', icon: 'none' })
    return
  }
  if (!form.value.date) {
    uni.showToast({ title: '请选择预约日期', icon: 'none' })
    return
  }
  if (!form.value.time_slot) {
    uni.showToast({ title: '请选择预约时段', icon: 'none' })
    return
  }
  if (!form.value.contact_name) {
    uni.showToast({ title: '请填写联系人', icon: 'none' })
    return
  }
  if (!/^1[3-9]\d{9}$/.test(form.value.contact_phone)) {
    uni.showToast({ title: '请输入正确的手机号', icon: 'none' })
    return
  }

  submitting.value = true
  try {
    await request({
      url: '/api/v1/test-sites/' + encodeURIComponent(siteId) + '/book',
      method: 'POST',
      data: {
        purpose: form.value.purpose,
        date: form.value.date,
        time_slot: form.value.time_slot,
        contact_name: form.value.contact_name,
        contact_phone: form.value.contact_phone,
      },
    })
    // 保存预约摘要，供确认页（pay）展示
    uni.setStorageSync('testBookingDraft', {
      siteId: siteId,
      siteName: site.value.name,
      date: form.value.date,
      time_slot: form.value.time_slot,
      purpose: form.value.purpose,
      contact_name: form.value.contact_name,
      contact_phone: form.value.contact_phone,
      price_fen: site.value.price_fen,
    })
    uni.redirectTo({ url: '/pages/testsites/pay' })
  } catch (e) {
    uni.showToast({ title: '预约提交失败，请稍后重试', icon: 'none' })
  } finally {
    submitting.value = false
  }
}

function goBack() {
  uni.navigateBack()
}

onLoad((options) => {
  siteId = (options && options.id) || ''
  minDate.value = today()
  form.value.date = today()

  var token = authStorage.getAccessToken()
  if (!token) {
    uni.showToast({ title: '请先登录', icon: 'none' })
    setTimeout(function () {
      uni.navigateTo({ url: '/pages/login/index' })
    }, 500)
    return
  }
  // 预填联系电话（微信登录用户可能无手机号，留空需手动填写）
  var u = getStoredUser()
  if (u && u.phone) {
    form.value.contact_phone = u.phone
  }
  fetchSite()
})
</script>

<style scoped>
.tsb-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: 96px;
}

/* 场地摘要 */
.summary-card {
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  padding: 14px 16px;
  margin: 12px 12px 8px;
}

.sum-top {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.sum-name {
  flex: 1;
  min-width: 0;
  font-size: 16px;
  font-weight: 600;
  color: #17212B;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.sum-type {
  flex-shrink: 0;
  font-size: 11px;
  color: #0A66C2;
  background: #EAF3FB;
  padding: 3px 8px;
  border-radius: 4px;
}

.sum-row {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.sum-label {
  font-size: 12px;
  color: #98A2B3;
}

.sum-price {
  font-size: 16px;
  font-weight: 700;
  color: #E96012;
}

.sum-rule {
  display: block;
  margin-top: 8px;
  font-size: 12px;
  color: #667085;
  line-height: 1.6;
}

/* 表单 */
.form-wrap {
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  margin: 0 12px 8px;
  padding: 16px;
}

.form-group {
  margin-bottom: 16px;
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-label {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  font-weight: 600;
  color: #344054;
  margin-bottom: 6px;
}

.required {
  color: #D92D20;
  font-size: 12px;
}

.form-input {
  width: 100%;
  box-sizing: border-box;
  min-height: 44px;
  padding: 10px 12px;
  border: 1px solid #E4E7EC;
  border-radius: 8px;
  font-size: 14px;
  color: #17212B;
  background: #fff;
}

.ph {
  color: #98A2B3;
}

.form-hint {
  display: block;
  margin-top: 6px;
  font-size: 11px;
  color: #98A2B3;
}

.form-error {
  display: block;
  margin-top: 4px;
  font-size: 11px;
  color: #D92D20;
}

/* 日期选择 */
.picker-box {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 44px;
  padding: 0 12px;
  border: 1px solid #E4E7EC;
  border-radius: 8px;
  background: #fff;
}

.picker-text {
  font-size: 14px;
  color: #17212B;
}

.picker-text.placeholder {
  color: #98A2B3;
}

.picker-arrow {
  font-size: 18px;
  color: #98A2B3;
}

/* 时段选择 */
.slot-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.slot-item {
  flex: none;
  width: calc((100% - 16px) / 3);
  box-sizing: border-box;
  padding: 10px 4px;
  border: 1px solid #E4E7EC;
  border-radius: 8px;
  text-align: center;
  font-size: 13px;
  color: #344054;
  background: #fff;
  transition: all 0.2s;
}

.slot-item.selected {
  border-color: #0A66C2;
  background: #0A66C2;
  color: #fff;
  font-weight: 600;
}

/* 资金说明 */
.notice-block {
  background: #F4F8FC;
  border-radius: 8px;
  padding: 12px 16px;
  margin: 0 12px 8px;
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

/* 底部提交栏 */
.submit-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 12px;
  padding-bottom: calc(12px + env(safe-area-inset-bottom));
  background: #fff;
  border-top: 1px solid #EEF1F4;
}

.submit-btn {
  height: 46px;
  border-radius: 8px;
  background: #0A66C2;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 15px;
  font-weight: 600;
}

.submit-btn.disabled {
  background: #98A2B3;
}
</style>
