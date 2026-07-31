<template>
  <view class="recommend-page">
    <van-nav-bar title="智能推荐" left-arrow @click-left="goBack" />

    <van-notice-bar left-icon="info-o" scrollable>
      💡 基于您的浏览和位置推荐需求
    </van-notice-bar>

    <!-- Loading -->
    <view v-if="loading && list.length === 0" class="loading-state">
      <van-loading size="24">推荐加载中...</van-loading>
    </view>

    <!-- Error -->
    <view v-else-if="errorMsg && list.length === 0" class="error-state">
      <van-empty description="加载失败" image="error" />
      <view class="retry-btn" @tap="fetchRecommendations(true)">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Empty -->
    <view v-else-if="!loading && list.length === 0" class="empty-state-wrapper">
      <van-empty image="search" description="暂无推荐需求" />
    </view>

    <!-- Normal -->
    <view v-else class="list-body">
      <van-cell-group inset>
        <van-cell
          v-for="item in list"
          :key="item.id"
          is-link
          @tap="goDetail(item)"
        >
          <template #title>
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
          </template>
          <template #label>
            <view class="cell-meta">
              <van-tag
                :type="bizTypeTagType(item.biz_type)"
                size="small"
              >
                {{ bizTypeLabel(item.biz_type) }}
              </van-tag>
              <text v-if="item.district" class="meta-text">{{ item.district }}</text>
              <text class="meta-text">{{ formatBudget(item.budget_fen) }}</text>
            </view>
            <view class="card-footer">
              <text class="footer-text">{{ item.publisher_name || '匿名用户' }}</text>
              <text class="footer-text">{{ formatDate(item.created_at) }}</text>
            </view>
          </template>
        </van-cell>
      </van-cell-group>
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
  background: #f7f8fa;
  padding-bottom: env(safe-area-inset-bottom);
}

/* States */
.loading-state {
  display: flex;
  justify-content: center;
  padding: 80px 0;
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
  background: #0A66C2;
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
  color: #1a1a1a;
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
  background: #34c759;
}

.match-mid {
  background: #ff9f0a;
}

.match-low {
  background: #86868b;
}

/* Cell meta */
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

/* Footer */
.card-footer {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 4px;
}

.footer-text {
  font-size: 12px;
  color: #c8c9cc;
}
</style>
