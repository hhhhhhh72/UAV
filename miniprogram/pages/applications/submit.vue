<template>
  <view class="submit-page">
    <u-nav-bar
      title="项目申报"
      show-back
      @back="goBack"
    />

    <!-- Loading state (initial auth check) -->
    <view v-if="checkingAuth" class="state-view">
      <view class="loading-inline">
        <u-loading size="24rpx" />
        <text>验证登录信息...</text>
      </view>
    </view>

    <!-- Not logged in -->
    <view v-else-if="!isLoggedIn" class="state-view">
      <u-empty description="请先登录后申报项目" />
      <view class="retry-btn" @tap="goLogin">
        <text>去登录</text>
      </view>
    </view>

    <!-- Normal: form -->
    <view v-else class="form-body">
      <u-cell-group inset>
        <!-- Project name -->
        <u-field
          v-model="form.project_name"
          label="项目名称"
          placeholder="请输入项目名称"
        />

        <!-- Category picker -->
        <view class="field-row" @tap="showCategoryPicker = true">
          <u-field
            :model-value="form.category"
            label="申报类别"
            placeholder="请选择申报类别"
            disabled
          />
          <text class="field-arrow">›</text>
        </view>

        <!-- Budget -->
        <u-field
          v-model="form.budget"
          label="预算金额(元)"
          type="digit"
          placeholder="请输入预算金额"
        />

        <!-- Description -->
        <u-field
          v-model="form.description"
          label="项目描述"
          type="textarea"
          auto-height
          placeholder="请详细描述项目内容和预期成果"
        />

        <!-- Attachment (optional upload indication) -->
        <view class="field-row">
          <u-field
            label="附件"
            placeholder="选填（暂不支持上传）"
            disabled
          />
          <u-button size="small" type="primary" @click="uploadAttachment">
            上传
          </u-button>
        </view>
      </u-cell-group>

      <!-- Submit button -->
      <view class="submit-section">
        <u-button
          type="primary"
          block
          round
          :loading="submitting"
          @click="handleSubmit"
        >
          提交申报
        </u-button>
      </view>
    </view>

    <!-- Category picker popup -->
    <u-picker
      :show="showCategoryPicker"
      title="请选择申报类别"
      :columns="categoryOptions"
      @confirm="onCategoryConfirm"
      @update:show="showCategoryPicker = $event"
    />
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
    onCategoryConfirm(v) {
      this.form.category = v
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
  background: var(--color-bg);
  padding-bottom: 60px;
}

/* State views */
.state-view {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 120px;
}

.loading-inline {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: var(--color-text-secondary);
}

.retry-btn {
  margin-top: 12px;
  padding: 8px 24px;
  background: var(--color-primary);
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}

/* Form body */
.form-body {
  padding: 12px 0;
}

/* Picker / button row */
.field-row {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #fff;
  padding-right: 16px;
}

.field-arrow {
  color: #c8c9cc;
  font-size: 20px;
  flex-shrink: 0;
}

/* Submit section */
.submit-section {
  padding: 32px 16px;
}
</style>
