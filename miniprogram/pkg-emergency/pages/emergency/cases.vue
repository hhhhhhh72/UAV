<template>
  <view class="cases-page">
    <u-nav-bar
      title="救援案例库"
      show-back
      @back="goBack"
    />

    <u-sticky>
      <u-search
        v-model="searchText"
        placeholder="搜索救援案例"
        @search="onSearch"
      />
      <u-tabs
        v-model:active="typeIndex"
        :titles="typeTitles"
        @change="onTypeChange"
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
      <u-empty description="暂无案例" />
    </view>

    <!-- List -->
    <view v-else class="list-body">
      <view class="c-list">
        <view v-for="item in list" :key="item.id" class="c-card">
          <view class="c-top">
            <view class="c-icon" :style="eventIconStyle(item.event_type)"><text>{{ eventIcon(item.event_type) }}</text></view>
            <text class="c-title">{{ item.title || '未命名案例' }}</text>
          </view>

          <view class="c-meta">
            <u-tag :type="eventTagType(item.event_type)" size="mini" :round="false" plain>{{ item.event_type || '未知' }}</u-tag>
            <text v-if="item.date" class="c-text">{{ item.date }}</text>
          </view>

          <view v-if="item.location || item.drone_model" class="c-extra">
            <text v-if="item.location" class="c-text">{{ item.location }}</text>
            <text v-if="item.drone_model" class="c-text">{{ item.drone_model }}</text>
          </view>

          <view v-if="item.result" class="c-result">
            <text class="c-result-label">处置结果</text>
            <u-tag :type="resultTagType(item.result)" size="mini" :round="false" plain>{{ resultLabel(item.result) }}</u-tag>
          </view>
        </view>

        <!-- Load more -->
        <view v-if="list.length > 0" class="load-more">
          <view v-if="loadingMore" class="loading-inline">
            <u-loading size="24rpx" color="#667085" />
            <text>加载更多...</text>
          </view>
          <text v-else-if="!hasMore" class="no-more">没有更多了</text>
        </view>
      </view>
      <view class="bottom-space" />
    </view>
  </view>
</template>

<script>
import { request } from '../../../utils/request'

export default {
  data() {
    return {
      searchText: '',
      activeType: '',
      typeIndex: 0,
      typeTitles: ['全部', '山火', '洪水', '地震', '搜救', '其他'],
      typeMap: ['', '山火', '洪水', '地震', '搜救', '其他'],
      loading: false,
      loadingMore: false,
      errorMsg: '',
      list: [],
      page: 1,
      pageSize: 20,
      hasMore: true,
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
  onReachBottom() {
    if (!this.loadingMore && this.hasMore) {
      this.loadMore()
    }
  },
  methods: {
    async fetchList(reset) {
      if (reset) {
        this.page = 1
        this.hasMore = true
        this.loading = true
      } else {
        this.loadingMore = true
      }
      this.errorMsg = ''

      try {
        var params = {}
        if (this.activeType) params.event_type = this.activeType
        if (this.searchText) params.q = this.searchText
        params.page = this.page
        params.page_size = this.pageSize

        var res = await request({ url: '/api/v1/rescue-cases', data: params })
        var data = Array.isArray(res) ? res : (res && res.data) || res || {}
        var items = Array.isArray(data) ? data : (data && data.items) || []
        var total = (data && data.total) != null ? data.total : items.length

        if (reset) {
          this.list = items
        } else {
          this.list = this.list.concat(items)
        }
        this.hasMore = this.list.length < total
      } catch (e) {
        this.errorMsg = '网络异常，请稍后重试'
      } finally {
        this.loading = false
        this.loadingMore = false
      }
    },
    loadMore() {
      this.page++
      this.fetchList(false)
    },
    onSearch() {
      this.fetchList(true)
    },
    onTypeChange(index) {
      this.typeIndex = index
      this.activeType = this.typeMap[index]
      this.fetchList(true)
    },
    /* 事件类型字符图标（低饱和色块，非 emoji） */
    eventIcon(type) {
      var map = {
        '山火': '火',
        '洪水': '水',
        '地震': '地',
        '搜救': '救',
        '其他': '卫',
      }
      return map[type] || '卫'
    },
    eventIconStyle(type) {
      var map = {
        '山火': { background: '#FFF0E6', color: '#E96012' },
        '洪水': { background: '#EAF3FB', color: '#0A66C2' },
        '地震': { background: '#F6F4FF', color: '#667085' },
        '搜救': { background: '#E9F7F0', color: '#168A55' },
      }
      return map[type] || { background: '#F4F6F8', color: '#667085' }
    },
    eventTagType(type) {
      var map = {
        '山火': 'danger',
        '洪水': 'primary',
        '地震': 'warning',
        '搜救': 'success',
        '其他': 'default',
      }
      return map[type] || 'default'
    },
    resultLabel(result) {
      var map = {
        '成功': '成功',
        '部分': '部分成功',
        '失败': '失败',
      }
      return map[result] || result || '未知'
    },
    resultTagType(result) {
      var map = { '成功': 'success', '部分': 'warning', '失败': 'danger' }
      return map[result] || 'default'
    },
    goBack() {
      uni.navigateBack()
    },
  },
}
</script>

<style scoped>
.cases-page {
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

.c-list { padding: 0 24rpx; }

.c-card {
  background: #ffffff;
  border: 1px solid #EEF1F4;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}

.c-top {
  display: flex;
  align-items: flex-start;
  gap: 16rpx;
  margin-bottom: 14rpx;
}

.c-icon {
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

.c-title {
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

.c-meta {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin-bottom: 10rpx;
}

.c-extra {
  display: flex;
  align-items: center;
  gap: 24rpx;
  margin-bottom: 12rpx;
}

.c-text {
  font-size: 23rpx;
  color: #667085;
}

.c-result {
  display: flex;
  align-items: center;
  gap: 16rpx;
  background: #F4F6F8;
  border-radius: 8rpx;
  padding: 14rpx 16rpx;
}

.c-result-label {
  font-size: 24rpx;
  color: #98A2B3;
}

/* Load more */
.load-more {
  text-align: center;
  padding: 20rpx 0;
}

.no-more {
  font-size: 24rpx;
  color: #98A2B3;
}

.bottom-space { height: 24rpx; }
</style>
