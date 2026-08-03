<template>
  <view class="mine-page">
    <u-nav-bar
      title="我的需求"
      show-back
      @back="goBack"
    />

    <!-- Loading -->
    <view v-if="published.loading" class="loading-state">
      <view class="loading-inline">
        <u-loading size="24rpx" />
        <text>加载中...</text>
      </view>
    </view>

    <!-- Error -->
    <view v-else-if="published.error" class="state-view">
      <u-empty description="加载失败" />
      <view class="retry-btn" @tap="fetchPublished">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Empty -->
    <view v-else-if="published.list.length === 0" class="state-view">
      <u-empty description="暂无发布的需求" />
    </view>

    <!-- Normal -->
    <view v-else class="list-body">
      <u-cell-group inset>
        <u-cell
          v-for="item in published.list"
          :key="item.id"
          is-link
          @click="goDetail(item.id)"
        >
          <template #title>
            <view class="cell-content">
              <text class="cell-title">{{ item.title }}</text>
              <view class="cell-meta">
                <u-tag :type="statusTagType(item.status)" size="mini">
                  {{ statusLabel(item.status) }}
                </u-tag>
                <text class="meta-text">{{ formatBudget(item.budget_fen) }}</text>
                <text class="meta-date">{{ formatDate(item.created_at) }}</text>
              </view>
              <view v-if="canOperate(item.status)" class="cell-actions">
                <view
                  v-if="item.status === 'rejected'"
                  class="action-btn action-submit"
                  @tap.stop="submitDemand(item)"
                >
                  重新提交
                </view>
                <view
                  v-if="item.status === 'published'"
                  class="action-btn action-complete"
                  @tap.stop="completeDemand(item)"
                >
                  标记完成
                </view>
                <view
                  v-if="item.status === 'pending' || item.status === 'rejected' || item.status === 'published'"
                  class="action-btn action-cancel"
                  @tap.stop="cancelDemand(item)"
                >
                  取消
                </view>
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
      published: {
        loading: false,
        error: false,
        list: [],
      },
    }
  },
  onLoad() {
    this.fetchPublished()
  },
  onPullDownRefresh() {
    this.fetchPublished().then(function () {
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
    canOperate(status) {
      return status === 'pending' || status === 'rejected' || status === 'published'
    },
    async cancelDemand(item) {
      try {
        await request({
          url: '/api/v1/demands/' + encodeURIComponent(item.id) + '/cancel',
          method: 'POST',
        })
        uni.showToast({ title: '已取消', icon: 'success' })
        this.fetchPublished()
      } catch (e) {
        uni.showToast({
          title: (e && e.data && e.data.error && e.data.error.message) || '操作失败',
          icon: 'none',
        })
      }
    },
    async submitDemand(item) {
      try {
        await request({
          url: '/api/v1/demands/' + encodeURIComponent(item.id) + '/submit',
          method: 'POST',
        })
        uni.showToast({ title: '已重新提交，请等待审核', icon: 'success' })
        this.fetchPublished()
      } catch (e) {
        uni.showToast({
          title: (e && e.data && e.data.error && e.data.error.message) || '操作失败',
          icon: 'none',
        })
      }
    },
    async completeDemand(item) {
      try {
        await request({
          url: '/api/v1/demands/' + encodeURIComponent(item.id) + '/complete',
          method: 'POST',
        })
        uni.showToast({ title: '已标记完成', icon: 'success' })
        this.fetchPublished()
      } catch (e) {
        uni.showToast({
          title: (e && e.data && e.data.error && e.data.error.message) || '操作失败',
          icon: 'none',
        })
      }
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
        published: '已发布',
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
        completed: 'success',
        cancelled: 'default',
        rejected: 'danger',
      }
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
  background: var(--color-bg);
  padding-bottom: env(safe-area-inset-bottom);
}

/* State views */
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

.state-view {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 80px;
}

.retry-btn {
  margin-top: 12px;
  padding: 8px 24px;
  background: var(--color-primary);
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}

/* List */
.list-body {
  padding: 12px 0 24px;
}

.cell-content {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
}

.cell-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text);
}

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

.meta-date {
  font-size: 12px;
  color: var(--color-text-placeholder);
}

/* Actions */
.cell-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
}

.action-btn {
  padding: 4px 16px;
  border-radius: 20px;
  font-size: 12px;
  line-height: 1.7;
}

.action-submit {
  background: var(--color-primary);
  color: #fff;
}

.action-complete {
  background: var(--color-success);
  color: #fff;
}

.action-cancel {
  background: var(--color-bg);
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border);
}
</style>
