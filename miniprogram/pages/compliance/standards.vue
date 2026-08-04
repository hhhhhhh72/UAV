<template>
  <view class="standards-page">
    <u-nav-bar
      title="团体标准库"
      show-back
      @back="goBack"
    />

    <!-- Sticky search + tabs -->
    <u-sticky>
      <u-search
        v-model="searchText"
        placeholder="搜索团体标准"
        @search="onSearch"
      />
      <u-tabs
        :active="tabIndex"
        :titles="tabTitles"
        @change="onTabChange"
      />
    </u-sticky>

    <!-- Loading state -->
    <view v-if="loading && list.length === 0" class="state-view">
      <view class="loading-inline">
        <u-loading size="24rpx" />
        <text>加载中...</text>
      </view>
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg && list.length === 0" class="state-view">
      <u-empty description="加载失败" />
      <view class="retry-btn" @tap="fetchList(true)">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Empty state -->
    <view v-else-if="!loading && list.length === 0" class="state-view">
      <u-empty description="暂无团体标准" />
    </view>

    <!-- Normal: standards list -->
    <view v-else class="list-body">
      <u-cell-group inset>
        <u-cell
          v-for="item in list"
          :key="item.id"
          is-clickable
          @click="openStandard(item)"
        >
          <template #title>
            <view class="standard-title-row">
              <text class="standard-icon">文</text>
              <view class="standard-info">
                <text class="standard-title">{{ item.title || '--' }}</text>
                <view class="standard-meta">
                  <text v-if="item.version" class="meta-tag">v{{ item.version }}</text>
                  <text class="meta-date">{{ formatDate(item.publish_date || item.created_at) }}</text>
                </view>
              </view>
            </view>
          </template>
          <template #value>
            <text class="download-icon" @click.stop="downloadStandard(item)">↓</text>
          </template>
        </u-cell>
      </u-cell-group>

      <!-- Load more (if API supports pagination) -->
      <view v-if="list.length > 0" class="load-more">
        <text v-if="!hasMore" class="no-more">没有更多了</text>
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
      activeCategory: '',
      loading: false,
      errorMsg: '',
      list: [],
      page: 1,
      pageSize: 20,
      hasMore: true,
      categoryTabs: [
        { label: '全部', value: '' },
        { label: '低空', value: 'low_altitude' },
        { label: '无人机', value: 'drone' },
        { label: '通用', value: 'general' },
      ],
    }
  },
  computed: {
    tabTitles() {
      return this.categoryTabs.map(function (t) { return t.label })
    },
    tabIndex() {
      var idx = this.categoryTabs.findIndex(function (t) { return t.value === this.activeCategory }.bind(this))
      return idx >= 0 ? idx : 0
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
  methods: {
    async fetchList(reset) {
      if (reset) {
        this.page = 1
        this.hasMore = true
        this.loading = true
      } else {
        this.loading = true
      }
      this.errorMsg = ''

      try {
        var params = {
          page: this.page,
          page_size: this.pageSize,
        }
        if (this.activeCategory) params.category = this.activeCategory
        if (this.searchText) params.q = this.searchText

        var res = await request({
          url: '/api/v1/compliance-standards',
          data: params,
        })
        var data = Array.isArray(res) ? res : (res && res.items) || []
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
      }
    },
    onSearch() {
      this.fetchList(true)
    },
    onTabChange(index) {
      var tab = this.categoryTabs[index]
      this.activeCategory = tab ? tab.value : ''
      this.fetchList(true)
    },
    openStandard(item) {
      var content = item.content || item.description || item.standard_no || ''
      uni.showModal({
        title: item.title || '标准详情',
        content: content || '暂无详细内容',
        showCancel: false,
        confirmText: '知道了',
      })
    },
    downloadStandard(item) {
      if (item.file_url) {
        uni.downloadFile({
          url: item.file_url,
          success: function (res) {
            if (res.statusCode === 200) {
              uni.openDocument({
                filePath: res.tempFilePath,
                showMenu: true,
              })
            }
          },
          fail: function () {
            uni.showToast({ title: '下载失败', icon: 'none' })
          },
        })
      } else if (item.url) {
        uni.setClipboardData({
          data: item.url,
          success: function () {
            uni.showToast({ title: '链接已复制，请在浏览器打开下载', icon: 'none' })
          },
        })
      } else {
        uni.showToast({ title: '暂无下载资源', icon: 'none' })
      }
    },
    goBack() {
      uni.navigateBack()
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
.standards-page {
  min-height: 100vh;
  background: var(--color-bg);
  padding-bottom: env(safe-area-inset-bottom);
}

/* State views */
.state-view {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 120px;
}

.loading-inline {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: var(--color-text-secondary);
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

/* Standard item */
.standard-title-row {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.standard-icon {
  width: 40rpx;
  height: 40rpx;
  line-height: 40rpx;
  text-align: center;
  font-size: 24rpx;
  color: var(--color-primary);
  background: var(--color-primary-light);
  border-radius: 8rpx;
  flex-shrink: 0;
  margin-top: 2px;
}

.standard-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
  min-width: 0;
}

.standard-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text);
  line-height: 1.4;
}

.standard-meta {
  display: flex;
  align-items: center;
  gap: 10px;
}

.meta-tag {
  font-size: 11px;
  color: var(--color-primary);
  background: var(--color-primary-light);
  padding: 1px 8px;
  border-radius: 4px;
}

.meta-date {
  font-size: 11px;
  color: var(--color-text-placeholder);
}

.download-icon {
  width: 44rpx;
  height: 44rpx;
  line-height: 44rpx;
  text-align: center;
  font-size: 26rpx;
  color: var(--color-primary);
  background: var(--color-primary-light);
  border-radius: 50%;
  flex-shrink: 0;
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
</style>
