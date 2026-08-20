<template>
  <view class="demand-list-page">
    <!-- 搜索 -->
    <u-sticky>
      <u-search
        v-model="searchText"
        placeholder="搜索需求、项目"
        @search="onSearch"
      />
    </u-sticky>

    <!-- 筛选 + 排序 -->
    <view class="filter-bar">
      <scroll-view scroll-x :show-scrollbar="false" class="filter-scroll">
        <view class="filter-tabs">
          <view
            v-for="(tab, index) in bizTypeTabs"
            :key="index"
            class="filter-tab"
            :class="{ active: activeBizType === tab.value }"
            @tap="switchBizType(tab.value)"
          >
            {{ tab.label }}
          </view>
        </view>
      </scroll-view>
      <view class="sort-trigger" @tap="showSortPicker">
        <text class="sort-label">{{ currentSortLabel }}</text>
        <text class="sort-arrow">▾</text>
      </view>
    </view>

    <!-- 加载状态 -->
    <view v-if="loading && list.length === 0" class="state-wrap">
      <u-loading size="28rpx" />
      <text class="state-text">加载中...</text>
    </view>

    <!-- 筛选无结果 -->
    <view v-else-if="!loading && list.length === 0 && !errorMsg && (activeBizType || searchText)" class="state-wrap">
      <u-empty description="没有符合条件的需求" />
      <view class="state-reset" @tap="resetFilter">清除筛选</view>
    </view>

    <!-- 空数据 -->
    <view v-else-if="!loading && list.length === 0 && !errorMsg" class="state-wrap">
      <u-empty description="暂无需求" />
    </view>

    <!-- 错误状态 -->
    <view v-else-if="errorMsg && list.length === 0" class="state-wrap">
      <u-empty description="加载失败" />
      <view class="retry-btn" @tap="fetchList(true)"><text>重新加载</text></view>
    </view>

    <!-- 列表 -->
    <view v-else class="list-body">
      <!-- 首条重点卡 -->
      <view class="featured-card" @tap="goDetail(list[0])">
        <image :src="featuredImage(list[0])" mode="aspectFill" class="featured-img" />
        <view class="featured-mask"></view>
        <view class="featured-copy">
          <view class="featured-tags">
            <text class="tag-blue">{{ bizTypeLabel(list[0].biz_type) }}</text>
            <text v-if="list[0].district" class="tag-white">{{ list[0].district }}</text>
          </view>
          <text class="featured-title">{{ list[0].title }}</text>
          <text class="featured-meta">{{ formatBudget(list[0].budget_fen) }} · {{ formatDate(list[0].created_at) }}</text>
        </view>
      </view>

      <!-- 紧凑卡 -->
      <view
        v-for="item in list.slice(1)"
        :key="item.id"
        class="compact-card"
        @tap="goDetail(item)"
      >
        <text class="compact-title">{{ item.title }}</text>
        <view class="compact-meta">
          <text class="tag-blue tag-mini">{{ bizTypeLabel(item.biz_type) }}</text>
          <text v-if="item.district" class="meta-text">{{ item.district }}</text>
          <text class="meta-text">{{ formatBudget(item.budget_fen) }}</text>
          <text class="meta-date">{{ formatDate(item.created_at) }}</text>
        </view>
      </view>

      <!-- 加载更多 -->
      <view class="load-more">
        <view v-if="loadingMore" class="loading-inline">
          <u-loading size="24rpx" />
          <text>加载更多...</text>
        </view>
        <text v-else-if="!hasMore" class="no-more">没有更多了</text>
      </view>
    </view>

    <!-- 排序弹层 -->
    <u-popup
      :show="sortPickerVisible"
      position="bottom"
      round
      @close="sortPickerVisible = false"
    >
      <view class="sheet">
        <view class="sheet-title">排序方式</view>
        <view
          v-for="a in sortActions"
          :key="a.value"
          class="sheet-item"
          :class="{ on: a.value === currentSort }"
          @tap="onSortSelect(a)"
        >
          <text class="sheet-name">{{ a.name }}</text>
          <text v-if="a.value === currentSort" class="sheet-check">✓</text>
        </view>
        <view class="sheet-cancel" @tap="sortPickerVisible = false">取消</view>
      </view>
    </u-popup>
  </view>
</template>

<script>
import { request, getStoredUser, BASE_URL } from '../../utils/request'
import { BIZ_TYPE_TABS, bizTypeLabel as bizTypeLabelOf } from '../../utils/enums'

export default {
  data() {
    return {
      searchText: '',
      activeBizType: '',
      currentSort: 'newest',
      sortPickerVisible: false,
      loading: false,
      loadingMore: false,
      errorMsg: '',
      list: [],
      page: 1,
      pageSize: 20,
      hasMore: true,
      bizTypeTabs: BIZ_TYPE_TABS,
      sortActions: [
        { name: '最新发布', value: 'newest' },
        { name: '预算最高', value: 'budget_desc' },
        { name: '预算最低', value: 'budget_asc' },
      ],
    }
  },
  computed: {
    currentSortLabel() {
      const found = this.sortActions.find(function (a) { return a.value === this.currentSort }.bind(this))
      return found ? found.name : '最新发布'
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
        const params = {}
        if (this.activeBizType) params.biz_type = this.activeBizType
        if (this.currentSort) params.sort = this.currentSort
        if (this.searchText) params.q = this.searchText
        params.page = this.page
        params.page_size = this.pageSize

        const res = await request({ url: '/api/v1/demands', data: params })
        const data = Array.isArray(res) ? res : (res && res.data) || res || {}
        const items = Array.isArray(data) ? data : (data && data.items) || []
        const total = (data && data.total) != null ? data.total : items.length

        if (reset) {
          this.list = items
        } else {
          this.list = this.list.concat(items)
        }
        this.hasMore = this.list.length < total
        // 后端忽略 sort 参数：预算排序在前端本地完成（缺失 budget_fen 视为 0）
        this.applySort()
      } catch (e) {
        this.errorMsg = '网络异常，请稍后重试'
      } finally {
        this.loading = false
        this.loadingMore = false
      }
    },
    applySort() {
      if (this.currentSort === 'budget_desc') {
        this.list.sort(function (a, b) {
          return (Number(b.budget_fen) || 0) - (Number(a.budget_fen) || 0)
        })
      } else if (this.currentSort === 'budget_asc') {
        this.list.sort(function (a, b) {
          return (Number(a.budget_fen) || 0) - (Number(b.budget_fen) || 0)
        })
      }
    },
    async loadMore() {
      this.page++
      await this.fetchList(false)
    },
    onSearch() {
      this.fetchList(true)
    },
    switchBizType(value) {
      this.activeBizType = value
      this.fetchList(true)
    },
    resetFilter() {
      this.activeBizType = ''
      this.searchText = ''
      this.fetchList(true)
    },
    showSortPicker() {
      this.sortPickerVisible = true
    },
    onSortSelect(action) {
      this.currentSort = action.value
      this.sortPickerVisible = false
      this.fetchList(true)
    },
    goDetail(item) {
      uni.navigateTo({ url: '/pages/demands/detail?id=' + encodeURIComponent(item.id) })
    },
    bizTypeLabel(type) {
      return bizTypeLabelOf(type)
    },
    featuredImage(item) {
      try {
        const arr = typeof item.images === 'string' ? JSON.parse(item.images) : item.images
        // 存库为相对路径 /uploads/xxx，预览必须补全域名，否则小程序按本地包内资源加载 → 白图
        if (Array.isArray(arr) && arr[0]) return this.resolveUrl(arr[0])
      } catch {}
      return '/static/home/hero-inspection.jpg'
    },
    resolveUrl(u) {
      if (!u) return ''
      if (u.indexOf('http') === 0) return u
      return BASE_URL + u
    },
    formatBudget(fen) {
      if (fen == null || fen === 0) return '面议'
      var yuan = (fen / 100).toFixed(2)
      return yuan.replace(/\.00$/, '') + ' 元'
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
.demand-list-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: env(safe-area-inset-bottom);
}

/* 筛选 + 排序 */
.filter-bar {
  display: flex;
  align-items: center;
  background: #fff;
  padding: 8px 12px;
  gap: 8px;
}

.filter-scroll {
  flex: 1;
  white-space: nowrap;
}

.filter-tabs {
  display: inline-flex;
  gap: 8px;
}

.filter-tab {
  flex-shrink: 0;
  padding: 6px 16px;
  border-radius: 8px;
  font-size: 13px;
  color: #344054;
  background: #F4F6F8;
  border: 1px solid #EEF1F4;
  transition: all 0.2s;
}

.filter-tab.active {
  color: #fff;
  background: #0A66C2;
  border-color: #0A66C2;
}

.sort-trigger {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 0;
  flex-shrink: 0;
}

.sort-label {
  font-size: 13px;
  color: #344054;
}

.sort-arrow {
  font-size: 10px;
  color: #98A2B3;
}

/* 状态 */
.state-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 64px 0;
}

.state-text {
  font-size: 13px;
  color: #667085;
}

.state-reset {
  padding: 8px 24px;
  border-radius: 8px;
  border: 1px solid #0A66C2;
  color: #0A66C2;
  font-size: 13px;
}

.retry-btn {
  padding: 8px 24px;
  border-radius: 8px;
  background: #0A66C2;
  color: #fff;
  font-size: 13px;
}

/* 列表 */
.list-body {
  padding: 12px;
}

/* 首条重点卡 */
.featured-card {
  position: relative;
  height: 276rpx;
  border-radius: 8px;
  overflow: hidden;
  margin-bottom: 8px;
}

.featured-img {
  width: 100%;
  height: 100%;
}

.featured-mask {
  position: absolute;
  inset: 0;
  background: rgba(16, 24, 40, 0.45);
}

.featured-copy {
  position: absolute;
  left: 16px;
  right: 16px;
  bottom: 12px;
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.featured-tags {
  display: flex;
  gap: 6px;
}

.tag-blue {
  font-size: 11px;
  color: #0A66C2;
  background: #EAF3FB;
  padding: 2px 8px;
  border-radius: 4px;
}

.tag-white {
  font-size: 11px;
  color: #fff;
  background: rgba(255, 255, 255, 0.22);
  padding: 2px 8px;
  border-radius: 4px;
}

.featured-title {
  font-size: 16px;
  font-weight: 700;
  color: #fff;
  line-height: 1.35;
}

.featured-meta {
  font-size: 11px;
  color: rgba(255, 255, 255, 0.85);
}

/* 紧凑卡 */
.compact-card {
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 8px;
}

.compact-title {
  font-size: 14px;
  font-weight: 600;
  color: #17212B;
  line-height: 1.4;
  display: block;
}

.compact-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
}

.tag-mini {
  font-size: 10px;
}

.meta-text {
  font-size: 12px;
  color: #667085;
}

.meta-date {
  font-size: 11px;
  color: #98A2B3;
  margin-left: auto;
}

/* 加载更多 */
.load-more {
  text-align: center;
  padding: 16px 0;
}

.loading-inline {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #667085;
}

.no-more {
  color: #98A2B3;
  font-size: 12px;
}

/* 排序弹层 */
.sheet {
  background: #fff;
  border-radius: 16rpx 16rpx 0 0;
  padding-bottom: env(safe-area-inset-bottom);
}

.sheet-title {
  text-align: center;
  font-size: 15px;
  font-weight: 600;
  color: #17212B;
  padding: 16px 0 8px;
}

.sheet-item {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 8px;
  padding: 14px 24px;
  font-size: 14px;
  color: #17212B;
}

.sheet-item.on {
  color: #0A66C2;
  font-weight: 600;
}

.sheet-check {
  font-size: 14px;
}

.sheet-cancel {
  text-align: center;
  padding: 14px;
  font-size: 14px;
  color: #667085;
  border-top: 1px solid #EEF1F4;
}
</style>
