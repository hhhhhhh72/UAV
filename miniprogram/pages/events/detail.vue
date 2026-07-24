<template>
  <view class="page-container">
    <!-- Nav -->
    <van-nav-bar
      title="活动详情"
      left-arrow
      @click-left="goBack"
      custom-style="background: #f59e0b;"
    >
      <template #title>
        <text style="color: #fff;">活动详情</text>
      </template>
    </van-nav-bar>

    <!-- Loading state -->
    <view v-if="loading" class="loading-state">
      <van-loading size="24">加载中...</van-loading>
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg" class="error-state">
      <van-empty image="network" description="加载失败" />
      <view class="retry-btn" @tap="fetchDetail">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Normal state -->
    <template v-else-if="detail">
      <!-- Hero image -->
      <view class="hero-wrap">
        <van-image
          v-if="detail.cover_image || detail.image"
          :src="detail.cover_image || detail.image"
          width="100%"
          height="200"
          fit="cover"
          class="hero-img"
        />
      </view>

      <!-- Header info card -->
      <view class="header-card">
        <text class="event-title">{{ detail.title }}</text>
        <view class="tags-row">
          <van-tag :type="typeTagType(detail.type)" size="medium">
            {{ detail.type || '通用' }}
          </van-tag>
          <van-tag :type="statusTagType(detail.status)" size="medium">
            {{ statusLabel(detail.status) }}
          </van-tag>
        </view>
        <view class="info-rows">
          <view v-if="detail.date" class="info-row">
            <van-icon name="clock-o" size="14" color="#969799" />
            <text class="info-text">{{ detail.date }}</text>
          </view>
          <view v-if="detail.location" class="info-row">
            <van-icon name="location-o" size="14" color="#969799" />
            <text class="info-text">{{ detail.location }}</text>
          </view>
          <view class="info-row">
            <van-icon name="friends-o" size="14" color="#969799" />
            <text class="info-text">{{ detail.registration_count || 0 }} / {{ detail.capacity || '不限' }} 人</text>
          </view>
        </view>
      </view>

      <!-- Description -->
      <view v-if="detail.description" class="section">
        <van-cell-group inset>
          <van-cell title="活动介绍">
            <template #default>
              <view class="desc-content">{{ detail.description }}</view>
            </template>
          </van-cell>
        </van-cell-group>
      </view>

      <!-- Agenda -->
      <view v-if="detail.agenda && detail.agenda.length > 0" class="section">
        <van-cell-group inset>
          <view class="section-title">活动议程</view>
          <van-cell
            v-for="(item, idx) in detail.agenda"
            :key="idx"
            :title="item.time || ''"
            :label="item.topic || item.title || item"
          />
        </van-cell-group>
      </view>

      <!-- Speakers -->
      <view v-if="detail.speakers && detail.speakers.length > 0" class="section">
        <van-cell-group inset>
          <view class="section-title">嘉宾/讲者</view>
          <van-cell
            v-for="(speaker, idx) in detail.speakers"
            :key="idx"
            :title="speaker.name || speaker"
            :label="speaker.title || speaker.org || ''"
          />
        </van-cell-group>
      </view>

      <!-- Bottom action bar -->
      <view class="action-bar">
        <van-button
          v-if="canRegister"
          type="primary"
          block
          round
          :loading="registering"
          custom-style="background: #f59e0b; border-color: #f59e0b;"
          @tap="showRegPopup"
        >
          立即报名
        </van-button>
        <van-button
          v-else
          type="default"
          block
          round
          disabled
        >
          {{ registrationClosedReason }}
        </van-button>
      </view>
    </template>

    <!-- Registration popup -->
    <van-popup
      :show="regPopupVisible"
      position="bottom"
      round
      custom-style="padding: 24px 16px 40px;"
      @close="regPopupVisible = false"
    >
      <view class="popup-title">活动报名</view>
      <view class="popup-form">
        <van-field
          v-model="regForm.name"
          label="姓名"
          placeholder="请输入姓名"
          :border="true"
        />
        <van-field
          v-model="regForm.phone"
          label="手机号"
          type="number"
          placeholder="请输入手机号"
          :border="true"
        />
        <van-field
          v-model="regForm.company"
          label="单位"
          placeholder="请输入单位名称"
          :border="true"
        />
        <van-field
          v-model="regForm.dietary_notes"
          label="饮食备注"
          placeholder="如有特殊饮食需求请注明（选填）"
          type="textarea"
          :border="false"
        />
      </view>
      <view class="popup-actions">
        <van-button
          type="primary"
          block
          round
          :loading="submittingReg"
          custom-style="background: #f59e0b; border-color: #f59e0b;"
          @tap="submitRegistration"
        >
          确认报名
        </van-button>
        <view class="cancel-link" @tap="regPopupVisible = false">
          <text class="cancel-text">取消</text>
        </view>
      </view>
    </van-popup>
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
      detail: null,
      registering: false,

      // Registration popup
      regPopupVisible: false,
      submittingReg: false,
      regForm: {
        name: '',
        phone: '',
        company: '',
        dietary_notes: '',
      },
    }
  },
  computed: {
    canRegister() {
      if (!this.detail) return false
      var status = this.detail.status
      if (status === 'closed' || status === 'ended') return false
      if (this.detail.capacity && this.detail.registration_count >= this.detail.capacity) return false
      return true
    },
    registrationClosedReason() {
      if (!this.detail) return ''
      if (this.detail.status === 'closed' || this.detail.status === 'ended') return '已结束'
      if (this.detail.capacity && this.detail.registration_count >= this.detail.capacity) return '已满额'
      return '已报名'
    },
  },
  onLoad(options) {
    this.id = options.id || ''
    this.fetchDetail()
  },
  methods: {
    async fetchDetail() {
      this.loading = true
      this.errorMsg = ''
      try {
        var res = await request({ url: '/api/v1/events/' + encodeURIComponent(this.id) })
        this.detail = (res && res.data) || res || null
        if (this.detail && this.detail.title) {
          uni.setNavigationBarTitle({ title: this.detail.title })
        }
      } catch (e) {
        this.errorMsg = '网络异常，请稍后重试'
      } finally {
        this.loading = false
      }
    },

    showRegPopup() {
      this.regForm = { name: '', phone: '', company: '', dietary_notes: '' }
      this.regPopupVisible = true
    },

    async submitRegistration() {
      if (!this.regForm.name) {
        uni.showToast({ title: '请填写姓名', icon: 'none' })
        return
      }
      if (!this.regForm.phone) {
        uni.showToast({ title: '请填写手机号', icon: 'none' })
        return
      }

      this.submittingReg = true
      try {
        await request({
          url: '/api/v1/events/' + encodeURIComponent(this.id) + '/register',
          method: 'POST',
          data: this.regForm,
        })
        uni.showToast({ title: '报名成功', icon: 'success' })
        this.regPopupVisible = false
        // Refresh detail to update count
        this.fetchDetail()
      } catch (e) {
        uni.showToast({ title: '报名失败，请重试', icon: 'none' })
      } finally {
        this.submittingReg = false
      }
    },

    goBack() {
      uni.navigateBack()
    },

    typeTagType(type) {
      var map = {
        '论坛': 'primary',
        '走访': 'warning',
        '沙龙': 'success',
        '培训': 'danger',
        '赛事': 'primary',
      }
      return map[type] || 'default'
    },

    statusLabel(status) {
      var map = {
        'open': '报名中',
        'closed': '已结束',
        'ended': '已结束',
        'full': '已满额',
      }
      return map[status] || status || '未知'
    },

    statusTagType(status) {
      var map = {
        'open': 'success',
        'closed': 'default',
        'ended': 'default',
        'full': 'danger',
      }
      return map[status] || 'default'
    },
  },
}
</script>

<style scoped>
.page-container {
  min-height: 100vh;
  background: #f7f8fa;
  padding-bottom: 80px;
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
  background: #f59e0b;
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}

/* Hero */
.hero-wrap {
  width: 100%;
}

.hero-img {
  display: block;
}

/* Header card */
.header-card {
  background: #fff;
  margin: 12px;
  border-radius: 12px;
  padding: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.03);
}

.event-title {
  font-size: 18px;
  font-weight: 700;
  color: #323233;
  display: block;
  margin-bottom: 10px;
  line-height: 1.4;
}

.tags-row {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.info-rows {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.info-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.info-text {
  font-size: 13px;
  color: #646566;
}

/* Sections */
.section {
  margin-bottom: 12px;
}

.section-title {
  padding: 12px 16px 4px;
  font-size: 15px;
  font-weight: 600;
  color: #323233;
}

.desc-content {
  font-size: 14px;
  color: #333;
  line-height: 1.7;
  white-space: pre-wrap;
  padding: 8px 0;
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
  padding-bottom: 28px;
  z-index: 100;
}

/* Popup */
.popup-title {
  font-size: 18px;
  font-weight: 700;
  color: #323233;
  text-align: center;
  margin-bottom: 20px;
}

.popup-form {
  margin-bottom: 20px;
}

.popup-actions {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.cancel-link {
  text-align: center;
  padding: 8px 0;
}

.cancel-text {
  font-size: 14px;
  color: #969799;
}
</style>
