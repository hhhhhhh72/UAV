<template>
  <view class="depts-page">
    <u-nav-bar
      title="部门对接"
      show-back
      @back="goBack"
    />

    <!-- Loading -->
    <view v-if="loading" class="loading-state">
      <view class="loading-inline">
        <u-loading size="24rpx" color="#667085" />
        <text>加载中...</text>
      </view>
    </view>

    <!-- Error -->
    <view v-else-if="errorMsg && depts.length === 0 && drills.length === 0" class="state-view">
      <u-empty description="加载失败" />
      <view class="retry-btn" @tap="fetchData">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Empty -->
    <view v-else-if="!loading && depts.length === 0 && drills.length === 0 && !errorMsg" class="state-view">
      <u-empty description="暂无数据" />
    </view>

    <!-- Normal -->
    <template v-else>
      <!-- 对接部门 -->
      <view v-if="depts.length > 0" class="section">
        <view class="section-header">
          <text class="section-title">对接部门</text>
        </view>
        <view class="dp-list">
          <view v-for="item in depts" :key="item.id" class="dp-card">
            <view class="dp-header">
              <view class="dp-icon" :style="deptIconStyle(item.type || item.name)"><text>{{ deptIcon(item.type || item.name) }}</text></view>
              <view class="dp-info">
                <view class="dp-name-row">
                  <text class="dp-name">{{ item.name }}</text>
                  <u-tag
                    :type="item.agreement_status === '已签署' ? 'success' : 'danger'"
                    size="mini"
                    :round="false"
                    plain
                  >{{ item.agreement_status === '已签署' ? '已签署' : '未签署' }}</u-tag>
                </view>
                <view class="dp-meta">
                  <text v-if="item.contact_name" class="dp-meta-item">{{ item.contact_name }}</text>
                  <text v-if="item.contact_phone" class="dp-meta-item">{{ item.contact_phone }}</text>
                </view>
              </view>
            </view>
          </view>
        </view>
      </view>

      <!-- 演练记录 -->
      <view v-if="drills.length > 0" class="section">
        <view class="section-header">
          <text class="section-title">演练记录</text>
        </view>
        <view class="tl-card">
          <view
            v-for="(item, idx) in drills"
            :key="item.id || idx"
            class="tl-item"
          >
            <view class="tl-line">
              <view class="tl-dot" :class="{ active: idx === 0 }" />
              <view v-if="idx < drills.length - 1" class="tl-bar" />
            </view>
            <view class="tl-content">
              <text class="tl-date">{{ formatDate(item.date || item.created_at) }}</text>
              <text class="tl-event">{{ item.event_name || item.title }}</text>
              <text v-if="item.description" class="tl-desc">{{ item.description }}</text>
            </view>
          </view>
        </view>
      </view>
      <view class="bottom-space" />
    </template>
  </view>
</template>

<script>
import { request } from '../../utils/request'

export default {
  data() {
    return {
      loading: false,
      errorMsg: '',
      depts: [],
      drills: [],
    }
  },
  onLoad() {
    this.fetchData()
  },
  methods: {
    async fetchData() {
      this.loading = true
      this.errorMsg = ''

      try {
        var results = await Promise.all([
          request({ url: '/api/v1/emergency-depts' }),
          request({ url: '/api/v1/emergency-drills' }),
        ])

        var deptRes = results[0]
        var drillRes = results[1]

        var deptData = Array.isArray(deptRes) ? deptRes : (deptRes && deptRes.data) || deptRes || {}
        var drillData = Array.isArray(drillRes) ? drillRes : (drillRes && drillRes.data) || drillRes || {}

        var deptItems = Array.isArray(deptData) ? deptData : (deptData && deptData.items) || (deptData && deptData.list) || []
        var drillItems = Array.isArray(drillData) ? drillData : (drillData && drillData.items) || (drillData && drillData.list) || []

        this.depts = deptItems
        this.drills = drillItems
      } catch (e) {
        this.errorMsg = '网络异常，请稍后重试'
      } finally {
        this.loading = false
      }
    },
    /* 部门类型字符图标（低饱和色块，非 emoji） */
    deptIcon(name) {
      var nameStr = (name || '').toLowerCase()
      if (nameStr.indexOf('消防') !== -1) return '防'
      if (nameStr.indexOf('公安') !== -1) return '警'
      if (nameStr.indexOf('医疗') !== -1 || nameStr.indexOf('卫生') !== -1) return '医'
      if (nameStr.indexOf('应急') !== -1) return '应'
      return '部'
    },
    deptIconStyle(name) {
      var nameStr = (name || '').toLowerCase()
      if (nameStr.indexOf('消防') !== -1) return { background: '#FFF0E6', color: '#E96012' }
      if (nameStr.indexOf('公安') !== -1) return { background: '#EAF3FB', color: '#0A66C2' }
      if (nameStr.indexOf('医疗') !== -1 || nameStr.indexOf('卫生') !== -1) return { background: '#E9F7F0', color: '#168A55' }
      if (nameStr.indexOf('应急') !== -1) return { background: '#EAF3FB', color: '#0A66C2' }
      return { background: '#F4F6F8', color: '#667085' }
    },
    formatDate(iso) {
      if (!iso) return ''
      var d = new Date(iso)
      var m = d.getMonth() + 1
      var day = d.getDate()
      return d.getFullYear() + '-' + (m < 10 ? '0' : '') + m + '-' + (day < 10 ? '0' : '') + day
    },
    goBack() {
      uni.navigateBack()
    },
  },
}
</script>

<style scoped>
.depts-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(env(safe-area-inset-bottom) + 24rpx);
}

/* State views */
.loading-state {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}

.loading-inline {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #667085;
}

.state-view {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 80px;
}

.retry-btn {
  margin-top: 12px;
  padding: 8px 24px;
  background: #0A66C2;
  color: #fff;
  border-radius: 8px;
  font-size: 14px;
}

/* Section */
.section {
  margin-bottom: 16px;
}

.section-header {
  padding: 20rpx 24rpx 12rpx;
}

.section-title {
  font-size: 28rpx;
  font-weight: 600;
  color: #17212B;
}

/* 部门卡片 */
.dp-list { padding: 0 24rpx; }

.dp-card {
  background: #ffffff;
  border: 1px solid #EEF1F4;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}

.dp-header {
  display: flex;
  align-items: center;
  gap: 16rpx;
}

.dp-icon {
  width: 64rpx;
  height: 64rpx;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  font-weight: 600;
  flex-shrink: 0;
}

.dp-info {
  flex: 1;
  min-width: 0;
}

.dp-name-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
  flex-wrap: wrap;
}

.dp-name {
  font-size: 28rpx;
  font-weight: 600;
  color: #17212B;
}

.dp-meta {
  display: flex;
  align-items: center;
  gap: 24rpx;
  margin-top: 6rpx;
}

.dp-meta-item {
  font-size: 23rpx;
  color: #667085;
}

/* 演练记录时间线 */
.tl-card {
  margin: 0 24rpx;
  padding: 24rpx 24rpx 4rpx;
  background: #ffffff;
  border: 1px solid #EEF1F4;
  border-radius: 16rpx;
}

.tl-item {
  display: flex;
  align-items: flex-start;
  padding-bottom: 20rpx;
}

.tl-line {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-right: 20rpx;
  flex-shrink: 0;
}

.tl-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: #D0D5DD;
  flex-shrink: 0;
  margin-top: 6rpx;
}

.tl-dot.active {
  background: #0A66C2;
}

.tl-bar {
  width: 2px;
  flex: 1;
  background: #EEF1F4;
  margin-top: 6rpx;
  min-height: 100%;
}

.tl-content {
  flex: 1;
  padding-bottom: 4rpx;
}

.tl-date {
  font-size: 22rpx;
  color: #98A2B3;
  display: block;
  margin-bottom: 4rpx;
}

.tl-event {
  font-size: 28rpx;
  font-weight: 600;
  color: #17212B;
  display: block;
  line-height: 1.4;
}

.tl-desc {
  font-size: 24rpx;
  color: #667085;
  display: block;
  margin-top: 4rpx;
  line-height: 1.5;
}

.bottom-space { height: 24rpx; }
</style>
