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

    <!-- Step 1: 基本信息 -->
    <view v-if="currentStep === 0" class="step-body">
      <u-cell-group inset>
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
        <u-field
          v-model="form.legal_person"
          label="法人代表"
          placeholder="请输入法人代表姓名"
        />
        <u-field
          v-model="form.contact_phone"
          label="联系电话"
          type="number"
          placeholder="请输入联系电话"
        />
        <u-field
          v-model="form.account_name"
          label="对公账户"
          placeholder="请输入对公账户名称（选填）"
        />
      </u-cell-group>
    </view>

    <!-- Step 2: 经营信息 -->
    <view v-if="currentStep === 1" class="step-body">
      <u-cell-group inset>
        <view class="field-row" @tap="showCategoryPicker">
          <u-field
            :model-value="form.industry_category ? categoryLabel(form.industry_category) : ''"
            label="行业类别"
            placeholder="请选择行业类别"
            disabled
          />
          <text class="field-arrow">›</text>
        </view>
        <u-field
          v-model="form.scale"
          label="企业规模"
          placeholder="如：50-100人"
        />
        <u-field
          v-model="form.address"
          label="企业地址"
          placeholder="请输入详细地址"
        />
      </u-cell-group>
    </view>

    <!-- Step 3: 附件与描述 -->
    <view v-if="currentStep === 2" class="step-body">
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
        <u-field
          v-model="form.description"
          label="企业描述"
          type="textarea"
          auto-height
          placeholder="请简要介绍企业经营范围与能力"
        />
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
        提交
      </u-button>
    </view>

    <!-- Category action sheet -->
    <u-popup
      :show="categoryPickerShow"
      position="bottom"
      round
      @close="closeCategoryPicker"
    >
      <view class="action-sheet">
        <view class="action-sheet-title">选择行业类别</view>
        <view
          v-for="opt in categoryOptions"
          :key="opt.value"
          class="action-sheet-item"
          @tap="onCategorySelect(opt.value)"
        >
          {{ opt.name }}
        </view>
        <view class="action-sheet-cancel" @tap="closeCategoryPicker">取消</view>
      </view>
    </u-popup>
  </view>
</template>

<script>
import { request, authStorage, BASE_URL } from '../../../utils/request'

export default {
  data() {
    return {
      currentStep: 0,
      steps: ['基本信息', '经营信息', '附件上传'],
      submitting: false,
      categoryPickerShow: false,
      form: {
        name: '',
        credit_code: '',
        legal_person: '',
        contact_phone: '',
        account_name: '',
        industry_category: '',
        scale: '',
        address: '',
        description: '',
      },
      licenseUrl: '',
      categoryOptions: [
        { name: '整机', value: '整机' },
        { name: '零部件', value: '零部件' },
        { name: '飞控', value: '飞控' },
        { name: '载荷', value: '载荷' },
        { name: '运营商', value: '运营商' },
        { name: '实训院校', value: '实训院校' },
      ],
    }
  },
  methods: {
    goBack() {
      uni.navigateBack()
    },
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
        if (!this.form.contact_phone) {
          return uni.showToast({ title: '请填写联系电话', icon: 'none' })
        }
      }
      this.currentStep++
    },
    showCategoryPicker() {
      this.categoryPickerShow = true
    },
    onCategorySelect(value) {
      this.form.industry_category = value
      this.categoryPickerShow = false
    },
    closeCategoryPicker() {
      this.categoryPickerShow = false
    },
    categoryLabel(value) {
      var found = this.categoryOptions.find(function (o) { return o.value === value })
      return found ? found.name : value
    },
    chooseLicense() {
      var self = this
      uni.chooseImage({
        count: 1,
        sizeType: ['compressed'],
        sourceType: ['album', 'camera'],
        success: function (res) {
          self.licenseUrl = res.tempFilePaths[0]
        },
        fail: function (err) {
          uni.showToast({ title: '选择图片失败', icon: 'none' })
        },
      })
    },
    previewLicense() {
      uni.previewImage({
        urls: [this.licenseUrl],
        current: this.licenseUrl,
      })
    },
    removeLicense() {
      this.licenseUrl = ''
    },
    async handleSubmit() {
      var self = this
      self.submitting = true

      try {
        var payload = {
          name: self.form.name,
          credit_code: self.form.credit_code,
          legal_person: self.form.legal_person,
          contact_phone: self.form.contact_phone,
          account_name: self.form.account_name,
          industry_category: self.form.industry_category,
          scale: self.form.scale,
          address: self.form.address,
          description: self.form.description,
        }

        // Upload license image if selected
        if (self.licenseUrl) {
          uni.showLoading({ title: '上传中...' })
          var uploadRes = await new Promise(function (resolve, reject) {
            uni.uploadFile({
              url: BASE_URL + '/api/v1/files/upload',
              filePath: self.licenseUrl,
              name: 'file',
              header: { Authorization: 'Bearer ' + authStorage.getAccessToken() },
              success: function (r) {
                var data = null
                try { data = JSON.parse(r.data) } catch (e) {}
                // 非 2xx 或缺少 file_id（含 data 信封内层）：透出后端具体错误（401/400/413 等）
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
          })
          uni.hideLoading()
          // /api/v1/files/upload 返回 {data:{file_id,...},request_id}（信封格式）
          var fid = uploadRes && (uploadRes.file_id || (uploadRes.data && uploadRes.data.file_id))
          if (!fid) {
            var reason = (uploadRes && uploadRes._error) || ''
            var tip = reason.indexOf('401') >= 0 || reason.indexOf('登录') >= 0 || reason.indexOf('token') >= 0
              ? '登录已过期，请重新登录后重试'
              : ('营业执照上传失败：' + (reason || '请重试'))
            uni.showToast({ title: tip, icon: 'none', duration: 2500 })
            return
          }
          payload.license_url = '/uploads/' + fid
        }

        const ent = await request({
          url: '/api/v1/enterprises',
          method: 'POST',
          data: payload,
        })

        // 创建为草稿 → 立即提交审核（submit 后进入管理员审核队列）
        const entId = ent && (ent.data && ent.data.id ? ent.data.id : ent.id)
        if (entId) {
          await request({
            url: '/api/v1/enterprises/' + encodeURIComponent(entId) + '/submit',
            method: 'POST',
          })
        }

        uni.showToast({ title: '已提交审核' })
        setTimeout(function () {
          uni.navigateBack()
        }, 1200)
      } catch (e) {
        var msg = (e && e.data && e.data.message) || (e && e.message) || '提交失败，请重试'
        uni.showToast({ title: msg, icon: 'none' })
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

.upload-hint {
  font-size: 12px;
  color: var(--color-text-secondary);
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

/* Action sheet */
.action-sheet {
  padding: 24rpx 0 env(safe-area-inset-bottom);
}

.action-sheet-title {
  text-align: center;
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text);
  padding: 8px 0 16px;
}

.action-sheet-item {
  text-align: center;
  padding: 16px;
  font-size: 15px;
  color: var(--color-text);
  border-top: 1rpx solid var(--color-divider);
}

.action-sheet-item:active {
  background: var(--color-bg);
  color: var(--color-primary);
}

.action-sheet-cancel {
  text-align: center;
  padding: 16px;
  margin-top: 8px;
  font-size: 15px;
  color: var(--color-text-secondary);
  border-top: 1rpx solid var(--color-divider);
}
</style>
