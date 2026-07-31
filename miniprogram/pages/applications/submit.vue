<template>
  <view class="submit-page">
    <van-nav-bar
      title="项目申报"
      fixed
      placeholder
      left-arrow
      @click-left="goBack"
    />

    <!-- Loading state (initial auth check) -->
    <view v-if="checkingAuth" class="state-view">
      <van-loading size="24">验证登录信息...</van-loading>
    </view>

    <!-- Not logged in -->
    <view v-else-if="!isLoggedIn" class="state-view">
      <van-empty description="请先登录后申报项目" image="error" />
      <view class="retry-btn" @tap="goLogin">
        <text>去登录</text>
      </view>
    </view>

    <!-- Normal: form -->
    <view v-else class="form-body">
      <van-cell-group inset>
        <!-- Project name -->
        <van-field
          v-model="form.project_name"
          label="项目名称"
          placeholder="请输入项目名称"
          required
          :border="true"
        />

        <!-- Category picker -->
        <view class="picker-field" @tap="showCategoryPicker = true">
          <view class="picker-field-inner">
            <text class="picker-label">
              <text class="required">*</text>
              申报类别
            </text>
            <view class="picker-value-wrapper">
              <text :class="{ placeholder: !form.category }">
                {{ form.category || '请选择申报类别' }}
              </text>
              <van-icon name="arrow" size="12" color="#969799" />
            </view>
          </view>
        </view>

        <!-- Budget -->
        <van-field
          v-model="form.budget"
          label="预算金额(元)"
          type="digit"
          placeholder="请输入预算金额"
          :border="true"
        />

        <!-- Description -->
        <van-field
          v-model="form.description"
          label="项目描述"
          type="textarea"
          autosize
          placeholder="请详细描述项目内容和预期成果"
          :border="true"
          rows="4"
        />

        <!-- Attachment (optional upload indication) -->
        <van-field
          label="附件"
          placeholder="选填（暂不支持上传）"
          readonly
          :border="true"
          right-icon="plus"
          @click-right-icon="uploadAttachment"
        >
          <template #right-icon>
            <van-button size="small" hairline type="primary" @tap="uploadAttachment">
              上传
            </van-button>
          </template>
        </van-field>
      </van-cell-group>

      <!-- Submit button -->
      <view class="submit-section">
        <van-button
          type="primary"
          block
          round
          :loading="submitting"
          @tap="handleSubmit"
        >
          提交申报
        </van-button>
      </view>
    </view>

    <!-- Category picker popup -->
    <van-popup
      :show="showCategoryPicker"
      position="bottom"
      round
      @close="showCategoryPicker = false"
    >
      <van-picker
        :columns="categoryOptions"
        @confirm="onCategoryConfirm"
        @cancel="showCategoryPicker = false"
      />
    </van-popup>
  </view>
</template>

<script>
import { request, getStoredUser, authStorage } from '../../utils/request'

export default {
  data() {
    return {
      checkingAuth: true,
      submitting: false,
      showCategoryPicker: false,
      form: {
        project_name: '',
        category: '',
        budget: '',
        description: '',
      },
      categoryOptions: ['资质申请', '政府补贴', '专项资金'],
    }
  },
  computed: {
    isLoggedIn() {
      var token = authStorage.getAccessToken()
      var user = getStoredUser()
      return !!(token && user)
    },
  },
  onLoad() {
    this.checkAuth()
  },
  onShow() {
    // Re-check auth when returning from login
    if (!this.checkingAuth) {
      this.checkAuth()
    }
  },
  methods: {
    checkAuth() {
      this.checkingAuth = true
      // Small delay to ensure storage is consistent
      var self = this
      setTimeout(function () {
        self.checkingAuth = false
        if (self.isLoggedIn) {
          // Pre-fill user info
          var user = getStoredUser()
          if (!self.form.project_name && user) {
            // Keep form empty - user fills themselves
          }
        }
      }, 300)
    },
    goLogin() {
      uni.navigateTo({ url: '/pages/login/index' })
    },
    onCategoryConfirm(e) {
      this.form.category = e.detail.value
      this.showCategoryPicker = false
    },
    uploadAttachment() {
      uni.showToast({ title: '附件上传功能即将上线', icon: 'none' })
    },
    async handleSubmit() {
      // Validation
      if (!this.form.project_name.trim()) {
        uni.showToast({ title: '请填写项目名称', icon: 'none' })
        return
      }
      if (!this.form.category) {
        uni.showToast({ title: '请选择申报类别', icon: 'none' })
        return
      }
      if (!this.form.description.trim()) {
        uni.showToast({ title: '请填写项目描述', icon: 'none' })
        return
      }

      this.submitting = true
      try {
        var payload = {
          project_name: this.form.project_name.trim(),
          category: this.form.category,
          budget: parseFloat(this.form.budget) || 0,
          description: this.form.description.trim(),
        }

        await request({
          url: '/api/v1/project-applications',
          method: 'POST',
          data: payload,
        })

        uni.showToast({ title: '申报成功', icon: 'success' })
        // Navigate back after success
        setTimeout(function () {
          uni.navigateBack()
        }, 1500)
      } catch (e) {
        uni.showToast({ title: '申报失败，请稍后重试', icon: 'none' })
      } finally {
        this.submitting = false
      }
    },
    goBack() {
      uni.navigateBack()
    },
  },
}
</script>

<style scoped>
.submit-page {
  min-height: 100vh;
  background: #f7f8fa;
  padding-bottom: 60px;
}

/* State views */
.state-view {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 120px;
}

.retry-btn {
  margin-top: 12px;
  padding: 8px 24px;
  background: #0A66C2;
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}

/* Form body */
.form-body {
  padding: 12px 0;
}

/* Picker field */
.picker-field {
  background: #fff;
  padding: 0 16px;
}

.picker-field-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid #f7f8fa;
}

.picker-label {
  font-size: 14px;
  color: #323233;
  flex-shrink: 0;
  margin-right: 12px;
}

.required {
  color: #ee0a24;
  margin-right: 4px;
}

.picker-value-wrapper {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  color: #323233;
}

.picker-value-wrapper .placeholder {
  color: #c8c9cc;
}

/* Submit section */
.submit-section {
  padding: 32px 16px;
}
</style>
