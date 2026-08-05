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
        <u-loading size="24rpx" color="#667085" />
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
      <view class="d-list">
        <view
          v-for="item in list"
          :key="item.id"
          class="d-card"
        >
          <view class="d-top">
            <text class="d-title">{{ item.event_desc || item.title || '未命名事件' }}</text>
            <u-tag :type="statusTagType(item.status)" size="mini" :round="false" plain>{{ statusLabel(item.status) }}</u-tag>
          </view>

          <view v-if="item.location" class="d-loc-row">
            <u-icon name="location" size="24rpx" color="#667085" />
            <text class="d-loc">{{ item.location }}</text>
          </view>

          <view class="d-meta">
            <text class="d-time">{{ formatDateTime(item.start_time || item.created_at) }}</text>
            <text v-if="item.commander" class="d-commander">负责人 · {{ item.commander }}</text>
          </view>

          <view v-if="item.result" class="d-result">
            <text class="d-result-text">{{ item.result }}</text>
          </view>
        </view>
      </view>
      <view class="d-bottom-space" />
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
      tabTitles: ['全部', '待响应', '已调度', '已完成'],
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
    /* 状态归一：pending/dispatched/completed/ongoing/done/cancelled */
    statusKey(status) {
      if (status === 'completed' || status === 'done') return 'completed'
      if (status === 'pending') return 'pending'
      if (status === 'ongoing' || status === 'dispatched') return 'ongoing'
      if (status === 'cancelled') return 'cancelled'
      return 'pending'
    },
    statusLabel(status) {
      var map = {
        pending: '待响应',
        dispatched: '已调度',
        completed: '已完成',
        ongoing: '进行中',
        done: '已完成',
        cancelled: '已取消',
      }
      return map[status] || status || '未知'
    },
    statusTagType(status) {
      var key = this.statusKey(status)
      var map = {
        pending: 'warning',
        ongoing: 'primary',
        completed: 'success',
        cancelled: 'default',
      }
      return map[key] || 'default'
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
  background: #F4F6F8;
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

/* List */
.list-body {
  padding-top: 12px;
}

.d-list { padding: 0 24rpx; }

.d-card {
  background: #ffffff;
  border: 1px solid #EEF1F4;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}

.d-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12rpx;
  margin-bottom: 12rpx;
}

.d-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #17212B;
  flex: 1;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.d-loc-row {
  display: flex;
  align-items: center;
  gap: 8rpx;
  margin-bottom: 10rpx;
}

.d-loc {
  font-size: 25rpx;
  color: #667085;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.d-meta {
  display: flex;
  align-items: center;
  gap: 20rpx;
  flex-wrap: wrap;
  margin-bottom: 12rpx;
}

.d-time { font-size: 23rpx; color: #98A2B3; }
.d-commander { font-size: 23rpx; color: #98A2B3; }

.d-result {
  background: #F4F6F8;
  border-radius: 8rpx;
  padding: 14rpx 16rpx;
}

.d-result-text {
  font-size: 24rpx;
  color: #344054;
  line-height: 1.6;
}

.d-bottom-space { height: 24rpx; }
</style>
