<template>
  <view class="page">
    <!-- ========== Hero ========== -->
    <view class="hero">
      <view class="hero-row">
        <view class="hero-text">
          <text class="hero-eyebrow">POLICY UPDATE</text>
          <text class="hero-title">政策动态，边界清晰</text>
          <text class="hero-desc">聚合低空、通航与无人机政策，厘清不同业态适用边界。</text>
        </view>
        <view class="hero-icon-box">
          <image :src="heroIconSrc" class="hero-icon-img" mode="aspectFit" />
        </view>
      </view>
      <view class="hero-search">
        <image :src="searchIconSrc" class="search-icon" mode="aspectFit" />
        <input
          v-model="searchText"
          class="search-input"
          placeholder="搜索已加载的政策资讯"
          placeholder-style="color:#98A2B3"
          @input="onSearchInput"
        />
      </view>
    </view>

    <!-- ========== Module tabs ========== -->
    <view class="module-tabs">
      <view
        v-for="tab in moduleTabs"
        :key="tab.key"
        class="module-tab"
        :class="{ active: activeModule === tab.key }"
        @tap="onModuleTab(tab.key)"
      >
        <image :src="tabIcon(tab.key)" class="tab-icon-img" mode="aspectFit" />
        <text class="tab-label">{{ tab.label }}</text>
      </view>
    </view>

    <!-- ========== Category chips ========== -->
    <scroll-view scroll-x class="chips-scroll" :show-scrollbar="false">
      <view class="chips">
        <view
          v-for="chip in categoryChips"
          :key="chip.value"
          class="chip"
          :class="{ active: activeCategory === chip.value }"
          @tap="onCategoryChange(chip.value)"
        >
          {{ chip.label }}
        </view>
      </view>
    </scroll-view>

    <!-- ========== Section head ========== -->
    <view class="section-head">
      <text class="section-title">最新资讯</text>
      <text class="section-count">{{ filteredList.length }} 条</text>
    </view>

    <!-- ========== Loading state ========== -->
    <view v-if="loading && list.length === 0" class="state-view">
      <view class="loading-inline">
        <u-loading size="24rpx" color="#0A66C2" />
        <text>加载中...</text>
      </view>
    </view>

    <!-- ========== Error state ========== -->
    <view v-else-if="errorMsg && list.length === 0" class="state-view">
      <view class="empty-icon-wrap">
        <image :src="emptyIconSrc" class="empty-icon-img" mode="aspectFit" />
      </view>
      <text class="empty-title">加载失败</text>
      <text class="empty-desc">{{ errorMsg }}</text>
      <view class="retry-btn" @tap="fetchList(true)">重新加载</view>
    </view>

    <!-- ========== Empty state ========== -->
    <view v-else-if="!loading && filteredList.length === 0 && list.length === 0" class="state-view">
      <view class="empty-icon-wrap">
        <image :src="folderIconSrc" class="empty-icon-img" mode="aspectFit" />
      </view>
      <text class="empty-title">暂无政策资讯</text>
      <text class="empty-desc">调整关键词或分类后重试</text>
    </view>

    <!-- ========== Empty search result ========== -->
    <view v-else-if="!loading && filteredList.length === 0 && list.length > 0" class="state-view">
      <view class="empty-icon-wrap">
        <image :src="folderIconSrc" class="empty-icon-img" mode="aspectFit" />
      </view>
      <text class="empty-title">暂无匹配资讯</text>
      <text class="empty-desc">调整关键词或分类后重试</text>
    </view>

    <!-- ========== Data state: article list ========== -->
    <view v-else class="list-body">
      <view
        v-for="item in filteredList"
        :key="item.id"
        class="card"
        @tap="openDetail(item)"
      >
        <view class="card-top">
          <view class="card-badges">
            <text class="badge" :class="categoryBadgeClass(item.category)">
              {{ categoryLabel(item.category) }}
            </text>
          </view>
          <text v-if="item.is_pinned" class="badge badge-pin">置顶</text>
        </view>
        <text class="card-title">{{ item.title }}</text>
        <text class="card-summary">{{ item.summary || '' }}</text>
        <view class="card-meta">
          <text class="meta-source">{{ item.source || '' }}</text>
          <text class="meta-date">{{ formatDate(item.created_at) }}</text>
          <text class="meta-arrow">›</text>
        </view>
      </view>

      <!-- Load more -->
      <view v-if="showLoadMore" class="load-more">
        <view v-if="loadingMore" class="loading-inline">
          <u-loading size="20rpx" color="#0A66C2" />
          <text>加载更多...</text>
        </view>
        <text v-else-if="!hasMore" class="no-more">- 没有更多了 -</text>
        <view v-else class="load-more-btn" @tap="loadMore">点击加载更多</view>
      </view>
    </view>

    <!-- ========== Error banner when data exists ========== -->
    <view v-if="errorMsg && list.length > 0" class="error-banner">
      <text>{{ errorMsg }}</text>
      <text class="error-retry" @tap="fetchList(true)">重试</text>
    </view>

    <!-- ========== Detail bottom sheet ========== -->
    <u-popup
      :show="sheetVisible"
      position="bottom"
      round
      @close="closeSheet"
    >
      <view v-if="selectedItem" class="sheet-body">
        <view class="sheet-handle"></view>
        <view class="sheet-head">
          <text class="sheet-title">{{ selectedItem.title }}</text>
          <view class="sheet-close" @tap="closeSheet">
            <image :src="xIconSrc" class="sheet-close-icon" mode="aspectFit" />
          </view>
        </view>
        <scroll-view scroll-y class="sheet-scroll">
          <view class="detail-meta">
            <text class="badge" :class="categoryBadgeClass(selectedItem.category)">
              {{ categoryLabel(selectedItem.category) }}
            </text>
            <text class="detail-text">{{ selectedItem.source || '' }}</text>
            <text class="detail-text">{{ selectedItem.author || '' }}</text>
            <text class="detail-text">{{ formatDate(selectedItem.created_at) }}</text>
          </view>
          <text class="detail-content">{{ selectedItem.content || selectedItem.summary || '暂无详细内容' }}</text>
        </scroll-view>
      </view>
    </u-popup>
  </view>
</template>

<script>
import { request } from '../../utils/request'

const MODULE_TABS = [
  { key: 'news', label: '政策' },
  { key: 'knowledge', label: '知识' },
  { key: 'standards', label: '标准' },
  { key: 'applications', label: '申报' },
  { key: 'sentiment', label: '舆情' },
]

const CATEGORY_CHIPS = [
  { label: '全部', value: '' },
  { label: '低空经济', value: 'low_altitude_policy' },
  { label: '无人机法规', value: 'uav_regulation' },
  { label: '空域管理', value: 'airspace_management' },
  { label: '补贴政策', value: 'subsidy_policy' },
  { label: '行业标准', value: 'industry_standard' },
]

const CATEGORY_MAP = {
  low_altitude_policy: '低空经济',
  uav_regulation: '无人机法规',
  airspace_management: '空域管理',
  subsidy_policy: '补贴政策',
  industry_standard: '行业标准',
}

const PAGE_SIZE = 10

/* ---------- inline SVG helpers ---------- */
function svgUri(pathD, color) {
  return 'data:image/svg+xml,' + encodeURIComponent(
    '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#' +
    color +
    '" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="' +
    pathD +
    '"/></svg>'
  )
}

const TAB_ICONS_INACTIVE = {
  news: svgUri('M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z M14 2v6h6 M16 13H8 M16 17H8', '718096'),
  knowledge: svgUri('M4 19.5A2.5 2.5 0 0 1 6.5 17H20 M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2Z', '718096'),
  standards: svgUri('M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z', '718096'),
  applications: svgUri('M16 21V5a2 2 0 0 0-2-2H10a2 2 0 0 0-2 2v16 M2 21h20 M8 10h8', '718096'),
  sentiment: svgUri('M22 12h-4l-3 9L9 3l-3 9H2', '718096'),
}

const TAB_ICONS_ACTIVE = {
  news: svgUri('M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z M14 2v6h6 M16 13H8 M16 17H8', '0A66C2'),
  knowledge: svgUri('M4 19.5A2.5 2.5 0 0 1 6.5 17H20 M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2Z', '0A66C2'),
  standards: svgUri('M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z', '0A66C2'),
  applications: svgUri('M16 21V5a2 2 0 0 0-2-2H10a2 2 0 0 0-2 2v16 M2 21h20 M8 10h8', '0A66C2'),
  sentiment: svgUri('M22 12h-4l-3 9L9 3l-3 9H2', '0A66C2'),
}

const ICON_HERO_NEWS = svgUri('M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z M14 2v6h6 M16 13H8 M16 17H8', 'ffffff')
const ICON_SEARCH = svgUri('M11 19a8 8 0 1 0 0-16 8 8 0 0 0 0 16z M21 21l-4.35-4.35', '98A2B3')
const ICON_EMPTY = svgUri('M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z', 'c0cad8')
const ICON_X = svgUri('M18 6 6 18 M6 6l12 12', '101828')

export default {
  data() {
    return {
      searchText: '',
      activeModule: 'news',
      activeCategory: '',
      loading: false,
      loadingMore: false,
      errorMsg: '',
      list: [],
      page: 1,
      pageSize: PAGE_SIZE,
      hasMore: true,
      requestId: 0,
      sheetVisible: false,
      selectedItem: null,

      moduleTabs: MODULE_TABS,
      categoryChips: CATEGORY_CHIPS,
      heroIconSrc: ICON_HERO_NEWS,
      searchIconSrc: ICON_SEARCH,
      emptyIconSrc: ICON_EMPTY,
      folderIconSrc: ICON_EMPTY,
      xIconSrc: ICON_X,
    }
  },
  computed: {
    filteredList() {
      const q = this.searchText.trim().toLowerCase()
      if (!q) return this.list
      return this.list.filter(function (item) {
        const target = [item.title, item.summary, item.source, item.category]
          .filter(Boolean)
          .join(' ')
          .toLowerCase()
        return target.indexOf(q) !== -1
      })
    },
    showLoadMore() {
      return this.list.length > 0 && !this.searchText
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
    if (!this.loadingMore && this.hasMore && !this.searchText) {
      this.loadMore()
    }
  },
  methods: {
    async fetchList(reset) {
      if (reset) {
        this.page = 1
        this.hasMore = true
        this.loading = true
        this.requestId++
      } else {
        this.loadingMore = true
      }

      const reqId = this.requestId
      this.errorMsg = ''

      try {
        const params = {
          page: this.page,
          page_size: this.pageSize,
        }
        if (this.activeCategory) {
          params.category = this.activeCategory
        }

        const res = await request({
          url: '/api/v1/articles',
          data: params,
        })

        // Ignore stale responses when category changed during fetch
        if (reqId !== this.requestId) return

        const items = Array.isArray(res) ? res : []
        this.hasMore = items.length === this.pageSize

        if (reset) {
          this.list = items
        } else {
          this.list = this.list.concat(items)
        }
      } catch (e) {
        if (reqId !== this.requestId) return
        if (reset) {
          this.errorMsg = '网络异常，请稍后重试'
        } else {
          // On load-more error, just show a brief message, keep existing data
          uni.showToast({ title: '加载失败', icon: 'none', duration: 2000 })
          this.page--
        }
      } finally {
        if (reqId === this.requestId) {
          this.loading = false
          this.loadingMore = false
        }
      }
    },

    async loadMore() {
      this.page++
      await this.fetchList(false)
    },

    onSearchInput() {
      // Client-side filtering only — API does not support q/keyword
    },

    onCategoryChange(value) {
      this.activeCategory = value
      this.fetchList(true)
    },

    onModuleTab(key) {
      if (key === 'news') return
      if (key === 'knowledge') {
        uni.navigateTo({ url: '/pages/compliance/knowledge' })
      } else if (key === 'standards') {
        uni.navigateTo({ url: '/pages/compliance/standards' })
      } else {
        uni.showToast({ title: '功能开发中，敬请期待', icon: 'none', duration: 2000 })
      }
    },

    openDetail(item) {
      this.selectedItem = item
      this.sheetVisible = true
    },

    closeSheet() {
      this.sheetVisible = false
      this.selectedItem = null
    },

    tabIcon(key) {
      return key === this.activeModule ? TAB_ICONS_ACTIVE[key] : TAB_ICONS_INACTIVE[key]
    },

    categoryLabel(cat) {
      return CATEGORY_MAP[cat] || cat || '其他'
    },

    categoryBadgeClass(cat) {
      if (cat === 'uav_regulation') return 'badge-amber'
      if (cat === 'industry_standard') return 'badge-green'
      return 'badge-blue'
    },

    formatDate(iso) {
      if (!iso) return ''
      var d = new Date(iso)
      if (isNaN(d.getTime())) return iso
      var m = d.getMonth() + 1
      var day = d.getDate()
      return d.getFullYear() + '-' + (m < 10 ? '0' : '') + m + '-' + (day < 10 ? '0' : '') + day
    },
  },
}
</script>

<style scoped>
/* ========== Page ========== */
.page {
  min-height: 100vh;
  background-color: var(--color-bg);
  padding-bottom: calc(env(safe-area-inset-bottom) + 32rpx);
}

/* ========== Hero ========== */
.hero {
  min-height: 316rpx;
  padding: 40rpx 36rpx 32rpx;
  background: #071225;
  border-radius: 16rpx;
  margin: 28rpx 28rpx 0;
  position: relative;
  overflow: hidden;
}

.hero-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  position: relative;
  z-index: 1;
}

.hero-text {
  flex: 1;
  min-width: 0;
}

.hero-eyebrow {
  display: block;
  font-size: 22rpx;
  font-weight: 700;
  color: #9FBAFF;
  margin-bottom: 16rpx;
  letter-spacing: 2rpx;
}

.hero-title {
  display: block;
  font-size: 48rpx;
  font-weight: 760;
  color: #FFFFFF;
  line-height: 1.25;
}

.hero-desc {
  display: block;
  margin-top: 18rpx;
  max-width: 500rpx;
  font-size: 26rpx;
  color: #BFC9DC;
  line-height: 1.55;
}

.hero-icon-box {
  width: 88rpx;
  height: 88rpx;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(255, 255, 255, 0.25);
  background: rgba(255, 255, 255, 0.08);
  border-radius: 16rpx;
}

.hero-icon-img {
  width: 40rpx;
  height: 40rpx;
}

/* Hero search */
.hero-search {
  position: relative;
  z-index: 2;
  height: 88rpx;
  display: flex;
  align-items: center;
  margin-top: 36rpx;
  padding: 0 24rpx;
  background: #FFFFFF;
  border: 1px solid #E5EAF1;
  border-radius: 16rpx;
  box-shadow: 0 12rpx 36rpx rgba(16, 24, 40, 0.1);
}

.search-icon {
  width: 36rpx;
  height: 36rpx;
  flex-shrink: 0;
}

.search-input {
  flex: 1;
  min-width: 0;
  height: 100%;
  margin-left: 16rpx;
  font-size: 28rpx;
  color: var(--color-text);
  background: transparent;
}

/* ========== Module tabs ========== */
.module-tabs {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 8rpx;
  margin: 24rpx 28rpx 8rpx;
  padding: 10rpx;
  background: #E9EDF4;
  border: 1px solid #DFE5EE;
  border-radius: 16rpx;
}

.module-tab {
  min-height: 96rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6rpx;
  color: #718096;
  background: transparent;
  border-radius: 12rpx;
  font-size: 20rpx;
  font-weight: 680;
  transition: all 0.2s;
}

.module-tab.active {
  color: var(--color-primary);
  background: #FFFFFF;
  box-shadow: 0 6rpx 20rpx rgba(16, 24, 40, 0.08);
}

.tab-icon-img {
  width: 34rpx;
  height: 34rpx;
}

.tab-label {
  font-size: 20rpx;
  line-height: 1;
}

/* ========== Category chips ========== */
.chips-scroll {
  white-space: nowrap;
  margin: 28rpx -28rpx 0;
  padding: 0 28rpx;
}

.chips {
  display: flex;
  gap: 16rpx;
}

.chip {
  flex-shrink: 0;
  min-height: 72rpx;
  padding: 0 30rpx;
  display: flex;
  align-items: center;
  color: #66758E;
  background: var(--color-bg-card);
  border: 1px solid #E5EAF1;
  border-radius: 36rpx;
  font-size: 26rpx;
  font-weight: 650;
  transition: all 0.2s;
}

.chip.active {
  color: #FFFFFF;
  background: var(--color-primary);
  border-color: var(--color-primary);
  box-shadow: 0 10rpx 28rpx rgba(10, 102, 194, 0.25);
}

/* ========== Section head ========== */
.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 36rpx 32rpx 20rpx;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 16rpx;
  font-size: 34rpx;
  font-weight: 760;
  color: var(--color-text);
}

.section-title::before {
  content: '';
  width: 8rpx;
  height: 36rpx;
  background: var(--color-primary);
  border-radius: 4rpx;
  flex-shrink: 0;
}

.section-count {
  font-size: 24rpx;
  color: #98A2B3;
}

/* ========== States ========== */
.state-view {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 120rpx 56rpx 40rpx;
  text-align: center;
}

.empty-icon-wrap {
  width: 80rpx;
  height: 80rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24rpx;
}

.empty-icon-img {
  width: 72rpx;
  height: 72rpx;
}

.empty-title {
  display: block;
  font-size: 30rpx;
  font-weight: 650;
  color: var(--color-text);
}

.empty-desc {
  display: block;
  margin-top: 12rpx;
  font-size: 24rpx;
  color: #98A2B3;
}

.retry-btn {
  margin-top: 32rpx;
  padding: 18rpx 48rpx;
  color: #FFFFFF;
  background: var(--color-primary);
  border-radius: 16rpx;
  font-size: 28rpx;
  font-weight: 650;
}

/* ========== Error banner ========== */
.error-banner {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16rpx;
  padding: 16rpx 28rpx;
  margin: 0 28rpx;
  font-size: 24rpx;
  color: #B42318;
  background: #FEF0EF;
  border-radius: 16rpx;
}

.error-retry {
  color: var(--color-primary);
  font-weight: 650;
}

/* ========== Card list ========== */
.list-body {
  padding: 0 28rpx;
}

.card {
  padding: 32rpx;
  margin-bottom: 20rpx;
  background: var(--color-bg-card);
  border: 1px solid #E5EAF1;
  border-radius: 16rpx;
  box-shadow: 0 10rpx 36rpx rgba(16, 24, 40, 0.05);
}

.card:active {
  transform: translateY(2rpx);
}

.card-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16rpx;
}

.card-badges {
  display: flex;
  gap: 12rpx;
  flex-wrap: wrap;
}

.badge {
  display: inline-flex;
  align-items: center;
  min-height: 48rpx;
  padding: 0 18rpx;
  border-radius: 8rpx;
  font-size: 22rpx;
  font-weight: 700;
}

.badge-blue {
  color: var(--color-primary);
  background: #EAF1FF;
}

.badge-amber {
  color: #A76513;
  background: #FFF4DF;
}

.badge-green {
  color: #178A63;
  background: #EAF7F2;
}

.badge-pin {
  color: #FFFFFF;
  background: var(--color-primary);
}

.card-title {
  display: block;
  margin-top: 20rpx;
  font-size: 32rpx;
  font-weight: 740;
  color: var(--color-text);
  line-height: 1.45;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-summary {
  display: block;
  margin-top: 16rpx;
  font-size: 26rpx;
  color: #66758E;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-meta {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin-top: 24rpx;
  padding-top: 22rpx;
  border-top: 1px solid #EEF1F5;
  font-size: 22rpx;
  color: #98A2B3;
}

.meta-source {
  color: #98A2B3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 200rpx;
}

.meta-date {
  color: #98A2B3;
  flex-shrink: 0;
}

.meta-arrow {
  margin-left: auto;
  font-size: 36rpx;
  color: var(--color-primary);
  font-weight: 300;
}

/* ========== Load more ========== */
.load-more {
  display: flex;
  justify-content: center;
  padding: 32rpx 0;
}

.no-more {
  font-size: 24rpx;
  color: #C8C9CC;
}

.load-more-btn {
  font-size: 26rpx;
  color: var(--color-primary);
  font-weight: 650;
}

/* ========== Sheet ========== */
.sheet-body {
  display: flex;
  flex-direction: column;
}

.sheet-handle {
  width: 76rpx;
  height: 8rpx;
  margin: 18rpx auto 6rpx;
  background: #D0D5DD;
  border-radius: 4rpx;
}

.sheet-head {
  display: flex;
  align-items: flex-start;
  gap: 24rpx;
  padding: 28rpx 36rpx 26rpx;
  border-bottom: 1px solid #E5EAF1;
}

.sheet-title {
  flex: 1;
  font-size: 38rpx;
  font-weight: 740;
  color: var(--color-text);
  line-height: 1.4;
}

.sheet-close {
  width: 60rpx;
  height: 60rpx;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 16rpx;
}

.sheet-close:active {
  background: #F2F4F7;
}

.sheet-close-icon {
  width: 36rpx;
  height: 36rpx;
}

.sheet-scroll {
  max-height: calc(66vh - 108rpx);
  padding: 32rpx 36rpx 52rpx;
}

.detail-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 16rpx 24rpx;
}

.detail-text {
  font-size: 24rpx;
  color: #98A2B3;
}

.detail-content {
  display: block;
  margin-top: 34rpx;
  font-size: 28rpx;
  color: #344054;
  line-height: 1.8;
  white-space: pre-line;
}
</style>
