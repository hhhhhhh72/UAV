<template>
  <view class="recommend-page">
    <u-nav-bar title="智能推荐" show-back @back="goBack" />

    <!-- Notice bar (CSS 实现) -->
    <view class="notice-bar">
      <view class="notice-icon">荐</view>
      <text class="notice-text">基于您的浏览和位置推荐需求</text>
    </view>

    <!-- Loading -->
    <view v-if="loading && list.length === 0" class="loading-state">
      <view class="loading-inline">
        <u-loading size="24rpx" />
        <text>推荐加载中...</text>
      </view>
    </view>

    <!-- Error -->
    <view v-else-if="errorMsg && list.length === 0" class="error-state">
      <u-empty description="加载失败" />
      <view class="retry-btn" @tap="fetchRecommendations(true)">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Empty -->
    <view v-else-if="!loading && list.length === 0" class="empty-state-wrapper">
      <u-empty description="暂无推荐需求" />
    </view>

    <!-- Normal -->
    <view v-else class="list-body">
      <u-cell-group inset>
        <u-cell
          v-for="item in list"
          :key="item.id"
          is-link
          @click="goDetail(item)"
        >
          <template #title>
            <view class="card-content">
              <view class="card-row">
                <view class="card-left">
                  <text class="card-title">{{ item.title }}</text>
                </view>
                <view class="card-right">
                  <view
                    class="match-ring"
                    :class="matchRingClass(item.match_score)"
                  >
                    <text>{{ matchScoreText(item.match_score) }}</text>
                  </view>
                </view>
              </view>
              <view class="cell-meta">
                <u-tag
                  :type="bizTypeTagType(item.biz_type)"
                  size="mini"
                >
                  {{ bizTypeLabel(item.biz_type) }}
                </u-tag>
                <text v-if="item.district" class="meta-text">{{ item.district }}</text>
                <text class="meta-text">{{ formatBudget(item.budget_fen) }}</text>
              </view>
              <view class="card-footer">
                <text class="footer-text">{{ item.publisher_name || '匿名用户' }}</text>
                <text class="footer-text">{{ formatDate(item.created_at) }}</text>
              </view>
            </view>
          </template>
        </u-cell>
      </u-cell-group>
    </view>
  </view>
</template>

<script>
import { request } from '../../utils/request'

export default {
  data() {
    return {
      loading: false,
      errorMsg: '',
      list: [],
    }
  },
  onLoad() {
    this.fetchRecommendations(true)
  },
  onPullDownRefresh() {
    this.fetchRecommendations(true).then(function () {
      uni.stopPullDownRefresh()
    })
  },
  methods: {
    async fetchRecommendations(reset) {
      if (reset) {
        this.loading = true
      }
      this.errorMsg = ''

      try {
        var res
        // Try recommendations endpoint first
        try {
          res = await request({ url: '/api/v1/recommendations' })
        } catch (e) {
          // Fallback to demands sorted by newest
          res = await request({ url: '/api/v1/demands', data: { sort: 'newest', page_size: 20 } })
        }
        var data = Array.isArray(res) ? res : (res && res.items) || []
        this.list = data
      } catch (e) {
        this.errorMsg = '网络异常，请稍后重试'
      } finally {
        this.loading = false
      }
    },
    matchRingClass(score) {
      if (score == null) return 'match-mid'
      if (score >= 80) return 'match-high'
      if (score >= 50) return 'match-mid'
      return 'match-low'
    },
    matchScoreText(score) {
      if (score == null) return '--'
      return Math.round(score) + '%'
    },
    bizTypeLabel(type) {
      var map = {
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
      var map = {
        cable_inspection: 'primary',
        plant_transport: 'success',
        spray_pesticide: 'warning',
        trade_lease: 'danger',
        clean_paint: 'primary',
        other: 'default',
      }
      return map[type] || 'default'
    },
    formatBudget(fen) {
      if (fen == null || fen === 0) return '面议'
      var yuan = (fen / 100).toFixed(2)
      return yuan.replace(/\.00$/, '') + ' 元'
    },
    formatDate(iso) {
      if (!iso) return ''
      var d = new Date(iso)
      var m = d.getMonth() + 1
      var day = d.getDate()
      return d.getFullYear() + '-' + (m < 10 ? '0' : '') + m + '-' + (day < 10 ? '0' : '') + day
    },
    goBack() {
      uni.navigateBack()
    },
    goDetail(item) {
      uni.navigateTo({ url: '/pages/demands/detail?id=' + encodeURIComponent(item.id) })
    },
  },
}
</script>

<style scoped>
.recommend-page {
  min-height: 100vh;
  background: var(--color-bg);
  padding-bottom: env(safe-area-inset-bottom);
}

/* Notice bar */
.notice-bar {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin: 16rpx 24rpx 0;
  padding: 14rpx 24rpx;
  background: var(--color-primary-light);
  border-radius: 12rpx;
}

.notice-icon {
  width: 36rpx;
  height: 36rpx;
  line-height: 36rpx;
  text-align: center;
  font-size: 22rpx;
  color: #fff;
  background: var(--color-primary);
  border-radius: 50%;
  flex-shrink: 0;
}

.notice-text {
  font-size: 13px;
  color: var(--color-primary);
}

/* States */
.loading-state {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}

.loading-inline {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: var(--color-text-secondary);
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
  background: var(--color-primary);
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}

.empty-state-wrapper {
  padding-top: 60px;
}

/* List */
.list-body {
  padding: 12px 0 24px;
}

/* Card row with title and match ring */
.card-content {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
}

.card-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  width: 100%;
}

.card-left {
  flex: 1;
  min-width: 0;
}

.card-title {
  font-size: 15px;
  font-weight: bold;
  color: var(--color-text);
  line-height: 1.4;
}

.card-right {
  margin-left: 12px;
}

/* Match ring */
.match-ring {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: bold;
  color: #fff;
  flex-shrink: 0;
}

.match-high {
  background: var(--color-success);
}

.match-mid {
  background: var(--color-warning);
}

.match-low {
  background: var(--color-text-secondary);
}

/* Cell meta */
.cell-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.meta-text {
  font-size: 12px;
  color: var(--color-text-secondary);
}

/* Footer */
.card-footer {
  display: flex;
  align-items: center;
  gap: 12px;
}

.footer-text {
  font-size: 12px;
  color: var(--color-text-placeholder);
}
</style>
