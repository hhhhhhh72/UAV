<template>
  <view class="register-page">
    <u-nav-bar
      title="企业入驻"
      show-back
      @back="goBack"
    />

    <!-- Step indicator -->
    <view class="step-indicator">
      <view
        v-for="(step, index) in steps"
        :key="index"
        class="step-item"
        :class="{ active: currentStep >= index, done: currentStep > index }"
      >
        <view class="step-circle">{{ currentStep > index ? '' : (index + 1) }}</view>
        <text class="step-label">{{ step }}</text>
        <view v-if="index < steps.length - 1" class="step-line" :class="{ filled: currentStep > index }" />
      </view>
    </view>

    <!-- Step 1: 基本信息（PRD FR-2.1：名称/信用代码/logo/成立时间/注册地/简介） -->
    <view v-if="currentStep === 0" class="step-body">
      <view class="section-title">企业档案</view>
      <u-cell-group inset>
        <!-- logo 上传 -->
        <view class="upload-cell">
          <text class="upload-label">企业 logo</text>
          <view class="upload-area">
            <view v-if="logoUrl" class="logo-preview" @tap="previewLogo">
              <image :src="logoUrl" mode="aspectFill" class="logo-img" />
              <view class="upload-remove" @tap.stop="removeLogo">
                <u-icon name="close" size="24rpx" color="var(--color-danger)" />
              </view>
            </view>
            <view v-else class="logo-upload-btn" @tap="chooseLogo">
              <u-icon name="plus" size="28rpx" color="var(--color-text-secondary)" />
              <text class="upload-hint">上传</text>
            </view>
          </view>
        </view>
        <u-field
          v-model="form.name"
          label="企业名称"
          placeholder="请输入企业名称"
        />
        <u-field
          v-model="form.credit_code"
          label="信用代码"
          placeholder="统一社会信用代码"
        />
        <picker mode="date" fields="month" :value="form.founded_at" @change="onDateChange">
          <view class="field-row">
            <u-field
              :model-value="form.founded_at"
              label="成立时间"
              placeholder="请选择成立时间"
              disabled
            />
            <text class="field-arrow">›</text>
          </view>
        </picker>
        <u-field
          v-model="form.address"
          label="注册地"
          placeholder="请输入办公地址"
        />
        <u-field
          v-model="form.description"
          label="企业简介"
          type="textarea"
          auto-height
          placeholder="简要介绍企业经营范围与能力（将公开展示）"
        />
      </u-cell-group>
    </view>

    <!-- Step 2: 分类与能力（PRD FR-2.1：8 类多选 + 预设能力标签） -->
    <view v-if="currentStep === 1" class="step-body">
      <view class="section-title">企业分类</view>
      <view class="chips-card">
        <view
          v-for="opt in categoryOptions"
          :key="opt.value"
          class="chip"
          :class="{ 'chip--active': form.industry_categories.includes(opt.value) }"
          @tap="toggleCategory(opt.value)"
        >
          {{ opt.name }}
        </view>
      </view>
      <view class="section-title">能力标签（多选）</view>
      <view class="chips-card">
        <view
          v-for="opt in tagOptions"
          :key="opt.value"
          class="chip"
          :class="{ 'chip--active': form.capability_tags.includes(opt.value) }"
          @tap="toggleTag(opt.value)"
        >
          {{ opt.name }}
        </view>
      </view>
      <view class="scale-card">
        <u-cell-group inset>
          <u-field
            v-model="form.scale"
            label="企业规模"
            placeholder="如：50-100人"
          />
        </u-cell-group>
      </view>
    </view>

    <!-- Step 3: 联系与资质（PRD FR-2.1：法人/联系人/电话/邮箱 + 营业执照） -->
    <view v-if="currentStep === 2" class="step-body">
      <view class="section-title">联系信息</view>
      <u-cell-group inset>
        <u-field
          v-model="form.legal_person"
          label="法人代表"
          placeholder="请输入法人代表姓名"
        />
        <u-field
          v-model="form.contact_person"
          label="联系人"
          placeholder="请输入联系人姓名"
        />
        <u-field
          v-model="form.contact_phone"
          label="联系电话"
          type="number"
          placeholder="请输入联系电话"
        />
        <u-field
          v-model="form.email"
          label="邮箱"
          placeholder="请输入企业邮箱"
        />
      </u-cell-group>
      <view class="section-title">企业资质</view>
      <u-cell-group inset>
        <view class="upload-cell">
          <text class="upload-label">营业执照</text>
          <view class="upload-area">
            <view v-if="licenseUrl" class="upload-preview" @tap="previewLicense">
              <image :src="licenseUrl" mode="aspectFill" class="upload-img" />
              <view class="upload-remove" @tap.stop="removeLicense">
                <u-icon name="close" size="24rpx" color="var(--color-danger)" />
              </view>
            </view>
            <view v-else class="upload-btn" @tap="chooseLicense">
              <u-icon name="plus" size="28rpx" color="var(--color-text-secondary)" />
              <text class="upload-hint">点击上传</text>
            </view>
          </view>
        </view>
        <view class="upload-tip">营业执照用于审核，审核通过后仅管理员可见</view>
      </u-cell-group>
    </view>

    <!-- Bottom action buttons -->
    <view class="action-bar">
      <u-button
        v-if="currentStep > 0"
        type="default"
        round
        class="prev-btn"
        @click="prevStep"
      >
        上一步
      </u-button>
      <u-button
        v-if="currentStep < 2"
        type="primary"
        round
        class="next-btn"
        @click="nextStep"
      >
        下一步
      </u-button>
      <u-button
        v-else
        type="primary"
        round
        class="submit-btn"
        :loading="submitting"
        @click="handleSubmit"
      >
        {{ isEdit ? '保存并重新提交' : '提交审核' }}
      </u-button>
    </view>

  </view>
</template>

<script>
import { request, authStorage, BASE_URL } from '../../../utils/request'

// PRD FR-2.1 企业分类（多选标签）
const CATEGORY_OPTIONS = [
  { name: '整机研发', value: '整机研发' },
  { name: '零部件制造', value: '零部件制造' },
  { name: '飞控系统', value: '飞控系统' },
  { name: '载荷设备', value: '载荷设备' },
  { name: '运营服务', value: '运营服务' },
  { name: '实训院校', value: '实训院校' },
  { name: '通航机场', value: '通航机场' },
  { name: '检测机构', value: '检测机构' },
]

// PRD FR-2.1 能力标签（预设标签库）
const TAG_OPTIONS = [
  { name: '航拍服务', value: '航拍服务' },
  { name: '农林植保', value: '农林植保' },
  { name: '电力巡检', value: '电力巡检' },
  { name: '测绘勘察', value: '测绘勘察' },
  { name: '应急救援', value: '应急救援' },
  { name: '物流配送', value: '物流配送' },
  { name: '巡防安防', value: '巡防安防' },
  { name: '消防救援', value: '消防救援' },
  { name: '桥梁检测', value: '桥梁检测' },
  { name: '环保监测', value: '环保监测' },
]

export default {
  data() {
    return {
      currentStep: 0,
      steps: ['基本信息', '分类能力', '联系资质'],
      submitting: false,
      categoryOptions: CATEGORY_OPTIONS,
      tagOptions: TAG_OPTIONS,
      form: {
        name: '',
        credit_code: '',
        legal_person: '',
        contact_person: '',
        contact_phone: '',
        email: '',
        industry_categories: [],
        capability_tags: [],
        scale: '',
        address: '',
        founded_at: '',
        description: '',
      },
      logoUrl: '',
      licenseUrl: '',
      // 编辑模式：status 页「需补充资料/驳回」后带 entId 进入
      editEntId: '',
    }
  },
  computed: {
    isEdit() {
      return !!this.editEntId
    },
  },
  onLoad(options) {
    // 登录守卫：入驻表单提交需要 token，未登录先引导登录
    if (!uni.getStorageSync('accessToken')) {
      uni.navigateTo({ url: '/pages/login/index' })
      return
    }
    if (options && options.entId) {
      this.editEntId = options.entId
      uni.setNavigationBarTitle({ title: '编辑企业资料' })
      this.loadEnterprise()
    }
  },
  methods: {
    goBack() {
      uni.navigateBack()
    },
    // ---- 编辑模式：预填已有资料 ----
    async loadEnterprise() {
      try {
        var res = await request({ url: '/api/v1/enterprises' })
        var data = (res && res.data) || res || {}
        var items = Array.isArray(data) ? data : (data && data.items) || []
        var ent = null
        for (var i = 0; i < items.length; i++) {
          if (items[i].id === this.editEntId) {
            ent = items[i]
            break
          }
        }
        if (!ent) {
          uni.showToast({ title: '未找到企业资料', icon: 'none' })
          return
        }
        this.form.name = ent.name || ''
        this.form.credit_code = ent.credit_code || ''
        this.form.legal_person = ent.legal_person || ''
        this.form.contact_person = ent.contact_person || ''
        this.form.contact_phone = ent.contact_phone || ''
        this.form.email = ent.email || ''
        this.form.industry_categories = this.splitTags(ent.industry_category)
        this.form.capability_tags = this.splitTags(ent.capability_tags)
        this.form.scale = ent.scale || ''
        this.form.address = ent.address || ''
        this.form.founded_at = ent.founded_at || ''
        this.form.description = ent.description || ''
        this.logoUrl = this.resolveUrl(ent.logo)
        this.licenseUrl = this.resolveUrl(ent.license_url)
      } catch (e) {
        uni.showToast({ title: '加载企业资料失败', icon: 'none' })
      }
    },
    splitTags(str) {
      if (!str) return []
      return String(str).split(',').map(function (t) { return t.trim() }).filter(Boolean)
    },
    // 相对路径（存库格式）→ 完整 URL（预览格式）
    resolveUrl(u) {
      if (!u) return ''
      if (u.indexOf('http') === 0) return u
      return BASE_URL + u
    },
    // 完整 URL（预览格式）→ 相对路径（存库格式）
    storageUrl(u) {
      if (!u) return ''
      if (u.indexOf(BASE_URL) === 0) return u.slice(BASE_URL.length)
      return u
    },
    // ---- 步骤切换 ----
    prevStep() {
      if (this.currentStep > 0) {
        this.currentStep--
      }
    },
    nextStep() {
      if (this.currentStep === 0) {
        if (!this.form.name) {
          return uni.showToast({ title: '请填写企业名称', icon: 'none' })
        }
        if (!this.form.credit_code) {
          return uni.showToast({ title: '请填写信用代码', icon: 'none' })
        }
      }
      if (this.currentStep === 1 && this.form.industry_categories.length === 0) {
        return uni.showToast({ title: '请至少选择一个企业分类', icon: 'none' })
      }
      this.currentStep++
    },
    // ---- 分类 / 能力标签多选 ----
    toggleCategory(value) {
      var list = this.form.industry_categories.slice()
      var idx = list.indexOf(value)
      if (idx >= 0) {
        list.splice(idx, 1)
      } else {
        list.push(value)
      }
      this.form.industry_categories = list
    },
    toggleTag(value) {
      var list = this.form.capability_tags.slice()
      var idx = list.indexOf(value)
      if (idx >= 0) {
        list.splice(idx, 1)
      } else {
        list.push(value)
      }
      this.form.capability_tags = list
    },
    // ---- 成立时间 ----
    onDateChange(e) {
      this.form.founded_at = e.detail.value
    },
    // ---- 图片上传（logo / 营业执照 共用） ----
    // 注意：预览必须用完整 URL（BASE_URL + 相对路径），
    // 小程序 image 的 src 若为 /uploads/xxx 会被当作本地包内资源 → 白图
    chooseLogo() {
      var self = this
      this.uploadImage(function (path) {
        self.logoUrl = self.resolveUrl(path)
      })
    },
    chooseLicense() {
      var self = this
      this.uploadImage(function (path) {
        self.licenseUrl = self.resolveUrl(path)
      })
    },
    uploadImage(onSuccess) {
      var self = this
      uni.chooseImage({
        count: 1,
        sizeType: ['compressed'],
        sourceType: ['album', 'camera'],
        success: function (res) {
          uni.showLoading({ title: '上传中...' })
          uni.uploadFile({
            url: BASE_URL + '/api/v1/files/upload',
            filePath: res.tempFilePaths[0],
            name: 'file',
            header: { Authorization: 'Bearer ' + authStorage.getAccessToken() },
            success: function (r) {
              uni.hideLoading()
              var data = null
              try { data = JSON.parse(r.data) } catch (e) {}
              if (r.statusCode >= 400 || !data || (!data.file_id && !(data.data && data.data.file_id))) {
                var msg = ''
                if (data && data.error) msg = data.error.message || data.error.code || ''
                if (data && data.message) msg = data.message
                var reason = msg || ('HTTP ' + r.statusCode)
                var tip = reason.indexOf('401') >= 0 || reason.indexOf('登录') >= 0 || reason.indexOf('token') >= 0
                  ? '登录已过期，请重新登录后重试'
                  : ('上传失败：' + reason)
                uni.showToast({ title: tip, icon: 'none', duration: 2500 })
                return
              }
              var fid = data.file_id || (data.data && data.data.file_id)
              if (!fid) {
                uni.showToast({ title: '上传失败，请重试', icon: 'none' })
                return
              }
              onSuccess('/uploads/' + fid)
            },
            fail: function () {
              uni.hideLoading()
              uni.showToast({ title: '上传失败，请重试', icon: 'none' })
            },
          })
        },
        fail: function () {
          uni.showToast({ title: '选择图片失败', icon: 'none' })
        },
      })
    },
    previewLogo() {
      uni.previewImage({ urls: [this.logoUrl], current: this.logoUrl })
    },
    previewLicense() {
      uni.previewImage({ urls: [this.licenseUrl], current: this.licenseUrl })
    },
    removeLogo() {
      this.logoUrl = ''
    },
    removeLicense() {
      this.licenseUrl = ''
    },
    // ---- 提交 ----
    async handleSubmit() {
      var self = this
      self.submitting = true

      try {
        var payload = {
          name: self.form.name,
          credit_code: self.form.credit_code,
          legal_person: self.form.legal_person,
          contact_person: self.form.contact_person,
          contact_phone: self.form.contact_phone,
          email: self.form.email,
          industry_category: self.form.industry_categories.join(','),
          capability_tags: self.form.capability_tags.join(','),
          scale: self.form.scale,
          address: self.form.address,
          founded_at: self.form.founded_at,
          description: self.form.description,
          logo: self.storageUrl(self.logoUrl),
          license_url: self.storageUrl(self.licenseUrl),
        }

        var ent
        if (self.isEdit) {
          ent = await request({
            url: '/api/v1/enterprises/' + encodeURIComponent(self.editEntId),
            method: 'PATCH',
            data: payload,
          })
        } else {
          ent = await request({
            url: '/api/v1/enterprises',
            method: 'POST',
            data: payload,
          })
        }

        // 新建时创建为草稿 → 立即提交审核；编辑时（draft/supplement_required）重提审核
        var entId = self.editEntId || (ent && (ent.data && ent.data.id ? ent.data.id : ent.id))
        if (entId) {
          await request({
            url: '/api/v1/enterprises/' + encodeURIComponent(entId) + '/submit',
            method: 'POST',
          })
        }

        uni.showToast({ title: self.isEdit ? '已重新提交审核' : '已提交审核' })
        setTimeout(function () {
          uni.navigateBack()
        }, 1200)
      } catch (e) {
        var msg = (e && e.data && e.data.message) || (e && e.message) || '提交失败，请重试'
        uni.showToast({ title: msg, icon: 'none', duration: 2500 })
      } finally {
        self.submitting = false
      }
    },
  },
}
</script>

<style scoped>
.register-page {
  min-height: 100vh;
  background: var(--color-bg);
  padding-bottom: 100px;
}

/* Step indicator */
.step-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px 32px 16px;
  background: var(--color-bg-card);
  margin-bottom: 12px;
}

.step-item {
  display: flex;
  align-items: center;
  position: relative;
}

.step-circle {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: var(--color-divider);
  color: var(--color-text-secondary);
  font-size: 14px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: all 0.3s;
}

.step-item.active .step-circle {
  background: var(--color-primary);
  color: #fff;
}

.step-item.done .step-circle {
  background: var(--color-success);
  color: #fff;
}

.step-item.done .step-circle::after {
  content: '\2713';
}

.step-label {
  margin-left: 6px;
  font-size: 12px;
  color: var(--color-text-secondary);
  white-space: nowrap;
}

.step-item.active .step-label {
  color: var(--color-primary);
  font-weight: 600;
}

.step-item.done .step-label {
  color: var(--color-success);
}

.step-line {
  width: 40px;
  height: 2px;
  background: var(--color-divider);
  margin: 0 10px;
  transition: all 0.3s;
}

.step-line.filled {
  background: var(--color-success);
}

/* Step body */
.step-body {
  padding: 12px 0;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text-secondary);
  padding: 8px 20px 4px;
}

/* Field row with arrow */
.field-row {
  display: flex;
  align-items: center;
  background: #fff;
  padding-right: 16px;
}

.field-arrow {
  color: var(--color-text-placeholder);
  font-size: 20px;
  flex-shrink: 0;
}

/* 多选 chips */
.chips-card {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 16px;
  background: var(--color-bg-card);
  border-radius: 12px;
  margin: 0 12px 12px;
}

.chip {
  padding: 10px 18px;
  border-radius: 24rpx;
  background: var(--color-bg);
  border: 1px solid var(--color-divider);
  font-size: 13px;
  color: var(--color-text-secondary);
  transition: all 0.2s;
}

.chip--active {
  background: rgba(10, 102, 194, 0.08);
  border-color: var(--color-primary);
  color: var(--color-primary);
  font-weight: 500;
}

.scale-card {
  margin: 0 12px;
}

/* Upload area */
.upload-cell {
  padding: 12px 16px;
  display: flex;
  align-items: flex-start;
}

.upload-label {
  width: 68px;
  font-size: 14px;
  color: var(--color-text);
  flex-shrink: 0;
  line-height: 24px;
  padding-top: 8px;
}

.upload-area {
  flex: 1;
}

.upload-btn {
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

.logo-upload-btn {
  width: 72px;
  height: 72px;
  border: 1px dashed var(--color-text-placeholder);
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.upload-hint {
  font-size: 12px;
  color: var(--color-text-secondary);
}

.logo-preview {
  position: relative;
  width: 72px;
  height: 72px;
  border-radius: 12px;
  overflow: hidden;
}

.logo-img {
  width: 100%;
  height: 100%;
}

.upload-preview {
  position: relative;
  width: 200px;
  height: 120px;
  border-radius: 8px;
  overflow: hidden;
}

.upload-img {
  width: 100%;
  height: 100%;
}

.upload-remove {
  position: absolute;
  top: 4px;
  right: 4px;
  background: rgba(255, 255, 255, 0.9);
  border-radius: 50%;
  padding: 2px;
}

.upload-tip {
  padding: 0 16px 12px;
  font-size: 12px;
  color: var(--color-text-placeholder);
}

/* Action bar */
.action-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 12px 16px;
  background: var(--color-bg-card);
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.04);
  padding-bottom: calc(12px + env(safe-area-inset-bottom));
  display: flex;
  gap: 12px;
  z-index: 100;
}

.prev-btn {
  flex: 1;
}

.next-btn,
.submit-btn {
  flex: 1;
}

</style>
