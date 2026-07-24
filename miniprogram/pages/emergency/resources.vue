<template>
  <view class="resources-page">
    <!-- Nav -->
    <van-nav-bar
      title="应急资源"
      left-arrow
      @click-left="goBack"
    />

    <!-- Search -->
    <van-sticky>
      <van-search
        v-model="searchText"
        placeholder="搜索应急资源"
        shape="round"
        @search="onSearch"
      />
    </van-sticky>

    <!-- Tabs -->
    <van-tabs
      :active="activeType"
      color="#f97316"
      @change="onTabChange"
    >
      <van-tab
        v-for="tab in typeTabs"
        :key="tab.value"
        :title="tab.label"
        :name="tab.value"
      />
    </van-tabs>

    <!-- Loading state -->
    <view v-if="loading && list.length === 0" class="loading-state">
      <van-loading size="24">加载中...</van-loading>
    </view>

    <!-- Empty state -->
    <view v-else-if="!loading && list.length === 0 && !errorMsg" class="empty-state-wrapper">
      <van-empty image="search" description="暂无资源" />
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg && list.length === 0" class="error-state">
      <van-empty description="加载失败" image="error" />
      <view class="retry-btn" @tap="fetchList(true)">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Normal state -->
    <view v-else class="list-body">
      <van-grid :column-num="2" :gutter="12">
        <van-grid-item
          v-for="item in list"
          :key="item.id"
          @tap="goDetail(item)"
        >
          <view class="grid-card">
            <text class="grid-emoji">{{ resEmoji(item.res_type) }}</text>
            <text class="grid-name">{{ item.name }}</text>
            <text v-if="item.model" class="grid-model">{{ item.model }}</text>
            <view class="grid-footer">
              <van-tag type="warning" size="small">
                {{ item.quantity || 0 }}
              </van-tag>
              <view v-if="item.location" class="grid-location">
                <text class="location-text">{{ item.location }}</text>
              </view>
            </view>
          </view>
        </van-grid-item>
      </van-grid>

      <!-- Load more -->
      <view v-if="list.length > 0" class="load-more">
        <van-loading v-if="loadingMore" size="20">加载更多...</van-loading>
        <text v-else-if="!hasMore" class="no-more">没有更多了</text>
      </view>
    </view>
  </view>
</template>

<script>
import { request } from '../../utils/request'

export default {
  data() {
    return {
      searchText: '',
      activeType: '',
      loading: false,
      loadingMore: false,
      errorMsg: '',
      list: [],
      page: 1,
      pageSize: 20,
      hasMore: true,
      typeTabs: [
        { label: '全部', value: '' },
        { label: '无人机', value: '无人机' },
        { label: '通信', value: '通信' },
        { label: '照明', value: '照明' },
        { label: '医疗', value: '医疗' },
        { label: '其他', value: '其他' },
      ],
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
        if (this.activeType) params.res_type = this.activeType
        if (this.searchText) params.q = this.searchText
        params.page = this.page
        params.page_size = this.pageSize

        var res = await request({ url: '/api/v1/emergency-resources', data: params })
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
    async loadMore() {
      this.page++
      await this.fetchList(false)
    },
    onSearch() {
      this.fetchList(true)
    },
    onTabChange(e) {
      var name = e.detail ? e.detail.name : e
      this.activeType = name
      this.fetchList(true)
    },
    resEmoji(type) {
      var map = {
        '无人机': '🚁',
        '通信': '📡',
        '照明': '💡',
        '医疗': '🏥',
        '其他': '📦',
      }
      return map[type] || '📦'
    },
    goDetail(item) {
      uni.showToast({ title: '即将上线', icon: 'none' })
    },
    goBack() {
      uni.navigateBack()
    },
  },
}
</script>

<style scoped>
.resources-page {
  min-height: 100vh;
  background: #f7f8fa;
  padding-bottom: env(safe-area-inset-bottom);
}

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
  background: #f97316;
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}

.list-body {
  padding: 12px 0 24px;
}

/* Grid card */
.grid-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 12px 8px;
  width: 100%;
  box-sizing: border-box;
}

.grid-emoji {
  font-size: 36px;
  margin-bottom: 8px;
}

.grid-name {
  font-size: 14px;
  font-weight: 600;
  color: #323233;
  text-align: center;
  line-height: 1.3;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100%;
}

.grid-model {
  font-size: 12px;
  color: #969799;
  margin-bottom: 10px;
  text-align: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100%;
}

.grid-footer {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  width: 100%;
}

.grid-location {
  display: flex;
  align-items: center;
  gap: 2px;
}

.location-text {
  font-size: 11px;
  color: #c8c9cc;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 120px;
}

/* Load more */
.load-more {
  text-align: center;
  padding: 16px 0;
}

.no-more {
  color: #c8c9cc;
  font-size: 13px;
}
</style>
