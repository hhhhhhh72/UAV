<template>
  <view class="pub-page" :style="{ paddingTop: topPad + 'px' }">
    <!-- 顶栏（与发布页同款） -->
    <view class="pub-nav">
      <view class="pub-back" hover-class="pub-fade" @tap="goBack">‹</view>
      <view class="pub-nav-title">{{ isEdit ? '编辑企业资料' : '企业入驻' }}</view>
    </view>

    <!-- Step indicator（3 步向导） -->
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
    <view v-if="currentStep === 0">
      <view class="pub-section">
        <view class="pub-section-title">企业档案</view>
        <view class="pub-form-card">
          <!-- logo 上传 -->
          <view class="pub-field">
            <view class="pub-field-label">企业 logo</view>
            <view class="pub-upload-row pub-upload-inline">
              <view v-if="logoUrl" class="pub-photo" @tap="previewLogo">
                <image :src="logoUrl" mode="aspectFill" class="pub-photo-img" />
                <view class="pub-photo-remove" @tap.stop="removeLogo">×</view>
              </view>
              <view v-else class="pub-add-photo" hover-class="pub-fade" @tap="chooseLogo">＋</view>
            </view>
          </view>
          <view class="pub-field">
            <view class="pub-field-label">企业名称<text class="pub-required">*</text></view>
            <input
              v-model="form.name"
              class="pub-input"
              placeholder="请输入企业名称"
              placeholder-class="pub-placeholder"
            />
          </view>
          <view class="pub-field">
            <view class="pub-field-label">信用代码<text class="pub-required">*</text></view>
            <input
              v-model="form.credit_code"
              class="pub-input"
              placeholder="统一社会信用代码"
              placeholder-class="pub-placeholder"
            />
          </view>
          <picker mode="date" fields="month" :value="form.founded_at" @change="onDateChange">
            <view class="pub-field pub-field--pick">
              <view class="pub-field-label">成立时间</view>
              <view class="pub-select-field">
                <text :class="form.founded_at ? 'pub-select-value' : 'pub-placeholder'">
                  {{ form.founded_at || '请选择成立时间' }}
                </text>
                <text class="pub-arrow">›</text>
              </view>
            </view>
          </picker>
          <view class="pub-field">
            <view class="pub-field-label">注册地</view>
            <input
              v-model="form.address"
              class="pub-input"
              placeholder="请输入办公地址"
              placeholder-class="pub-placeholder"
            />
          </view>
          <view class="pub-field">
            <view class="pub-field-label">企业简介</view>
            <textarea
              v-model="form.description"
              class="pub-input pub-input--textarea"
              auto-height
              placeholder="简要介绍企业经营范围与能力（将公开展示）"
              placeholder-class="pub-placeholder"
            />
          </view>
        </view>
      </view>
    </view>

    <!-- Step 2: 分类与能力（PRD FR-2.1：8 类多选 + 预设能力标签） -->
    <view v-if="currentStep === 1">
      <view class="pub-section">
        <view class="pub-section-title">企业分类</view>
        <view class="pub-form-card">
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
        </view>
      </view>
      <view class="pub-section">
        <view class="pub-section-title">能力标签（多选）</view>
        <view class="pub-form-card">
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
        </view>
      </view>
      <view class="pub-section">
        <view class="pub-section-title">企业规模</view>
        <view class="pub-form-card">
          <view class="pub-field">
            <view class="pub-field-label">企业规模</view>
            <input
              v-model="form.scale"
              class="pub-input"
              placeholder="如：50-100人"
              placeholder-class="pub-placeholder"
            />
          </view>
        </view>
      </view>
    </view>

    <!-- Step 3: 联系与资质（PRD FR-2.1：法人/联系人/电话/邮箱 + 营业执照） -->
    <view v-if="currentStep === 2">
      <view class="pub-section">
        <view class="pub-section-title">联系信息</view>
        <view class="pub-form-card">
          <view class="pub-field">
            <view class="pub-field-label">法人代表</view>
            <input
              v-model="form.legal_person"
              class="pub-input"
              placeholder="请输入法人代表姓名"
              placeholder-class="pub-placeholder"
            />
          </view>
          <view class="pub-field">
            <view class="pub-field-label">联系人</view>
            <input
              v-model="form.contact_person"
              class="pub-input"
              placeholder="请输入联系人姓名"
              placeholder-class="pub-placeholder"
            />
          </view>
          <view class="pub-field">
            <view class="pub-field-label">联系电话</view>
            <input
              v-model="form.contact_phone"
              class="pub-input"
              type="number"
              placeholder="请输入联系电话"
              placeholder-class="pub-placeholder"
            />
          </view>
          <view class="pub-field">
            <view class="pub-field-label">邮箱</view>
            <input
              v-model="form.email"
              class="pub-input"
              placeholder="请输入企业邮箱"
              placeholder-class="pub-placeholder"
            />
          </view>
        </view>
      </view>
      <view class="pub-section">
        <view class="pub-section-title">企业资质</view>
        <view class="pub-form-card">
          <view class="pub-field">
            <view class="pub-field-label">营业执照</view>
            <view class="pub-upload-row pub-upload-inline">
              <view v-if="licenseUrl" class="pub-photo" @tap="previewLicense">
                <image :src="licenseUrl" mode="aspectFill" class="pub-photo-img" />
                <view class="pub-photo-remove" @tap.stop="removeLicense">×</view>
              </view>
              <view v-else class="pub-add-photo" hover-class="pub-fade" @tap="chooseLicense">＋</view>
            </view>
          </view>
          <view class="pub-upload-tip">营业执照用于审核，审核通过后仅管理员可见</view>
        </view>
      </view>
    </view>

    <!-- 固定底部操作区（与发布页同款） -->
    <view class="pub-sticky">
      <view
        v-if="currentStep > 0"
        class="pub-btn pub-btn--ghost"
        hover-class="pub-btn--active"
        @tap="prevStep"
      >
        上一步
      </view>
      <view
        v-if="currentStep < 2"
        class="pub-btn pub-btn--primary"
        hover-class="pub-btn--active"
        @tap="nextStep"
      >
        下一步
      </view>
      <view
        v-else
        class="pub-btn pub-btn--primary"
        hover-class="pub-btn--active"
        @tap="handleSubmit"
      >
        {{ submitting ? '提交中...' : (isEdit ? '保存并重新提交' : '提交审核') }}
      </view>
    </view>

  </view>
</template>

<script>
import { request, authStorage, BASE_URL } from '../../../utils/request'
import { useSafeTop } from '../../../utils/safeTop'

// 自定义顶栏安全区（与发布页同款 pub-nav）
const { topPad: safeTopPad, initSafeTop } = useSafeTop(true)

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
      topPad: 24,
      currentStep: 0,
      steps: ['基本信息', '分类能力', '联系资质'],
      submitting: false,
      backTimer: null,
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
      // 营业执照私有文件的服务端相对路径（/uploads/private/{fid}）；预览用 licenseUrl（本地临时路径）
      licensePath: '',
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
    // 顶栏安全区（pub-nav 自定义顶栏）
    initSafeTop()
    this.topPad = safeTopPad.value
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
  onUnload() {
    if (this.backTimer) clearTimeout(this.backTimer)
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
        var lic = ent.license_url || ''
        this.licensePath = lic
        // 私有文件（/uploads/private/）需 Authorization 头，<image> 无法携带 → 预览留空待重新上传
        this.licenseUrl = lic.indexOf('/uploads/private/') === 0 ? '' : this.resolveUrl(lic)
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
    // 营业执照走私有上传（formData.private=true → 落 /uploads/private/{fid}，仅鉴权可读）；
    // 私有文件 <image>/previewImage 无法携带 Authorization 头，预览用本地临时路径
    chooseLicense() {
      var self = this
      this.uploadImage(function (path, localPreview) {
        self.licensePath = path
        self.licenseUrl = localPreview || self.resolveUrl(path)
      }, true)
    },
    uploadImage(onSuccess, isPrivate) {
      var self = this
      uni.chooseImage({
        count: 1,
        sizeType: ['compressed'],
        sourceType: ['album', 'camera'],
        success: function (res) {
          var localPath = res.tempFilePaths[0]
          uni.showLoading({ title: '上传中...' })
          uni.uploadFile({
            url: BASE_URL + '/api/v1/files/upload',
            filePath: localPath,
            name: 'file',
            // P0 修复：营业执照等敏感证件走私有上传（后端 FormValue("private")=="true" → /uploads/private/）
            formData: { private: isPrivate ? 'true' : 'false' },
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
              // 优先用后端返回的 url（私有上传为 /uploads/private/{fid}），缺失时按私有标记拼路径
              var url = data.url || (data.data && data.data.url)
              var rel = url || ('/uploads/' + (isPrivate ? 'private/' : '') + fid)
              onSuccess(rel, isPrivate ? localPath : '')
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
      this.licensePath = ''
    },
    // ---- 提交 ----
    async handleSubmit() {
      if (this.submitting) return
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
          license_url: self.licensePath || self.storageUrl(self.licenseUrl),
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
        // P0 修复：entId 取不到时不得静默跳过提审并提示成功，必须报失败可重试
        if (!entId) {
          uni.showToast({ title: '提交失败：未获取到企业ID，请重试', icon: 'none', duration: 2500 })
          return
        }
        await request({
          url: '/api/v1/enterprises/' + encodeURIComponent(entId) + '/submit',
          method: 'POST',
        })

        uni.showToast({ title: self.isEdit ? '已重新提交审核' : '已提交审核' })
        self.backTimer = setTimeout(function () {
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
@import '../../../pages/publish/pub-style.css';

.pub-fade { opacity: 0.6; }
.pub-photo-img {
  width: 100%;
  height: 100%;
  display: block;
}

/* 步骤指示器（配色对齐 pub：进行中蓝 / 完成绿 / 未到灰） */
.step-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 4px 12px 20px;
}

.step-item {
  display: flex;
  align-items: center;
  position: relative;
}

.step-circle {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: #EDF1F5;
  color: #98A2B3;
  font-size: 13px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: all 0.3s;
}

.step-item.active .step-circle {
  background: #0A66C2;
  color: #fff;
}

.step-item.done .step-circle {
  background: #219653;
  color: #fff;
}

.step-item.done .step-circle::after {
  content: '\2713';
}

.step-label {
  margin-left: 6px;
  font-size: 12px;
  color: #98A2B3;
  white-space: nowrap;
}

.step-item.active .step-label {
  color: #0A66C2;
  font-weight: 600;
}

.step-item.done .step-label {
  color: #219653;
}

.step-line {
  width: 32px;
  height: 2px;
  background: #DDE6F0;
  margin: 0 8px;
  transition: all 0.3s;
}

.step-line.filled {
  background: #219653;
}

/* 多选 chips（选中态浅蓝底蓝字、圆角 5px，对齐 pub / resume.vue skill-tag） */
.chips-card {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 13px;
}

.chip {
  padding: 7px 13px;
  border-radius: 5px;
  background: #F5F6F8;
  border: 1px solid #EEF1F4;
  font-size: 12px;
  color: #667085;
  transition: all 0.2s;
}

.chip--active {
  background: #E8F2FC;
  border-color: #A6C9EE;
  color: #0A66C2;
  font-weight: 700;
}

/* 卡片字段内嵌上传行：去掉 pub-upload-row 自带内边距（字段本身已有 padding） */
.pub-upload-inline {
  padding: 0;
}

/* 成立时间：picker 组件内的 pub-field 会命中 :first-child 去线规则，补回顶部分隔线 */
.pub-field.pub-field--pick {
  border-top: 1px solid #EEF1F4;
}

/* 上传预览删除角标（同 resume.vue） */
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
