<template>
  <view class="dispatches-page">
    <u-nav-bar
      title="调度记录"
      show-back
      @back="goBack"
    />

    <u-sticky>
      <u-tabs
        v-model:active="activeTabIndex"
        :titles="['全部', '待响应', '已调度', '已完成']"
        @change="onTabChange"
      />
    </u-sticky>

    <!-- Loading -->
    <view v-if="loading && list.length === 0" class="loading-state">
      <view class="loading-inline">
        <u-loading size="24rpx" />
        <text>加载中...</text>
      </view>
    </view>

    <!-- Error -->
    <view v-else-if="errorMsg && list.length === 0" class="state-view">
      <u-empty description="加载失败" />
      <view class="retry-btn" @tap="fetchList(true)">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Empty -->
    <view v-else-if="!loading && list.length === 0" class="state-view">
      <u-empty description="暂无调度记录" />
    </view>

    <!-- Normal -->
    <view v-else class="list-body">
      <u-cell-group inset>
        <u-cell
          v-for="item in list"
          :key="item.id"
        >
          <template #title>
            <view class="cell-content">
              <text class="cell-title">{{ item.title || item.event_title }}</text>
              <view class="cell-meta">
                <text class="meta-text">{{ item.resource_name || item.resource }}</text>
                <text class="meta-date">{{ formatDateTime(item.dispatch_time || item.created_at) }}</text>
              </view>
            </view>
          </template>
          <template #value>
            <u-tag
              :type="statusTagType(item.status)"
              size="mini"
            >
              {{ statusLabel(item.status) }}
            </u-tag>
          </template>
        </u-cell>
      </u-cell-group>
    </view>
  </view>
</template>

<script>
import { request } from '../../utils/request'

export default {
  data() {
    return {
      activeTabIndex: 0,
      loading: false,
      errorMsg: '',
      list: [],
      statusMap: ['', 'pending', 'dispatched', 'completed'],
    }
  },
  onLoad() {
    this.fetchList(true)
  },
  onPullDownRefresh() {
    this.fetchList(true).then(function () {
      uni.stopPullDownRefresh()
    })
  },
  methods: {
    async fetchList(reset) {
      if (reset) {
        this.loading = true
      }
      this.errorMsg = ''

      try {
        var params = {}
        var statusVal = this.statusMap[this.activeTabIndex]
        if (statusVal) params.status = statusVal

        var res = await request({
          url: '/api/v1/emergency-dispatches',
          data: params,
        })
        var data = Array.isArray(res) ? res : (res && res.data) || res || {}
        var items = Array.isArray(data) ? data : (data && data.items) || []

        this.list = items
      } catch (e) {
        this.errorMsg = '网络异常，请稍后重试'
      } finally {
        this.loading = false
      }
    },
    onTabChange(index) {
      this.activeTabIndex = index
      this.fetchList(true)
    },
    goBack() {
      uni.navigateBack()
    },
    statusLabel(status) {
      var map = {
        pending: '待响应',
        dispatched: '已调度',
        completed: '已完成',
      }
      return map[status] || status || '未知'
    },
    statusTagType(status) {
      var map = {
        pending: 'warning',
        dispatched: 'primary',
        completed: 'success',
      }
      return map[status] || 'default'
    },
    formatDateTime(iso) {
      if (!iso) return ''
      var d = new Date(iso)
      var m = d.getMonth() + 1
      var day = d.getDate()
      var h = d.getHours()
      var min = d.getMinutes()
      return (
        d.getFullYear() +
        '-' +
        (m < 10 ? '0' : '') +
        m +
        '-' +
        (day < 10 ? '0' : '') +
        day +
        ' ' +
        (h < 10 ? '0' : '') +
        h +
        ':' +
        (min < 10 ? '0' : '') +
        min
      )
    },
  },
}
</script>

<style scoped>
.dispatches-page {
  min-height: 100vh;
  background: var(--color-bg);
  padding-bottom: env(safe-area-inset-bottom);
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
  color: var(--color-text-secondary);
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
  background: var(--color-primary);
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}

/* List */
.list-body {
  padding: 12px 0 24px;
}

.cell-content {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
}

.cell-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text);
}

.cell-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.meta-text {
  font-size: 12px;
  color: var(--color-text-secondary);
}

.meta-date {
  font-size: 12px;
  color: var(--color-text-placeholder);
}
</style>
