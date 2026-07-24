<template>
  <view class="courses-page">
    <!-- Nav -->
    <van-nav-bar
      title="培训认证"
      left-arrow
      @click-left="goBack"
    />

    <!-- Search -->
    <van-sticky>
      <van-search
        v-model="searchText"
        placeholder="搜索课程"
        shape="round"
        @search="onSearch"
      />
    </van-sticky>

    <!-- Type filter tabs -->
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
      <van-empty image="search" description="暂无课程" />
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg && list.length === 0" class="error-state">
      <van-empty description="加载失败" image="error" />
      <view class="retry-btn" @tap="fetchList(true)">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Normal state: course list -->
    <view v-else class="list-body">
      <view class="card-list">
        <van-card
          v-for="item in list"
          :key="item.id"
          :title="item.title"
          :desc="item.description || ''"
          :thumb="item.cover_image || item.image || ''"
          :price="formatPrice(item.price_fen)"
          :num="'已有 ' + (item.enrollment_count || 0) + ' 人报名'"
          :tag="item.course_type || ''"
          thumb-mode="aspectFill"
          @tap="goEnroll(item)"
        >
          <template #tags>
            <view class="card-tags">
              <van-tag :type="courseTypeTag(item.course_type)" size="medium">
                {{ item.course_type || '通用' }}
              </van-tag>
              <text v-if="item.duration" class="card-duration">
                {{ item.duration }}
              </text>
            </view>
          </template>
        </van-card>
      </view>

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
        { label: 'CAAC', value: 'CAAC' },
        { label: 'UTC', value: 'UTC' },
        { label: '人社', value: '人社' },
        { label: '飞手', value: '飞手' },
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

        var res = await request({ url: '/api/v1/training-courses', data: params })
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
    goEnroll(item) {
      uni.navigateTo({ url: '/pages/training/enroll?id=' + encodeURIComponent(item.id) })
    },
    goBack() {
      uni.navigateBack()
    },
    courseTypeTag(type) {
      var map = {
        'CAAC': 'primary',
        'UTC': 'success',
        '人社': 'warning',
        '飞手': 'danger',
      }
      return map[type] || 'default'
    },
    formatPrice(fen) {
      if (fen == null || fen === 0) return '免费'
      var yuan = (fen / 100).toFixed(2)
      return '¥' + yuan.replace(/\.00$/, '')
    },
  },
}
</script>

<style scoped>
.courses-page {
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

.card-list {
  display: flex;
  flex-direction: column;
}

.card-tags {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
}

.card-duration {
  font-size: 12px;
  color: #969799;
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
