<template>
  <view class="resume-page">
    <!-- Nav -->
    <u-nav-bar
      title="我的简历"
      show-back
      @back="goBack"
    />

    <!-- Loading state -->
    <view v-if="loading" class="loading-state">
      <view class="loading-inline">
        <u-loading size="28rpx" />
        <text>加载中...</text>
      </view>
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg && !resumeLoaded" class="error-state">
      <u-empty description="加载失败" />
      <view class="retry-btn" @tap="fetchResume">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Empty / Not logged in -->
    <view v-else-if="!isAuth" class="empty-state-wrapper">
      <u-empty description="请先登录" />
    </view>

    <!-- Normal state -->
    <template v-else>
      <view class="section-title-wrapper">
        <text class="section-title">基本信息</text>
      </view>
      <u-cell-group inset>
        <u-field
          v-model="form.name"
          label="姓名"
          placeholder="请输入姓名"
        />
        <u-field
          v-model="form.phone"
          label="手机号"
          type="number"
          placeholder="请输入手机号"
        />
        <u-field
          v-model="form.email"
          label="邮箱"
          placeholder="请输入邮箱"
        />
        <view class="field-row" @tap="showEducationPicker">
          <u-field
            :model-value="educationLabel(form.education)"
            label="学历"
            placeholder="请选择学历"
            disabled
          />
          <text class="field-arrow">›</text>
        </view>
      </u-cell-group>

      <view class="section-title-wrapper">
        <text class="section-title">工作经历</text>
      </view>
      <u-cell-group inset>
        <u-field
          v-model="form.work_experience"
          label="工作经历"
          type="textarea"
          placeholder="请描述您的工作经历"
          auto-height
        />
      </u-cell-group>

      <view class="section-title-wrapper">
        <text class="section-title">技能标签</text>
      </view>
      <u-cell-group inset>
        <view class="skills-cell">
          <view class="skills-tags">
            <view
              v-for="(skill, idx) in form.skills"
              :key="idx"
              class="skill-tag-wrap"
              @tap="removeSkill(idx)"
            >
              <u-tag type="primary">{{ skill }}</u-tag>
              <text class="skill-tag-close">×</text>
            </view>
          </view>
          <view class="skill-input-row">
            <view class="skill-input">
              <u-field
                v-model="skillInput"
                placeholder="输入技能名称"
              />
            </view>
            <u-button
              type="primary"
              size="small"
              @click="addSkill"
            >
              添加
            </u-button>
          </view>
        </view>
      </u-cell-group>

      <view class="section-title-wrapper">
        <text class="section-title">证书上传</text>
      </view>
      <u-cell-group inset>
        <view class="cert-cell">
          <view v-if="certImageUrl" class="cert-preview" @tap="previewCert">
            <image :src="certImageUrl" mode="aspectFill" class="cert-img" />
            <view class="cert-remove" @tap.stop="removeCert">
              <u-icon name="close" size="18" color="var(--color-danger)" />
            </view>
          </view>
          <view v-else class="cert-btn" @tap="chooseCert">
            <u-icon name="plus" size="28" color="#969799" />
            <text class="cert-hint">上传证书</text>
          </view>
        </view>
      </u-cell-group>

      <!-- Submit -->
      <view class="submit-area">
        <u-button
          type="primary"
          block
          :loading="saving"
          @click="handleSave"
        >
          保存简历
        </u-button>
      </view>
    </template>

    <!-- Education picker -->
    <u-popup
      :show="educationPickerShow"
      position="bottom"
      round
      @close="educationPickerShow = false"
    >
      <view class="sheet">
        <view class="sheet-title">选择学历</view>
        <view
          v-for="opt in educationOptions"
          :key="opt.value"
          class="sheet-item"
          :class="{ on: form.education === opt.value }"
          @tap="onEducationSelect(opt)"
        >
          <text class="sheet-name">{{ opt.name }}</text>
          <text v-if="form.education === opt.value" class="sheet-check">✓</text>
        </view>
        <view class="sheet-cancel" @tap="educationPickerShow = false">取消</view>
      </view>
    </u-popup>
  </view>
</template>

<script>
import { request, getStoredUser, authStorage, BASE_URL } from '../../utils/request'

export default {
  data() {
    return {
      loading: false,
      errorMsg: '',
      resumeLoaded: false,
      isAuth: false,
      resumeId: '',
      saving: false,
      educationPickerShow: false,
      skillInput: '',
      certImageUrl: '',
      form: {
        name: '',
        phone: '',
        email: '',
        education: '',
        work_experience: '',
        skills: [],
      },
      educationOptions: [
        { name: '高中', value: '高中' },
        { name: '大专', value: '大专' },
        { name: '本科', value: '本科' },
        { name: '硕士', value: '硕士' },
        { name: '博士', value: '博士' },
      ],
    }
  },
  onLoad() {
    var user = getStoredUser()
    this.isAuth = !!user
    if (this.isAuth) {
      this.fetchResume()
    }
  },
  methods: {
    async fetchResume() {
      this.loading = true
      this.errorMsg = ''

      try {
        var res = await request({ url: '/api/v1/resumes/mine' })
        var data = (res && res.data) || res || null
        if (data && (data.id || data._id)) {
          this.resumeId = data.id || data._id
          this.form.name = data.name || ''
          this.form.phone = data.phone || ''
          this.form.email = data.email || ''
          this.form.education = data.education || ''
          this.form.work_experience = data.work_experience || ''
          this.form.skills = Array.isArray(data.skills) ? data.skills : []
          this.certImageUrl = data.certificate_url || ''
        }
        this.resumeLoaded = true
      } catch (e) {
        this.errorMsg = '网络异常，请稍后重试'
      } finally {
        this.loading = false
      }
    },
    showEducationPicker() {
      this.educationPickerShow = true
    },
    onEducationSelect(opt) {
      this.form.education = opt.value
      this.educationPickerShow = false
    },
    educationLabel(value) {
      var found = this.educationOptions.find(function (o) { return o.value === value })
      return found ? found.name : value
    },
    addSkill() {
      var skill = this.skillInput.trim()
      if (!skill) return
      if (this.form.skills.indexOf(skill) !== -1) {
        uni.showToast({ title: '技能已存在', icon: 'none' })
        return
      }
      this.form.skills = this.form.skills.concat([skill])
      this.skillInput = ''
    },
    removeSkill(idx) {
      var newSkills = this.form.skills.slice()
      newSkills.splice(idx, 1)
      this.form.skills = newSkills
    },
    chooseCert() {
      var self = this
      uni.chooseImage({
        count: 1,
        sizeType: ['compressed'],
        sourceType: ['album', 'camera'],
        success: function (res) {
          self.certImageUrl = res.tempFilePaths[0]
        },
        fail: function () {
          uni.showToast({ title: '选择图片失败', icon: 'none' })
        },
      })
    },
    previewCert() {
      uni.previewImage({
        urls: [this.certImageUrl],
        current: this.certImageUrl,
      })
    },
    removeCert() {
      this.certImageUrl = ''
    },
    validateForm() {
      if (!this.form.name.trim()) {
        uni.showToast({ title: '请填写姓名', icon: 'none' })
        return false
      }
      if (!this.form.phone.trim()) {
        uni.showToast({ title: '请填写手机号', icon: 'none' })
        return false
      }
      return true
    },
    async handleSave() {
      if (!this.validateForm()) return

      this.saving = true
      try {
        var payload = {
          name: this.form.name,
          phone: this.form.phone,
          email: this.form.email,
          education: this.form.education,
          work_experience: this.form.work_experience,
          skills: this.form.skills,
        }

        // Upload certificate if selected
        if (this.certImageUrl && this.certImageUrl.indexOf('https') !== 0) {
          uni.showLoading({ title: '上传中...' })
          try {
            var uploadRes = await new Promise(function (resolve, reject) {
              uni.uploadFile({
                url: BASE_URL + '/api/v1/files/upload',
                filePath: this.certImageUrl,
                name: 'file',
                header: { Authorization: 'Bearer ' + authStorage.getAccessToken() },
                success: function (r) {
                  var data = null
                  try { data = JSON.parse(r.data) } catch (e) {}
                  // 非 2xx 或缺少 file_id（含 data 信封内层）：透出后端具体错误
                  if (r.statusCode >= 400 || !data || (!data.file_id && !(data.data && data.data.file_id))) {
                    var msg = ''
                    if (data && data.error) msg = data.error.message || data.error.code || ''
                    if (data && data.message) msg = data.message
                    resolve({ _error: msg || ('HTTP ' + r.statusCode) })
                    return
                  }
                  resolve(data)
                },
                fail: reject,
              })
            }.bind(this))
            uni.hideLoading()
            // /api/v1/files/upload 返回 {data:{file_id,...},request_id}（信封格式）
            var fid = uploadRes && (uploadRes.file_id || (uploadRes.data && uploadRes.data.file_id))
            if (!fid) {
              var reason = (uploadRes && uploadRes._error) || ''
              var tip = reason.indexOf('401') >= 0 || reason.indexOf('登录') >= 0 || reason.indexOf('token') >= 0
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
        } else if (this.certImageUrl) {
          payload.certificate_url = this.certImageUrl
        }

        var url = this.resumeId
          ? '/api/v1/resumes/' + encodeURIComponent(this.resumeId)
          : '/api/v1/resumes'
        var method = this.resumeId ? 'PATCH' : 'POST'

        await request({ url: url, method: method, data: payload })
        uni.showToast({ title: '保存成功', icon: 'success' })
        setTimeout(function () {
          uni.navigateBack()
        }, 1200)
      } catch (e) {
        var msg = (e && e.data && e.data.message) || '保存失败，请重试'
        uni.showToast({ title: msg, icon: 'none' })
      } finally {
        this.saving = false
      }
    },
    goBack() {
      uni.navigateBack()
    },
  },
}
</script>

<style scoped>
.resume-page {
  min-height: 100vh;
  background: var(--color-bg);
  padding-bottom: 80px;
}

.loading-state {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}

.loading-inline {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 13px;
  color: var(--color-text-secondary);
}

/* 只读选择字段 */
.field-row {
  position: relative;
  padding: 8px 0;
}

.field-row .u-field {
  padding-right: 56rpx;
}

.field-arrow {
  position: absolute;
  right: 24rpx;
  top: 50%;
  transform: translateY(-50%);
  font-size: 34rpx;
  color: var(--color-text-placeholder);
  line-height: 1;
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
  background: var(--color-primary);
  color: #fff;
  border-radius: 8px;
  font-size: 14px;
}

.empty-state-wrapper {
  padding-top: 60px;
}

.section-title-wrapper {
  padding: 16px 16px 8px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
}

/* Skills */
.skills-cell {
  padding: 12px 16px;
}

.skills-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}

.skill-tag-wrap {
  display: inline-flex;
  align-items: center;
  gap: 4rpx;
}

.skill-tag-close {
  font-size: 24rpx;
  color: var(--color-text-placeholder);
  padding: 4rpx;
  margin-right: 4rpx;
}

.skill-input-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.skill-input-row .skill-input {
  flex: 1;
}

/* Certificate */
.cert-cell {
  padding: 12px 16px;
}

.cert-btn {
  width: 80px;
  height: 80px;
  border: 1px dashed var(--color-text-placeholder);
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.cert-hint {
  font-size: 12px;
  color: var(--color-text-secondary);
}

.cert-preview {
  position: relative;
  width: 200px;
  height: 120px;
  border-radius: 8px;
  overflow: hidden;
}

.cert-img {
  width: 100%;
  height: 100%;
}

.cert-remove {
  position: absolute;
  top: 4px;
  right: 4px;
  background: rgba(255, 255, 255, 0.9);
  border-radius: 50%;
  padding: 2px;
}

/* Submit */
.submit-area {
  padding: 24px 16px;
}

/* 学历选择弹层 */
.sheet {
  background: #fff;
  border-radius: 16rpx 24rpx 0 0;
  padding-bottom: env(safe-area-inset-bottom);
}

.sheet-title {
  text-align: center;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
  padding: 16px 0 8px;
}

.sheet-item {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 8px;
  padding: 14px 24px;
  font-size: 14px;
  color: var(--color-text);
}

.sheet-item.on {
  color: var(--color-primary);
  font-weight: 600;
}

.sheet-name {
  color: inherit;
}

.sheet-check {
  font-size: 14px;
}

.sheet-cancel {
  text-align: center;
  padding: 14px;
  font-size: 14px;
  color: var(--color-text-secondary);
  border-top: 1px solid var(--color-divider);
}
</style>
