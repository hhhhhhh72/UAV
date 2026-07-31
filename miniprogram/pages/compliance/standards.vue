<template>
  <view class="standards-page">
    <van-nav-bar
      title="团体标准库"
      fixed
      placeholder
      left-arrow
      @click-left="goBack"
    />

    <!-- Sticky search + tabs -->
    <van-sticky>
      <van-search
        v-model="searchText"
        placeholder="搜索团体标准"
        shape="round"
        @search="onSearch"
      />
      <van-tabs
        :active="activeCategory"
        @change="onTabChange"
        :ellipsis="false"
        swipeable
      >
        <van-tab
          v-for="tab in categoryTabs"
          :key="tab.value"
          :title="tab.label"
          :name="tab.value"
        />
      </van-tabs>
    </van-sticky>

    <!-- Loading state -->
    <view v-if="loading && list.length === 0" class="state-view">
      <van-loading size="24">加载中...</van-loading>
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg && list.length === 0" class="state-view">
      <van-empty description="加载失败" image="error" />
      <view class="retry-btn" @tap="fetchList(true)">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Empty state -->
    <view v-else-if="!loading && list.length === 0" class="state-view">
      <van-empty image="search" description="暂无团体标准" />
    </view>

    <!-- Normal: standards list -->
    <view v-else class="list-body">
      <van-cell-group inset>
        <van-cell
          v-for="item in list"
          :key="item.id"
          :title="item.title || '--'"
          :title-width="'100%'"
          @tap="openStandard(item)"
        >
          <template #title>
            <view class="standard-title-row">
              <text class="standard-icon">📄</text>
              <view class="standard-info">
                <text class="standard-title">{{ item.title || '--' }}</text>
                <view class="standard-meta">
                  <text v-if="item.version" class="meta-tag">v{{ item.version }}</text>
                  <text class="meta-date">{{ formatDate(item.publish_date || item.created_at) }}</text>
                </view>
              </view>
            </view>
          </template>
          <template #right-icon>
            <van-icon
              name="down"
              size="20"
              color="#0A66C2"
              @tap.stop="downloadStandard(item)"
            />
          </template>
        </van-cell>
      </van-cell-group>

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
    onTabChange(e) {
      this.activeCategory = e.detail ? e.detail.name : e
      this.fetchList(true)
    },
    openStandard(item) {
      if (item.content || item.url) {
        uni.showModal({
          title: item.title || '标准详情',
          content: item.content || '查看详情链接',
          showCancel: false,
          confirmText: '知道了',
        })
      } else {
        uni.showToast({ title: '即将上线', icon: 'none' })
      }
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
  background: #f7f8fa;
  padding-bottom: env(safe-area-inset-bottom);
}

/* State views */
.state-view {
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
  font-size: 20px;
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
  color: #323233;
  line-height: 1.4;
}

.standard-meta {
  display: flex;
  align-items: center;
  gap: 10px;
}

.meta-tag {
  font-size: 11px;
  color: #0A66C2;
  background: #e8f2ff;
  padding: 1px 8px;
  border-radius: 4px;
}

.meta-date {
  font-size: 11px;
  color: #c8c9cc;
}

/* Load more */
.load-more {
  text-align: center;
  padding: 16px 0;
}

.no-more {
  color: #c8c9cc;
  font-size: 13px;
}
</style>
