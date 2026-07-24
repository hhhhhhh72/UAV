<template>
  <view class="page-container">
    <!-- Nav -->
    <van-nav-bar
      title="展会排期"
      left-arrow
      @click-left="goBack"
      custom-style="background: #f59e0b;"
    >
      <template #title>
        <text style="color: #fff;">展会排期</text>
      </template>
    </van-nav-bar>

    <!-- Search -->
    <van-sticky>
      <van-search
        v-model="searchText"
        placeholder="搜索展会"
        shape="round"
        @search="onSearch"
      />
    </van-sticky>

    <!-- Category tabs -->
    <van-tabs
      :active="activeCategory"
      @change="onTabChange"
    >
      <van-tab
        v-for="tab in categoryTabs"
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
      <van-empty image="search" description="暂无展会" />
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg && list.length === 0" class="error-state">
      <van-empty image="network" description="加载失败" />
      <view class="retry-btn" @tap="fetchList(true)">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Normal state -->
    <view v-else class="list-body">
      <view class="card-list">
        <van-card
          v-for="item in list"
          :key="item.id"
          :title="item.title"
          :desc="item.location || ''"
          :num="'共 ' + (item.booth_count || 0) + ' 个展位'"
          :thumb="item.poster_image || item.cover_image || item.image || ''"
          thumb-mode="aspectFill"
          @tap="goDetail(item)"
        >
          <template #tags>
            <view class="card-tags">
              <van-tag :type="categoryTag(item.category)" size="medium">
                {{ item.category || '通用' }}
              </van-tag>
              <van-tag :type="statusTagType(item.status)" size="medium">
                {{ statusLabel(item.status) }}
              </van-tag>
            </view>
            <view v-if="item.dates || item.start_date" class="card-extra">
              <text class="extra-text">{{ formatDateRange(item) }}</text>
            </view>
          </template>
        </van-card>
      </view>

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
    onTabChange(e) {
      var name = e.detail ? e.detail.name : e
      this.activeCategory = name
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
  background: #f7f8fa;
  padding-bottom: 40px;
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
  background: #f59e0b;
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
  margin-top: 4px;
}

.extra-text {
  font-size: 12px;
  color: #969799;
}

.load-more {
  text-align: center;
  padding: 16px 0;
}

.no-more {
  color: #c8c9cc;
  font-size: 13px;
}
</style>
