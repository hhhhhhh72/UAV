<template>
  <view class="status-page">
    <van-nav-bar
      title="企业审核"
      fixed
      placeholder
      left-arrow
      @click-left="goBack"
    />

    <!-- Loading -->
    <view v-if="loading" class="loading-state">
      <van-loading size="24">加载中...</van-loading>
    </view>

    <!-- Error -->
    <view v-else-if="errorMsg" class="state-view">
      <van-empty description="加载失败" image="error" />
      <view class="retry-btn" @tap="fetchEnterprise">
        <text>重新加载</text>
      </view>
    </view>

    <!-- No enterprise -->
    <view v-else-if="!enterprise" class="empty-state">
      <van-empty image="search" description="暂无企业信息" />
      <view class="action-wrap">
        <van-button type="primary" round @tap="goRegister">
          立即入驻
        </van-button>
      </view>
    </view>

    <!-- Enterprise info -->
    <view v-else class="info-body">
      <!-- Status card -->
      <view class="status-card">
        <van-tag :type="statusTagType(enterprise.status)" size="large">
          {{ statusLabel(enterprise.status) }}
        </van-tag>
        <text v-if="enterprise.review_comment" class="review-comment">
          {{ enterprise.review_comment }}
        </text>
      </view>

      <van-cell-group inset>
        <van-cell title="企业名称" :value="enterprise.name" />
        <van-cell title="信用代码" :value="enterprise.credit_code" />
        <van-cell title="法人代表" :value="enterprise.legal_person || '未填写'" />
        <van-cell title="联系电话" :value="enterprise.contact_phone || '未填写'" />
        <van-cell title="行业类别" :value="enterprise.industry_category || '未填写'" />
        <van-cell title="企业规模" :value="enterprise.scale || '未填写'" />
        <van-cell title="企业地址" :value="enterprise.address || '未填写'" />
        <van-cell title="申请时间" :value="formatDate(enterprise.created_at)" />
      </van-cell-group>

      <view v-if="enterprise.description" class="desc-section">
        <van-cell-group inset>
          <view class="desc-cell">
            <text class="desc-label">企业描述</text>
            <text class="desc-text">{{ enterprise.description }}</text>
          </view>
        </van-cell-group>
      </view>
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
      enterprise: null,
    }
  },
  onLoad() {
    if (!uni.getStorageSync('accessToken')) {
      uni.navigateTo({ url: '/pages/login/index' })
      return
    }
    this.fetchEnterprise()
  },
  onPullDownRefresh() {
    this.fetchEnterprise().then(function () {
      uni.stopPullDownRefresh()
    })
  },
  methods: {
    async fetchEnterprise() {
      this.loading = true
      this.errorMsg = ''

      try {
        var res = await request({ url: '/api/v1/enterprises' })
        var data = (res && res.data) || res || {}
        var items = Array.isArray(data) ? data : (data && data.items) || []

        if (items.length > 0) {
          this.enterprise = items[0]
        } else {
          this.enterprise = null
        }
      } catch (e) {
        this.errorMsg = '网络异常，请稍后重试'
      } finally {
        this.loading = false
      }
    },
    goRegister() {
      uni.navigateTo({ url: '/pages/enterprise/register' })
    },
    goBack() {
      uni.navigateBack()
    },
    statusLabel(status) {
      var map = {
        draft: '草稿',
        submitted: '已提交',
        supplement_required: '需补充资料',
        approved: '已通过',
        rejected: '未通过',
      }
      return map[status] || status || '未知'
    },
    statusTagType(status) {
      var map = {
        draft: 'default',
        submitted: 'primary',
        supplement_required: 'warning',
        approved: 'success',
        rejected: 'danger',
      }
      return map[status] || 'default'
    },
    formatDate(iso) {
      if (!iso) return ''
      var d = new Date(iso)
      var m = d.getMonth() + 1
      var day = d.getDate()
      var h = d.getHours()
      var min = d.getMinutes()
      return d.getFullYear() + '-' + (m < 10 ? '0' : '') + m + '-' + (day < 10 ? '0' : '') + day + ' ' + (h < 10 ? '0' : '') + h + ':' + (min < 10 ? '0' : '') + min
    },
  },
}
</script>

<style scoped>
.status-page {
  min-height: 100vh;
  background: #f7f8fa;
  padding-bottom: env(safe-area-inset-bottom);
}

.loading-state {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}

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

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 80px;
}

.action-wrap {
  margin-top: 16px;
}

/* Status card */
.status-card {
  background: #fff;
  margin: 12px 16px;
  padding: 16px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  gap: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.02);
}

.review-comment {
  font-size: 13px;
  color: #969799;
  line-height: 1.5;
  flex: 1;
}

/* Info body */
.info-body {
  padding: 12px 0 24px;
}

/* Description */
.desc-section {
  margin-top: 12px;
}

.desc-cell {
  padding: 16px;
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
</style>
