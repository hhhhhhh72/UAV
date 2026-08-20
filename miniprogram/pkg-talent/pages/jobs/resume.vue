<template>
  <view class="pub-page" :style="{ paddingTop: topPad + 'px' }">
    <!-- 顶栏（与发布页同款） -->
    <view class="pub-nav">
      <view class="pub-back" hover-class="pub-fade" @tap="goBack">‹</view>
      <view class="pub-nav-title">我的简历</view>
    </view>

    <!-- 加载中 -->
    <view v-if="loading" class="loading-state">
      <u-loading size="28rpx" />
      <text>加载中...</text>
    </view>

    <!-- 加载失败 -->
    <view v-else-if="errorMsg && !resumeLoaded" class="pub-empty">
      <view class="pub-empty-mark">!</view>
      <view class="pub-empty-title">加载失败</view>
      <view class="pub-empty-desc">网络异常，请稍后重试</view>
      <view class="pub-btn pub-btn--primary retry-btn" hover-class="pub-btn--active" @tap="fetchResume">重新加载</view>
    </view>

    <!-- 未登录 -->
    <view v-else-if="!isAuth" class="pub-empty">
      <view class="pub-empty-title">请先登录</view>
      <view class="pub-empty-desc">登录后可创建并编辑简历</view>
    </view>

    <!-- 表单 -->
    <template v-else>
      <view class="pub-form-intro">
        <view class="pub-form-intro-h2">我的简历</view>
        <view class="pub-form-intro-p">完善简历信息，投递职位时自动展示给招聘方</view>
      </view>

      <!-- 基本信息 -->
      <view class="pub-section">
        <view class="pub-section-title">基本信息</view>
        <view class="pub-form-card">
          <view class="pub-field">
            <view class="pub-field-label">姓名<text class="pub-required">*</text></view>
            <input
              v-model="form.name"
              class="pub-input"
              placeholder="请输入姓名"
              placeholder-class="pub-placeholder"
            />
          </view>
          <view class="pub-field">
            <view class="pub-field-label">手机号<text class="pub-required">*</text></view>
            <input
              v-model="form.phone"
              class="pub-input"
              type="number"
              placeholder="请输入手机号"
              placeholder-class="pub-placeholder"
            />
          </view>
          <view class="pub-field">
            <view class="pub-field-label">邮箱</view>
            <input
              v-model="form.email"
              class="pub-input"
              placeholder="请输入邮箱"
              placeholder-class="pub-placeholder"
            />
          </view>
          <view class="pub-field" @tap="showEducationPicker">
            <view class="pub-field-label">学历</view>
            <view class="pub-select-field">
              <text :class="form.education ? 'pub-select-value' : 'pub-placeholder'">
                {{ form.education || '请选择学历' }}
              </text>
              <text class="pub-arrow">›</text>
            </view>
          </view>
        </view>
      </view>

      <!-- 工作经历 -->
      <view class="pub-section">
        <view class="pub-section-title">工作经历</view>
        <view class="pub-form-card">
          <view class="pub-field">
            <textarea
              v-model="form.work_experience"
              class="pub-input pub-input--textarea"
              placeholder="请描述您的工作经历"
              placeholder-class="pub-placeholder"
              auto-height
            />
          </view>
        </view>
      </view>

      <!-- 技能标签 -->
      <view class="pub-section">
        <view class="pub-section-title">技能标签</view>
        <view class="pub-section-note">可添加多个技能，如：无人机飞行、航拍测绘</view>
        <view class="pub-form-card">
          <view v-if="form.skills.length" class="skill-tags">
            <view
              v-for="skill in form.skills"
              :key="skill"
              class="skill-tag"
              @tap="removeSkill(form.skills.indexOf(skill))"
            >
              <text>{{ skill }}</text>
              <text class="skill-tag-x">×</text>
            </view>
          </view>
          <view class="pub-field skill-add-row">
            <input
              v-model="skillInput"
              class="pub-input"
              placeholder="输入技能名称"
              placeholder-class="pub-placeholder"
              @confirm="addSkill"
            />
            <view class="pub-btn pub-btn--ghost skill-add-btn" hover-class="pub-btn--active" @tap="addSkill">添加</view>
          </view>
        </view>
      </view>

      <!-- 证书上传 -->
      <view class="pub-section">
        <view class="pub-section-title">证书上传</view>
        <view class="pub-section-note">上传无人机相关资质证书，提升求职竞争力</view>
        <view class="pub-form-card">
          <view class="pub-upload-row">
            <view v-if="certImageUrl" class="pub-photo" @tap="previewCert">
              <image :src="fullUrl(certImageUrl)" mode="aspectFill" class="pub-photo-img" />
              <view class="pub-photo-remove" @tap.stop="removeCert">×</view>
            </view>
            <view class="pub-add-photo" hover-class="pub-fade" @tap="chooseCert">＋</view>
          </view>
          <view class="pub-upload-tip">支持上传 1 张证书图片，点击图片可预览</view>
        </view>
      </view>
    </template>

    <!-- 固定底部操作区（与发布页同款） -->
    <view v-if="!loading && !(errorMsg && !resumeLoaded) && isAuth" class="pub-sticky">
      <view class="pub-btn pub-btn--primary" hover-class="pub-btn--active" @tap="handleSave">
        {{ saving ? '保存中...' : '保存简历' }}
      </view>
    </view>

    <!-- 学历选择抽屉（与发布页同款） -->
    <view v-if="educationPickerShow" class="pub-overlay" @tap="closeEducation">
      <view class="pub-sheet" @tap.stop>
        <view class="pub-grab"></view>
        <view class="pub-sheet-head">
          <view class="pub-sheet-head-title">选择学历</view>
          <view class="pub-sheet-cancel" @tap="closeEducation">取消</view>
        </view>
        <view
          v-for="opt in educationOptions"
          :key="opt.value"
          class="pub-option"
          :class="{ 'pub-option--selected': form.education === opt.value }"
          @tap="onEducationSelect(opt)"
        >
          <text>{{ opt.name }}</text>
          <text v-if="form.education === opt.value" class="pub-option-check">✓</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { onLoad, onUnload } from '@dcloudio/uni-app'
import { request, getStoredUser, authStorage, BASE_URL, getErrorMessage } from '../../../utils/request'
import { useSafeTop } from '../../../utils/safeTop'

const { topPad, initSafeTop } = useSafeTop(true)

// 证书图片完整地址：服务端相对路径拼 BASE_URL，本地临时路径原样返回
const fullUrl = (u) => {
  if (!u) return ''
  if (u.startsWith('http') || u.startsWith('wxfile') || u.startsWith('blob:')) return u
  return BASE_URL + u
}

const loading = ref(false)
const errorMsg = ref('')
const resumeLoaded = ref(false)
const isAuth = ref(false)
const resumeId = ref('')
const saving = ref(false)
const educationPickerShow = ref(false)
const skillInput = ref('')
const certImageUrl = ref('')
let backTimer = null

const form = reactive({
  name: '',
  phone: '',
  email: '',
  education: '',
  work_experience: '',
  skills: [],
})

const educationOptions = [
  { name: '高中', value: '高中' },
  { name: '大专', value: '大专' },
  { name: '本科', value: '本科' },
  { name: '硕士', value: '硕士' },
  { name: '博士', value: '博士' },
]

onLoad(() => {
  initSafeTop()
  const user = getStoredUser()
  isAuth.value = !!user
  if (isAuth.value) {
    fetchResume()
  }
})

async function fetchResume() {
  loading.value = true
  errorMsg.value = ''

  try {
    const res = await request({ url: '/api/v1/resumes/mine' })
    // GET /api/v1/resumes/mine 经 request.js unwrap 后是数组，取第一条
    const arr = Array.isArray(res) ? res : ((res && res.data) || [])
    const data = arr[0] || {}
    if (data && (data.id || data._id)) {
      resumeId.value = data.id || data._id
      form.name = data.name || ''
      form.phone = data.phone || ''
      form.email = data.email || ''
      form.education = data.education || ''
      form.work_experience = data.work_experience || ''
      form.skills = Array.isArray(data.skills) ? data.skills : []
      certImageUrl.value = data.certificate_url || ''
    }
    resumeLoaded.value = true
  } catch (e) {
    errorMsg.value = '网络异常，请稍后重试'
  } finally {
    loading.value = false
  }
}

function showEducationPicker() {
  educationPickerShow.value = true
}

function closeEducation() {
  educationPickerShow.value = false
}

function onEducationSelect(opt) {
  form.education = opt.value
  educationPickerShow.value = false
}

function addSkill() {
  const skill = skillInput.value.trim()
  if (!skill) return
  if (form.skills.indexOf(skill) !== -1) {
    uni.showToast({ title: '技能已存在', icon: 'none' })
    return
  }
  form.skills = form.skills.concat([skill])
  skillInput.value = ''
}

function removeSkill(idx) {
  const next = form.skills.slice()
  next.splice(idx, 1)
  form.skills = next
}

function chooseCert() {
  uni.chooseImage({
    count: 1,
    sizeType: ['compressed'],
    sourceType: ['album', 'camera'],
    success: (res) => {
      certImageUrl.value = res.tempFilePaths[0]
    },
    fail: () => {
      uni.showToast({ title: '选择图片失败', icon: 'none' })
    },
  })
}

function previewCert() {
  uni.previewImage({
    urls: [fullUrl(certImageUrl.value)],
    current: fullUrl(certImageUrl.value),
  })
}

function removeCert() {
  certImageUrl.value = ''
}

function validateForm() {
  if (!form.name.trim()) {
    uni.showToast({ title: '请填写姓名', icon: 'none' })
    return false
  }
  if (!form.phone.trim()) {
    uni.showToast({ title: '请填写手机号', icon: 'none' })
    return false
  }
  return true
}

async function handleSave() {
  if (saving.value) return
  if (!validateForm()) return

  saving.value = true
  try {
    const payload = {
      name: form.name,
      phone: form.phone,
      email: form.email,
      education: form.education,
      work_experience: form.work_experience,
      skills: form.skills,
    }

    // 新选的本地图片先上传，成功后存证书路径（服务端已存的 /uploads/ 路径直接透传）
    const isLocalCert = certImageUrl.value
      && certImageUrl.value.indexOf('https') !== 0
      && certImageUrl.value.indexOf('/uploads/') !== 0
    if (isLocalCert) {
      uni.showLoading({ title: '上传中...' })
      try {
        const uploadRes = await new Promise((resolve, reject) => {
          uni.uploadFile({
            url: BASE_URL + '/api/v1/files/upload',
            filePath: certImageUrl.value,
            name: 'file',
            header: { Authorization: 'Bearer ' + authStorage.getAccessToken() },
            success: (r) => {
              let data = null
              try { data = JSON.parse(r.data) } catch (e) {}
              // 非 2xx 或缺少 file_id（含 data 信封内层）：透出后端具体错误
              if (r.statusCode >= 400 || !data || (!data.file_id && !(data.data && data.data.file_id))) {
                let msg = ''
                if (data && data.error) msg = data.error.message || data.error.code || ''
                if (data && data.message) msg = data.message
                resolve({ _error: msg || ('HTTP ' + r.statusCode) })
                return
              }
              resolve(data)
            },
            fail: reject,
          })
        })
        uni.hideLoading()
        // /api/v1/files/upload 返回 {data:{file_id,...},request_id}（信封格式）
        const fid = uploadRes && (uploadRes.file_id || (uploadRes.data && uploadRes.data.file_id))
        if (!fid) {
          const reason = (uploadRes && uploadRes._error) || ''
          const tip = reason.indexOf('401') >= 0 || reason.indexOf('登录') >= 0 || reason.indexOf('token') >= 0
            ? '登录已过期，请重新登录后重试'
            : ('证书上传失败：' + (reason || '请重试'))
          uni.showToast({ title: tip, icon: 'none', duration: 2500 })
          return
        }
        payload.certificate_url = '/uploads/' + fid
      } catch (uploadErr) {
        uni.hideLoading()
        uni.showToast({ title: '证书上传失败，请重试', icon: 'none' })
        return
      }
    } else if (certImageUrl.value) {
      payload.certificate_url = certImageUrl.value
    }

    const url = resumeId.value
      ? '/api/v1/resumes/' + encodeURIComponent(resumeId.value)
      : '/api/v1/resumes'
    const method = resumeId.value ? 'PATCH' : 'POST'

    await request({ url: url, method: method, data: payload })
    uni.showToast({ title: '保存成功', icon: 'success' })
    backTimer = setTimeout(() => {
      uni.navigateBack()
    }, 1200)
  } catch (e) {
    const msg = getErrorMessage(e) || '保存失败，请重试'
    uni.showToast({ title: msg, icon: 'none' })
  } finally {
    saving.value = false
  }
}

function goBack() {
  uni.navigateBack()
}

onUnload(() => {
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
.pub-form-intro-p {
  font-size: 12px;
  color: #667085;
  margin: 0;
  line-height: 1.5;
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

/* 技能标签 */
.skill-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
  padding: 13px 13px 0;
}
.skill-tag {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  color: #0A66C2;
  background: #E8F2FC;
  border-radius: 5px;
  padding: 5px 8px;
  font-size: 12px;
  font-weight: 700;
}
.skill-tag-x {
  color: #98A2B3;
  font-size: 13px;
  line-height: 1;
}
.skill-add-row {
  display: flex;
  align-items: center;
  gap: 9px;
}
.skill-add-row .pub-input {
  flex: 1;
  min-width: 0;
}
.skill-add-btn {
  flex: none;
  min-height: 34px;
  padding: 0 14px;
  font-size: 13px;
}

/* 证书预览删除角标 */
.pub-photo-remove {
  position: absolute;
  top: 3px;
  right: 3px;
  width: 17px;
  height: 17px;
  border-radius: 50%;
  background: rgba(23, 33, 43, 0.55);
  color: #fff;
  font-size: 12px;
  line-height: 17px;
  text-align: center;
}
</style>
