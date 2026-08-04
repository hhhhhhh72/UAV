<template>
  <view class="jobs-page">
    <!-- Nav -->
    <u-nav-bar
      title="招聘求职"
      show-back
      @back="goBack"
    />

    <!-- Search -->
    <u-sticky>
      <u-search
        v-model="searchText"
        placeholder="搜索职位"
        @search="onSearch"
      />
    </u-sticky>

    <!-- Tabs -->
    <u-tabs
      :active="activeTabIndex"
      :titles="typeTitles"
      @change="onTabChange"
    />

    <!-- Loading state -->
    <view v-if="loading && list.length === 0" class="loading-state">
      <view class="loading-inline">
        <u-loading size="28rpx" />
        <text>加载中...</text>
      </view>
    </view>

    <!-- Empty state -->
    <view v-else-if="!loading && list.length === 0 && !errorMsg" class="empty-state-wrapper">
      <u-empty description="暂无职位" />
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
      <u-cell-group inset>
        <u-cell
          v-for="item in list"
          :key="item.id"
        >
          <template #title>
            <view class="cell-content">
              <text class="job-title">{{ item.title }}</text>
              <view class="cell-meta">
                <text class="company-text">{{ item.company }}</text>
                <text v-if="item.salary" class="salary-text">{{ item.salary }}</text>
              </view>
              <view class="cell-extra">
                <u-tag v-if="item.type" type="primary" size="mini">{{ item.type }}</u-tag>
                <text v-if="item.location" class="extra-text">{{ item.location }}</text>
                <text class="date-text">{{ formatDate(item.created_at || item.date) }}</text>
              </view>
            </view>
          </template>
        </u-cell>
      </u-cell-group>

      <!-- Load more -->
      <view v-if="list.length > 0" class="load-more">
        <view v-if="loadingMore" class="loading-inline">
          <u-loading size="24rpx" />
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
        { label: '全职', value: '全职' },
        { label: '兼职', value: '兼职' },
        { label: '实习', value: '实习' },
        { label: '项目制', value: '项目制' },
      ],
    }
  },
  computed: {
    // u-tabs 只接受字符串标题数组 + 数字 active 索引，映射自 typeTabs
    typeTitles() {
      return this.typeTabs.map(function (t) { return t.label })
    },
    activeTabIndex() {
      for (var i = 0; i < this.typeTabs.length; i++) {
        if (this.typeTabs[i].value === this.activeType) return i
      }
      return 0
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
        if (this.activeType) params.type = this.activeType
        if (this.searchText) params.q = this.searchText
        params.page = this.page
        params.page_size = this.pageSize

        var res = await request({ url: '/api/v1/jobs', data: params })
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
      this.activeType = this.typeTabs[index].value
      this.fetchList(true)
    },
    goBack() {
      uni.navigateBack()
    },
    formatDate(iso) {
      if (!iso) return ''
      var d = new Date(iso)
      var m = d.getMonth() + 1
      var day = d.getDate()
      return d.getFullYear() + '-' + (m < 10 ? '0' : '') + m + '-' + (day < 10 ? '0' : '') + day
    },
  },
}
</script>

<style scoped>
.jobs-page {
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
  justify-content: center;
  gap: 8px;
  font-size: 13px;
  color: var(--color-text-secondary);
}

.cell-content {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
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
  border-radius: 20px;
  font-size: 14px;
}

.list-body {
  padding: 12px 0 24px;
}

.job-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--color-text);
}

.cell-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 4px;
}

.company-text {
  font-size: 13px;
  color: var(--color-text-secondary);
}

.salary-text {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-success);
}

.cell-extra {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  flex-wrap: wrap;
}

.extra-text {
  font-size: 12px;
  color: var(--color-text-secondary);
}

.date-text {
  font-size: 12px;
  color: var(--color-text-placeholder);
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
