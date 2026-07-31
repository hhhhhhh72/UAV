<template>
  <view class="register-page">
    <van-nav-bar
      title="企业入驻"
      fixed
      placeholder
      left-arrow
      @click-left="goBack"
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
      <van-cell-group inset>
        <van-field
          v-model="form.name"
          label="企业名称"
          placeholder="请输入企业名称"
          required
          :border="true"
        />
        <van-field
          v-model="form.credit_code"
          label="信用代码"
          placeholder="统一社会信用代码"
          required
          :border="true"
        />
        <van-field
          v-model="form.legal_person"
          label="法人代表"
          placeholder="请输入法人代表姓名"
          :border="true"
        />
        <van-field
          v-model="form.contact_phone"
          label="联系电话"
          type="number"
          placeholder="请输入联系电话"
          required
          :border="true"
        />
      </van-cell-group>
    </view>

    <!-- Step 2: 经营信息 -->
    <view v-if="currentStep === 1" class="step-body">
      <van-cell-group inset>
        <van-field
          :model-value="form.industry_category ? categoryLabel(form.industry_category) : ''"
          label="行业类别"
          placeholder="请选择行业类别"
          readonly
          is-link
          :border="true"
          @click="showCategoryPicker"
        />
        <van-field
          v-model="form.scale"
          label="企业规模"
          placeholder="如：50-100人"
          :border="true"
        />
        <van-field
          v-model="form.address"
          label="企业地址"
          placeholder="请输入详细地址"
          :border="true"
        />
      </van-cell-group>
    </view>

    <!-- Step 3: 附件与描述 -->
    <view v-if="currentStep === 2" class="step-body">
      <van-cell-group inset>
        <view class="upload-cell">
          <text class="upload-label">营业执照</text>
          <view class="upload-area">
            <view v-if="licenseUrl" class="upload-preview" @tap="previewLicense">
              <image :src="licenseUrl" mode="aspectFill" class="upload-img" />
              <view class="upload-remove" @tap.stop="removeLicense">
                <van-icon name="clear" size="18" color="#ee0a24" />
              </view>
            </view>
            <view v-else class="upload-btn" @tap="chooseLicense">
              <van-icon name="photograph" size="28" color="#969799" />
              <text class="upload-hint">点击上传</text>
            </view>
          </view>
        </view>
        <van-field
          v-model="form.description"
          label="企业描述"
          type="textarea"
          placeholder="请简要介绍企业经营范围与能力"
          rows="4"
          autosize
          :border="true"
        />
      </van-cell-group>
    </view>

    <!-- Bottom action buttons -->
    <view class="action-bar">
      <van-button
        v-if="currentStep > 0"
        plain
        round
        class="prev-btn"
        @tap="prevStep"
      >
        上一步
      </van-button>
      <van-button
        v-if="currentStep < 2"
        type="primary"
        round
        class="next-btn"
        @tap="nextStep"
      >
        下一步
      </van-button>
      <van-button
        v-else
        type="primary"
        round
        class="submit-btn"
        :loading="submitting"
        @tap="handleSubmit"
      >
        提交
      </van-button>
    </view>

    <!-- Category picker -->
    <van-action-sheet
      :show="categoryPickerShow"
      :actions="categoryOptions"
      cancel-text="取消"
      @select="onCategorySelect"
      @close="categoryPickerShow = false"
      @cancel="categoryPickerShow = false"
    />
  </view>
</template>

<script>
import { request } from '../../utils/request'

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
    onCategorySelect(e) {
      this.form.industry_category = e.detail.value
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
              url: 'http://localhost:8080/api/v1/upload',
              filePath: self.licenseUrl,
              name: 'file',
              success: function (r) {
                try {
                  var data = JSON.parse(r.data)
                  resolve(data)
                } catch (e) {
                  resolve({ url: self.licenseUrl })
                }
              },
              fail: reject,
            })
          })
          uni.hideLoading()
          payload.license_url = (uploadRes && (uploadRes.url || uploadRes.data && uploadRes.data.url)) || self.licenseUrl
        }

        await request({
          url: '/api/v1/enterprises',
          method: 'POST',
          data: payload,
        })

        uni.showToast({ title: '提交成功' })
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
  background: #f7f8fa;
  padding-bottom: 100px;
}

/* Step indicator */
.step-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px 32px 16px;
  background: #fff;
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
  background: #ebedf0;
  color: #969799;
  font-size: 14px;
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
  background: #07c160;
  color: #fff;
}

.step-item.done .step-circle::after {
  content: '\2713';
}

.step-label {
  margin-left: 6px;
  font-size: 12px;
  color: #969799;
  white-space: nowrap;
}

.step-item.active .step-label {
  color: #0A66C2;
  font-weight: 600;
}

.step-item.done .step-label {
  color: #07c160;
}

.step-line {
  width: 40px;
  height: 2px;
  background: #ebedf0;
  margin: 0 10px;
  transition: all 0.3s;
}

.step-line.filled {
  background: #07c160;
}

/* Step body */
.step-body {
  padding: 12px 0;
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
  color: #323233;
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
  border: 1px dashed #c8c9cc;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.upload-hint {
  font-size: 12px;
  color: #969799;
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
  background: #fff;
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
