<template>
  <view class="search-page">
    <!-- ═══════ 深蓝顶部：返回 + 标题 + 搜索框（对齐需求大厅） ═══════ -->
    <view class="topbar" :style="{ paddingTop: (statusBarH + 4) + 'px' }">
      <view class="topbar-row">
        <view class="back-btn" hover-class="tap-fade" hover-stay-time="120" @tap="goBack">
          <view class="back-arrow"></view>
        </view>
        <text class="top-title">全局搜索</text>
        <view class="topbar-spacer"></view>
      </view>
      <view class="search-box">
        <view class="search-icon"></view>
        <input
          class="search-input"
          v-model="searchText"
          placeholder="搜索需求、企业..."
          placeholder-class="search-ph"
          confirm-type="search"
          @input="onSearchInput($event.detail.value)"
          @confirm="onSearch"
        />
      </view>
    </view>

    <!-- ═══════ 白底下划线 Tabs ═══════ -->
    <view class="tabs">
      <view
        class="tab"
        :class="{ active: activeTab === 'demand' }"
        @tap="onTabChange(0)"
      >搜需求</view>
      <view
        class="tab"
        :class="{ active: activeTab === 'enterprise' }"
        @tap="onTabChange(1)"
      >搜企业</view>
    </view>

    <!-- ====== No search query: show history ====== -->
    <view v-if="!searchText" class="history-section">
      <view v-if="historyList.length > 0" class="history-header">
        <text class="history-title">搜索历史</text>
        <view class="history-clear" @tap="clearHistory">
          <text class="history-clear">清除</text>
        </view>
      </view>

      <view v-if="historyList.length > 0" class="history-tags">
        <text
          v-for="(tag, index) in historyList"
          :key="index"
          class="history-tag"
          @tap="fillSearch(tag)"
        >{{ tag }}</text>
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

      <!-- Result list（紧凑白卡，对齐需求大厅卡片） -->
      <view v-else-if="results.length > 0" class="result-list">
        <!-- Demand results -->
        <template v-if="activeTab === 'demand'">
          <view
            v-for="item in results"
            :key="item.id"
            class="result-card"
            hover-class="tap-fade"
            hover-stay-time="120"
            @tap="goDemandDetail(item)"
          >
            <text class="result-title">{{ item.title }}</text>
            <view class="result-meta">
              <text class="meta-tag" :class="'meta-tag--' + bizTypeTagType(item.biz_type)">{{ bizTypeLabel(item.biz_type) }}</text>
              <text v-if="item.district" class="meta-text">{{ item.district }}</text>
              <text class="meta-text">{{ formatBudget(item.budget_fen) }}</text>
              <text class="meta-date">{{ formatDate(item.created_at) }}</text>
            </view>
          </view>
        </template>

        <!-- Enterprise results -->
        <template v-else>
          <view
            v-for="item in results"
            :key="item.id"
            class="result-card ent-card"
            hover-class="tap-fade"
            hover-stay-time="120"
            @tap="goEnterpriseDetail(item)"
          >
            <view class="ent-icon">
              <text class="ent-icon-text">企</text>
            </view>
            <view class="ent-title-row">
              <text class="ent-name">{{ item.name || item.enterprise_name }}</text>
              <text v-if="item.description || item.business_scope" class="ent-desc">{{ item.description || item.business_scope || '' }}</text>
            </view>
          </view>
        </template>
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
      statusBarH: 20, // 状态栏高度（custom 导航避让，动态取系统信息）
      historyList: [],
      results: [],
      searchLoading: false,
      searchError: '',
      searched: false,
    }
  },
  onLoad() {
    this.loadHistory()
    try {
      this.statusBarH = uni.getSystemInfoSync().statusBarHeight || 20
    } catch (e) { /* 默认 20 */ }
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
      uni.navigateTo({ url: '/pkg-eco/pages/enterprise/status?id=' + encodeURIComponent(item.id) })
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

/* ═══════ 深蓝顶部（对齐需求大厅） ═══════ */
.topbar {
  background: #074D92;
  color: #fff;
  padding: 16rpx 24rpx 28rpx;
  /* 顶部避让由 :style 动态 paddingTop 控制（对齐首页 headerPadTop 方案） */
}
.topbar-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
}
.back-btn {
  width: 60rpx;
  height: 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}
.back-arrow {
  width: 20rpx;
  height: 20rpx;
  border-left: 4rpx solid #fff;
  border-bottom: 4rpx solid #fff;
  transform: rotate(45deg);
  margin-left: 10rpx;
}
.top-title {
  flex: 1;
  font-size: 38rpx;
  font-weight: 700;
  text-align: center;
}
.topbar-spacer { width: 60rpx; }

/* 搜索框：白底圆角，对齐需求大厅 search-trigger */
.search-box {
  width: 100%;
  height: 44px;
  margin-top: 24rpx;
  border-radius: 7px;
  background: #fff;
  display: flex;
  align-items: center;
  gap: 14rpx;
  padding: 0 24rpx;
  box-sizing: border-box;
}
.search-icon {
  width: 30rpx;
  height: 30rpx;
  border: 4rpx solid #98A2B3;
  border-radius: 50%;
  position: relative;
  flex-shrink: 0;
}
.search-icon::after {
  content: '';
  position: absolute;
  right: -12rpx;
  bottom: -7rpx;
  width: 14rpx;
  height: 4rpx;
  border-radius: 4rpx;
  background: #98A2B3;
  transform: rotate(45deg);
}
.search-input { flex: 1; font-size: 28rpx; color: #17212B; }
.search-ph { color: #98A2B3; }

/* ═══════ 白底下划线 Tabs ═══════ */
.tabs {
  display: flex;
  background: #fff;
  border-bottom: 1px solid #EEF1F4;
  padding: 0 32rpx;
}
.tab {
  flex: 1;
  height: 92rpx;
  line-height: 92rpx;
  text-align: center;
  position: relative;
  color: #667085;
  font-weight: 600;
  font-size: 28rpx;
}
.tab.active { color: #0A66C2; }
.tab.active::after {
  content: '';
  position: absolute;
  width: 56rpx;
  height: 6rpx;
  background: #0A66C2;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  border-radius: 3rpx;
}

/* History section */
.history-section { padding: 24rpx; }
.history-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20rpx;
}
.history-title {
  font-size: 28rpx;
  font-weight: 600;
  color: var(--color-text);
}
.history-clear {
  font-size: 24rpx;
  color: var(--color-text-secondary);
}
.history-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 20rpx;
}
.history-tag {
  padding: 10rpx 28rpx;
  background: #fff;
  border: 1rpx solid var(--color-border);
  border-radius: 8rpx;
  font-size: 26rpx;
  color: #344054;
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
.empty-state-wrapper { padding-top: 60px; }

/* Results section：紧凑白卡（对齐需求大厅卡片） */
.results-section { padding-top: 8px; }
.result-list { padding: 20rpx 24rpx 24rpx; }
.result-card {
  background: #fff;
  border: 1rpx solid var(--color-border);
  border-radius: 16rpx;
  box-shadow: var(--shadow-sm);
  padding: 24rpx;
  margin-bottom: 16rpx;
}
.result-title {
  display: block;
  font-size: 28rpx;
  font-weight: 700;
  color: var(--color-text);
  line-height: 1.4;
}
.result-meta {
  display: flex;
  align-items: center;
  gap: 16rpx;
  flex-wrap: wrap;
  margin-top: 14rpx;
}
.meta-tag {
  padding: 4rpx 14rpx;
  border-radius: 8rpx;
  font-size: 20rpx;
  font-weight: 600;
}
.meta-tag--primary { background: var(--color-primary-light); color: var(--color-primary); }
.meta-tag--success { background: #E8F7EF; color: #16A34A; }
.meta-tag--warning { background: #FDF1E7; color: #E46426; }
.meta-tag--danger { background: #FDECEC; color: #E84C3D; }
.meta-tag--default { background: #F0F3F6; color: #667085; }
.meta-text { font-size: 24rpx; color: var(--color-text-secondary); }
.meta-date { font-size: 24rpx; color: var(--color-text-placeholder); }

/* Enterprise result */
.ent-card { display: flex; align-items: center; gap: 20rpx; }
.ent-icon {
  width: 72rpx;
  height: 72rpx;
  flex-shrink: 0;
  border-radius: 50%;
  background: var(--color-primary-light);
  display: flex;
  align-items: center;
  justify-content: center;
}
.ent-icon-text { font-size: 30rpx; font-weight: 600; color: var(--color-primary); }
.ent-title-row {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}
.ent-name {
  font-size: 28rpx;
  font-weight: 700;
  color: var(--color-text);
}
.ent-desc {
  font-size: 24rpx;
  color: var(--color-text-secondary);
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

/* 按压反馈 */
.tap-fade { opacity: .8; }
</style>
