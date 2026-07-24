<template>
  <view class="resource-list-page">
    <!-- Sticky header: search + tabs -->
    <van-sticky>
      <van-search
        v-model="searchText"
        placeholder="搜索产业资源"
        shape="round"
        @search="onSearch"
      />
      <van-tabs
        :active="activeType"
        @change="onTabChange"
        :ellipsis="false"
        swipeable
      >
        <van-tab
          v-for="tab in typeTabs"
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
      <van-empty image="search" description="暂无产业资源" />
    </view>

    <!-- Normal state: 2-column grid -->
    <view v-else class="list-body">
      <van-grid column-num="2" :gutter="12" :border="false">
        <van-grid-item
          v-for="item in list"
          :key="item.id"
          @tap="goDetail(item)"
        >
          <view class="resource-card">
            <view class="card-emoji">{{ resourceEmoji(item.res_type) }}</view>
            <text class="card-name">{{ item.name || '未命名资源' }}</text>
            <text v-if="item.model" class="card-model">{{ item.model }}</text>
            <text class="card-fee">¥{{ item.daily_fee || 0 }}/天</text>
            <text v-if="item.location" class="card-location">
              <van-icon name="location-o" size="10" color="#969799" />
              {{ item.location }}
            </text>
          </view>
        </van-grid-item>
      </van-grid>

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
        { label: '无人机', value: 'drone' },
        { label: '机场', value: 'airport' },
        { label: '试飞场地', value: 'test_site' },
        { label: '测试基地', value: 'test_base' },
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
        var params = {
          page: this.page,
          page_size: this.pageSize,
        }
        if (this.activeType) params.res_type = this.activeType
        if (this.searchText) params.q = this.searchText

        var res = await request({ url: '/api/v1/industry-resources', data: params })
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
      this.activeType = e.detail ? e.detail.name : e
      this.fetchList(true)
    },
    goDetail(item) {
      uni.navigateTo({ url: '/pages/resources/detail?id=' + encodeURIComponent(item.id) })
    },
    resourceEmoji(type) {
      var map = {
        drone: '🚁',
        airport: '🏪',
        test_site: '🏞',
        test_base: '🏗',
      }
      return map[type] || '📁'
    },
  },
}
</script>

<style scoped>
.resource-list-page {
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
  background: #1989fa;
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}

/* List body */
.list-body {
  padding: 12px 8px 24px;
}

/* Resource card */
.resource-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 16px 10px 12px;
  text-align: center;
  width: 100%;
  box-sizing: border-box;
}

.card-emoji {
  font-size: 36px;
  line-height: 1.2;
  margin-bottom: 8px;
}

.card-name {
  font-size: 14px;
  font-weight: 600;
  color: #323233;
  display: block;
  margin-bottom: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100%;
}

.card-model {
  font-size: 12px;
  color: #969799;
  display: block;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100%;
}

.card-fee {
  font-size: 14px;
  font-weight: 600;
  color: #ee0a24;
  display: block;
  margin-bottom: 4px;
}

.card-location {
  font-size: 11px;
  color: #969799;
  display: flex;
  align-items: center;
  gap: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100%;
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
