<template>
  <view class="search-page">
    <!-- Nav bar: back arrow only, no title -->
    <u-nav-bar show-back @back="goBack" />

    <!-- Search bar -->
    <view class="search-bar-wrap">
      <u-search
        v-model="searchText"
        placeholder="搜索需求、企业..."
        @search="onSearch"
        @change="onSearchInput"
      />
    </view>

    <!-- Tabs -->
    <u-sticky>
      <u-tabs
        :active="tabIndex"
        :titles="tabTitles"
        @change="onTabChange"
      />
    </u-sticky>

    <!-- ====== No search query: show history ====== -->
    <view v-if="!searchText" class="history-section">
      <view v-if="historyList.length > 0" class="history-header">
        <text class="history-title">搜索历史</text>
        <view class="history-clear" @tap="clearHistory">
          <text class="clear-text">清除</text>
        </view>
      </view>

      <view v-if="historyList.length > 0" class="history-tags">
        <u-tag
          v-for="(tag, index) in historyList"
          :key="index"
          type="default"
          class="history-tag"
          @tap="fillSearch(tag)"
        >
          {{ tag }}
        </u-tag>
      </view>

      <view v-else class="empty-state-wrapper">
        <u-empty description="暂无搜索历史" />
      </view>
    </view>

    <!-- ====== Has search query: show results ====== -->
    <view v-else class="results-section">
      <!-- Loading -->
      <view v-if="searchLoading" class="loading-state">
        <view class="loading-inline">
          <u-loading size="28rpx" />
          <text>搜索中...</text>
        </view>
      </view>

      <!-- Error -->
      <view v-else-if="searchError && results.length === 0" class="error-state">
        <u-empty description="搜索失败" />
        <view class="retry-btn" @tap="doSearch">
          <text>重新加载</text>
        </view>
      </view>

      <!-- Empty results -->
      <view v-else-if="!searchLoading && results.length === 0 && searched" class="empty-state-wrapper">
        <u-empty description="未找到相关内容" />
      </view>

      <!-- Result list -->
      <view v-else-if="results.length > 0" class="result-list">
        <!-- Demand results -->
        <u-cell-group v-if="activeTab === 'demand'" inset>
          <u-cell
            v-for="item in results"
            :key="item.id"
            is-link
            @click="goDemandDetail(item)"
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

        <!-- Enterprise results -->
        <u-cell-group v-else inset>
          <u-cell
            v-for="item in results"
            :key="item.id"
            is-link
            @click="goEnterpriseDetail(item)"
          >
            <template #icon>
              <view class="ent-icon-wrapper">
                <text class="ent-icon-text">企</text>
              </view>
            </template>
            <template #title>
              <view class="ent-title-row">
                <text class="ent-name">{{ item.name || item.enterprise_name }}</text>
                <text v-if="item.description || item.business_scope" class="ent-desc">{{ item.description || item.business_scope || '' }}</text>
              </view>
            </template>
          </u-cell>
        </u-cell-group>
      </view>
    </view>
  </view>
</template>

<script>
import { request } from '../../utils/request'

var HISTORY_KEY = 'searchHistory'
var MAX_HISTORY = 10

export default {
  data() {
    return {
      searchText: '',
      activeTab: 'demand',
      historyList: [],
      results: [],
      searchLoading: false,
      searchError: '',
      searched: false,
    }
  },
  computed: {
    // u-tabs 只接受字符串标题数组 + 数字 active 索引
    tabTitles() {
      return ['搜需求', '搜企业']
    },
    tabIndex() {
      return this.activeTab === 'enterprise' ? 1 : 0
    },
  },
  onLoad() {
    this.loadHistory()
  },
  methods: {
    // --- History ---
    loadHistory() {
      var self = this
      try {
        var raw = uni.getStorageSync(HISTORY_KEY)
        var arr = raw ? JSON.parse(raw) : []
        self.historyList = Array.isArray(arr) ? arr : []
      } catch (e) {
        self.historyList = []
      }
    },
    saveHistory() {
      try {
        uni.setStorageSync(HISTORY_KEY, JSON.stringify(this.historyList))
      } catch (e) {
        // ignore
      }
    },
    addToHistory(keyword) {
      if (!keyword || !keyword.trim()) return
      var kw = keyword.trim()
      var list = this.historyList.slice()
      // Remove duplicate
      var idx = list.indexOf(kw)
      if (idx !== -1) {
        list.splice(idx, 1)
      }
      // Add to front
      list.unshift(kw)
      // Cap at MAX_HISTORY
      if (list.length > MAX_HISTORY) {
        list = list.slice(0, MAX_HISTORY)
      }
      this.historyList = list
      this.saveHistory()
    },
    clearHistory() {
      this.historyList = []
      uni.removeStorageSync(HISTORY_KEY)
    },
    fillSearch(tag) {
      this.searchText = tag
      this.doSearch()
    },

    // --- Search ---
    onSearchInput(val) {
      this.searchText = val
      if (!this.searchText) {
        this.results = []
        this.searched = false
        this.searchError = ''
      }
    },
    onSearch() {
      if (this.searchText && this.searchText.trim()) {
        this.doSearch()
      }
    },
    async doSearch() {
      var kw = this.searchText && this.searchText.trim()
      if (!kw) {
        this.results = []
        this.searched = false
        return
      }

      this.searchLoading = true
      this.searchError = ''
      this.results = []

      try {
        var res = await request({
          url: '/api/v1/search',
          data: {
            q: kw,
            type: this.activeTab,
          },
        })
        var data = Array.isArray(res) ? res : (res && res.items) || []
        this.results = data
        this.searched = true
        // Save to history on successful search
        this.addToHistory(kw)
      } catch (e) {
        this.searchError = '网络异常，请稍后重试'
      } finally {
        this.searchLoading = false
      }
    },

    // --- Tabs ---
    onTabChange(index) {
      this.activeTab = index === 1 ? 'enterprise' : 'demand'
      this.results = []
      this.searched = false
      this.searchError = ''
      if (this.searchText && this.searchText.trim()) {
        this.doSearch()
      }
    },

    // --- Navigation ---
    goBack() {
      uni.navigateBack()
    },
    goDemandDetail(item) {
      uni.navigateTo({ url: '/pages/demands/detail?id=' + encodeURIComponent(item.id) })
    },
    goEnterpriseDetail(item) {
      uni.navigateTo({ url: '/pages/enterprise/status?id=' + encodeURIComponent(item.id) })
    },

    // --- Helpers ---
    bizTypeLabel(type) {
      var map = {
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
.search-page {
  min-height: 100vh;
  background: var(--color-bg);
  padding-bottom: env(safe-area-inset-bottom);
}

/* Search bar wrap */
.search-bar-wrap {
  background: #fff;
}

/* States */
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

.empty-state-wrapper {
  padding-top: 60px;
}

/* History section */
.history-section {
  padding: 16px;
}

.history-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.history-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text);
}

.clear-text {
  font-size: 13px;
  color: var(--color-text-secondary);
}

.history-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.history-tags .history-tag {
  padding: 4px 12px;
  border-radius: 4px;
  font-size: 13px;
  background: #fff;
  color: var(--color-text);
  border: 1px solid var(--color-divider);
}

/* Results section */
.results-section {
  padding-top: 8px;
}

.result-list {
  padding: 12px 0 24px;
}

/* Cell content */
.cell-content {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
}

.cell-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text);
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
  color: var(--color-text-secondary);
}

.meta-date {
  font-size: 12px;
  color: var(--color-text-placeholder);
}

/* Enterprise icon */
.ent-icon-wrapper {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--color-primary-light);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 12px;
}

.ent-icon-text {
  font-size: 18px;
  font-weight: 600;
  color: var(--color-primary);
}

.ent-title-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
}

.ent-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text);
}

.ent-desc {
  font-size: 12px;
  color: var(--color-text-secondary);
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
</style>
