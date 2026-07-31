<template>
  <view class="page-container">
    <!-- Nav -->
    <van-nav-bar
      title="行业报告"
      left-arrow
      @click-left="goBack"
    />

    <!-- Search -->
    <van-sticky>
      <van-search
        v-model="searchText"
        placeholder="搜索报告"
        shape="round"
        @search="onSearch"
      />
    </van-sticky>

    <!-- Type tabs -->
    <van-tabs
      :active="activeType"
      @change="onTabChange"
    >
      <van-tab
        v-for="tab in typeTabs"
        :key="tab.value"
        :title="tab.label"
        :name="tab.value"
      />
    </van-tabs>

    <!-- Loading state -->
    <view v-if="loading && list.length === 0" class="loading-state">
      <van-loading size="24">加载中...</van-loading>
    </view>

    <!-- Empty state -->
    <view v-else-if="!loading && list.length === 0 && !errorMsg" class="empty-state-wrapper">
      <van-empty image="search" description="暂无报告" />
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg && list.length === 0" class="error-state">
      <van-empty image="network" description="加载失败" />
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
          :title-width="'100%'"
        >
          <template #icon>
            <text class="report-icon">📄</text>
          </template>
          <template #label>
            <view class="cell-meta">
              <van-tag :type="typeTagType(item.report_type || item.type)" size="small">
                {{ typeLabel(item.report_type || item.type) }}
              </van-tag>
              <text v-if="item.publish_date" class="meta-text">{{ item.publish_date }}</text>
            </view>
          </template>
          <template #default>
            <view class="cell-action" @tap.stop="downloadReport(item)">
              <van-icon name="down" size="18" color="#0A66C2" />
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
    onTabChange(e) {
      var name = e.detail ? e.detail.name : e
      this.activeType = name
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
  background: #f7f8fa;
  padding-bottom: 40px;
}

.loading-state {
  display: flex;
  justify-content: center;
  padding: 80px 0;
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
  background: #0A66C2;
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}

.list-body {
  padding: 12px 0 24px;
}

.report-icon {
  font-size: 24px;
  margin-right: 8px;
  align-self: flex-start;
  margin-top: 12px;
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

.cell-action {
  display: flex;
  align-items: center;
  padding: 4px;
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
