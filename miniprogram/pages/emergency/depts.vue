<template>
  <view class="depts-page">
    <!-- Nav -->
    <u-nav-bar
      title="部门对接"
      show-back
      @back="goBack"
    />

    <!-- Loading state -->
    <view v-if="loading" class="loading-state">
      <view class="loading-inline">
        <u-loading size="24rpx" />
        <text>加载中...</text>
      </view>
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg && depts.length === 0 && drills.length === 0" class="error-state">
      <u-empty description="加载失败" />
      <view class="retry-btn" @tap="fetchData">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Empty state -->
    <view v-else-if="!loading && depts.length === 0 && drills.length === 0 && !errorMsg" class="empty-state-wrapper">
      <u-empty description="暂无数据" />
    </view>

    <!-- Normal state -->
    <template v-else>
      <!-- Department list -->
      <view v-if="depts.length > 0" class="section">
        <view class="section-header">
          <text class="section-title">对接部门</text>
        </view>
        <u-cell-group inset>
          <u-cell
            v-for="item in depts"
            :key="item.id"
          >
            <template #title>
              <view class="dept-content">
                <view class="dept-title-row">
                  <text class="dept-emoji">{{ deptEmoji(item.type || item.name) }}</text>
                  <text class="dept-name">{{ item.name }}</text>
                  <u-tag
                    :type="item.agreement_status === '已签署' ? 'success' : 'danger'"
                    size="mini"
                  >
                    {{ item.agreement_status === '已签署' ? '已签署' : '未签署' }}
                  </u-tag>
                </view>
                <view class="dept-meta">
                  <text v-if="item.contact_name" class="meta-item">{{ item.contact_name }}</text>
                  <text v-if="item.contact_phone" class="meta-item">{{ item.contact_phone }}</text>
                </view>
              </view>
            </template>
          </u-cell>
        </u-cell-group>
      </view>

      <!-- Drill records timeline -->
      <view v-if="drills.length > 0" class="section">
        <view class="section-header">
          <text class="section-title">演练记录</text>
        </view>
        <view class="timeline-card">
          <view class="timeline">
            <view
              v-for="(item, idx) in drills"
              :key="item.id || idx"
              class="timeline-item"
              :class="{ last: idx === drills.length - 1 }"
            >
              <view class="timeline-line">
                <view class="timeline-dot" :class="{ active: idx === 0 }" />
                <view v-if="idx < drills.length - 1" class="timeline-bar" />
              </view>
              <view class="timeline-content">
                <view class="timeline-header">
                  <text class="timeline-date">{{ formatDate(item.date || item.created_at) }}</text>
                </view>
                <text class="timeline-event">{{ item.event_name || item.title }}</text>
                <text v-if="item.description" class="timeline-desc">{{ item.description }}</text>
              </view>
            </view>
          </view>
        </view>
      </view>
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
    deptEmoji(name) {
      var nameStr = (name || '').toLowerCase()
      if (nameStr.indexOf('消防') !== -1) return '防'
      if (nameStr.indexOf('公安') !== -1) return '警'
      if (nameStr.indexOf('医疗') !== -1 || nameStr.indexOf('卫生') !== -1) return '医'
      if (nameStr.indexOf('应急') !== -1) return '应'
      return '部'
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
  background: var(--color-bg);
  padding-bottom: calc(env(safe-area-inset-bottom) + 40px);
}

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
  color: var(--color-text-secondary);
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
  background: var(--color-primary);
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}

.empty-state-wrapper {
  padding-top: 60px;
}

/* Section */
.section {
  margin-bottom: 16px;
}

.section-header {
  padding: 16px 16px 10px;
}

.section-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--color-text);
}

/* Department */
.dept-content {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
}

.dept-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.dept-emoji {
  width: 44rpx;
  height: 44rpx;
  line-height: 44rpx;
  text-align: center;
  font-size: 24rpx;
  color: var(--color-primary);
  background: var(--color-primary-light);
  border-radius: 8rpx;
  flex-shrink: 0;
}

.dept-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text);
}

.dept-meta {
  display: flex;
  align-items: center;
  gap: 12px;
}

.meta-item {
  font-size: 13px;
  color: var(--color-text-secondary);
}

/* Timeline */
.timeline-card {
  margin: 0 12px;
  padding: 16px 16px 4px;
  background: var(--color-bg-card);
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.02);
}

.timeline {
  display: flex;
  flex-direction: column;
}

.timeline-item {
  display: flex;
  align-items: flex-start;
  padding-bottom: 20px;
}

.timeline-item.last {
  padding-bottom: 4px;
}

.timeline-line {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-right: 14px;
  flex-shrink: 0;
}

.timeline-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: var(--color-text-placeholder);
  border: 2px solid #fff;
  box-shadow: 0 0 0 2px var(--color-text-placeholder);
  flex-shrink: 0;
  margin-top: 2px;
}

.timeline-dot.active {
  background: var(--color-primary);
  box-shadow: 0 0 0 2px var(--color-primary);
}

.timeline-bar {
  width: 2px;
  flex: 1;
  background: var(--color-divider);
  margin-top: 6px;
  min-height: 100%;
}

.timeline-content {
  flex: 1;
  padding-bottom: 4px;
}

.timeline-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.timeline-date {
  font-size: 12px;
  color: var(--color-text-secondary);
}

.timeline-event {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text);
  display: block;
  line-height: 1.4;
}

.timeline-desc {
  font-size: 13px;
  color: var(--color-text);
  opacity: 0.85;
  display: block;
  margin-top: 4px;
  line-height: 1.5;
}
</style>
