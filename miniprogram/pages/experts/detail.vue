<template>
  <view class="expert-detail-page">
    <u-nav-bar
      title="专家详情"
      show-back
      @back="goBack"
    />

    <!-- Loading -->
    <view v-if="loading" class="loading-state">
      <view class="loading-inline">
        <u-loading size="28rpx" />
        <text>加载中...</text>
      </view>
    </view>

    <!-- Error -->
    <view v-else-if="errorMsg" class="state-view">
      <u-empty description="加载失败" />
      <view class="retry-btn" @tap="fetchExpert">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Normal -->
    <template v-else-if="expert">
      <!-- Header card -->
      <view class="header-card">
        <view class="avatar-wrap">
          <image
            v-if="expert.avatar"
            :src="expert.avatar"
            mode="aspectFill"
            class="avatar-img"
          />
          <view v-else class="avatar-placeholder">
            <text class="avatar-placeholder-text">专</text>
          </view>
        </view>
        <view class="header-info">
          <text class="expert-name">{{ expert.name }}</text>
          <text v-if="expert.title" class="expert-title">{{ expert.title }}</text>
          <text v-if="expert.organization" class="expert-org">{{ expert.organization }}</text>
          <view class="field-tags">
            <u-tag
              v-for="(f, fi) in parseFields(expert.field)"
              :key="fi"
              type="primary"
              size="mini"
            >
              {{ f }}
            </u-tag>
          </view>
        </view>
      </view>

      <!-- Info sections -->
      <view class="info-sections">
        <view v-if="expert.bio" class="section-card">
          <view class="section-title">个人简介</view>
          <text class="section-text">{{ expert.bio }}</text>
        </view>

        <view v-if="expert.achievements" class="section-card">
          <view class="section-title">主要成就</view>
          <text class="section-text">{{ expert.achievements }}</text>
        </view>

        <u-cell-group inset>
          <u-cell
            v-if="expert.organization"
            title="所属机构"
            :value="expert.organization"
          />
          <u-cell
            v-if="expert.email"
            title="邮箱"
            :value="expert.email"
          />
          <u-cell
            v-if="expert.phone"
            title="联系电话"
            :value="expert.phone"
          />
          <u-cell
            v-if="expert.created_at"
            title="入驻时间"
            :value="formatDate(expert.created_at)"
          />
        </u-cell-group>
      </view>
    </template>
  </view>
</template>

<script>
import { request } from '../../utils/request'

export default {
  data() {
    return {
      id: '',
      loading: false,
      errorMsg: '',
      expert: null,
    }
  },
  onLoad(options) {
    this.id = options.id || ''
    if (this.id) {
      this.fetchExpert()
    } else {
      this.errorMsg = '缺少专家ID参数'
      this.loading = false
    }
  },
  methods: {
    async fetchExpert() {
      this.loading = true
      this.errorMsg = ''

      try {
        var res = await request({ url: '/api/v1/experts/' + encodeURIComponent(this.id) })
        this.expert = (res && res.data) || res || null

        if (this.expert && this.expert.name) {
          uni.setNavigationBarTitle({ title: this.expert.name })
        }
      } catch (e) {
        this.errorMsg = '网络异常，请稍后重试'
      } finally {
        this.loading = false
      }
    },
    goBack() {
      uni.navigateBack()
    },
    parseFields(field) {
      if (!field) return []
      if (typeof field === 'string') {
        return field.split(/[,，]/).filter(Boolean)
      }
      if (Array.isArray(field)) return field
      return []
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
.expert-detail-page {
  min-height: 100vh;
  background: var(--color-bg);
  padding-bottom: 40px;
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
  justify-content: center;
  gap: 8px;
  font-size: 13px;
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

/* Header card */
.header-card {
  background: #fff;
  margin: 12px 16px;
  padding: 24px 16px;
  border-radius: 8px;
  display: flex;
  align-items: flex-start;
  gap: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.02);
}

.avatar-wrap {
  flex-shrink: 0;
}

.avatar-img {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  display: block;
}

.avatar-placeholder {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: var(--color-bg);
  display: flex;
  align-items: center;
  justify-content: center;
}

.avatar-placeholder-text {
  font-size: 32px;
  color: var(--color-text-placeholder);
  font-weight: 500;
}

.header-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.expert-name {
  font-size: 18px;
  font-weight: 700;
  color: var(--color-text);
  display: block;
}

.expert-title {
  font-size: 13px;
  color: var(--color-primary);
  display: block;
}

.expert-org {
  font-size: 13px;
  color: var(--color-text-secondary);
  display: block;
}

.field-tags {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  margin-top: 4px;
}

/* Info sections */
.info-sections {
  padding: 0 0 24px;
}

.section-card {
  background: #fff;
  margin: 12px 16px;
  padding: 16px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.02);
}

.section-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--color-text);
  margin-bottom: 10px;
  padding-left: 10px;
  border-left: 3px solid var(--color-primary);
}

.section-text {
  font-size: 14px;
  color: var(--color-text-secondary);
  line-height: 1.7;
  display: block;
  white-space: pre-wrap;
}
</style>
