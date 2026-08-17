<template>
  <view class="ent-detail-page">
    <!-- ═══════ 渐变 Hero：返回 + 标题 ═══════ -->
    <view class="hero" :style="{ paddingTop: (statusBarH + 4) + 'px' }">
      <view class="hero-glow hero-glow-a" />
      <view class="hero-glow hero-glow-b" />
      <view class="topbar-row">
        <view class="back-btn" hover-class="tap-fade" hover-stay-time="120" @tap="goBack">
          <view class="back-arrow"></view>
        </view>
        <view class="topbar-center">
          <text class="top-title">企业详情</text>
        </view>
        <view class="topbar-spacer"></view>
      </view>
    </view>

    <!-- 加载中：骨架屏 -->
    <view v-if="loading" class="skeleton-wrap">
      <view class="skeleton-cover"></view>
      <view class="skeleton-card"></view>
      <view class="skeleton-card"></view>
    </view>

    <!-- 不存在/未审核：统一 404 态 -->
    <view v-else-if="notFound" class="state-panel">
      <view class="state-mark">
        <view class="state-mark-inner">
          <view class="state-building">
            <view class="state-win state-win-1" />
            <view class="state-win state-win-2" />
          </view>
        </view>
      </view>
      <text class="state-title">企业不存在或暂未公开</text>
      <text class="state-desc">仅完成入驻审核的企业在此公示</text>
      <view class="state-btn" hover-class="tap-fade" hover-stay-time="120" @tap="goBack">
        <text>返回列表</text>
      </view>
    </view>

    <template v-else>
      <!-- ═══════ 品牌区：封面 + 叠压 Logo + 名称/认证 ═══════ -->
      <view class="brand-panel">
        <view class="cover-box">
          <image v-if="ent.cover_image" :src="resolveUrl(ent.cover_image)" mode="aspectFill" class="cover-img" @error="ent.cover_image = ''" />
          <view v-else class="cover-fallback">
            <view class="cover-ring cover-ring-a" />
            <view class="cover-ring cover-ring-b" />
          </view>
          <view class="cover-shade" />
        </view>

        <view class="brand-id">
          <view class="brand-logo">
            <image v-if="ent.logo" :src="resolveUrl(ent.logo)" mode="aspectFill" class="brand-logo-img" @error="ent.logo = ''" />
            <view v-else class="brand-logo-fallback">{{ ent.name ? ent.name.charAt(0) : '企' }}</view>
          </view>
          <view class="brand-info">
            <view class="brand-name-row">
              <text class="brand-name">{{ ent.name }}</text>
              <text v-if="ent.is_member" class="member-badge">会员</text>
            </view>
            <view class="ent-verified">
              <view class="verified-dot" />
              <text class="verified-text">协会已认证企业</text>
            </view>
          </view>
        </view>
      </view>

      <!-- ═══════ 行业分类 ═══════ -->
      <view class="section-card">
        <view class="section-head">
          <view class="head-bar" />
          <text class="section-title">行业分类</text>
        </view>
        <view class="tag-wrap">
          <text v-for="c in categoryList(ent)" :key="c" class="tag tag-blue">{{ c }}</text>
          <text v-if="categoryList(ent).length === 0" class="tag-empty">未填写</text>
        </view>
      </view>

      <!-- ═══════ 企业简介 ═══════ -->
      <view class="section-card">
        <view class="section-head">
          <view class="head-bar head-bar-teal" />
          <text class="section-title">企业简介</text>
        </view>
        <text class="desc-text">{{ ent.description || '该企业暂未填写简介' }}</text>
      </view>

      <!-- ═══════ 核心能力 ═══════ -->
      <view v-if="tagList(ent).length" class="section-card">
        <view class="section-head">
          <view class="head-bar head-bar-teal" />
          <text class="section-title">核心能力</text>
        </view>
        <view class="tag-wrap">
          <text v-for="t in tagList(ent)" :key="t" class="tag tag-gray">{{ t }}</text>
        </view>
      </view>

      <!-- ═══════ 基本信息 ═══════ -->
      <view class="section-card">
        <view class="section-head">
          <view class="head-bar" />
          <text class="section-title">基本信息</text>
        </view>
        <view class="info-rows">
          <view class="info-row">
            <text class="info-label">企业规模</text>
            <text class="info-value">{{ ent.scale || '-' }}</text>
          </view>
          <view class="info-row">
            <text class="info-label">所在地区</text>
            <text class="info-value">{{ ent.address || '-' }}</text>
          </view>
          <view class="info-row">
            <text class="info-label">营业时间</text>
            <text class="info-value">{{ ent.business_hours || '-' }}</text>
          </view>
          <view class="info-row">
            <text class="info-label">成立时间</text>
            <text class="info-value">{{ ent.founded_at || '-' }}</text>
          </view>
          <view class="info-row info-row-last">
            <text class="info-label">入驻时间</text>
            <text class="info-value">{{ formatDate(ent.created_at) }}</text>
          </view>
        </view>
      </view>

      <text class="foot-note">信息由企业提交，经协会审核后公示</text>
    </template>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request, BASE_URL } from '../../../utils/request'

const loading = ref(true)
const notFound = ref(false)
const ent = ref({})
const statusBarH = ref(20)

const goBack = () => {
  // 栈感知：有上级页面则返回，否则兜底回列表页
  const pages = getCurrentPages()
  if (pages.length > 1) uni.navigateBack()
  else uni.redirectTo({ url: '/pkg-eco/pages/enterprise/list' })
}

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

onLoad(async (query) => {
  try {
    statusBarH.value = uni.getSystemInfoSync().statusBarHeight || 20
  } catch (e) {
    // 默认 20
  }
  const id = query && query.id
  if (!id) {
    notFound.value = true
    loading.value = false
    return
  }
  try {
    const res = await request({ url: '/api/v1/enterprises/public/detail?id=' + encodeURIComponent(id) })
    ent.value = res || {}
  } catch (e) {
    // 404（未审核/不存在）与其他错误统一进入空态
    notFound.value = true
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.ent-detail-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(48rpx + env(safe-area-inset-bottom));
}

.tap-fade { opacity: 0.85; }

/* ═══════ 渐变 Hero ═══════ */
.hero {
  position: relative;
  overflow: hidden;
  padding: 16rpx 24rpx 96rpx;
  background: linear-gradient(160deg, #074D92 0%, #0A66C2 62%, #0D7AE0 100%);
  color: #fff;
}
.hero-glow {
  position: absolute;
  border-radius: 50%;
  pointer-events: none;
}
.hero-glow-a {
  top: -120rpx;
  right: -80rpx;
  width: 320rpx;
  height: 320rpx;
  background: rgba(255, 255, 255, 0.07);
}
.hero-glow-b {
  top: -30rpx;
  right: 10rpx;
  width: 200rpx;
  height: 200rpx;
  background: rgba(29, 212, 168, 0.12);
}
.topbar-row {
  position: relative;
  z-index: 2;
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
.topbar-center {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.top-title {
  font-size: 38rpx;
  font-weight: 700;
  letter-spacing: 2rpx;
}
.topbar-spacer { width: 60rpx; }

/* ═══════ 品牌区（上叠 Hero 的白色面板） ═══════ */
.brand-panel {
  position: relative;
  margin: -64rpx 24rpx 0;
  background: #fff;
  border-radius: 20rpx;
  overflow: hidden;
  box-shadow: 0 8px 24px rgba(16, 24, 40, 0.08);
}
.cover-box {
  position: relative;
  width: 100%;
  height: 260rpx;
}
.cover-img {
  width: 100%;
  height: 100%;
}
.cover-fallback {
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #0A66C2 0%, #0D7AE0 45%, #1DD4A8 130%);
  overflow: hidden;
}
.cover-ring {
  position: absolute;
  border-radius: 50%;
  border: 2rpx solid rgba(255, 255, 255, 0.25);
}
.cover-ring-a {
  top: -70rpx;
  right: -40rpx;
  width: 240rpx;
  height: 240rpx;
}
.cover-ring-b {
  bottom: -100rpx;
  left: 20%;
  width: 300rpx;
  height: 300rpx;
}
.cover-shade {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 60rpx;
  background: linear-gradient(180deg, transparent, rgba(255, 255, 255, 0.35));
}

/* Logo 叠压封面底部 */
.brand-id {
  position: relative;
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 0 28rpx 28rpx;
}
.brand-logo {
  position: relative;
  width: 128rpx;
  height: 128rpx;
  margin-top: -48rpx;
  flex-shrink: 0;
  border-radius: 24rpx;
  overflow: hidden;
  background: #fff;
  border: 4rpx solid #fff;
  box-shadow: 0 8rpx 20rpx rgba(16, 24, 40, 0.14);
}
.brand-logo-img {
  width: 100%;
  height: 100%;
}
.brand-logo-fallback {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 48rpx;
  font-weight: 700;
  color: #0A66C2;
  background: linear-gradient(150deg, #EAF3FB, #DCEBFA);
}
.brand-info {
  flex: 1;
  min-width: 0;
  margin-top: -4rpx;
}
.brand-name-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
}
.brand-name {
  max-width: 340rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 32rpx;
  font-weight: 700;
  color: #17212B;
}
.member-badge {
  flex-shrink: 0;
  font-size: 18rpx;
  font-weight: 600;
  color: #fff;
  background: linear-gradient(135deg, #0FC293, #1DD4A8);
  border-radius: 999rpx;
  padding: 4rpx 14rpx;
  line-height: 1.3;
  box-shadow: 0 4rpx 10rpx rgba(29, 212, 168, 0.32);
}
.ent-verified {
  display: flex;
  align-items: center;
  gap: 8rpx;
  margin-top: 12rpx;
}
.verified-dot {
  width: 10rpx;
  height: 10rpx;
  border-radius: 50%;
  background: #0FC293;
}
.verified-text {
  font-size: 22rpx;
  color: #0B8A63;
  font-weight: 500;
}

/* ═══════ 内容卡片 ═══════ */
.section-card {
  margin: 20rpx 24rpx 0;
  padding: 26rpx 28rpx;
  background: #fff;
  border-radius: 20rpx;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.045);
}
.section-head {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-bottom: 20rpx;
}
.head-bar {
  width: 8rpx;
  height: 28rpx;
  border-radius: 4rpx;
  background: linear-gradient(180deg, #0D7AE0, #0A66C2);
}
.head-bar-teal {
  background: linear-gradient(180deg, #2EE0B2, #1DD4A8);
}
.section-title {
  font-size: 28rpx;
  font-weight: 700;
  color: #17212B;
}

.tag-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 14rpx;
}
.tag {
  border-radius: 8rpx;
  padding: 8rpx 18rpx;
  font-size: 22rpx;
  line-height: 1.4;
}
.tag-blue {
  color: #0A66C2;
  background: #EAF3FB;
  border: 1rpx solid rgba(10, 102, 194, 0.12);
}
.tag-gray {
  color: #667085;
  background: #F1F3F5;
  border: 1rpx solid rgba(102, 112, 133, 0.1);
}
.tag-empty {
  font-size: 22rpx;
  color: #98A2B3;
}

.desc-text {
  display: block;
  font-size: 24rpx;
  color: #475467;
  line-height: 1.7;
}

.info-rows {
  display: flex;
  flex-direction: column;
}
.info-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24rpx;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #F0F2F5;
}
.info-row-last {
  border-bottom: none;
}
.info-label {
  flex-shrink: 0;
  font-size: 24rpx;
  color: #98A2B3;
}
.info-value {
  font-size: 24rpx;
  color: #344054;
  font-weight: 500;
  text-align: right;
  word-break: break-all;
}

.foot-note {
  display: block;
  margin-top: 32rpx;
  text-align: center;
  font-size: 20rpx;
  color: #B0B9C4;
}

/* ═══════ 骨架屏 ═══════ */
.skeleton-wrap {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
  padding: 24rpx;
}
.skeleton-cover {
  height: 320rpx;
  border-radius: 20rpx;
  background: linear-gradient(90deg, #E9EDF1 25%, #F5F7F9 37%, #E9EDF1 63%);
  background-size: 400% 100%;
  animation: shimmer 1.3s infinite;
}
.skeleton-card {
  height: 200rpx;
  border-radius: 20rpx;
  background: linear-gradient(90deg, #E9EDF1 25%, #F5F7F9 37%, #E9EDF1 63%);
  background-size: 400% 100%;
  animation: shimmer 1.3s infinite;
}
@keyframes shimmer {
  0% { background-position: 100% 0; }
  100% { background-position: 0 0; }
}

/* ═══════ 404 空态 ═══════ */
.state-panel {
  min-height: 640rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 56rpx;
  text-align: center;
}
.state-mark {
  width: 132rpx;
  height: 132rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24rpx;
  border-radius: 50%;
  background: linear-gradient(160deg, #EAF3FB, #F0FAF6);
}
.state-mark-inner {
  width: 92rpx;
  height: 92rpx;
  border-radius: 24rpx;
  background: #fff;
  box-shadow: 0 8rpx 20rpx rgba(10, 102, 194, 0.12);
  display: flex;
  align-items: center;
  justify-content: center;
}
/* CSS 楼宇图标（非 emoji） */
.state-building {
  width: 52rpx;
  height: 44rpx;
  position: relative;
  background: linear-gradient(180deg, #0D7AE0, #0A66C2);
  border-radius: 6rpx 6rpx 2rpx 2rpx;
}
.state-building::after {
  content: '';
  position: absolute;
  left: 14rpx;
  bottom: -8rpx;
  width: 24rpx;
  height: 8rpx;
  background: #1DD4A8;
  border-radius: 0 0 4rpx 4rpx;
}
.state-win {
  position: absolute;
  top: 12rpx;
  width: 8rpx;
  height: 12rpx;
  background: #fff;
  border-radius: 2rpx;
  opacity: 0.85;
}
.state-win-1 { left: 12rpx; }
.state-win-2 { right: 12rpx; }
.state-title { font-size: 28rpx; font-weight: 700; color: #17212B; }
.state-desc { margin: 12rpx 0 0; font-size: 22rpx; color: #98A2B3; }
.state-btn {
  margin-top: 36rpx;
  padding: 16rpx 64rpx;
  border-radius: 50rpx;
  background: linear-gradient(135deg, #0A66C2, #0D7AE0);
  box-shadow: 0 8rpx 20rpx rgba(10, 102, 194, 0.28);
  font-size: 26rpx;
  font-weight: 600;
  color: #fff;
}
</style>
