<template>
  <view class="search-page">
    <!-- Nav bar: back arrow only, no title -->
    <van-nav-bar left-arrow @click-left="goBack" />

    <!-- Search bar -->
    <view class="search-bar-wrap">
      <van-search
        :value="searchText"
        placeholder="搜索需求、企业..."
        shape="round"
        focus
        @search="onSearch"
        @input="onSearchInput"
      />
    </view>

    <!-- Tabs -->
    <van-tabs
      :active="activeTab"
      @change="onTabChange"
      :color="tabColor"
      :title-active-color="tabColor"
      sticky
    >
      <van-tab title="搜需求" name="demand" />
      <van-tab title="搜企业" name="enterprise" />
    </van-tabs>

    <!-- ====== No search query: show history ====== -->
    <view v-if="!searchText" class="history-section">
      <view v-if="historyList.length > 0" class="history-header">
        <text class="history-title">搜索历史</text>
        <view class="history-clear" @tap="clearHistory">
          <text class="clear-text">🗑 清除</text>
        </view>
      </view>

      <view v-if="historyList.length > 0" class="history-tags">
        <van-tag
          v-for="(tag, index) in historyList"
          :key="index"
          size="medium"
          type="default"
          class="history-tag"
          @tap="fillSearch(tag)"
        >
          {{ tag }}
        </van-tag>
      </view>

      <view v-else class="empty-state-wrapper">
        <van-empty description="暂无搜索历史" />
      </view>
    </view>

    <!-- ====== Has search query: show results ====== -->
    <view v-else class="results-section">
      <!-- Loading -->
      <view v-if="searchLoading" class="loading-state">
        <van-loading size="24">搜索中...</van-loading>
      </view>

      <!-- Error -->
      <view v-else-if="searchError && results.length === 0" class="error-state">
        <van-empty description="搜索失败" image="error" />
        <view class="retry-btn" @tap="doSearch">
          <text>重新加载</text>
        </view>
      </view>

      <!-- Empty results -->
      <view v-else-if="!searchLoading && results.length === 0 && searched" class="empty-state-wrapper">
        <van-empty description="未找到相关内容" image="search" />
      </view>

      <!-- Result list -->
      <view v-else-if="results.length > 0" class="result-list">
        <!-- Demand results -->
        <van-cell-group v-if="activeTab === 'demand'" inset>
          <van-cell
            v-for="item in results"
            :key="item.id"
            :title="item.title"
            is-link
            @tap="goDemandDetail(item)"
          >
            <template #label>
              <view class="cell-meta">
                <van-tag
                  :type="bizTypeTagType(item.biz_type)"
                  size="small"
                >
                  {{ bizTypeLabel(item.biz_type) }}
                </van-tag>
                <text v-if="item.district" class="meta-text">{{ item.district }}</text>
                <text class="meta-text">{{ formatBudget(item.budget_fen) }}</text>
                <text class="meta-date">{{ formatDate(item.created_at) }}</text>
              </view>
            </template>
          </van-cell>
        </van-cell-group>

        <!-- Enterprise results -->
        <van-cell-group v-else inset>
          <van-cell
            v-for="item in results"
            :key="item.id"
            :title="item.name || item.enterprise_name"
            :label="item.description || item.business_scope || ''"
            is-link
            @tap="goEnterpriseDetail(item)"
          >
            <template #icon>
              <view class="ent-icon-wrapper">
                <van-icon name="shop" size="20" color="#0A66C2" />
              </view>
            </template>
          </van-cell>
        </van-cell-group>
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
    tabColor() {
      return this.activeTab === 'demand' ? '#1565C0' : '#34c759'
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
    onSearchInput(e) {
      this.searchText = e.detail
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
    onTabChange(e) {
      this.activeTab = e.detail.name
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
  background: #f7f8fa;
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
  color: #1a1a1a;
}

.clear-text {
  font-size: 13px;
  color: #969799;
}

.history-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.history-tag {
  padding: 4px 12px;
  border-radius: 4px;
  font-size: 13px;
  background: #fff;
  color: #323233;
  border: 1px solid #ebedf0;
}

/* Results section */
.results-section {
  padding-top: 8px;
}

.result-list {
  padding: 12px 0 24px;
}

/* Cell meta */
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

/* Enterprise icon */
.ent-icon-wrapper {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: #e8f4fd;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 12px;
}
</style>
