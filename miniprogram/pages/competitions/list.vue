<template>
  <view class="competitions-page">
    <!-- Nav -->
    <van-nav-bar
      title="赛事中心"
      left-arrow
      @click-left="goBack"
    />

    <!-- Search -->
    <van-sticky>
      <van-search
        v-model="searchText"
        placeholder="搜索赛事"
        shape="round"
        @search="onSearch"
      />
    </van-sticky>

    <!-- Category filter tabs -->
    <van-tabs
      :active="activeCategory"
      @change="onTabChange"
    >
      <van-tab
        v-for="tab in categoryTabs"
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
      <van-empty image="search" description="暂无赛事" />
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg && list.length === 0" class="error-state">
      <van-empty description="加载失败" image="error" />
      <view class="retry-btn" @tap="fetchList(true)">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Normal state: competition list -->
    <view v-else class="list-body">
      <view class="card-list">
        <van-card
          v-for="item in list"
          :key="item.id"
          :title="item.title"
          :desc="item.description || ''"
          :thumb="item.cover_image || item.image || ''"
          :num="'已有 ' + (item.registration_count || 0) + ' 人报名'"
          thumb-mode="aspectFill"
          @tap="goDetail(item)"
        >
          <template #tags>
            <view class="card-tags">
              <van-tag :type="categoryTag(item.category)" size="medium">
                {{ item.category || '通用' }}
              </van-tag>
              <van-tag :type="registrationTagType(item.registration_status)" size="medium">
                {{ registrationLabel(item.registration_status) }}
              </van-tag>
            </view>
            <view v-if="item.date || item.location" class="card-extra">
              <text v-if="item.date" class="extra-text">{{ item.date }}</text>
              <text v-if="item.location" class="extra-text">{{ item.location }}</text>
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
      activeCategory: '',
      loading: false,
      loadingMore: false,
      errorMsg: '',
      list: [],
      page: 1,
      pageSize: 20,
      hasMore: true,
      categoryTabs: [
        { label: '全部', value: '' },
        { label: '无人机竞技', value: '无人机竞技' },
        { label: '创新创业', value: '创新创业' },
        { label: '技能大赛', value: '技能大赛' },
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
        if (this.activeCategory) params.category = this.activeCategory
        if (this.searchText) params.q = this.searchText
        params.page = this.page
        params.page_size = this.pageSize

        var res = await request({ url: '/api/v1/competitions', data: params })
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
      this.activeCategory = name
      this.fetchList(true)
    },
    goDetail(item) {
      uni.showToast({ title: '即将上线', icon: 'none' })
    },
    goBack() {
      uni.navigateBack()
    },
    categoryTag(category) {
      var map = {
        '无人机竞技': 'primary',
        '创新创业': 'warning',
        '技能大赛': 'success',
      }
      return map[category] || 'default'
    },
    registrationLabel(status) {
      var map = {
        'open': '报名中',
        'closed': '已结束',
        'full': '已满额',
      }
      return map[status] || status || '未知'
    },
    registrationTagType(status) {
      var map = {
        'open': 'success',
        'closed': 'default',
        'full': 'danger',
      }
      return map[status] || 'default'
    },
  },
}
</script>

<style scoped>
.competitions-page {
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
  margin-top: 4px;
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
