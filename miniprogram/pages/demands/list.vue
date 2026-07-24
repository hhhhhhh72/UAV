<template>
  <view class="demand-list-page">
    <!-- Search -->
    <van-sticky>
      <van-search
        v-model="searchText"
        placeholder="搜索需求"
        shape="round"
        @search="onSearch"
      />
    </van-sticky>

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

    <!-- Sort picker -->
    <view class="sort-bar">
      <view class="sort-trigger" @tap="showSortPicker">
        <text class="sort-label">{{ currentSortLabel }}</text>
        <van-icon name="arrow-down" size="12" color="#969799" />
      </view>
    </view>

    <!-- Loading state -->
    <view v-if="loading && list.length === 0" class="loading-state">
      <van-loading size="24" vertical>加载中...</van-loading>
    </view>

    <!-- Empty state -->
    <view v-else-if="!loading && list.length === 0" class="empty-state">
      <van-empty description="暂无需求" />
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg" class="error-state">
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
          :title="item.title"
          is-link
          @tap="goDetail(item)"
        >
          <template #label>
            <view class="cell-meta">
              <van-tag
                :type="bizTypeTagType(item.biz_type)"
                size="small"
              >
                {{ bizTypeLabel(item.biz_type) }}
              </van-tag>
              <text class="meta-text">{{ formatBudget(item.budget_fen) }}</text>
              <text v-if="item.district" class="meta-text">{{ item.district }}</text>
              <text class="meta-date">{{ formatDate(item.created_at) }}</text>
            </view>
          </template>
        </van-cell>
      </van-cell-group>

      <!-- Load more -->
      <view v-if="list.length > 0" class="load-more">
        <van-loading v-if="loadingMore" size="20">加载更多...</van-loading>
        <text v-else-if="!hasMore" class="no-more">没有更多了</text>
      </view>
    </view>

    <!-- Sort action sheet -->
    <van-action-sheet
      :show="sortPickerVisible"
      :actions="sortActions"
      cancel-text="取消"
      @select="onSortSelect"
      @close="sortPickerVisible = false"
      @cancel="sortPickerVisible = false"
    />
  </view>
</template>

<script>
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
      bizTypeTabs: [
        { label: '全部', value: '' },
        { label: '巡检', value: 'cable_inspection' },
        { label: '植保', value: 'plant_transport' },
        { label: '农药', value: 'spray_pesticide' },
        { label: '租赁', value: 'trade_lease' },
        { label: '清洗', value: 'clean_paint' },
        { label: '其他', value: 'other' },
      ],
      sortActions: [
        { name: '最新发布', value: 'newest' },
        { name: '预算最高', value: 'budget_desc' },
        { name: '预算最低', value: 'budget_asc' },
      ],
    }
  },
  computed: {
    currentSortLabel() {
      const found = this.sortActions.find((a) => a.value === this.currentSort)
      return found ? found.name : '最新发布'
    },
  },
  onLoad() {
    this.fetchList(true)
  },
  onPullDownRefresh() {
    this.fetchList(true).then(() => {
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
      }
      if (reset) {
        this.loading = true
      } else {
        this.loadingMore = true
      }
      this.errorMsg = ''

      const accessToken = uni.getStorageSync('accessToken') || ''
      const params = []
      if (this.activeBizType) params.push('biz_type=' + encodeURIComponent(this.activeBizType))
      if (this.currentSort) params.push('sort=' + encodeURIComponent(this.currentSort))
      if (this.searchText) params.push('q=' + encodeURIComponent(this.searchText))
      params.push('page=' + this.page)
      params.push('page_size=' + this.pageSize)

      const url = 'http://localhost:8080/api/v1/demands?' + params.join('&')
      try {
        const [respErr, resp] = await uniRequest(url, {
          header: accessToken ? { Authorization: 'Bearer ' + accessToken } : {},
        })
        if (respErr) {
          this.errorMsg = respErr.message || '加载失败'
          return
        }
        const data = resp.data
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
    onSortSelect(e) {
      this.currentSort = e.detail.value
      this.sortPickerVisible = false
      this.fetchList(true)
    },
    goDetail(item) {
      uni.navigateTo({ url: '/pages/demands/detail?id=' + encodeURIComponent(item.id) })
    },
    bizTypeLabel(type) {
      const map = {
        cable_inspection: '巡检',
        plant_transport: '植保',
        spray_pesticide: '农药',
        trade_lease: '租赁',
        clean_paint: '清洗',
        other: '其他',
      }
      return map[type] || type || '其他'
    },
    bizTypeTagType(type) {
      const map = {
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
      const yuan = (fen / 100).toFixed(2)
      return yuan.replace(/\.00$/, '') + ' 元'
    },
    formatDate(iso) {
      if (!iso) return ''
      return iso.slice(0, 10)
    },
  },
}

function uniRequest(url, options) {
  return new Promise((resolve) => {
    uni.request({
      url,
      method: 'GET',
      header: options.header || {},
      success: (res) => {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve([null, res.data])
        } else {
          const msg =
            (res.data && res.data.error && res.data.error.message) ||
            '请求失败 (' + res.statusCode + ')'
          resolve([new Error(msg), null])
        }
      },
      fail: (err) => {
        resolve([err || new Error('网络异常'), null])
      },
    })
  })
}
</script>

<style scoped>
.demand-list-page {
  min-height: 100vh;
  background: #f7f8fa;
  padding-bottom: env(safe-area-inset-bottom);
}

.filter-tabs {
  display: flex;
  padding: 10px 12px;
  gap: 8px;
  background: #fff;
  overflow-x: auto;
  white-space: nowrap;
}

.filter-tab {
  flex-shrink: 0;
  padding: 6px 16px;
  border-radius: 20px;
  font-size: 13px;
  color: #646566;
  background: #f7f8fa;
}

.filter-tab.active {
  color: #fff;
  background: #1989fa;
}

.sort-bar {
  padding: 8px 16px;
  display: flex;
  justify-content: flex-end;
  background: #fff;
  border-bottom: 1px solid #f2f3f5;
}

.sort-trigger {
  display: flex;
  align-items: center;
  gap: 4px;
}

.sort-label {
  font-size: 13px;
  color: #969799;
}

.loading-state,
.empty-state,
.error-state {
  padding-top: 120px;
  text-align: center;
}

.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
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

.meta-date {
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
