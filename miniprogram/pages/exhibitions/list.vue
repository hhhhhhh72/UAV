<template>
  <view class="page-container">
    <!-- Nav -->
    <u-nav-bar
      title="展会排期"
      show-back
      @back="goBack"
    />

    <!-- Search -->
    <u-sticky>
      <u-search
        v-model="searchText"
        placeholder="搜索展会"
        @search="onSearch"
      />
    </u-sticky>

    <!-- Category tabs -->
    <u-tabs
      :active="tabIndex"
      :titles="tabTitles"
      @change="onTabChange"
    />

    <!-- Loading state -->
    <view v-if="loading && list.length === 0" class="loading-state">
      <view class="loading-inline">
        <u-loading size="24rpx" />
        <text>加载中...</text>
      </view>
    </view>

    <!-- Empty state -->
    <view v-else-if="!loading && list.length === 0 && !errorMsg" class="empty-state-wrapper">
      <u-empty description="暂无展会" />
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg && list.length === 0" class="error-state">
      <u-empty description="加载失败" />
      <view class="retry-btn" @tap="fetchList(true)">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Normal state -->
    <view v-else class="list-body">
      <view class="card-list">
        <view
          v-for="item in list"
          :key="item.id"
          class="expo-card"
          @tap="goDetail(item)"
        >
          <image
            v-if="item.poster_image || item.cover_image || item.image"
            :src="item.poster_image || item.cover_image || item.image"
            mode="aspectFill"
            class="card-thumb"
          />
          <view v-else class="card-thumb card-thumb--placeholder"><text>展</text></view>
          <view class="card-body">
            <text class="card-title">{{ item.title }}</text>
            <text v-if="item.location" class="card-desc">{{ item.location }}</text>
            <view class="card-tags">
              <u-tag :type="categoryTag(item.category)">
                {{ item.category || '通用' }}
              </u-tag>
              <u-tag :type="statusTagType(item.status)">
                {{ statusLabel(item.status) }}
              </u-tag>
            </view>
            <view v-if="item.dates || item.start_date" class="card-extra">
              <text class="extra-text">{{ formatDateRange(item) }}</text>
            </view>
            <text class="card-num">共 {{ item.booth_count || 0 }} 个展位</text>
          </view>
        </view>
      </view>

      <!-- Load more -->
      <view v-if="list.length > 0" class="load-more">
        <view v-if="loadingMore" class="loading-inline">
          <u-loading size="20rpx" />
          <text>加载更多...</text>
        </view>
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
      activeCategory: '',
      loading: false,
      loadingMore: false,
      errorMsg: '',
      list: [],
      page: 1,
      pageSize: 20,
      hasMore: true,
      categoryTabs: [
        { label: '全部', value: '' },
        { label: '无人机展', value: '无人机展' },
        { label: '航模展', value: '航模展' },
        { label: '低空经济', value: '低空经济' },
        { label: '其他', value: '其他' },
      ],
    }
  },
  computed: {
    tabTitles() {
      return this.categoryTabs.map(function (t) { return t.label })
    },
    tabIndex() {
      var idx = this.categoryTabs.findIndex(function (t) { return t.value === this.activeCategory }.bind(this))
      return idx >= 0 ? idx : 0
    },
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
        if (this.activeCategory) params.category = this.activeCategory
        if (this.searchText) params.q = this.searchText
        params.page = this.page
        params.page_size = this.pageSize

        var res = await request({ url: '/api/v1/exhibitions', data: params })
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
    onTabChange(index) {
      var tab = this.categoryTabs[index]
      this.activeCategory = tab ? tab.value : ''
      this.fetchList(true)
    },
    goDetail(item) {
      uni.navigateTo({ url: '/pages/exhibitions/booth?id=' + encodeURIComponent(item.id) })
    },
    goBack() {
      uni.navigateBack()
    },
    categoryTag(category) {
      var map = {
        '无人机展': 'primary',
        '航模展': 'warning',
        '低空经济': 'success',
        '其他': 'default',
      }
      return map[category] || 'default'
    },
    statusLabel(status) {
      var map = {
        'open': '报名中',
        'ongoing': '进行中',
        'closed': '已结束',
      }
      return map[status] || status || '未知'
    },
    statusTagType(status) {
      var map = {
        'open': 'success',
        'ongoing': 'primary',
        'closed': 'default',
      }
      return map[status] || 'default'
    },
    formatDateRange(item) {
      if (item.dates) return item.dates
      var start = item.start_date || ''
      var end = item.end_date || ''
      if (start && end) return start + ' ~ ' + end
      return start || end || ''
    },
  },
}
</script>

<style scoped>
.page-container {
  min-height: 100vh;
  background: var(--color-bg);
  padding-bottom: 40px;
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
  background: var(--color-warning);
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}

.list-body {
  padding: 12px 0 24px;
}

.card-list {
  display: flex;
  flex-direction: column;
  padding: 0 12px;
  gap: 12px;
}

/* Expo card 自定义样式 */
.expo-card {
  display: flex;
  gap: 12px;
  background: var(--color-bg-card);
  border-radius: 12px;
  padding: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.03);
}

.expo-card:active {
  transform: translateY(2rpx);
}

.card-thumb {
  width: 160rpx;
  height: 160rpx;
  border-radius: 8px;
  flex-shrink: 0;
  background: var(--color-primary-light);
}

.card-thumb--placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 36rpx;
  color: var(--color-primary);
}

.card-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.card-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text);
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-desc {
  font-size: 13px;
  color: var(--color-text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-tags {
  display: flex;
  align-items: center;
  gap: 8px;
}

.card-extra {
  display: flex;
  align-items: center;
  gap: 12px;
}

.extra-text {
  font-size: 12px;
  color: var(--color-text-secondary);
}

.card-num {
  font-size: 12px;
  color: var(--color-primary);
  font-weight: 600;
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
