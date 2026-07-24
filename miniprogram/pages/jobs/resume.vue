<template>
  <view class="resume-page">
    <!-- Nav -->
    <van-nav-bar
      title="我的简历"
      left-arrow
      fixed
      placeholder
      @click-left="goBack"
    />

    <!-- Loading state -->
    <view v-if="loading" class="loading-state">
      <van-loading size="24">加载中...</van-loading>
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg && !resumeLoaded" class="error-state">
      <van-empty description="加载失败" image="error" />
      <view class="retry-btn" @tap="fetchResume">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Empty / Not logged in -->
    <view v-else-if="!isAuth" class="empty-state-wrapper">
      <van-empty image="user-o" description="请先登录" />
    </view>

    <!-- Normal state -->
    <template v-else>
      <view class="section-title-wrapper">
        <text class="section-title">基本信息</text>
      </view>
      <van-cell-group inset>
        <van-field
          v-model="form.name"
          label="姓名"
          placeholder="请输入姓名"
          required
          :border="true"
        />
        <van-field
          v-model="form.phone"
          label="手机号"
          type="number"
          placeholder="请输入手机号"
          required
          :border="true"
        />
        <van-field
          v-model="form.email"
          label="邮箱"
          placeholder="请输入邮箱"
          :border="true"
        />
        <van-field
          :model-value="educationLabel(form.education)"
          label="学历"
          placeholder="请选择学历"
          readonly
          is-link
          :border="true"
          @click="showEducationPicker"
        />
      </van-cell-group>

      <view class="section-title-wrapper">
        <text class="section-title">工作经历</text>
      </view>
      <van-cell-group inset>
        <van-field
          v-model="form.work_experience"
          label="工作经历"
          type="textarea"
          placeholder="请描述您的工作经历"
          rows="4"
          autosize
          :border="true"
        />
      </van-cell-group>

      <view class="section-title-wrapper">
        <text class="section-title">技能标签</text>
      </view>
      <van-cell-group inset>
        <view class="skills-cell">
          <view class="skills-tags">
            <van-tag
              v-for="(skill, idx) in form.skills"
              :key="idx"
              type="primary"
              size="medium"
              closeable
              @close="removeSkill(idx)"
            >
              {{ skill }}
            </van-tag>
          </view>
          <view class="skill-input-row">
            <van-field
              v-model="skillInput"
              placeholder="输入技能名称"
              :border="false"
              custom-style="background: #f7f8fa; border-radius: 8px; padding: 4px 12px;"
            />
            <van-button
              type="primary"
              size="small"
              round
              @tap="addSkill"
            >
              添加
            </van-button>
          </view>
        </view>
      </van-cell-group>

      <view class="section-title-wrapper">
        <text class="section-title">证书上传</text>
      </view>
      <van-cell-group inset>
        <view class="cert-cell">
          <view v-if="certImageUrl" class="cert-preview" @tap="previewCert">
            <image :src="certImageUrl" mode="aspectFill" class="cert-img" />
            <view class="cert-remove" @tap.stop="removeCert">
              <van-icon name="clear" size="18" color="#ee0a24" />
            </view>
          </view>
          <view v-else class="cert-btn" @tap="chooseCert">
            <van-icon name="photograph" size="28" color="#969799" />
            <text class="cert-hint">上传证书</text>
          </view>
        </view>
      </van-cell-group>

      <!-- Submit -->
      <view class="submit-area">
        <van-button
          type="primary"
          block
          round
          :loading="saving"
          @tap="handleSave"
        >
          保存简历
        </van-button>
      </view>
    </template>

    <!-- Education picker -->
    <van-action-sheet
      :show="educationPickerShow"
      :actions="educationOptions"
      cancel-text="取消"
      @select="onEducationSelect"
      @close="educationPickerShow = false"
      @cancel="educationPickerShow = false"
    />
  </view>
</template>

<script>
import { request, getStoredUser } from '../../utils/request'

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
    onEducationSelect(e) {
      this.form.education = e.detail.value
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
                url: 'http://localhost:8080/api/v1/upload',
                filePath: this.certImageUrl,
                name: 'file',
                success: function (r) {
                  try {
                    var data = JSON.parse(r.data)
                    resolve(data)
                  } catch (e) {
                    resolve({ url: this.certImageUrl })
                  }
                },
                fail: reject,
              })
            }.bind(this))
            uni.hideLoading()
            payload.certificate_url = (uploadRes && (uploadRes.url || uploadRes.data && uploadRes.data.url)) || this.certImageUrl
          } catch (uploadErr) {
            uni.hideLoading()
            payload.certificate_url = this.certImageUrl
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
  background: #f7f8fa;
  padding-bottom: calc(env(safe-area-inset-bottom) + 40px);
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

.empty-state-wrapper {
  padding-top: 60px;
}

.section-title-wrapper {
  padding: 16px 16px 8px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: #323233;
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

.skill-input-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.skill-input-row > :first-child {
  flex: 1;
}

/* Certificate */
.cert-cell {
  padding: 12px 16px;
}

.cert-btn {
  width: 80px;
  height: 80px;
  border: 1px dashed #c8c9cc;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.cert-hint {
  font-size: 12px;
  color: #969799;
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
</style>
