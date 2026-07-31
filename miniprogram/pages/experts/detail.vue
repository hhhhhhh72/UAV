<template>
  <view class="expert-detail-page">
    <van-nav-bar
      title="专家详情"
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
      <view class="retry-btn" @tap="fetchExpert">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Normal -->
    <template v-else-if="expert">
      <!-- Header card -->
      <view class="header-card">
        <view class="avatar-wrap">
          <van-image
            v-if="expert.avatar"
            :src="expert.avatar"
            width="80"
            height="80"
            radius="50%"
            fit="cover"
          />
          <view v-else class="avatar-placeholder">
            <van-icon name="manager" size="36" color="#c8c9cc" />
          </view>
        </view>
        <view class="header-info">
          <text class="expert-name">{{ expert.name }}</text>
          <text v-if="expert.title" class="expert-title">{{ expert.title }}</text>
          <text v-if="expert.organization" class="expert-org">{{ expert.organization }}</text>
          <view class="field-tags">
            <van-tag
              v-for="(f, fi) in parseFields(expert.field)"
              :key="fi"
              type="primary"
              size="small"
            >
              {{ f }}
            </van-tag>
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

        <van-cell-group inset>
          <van-cell
            v-if="expert.organization"
            title="所属机构"
            :value="expert.organization"
          />
          <van-cell
            v-if="expert.email"
            title="邮箱"
            :value="expert.email"
          />
          <van-cell
            v-if="expert.phone"
            title="联系电话"
            :value="expert.phone"
          />
          <van-cell
            v-if="expert.created_at"
            title="入驻时间"
            :value="formatDate(expert.created_at)"
          />
        </van-cell-group>
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
  background: #f7f8fa;
  padding-bottom: 40px;
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

/* Header card */
.header-card {
  background: #fff;
  margin: 12px 16px;
  padding: 24px 16px;
  border-radius: 12px;
  display: flex;
  align-items: flex-start;
  gap: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.02);
}

.avatar-wrap {
  flex-shrink: 0;
}

.avatar-placeholder {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: #f7f8fa;
  display: flex;
  align-items: center;
  justify-content: center;
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
  color: #323233;
  display: block;
}

.expert-title {
  font-size: 13px;
  color: #0A66C2;
  display: block;
}

.expert-org {
  font-size: 13px;
  color: #646566;
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
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.02);
}

.section-title {
  font-size: 15px;
  font-weight: 700;
  color: #323233;
  margin-bottom: 10px;
  padding-left: 10px;
  border-left: 3px solid #0A66C2;
}

.section-text {
  font-size: 14px;
  color: #646566;
  line-height: 1.7;
  display: block;
  white-space: pre-wrap;
}
</style>
