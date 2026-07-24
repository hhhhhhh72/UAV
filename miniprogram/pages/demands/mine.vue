<template>
  <view class="mine-page">
    <van-nav-bar
      title="我的需求"
      fixed
      placeholder
      left-arrow
      @click-left="goBack"
    />

    <van-tabs
      :active="activeTab"
      @change="onTabChange"
      sticky
      :offset-top="0"
    >
      <!-- Tab: 我发布的 -->
      <van-tab title="我发布的">
        <!-- Loading -->
        <view v-if="published.loading" class="loading-state">
          <van-loading size="24">加载中...</van-loading>
        </view>

        <!-- Error -->
        <view v-else-if="published.error" class="state-view">
          <van-empty description="加载失败" image="error" />
          <view class="retry-btn" @tap="fetchPublished">
            <text>重新加载</text>
          </view>
        </view>

        <!-- Empty -->
        <view v-else-if="published.list.length === 0" class="state-view">
          <van-empty image="search" description="暂无发布的需求" />
        </view>

        <!-- Normal -->
        <view v-else class="list-body">
          <van-cell-group inset>
            <van-cell
              v-for="item in published.list"
              :key="item.id"
              :title="item.title"
              is-link
              @tap="goDetail(item.id)"
            >
              <template #label>
                <view class="cell-meta">
                  <van-tag :type="statusTagType(item.status)" size="small">
                    {{ statusLabel(item.status) }}
                  </van-tag>
                  <text class="meta-text">{{ formatBudget(item.budget_fen) }}</text>
                  <text class="meta-date">{{ formatDate(item.created_at) }}</text>
                </view>
              </template>
            </van-cell>
          </van-cell-group>
        </view>
      </van-tab>

      <!-- Tab: 我竞标的 -->
      <van-tab title="我竞标的">
        <!-- Loading -->
        <view v-if="bids.loading" class="loading-state">
          <van-loading size="24">加载中...</van-loading>
        </view>

        <!-- Error -->
        <view v-else-if="bids.error" class="state-view">
          <van-empty description="加载失败" image="error" />
          <view class="retry-btn" @tap="fetchBids">
            <text>重新加载</text>
          </view>
        </view>

        <!-- Empty -->
        <view v-else-if="bids.list.length === 0" class="state-view">
          <van-empty image="search" description="暂无竞标记录" />
        </view>

        <!-- Normal -->
        <view v-else class="list-body">
          <van-cell-group inset>
            <van-cell
              v-for="item in bids.list"
              :key="item.id"
              :title="item.demand_title || '需求 #' + item.demand_id"
              :label="formatBudget(item.amount_fen)"
              is-link
              @tap="goDetail(item.demand_id)"
            >
              <template #right-icon>
                <van-tag :type="bidStatusTagType(item.status)" size="small">
                  {{ bidStatusLabel(item.status) }}
                </van-tag>
              </template>
            </van-cell>
          </van-cell-group>
        </view>
      </van-tab>
    </van-tabs>
  </view>
</template>

<script>
import { request, getStoredUser } from '../../utils/request'

export default {
  data() {
    return {
      activeTab: 0,
      published: {
        loading: false,
        error: false,
        list: [],
      },
      bids: {
        loading: false,
        error: false,
        list: [],
      },
    }
  },
  onLoad() {
    this.fetchPublished()
    this.fetchBids()
  },
  onPullDownRefresh() {
    var fn = this.activeTab === 0 ? this.fetchPublished : this.fetchBids
    fn.call(this).then(function () {
      uni.stopPullDownRefresh()
    })
  },
  methods: {
    async fetchPublished() {
      this.published.loading = true
      this.published.error = false

      try {
        const res = await request({
          url: '/api/v1/demands',
          data: { mine: 1, page: 1, page_size: 50 },
        })
        const data = (res && res.data) || res || []
        this.published = {
          loading: false,
          error: false,
          list: Array.isArray(data) ? data : (data && data.items) || [],
        }
      } catch (e) {
        this.published = { loading: false, error: true, list: this.published.list }
      }
    },
    async fetchBids() {
      this.bids.loading = true
      this.bids.error = false

      try {
        const res = await request({ url: '/api/v1/demands/bids/mine' })
        const data = Array.isArray(res) ? res : ((res && res.data) || res || [])
        const rawList = Array.isArray(data) ? data : (data && data.items) || []

        // Enrich bid items with demand titles
        var enriched = []
        for (var i = 0; i < rawList.length; i++) {
          var bid = rawList[i]
          var demandTitle = '需求 #' + bid.demand_id
          try {
            const detailRes = await request({
              url: '/api/v1/demands/' + encodeURIComponent(bid.demand_id),
            })
            var detail = (detailRes && detailRes.data) || detailRes
            if (detail && detail.title) {
              demandTitle = detail.title
            }
          } catch (e) {
            // fallback to demand_id
          }
          enriched.push({
            id: bid.id,
            demand_id: bid.demand_id,
            demand_title: demandTitle,
            amount_fen: bid.amount_fen,
            status: bid.status,
            created_at: bid.created_at,
            updated_at: bid.updated_at,
          })
        }
        this.bids = { loading: false, error: false, list: enriched }
      } catch (e) {
        this.bids = { loading: false, error: true, list: this.bids.list }
      }
    },
    onTabChange(e) {
      this.activeTab = e.detail.index
    },
    goDetail(id) {
      uni.navigateTo({ url: '/pages/demands/detail?id=' + encodeURIComponent(id) })
    },
    goBack() {
      uni.navigateBack()
    },
    statusLabel(status) {
      var map = {
        pending: '待审核',
        published: '进行中',
        matched: '已匹配',
        completed: '已完成',
        cancelled: '已取消',
        rejected: '已驳回',
      }
      return map[status] || status || '未知'
    },
    statusTagType(status) {
      var map = {
        pending: 'warning',
        published: 'primary',
        matched: 'success',
        completed: 'success',
        cancelled: 'default',
        rejected: 'danger',
      }
      return map[status] || 'default'
    },
    bidStatusLabel(status) {
      var map = { pending: '待选', accepted: '已中标', rejected: '未选中' }
      return map[status] || status || '未知'
    },
    bidStatusTagType(status) {
      var map = { pending: 'warning', accepted: 'success', rejected: 'danger' }
      return map[status] || 'default'
    },
    formatBudget(fen) {
      if (fen == null || fen === 0) return '面议'
      var yuan = (fen / 100).toFixed(2)
      return '¥' + yuan.replace(/\.00$/, '')
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
.mine-page {
  min-height: 100vh;
  background: #f7f8fa;
  padding-bottom: env(safe-area-inset-bottom);
}

/* State views */
.loading-state {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}

.state-view {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 80px;
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

.meta-date {
  font-size: 12px;
  color: #c8c9cc;
}
</style>
