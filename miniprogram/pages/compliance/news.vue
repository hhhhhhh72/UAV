<template>
  <view class="news-page">
    <van-nav-bar
      title="政策资讯"
      fixed
      placeholder
      left-arrow
      @click-left="goBack"
    />

    <!-- Sticky search + tabs -->
    <van-sticky>
      <van-search
        v-model="searchText"
        placeholder="搜索政策资讯"
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
      <van-empty image="search" description="暂无政策资讯" />
    </view>

    <!-- Normal: article list -->
    <view v-else class="list-body">
      <van-cell-group inset>
        <van-cell
          v-for="item in list"
          :key="item.id"
          :title="item.title || '--'"
          :title-width="'100%'"
          is-link
          @tap="showDetail(item)"
        >
          <template #label>
            <view class="cell-meta">
              <van-tag
                :color="categoryTagColor(item.category)"
                size="small"
                text-color="#fff"
              >
                {{ categoryLabel(item.category) }}
              </van-tag>
              <text v-if="item.source" class="meta-text">{{ item.source }}</text>
              <text class="meta-text">{{ formatDate(item.publish_date || item.created_at) }}</text>
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

    <!-- Detail popup -->
    <van-popup
      :show="detailPopupVisible"
      position="bottom"
      round
      closeable
      @close="detailPopupVisible = false"
      custom-style="padding: 24px 16px 32px; max-height: 80vh; overflow-y: auto;"
    >
      <view v-if="selectedArticle" class="article-detail">
        <text class="article-title">{{ selectedArticle.title }}</text>
        <view class="article-meta">
          <van-tag
            :color="categoryTagColor(selectedArticle.category)"
            size="small"
            text-color="#fff"
          >
            {{ categoryLabel(selectedArticle.category) }}
          </van-tag>
          <text class="article-source">{{ selectedArticle.source || '' }}</text>
          <text class="article-date">{{ formatDate(selectedArticle.publish_date || selectedArticle.created_at) }}</text>
        </view>
        <view class="article-content">
          <text>{{ selectedArticle.content || selectedArticle.summary || '暂无详细内容' }}</text>
        </view>
      </view>
    </van-popup>
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
        { label: '政策', value: 'policy' },
        { label: '法规', value: 'regulation' },
        { label: '标准', value: 'standard' },
        { label: '指南', value: 'guide' },
      ],
      detailPopupVisible: false,
      selectedArticle: null,
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
        if (this.activeCategory) params.category = this.activeCategory
        if (this.searchText) params.q = this.searchText

        var res = await request({ url: '/api/v1/articles', data: params })
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
      this.activeCategory = e.detail ? e.detail.name : e
      this.fetchList(true)
    },
    showDetail(item) {
      this.selectedArticle = item
      this.detailPopupVisible = true
    },
    goBack() {
      uni.navigateBack()
    },
    categoryLabel(cat) {
      var map = {
        policy: '政策',
        regulation: '法规',
        standard: '标准',
        guide: '指南',
      }
      return map[cat] || cat || '其他'
    },
    categoryTagColor(cat) {
      var map = {
        policy: '#1565C0',
        regulation: '#ef4444',
        standard: '#34c759',
        guide: '#7c3aed',
      }
      return map[cat] || '#999'
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
.news-page {
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

/* List */
.list-body {
  padding: 12px 0 24px;
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

/* Load more */
.load-more {
  text-align: center;
  padding: 16px 0;
}

.no-more {
  color: #c8c9cc;
  font-size: 13px;
}

/* Article detail popup */
.article-detail {
  display: flex;
  flex-direction: column;
}

.article-title {
  font-size: 18px;
  font-weight: 700;
  color: #323233;
  line-height: 1.4;
  display: block;
  margin-bottom: 12px;
}

.article-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f2f3f5;
}

.article-source {
  font-size: 12px;
  color: #969799;
}

.article-date {
  font-size: 12px;
  color: #c8c9cc;
}

.article-content {
  font-size: 14px;
  color: #323233;
  line-height: 1.8;
  white-space: pre-wrap;
}
</style>
