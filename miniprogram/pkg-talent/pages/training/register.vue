<template>
  <view class="rg-page">
    <u-nav-bar title="培训报名" show-back @back="goBack" />

    <StateView
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && !course"
      empty-text="课程不存在"
      @retry="fetchCourse"
    >
      <template v-if="course">
        <view class="rg-content">
          <!-- 课程摘要（真实课程数据）⭐ 左侧渐变色条 + 类型胶囊 + 价格 -->
          <view class="summary-card">
            <view class="summary-bar" :class="'summary-bar--' + certTypeKey" />
            <view class="summary-main">
              <view class="summary-top">
                <text class="summary-title">{{ courseTitle }}</text>
                <view v-if="certLabel" class="cert-pill">{{ certLabel }}</view>
              </view>
              <view class="summary-bottom">
                <view class="summary-price">
                  <text v-if="currentPrice > 0" class="summary-price-symbol">¥</text>
                  <text class="summary-price-value">{{ feeText }}</text>
                  <text v-if="currentPrice > 0" class="summary-price-unit">/人</text>
                </view>
              </view>
            </view>
            <view v-if="courseOptions.length > 1" class="picker-row" @tap="showPicker = true">
              <text class="picker-label">选择课程</text>
              <text class="picker-value">{{ selectedName }}</text>
              <text class="picker-arrow">›</text>
            </view>
          </view>

          <!-- 个人信息 -->
          <view class="section-block">
            <view class="section-head">
              <text class="section-title">个人信息</text>
              <text class="required-pill">必填</text>
            </view>

            <view class="field">
              <text class="field-label">姓名<text class="field-star">*</text></text>
              <input class="field-input" v-model="form.name" placeholder="请输入报名人姓名" />
            </view>
            <view class="field">
              <text class="field-label">手机号<text class="field-star">*</text></text>
              <input class="field-input" v-model="form.phone" type="number" maxlength="11" placeholder="请输入手机号码" />
            </view>
            <view class="field">
              <text class="field-label">身份证号<text class="field-star">*</text></text>
              <input class="field-input" v-model="form.idCard" maxlength="18" placeholder="请输入身份证号码" />
            </view>
          </view>

          <!-- 证件上传（上传至 /api/v1/files/upload，提交服务端路径） -->
          <view class="section-block">
            <view class="section-head">
              <text class="section-title">证件上传</text>
              <text class="required-pill">必传</text>
            </view>

            <view class="upload-row">
              <view class="upload-box" @tap="chooseImage('photo')">
                <image v-if="photoPreview" :src="photoPreview" class="upload-preview" mode="aspectFill" />
                <view v-else class="upload-placeholder">
                  <view class="upload-cam"><text class="upload-cam-icon">＋</text></view>
                  <text class="upload-title">白底免冠证件照</text>
                  <text class="upload-hint">点击上传</text>
                </view>
              </view>
              <view class="upload-box" @tap="chooseImage('idCard')">
                <image v-if="idCardPreview" :src="idCardPreview" class="upload-preview" mode="aspectFill" />
                <view v-else class="upload-placeholder">
                  <view class="upload-cam"><text class="upload-cam-icon">＋</text></view>
                  <text class="upload-title">身份证正面</text>
                  <text class="upload-hint">点击上传</text>
                </view>
              </view>
            </view>

            <view class="checkbox-row" @tap="form.noCrime = !form.noCrime">
              <view class="checkbox-box" :class="{ checked: form.noCrime }">
                <text v-if="form.noCrime" class="check-mark">✓</text>
              </view>
              <text class="checkbox-text">本人无犯罪记录，声明属实</text>
            </view>
          </view>

          <!-- 费用（真实价格） -->
          <view class="fee-card">
            <text class="fee-label">课程费用</text>
            <view class="fee-right">
              <text v-if="currentPrice > 0" class="fee-symbol">¥</text>
              <text class="fee-value">{{ feeText }}</text>
              <text v-if="currentPrice > 0" class="fee-unit">/人</text>
            </view>
          </view>

          <!-- 隐私提示 -->
          <text class="privacy-text">报名信息仅用于课程注册，受隐私政策保护</text>
          <text class="consult-text" @tap="handleConsult">报名前想咨询？联系客服 {{ HOTLINE }}</text>
        </view>
      </template>
    </StateView>

    <!-- 底部固定 CTA 栏（联系咨询 + 确认报名） -->
    <view v-if="course" class="bottom-cta-bar">
      <view class="cta-consult" hover-class="press-feedback" :hover-stay-time="120" @tap="handleConsult">
        <text>联系咨询</text>
      </view>
      <view class="cta-submit" :class="{ submitting: submitting }" hover-class="press-feedback" :hover-stay-time="120" @tap="handleSubmit">
        <text>{{ submitting ? '提交中...' : '确认报名' }}</text>
      </view>
    </view>

    <!-- 课程选择弹窗（仅当有多个真实课程选项时展示） -->
    <template v-if="courseOptions.length > 1">
      <u-picker
        :show="showPicker"
        :columns="optionNames"
        :model-value="selectedName"
        title="选择课程"
        @confirm="onCourseChange"
        @update:show="showPicker = $event"
      />
    </template>
  </view>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request, BASE_URL, authStorage } from '../../../utils/request'
import StateView from '../../../components/StateView.vue'

const HOTLINE = '400-116-0851'

const id = ref('')
const loading = ref(false)
const errorMsg = ref('')
const course = ref(null)

const showPicker = ref(false)
const selectedIndex = ref(0)
const submitting = ref(false)
const photoPreview = ref('')
const idCardPreview = ref('')

/* 与后端 EnrollmentForm 契约一致：name/phone/idCard/gender/birthday/email/education/experience/photo/idCardImage/noCrime */
const form = reactive({
  name: '',
  phone: '',
  idCard: '',
  gender: '',
  birthday: '',
  email: '',
  education: '',
  experience: '',
  photo: '',        // 服务端文件路径（/uploads/xxx）
  idCardImage: '',  // 服务端文件路径（/uploads/xxx）
  noCrime: false,
})

const CERT_LABELS = { caac: 'CAAC执照', utc_dji: '大疆UTC认证', gov_level: '人社等级证书' }

const courseTitle = computed(function () {
  const c = course.value
  return c ? (c.title || c.name || '培训机构') : ''
})

const certLabel = computed(function () {
  const c = course.value
  if (!c) return ''
  if (c.cert_type && CERT_LABELS[c.cert_type]) return CERT_LABELS[c.cert_type]
  return c.cert_type || ''
})

/* 课程卡色条类型：caac=蓝 / utc_dji=紫 / gov_level=金 */
const certTypeKey = computed(function () {
  const t = course.value && course.value.cert_type
  return t || 'caac'
})

/* 真实课程选项：course.courses 数组（若后端下发）逐条展示，否则仅本课程一条 */
const courseOptions = computed(function () {
  const c = course.value
  if (!c) return []
  if (Array.isArray(c.courses) && c.courses.length > 0) {
    return c.courses.map(function (x) {
      return {
        name: x.name || x.title || c.title,
        price: x.price != null ? x.price : (x.price_fen ? x.price_fen / 100 : 0),
      }
    })
  }
  return [{ name: c.title || '课程', price: c.price_fen ? c.price_fen / 100 : 0 }]
})

const optionNames = computed(function () {
  return courseOptions.value.map(function (o) { return o.name })
})

const selectedName = computed(function () {
  const o = courseOptions.value[selectedIndex.value]
  return o ? o.name : ''
})

const currentPrice = computed(function () {
  const o = courseOptions.value[selectedIndex.value]
  return (o && o.price > 0) ? o.price : 0
})

const feeText = computed(function () {
  const o = courseOptions.value[selectedIndex.value]
  if (o && o.price > 0) return Number(o.price).toLocaleString()
  return '面议'
})

/* 身份证号 18 位时自动推导生日与性别 */
watch(function () { return form.idCard }, function (val) {
  if (val && val.length === 18) {
    const birth = val.substring(6, 14)
    form.birthday = birth.substring(0, 4) + '-' + birth.substring(4, 6) + '-' + birth.substring(6, 8)
    form.gender = parseInt(val.charAt(16), 10) % 2 === 0 ? '女' : '男'
  }
})

function onCourseChange(val) {
  const idx = courseOptions.value.findIndex(function (o) { return o.name === val })
  if (idx < 0) { showPicker.value = false; return }
  selectedIndex.value = idx
  showPicker.value = false
}

/* === 证件上传：uni.uploadFile → /api/v1/files/upload，提交服务端路径 === */
function chooseImage(key) {
  uni.chooseImage({
    count: 1,
    sourceType: ['album', 'camera'],
    success: function (res) {
      uploadFile(key, res.tempFilePaths[0])
    },
  })
}

async function uploadFile(key, filePath) {
  const token = authStorage.getAccessToken()
  if (!token) {
    uni.showToast({ title: '请先登录', icon: 'none' })
    setTimeout(function () { uni.navigateTo({ url: '/pages/login/index' }) }, 500)
    return
  }

  uni.showLoading({ title: '上传中...' })
  try {
    const data = await new Promise(function (resolve, reject) {
      uni.uploadFile({
        url: BASE_URL + '/api/v1/files/upload',
        filePath: filePath,
        name: 'file',
        header: { Authorization: 'Bearer ' + token },
        success: function (r) {
          if (r.statusCode >= 200 && r.statusCode < 300) {
            try { resolve(JSON.parse(r.data)) } catch (e) { reject(e) }
          } else {
            reject(new Error('upload failed ' + r.statusCode))
          }
        },
        fail: reject,
      })
    })
    // /api/v1/files/upload 信封格式 {data:{file_id,...}}，保存为 /uploads/{file_id} 路径
    const fid = data && (data.file_id || (data.data && data.data.file_id))
    if (!fid) {
      uni.showToast({ title: '上传失败，请重试', icon: 'none' })
      return
    }
    if (key === 'photo') {
      form.photo = '/uploads/' + fid
      photoPreview.value = filePath
    } else {
      form.idCardImage = '/uploads/' + fid
      idCardPreview.value = filePath
    }
  } catch (e) {
    uni.showToast({ title: '上传失败，请重试', icon: 'none' })
  } finally {
    uni.hideLoading()
  }
}

function validate() {
  if (!form.name) return '请输入姓名'
  if (!/^1[3-9]\d{9}$/.test(form.phone)) return '请输入正确的手机号'
  if (!/^\d{17}[\dXx]$/.test(form.idCard)) return '请输入正确的身份证号'
  if (!form.photo) return '请上传白底免冠证件照'
  if (!form.idCardImage) return '请上传身份证正面'
  if (!form.noCrime) return '请勾选无犯罪记录声明'
  return null
}

/* === 提交：POST /api/v1/training-courses/{id}/enroll（与后端 EnrollmentForm 契约一致） === */
async function handleSubmit() {
  if (submitting.value) return
  const err = validate()
  if (err) {
    uni.showToast({ title: err, icon: 'none' })
    return
  }

  submitting.value = true
  try {
    // noCrime 布尔勾选 → 后端按字符串存储，提交前转换
    const payload = Object.assign({}, form, {
      noCrime: form.noCrime ? '无犯罪记录' : '',
    })
    await request({
      url: '/api/v1/training-courses/' + encodeURIComponent(id.value) + '/enroll',
      method: 'POST',
      data: payload,
    })
    uni.showToast({ title: '报名成功' })
    setTimeout(function () { uni.navigateBack() }, 1500)
  } catch (e) {
    // 后端统一错误信封 {error:{code,message}}，409 重复报名等场景展示真实原因
    const msg = (e && e.data && e.data.error && e.data.error.message) || ''
    uni.showToast({ title: msg || '报名失败，请重试', icon: 'none', duration: 2500 })
  } finally {
    submitting.value = false
  }
}

function handleConsult() {
  uni.showToast({ title: '请联系客服 ' + HOTLINE, icon: 'none' })
}

function goBack() { uni.navigateBack({ delta: 1 }) }

async function fetchCourse() {
  loading.value = true
  errorMsg.value = ''

  try {
    const res = await request({ url: '/api/v1/training-courses/' + encodeURIComponent(id.value) })
    course.value = res
    if (!res) errorMsg.value = '课程不存在'
  } catch (e) {
    errorMsg.value = '加载失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

onLoad(function (options) {
  id.value = options.id || ''
  fetchCourse()
})
</script>

<style scoped>
.rg-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(160rpx + env(safe-area-inset-bottom));
}

.rg-content {
  padding: 20rpx 24rpx 0;
}

/* ===== 课程摘要 ⭐ 左侧渐变色条 + 类型胶囊 + 价格 ===== */
.summary-card {
  background: #FFFFFF;
  border: 1rpx solid #EEF1F4;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
  position: relative;
  overflow: hidden;
  display: flex;
  box-shadow: 0 4rpx 16rpx rgba(10, 31, 68, 0.06);
}

/* 左侧渐变色条（按类型配色） */
.summary-bar {
  width: 6rpx;
  border-radius: 3rpx;
  flex-shrink: 0;
  margin-right: 16rpx;
  align-self: stretch;
}
.summary-bar--caac { background: linear-gradient(180deg, #074D92, #0A66C2); }
.summary-bar--utc_dji { background: linear-gradient(180deg, #6D28D9, #DB2777); }
.summary-bar--gov_level { background: linear-gradient(180deg, #D97706, #FB923C); }

.summary-main {
  flex: 1;
  min-width: 0;
}

.summary-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12rpx;
  margin-bottom: 16rpx;
}

.summary-title {
  font-size: 30rpx;
  font-weight: 700;
  color: #17212B;
  line-height: 1.4;
  flex: 1;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

/* 右侧类型胶囊 */
.cert-pill {
  flex-shrink: 0;
  padding: 4rpx 14rpx;
  border: 1rpx solid rgba(10, 102, 194, 0.3);
  border-radius: 999rpx;
  color: #0A66C2;
  font-size: 20rpx;
  font-weight: 500;
  background: rgba(10, 102, 194, 0.06);
}

.summary-bottom {
  display: flex;
  align-items: baseline;
}

.summary-price {
  display: flex;
  align-items: baseline;
}

.summary-price-symbol {
  font-size: 26rpx;
  color: #E96012;
  font-weight: 700;
}

.summary-price-value {
  font-size: 40rpx;
  color: #E96012;
  font-weight: 800;
  margin: 0 4rpx;
}

.summary-price-unit {
  font-size: 22rpx;
  color: #98A2B3;
}

/* 多课程选择行 */
.picker-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-top: 20rpx;
  padding: 20rpx 24rpx;
  background: #F4F6F8;
  border-radius: 12rpx;
}

.picker-label {
  font-size: 24rpx;
  color: #98A2B3;
}

.picker-value {
  flex: 1;
  font-size: 28rpx;
  font-weight: 500;
  color: #17212B;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.picker-arrow {
  font-size: 32rpx;
  color: #98A2B3;
}

/* ===== 表单分组白卡 ===== */
.section-block {
  background: #FFFFFF;
  border: 1rpx solid #EEF1F4;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}

.section-head {
  display: flex;
  align-items: baseline;
  margin-bottom: 16rpx;
}

.section-title {
  font-size: 30rpx;
  font-weight: 700;
  color: #17212B;
}

/* 必填红色 pill */
.required-pill {
  margin-left: 12rpx;
  padding: 2rpx 12rpx;
  background: #FEE2E2;
  color: #EF4444;
  font-size: 20rpx;
  font-weight: 600;
  border-radius: 999rpx;
  line-height: 1.6;
}

/* 输入控件：高 44px，圆角 8px */
.field {
  margin-bottom: 20rpx;
}

.field-label {
  font-size: 26rpx;
  color: #344054;
  font-weight: 500;
  display: block;
  margin-bottom: 10rpx;
}

.field-star {
  color: #E96012;
  margin-left: 4rpx;
}

.field-input {
  box-sizing: border-box;
  width: 100%;
  height: 88rpx;
  line-height: 88rpx;
  background: #FFFFFF;
  border: 2rpx solid #E4E7EC;
  border-radius: 12rpx;
  padding: 0 20rpx;
  font-size: 28rpx;
  color: #17212B;
}

/* ===== 证件上传 ===== */
.upload-row {
  display: flex;
  gap: 20rpx;
  margin-bottom: 12rpx;
}

.upload-box {
  flex: 1;
  height: 240rpx;
  background: linear-gradient(180deg, #FAFBFC, #F5F8FC);
  border: 2rpx dashed #CBD5E1;
  border-radius: 12rpx;
  overflow: hidden;
  position: relative;
}

.upload-preview {
  width: 100%;
  height: 100%;
}

.upload-placeholder {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8rpx;
}

/* 圆底相机图标（浅蓝底 + 蓝色图标） */
.upload-cam {
  width: 72rpx;
  height: 72rpx;
  border-radius: 50%;
  background: #EAF3FB;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 4rpx;
}

.upload-cam-icon {
  font-size: 40rpx;
  color: #0A66C2;
  font-weight: 400;
}

.upload-title {
  font-size: 24rpx;
  color: #344054;
  font-weight: 500;
}

.upload-hint {
  font-size: 22rpx;
  color: #98A2B3;
}

/* 无犯罪记录 */
.checkbox-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
  padding: 16rpx 0 4rpx;
}

.checkbox-box {
  width: 40rpx;
  height: 40rpx;
  border: 2rpx solid #D0D5DD;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: #FFFFFF;
}

.checkbox-box.checked {
  background: #0A66C2;
  border-color: #0A66C2;
}

.check-mark {
  color: #FFFFFF;
  font-size: 24rpx;
  font-weight: 700;
}

.checkbox-text {
  font-size: 26rpx;
  color: #344054;
}

/* ===== 费用（浅灰底圆角 + 价格集中强化） ===== */
.fee-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #F5F8FC;
  border: 1rpx solid #EEF1F4;
  border-radius: 12rpx;
  padding: 20rpx 24rpx;
  margin-bottom: 24rpx;
}

.fee-label {
  font-size: 24rpx;
  color: #667085;
}

.fee-right {
  display: flex;
  align-items: baseline;
}

.fee-symbol {
  font-size: 26rpx;
  color: #E96012;
  font-weight: 700;
}

.fee-value {
  font-size: 48rpx;
  font-weight: 800;
  color: #E96012;
  margin: 0 4rpx;
}

.fee-unit {
  font-size: 24rpx;
  color: #98A2B3;
}

/* ===== 底部固定 CTA 栏（联系咨询 + 确认报名） ===== */
.bottom-cta-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  gap: 20rpx;
  padding: 16rpx 32rpx;
  padding-bottom: calc(16rpx + env(safe-area-inset-bottom));
  background: #FFFFFF;
  z-index: 100;
  box-shadow: 0 -4rpx 20rpx rgba(0, 0, 0, 0.06);
}

.cta-consult {
  flex: 1;
  height: 92rpx;
  border: 2rpx solid #0A66C2;
  border-radius: 999rpx;
  color: #0A66C2;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 30rpx;
  font-weight: 600;
  background: #FFFFFF;
}

.cta-submit {
  flex: 2;
  height: 92rpx;
  border-radius: 999rpx;
  background: linear-gradient(135deg, #074D92, #0A66C2);
  color: #FFFFFF;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 30rpx;
  font-weight: 600;
  letter-spacing: 2rpx;
  box-shadow: 0 4rpx 16rpx rgba(10, 102, 194, 0.25);
  animation: ctaPulse 2s ease-in-out infinite;
}

.cta-submit.submitting {
  opacity: 0.6;
  animation: none;
}

@keyframes ctaPulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(10, 102, 194, 0.3); }
  50% { box-shadow: 0 0 0 8rpx rgba(10, 102, 194, 0); }
}

.press-feedback {
  transform: scale(0.98);
  opacity: 0.85;
}

.privacy-text {
  display: block;
  text-align: center;
  font-size: 22rpx;
  color: #98A2B3;
  margin-top: 20rpx;
}

.consult-text {
  display: block;
  text-align: center;
  font-size: 24rpx;
  color: #0A66C2;
  margin-top: 12rpx;
}
</style>
