<template>
  <view class="page-container">
    <!-- Nav -->
    <u-nav-bar
      title="行业报告"
      show-back
      @back="goBack"
    />

    <!-- Search -->
    <u-sticky>
      <u-search
        v-model="searchText"
        placeholder="搜索报告"
        @search="onSearch"
      />
    </u-sticky>

    <!-- Type tabs -->
    <u-tabs
      :active="tabIndex"
      :titles="tabTitles"
      @change="onTabChange"
    />

    <!-- Loading state -->
    <view v-if="loading && list.length === 0" class="loading-state">
      <view class="loading-inline">
        <u-loading size="24rpx" />
        <text>加载中...</text>
      </view>
    </view>

    <!-- Empty state -->
    <view v-else-if="!loading && list.length === 0 && !errorMsg" class="empty-state-wrapper">
      <u-empty description="暂无报告" />
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
        >
          <template #title>
            <view class="cell-content">
              <view class="cell-title-row">
                <text class="report-icon">文</text>
                <text class="cell-title">{{ item.title }}</text>
              </view>
              <view class="cell-meta">
                <u-tag :type="typeTagType(item.report_type || item.type)" size="mini">
                  {{ typeLabel(item.report_type || item.type) }}
                </u-tag>
                <text v-if="item.publish_date" class="meta-text">{{ item.publish_date }}</text>
              </view>
            </view>
          </template>
          <template #value>
            <view class="cell-action" @click.stop="downloadReport(item)">
              <text class="download-icon">↓</text>
            </view>
          </template>
        </u-cell>
      </u-cell-group>

      <!-- Load more -->
      <view v-if="list.length > 0" class="load-more">
        <view v-if="loadingMore" class="loading-inline">
          <u-loading size="20rpx" />
          <text>加载更多...</text>
        </view>
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
      activeType: '',
      loading: false,
      loadingMore: false,
      errorMsg: '',
      list: [],
      page: 1,
      pageSize: 20,
      hasMore: true,
      typeTabs: [
        { label: '全部', value: '' },
        { label: '白皮书', value: '白皮书' },
        { label: '调研报告', value: '调研报告' },
        { label: '年度报告', value: '年度报告' },
      ],
    }
  },
  computed: {
    tabTitles() {
      return this.typeTabs.map(function (t) { return t.label })
    },
    tabIndex() {
      var idx = this.typeTabs.findIndex(function (t) { return t.value === this.activeType }.bind(this))
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
        if (this.activeType) params.type = this.activeType
        if (this.searchText) params.q = this.searchText
        params.page = this.page
        params.page_size = this.pageSize

        var res = await request({ url: '/api/v1/industry-reports', data: params })
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
    onTabChange(index) {
      var tab = this.typeTabs[index]
      this.activeType = tab ? tab.value : ''
      this.fetchList(true)
    },

    downloadReport(item) {
      var fileUrl = item.file_url || item.download_url || item.url
      if (fileUrl) {
        // Attempt to download
        uni.downloadFile({
          url: fileUrl,
          success: function (res) {
            if (res.statusCode === 200) {
              uni.showToast({ title: '下载成功', icon: 'success' })
              // Open the downloaded file
              uni.openDocument({
                filePath: res.tempFilePath,
                showMenu: true,
              })
            } else {
              uni.showToast({ title: '下载失败', icon: 'none' })
            }
          },
          fail: function () {
            // Fallback: copy link
            uni.setClipboardData({
              data: fileUrl,
              success: function () {
                uni.showToast({ title: '链接已复制，请在浏览器打开', icon: 'none' })
              },
            })
          },
        })
      } else {
        uni.showToast({ title: '暂无下载链接', icon: 'none' })
      }
    },

    typeLabel(type) {
      var map = {
        '白皮书': '白皮书',
        '调研报告': '调研报告',
        '年度报告': '年度报告',
      }
      return map[type] || type || '其他'
    },

    typeTagType(type) {
      var map = {
        '白皮书': 'primary',
        '调研报告': 'warning',
        '年度报告': 'success',
      }
      return map[type] || 'default'
    },

    goBack() {
      uni.navigateBack()
    },
  },
}
</script>

<style scoped>
.page-container {
  min-height: 100vh;
  background: var(--color-bg);
  padding-bottom: 40px;
}

.loading-state {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}

.loading-inline {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
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
  border-radius: 8px;
  font-size: 14px;
}

.list-body {
  padding: 12px 0 24px;
}

.cell-content {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
}

.cell-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.report-icon {
  width: 40rpx;
  height: 40rpx;
  line-height: 40rpx;
  text-align: center;
  font-size: 24rpx;
  color: var(--color-primary);
  background: var(--color-primary-light);
  border-radius: 8rpx;
  flex-shrink: 0;
}

.cell-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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

.cell-action {
  display: flex;
  align-items: center;
  padding: 4px;
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

.load-more {
  text-align: center;
  padding: 16px 0;
}

.no-more {
  color: var(--color-text-placeholder);
  font-size: 13px;
}
</style>
