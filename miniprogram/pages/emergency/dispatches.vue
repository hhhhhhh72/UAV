<template>
  <view class="dispatches-page">
    <van-nav-bar
      title="调度记录"
      fixed
      placeholder
      left-arrow
      @click-left="goBack"
    />

    <van-tabs
      :active="activeTab"
      @change="onTabChange"
      sticky
      :offset-top="0"
    >
      <van-tab title="全部" name="all" />
      <van-tab title="待响应" name="pending" />
      <van-tab title="已调度" name="dispatched" />
      <van-tab title="已完成" name="completed" />
    </van-tabs>

    <!-- Loading -->
    <view v-if="loading && list.length === 0" class="loading-state">
      <van-loading size="24">加载中...</van-loading>
    </view>

    <!-- Error -->
    <view v-else-if="errorMsg && list.length === 0" class="state-view">
      <van-empty description="加载失败" image="error" />
      <view class="retry-btn" @tap="fetchList(true)">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Empty -->
    <view v-else-if="!loading && list.length === 0" class="state-view">
      <van-empty image="search" description="暂无调度记录" />
    </view>

    <!-- Normal -->
    <view v-else class="list-body">
      <van-cell-group inset>
        <van-cell
          v-for="item in list"
          :key="item.id"
          :title="item.title || item.event_title"
        >
          <template #label>
            <view class="cell-meta">
              <text class="meta-text">{{ item.resource_name || item.resource }}</text>
              <text class="meta-date">{{ formatDateTime(item.dispatch_time || item.created_at) }}</text>
            </view>
          </template>
          <template #value>
            <van-tag
              :type="statusTagType(item.status)"
              size="small"
            >
              {{ statusLabel(item.status) }}
            </van-tag>
          </template>
        </van-cell>
      </van-cell-group>
    </view>
  </view>
</template>

<script>
import { request } from '../../utils/request'

export default {
  data() {
    return {
      activeTab: 'all',
      loading: false,
      errorMsg: '',
      list: [],
      statusMap: {
        all: '',
        pending: 'pending',
        dispatched: 'dispatched',
        completed: 'completed',
      },
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
        var statusVal = this.statusMap[this.activeTab]
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
    onTabChange(e) {
      this.activeTab = e.detail.name
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
  background: #f7f8fa;
  padding-bottom: env(safe-area-inset-bottom);
}

/* State views */
.loading-state {
  display: flex;
  justify-content: center;
  padding: 80px 0;
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
  background: #1989fa;
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
  margin-top: 6px;
}

.meta-text {
  font-size: 12px;
  color: #969799;
}

.meta-date {
  font-size: 12px;
  color: #c8c9cc;
}
</style>
