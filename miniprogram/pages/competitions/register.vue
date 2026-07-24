<template>
  <view class="register-page">
    <!-- Nav -->
    <van-nav-bar
      title="赛事报名"
      left-arrow
      fixed
      placeholder
      custom-style="background: #7c3aed;"
      title-style="color: #fff;"
      left-icon-style="color: #fff;"
      @click-left="goBack"
    />

    <!-- Loading state -->
    <view v-if="loading" class="loading-state">
      <van-loading size="24">加载中...</van-loading>
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg" class="error-state">
      <van-empty description="加载失败" image="error" />
      <view class="retry-btn" @tap="fetchCompetition">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Empty state -->
    <view v-else-if="!competition" class="empty-state-wrapper">
      <van-empty image="search" description="赛事信息不存在" />
    </view>

    <!-- Normal state -->
    <template v-else>
      <!-- Competition summary card -->
      <view class="summary-card">
        <view class="summary-header">
          <text class="comp-name">{{ competition.name || competition.title }}</text>
          <van-tag :type="statusTagType(competition.status)" size="medium">
            {{ statusLabel(competition.status) }}
          </van-tag>
        </view>
        <view class="summary-meta">
          <view v-if="competition.date" class="meta-item">
            <van-icon name="calendar-o" size="14" color="#969799" />
            <text class="meta-text">{{ competition.date }}</text>
          </view>
          <view v-if="competition.location" class="meta-item">
            <van-icon name="location-o" size="14" color="#969799" />
            <text class="meta-text">{{ competition.location }}</text>
          </view>
        </view>
      </view>

      <!-- Registration form -->
      <view class="section-title-wrapper">
        <text class="section-title">报名信息</text>
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
          v-model="form.id_number"
          label="身份证号"
          placeholder="请输入身份证号"
          required
          :border="true"
        />
        <van-field
          :model-value="categoryLabel(form.category)"
          label="参赛类别"
          placeholder="请选择参赛类别"
          readonly
          is-link
          :border="true"
          @click="showCategoryPicker"
        />
        <van-field
          v-model="form.emergency_contact"
          label="紧急联系人"
          placeholder="姓名及电话"
          :border="true"
        />
        <van-field
          v-model="form.notes"
          label="备注"
          type="textarea"
          placeholder="其他需要说明的信息"
          rows="3"
          autosize
          :border="true"
        />
      </van-cell-group>

      <!-- Submit button -->
      <view class="submit-area">
        <van-button
          type="primary"
          block
          round
          :loading="submitting"
          custom-style="background: #7c3aed; border-color: #7c3aed;"
          @tap="handleSubmit"
        >
          提交报名
        </van-button>
      </view>
    </template>

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
      id: '',
      loading: false,
      errorMsg: '',
      competition: null,
      submitting: false,
      categoryPickerShow: false,
      form: {
        name: '',
        phone: '',
        id_number: '',
        category: '',
        emergency_contact: '',
        notes: '',
      },
      categoryOptions: [
        { name: '运动员', value: '运动员' },
        { name: '教练', value: '教练' },
        { name: '裁判', value: '裁判' },
        { name: '俱乐部', value: '俱乐部' },
      ],
    }
  },
  onLoad(options) {
    this.id = options.id || ''
    this.fetchCompetition()
  },
  methods: {
    async fetchCompetition() {
      this.loading = true
      this.errorMsg = ''

      try {
        var res = await request({ url: '/api/v1/competitions/' + encodeURIComponent(this.id) })
        this.competition = (res && res.data) || res || null
      } catch (e) {
        this.errorMsg = '网络异常，请稍后重试'
      } finally {
        this.loading = false
      }
    },
    showCategoryPicker() {
      this.categoryPickerShow = true
    },
    onCategorySelect(e) {
      this.form.category = e.detail.value
      this.categoryPickerShow = false
    },
    categoryLabel(value) {
      var found = this.categoryOptions.find(function (o) { return o.value === value })
      return found ? found.name : value
    },
    statusTagType(status) {
      var map = { open: 'success', closed: 'default', full: 'danger' }
      return map[status] || 'default'
    },
    statusLabel(status) {
      var map = { open: '报名中', closed: '已结束', full: '已满额' }
      return map[status] || status || '未知'
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
      if (!/^1\d{10}$/.test(this.form.phone.trim())) {
        uni.showToast({ title: '请输入11位手机号', icon: 'none' })
        return false
      }
      if (!this.form.id_number.trim()) {
        uni.showToast({ title: '请填写身份证号', icon: 'none' })
        return false
      }
      return true
    },
    async handleSubmit() {
      if (!this.validateForm()) return

      this.submitting = true
      try {
        await request({
          url: '/api/v1/competitions/' + encodeURIComponent(this.id) + '/register',
          method: 'POST',
          data: {
            name: this.form.name,
            phone: this.form.phone,
            id_number: this.form.id_number,
            category: this.form.category,
            emergency_contact: this.form.emergency_contact,
            notes: this.form.notes,
          },
        })
        uni.showToast({ title: '报名成功', icon: 'success' })
        setTimeout(function () {
          uni.navigateBack()
        }, 1200)
      } catch (e) {
        var msg = (e && e.data && e.data.message) || '报名失败，请重试'
        uni.showToast({ title: msg, icon: 'none' })
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
.register-page {
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
  background: #7c3aed;
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}

.empty-state-wrapper {
  padding-top: 60px;
}

/* Summary card */
.summary-card {
  margin: 12px 12px;
  padding: 16px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.02);
}

.summary-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.comp-name {
  font-size: 17px;
  font-weight: 700;
  color: #323233;
  flex: 1;
  line-height: 1.4;
}

.summary-meta {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.meta-text {
  font-size: 13px;
  color: #969799;
}

/* Section title */
.section-title-wrapper {
  padding: 16px 16px 8px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: #323233;
}

/* Submit */
.submit-area {
  padding: 24px 16px;
}
</style>
