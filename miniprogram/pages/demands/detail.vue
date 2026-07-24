<template>
  <view class="demand-detail-page">
    <van-nav-bar
      title="需求详情"
      fixed
      placeholder
      left-arrow
      @click-left="goBack"
    />

    <!-- Loading -->
    <view v-if="loading" class="state-view">
      <van-loading size="24" vertical>加载中...</van-loading>
    </view>

    <!-- Error -->
    <view v-else-if="errorMsg" class="state-view">
      <van-empty description="加载失败" image="error" />
      <view class="retry-btn" @tap="fetchDetail">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Normal -->
    <template v-else-if="detail">
      <!-- Info card -->
      <view class="info-card">
        <text class="demand-title">{{ detail.title }}</text>
        <view class="tags-row">
          <van-tag :type="bizTypeTagType(detail.biz_type)" size="medium">
            {{ bizTypeLabel(detail.biz_type) }}
          </van-tag>
          <van-tag :type="statusTagType(detail.status)" size="medium">
            {{ statusLabel(detail.status) }}
          </van-tag>
        </view>
        <view class="info-rows">
          <view class="info-row">
            <text class="info-label">预算</text>
            <text class="info-value budget">{{ formatBudget(detail.budget_fen) }}</text>
          </view>
          <view class="info-row">
            <text class="info-label">地区</text>
            <text class="info-value">{{ detail.district || '未指定' }}</text>
          </view>
          <view class="info-row">
            <text class="info-label">发布者</text>
            <text class="info-value">{{ detail.publisher_name || '匿名' }}</text>
          </view>
          <view class="info-row">
            <text class="info-label">发布时间</text>
            <text class="info-value">{{ formatDate(detail.created_at) }}</text>
          </view>
        </view>
        <view v-if="detail.description" class="desc-section">
          <text class="desc-label">需求描述</text>
          <text class="desc-text">{{ detail.description }}</text>
        </view>
      </view>

      <!-- Bids section -->
      <view class="bids-section">
        <text class="section-title">竞标列表 ({{ bids.length }})</text>

        <!-- Loading bids -->
        <view v-if="bidsLoading" class="bids-loading">
          <van-loading size="18" vertical>加载竞标...</van-loading>
        </view>

        <!-- Empty bids -->
        <view v-else-if="bids.length === 0" class="bids-empty">
          <van-empty description="暂无竞标" />
        </view>

        <!-- Bid list -->
        <view v-else>
          <van-cell-group inset>
            <van-cell
              v-for="bid in bids"
              :key="bid.id"
              :title="bid.bidder_name || '匿名用户'"
              :label="bid.proposal || '无具体说明'"
              :value="formatBudget(bid.amount_fen)"
            >
              <template #right-icon>
                <text class="bid-time">{{ formatDate(bid.created_at) }}</text>
              </template>
            </van-cell>
          </van-cell-group>
        </view>
      </view>
    </template>

    <!-- Bottom action bar -->
    <view
      v-if="detail && detail.status === 'published' && !isOwnDemand"
      class="action-bar"
    >
      <van-button type="primary" block round @tap="goBid">
        参与竞标
      </van-button>
    </view>
  </view>
</template>

<script>
export default {
  data() {
    return {
      id: '',
      loading: false,
      errorMsg: '',
      detail: null,
      bids: [],
      bidsLoading: false,
    }
  },
  computed: {
    isOwnDemand() {
      return false // Simplified — real check would compare with current user ID
    },
  },
  onLoad(options) {
    this.id = options.id || ''
    this.fetchDetail()
  },
  methods: {
    async fetchDetail() {
      this.loading = true
      this.errorMsg = ''

      const accessToken = uni.getStorageSync('accessToken') || ''
      const url = 'http://localhost:8080/api/v1/demands/' + encodeURIComponent(this.id)

      try {
        const [err, resp] = await uniRequest(url, {
          header: accessToken ? { Authorization: 'Bearer ' + accessToken } : {},
        })
        if (err) {
          this.errorMsg = err.message || '加载失败'
          return
        }
        this.detail = resp.data || resp
        uni.setNavigationBarTitle({ title: (this.detail && this.detail.title) || '需求详情' })
        // Load bids
        this.fetchBids()
      } catch (e) {
        this.errorMsg = '网络异常'
      } finally {
        this.loading = false
      }
    },
    async fetchBids() {
      this.bidsLoading = true
      const accessToken = uni.getStorageSync('accessToken') || ''
      const url = 'http://localhost:8080/api/v1/demands/' + encodeURIComponent(this.id) + '/applications'
      try {
        const [err, resp] = await uniRequest(url, {
          header: accessToken ? { Authorization: 'Bearer ' + accessToken } : {},
        })
        if (err) {
          console.error('Failed to load bids:', err.message)
          return
        }
        this.bids = Array.isArray(resp.data) ? resp.data : (resp.data && resp.data.items) || []
      } catch (e) {
        console.error('Failed to load bids:', e)
      } finally {
        this.bidsLoading = false
      }
    },
    goBid() {
      uni.navigateTo({ url: '/pages/demands/bid?id=' + encodeURIComponent(this.id) })
    },
    goBack() {
      uni.navigateBack()
    },
    bizTypeLabel(type) {
      const map = {
        cable_inspection: '巡检',
        plant_transport: '植保',
        spray_pesticide: '农药',
        trade_lease: '租赁',
        clean_paint: '清洗',
        other: '其他',
      }
      return map[type] || type || '其他'
    },
    bizTypeTagType(type) {
      const map = {
        cable_inspection: 'primary',
        plant_transport: 'success',
        spray_pesticide: 'warning',
        trade_lease: 'danger',
        clean_paint: 'primary',
      }
      return map[type] || 'default'
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
.demand-detail-page {
  min-height: 100vh;
  background: #f7f8fa;
  padding-bottom: 80px;
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

.info-card {
  background: #fff;
  padding: 20px 16px;
  margin: 12px 12px 0;
  border-radius: 12px;
}

.demand-title {
  font-size: 18px;
  font-weight: 700;
  color: #323233;
  display: block;
  margin-bottom: 12px;
  line-height: 1.4;
}

.tags-row {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}

.info-rows {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.info-row {
  display: flex;
  align-items: center;
}

.info-label {
  width: 60px;
  font-size: 14px;
  color: #969799;
  flex-shrink: 0;
}

.info-value {
  font-size: 14px;
  color: #323233;
}

.info-value.budget {
  color: #ee0a24;
  font-weight: 600;
}

.desc-section {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #f2f3f5;
}

.desc-label {
  font-size: 14px;
  font-weight: 600;
  color: #323233;
  display: block;
  margin-bottom: 8px;
}

.desc-text {
  font-size: 14px;
  color: #646566;
  line-height: 1.6;
  display: block;
  white-space: pre-wrap;
}

.bids-section {
  margin: 12px 12px 0;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: #323233;
  padding: 8px 4px 8px;
  display: block;
}

.bids-loading,
.bids-empty {
  padding: 24px 0;
  text-align: center;
}

.bid-time {
  font-size: 12px;
  color: #c8c9cc;
}

.action-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 12px 16px;
  background: #fff;
  border-top: 1px solid #f2f3f5;
  padding-bottom: calc(12px + env(safe-area-inset-bottom));
  z-index: 100;
}
</style>
