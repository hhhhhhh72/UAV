<template>
  <view class="expert-list-page">
    <van-nav-bar
      title="专家智库"
      fixed
      placeholder
      left-arrow
      @click-left="goBack"
    />

    <!-- Search bar -->
    <van-search
      v-model="searchText"
      placeholder="搜索专家姓名或领域"
      shape="round"
      @search="onSearch"
    />

    <!-- Field filter tabs -->
    <view class="filter-tabs">
      <view
        v-for="(tab, index) in fieldTabs"
        :key="index"
        class="filter-tab"
        :class="{ active: activeField === tab.value }"
        @tap="switchField(tab.value)"
      >
        {{ tab.label }}
      </view>
    </view>

    <!-- Loading state -->
    <view v-if="loading && list.length === 0" class="loading-state">
      <van-loading size="24">加载中...</van-loading>
    </view>

    <!-- Empty state -->
    <view v-else-if="!loading && list.length === 0 && !errorMsg" class="empty-state-wrapper">
      <van-empty image="search" description="暂无专家信息" />
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
          :title="item.name"
          is-link
          @tap="goDetail(item)"
        >
          <template #label>
            <view class="cell-meta">
              <van-tag
                v-for="(f, fi) in parseFields(item.field)"
                :key="fi"
                type="primary"
                size="small"
              >
                {{ f }}
              </van-tag>
              <text v-if="item.organization" class="meta-text">{{ item.organization }}</text>
            </view>
          </template>
          <template #value>
            <text class="cell-value-text">{{ item.title || '' }}</text>
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
      activeField: '',
      loading: false,
      loadingMore: false,
      errorMsg: '',
      list: [],
      page: 1,
      pageSize: 20,
      hasMore: true,
      fieldTabs: [
        { label: '全部', value: '' },
        { label: '低空管控', value: '低空管控' },
        { label: '适航认证', value: '适航认证' },
        { label: '无人机研发', value: '无人机研发' },
        { label: '应急救援', value: '应急救援' },
        { label: '政策法规', value: '政策法规' },
        { label: '投融资', value: '投融资' },
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
        if (this.activeField) params.field = this.activeField
        if (this.searchText) params.q = this.searchText
        params.page = this.page
        params.page_size = this.pageSize

        var res = await request({ url: '/api/v1/experts', data: params })
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
    switchField(value) {
      this.activeField = value
      this.fetchList(true)
    },
    goDetail(item) {
      uni.navigateTo({ url: '/pages/experts/detail?id=' + encodeURIComponent(item.id) })
    },
    goBack() {
      uni.navigateBack()
    },
    parseFields(field) {
      if (!field) return []
      if (typeof field === 'string') {
        return field.split(/[,，]/).filter(Boolean)
      }
      if (Array.isArray(field)) return field
      return []
    },
  },
}
</script>

<style scoped>
.expert-list-page {
  min-height: 100vh;
  background: #f7f8fa;
  padding-bottom: env(safe-area-inset-bottom);
}

/* Filter tabs */
.filter-tabs {
  display: flex;
  padding: 10px 12px;
  gap: 8px;
  background: #fff;
  overflow-x: auto;
  white-space: nowrap;
  -webkit-overflow-scrolling: touch;
}

.filter-tabs::-webkit-scrollbar {
  display: none;
}

.filter-tab {
  flex-shrink: 0;
  padding: 6px 16px;
  border-radius: 20px;
  font-size: 13px;
  color: #646566;
  background: #f7f8fa;
  transition: all 0.2s;
}

.filter-tab.active {
  color: #fff;
  background: #0A66C2;
}

/* State views */
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
  background: #0A66C2;
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

.cell-value-text {
  font-size: 12px;
  color: #c8c9cc;
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
