<template>
  <view class="status-page">
    <u-nav-bar
      title="企业审核"
      show-back
      @back="goBack"
    />

    <!-- Loading -->
    <view v-if="loading" class="loading-state">
      <view class="loading-inline">
        <u-loading size="24rpx" />
        <text>加载中...</text>
      </view>
    </view>

    <!-- Error -->
    <view v-else-if="errorMsg" class="state-view">
      <u-empty description="加载失败" />
      <view class="retry-btn" @tap="fetchEnterprise">
        <text>重新加载</text>
      </view>
    </view>

    <!-- No enterprise -->
    <view v-else-if="!enterprise" class="empty-state">
      <u-empty description="暂无企业信息" />
      <view class="action-wrap">
        <u-button type="primary" round @click="goRegister">
          立即入驻
        </u-button>
        <u-button type="default" round class="view-list-btn" @click="goList">
          查看已认证企业
        </u-button>
      </view>
    </view>

    <!-- Enterprise info -->
    <view v-else class="info-body">
      <!-- Header card: logo + name + status -->
      <view class="header-card">
        <view class="header-logo">
          <image
            v-if="enterprise.logo"
            :src="resolveUrl(enterprise.logo)"
            mode="aspectFill"
            class="header-logo-img"
          />
          <view v-else class="header-logo-fallback">{{ enterprise.name ? enterprise.name.charAt(0) : '企' }}</view>
        </view>
        <view class="header-info">
          <text class="header-name">{{ enterprise.name }}</text>
          <view class="header-status-row">
            <u-tag :type="statusTagType(enterprise.status)" size="mini">
              {{ statusLabel(enterprise.status) }}
            </u-tag>
            <text v-if="isUnderReview" class="header-hint">审核中，请耐心等待</text>
          </view>
        </view>
      </view>

      <!-- Review comment (需补充/驳回) -->
      <view
        v-if="enterprise.review_comment"
        class="review-card"
        :class="{ 'review-card--danger': enterprise.status === 'rejected' }"
      >
        <text class="review-title">
          {{ enterprise.status === 'rejected' ? '驳回原因' : '补充要求' }}
        </text>
        <text class="review-text">{{ enterprise.review_comment }}</text>
      </view>

      <!-- Action buttons by status -->
      <view v-if="canEdit" class="action-section">
        <u-button
          type="primary"
          round
          class="action-btn"
          @click="goEdit"
        >
          {{ enterprise.status === 'rejected' ? '重新编辑并提交' : '去补充资料' }}
        </u-button>
      </view>
      <view v-else-if="enterprise.status === 'approved'" class="action-section">
        <u-button
          type="default"
          round
          class="action-btn"
          @click="goList"
        >
          查看企业展示
        </u-button>
      </view>

      <!-- Profile fields -->
      <view class="section-title">企业档案</view>
      <u-cell-group inset>
        <u-cell title="企业名称" :value="enterprise.name" />
        <u-cell title="信用代码" :value="enterprise.credit_code" />
        <u-cell title="成立时间" :value="enterprise.founded_at || '未填写'" />
        <u-cell title="法人代表" :value="enterprise.legal_person || '未填写'" />
        <u-cell title="联系人" :value="enterprise.contact_person || '未填写'" />
        <u-cell title="联系电话" :value="enterprise.contact_phone || '未填写'" />
        <u-cell title="邮箱" :value="enterprise.email || '未填写'" />
        <u-cell title="企业规模" :value="enterprise.scale || '未填写'" />
        <u-cell title="注册地" :value="enterprise.address || '未填写'" />
        <u-cell title="申请时间" :value="formatDate(enterprise.created_at)" />
      </u-cell-group>

      <!-- 企业分类 chips -->
      <view v-if="categoryList.length" class="chips-section">
        <view class="section-title">企业分类</view>
        <view class="chips-row">
          <view v-for="c in categoryList" :key="c" class="chip">{{ c }}</view>
        </view>
      </view>

      <!-- 能力标签 chips -->
      <view v-if="tagList.length" class="chips-section">
        <view class="section-title">能力标签</view>
        <view class="chips-row">
          <view v-for="t in tagList" :key="t" class="chip chip--primary">{{ t }}</view>
        </view>
      </view>

      <!-- Description -->
      <view v-if="enterprise.description" class="desc-section">
        <view class="section-title">企业简介</view>
        <view class="desc-card">
          <text class="desc-text">{{ enterprise.description }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { request, BASE_URL, getErrorMessage } from '../../../utils/request'

export default {
  data() {
    return {
      loading: false,
      errorMsg: '',
      enterprise: null,
    }
  },
  computed: {
    isUnderReview() {
      return this.enterprise && (this.enterprise.status === 'submitted' || this.enterprise.status === 'draft')
    },
    // draft/supplement_required 可编辑；rejected 需重新编辑提交；approved 不可编辑
    canEdit() {
      if (!this.enterprise) return false
      return ['draft', 'supplement_required', 'rejected'].indexOf(this.enterprise.status) >= 0
    },
    categoryList() {
      return this.splitTags(this.enterprise && this.enterprise.industry_category)
    },
    tagList() {
      return this.splitTags(this.enterprise && this.enterprise.capability_tags)
    },
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
    splitTags(str) {
      if (!str) return []
      return String(str).split(',').map(function (t) { return t.trim() }).filter(Boolean)
    },
    // 相对路径（存库格式）→ 完整 URL（预览格式），空值/完整 URL 原样返回
    resolveUrl(u) {
      if (!u) return ''
      if (u.indexOf('http') === 0) return u
      return BASE_URL + u
    },
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
        this.errorMsg = getErrorMessage(e) || '网络异常，请稍后重试'
      } finally {
        this.loading = false
      }
    },
    goRegister() {
      uni.navigateTo({ url: '/pkg-eco/pages/enterprise/register' })
    },
    goEdit() {
      uni.navigateTo({
        url: '/pkg-eco/pages/enterprise/register?entId=' + encodeURIComponent(this.enterprise.id),
      })
    },
    goList() {
      uni.navigateTo({ url: '/pkg-eco/pages/enterprise/list' })
    },
    goBack() {
      uni.navigateBack()
    },
    statusLabel(status) {
      var map = {
        draft: '草稿',
        submitted: '审核中',
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
  background: var(--color-bg);
  padding-bottom: env(safe-area-inset-bottom);
}

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
  padding-top: 120px;
}

.retry-btn {
  margin-top: 12px;
  padding: 8px 24px;
  background: var(--color-primary);
  color: #fff;
  border-radius: 8px;
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

.view-list-btn {
  margin-top: 16rpx;
}

/* Header card */
.header-card {
  display: flex;
  align-items: center;
  gap: 16px;
  background: var(--color-bg-card);
  margin: 12px 16px;
  padding: 16px;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.02);
}

.header-logo {
  width: 64px;
  height: 64px;
  border-radius: 12px;
  overflow: hidden;
  flex-shrink: 0;
  background: var(--color-bg);
}

.header-logo-img {
  width: 100%;
  height: 100%;
}

.header-logo-fallback {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 26px;
  font-weight: 600;
  color: var(--color-primary);
  background: rgba(10, 102, 194, 0.08);
}

.header-info {
  flex: 1;
  min-width: 0;
}

.header-name {
  font-size: 17px;
  font-weight: 600;
  color: var(--color-text);
  display: block;
  margin-bottom: 6px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.header-status-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-hint {
  font-size: 12px;
  color: var(--color-text-placeholder);
}

/* Review comment card */
.review-card {
  background: #FFF7E6;
  margin: 0 16px 12px;
  padding: 12px 16px;
  border-radius: 12px;
  border-left: 1px solid #FAAD14;
}

.review-card--danger {
  background: #FFF1F0;
  border-left-color: #F5222D;
}

.review-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text);
  display: block;
  margin-bottom: 4px;
}

.review-text {
  font-size: 13px;
  color: var(--color-text);
  opacity: 0.85;
  line-height: 1.5;
}

/* Action section */
.action-section {
  padding: 0 16px 12px;
}

.action-btn {
  width: 100%;
}

/* Section */
.section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text-secondary);
  padding: 8px 20px 4px;
}

/* Chips */
.chips-section {
  margin-top: 4px;
}

.chips-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 8px 16px 12px;
}

.chip {
  padding: 6px 14px;
  border-radius: 24rpx;
  background: var(--color-bg-card);
  border: 1px solid var(--color-divider);
  font-size: 12px;
  color: var(--color-text-secondary);
}

.chip--primary {
  background: rgba(10, 102, 194, 0.06);
  border-color: rgba(10, 102, 194, 0.2);
  color: var(--color-primary);
}

/* Description */
.desc-section {
  margin-top: 4px;
}

.desc-card {
  background: var(--color-bg-card);
  margin: 0 16px 16px;
  padding: 14px 16px;
  border-radius: 12px;
}

.desc-text {
  font-size: 14px;
  color: var(--color-text);
  opacity: 0.85;
  line-height: 1.6;
  white-space: pre-wrap;
}
</style>
