<template>
  <view class="demand-list-page">
    <!-- Search -->
    <u-sticky>
      <u-search
        v-model="searchText"
        placeholder="搜索需求"
        @search="onSearch"
      />
    </u-sticky>

    <!-- Biz type filter tabs -->
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

    <!-- Sort bar -->
    <view class="sort-bar">
      <view class="sort-trigger" @tap="showSortPicker">
        <text class="sort-label">{{ currentSortLabel }}</text>
        <text class="sort-arrow">▼</text>
      </view>
    </view>

    <!-- Loading state -->
    <view v-if="loading && list.length === 0" class="loading-state">
      <view class="loading-inline">
        <u-loading size="28rpx" />
        <text>加载中...</text>
      </view>
    </view>

    <!-- Empty state -->
    <view v-else-if="!loading && list.length === 0 && !errorMsg" class="empty-state-wrapper">
      <u-empty description="暂无需求" />
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
          is-link
          @click="goDetail(item)"
        >
          <template #title>
            <view class="cell-content">
              <text class="cell-title">{{ item.title }}</text>
              <view class="cell-meta">
                <u-tag
                  :type="bizTypeTagType(item.biz_type)"
                  size="mini"
                >
                  {{ bizTypeLabel(item.biz_type) }}
                </u-tag>
                <text v-if="item.district" class="meta-text">{{ item.district }}</text>
                <text class="meta-text">{{ formatBudget(item.budget_fen) }}</text>
                <text class="meta-date">{{ formatDate(item.created_at) }}</text>
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

    <!-- Sort action sheet -->
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
import { request, getStoredUser } from '../../utils/request'
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
    switchBizType(value) {
      this.activeBizType = value
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
    bizTypeTagType(type) {
      var map = {
        cable_inspection: 'primary',
        plant_transport: 'success',
        spray_pesticide: 'warning',
        trade_lease: 'danger',
        clean_paint: 'primary',
        other: 'default',
      }
      return map[type] || 'default'
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
  background: var(--color-bg);
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
  color: var(--color-text-secondary);
  background: var(--color-bg);
  transition: all 0.2s;
}

.filter-tab.active {
  color: #fff;
  background: var(--color-primary);
}

/* Sort bar */
.sort-bar {
  padding: 8px 16px;
  display: flex;
  justify-content: flex-end;
  background: #fff;
  border-bottom: 1px solid var(--color-border);
}

.sort-trigger {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 0;
}

.sort-label {
  font-size: 13px;
  color: var(--color-text-secondary);
}

.sort-arrow {
  font-size: 10px;
  color: var(--color-text-placeholder);
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
  justify-content: center;
  gap: 8px;
}

.loading-inline text {
  font-size: 13px;
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
  border-radius: 20px;
  font-size: 14px;
}

/* List */
.list-body {
  padding: 12px 0 24px;
}

.cell-content {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
}

.cell-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text);
}

.cell-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.meta-text {
  font-size: 12px;
  color: var(--color-text-secondary);
}

.meta-date {
  font-size: 12px;
  color: var(--color-text-placeholder);
}

/* Load more */
.load-more {
  text-align: center;
  padding: 16px 0;
}

.no-more {
  color: var(--color-text-placeholder);
  font-size: 13px;
}

/* Sort sheet */
.sheet {
  background: #fff;
  border-radius: 24rpx 24rpx 0 0;
  padding-bottom: env(safe-area-inset-bottom);
}

.sheet-title {
  text-align: center;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
  padding: 16px 0 8px;
}

.sheet-item {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 8px;
  padding: 14px 24px;
  font-size: 14px;
  color: var(--color-text);
}

.sheet-item.on {
  color: var(--color-primary);
  font-weight: 600;
}

.sheet-name {
  color: inherit;
}

.sheet-check {
  font-size: 14px;
}

.sheet-cancel {
  text-align: center;
  padding: 14px;
  font-size: 14px;
  color: var(--color-text-secondary);
  border-top: 1px solid var(--color-divider);
}
</style>
