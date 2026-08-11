<template>
  <view class="ent-list-page">
    <u-nav-bar title="入驻企业" show-back @back="goBack" />

    <view v-if="loading" class="loading-state">
      <view class="loading-inline">
        <u-loading size="28rpx" />
        <text>加载中...</text>
      </view>
    </view>

    <view v-else-if="list.length === 0" class="empty-state">
      <u-empty description="暂无入驻企业" />
      <text class="empty-note">企业完成入驻审核后将在此公示</text>
    </view>

    <!-- 企业卡片列表（PRD FR-2.3：logo/名称/分类标签/核心能力/认证状态） -->
    <view v-else class="list-body">
      <view v-for="e in list" :key="e.id" class="ent-card">
        <view class="ent-card-header">
          <view class="ent-logo">
            <image v-if="e.logo" :src="resolveUrl(e.logo)" mode="aspectFill" class="ent-logo-img" />
            <view v-else class="ent-logo-fallback">{{ e.name ? e.name.charAt(0) : '企' }}</view>
          </view>
          <view class="ent-info">
            <view class="ent-name-row">
              <text class="ent-name">{{ e.name }}</text>
              <view v-if="e.is_member" class="member-badge">会员</view>
            </view>
            <text class="ent-date">入驻 {{ formatDate(e.created_at) }}</text>
          </view>
        </view>

        <view v-if="categoryList(e).length" class="ent-tags">
          <view v-for="c in categoryList(e)" :key="c" class="ent-tag">{{ c }}</view>
        </view>

        <view v-if="tagList(e).length" class="ent-tags">
          <view v-for="t in tagList(e)" :key="t" class="ent-tag ent-tag--primary">{{ t }}</view>
        </view>

        <text v-if="e.description" class="ent-desc">{{ e.description }}</text>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request, BASE_URL } from '../../../utils/request'

const loading = ref(false)
const list = ref([])

const goBack = () => uni.navigateBack()

const splitTags = (str) => {
  if (!str) return []
  return String(str).split(',').map((t) => t.trim()).filter(Boolean)
}
// 相对路径（存库格式）→ 完整 URL（预览格式）
const resolveUrl = (u) => {
  if (!u) return ''
  if (u.indexOf('http') === 0) return u
  return BASE_URL + u
}
const categoryList = (e) => splitTags(e.industry_category)
const tagList = (e) => splitTags(e.capability_tags)

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  return `${dt.getFullYear()}-${String(dt.getMonth() + 1).padStart(2, '0')}-${String(dt.getDate()).padStart(2, '0')}`
}

onLoad(async () => {
  loading.value = true
  try {
    const res = await request({ url: '/api/v1/enterprises/public' })
    list.value = Array.isArray(res) ? res : []
  } catch (e) {
    list.value = []
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.ent-list-page {
  min-height: 100vh;
  background: var(--color-bg);
  padding-bottom: env(safe-area-inset-bottom);
}

.loading-state {
  display: flex;
  justify-content: center;
  padding: 80rpx 0;
}

.loading-inline {
  display: flex;
  align-items: center;
  gap: 16rpx;
  color: var(--color-text-secondary);
}

.empty-state {
  padding: 100rpx 0;
  text-align: center;
}

.empty-note {
  display: block;
  margin-top: 16rpx;
  font-size: 24rpx;
  color: var(--color-text-secondary);
}

.list-body {
  padding: 24rpx;
}

/* Enterprise card */
.ent-card {
  background: var(--color-bg-card);
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 24rpx;
  box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.03);
}

.ent-card-header {
  display: flex;
  align-items: center;
  gap: 20rpx;
}

.ent-logo {
  width: 88rpx;
  height: 88rpx;
  border-radius: 16rpx;
  overflow: hidden;
  flex-shrink: 0;
  background: var(--color-bg);
}

.ent-logo-img {
  width: 100%;
  height: 100%;
}

.ent-logo-fallback {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 34rpx;
  font-weight: 600;
  color: var(--color-primary);
  background: rgba(10, 102, 194, 0.08);
}

.ent-info {
  flex: 1;
  min-width: 0;
}

.ent-name-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.ent-name {
  font-size: 30rpx;
  font-weight: 600;
  color: var(--color-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.member-badge {
  font-size: 20rpx;
  color: #fff;
  background: var(--color-primary);
  border-radius: 8rpx;
  padding: 2rpx 10rpx;
  flex-shrink: 0;
}

.ent-date {
  display: block;
  margin-top: 8rpx;
  font-size: 24rpx;
  color: var(--color-text-placeholder);
}

/* Tags */
.ent-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-top: 20rpx;
}

.ent-tag {
  padding: 8rpx 20rpx;
  border-radius: 24rpx;
  background: var(--color-bg);
  border: 1rpx solid var(--color-divider);
  font-size: 22rpx;
  color: var(--color-text-secondary);
}

.ent-tag--primary {
  background: rgba(10, 102, 194, 0.06);
  border-color: rgba(10, 102, 194, 0.2);
  color: var(--color-primary);
}

.ent-desc {
  display: block;
  margin-top: 20rpx;
  font-size: 26rpx;
  color: var(--color-text-secondary);
  line-height: 1.6;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}
</style>
