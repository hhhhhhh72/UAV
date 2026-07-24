<template>
  <view class="colleges-page">
    <!-- Nav -->
    <van-nav-bar
      title="院校展示"
      left-arrow
      @click-left="goBack"
    />

    <!-- Search -->
    <van-sticky>
      <van-search
        v-model="searchText"
        placeholder="搜索院校"
        shape="round"
        @search="onSearch"
      />
    </van-sticky>

    <!-- Loading state -->
    <view v-if="loading && list.length === 0" class="loading-state">
      <van-loading size="24">加载中...</van-loading>
    </view>

    <!-- Empty state -->
    <view v-else-if="!loading && list.length === 0 && !errorMsg" class="empty-state-wrapper">
      <van-empty image="search" description="暂无院校" />
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg && list.length === 0" class="error-state">
      <van-empty description="加载失败" image="error" />
      <view class="retry-btn" @tap="fetchList(true)">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Normal state -->
    <view v-else class="list-body">
      <van-card
        v-for="item in list"
        :key="item.id"
        :title="item.name"
        :desc="item.majors || ''"
        :thumb="item.logo || item.image || ''"
        thumb-mode="aspectFit"
        @tap="goDetail(item)"
      >
        <template #tags>
          <view class="card-meta">
            <text v-if="item.facilities" class="meta-text">{{ item.facilities }}</text>
            <text v-if="item.graduates_count != null" class="meta-count">
              {{ item.graduates_count }} 名毕业生
            </text>
          </view>
        </template>
        <template #thumb>
          <van-image
            v-if="item.logo || item.image"
            :src="item.logo || item.image"
            width="48"
            height="48"
            fit="cover"
            round
            custom-style="border-radius: 50%;"
          />
          <view v-else class="thumb-placeholder">
            <text class="thumb-text">{{ (item.name || '院').charAt(0) }}</text>
          </view>
        </template>
      </van-card>

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
      loading: false,
      loadingMore: false,
      errorMsg: '',
      list: [],
      page: 1,
      pageSize: 20,
      hasMore: true,
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
        if (this.searchText) params.q = this.searchText
        params.page = this.page
        params.page_size = this.pageSize

        var res = await request({ url: '/api/v1/colleges', data: params })
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
    goDetail(item) {
      uni.showToast({ title: '即将上线', icon: 'none' })
    },
    goBack() {
      uni.navigateBack()
    },
  },
}
</script>

<style scoped>
.colleges-page {
  min-height: 100vh;
  background: #f7f8fa;
  padding-bottom: env(safe-area-inset-bottom);
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
  background: #1989fa;
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}

.list-body {
  padding: 12px 0 24px;
}

.card-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.meta-text {
  font-size: 12px;
  color: #969799;
}

.meta-count {
  font-size: 12px;
  color: #1989fa;
  font-weight: 500;
}

.thumb-placeholder {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: #eef2ff;
  display: flex;
  align-items: center;
  justify-content: center;
}

.thumb-text {
  font-size: 20px;
  font-weight: 700;
  color: #7c3aed;
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
