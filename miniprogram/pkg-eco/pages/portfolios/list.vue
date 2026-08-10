<template>
  <view class="page-container">
    <!-- Nav -->
    <u-nav-bar
      title="品牌展示"
      show-back
      @back="goBack"
    />

    <!-- Search -->
    <u-sticky>
      <u-search
        v-model="searchText"
        placeholder="搜索品牌"
        @search="onSearch"
      />
    </u-sticky>

    <!-- Loading state -->
    <view v-if="loading && list.length === 0" class="loading-state">
      <view class="loading-inline">
        <u-loading size="24rpx" />
        <text>加载中...</text>
      </view>
    </view>

    <!-- Empty state -->
    <view v-else-if="!loading && list.length === 0 && !errorMsg" class="empty-state-wrapper">
      <u-empty description="暂无品牌展示" />
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg && list.length === 0" class="error-state">
      <u-empty description="加载失败" />
      <view class="retry-btn" @tap="fetchList(true)">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Normal state: brand grid -->
    <view v-else class="list-body">
      <view class="brand-grid">
        <view
          v-for="item in list"
          :key="item.id"
          class="brand-item"
        >
          <image
            v-if="item.logo || item.image"
            :src="item.logo || item.image"
            mode="aspectFill"
            class="grid-logo"
          />
          <view v-else class="grid-logo grid-logo--placeholder"><text>牌</text></view>
          <text class="grid-name">{{ item.name || item.company_name || '' }}</text>
          <u-tag
            v-if="item.industry"
            :type="industryTagType(item.industry)"
            size="mini"
            class="grid-tag"
          >
            {{ item.industry }}
          </u-tag>
        </view>
      </view>

      <!-- Load more -->
      <view v-if="list.length > 0" class="load-more">
        <view v-if="loadingMore" class="loading-inline">
          <u-loading size="20rpx" />
          <text>加载更多...</text>
        </view>
        <u-button
          v-else-if="hasMore"
          type="default"
          size="small"
          @click="loadMore"
        >
          加载更多
        </u-button>
        <text v-else class="no-more">没有更多了</text>
      </view>
    </view>
  </view>
</template>

<script>
import { request } from '../../../utils/request'

export default {
  data() {
    return {
      searchText: '',
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
        if (this.searchText) params.q = this.searchText
        params.page = this.page
        params.page_size = this.pageSize

        var res = await request({ url: '/api/v1/portfolios', data: params })
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
    goBack() {
      uni.navigateBack()
    },
    industryTagType(industry) {
      var map = {
        '无人机': 'primary',
        '飞控': 'warning',
        '载荷': 'success',
        '软件': 'danger',
        '材料': 'default',
        '电池': 'primary',
        '通信': 'success',
      }
      return map[industry] || 'default'
    },
  },
}
</script>

<style scoped>
.page-container {
  min-height: 100vh;
  background: var(--color-bg);
  padding-bottom: env(safe-area-inset-bottom);
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
  background: var(--color-primary);
  color: #fff;
  border-radius: 8px;
  font-size: 14px;
}

.list-body {
  padding: 12px 0 24px;
}

/* Brand grid CSS 两列 */
.brand-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  padding: 0 12px;
}

.brand-item {
  background: var(--color-bg-card);
  border-radius: 8px;
  padding: 20px 12px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.03);
}

.brand-item:active {
  transform: translateY(2rpx);
}

.grid-logo {
  width: 120rpx;
  height: 120rpx;
  border-radius: 50%;
  background: var(--color-bg);
  flex-shrink: 0;
}

.grid-logo--placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 36rpx;
  color: var(--color-primary);
}

.grid-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text);
  text-align: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 140px;
}

.grid-tag {
  margin-top: 2px;
}

.load-more {
  text-align: center;
  padding: 16px 0;
}

.no-more {
  color: var(--color-text-placeholder);
  font-size: 13px;
}
</style>
