<template>
  <view class="ent-list-page">
    <!-- ═══════ 深蓝顶部：返回 + 标题（对齐需求大厅） ═══════ -->
    <view class="topbar" :style="{ paddingTop: (statusBarH + 4) + 'px' }">
      <view class="topbar-row">
        <view class="back-btn" hover-class="tap-fade" hover-stay-time="120" @tap="goBack">
          <view class="back-arrow"></view>
        </view>
        <text class="top-title">入驻企业</text>
        <view class="topbar-spacer"></view>
      </view>
    </view>

    <!-- 加载中：骨架屏 -->
    <view v-if="loading" class="skeleton-list">
      <view class="skeleton-card"></view>
      <view class="skeleton-card"></view>
      <view class="skeleton-card"></view>
    </view>

    <!-- 空数据 -->
    <view v-else-if="list.length === 0" class="state-panel">
      <view class="state-mark">⌁</view>
      <text class="state-title">暂无入驻企业</text>
      <text class="state-desc">企业完成入驻审核后将在此公示</text>
    </view>

    <!-- 企业卡片列表（PRD FR-2.3：logo/名称/分类标签/核心能力/认证状态） -->
    <view v-else class="card-list">
      <view v-for="e in list" :key="e.id" class="ent-card" hover-class="tap-fade">
        <view class="ent-header">
          <view class="ent-logo">
            <image v-if="e.logo" :src="resolveUrl(e.logo)" mode="aspectFill" class="ent-logo-img" />
            <view v-else class="ent-logo-fallback">{{ e.name ? e.name.charAt(0) : '企' }}</view>
          </view>
          <view class="ent-body">
            <view class="ent-name-row">
              <text class="ent-name">{{ e.name }}</text>
              <text v-if="e.is_member" class="member-badge">会员</text>
            </view>
            <text class="ent-date">入驻 {{ formatDate(e.created_at) }}</text>
          </view>
        </view>

        <view v-if="categoryList(e).length || tagList(e).length" class="tag-row">
          <text v-for="c in categoryList(e)" :key="c" class="tag blue">{{ c }}</text>
          <text v-for="t in tagList(e)" :key="t" class="tag gray">{{ t }}</text>
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
const statusBarH = ref(20)

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
  try {
    statusBarH.value = uni.getSystemInfoSync().statusBarHeight || 20
  } catch (e) {
    // 默认 20
  }
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
  background: #F4F6F8;
  padding-bottom: env(safe-area-inset-bottom);
}

.tap-fade { opacity: 0.85; }

/* ═══════ 深蓝顶部（对齐需求大厅） ═══════ */
.topbar {
  background: #074D92;
  color: #fff;
  padding: 16rpx 24rpx 28rpx;
}
.topbar-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
}
.back-btn {
  width: 60rpx;
  height: 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}
.back-arrow {
  width: 20rpx;
  height: 20rpx;
  border-left: 4rpx solid #fff;
  border-bottom: 4rpx solid #fff;
  transform: rotate(45deg);
  margin-left: 10rpx;
}
.top-title {
  flex: 1;
  font-size: 38rpx;
  font-weight: 700;
  text-align: center;
}
.topbar-spacer { width: 60rpx; }

/* ═══════ 骨架屏（对齐需求大厅） ═══════ */
.skeleton-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
  padding: 24rpx 32rpx;
}
.skeleton-card {
  height: 216rpx;
  border-radius: 16rpx;
  background: linear-gradient(90deg, #E9EDF1 25%, #F5F7F9 37%, #E9EDF1 63%);
  background-size: 400% 100%;
  animation: shimmer 1.3s infinite;
}
@keyframes shimmer {
  0% { background-position: 100% 0; }
  100% { background-position: 0 0; }
}

/* ═══════ 空状态（对齐需求大厅 state-panel） ═══════ */
.state-panel {
  min-height: 620rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 56rpx;
  text-align: center;
  color: #667085;
  font-size: 26rpx;
}
.state-mark {
  width: 124rpx;
  height: 124rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24rpx;
  border-radius: 50%;
  background: #EAF3FB;
  color: #0A66C2;
  font-size: 54rpx;
}
.state-title { font-size: 28rpx; font-weight: 700; color: #17212B; }
.state-desc { margin: 12rpx 0 0; font-size: 22rpx; color: #98A2B3; }

/* ═══════ 企业卡片（对齐需求大厅 trade-card） ═══════ */
.card-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
  padding: 24rpx 32rpx;
}
.ent-card {
  background: #fff;
  border-radius: 16rpx;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.045);
  border: 1px solid rgba(228, 231, 236, 0.7);
  padding: 24rpx;
  overflow: hidden;
}
.ent-header {
  display: flex;
  align-items: center;
  gap: 22rpx;
}
.ent-logo {
  width: 88rpx;
  height: 88rpx;
  border-radius: 14rpx;
  overflow: hidden;
  flex-shrink: 0;
  background: #E8F2FC;
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
  font-size: 32rpx;
  font-weight: 600;
  color: #0A66C2;
}
.ent-body {
  flex: 1;
  min-width: 0;
}
.ent-name-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
}
.ent-name {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 28rpx;
  font-weight: 700;
  color: #17212B;
}
.member-badge {
  flex-shrink: 0;
  font-size: 20rpx;
  color: #fff;
  background: #0A66C2;
  border-radius: 8rpx;
  padding: 4rpx 12rpx;
  line-height: 1.2;
}
.ent-date {
  display: block;
  margin-top: 10rpx;
  font-size: 22rpx;
  color: #667085;
}

/* 标签行（分类蓝 / 能力灰，对齐需求大厅 .tag 系列） */
.tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-top: 18rpx;
}
.tag {
  max-width: 240rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  border-radius: 8rpx;
  padding: 6rpx 12rpx;
  font-size: 20rpx;
  line-height: 1;
}
.tag.blue { color: #0A66C2; background: #EAF3FB; }
.tag.gray { color: #667085; background: #F1F3F5; }

/* 描述：两行截断 */
.ent-desc {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  margin-top: 18rpx;
  font-size: 24rpx;
  color: #667085;
  line-height: 1.6;
}
</style>
