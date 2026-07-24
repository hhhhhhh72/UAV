<template>
  <view class="page-container">
    <!-- Nav -->
    <van-nav-bar
      title="成果转化"
      left-arrow
      @click-left="goBack"
      custom-style="background: #10b981;"
    >
      <template #title>
        <text style="color: #fff;">成果转化</text>
      </template>
    </van-nav-bar>

    <!-- Stage stepper (shown when tracking a specific item) -->
    <view v-if="activeTrack" class="stepper-card">
      <view class="stepper-title">{{ activeTrack.title || activeTrack.name }}</view>
      <view class="stepper">
        <view
          v-for="(stage, idx) in stages"
          :key="stage.key"
          class="stepper-step"
        >
          <!-- Dot -->
          <view class="stepper-dot-wrap">
            <view
              class="stepper-dot"
              :class="getDotClass(idx)"
            >
              <text v-if="idx < activeStageIdx" class="dot-icon">&#10003;</text>
              <text v-else class="dot-num">{{ idx + 1 }}</text>
            </view>
            <!-- Bar before this dot (not for first) -->
            <view
              v-if="idx > 0"
              class="stepper-bar"
              :class="idx <= activeStageIdx ? 'bar-done' : 'bar-pending'"
            />
          </view>
          <!-- Label -->
          <text
            class="stepper-label"
            :class="getLabelClass(idx)"
          >{{ stage.label }}</text>
          <!-- Date -->
          <text
            v-if="getStageDate(stage.key)"
            class="stepper-date"
          >{{ getStageDate(stage.key) }}</text>
        </view>
      </view>
    </view>

    <!-- Loading state -->
    <view v-if="loading && list.length === 0" class="loading-state">
      <van-loading size="24">加载中...</van-loading>
    </view>

    <!-- Empty state -->
    <view v-else-if="!loading && list.length === 0 && !errorMsg" class="empty-state-wrapper">
      <van-empty image="search" description="暂无转化成果" />
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg && list.length === 0" class="error-state">
      <van-empty image="network" description="加载失败" />
      <view class="retry-btn" @tap="fetchList">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Normal state: transformations list -->
    <view v-else class="list-body">
      <van-cell-group inset>
        <van-cell
          v-for="item in list"
          :key="item.id"
          :title="item.title || item.name"
          is-link
          @tap="viewTrack(item)"
        >
          <template #label>
            <view class="cell-meta">
              <van-tag :type="stageTagType(getCurrentStage(item))" size="small">
                {{ stageLabel(getCurrentStage(item)) }}
              </van-tag>
              <text v-if="item.org_name" class="meta-text">{{ item.org_name }}</text>
            </view>
          </template>
          <template #default>
            <text class="cell-date">{{ formatDate(item.updated_at || item.created_at) }}</text>
          </template>
        </van-cell>
      </van-cell-group>

      <view v-if="activeTrack" class="back-link" @tap="clearTrack">
        <text class="back-text">返回列表</text>
      </view>
    </view>
  </view>
</template>

<script>
import { request } from '../../utils/request'

export default {
  data() {
    return {
      loading: false,
      errorMsg: '',
      list: [],
      activeTrack: null,
      stages: [
        { key: 'lab', label: '实验室' },
        { key: 'pilot', label: '中试' },
        { key: 'industrial', label: '产业化' },
        { key: 'market', label: '上市' },
      ],
    }
  },
  computed: {
    activeStageIdx() {
      if (!this.activeTrack) return -1
      var stage = this.getCurrentStage(this.activeTrack)
      var idx = this.stages.findIndex(function (s) { return s.key === stage })
      return idx >= 0 ? idx : 0
    },
  },
  onLoad() {
    this.fetchList()
  },
  methods: {
    async fetchList() {
      this.loading = true
      this.errorMsg = ''
      try {
        var res = await request({ url: '/api/v1/transformations/mine' })
        var data = Array.isArray(res) ? res : (res && res.data) || res || {}
        this.list = Array.isArray(data) ? data : (data && data.items) || []
      } catch (e) {
        this.errorMsg = '网络异常，请稍后重试'
      } finally {
        this.loading = false
      }
    },

    viewTrack(item) {
      this.activeTrack = item
    },

    clearTrack() {
      this.activeTrack = null
    },

    getCurrentStage(item) {
      return item.current_stage || item.stage || 'lab'
    },

    getStageDate(stageKey) {
      if (!this.activeTrack) return ''
      var dates = this.activeTrack.stage_dates || this.activeTrack.dates || {}
      return dates[stageKey] || ''
    },

    getDotClass(idx) {
      if (idx < this.activeStageIdx) return 'dot-done'
      if (idx === this.activeStageIdx) return 'dot-active'
      return 'dot-pending'
    },

    getLabelClass(idx) {
      if (idx < this.activeStageIdx) return 'label-done'
      if (idx === this.activeStageIdx) return 'label-active'
      return 'label-pending'
    },

    stageLabel(stage) {
      var map = {
        'lab': '实验室',
        'pilot': '中试',
        'industrial': '产业化',
        'market': '上市',
      }
      return map[stage] || stage || '未知'
    },

    stageTagType(stage) {
      var map = {
        'lab': 'primary',
        'pilot': 'warning',
        'industrial': 'success',
        'market': 'danger',
      }
      return map[stage] || 'default'
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
.page-container {
  min-height: 100vh;
  background: #f7f8fa;
  padding-bottom: 40px;
}

/* Stepper card */
.stepper-card {
  background: #fff;
  margin: 12px 12px 0;
  border-radius: 12px;
  padding: 20px 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.03);
}

.stepper-title {
  font-size: 16px;
  font-weight: 600;
  color: #323233;
  margin-bottom: 24px;
  text-align: center;
}

.stepper {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  position: relative;
}

.stepper-step {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
  position: relative;
}

.stepper-dot-wrap {
  display: flex;
  align-items: center;
  width: 100%;
  position: relative;
}

.stepper-bar {
  position: absolute;
  top: 12px;
  left: -50%;
  right: 50%;
  height: 3px;
  z-index: 0;
}

.bar-done {
  background: #34c759;
}

.bar-pending {
  background: #ebedf0;
}

.stepper-dot {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto;
  position: relative;
  z-index: 1;
  font-size: 12px;
  font-weight: 600;
  transition: all 0.3s;
}

.dot-done {
  background: #34c759;
  color: #fff;
}

.dot-active {
  background: #1989fa;
  color: #fff;
  box-shadow: 0 0 0 4px rgba(25, 137, 250, 0.2);
}

.dot-pending {
  background: #ebedf0;
  color: #c8c9cc;
}

.dot-icon {
  font-size: 14px;
  line-height: 1;
}

.dot-num {
  font-size: 12px;
  line-height: 1;
}

.stepper-label {
  margin-top: 8px;
  font-size: 12px;
  text-align: center;
}

.label-done {
  color: #34c759;
  font-weight: 500;
}

.label-active {
  color: #1989fa;
  font-weight: 600;
}

.label-pending {
  color: #c8c9cc;
}

.stepper-date {
  margin-top: 4px;
  font-size: 10px;
  color: #969799;
  text-align: center;
}

/* Loading */
.loading-state {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}

.empty-state-wrapper {
  padding-top: 60px;
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
  background: #10b981;
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}

/* List */
.list-body {
  padding: 12px 0 24px;
}

.cell-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 4px;
}

.meta-text {
  font-size: 12px;
  color: #969799;
}

.cell-date {
  font-size: 12px;
  color: #c8c9cc;
}

.back-link {
  text-align: center;
  padding: 16px 0;
}

.back-text {
  font-size: 14px;
  color: #10b981;
}
</style>
