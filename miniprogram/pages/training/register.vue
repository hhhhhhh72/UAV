<template>
  <view class="page">
    <StateView
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && !course"
      empty-text="课程不存在"
      @retry="fetchCourse"
    >
      <template v-if="course">
        <!-- ====== ① 绿色 Banner ====== -->
        <view class="banner">
          <view class="banner-nav">
            <view class="back-btn" @click="goBack">
              <text class="back-icon">‹</text>
            </view>
            <text class="banner-nav-title">课程报名</text>
          </view>
          <text class="banner-org-name">{{ course.title || course.name || '培训机构' }}</text>
          <view class="banner-type-tags">
            <text
              v-for="t in courseTypesDisplay"
              :key="t"
              class="banner-tag"
            >{{ t }}</text>
          </view>
        </view>

        <!-- ====== 主卡片 ====== -->
        <view class="main-card">
          <!-- ② 课程与机型选择 -->
          <view class="course-picker" @click="showPicker = true">
            <view class="picker-left">
              <text class="picker-label">选择课程与机型</text>
              <text class="picker-value">{{ selectedCourse }}</text>
            </view>
            <text class="picker-arrow">›</text>
          </view>

          <!-- ③ 个人信息表单 -->
          <view class="section-header">
            <text class="section-title">个人信息</text>
            <text class="section-badge">*必填</text>
          </view>

          <view class="form-group">
            <view class="form-item">
              <text class="form-label">姓名</text>
              <input
                class="form-input"
                v-model="form.name"
                placeholder="请输入报名人姓名"
              />
              <text class="form-required">*</text>
            </view>
            <view class="form-item">
              <text class="form-label">手机号</text>
              <input
                class="form-input"
                v-model="form.phone"
                type="number"
                maxlength="11"
                placeholder="请输入手机号码"
              />
              <text class="form-required">*</text>
            </view>
            <view class="form-item">
              <text class="form-label">身份证号</text>
              <input
                class="form-input"
                v-model="form.idCard"
                maxlength="18"
                placeholder="请输入身份证号码"
              />
              <text class="form-required">*</text>
            </view>
          </view>

          <!-- 展开更多信息 -->
          <view class="expand-btn" @click="showMore = !showMore">
            <text>{{ showMore ? '收起更多信息' : '展开更多信息' }}</text>
            <text class="expand-arrow">{{ showMore ? '▲' : '▼' }}</text>
          </view>

          <!-- 展开区域 -->
          <view v-if="showMore" class="extra-form">
            <text class="extra-label">补充信息 <text class="optional">选填</text></text>

            <view class="extra-item">
              <text class="form-label">性别</text>
              <view class="radio-group">
                <view
                  v-for="g in ['男', '女']"
                  :key="g"
                  class="radio-item"
                  :class="{ active: form.gender === g }"
                  @click="form.gender = g"
                >
                  <text>{{ form.gender === g ? '●' : '○' }} {{ g }}</text>
                </view>
              </view>
            </view>

            <view class="extra-item">
              <text class="form-label">出生日期</text>
              <picker mode="date" :value="form.birthday" @change="onBirthdayChange">
                <text class="picker-text">{{ form.birthday || '____年__月__日' }}</text>
              </picker>
            </view>

            <view class="extra-item">
              <text class="form-label">电子邮箱</text>
              <input class="form-input" v-model="form.email" placeholder="请输入邮箱" />
            </view>

            <view class="extra-item">
              <text class="form-label">学历</text>
              <picker :range="educationList" :value="educationIdx" @change="onEducationChange">
                <text class="picker-text">{{ form.education || '请选择' }}</text>
              </picker>
            </view>

            <view class="extra-item extra-item-last">
              <text class="form-label">驾驶基础</text>
              <view class="radio-group">
                <view
                  v-for="ex in experienceList"
                  :key="ex"
                  class="radio-item radio-sm"
                  :class="{ active: form.experience === ex }"
                  @click="form.experience = ex"
                >
                  <text>{{ form.experience === ex ? '●' : '○' }} {{ ex }}</text>
                </view>
              </view>
            </view>
          </view>

          <!-- ④ 证件上传 -->
          <view class="section-header">
            <text class="section-title">证件上传</text>
          </view>

          <view class="upload-row">
            <view class="upload-box" @click="uploadImage('photo')">
              <image
                v-if="form.photo"
                :src="form.photo"
                class="upload-preview"
                mode="aspectFill"
              />
              <view v-else class="upload-placeholder">
                <text class="upload-icon">📷</text>
                <text class="upload-title">白底免冠证件照</text>
                <text class="upload-hint">点击上传</text>
              </view>
            </view>
            <view class="upload-box" @click="uploadImage('idCard')">
              <image
                v-if="form.idCardImage"
                :src="form.idCardImage"
                class="upload-preview"
                mode="aspectFill"
              />
              <view v-else class="upload-placeholder">
                <text class="upload-icon">🪪</text>
                <text class="upload-title">身份证正面</text>
                <text class="upload-hint">点击上传</text>
              </view>
            </view>
          </view>

          <!-- 无犯罪记录 -->
          <view class="checkbox-row" @click="form.noCrime = !form.noCrime">
            <view class="checkbox-box" :class="{ checked: form.noCrime }">
              <text v-if="form.noCrime" class="check-mark">✓</text>
            </view>
            <text class="checkbox-text">本人无犯罪记录，声明属实</text>
          </view>

          <!-- ⑤ 费用明细 -->
          <view class="price-section">
            <view class="price-row">
              <text class="price-label">课程费用</text>
              <text class="price-value">¥{{ currentPrice.toLocaleString() }}</text>
            </view>
            <view class="price-row">
              <text class="price-label">单位</text>
              <text class="price-unit">/人</text>
            </view>
          </view>

          <!-- ⑥ 底部操作栏 -->
          <view class="bottom-bar">
            <view class="bottom-btn outline" @click="handleConsult">联系咨询</view>
            <view class="bottom-btn primary" @click="handleSubmit">确认报名</view>
          </view>
          <text class="privacy-text">报名信息仅用于课程注册，受隐私政策保护</text>
        </view>
      </template>
    </StateView>

    <!-- Picker 弹窗 -->
    <van-popup :show="showPicker" position="bottom" round @click-overlay="showPicker = false">
      <van-picker
        :columns="courseTypes"
        value-key="label"
        :default-index="courseTypeIdx"
        @confirm="onCourseChange"
        @cancel="showPicker = false"
      />
    </van-popup>
  </view>
</template>

<script setup>
import { ref, reactive, watch, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '../../utils/request'
import StateView from '../../components/StateView.vue'

const id = ref('')
const loading = ref(false)
const errorMsg = ref('')
const course = ref(null)

const courseTypes = [
  { label: '视距内驾驶员 · 多旋翼', price: 5800 },
  { label: '视距内驾驶员 · 固定翼', price: 6800 },
  { label: '超视距驾驶员 · 多旋翼', price: 8800 },
  { label: '超视距驾驶员 · 固定翼', price: 9800 },
]

const courseTypeIdx = ref(0)
const showPicker = ref(false)
const selectedCourse = ref(courseTypes[0].label)
const currentPrice = ref(courseTypes[0].price)

const courseTypesDisplay = computed(function () {
  if (course.value && course.value.course_types) return course.value.course_types
  if (course.value && course.value.cert_type) return [course.value.cert_type + '视距内', course.value.cert_type + '超视距']
  return ['CAAC视距内', 'CAAC超视距']
})

const showMore = ref(false)
const educationList = ['高中', '大专', '本科', '硕士及以上']
const educationIdx = ref(-1)
const experienceList = ['零基础', '业余爱好者', '有飞行经验']

const form = reactive({
  name: '',
  phone: '',
  idCard: '',
  gender: '',
  birthday: '',
  email: '',
  education: '',
  experience: '',
  photo: '',
  idCardImage: '',
  noCrime: false,
})

watch(function () { return form.idCard }, function (val) {
  if (val && val.length === 18) {
    var birth = val.substring(6, 14)
    form.birthday = birth.substring(0, 4) + '-' + birth.substring(4, 6) + '-' + birth.substring(6, 8)
    form.gender = parseInt(val.charAt(16), 10) % 2 === 0 ? '女' : '男'
  }
})

function onCourseChange(e) {
  var idx = typeof e.detail === 'number' ? e.detail : (e.detail && e.detail.index)
  if (idx === undefined || idx === null) { showPicker.value = false; return }
  courseTypeIdx.value = idx
  selectedCourse.value = courseTypes[idx].label
  currentPrice.value = courseTypes[idx].price
  showPicker.value = false
}

function onBirthdayChange(e) { form.birthday = e.detail.value }

function onEducationChange(e) {
  educationIdx.value = e.detail.value
  form.education = educationList[e.detail.value]
}

function uploadImage(type) {
  uni.chooseImage({
    count: 1,
    sourceType: ['album', 'camera'],
    success: function (res) {
      form[type === 'photo' ? 'photo' : 'idCardImage'] = res.tempFilePaths[0]
    },
  })
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

async function handleSubmit() {
  var err = validate()
  if (err) {
    uni.showToast({ title: err, icon: 'none' })
    return
  }
  try {
    await request({
      url: '/api/v1/training-courses/' + encodeURIComponent(id.value) + '/pay-and-enroll',
      method: 'POST',
      data: form,
    })
    uni.showToast({ title: '报名成功' })
    setTimeout(function () { uni.navigateBack() }, 1500)
  } catch (e) {
    uni.showToast({ title: '报名失败，请重试', icon: 'none' })
  }
}

function handleConsult() {
  uni.showToast({ title: '请联系客服 400-116-0851', icon: 'none' })
}

function goBack() { uni.navigateBack({ delta: 1 }) }

async function fetchCourse() {
  loading.value = true
  errorMsg.value = ''

  try {
    var res = await request({ url: '/api/v1/training-courses' })
    var data = Array.isArray(res) ? res : (res && res.data) || res || {}
    var items = Array.isArray(data) ? data : (data && data.items) || data || []
    var found = null
    for (var i = 0; i < items.length; i++) {
      if (String(items[i].id) === String(id.value)) { found = items[i]; break }
    }
    course.value = found
    if (!found) errorMsg.value = '课程不存在'
  } catch (e) {
    errorMsg.value = '网络异常，请稍后重试'
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
.page { min-height: 100vh; background: #f5f6f8; padding-bottom: env(safe-area-inset-bottom); }

/* ① Banner */
.banner {
  background: linear-gradient(135deg, #07c160, #05a854);
  padding: 80rpx 32rpx 64rpx;
}

.banner-nav { display: flex; align-items: center; gap: 12rpx; margin-bottom: 24rpx; }

.back-btn {
  width: 64rpx; height: 64rpx; background: rgba(255,255,255,0.2);
  border-radius: 50%; display: flex; align-items: center; justify-content: center;
}

.back-icon { color: #ffffff; font-size: 40rpx; font-weight: 300; }
.banner-nav-title { color: rgba(255,255,255,0.85); font-size: 28rpx; font-weight: 500; }

.banner-org-name {
  color: #ffffff; font-size: 56rpx; font-weight: 700; line-height: 1.2; margin-bottom: 8rpx;
}

.banner-type-tags { display: flex; gap: 12rpx; margin-top: 16rpx; }
.banner-tag {
  padding: 6rpx 18rpx; border: 1rpx solid rgba(255,255,255,0.4);
  border-radius: 20rpx; color: rgba(255,255,255,0.9); font-size: 22rpx;
}

/* 主卡片 */
.main-card {
  background: #ffffff; border-radius: 32rpx 32rpx 0 0; margin-top: -32rpx;
  padding: 40rpx 32rpx 32rpx; position: relative; z-index: 2;
}

/* ② 课程选择 */
.course-picker {
  background: #f8f9fc; border-radius: 16rpx; padding: 24rpx;
  display: flex; justify-content: space-between; align-items: center; margin-bottom: 32rpx;
}
.picker-left { flex: 1; }
.picker-label { font-size: 24rpx; color: #969799; display: block; margin-bottom: 6rpx; }
.picker-value { font-size: 30rpx; font-weight: 500; color: #1a1a1a; }
.picker-arrow { font-size: 36rpx; color: #969799; flex-shrink: 0; }

/* ③ 表单 */
.section-header {
  display: flex; align-items: center; padding-left: 16rpx;
  border-left: 6rpx solid #5b8cff; margin-bottom: 24rpx;
}
.section-title { font-size: 30rpx; font-weight: 700; color: #1a1a1a; }
.section-badge { font-size: 22rpx; color: #ff6b35; margin-left: 8rpx; }

.form-group { margin-bottom: 8rpx; }
.form-item { display: flex; align-items: center; padding: 22rpx 0; border-bottom: 1rpx solid #ebedf0; }
.form-label { font-size: 28rpx; color: #1a1a1a; width: 130rpx; flex-shrink: 0; }
.form-input { flex: 1; font-size: 28rpx; color: #1a1a1a; }
.form-required { font-size: 22rpx; color: #ff6b35; }

.expand-btn {
  display: flex; align-items: center; justify-content: center; gap: 8rpx;
  padding: 20rpx 0; color: #5b8cff; font-size: 26rpx; font-weight: 500;
}
.expand-arrow { font-size: 20rpx; }

.extra-form { background: #fafbfc; border-radius: 16rpx; padding: 24rpx; margin-bottom: 8rpx; }
.extra-label { font-size: 24rpx; color: #969799; margin-bottom: 20rpx; display: block; }
.optional { color: #c0c4cc; font-size: 22rpx; }

.extra-item {
  display: flex; align-items: center; justify-content: space-between;
  padding: 18rpx 0; border-bottom: 1rpx solid #ebedf0;
}
.extra-item-last { border-bottom: none; }

.radio-group { display: flex; gap: 32rpx; }
.radio-item { font-size: 26rpx; color: #c0c4cc; }
.radio-sm { font-size: 24rpx; }
.radio-item.active { color: #07c160; font-weight: 500; }
.picker-text { font-size: 28rpx; color: #c0c4cc; }

/* ④ 证件上传 */
.upload-row { display: flex; gap: 20rpx; margin-bottom: 20rpx; }
.upload-box {
  flex: 1; height: 220rpx; background: #f8f9fc; border-radius: 16rpx;
  border: 2rpx dashed #d0d5dd; overflow: hidden;
}
.upload-preview { width: 100%; height: 100%; }
.upload-placeholder {
  height: 100%; display: flex; flex-direction: column;
  align-items: center; justify-content: center; gap: 6rpx;
}
.upload-icon { font-size: 48rpx; opacity: 0.5; }
.upload-title { font-size: 24rpx; color: #969799; }
.upload-hint { font-size: 22rpx; color: #c0c4cc; }

.checkbox-row {
  display: flex; align-items: center; gap: 12rpx; padding: 16rpx 0; margin-bottom: 8rpx;
}
.checkbox-box {
  width: 36rpx; height: 36rpx; border: 2rpx solid #d0d5dd; border-radius: 8rpx;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.checkbox-box.checked { background: #07c160; border-color: #07c160; }
.check-mark { color: #ffffff; font-size: 24rpx; font-weight: 700; }
.checkbox-text { font-size: 26rpx; color: #4a4a4a; }

/* ⑤ 费用明细 */
.price-section { border-top: 1rpx solid #ebedf0; padding-top: 24rpx; margin-bottom: 32rpx; }
.price-row { display: flex; justify-content: space-between; align-items: center; padding: 8rpx 0; }
.price-label { font-size: 26rpx; color: #969799; }
.price-value { font-size: 44rpx; font-weight: 700; color: #ff6b35; }
.price-unit { font-size: 26rpx; color: #1a1a1a; }

/* ⑥ 底部按钮 */
.bottom-bar { display: flex; gap: 20rpx; }
.bottom-btn {
  flex: 1; height: 96rpx; border-radius: 48rpx;
  display: flex; align-items: center; justify-content: center;
  font-size: 32rpx; font-weight: 600;
}
.bottom-btn.primary { background: linear-gradient(135deg, #07c160, #05a854); color: #ffffff; }
.bottom-btn.outline { border: 2rpx solid #07c160; background: #ffffff; color: #07c160; }
.privacy-text { display: block; text-align: center; font-size: 22rpx; color: #c0c4cc; margin-top: 16rpx; }
</style>
