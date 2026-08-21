<template>
  <view class="pub-page" :style="{ paddingTop: topPad + 'px' }">
    <!-- 顶栏（与发布页同款） -->
    <view class="pub-nav">
      <view class="pub-back" hover-class="pub-fade" @tap="goBack">‹</view>
      <view class="pub-nav-title">培训报名</view>
    </view>

    <!-- 加载中 -->
    <view v-if="loading" class="loading-state">
      <u-loading size="28rpx" />
      <text>加载中...</text>
    </view>

    <!-- 加载失败 -->
    <view v-else-if="errorMsg" class="pub-empty">
      <view class="pub-empty-mark">!</view>
      <view class="pub-empty-title">加载失败</view>
      <view class="pub-empty-desc">请检查网络后重试</view>
      <view class="pub-btn pub-btn--primary retry-btn" hover-class="pub-btn--active" @tap="fetchCourse">重新加载</view>
    </view>

    <!-- 课程不存在 -->
    <view v-else-if="!course" class="pub-empty">
      <view class="pub-empty-title">课程不存在</view>
    </view>

    <!-- 报名表单 -->
    <template v-else>
      <!-- 表单头部：课程摘要（真实课程数据）⭐ 类型色条 + 证书胶囊 -->
      <view class="pub-form-intro">
        <view class="course-bar" :class="'course-bar--' + certTypeKey" />
        <view class="course-head">
          <view class="pub-form-intro-h2">{{ courseTitle }}</view>
          <view v-if="certLabel" class="cert-pill">{{ certLabel }}</view>
        </view>
      </view>

      <!-- 课程信息（多课程选择 + 真实价格） -->
      <view class="pub-section">
        <view class="pub-section-title">课程信息</view>
        <view class="pub-form-card">
          <!-- 多课程选择（仅当有多个真实课程选项时展示） -->
          <view v-if="courseOptions.length > 1" class="pub-field" @tap="showPicker = true">
            <view class="pub-field-label">选择课程</view>
            <view class="pub-select-field">
              <text class="pub-select-value">{{ selectedName }}</text>
              <text class="pub-arrow">›</text>
            </view>
          </view>
          <!-- 课程费用（真实价格） -->
          <view class="pub-field">
            <view class="pub-field-label">课程费用</view>
            <view class="fee-price">
              <text v-if="currentPrice > 0" class="fee-price-symbol">¥</text>
              <text class="fee-price-value">{{ feeText }}</text>
              <text v-if="currentPrice > 0" class="fee-price-unit">/人</text>
            </view>
          </view>
        </view>
      </view>

      <!-- 个人信息 -->
      <view class="pub-section">
        <view class="pub-section-title">个人信息</view>
        <view class="pub-form-card">
          <view class="pub-field">
            <view class="pub-field-label">姓名<text class="pub-required">*</text></view>
            <input
              v-model="form.name"
              class="pub-input"
              placeholder="请输入报名人姓名"
              placeholder-class="pub-placeholder"
            />
          </view>
          <view class="pub-field">
            <view class="pub-field-label">手机号<text class="pub-required">*</text></view>
            <input
              v-model="form.phone"
              class="pub-input"
              type="number"
              maxlength="11"
              placeholder="请输入手机号码"
              placeholder-class="pub-placeholder"
            />
          </view>
          <view class="pub-field">
            <view class="pub-field-label">身份证号<text class="pub-required">*</text></view>
            <input
              v-model="form.idCard"
              class="pub-input"
              maxlength="18"
              placeholder="请输入身份证号码"
              placeholder-class="pub-placeholder"
            />
          </view>
        </view>
      </view>

      <!-- 证件上传（上传至 /api/v1/files/upload，提交服务端路径） -->
      <view class="pub-section">
        <view class="pub-section-title">证件上传</view>
        <view class="pub-section-note">必传 2 张：白底免冠证件照、身份证正面</view>
        <view class="pub-form-card">
          <view class="pub-upload-row">
            <view v-if="photoPreview" class="pub-photo" @tap="chooseImage('photo')">
              <image :src="photoPreview" mode="aspectFill" class="pub-photo-img" />
            </view>
            <view v-else class="pub-add-photo" hover-class="pub-fade" @tap="chooseImage('photo')">＋</view>

            <view v-if="idCardPreview" class="pub-photo" @tap="chooseImage('idCard')">
              <image :src="idCardPreview" mode="aspectFill" class="pub-photo-img" />
            </view>
            <view v-else class="pub-add-photo" hover-class="pub-fade" @tap="chooseImage('idCard')">＋</view>
          </view>
          <view class="pub-upload-tip">左侧白底免冠证件照，右侧身份证正面，点击可重新上传</view>

          <view class="pub-check-row" @tap="form.noCrime = !form.noCrime">
            <view class="no-crime-box" :class="{ 'no-crime-box--checked': form.noCrime }">
              <text v-if="form.noCrime" class="no-crime-mark">✓</text>
            </view>
            <text>本人无犯罪记录，声明属实</text>
          </view>
        </view>
      </view>

      <!-- 隐私提示 -->
      <text class="privacy-text">报名信息仅用于课程注册，受隐私政策保护</text>
      <text class="consult-text" @tap="handleConsult">报名前想咨询？联系客服 {{ HOTLINE }}</text>
    </template>

    <!-- 底部固定 CTA 栏（联系咨询 + 确认报名，主蓝 + 次白描边） -->
    <view v-if="course" class="pub-sticky">
      <view class="pub-btn pub-btn--secondary cta-consult" hover-class="pub-btn--active" @tap="handleConsult">
        <text>联系咨询</text>
      </view>
      <view
        class="pub-btn pub-btn--primary cta-submit"
        :class="{ 'cta-submit--busy': submitting }"
        hover-class="pub-btn--active"
        @tap="handleSubmit"
      >
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
import { onLoad, onUnload } from '@dcloudio/uni-app'
import { request, BASE_URL, authStorage } from '../../../utils/request'
import { useSafeTop } from '../../../utils/safeTop'

const { topPad, initSafeTop } = useSafeTop(true)

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
let backTimer = null

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

/* 课程卡色条类型：caac=蓝 / utc_dji=紫 / gov_level=橙 */
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

/* === 提交：POST /api/v1/training-courses/{id}/pay-and-enroll（与后端 EnrollmentForm 契约一致） ===
   报名即冻结托管金：payAndEnroll 先冻结课程费再报名；余额不足返回 402，
   前端提示充值引导（小程序暂无托管金充值入口，toast 引导）。 */
async function handleSubmit() {
  if (submitting.value) return
  const err = validate()
  if (err) {
    uni.showToast({ title: err, icon: 'none' })
    return
  }
  // 隐私二次确认：报名收集姓名+身份证号+证件照等高敏信息，提交前明确用途
  const confirmed = await new Promise((resolve) => {
    uni.showModal({
      title: '信息授权确认',
      content: '报名将提交您的姓名、身份证号与证件照片，用于实名报名与培训档案管理，仅平台与课程机构可见。是否确认提交？',
      confirmText: '确认提交',
      cancelText: '再想想',
      success: (r) => resolve(!!r.confirm),
      fail: () => resolve(false),
    })
  })
  if (!confirmed) return

  submitting.value = true
  try {
    // noCrime 布尔勾选 → 后端按字符串存储，提交前转换
    const payload = Object.assign({}, form, {
      noCrime: form.noCrime ? '无犯罪记录' : '',
    })
    await request({
      url: '/api/v1/training-courses/' + encodeURIComponent(id.value) + '/pay-and-enroll',
      method: 'POST',
      data: payload,
    })
    uni.showToast({ title: '报名成功' })
    backTimer = setTimeout(function () { uni.navigateBack() }, 1500)
  } catch (e) {
    // 后端统一错误信封 {error:{code,message}}，409 重复报名等场景展示真实原因
    const msg = (e && e.data && e.data.error && e.data.error.message) || ''
    // 402 托管金余额不足：报名即冻结学费，余额不足后端拒绝——小程序暂无充值入口，toast 引导
    if (e && e.statusCode === 402) {
      uni.showToast({ title: '托管金余额不足，请先充值', icon: 'none', duration: 2500 })
      return
    }
    uni.showToast({ title: msg || '报名失败，请重试', icon: 'none', duration: 2500 })
  } finally {
    submitting.value = false
  }
}

function handleConsult() {
  uni.showToast({ title: '请联系客服 ' + HOTLINE, icon: 'none' })
}

function goBack() { safeBack() }

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
  initSafeTop()
  id.value = options.id || ''
  fetchCourse()
})

onUnload(function () {
  if (backTimer) clearTimeout(backTimer)
})
</script>

<style scoped>
@import '../../../pages/publish/pub-style.css';

.pub-fade { opacity: 0.6; }
.pub-form-intro-h2 {
  font-size: 20px;
  margin: 0 0 4px;
  color: #17212B;
}
.pub-photo-img {
  width: 100%;
  height: 100%;
  display: block;
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

/* ===== 课程摘要：类型色条 + 证书胶囊（配色锚定 pub 色板） ===== */
.course-bar {
  width: 34px;
  height: 4px;
  border-radius: 3px;
  margin-bottom: 8px;
}
.course-bar--caac { background: linear-gradient(90deg, #0A66C2, rgba(10, 102, 194, 0.35)); }
.course-bar--utc_dji { background: linear-gradient(90deg, #7056D6, rgba(112, 86, 214, 0.35)); }
.course-bar--gov_level { background: linear-gradient(90deg, #F97316, rgba(249, 115, 22, 0.35)); }

.course-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}
.course-head .pub-form-intro-h2 {
  flex: 1;
  min-width: 0;
}

.cert-pill {
  flex-shrink: 0;
  padding: 3px 10px;
  border: 1px solid #A6C9EE;
  border-radius: 999px;
  color: #0A66C2;
  font-size: 11px;
  font-weight: 700;
  background: #E8F2FC;
}

/* ===== 课程费用（真实价格，橙强化） ===== */
.fee-price {
  display: flex;
  align-items: baseline;
}
.fee-price-symbol {
  font-size: 13px;
  color: #F97316;
  font-weight: 700;
}
.fee-price-value {
  font-size: 22px;
  font-weight: 800;
  color: #F97316;
  margin: 0 2px;
  line-height: 1.2;
}
.fee-price-unit {
  font-size: 12px;
  color: #98A2B3;
}

/* ===== 无犯罪记录声明（pub 蓝勾选） ===== */
.no-crime-box {
  width: 18px;
  height: 18px;
  border: 1px solid #C4CBD3;
  border-radius: 5px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: #FFFFFF;
}
.no-crime-box--checked {
  background: #0A66C2;
  border-color: #0A66C2;
}
.no-crime-mark {
  color: #FFFFFF;
  font-size: 12px;
  font-weight: 700;
  line-height: 1;
}

/* ===== 底部双 CTA（保持原比例 1:2，脉冲动画对齐 pub 蓝） ===== */
.cta-consult { flex: 1; }
.cta-submit {
  flex: 2;
  letter-spacing: 0.08em;
  animation: ctaPulse 2s ease-in-out infinite;
}
.cta-submit--busy {
  opacity: 0.6;
  animation: none;
}

@keyframes ctaPulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(10, 102, 194, 0.3); }
  50% { box-shadow: 0 0 0 8px rgba(10, 102, 194, 0); }
}

/* ===== 底部隐私提示 ===== */
.privacy-text {
  display: block;
  text-align: center;
  font-size: 11px;
  color: #98A2B3;
  margin-top: 14px;
}

.consult-text {
  display: block;
  text-align: center;
  font-size: 12px;
  color: #0A66C2;
  margin-top: 6px;
}
</style>
