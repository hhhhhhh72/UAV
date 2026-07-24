<template>
  <view class="page-container">
    <!-- Nav -->
    <van-nav-bar
      title="研发难题广场"
      left-arrow
      @click-left="goBack"
    />

    <!-- Search -->
    <van-sticky>
      <van-search
        v-model="searchText"
        placeholder="搜索研发难题"
        shape="round"
        @search="onSearch"
      />
    </van-sticky>

    <!-- Category filter tabs -->
    <van-tabs
      :active="activeField"
      @change="onTabChange"
    >
      <van-tab
        v-for="tab in fieldTabs"
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
      <van-empty image="search" description="暂无研发难题" />
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg && list.length === 0" class="error-state">
      <van-empty image="network" description="加载失败" />
      <view class="retry-btn" @tap="fetchList(true)">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Normal state: challenge list -->
    <view v-else class="list-body">
      <view class="card-list">
        <van-card
          v-for="item in list"
          :key="item.id"
          :title="item.title"
          :desc="item.org_name || ''"
          :thumb="item.cover_image || item.image || ''"
          thumb-mode="aspectFill"
          @tap="goDetail(item)"
        >
          <template #tags>
            <view class="card-tags">
              <van-tag :type="fieldTagType(item.field)" size="medium">
                {{ item.field || '其他' }}
              </van-tag>
            </view>
            <view class="card-extra">
              <text v-if="item.reward_amount" class="reward-text">
                {{ formatReward(item.reward_amount) }}
              </text>
              <text v-if="item.deadline" class="extra-text">
                截止: {{ formatDate(item.deadline) }}
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
      activeField: '',
      loading: false,
      loadingMore: false,
      errorMsg: '',
      list: [],
      page: 1,
      pageSize: 20,
      hasMore: true,
      fieldTabs: [
        { label: '全部', value: '' },
        { label: '电池', value: '电池' },
        { label: '飞控', value: '飞控' },
        { label: '通信', value: '通信' },
        { label: '材料', value: '材料' },
        { label: 'AI', value: 'AI' },
        { label: '其他', value: '其他' },
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
        if (this.activeField) params.field = this.activeField
        if (this.searchText) params.q = this.searchText
        params.page = this.page
        params.page_size = this.pageSize

        var res = await request({ url: '/api/v1/rd-challenges', data: params })
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
      this.activeField = name
      this.fetchList(true)
    },
    goDetail(item) {
      uni.showToast({ title: '即将上线', icon: 'none' })
    },
    goBack() {
      uni.navigateBack()
    },
    fieldTagType(field) {
      var map = {
        '电池': 'primary',
        '飞控': 'warning',
        '通信': 'success',
        '材料': 'danger',
        'AI': 'primary',
        '其他': 'default',
      }
      return map[field] || 'default'
    },
    formatReward(amount) {
      if (amount == null || amount === 0) return '面议'
      if (amount >= 10000) {
        return '¥' + (amount / 10000).toFixed(1).replace(/\.0$/, '') + '万'
      }
      return '¥' + amount.toLocaleString()
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
.page-container {
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
}

.card-extra {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 6px;
}

.reward-text {
  font-size: 15px;
  font-weight: 600;
  color: #ee0a24;
}

.extra-text {
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
