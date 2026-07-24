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
        <view v-if="published.loading" class="state-view">
          <van-loading size="24" vertical>加载中...</van-loading>
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
          <van-empty description="暂无发布的需求" />
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
        <view v-if="bids.loading" class="state-view">
          <van-loading size="24" vertical>加载中...</van-loading>
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
          <van-empty description="暂无竞标记录" />
        </view>

        <!-- Normal -->
        <view v-else class="list-body">
          <van-cell-group inset>
            <van-cell
              v-for="item in bids.list"
              :key="item.id"
              :title="item.demand_title || '需求 #' + item.demand_id"
              is-link
              @tap="goDetail(item.demand_id)"
            >
              <template #label>
                <view class="cell-meta">
                  <van-tag :type="bidStatusTagType(item.status)" size="small">
                    {{ bidStatusLabel(item.status) }}
                  </van-tag>
                  <text class="meta-text">{{ formatBudget(item.amount_fen) }}</text>
                  <text class="meta-date">{{ formatDate(item.created_at) }}</text>
                </view>
              </template>
            </van-cell>
          </van-cell-group>
        </view>
      </van-tab>
    </van-tabs>
  </view>
</template>

<script>
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
    const fn = this.activeTab === 0 ? this.fetchPublished : this.fetchBids
    fn().then(() => {
      uni.stopPullDownRefresh()
    })
  },
  methods: {
    async fetchPublished() {
      this.published = { ...this.published, loading: true, error: false }
      const accessToken = uni.getStorageSync('accessToken') || ''
      const url = 'http://localhost:8080/api/v1/demands?mine=1&page=1&page_size=50'
      try {
        const [err, resp] = await uniRequest(url, {
          header: accessToken ? { Authorization: 'Bearer ' + accessToken } : {},
        })
        if (err) {
          this.published = { ...this.published, error: true, loading: false }
          return
        }
        const data = resp.data
        this.published = {
          loading: false,
          error: false,
          list: Array.isArray(data) ? data : (data && data.items) || [],
        }
      } catch (e) {
        this.published = { ...this.published, error: true, loading: false }
      }
    },
    async fetchBids() {
      this.bids = { ...this.bids, loading: true, error: false }
      const accessToken = uni.getStorageSync('accessToken') || ''
      const url = 'http://localhost:8080/api/v1/demands/bids/mine'
      try {
        const [err, resp] = await uniRequest(url, {
          header: accessToken ? { Authorization: 'Bearer ' + accessToken } : {},
        })
        if (err) {
          this.bids = { ...this.bids, error: true, loading: false }
          return
        }
        const data = Array.isArray(resp.data) ? resp.data : (resp.data && resp.data.items) || []

        // Enrich bid items with demand titles
        const enriched = []
        for (const bid of data) {
          let demandTitle = '需求 #' + bid.demand_id
          try {
            const [dErr, dResp] = await uniRequest(
              'http://localhost:8080/api/v1/demands/' + encodeURIComponent(bid.demand_id),
              {
                header: accessToken ? { Authorization: 'Bearer ' + accessToken } : {},
              }
            )
            if (!dErr && dResp) {
              const detail = dResp.data || dResp
              if (detail && detail.title) {
                demandTitle = detail.title
              }
            }
          } catch (e) {
            // fallback to demand_id
          }
          enriched.push({
            ...bid,
            demand_title: demandTitle,
          })
        }
        this.bids = { loading: false, error: false, list: enriched }
      } catch (e) {
        this.bids = { ...this.bids, error: true, loading: false }
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
      const map = {
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
      const map = {
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
      const map = { pending: '待选', accepted: '已中标', rejected: '未选中' }
      return map[status] || status || '未知'
    },
    bidStatusTagType(status) {
      const map = { pending: 'warning', accepted: 'success', rejected: 'danger' }
      return map[status] || 'default'
    },
    formatBudget(fen) {
      if (fen == null || fen === 0) return '面议'
      const yuan = (fen / 100).toFixed(2)
      return yuan.replace(/\.00$/, '') + ' 元'
    },
    formatDate(iso) {
      if (!iso) return ''
      return iso.slice(0, 10)
    },
  },
}

function uniRequest(url, options) {
  return new Promise((resolve) => {
    uni.request({
      url,
      method: 'GET',
      header: options.header || {},
      success: (res) => {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve([null, res.data])
        } else {
          const msg =
            (res.data && res.data.error && res.data.error.message) ||
            '请求失败 (' + res.statusCode + ')'
          resolve([new Error(msg), null])
        }
      },
      fail: (err) => {
        resolve([err || new Error('网络异常'), null])
      },
    })
  })
}
</script>

<style scoped>
.mine-page {
  min-height: 100vh;
  background: #f7f8fa;
  padding-bottom: env(safe-area-inset-bottom);
}

.state-view {
  padding-top: 120px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
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
