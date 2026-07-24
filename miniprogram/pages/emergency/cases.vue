<template>
  <view class="cases-page">
    <!-- Nav -->
    <van-nav-bar
      title="救援案例库"
      left-arrow
      @click-left="goBack"
    />

    <!-- Search -->
    <van-sticky>
      <van-search
        v-model="searchText"
        placeholder="搜索救援案例"
        shape="round"
        @search="onSearch"
      />
    </van-sticky>

    <!-- Tabs -->
    <van-tabs
      :active="activeType"
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
      <van-empty image="search" description="暂无案例" />
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
      <van-cell-group inset>
        <van-cell
          v-for="item in list"
          :key="item.id"
          is-link
          @tap="goDetail(item)"
        >
          <template #title>
            <view class="case-title-row">
              <text class="case-emoji">{{ eventEmoji(item.event_type) }}</text>
              <text class="case-title">{{ item.title }}</text>
            </view>
          </template>
          <template #label>
            <view class="case-meta">
              <van-tag :type="eventTagType(item.event_type)" size="small">
                {{ item.event_type || '未知' }}
              </van-tag>
              <text v-if="item.date" class="meta-text">{{ item.date }}</text>
            </view>
            <view class="case-extra">
              <text v-if="item.location" class="extra-text">{{ item.location }}</text>
              <text v-if="item.drone_model" class="extra-text">{{ item.drone_model }}</text>
            </view>
          </template>
          <template #value>
            <van-tag
              :type="resultTagType(item.result)"
              :color="resultColor(item.result)"
              size="small"
            >
              {{ resultLabel(item.result) }}
            </van-tag>
          </template>
        </van-cell>
      </van-cell-group>

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
        { label: '山火', value: '山火' },
        { label: '洪水', value: '洪水' },
        { label: '地震', value: '地震' },
        { label: '搜救', value: '搜救' },
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
    eventEmoji(type) {
      var map = {
        '山火': '🔥',
        '洪水': '🌊',
        '地震': '🏔',
        '搜救': '🔍',
        '其他': '🛡',
      }
      return map[type] || '🛡'
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
    resultColor(result) {
      var map = { '成功': '#34c759', '部分': '#ff9f0a', '失败': '#ef4444' }
      return map[result] || ''
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
.cases-page {
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
  background: #1989fa;
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}

.list-body {
  padding: 12px 0 24px;
}

.case-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.case-emoji {
  font-size: 18px;
  flex-shrink: 0;
}

.case-title {
  font-size: 15px;
  font-weight: 600;
  color: #323233;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.case-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 6px;
}

.meta-text {
  font-size: 12px;
  color: #969799;
}

.case-extra {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 4px;
}

.extra-text {
  font-size: 12px;
  color: #c8c9cc;
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
