<template>
  <view class="resource-page">
    <view class="hero-card">
      <view class="hero-icon"><text class="hero-icon-text">资</text></view>
      <text class="hero-kicker">INDUSTRY ASSETS</text>
      <text class="hero-title">产业资源台账</text>
      <text class="hero-desc">集中查看无人机、机场、试飞场地与测试基地</text>
      <view class="hero-stats">
        <view>
          <text class="stat-value">{{ list.length }}</text>
          <text class="stat-label">已加载资源</text>
        </view>
        <view>
          <text class="stat-value">{{ activeType ? '1' : '4' }}</text>
          <text class="stat-label">资源类别</text>
        </view>
      </view>
    </view>

    <view class="search-card">
      <u-search
        v-model="searchText"
        placeholder="搜索资源名称、型号或位置"
      />
    </view>

    <scroll-view class="filter-scroll" scroll-x :show-scrollbar="false">
      <view class="filter-row">
        <view
          v-for="item in typeTabs"
          :key="item.value"
          class="filter-item"
          :class="{ active: activeType === item.value }"
          @tap="switchType(item.value)"
        >
          {{ item.label }}
        </view>
      </view>
    </scroll-view>

    <view v-if="loading && !list.length" class="state-panel">
      <view class="loading-inline">
        <u-loading size="26rpx" color="#0A66C2" />
        <text>正在加载产业资源</text>
      </view>
    </view>

    <view v-else-if="errorMsg && !list.length" class="state-panel">
      <u-empty description="产业资源加载失败" />
      <button class="retry-btn" @tap="fetchList(true)">重新加载</button>
    </view>

    <view v-else-if="!displayList.length" class="state-panel compact">
      <u-empty description="暂无匹配的产业资源" />
    </view>

    <view v-else class="resource-list">
      <view v-for="item in displayList" :key="item.id" class="resource-card" @tap="goDetail(item)">
        <view class="resource-icon">
          <text class="resource-icon-text">{{ resourceIcon(item.res_type) }}</text>
        </view>
        <view class="resource-main">
          <view class="title-row">
            <text class="resource-name text-ellipsis">{{ item.name || '未命名资源' }}</text>
            <text class="type-tag">{{ typeLabel(item.res_type) }}</text>
            <text v-if="item.status && item.status !== 'available'" class="status-tag" :class="'status-' + item.status">{{ statusLabel(item.status) }}</text>
          </view>
          <text class="resource-model text-ellipsis">{{ item.model || item.specs || '型号信息暂未填写' }}</text>
          <view class="meta-row">
            <text class="price">{{ formatPrice(item.price_fen) }}</text>
            <text class="location text-ellipsis">
              <u-icon name="location" size="26rpx" color="var(--color-text-secondary)" />
              {{ item.location || '位置待确认' }}
            </text>
          </view>
        </view>
        <text class="card-arrow">›</text>
      </view>

      <view class="load-more">
        <view v-if="loadingMore" class="loading-inline">
          <u-loading size="20rpx" color="#0A66C2" />
          <text>加载更多</text>
        </view>
        <text v-else-if="!hasMore" class="no-more">已加载全部资源</text>
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
      loading: true,
      loadingMore: false,
      errorMsg: '',
      list: [],
      page: 1,
      pageSize: 20,
      hasMore: true,
      typeTabs: [
        { label: '全部', value: '' },
        { label: '无人机', value: 'drone' },
        { label: '无人机机场', value: 'airport' },
        { label: '试飞场地', value: 'test_site' },
        { label: '测试基地', value: 'test_base' },
      ],
    }
  },
  computed: {
    displayList() {
      var keyword = this.searchText.trim().toLowerCase()
      if (!keyword) return this.list
      return this.list.filter(function (item) {
        return [item.name, item.model, item.specs, item.location]
          .filter(Boolean)
          .join(' ')
          .toLowerCase()
          .includes(keyword)
      })
    },
  },
  onLoad() {
    this.fetchList(true)
  },
  onPullDownRefresh() {
    this.fetchList(true).finally(function () {
      uni.stopPullDownRefresh()
    })
  },
  onReachBottom() {
    if (!this.loading && !this.loadingMore && this.hasMore) {
      this.page += 1
      this.fetchList(false)
    }
  },
  methods: {
    async fetchList(reset) {
      if (reset) {
        this.page = 1
        this.loading = true
      } else {
        this.loadingMore = true
      }
      this.errorMsg = ''
      try {
        var data = { page: this.page, page_size: this.pageSize }
        if (this.activeType) data.res_type = this.activeType
        var res = await request({ url: '/api/v1/industry-resources', data: data })
        var items = Array.isArray(res) ? res : (res && res.data) || []
        if (!Array.isArray(items)) items = []
        this.list = reset ? items : this.list.concat(items)
        this.hasMore = items.length === this.pageSize
      } catch (error) {
        if (!reset) this.page = Math.max(1, this.page - 1)
        if (reset) this.list = []
        this.errorMsg = '网络异常，请稍后重试'
        if (!reset) uni.showToast({ title: '加载更多失败', icon: 'none' })
      } finally {
        this.loading = false
        this.loadingMore = false
      }
    },
    switchType(value) {
      if (this.activeType === value) return
      this.activeType = value
      this.fetchList(true)
    },
    goDetail(item) {
      uni.setStorageSync('resource_detail_' + item.id, item)
      uni.navigateTo({ url: '/pages/resources/detail?id=' + encodeURIComponent(item.id) })
    },
    typeLabel(type) {
      var item = this.typeTabs.find(function (tab) { return tab.value === type })
      return item ? item.label : '产业资源'
    },
    statusLabel(status) {
      var map = { available: '可用', in_use: '使用中', maintenance: '维护中' }
      return map[status] || status || ''
    },
    resourceIcon(type) {
      var map = { drone: '机', airport: '场', test_site: '地', test_base: '基' }
      return map[type] || '源'
    },
    formatPrice(value) {
      var amount = Number(value || 0) / 100
      return amount ? '¥' + amount.toFixed(2) : '免费 / 面议'
    },
  },
}
</script>

<style scoped>
.resource-page {
  min-height: 100vh;
  box-sizing: border-box;
  padding: var(--space-md) var(--space-md) 60rpx;
  background: var(--color-bg);
}

.hero-card {
  position: relative;
  overflow: hidden;
  padding: 38rpx 36rpx 30rpx;
  color: #ffffff;
  background: #071225;
  border-radius: var(--radius-lg);
  box-shadow: inset -180rpx -100rpx 140rpx rgba(10, 102, 194, 0.4), var(--shadow-lg);
}

.hero-icon {
  position: absolute;
  top: 34rpx;
  right: 34rpx;
  width: 76rpx;
  height: 76rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2rpx solid rgba(255, 255, 255, 0.34);
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.08);
}

.hero-icon-text {
  font-size: 32rpx;
  font-weight: 600;
  color: #ffffff;
}

.hero-kicker,
.hero-title,
.hero-desc,
.stat-value,
.stat-label {
  display: block;
}

.hero-kicker {
  color: #91adff;
  font-size: var(--font-xs);
}

.hero-title {
  margin-top: 8rpx;
  font-size: var(--font-xxl);
  font-weight: 700;
}

.hero-desc {
  width: 76%;
  margin-top: 10rpx;
  color: #b8c4dc;
  font-size: var(--font-sm);
  line-height: 1.5;
}

.hero-stats {
  display: flex;
  gap: 90rpx;
  margin-top: 30rpx;
  padding-top: 24rpx;
  border-top: 2rpx solid rgba(255, 255, 255, 0.12);
}

.stat-value {
  font-size: var(--font-xl);
  font-weight: 700;
}

.stat-label {
  margin-top: 4rpx;
  color: #94a6c3;
  font-size: var(--font-xs);
}

.search-card {
  margin-top: var(--space-md);
  overflow: hidden;
  background: #ffffff;
  border: 2rpx solid #e3e8ef;
  border-radius: var(--radius-lg);
}

.filter-scroll {
  width: 100%;
  margin: 20rpx 0;
  white-space: nowrap;
}

.filter-row {
  display: inline-flex;
  gap: 12rpx;
  padding-right: 20rpx;
}

.filter-item {
  flex-shrink: 0;
  padding: 12rpx 26rpx;
  color: var(--color-text-secondary);
  font-size: var(--font-sm);
  background: #ffffff;
  border: 1rpx solid var(--color-border);
  border-radius: var(--radius-sm);
}

.filter-item.active {
  color: #ffffff;
  font-weight: 600;
  background: var(--color-primary);
  border-color: var(--color-primary);
}

.state-panel {
  min-height: 50vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.state-panel.compact {
  min-height: 38vh;
}

.retry-btn {
  min-width: 220rpx;
  height: 76rpx;
  margin: 8rpx 0 0;
  color: #ffffff;
  font-size: var(--font-sm);
  line-height: 76rpx;
  background: var(--color-primary);
  border-radius: var(--radius-round);
}

.retry-btn::after {
  border: 0;
}

.resource-list {
  display: flex;
  flex-direction: column;
  gap: 18rpx;
}

.resource-card {
  min-height: 164rpx;
  display: flex;
  align-items: center;
  gap: 22rpx;
  box-sizing: border-box;
  padding: 26rpx 24rpx;
  background: #ffffff;
  border: 2rpx solid #e3e8ef;
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
}

.resource-icon {
  width: 94rpx;
  height: 94rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: var(--color-primary-light);
  border-radius: var(--radius-md);
}

.resource-icon-text {
  font-size: 36rpx;
  font-weight: 600;
  color: var(--color-primary);
}

.resource-main {
  min-width: 0;
  flex: 1;
}

.title-row,
.meta-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.resource-name {
  max-width: 70%;
  color: var(--color-text);
  font-size: var(--font-lg);
  font-weight: 700;
}

.type-tag {
  flex-shrink: 0;
  padding: 4rpx 10rpx;
  color: var(--color-primary);
  font-size: var(--font-xs);
  background: var(--color-primary-light);
  border-radius: var(--radius-sm);
}
.status-tag {
  flex-shrink: 0;
  padding: 4rpx 10rpx;
  font-size: var(--font-xs);
  border-radius: var(--radius-sm);
}
.status-tag.status-in_use { color: var(--color-warning); background: var(--color-warning-light, #FFF4E6); }
.status-tag.status-maintenance { color: var(--color-text-secondary); background: var(--color-divider); }

.resource-model {
  display: block;
  margin-top: 8rpx;
  color: var(--color-text-secondary);
  font-size: var(--font-sm);
}

.meta-row {
  margin-top: 12rpx;
  justify-content: space-between;
}

.price {
  flex-shrink: 0;
  color: var(--color-primary);
  font-size: var(--font-sm);
  font-weight: 700;
}

.location {
  min-width: 0;
  color: var(--color-text-secondary);
  font-size: var(--font-xs);
}

.loading-inline {
  display: flex;
  align-items: center;
  gap: 8rpx;
  color: var(--color-text-secondary);
  font-size: var(--font-xs);
}

.card-arrow {
  font-size: 18px;
  color: var(--color-text-placeholder);
}

.load-more {
  min-height: 64rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.no-more {
  color: var(--color-text-placeholder);
  font-size: var(--font-xs);
}
</style>
