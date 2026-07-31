<template>
  <view class="certificates-page">
    <!-- Nav -->
    <van-nav-bar
      title="我的证书"
      left-arrow
      @click-left="goBack"
    />

    <!-- Loading state -->
    <view v-if="loading" class="loading-state">
      <van-loading size="24">加载中...</van-loading>
    </view>

    <!-- Empty state -->
    <view v-else-if="!loading && list.length === 0 && !errorMsg" class="empty-state-wrapper">
      <van-empty image="search" description="暂无证书" />
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg && list.length === 0" class="error-state">
      <van-empty description="加载失败" image="error" />
      <view class="retry-btn" @tap="fetchList">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Normal state: certificate list -->
    <view v-else class="list-body">
      <van-cell-group inset>
        <van-cell
          v-for="item in list"
          :key="item.id"
          :title="item.title || item.cert_name || '证书'"
          is-link
          @tap="viewCert(item)"
        >
          <template #label>
            <view class="cert-meta">
              <van-tag :type="certTypeTag(item.cert_type)" size="small">
                {{ item.cert_type || '通用' }}
              </van-tag>
              <van-tag :type="statusTagType(item.status)" size="small">
                {{ statusLabel(item.status) }}
              </van-tag>
            </view>
            <view class="cert-detail">
              <text v-if="item.cert_number" class="cert-text">编号：{{ item.cert_number }}</text>
              <text v-if="item.issue_date" class="cert-text">发证日期：{{ item.issue_date }}</text>
              <text v-if="item.expiry_date" class="cert-text">有效期至：{{ item.expiry_date }}</text>
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
    if (!uni.getStorageSync('accessToken')) {
      uni.navigateTo({ url: '/pages/login/index' })
      return
    }
    this.fetchList()
  },
  onPullDownRefresh() {
    this.fetchList().then(function () {
      uni.stopPullDownRefresh()
    })
  },
  methods: {
    async fetchList() {
      this.loading = true
      this.errorMsg = ''

      try {
        var res = await request({ url: '/api/v1/certificates/mine' })
        var data = Array.isArray(res) ? res : (res && res.data) || res || {}
        var items = Array.isArray(data) ? data : (data && data.items) || []
        this.list = items
      } catch (e) {
        this.errorMsg = '网络异常，请稍后重试'
      } finally {
        this.loading = false
      }
    },
    viewCert(item) {
      // Placeholder — could navigate to certificate detail or show image
      uni.showToast({ title: item.title || item.cert_name || '证书详情', icon: 'none' })
    },
    goBack() {
      uni.navigateBack()
    },
    certTypeTag(type) {
      var map = {
        'CAAC': 'primary',
        'UTC': 'success',
        '人社': 'warning',
        'AOPA': 'primary',
        'ASFC': 'success',
      }
      return map[type] || 'default'
    },
    statusLabel(status) {
      var map = {
        'active': '有效',
        'expired': '已过期',
        'pending': '审核中',
        'revoked': '已吊销',
      }
      return map[status] || status || '未知'
    },
    statusTagType(status) {
      var map = {
        'active': 'success',
        'expired': 'danger',
        'pending': 'warning',
        'revoked': 'danger',
      }
      return map[status] || 'default'
    },
  },
}
</script>

<style scoped>
.certificates-page {
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
  background: #0A66C2;
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}

.list-body {
  padding: 12px 0 24px;
}

.cert-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.cert-detail {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.cert-text {
  font-size: 12px;
  color: #969799;
}
</style>
