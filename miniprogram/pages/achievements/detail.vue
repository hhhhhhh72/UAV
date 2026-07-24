<template>
  <view class="detail-page">
    <van-nav-bar
      title="成果详情"
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
      <view class="retry-btn" @tap="fetchDetail">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Normal -->
    <template v-else-if="detail">
      <!-- Header card -->
      <view class="header-card">
        <text class="detail-title">{{ detail.title }}</text>
        <view class="tags-row">
          <van-tag :type="fieldTagType(detail.field)" size="medium">
            {{ detail.field || '未知' }}
          </van-tag>
          <van-tag :type="stageTagType(detail.stage)" size="medium">
            {{ stageLabel(detail.stage) }}
          </van-tag>
        </view>
        <view class="header-meta">
          <text v-if="detail.org_name" class="meta-item">{{ detail.org_name }}</text>
          <text class="meta-item">{{ formatDate(detail.created_at || detail.date) }}</text>
        </view>
      </view>

      <!-- Detail info -->
      <van-cell-group inset>
        <view v-if="detail.description" class="info-block">
          <text class="info-label">成果描述</text>
          <text class="info-text">{{ detail.description }}</text>
        </view>
      </van-cell-group>

      <van-cell-group inset>
        <van-cell
          v-if="detail.inventors"
          title="发明人"
          :value="detail.inventors"
          value-class="cell-value"
        />
        <van-cell
          v-if="detail.patent_number"
          title="专利号"
          :value="detail.patent_number"
          value-class="cell-value"
        />
        <van-cell
          v-if="detail.application_area"
          title="应用领域"
          :value="detail.application_area"
          value-class="cell-value"
        />
      </van-cell-group>
    </template>

    <!-- Not found -->
    <view v-else class="state-view">
      <van-empty description="未找到该成果" image="search" />
    </view>
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
      detail: null,
    }
  },
  onLoad(options) {
    this.id = options.id || ''
    this.fetchDetail()
  },
  methods: {
    async fetchDetail() {
      if (!this.id) {
        this.errorMsg = '缺少参数'
        return
      }
      this.loading = true
      this.errorMsg = ''

      try {
        var res = await request({
          url: '/api/v1/achievements/' + encodeURIComponent(this.id),
        })
        var data = (res && res.data) || res || null
        this.detail = data
        if (data && data.title) {
          uni.setNavigationBarTitle({ title: data.title })
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
    stageLabel(stage) {
      var map = {
        laboratory: '实验室',
        pilot: '中试',
        industrialization: '产业化',
        listed: '上市',
      }
      return map[stage] || stage || '未知'
    },
    stageTagType(stage) {
      var map = {
        laboratory: 'primary',
        pilot: 'warning',
        industrialization: 'success',
        listed: 'danger',
      }
      return map[stage] || 'default'
    },
    fieldTagType(field) {
      var map = {
        '无人机': 'primary',
        '飞控': 'warning',
        '载荷': 'success',
        '软件': 'danger',
        '材料': 'default',
      }
      return map[field] || 'default'
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
.detail-page {
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
  background: #1989fa;
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}

/* Header card */
.header-card {
  background: #fff;
  padding: 16px;
  margin: 12px;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.02);
}

.detail-title {
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
  margin-bottom: 12px;
}

.header-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.meta-item {
  font-size: 13px;
  color: #969799;
}

/* Info block inside cell group */
.info-block {
  padding: 16px;
}

.info-label {
  font-size: 14px;
  font-weight: 600;
  color: #323233;
  display: block;
  margin-bottom: 8px;
}

.info-text {
  font-size: 14px;
  color: #333;
  line-height: 1.6;
  display: block;
  white-space: pre-wrap;
}

.cell-value {
  font-size: 14px;
  color: #323233;
  text-align: right;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 200px;
}
</style>
