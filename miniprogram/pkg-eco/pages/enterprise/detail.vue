<template>
  <view class="ent-detail-page" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <!-- 统一导航（与课题/活动/难题详情同款 u-nav-bar：状态栏避让 + 44px 栏高 + 标题 32rpx） -->
    <u-nav-bar title="企业详情" show-back :fixed="true" @back="goBack" />

    <!-- ① Hero（对齐培训详情：全宽封面 + 三层蒙层 + 认证徽章 + 底部信息区） -->
    <view class="hero">
      <image
        v-if="ent.cover_image && !heroImgError"
        :src="resolveUrl(ent.cover_image)"
        mode="aspectFill"
        class="hero-img"
        lazy-load
        @error="heroImgError = true"
      />
      <view v-else class="hero-fallback">
        <view class="hero-fallback-icon">
          <text class="hero-fallback-char">{{ ent.name ? ent.name.charAt(0) : '企' }}</text>
        </view>
      </view>
      <view class="hero-mask" />
      <view class="status-badge"><text class="status-text">协会认证</text></view>
      <view v-if="ent && ent.name" class="hero-bottom">
        <text class="hero-title">{{ ent.name }}</text>
        <text class="hero-org">{{ firstCategory() }}</text>
        <view class="hero-meta-row">
          <text class="hero-meta-text">{{ ent.address || '地区待公开' }}</text>
        </view>
      </view>
    </view>

    <!-- 加载骨架 -->
    <view v-if="loading" class="skeleton-wrap">
      <view class="skeleton-card"></view>
      <view class="skeleton-card"></view>
    </view>

    <!-- 404 空态 -->
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

    <!-- 错误态 -->
    <view v-else-if="err" class="state-panel">
      <view class="state-mark">
        <view class="state-mark-inner">
          <view class="state-building">
            <view class="state-win state-win-1" />
            <view class="state-win state-win-2" />
          </view>
        </view>
      </view>
      <text class="state-title">加载失败</text>
      <text class="state-desc">请检查网络后重试</text>
      <view class="state-btn" hover-class="tap-fade" hover-stay-time="120" @tap="loadDetail">
        <text>重新加载</text>
      </view>
    </view>

    <!-- ② 白色内容区（浮起：28rpx 上圆角 + 上投影） -->
    <template v-else>
      <view class="content">
        <!-- 品牌卡：Logo + 名称 + 会员/认证 -->
        <view class="brand-card">
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
      </view>

      <!-- 底部操作栏：收藏 + 分享 + 联系企业 -->
      <view class="bb-space"></view>
      <view class="bb">
        <view class="bi" :class="{ fv: isFav }" aria-role="button" :aria-label="isFav ? '取消收藏' : '收藏'" @tap="toggleFav">
          <view class="heart" />
        </view>
        <button class="bo" open-type="share" hover-class="bo-hover" hover-start-time="0" hover-stay-time="300" aria-label="转发">分享</button>
        <view class="bp" :class="{ disabled: !canContact }" @tap="onContact">联系企业</view>
      </view>
    </template>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onShareAppMessage } from '@dcloudio/uni-app'
import { request, BASE_URL } from '../../../utils/request'

const loading = ref(true)
const notFound = ref(false)
const err = ref(false)
const ent = ref({})
const statusBarHeight = ref(20)
const heroImgError = ref(false)
const isFav = ref(false)
let detailId = ''

// 收藏（本地持久化，按企业 id 记忆）
const FAV_KEY = 'ent_favs'
const loadFavs = () => {
  try {
    const v = uni.getStorageSync(FAV_KEY)
    return Array.isArray(v) ? new Set(v) : new Set()
  } catch (e) { return new Set() }
}
const favs = loadFavs()
const saveFavs = () => {
  try { uni.setStorageSync(FAV_KEY, Array.from(favs)) } catch (e) { /* 忽略 */ }
}
const toggleFav = () => {
  if (!detailId) return
  if (favs.has(detailId)) { favs.delete(detailId); isFav.value = false }
  else { favs.add(detailId); isFav.value = true }
  saveFavs()
  uni.showToast({ title: isFav.value ? '已收藏' : '已取消收藏', icon: 'none' })
}

// 联系企业：公开响应里可能无电话（脱敏/未公开）——没有就给去向，不装可点
const canContact = computed(() => !!(ent.value && (ent.value.contact_phone || ent.value.contact || ent.value.phone)))
const onContact = () => {
  const phone = ent.value && (ent.value.contact_phone || ent.value.contact || ent.value.phone)
  if (phone) {
    uni.makePhoneCall({ phoneNumber: String(phone).replace(/\s/g, '') })
    return
  }
  uni.showModal({
    title: '联系方式暂未公开',
    content: '该企业暂未公开联系方式，可稍后重试，或通过协会秘书处联系。',
    confirmText: '知道了',
    showCancel: false,
  })
}

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
const firstCategory = () => {
  const arr = categoryList(ent.value || {})
  return arr[0] || '入驻企业'
}

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  return `${dt.getFullYear()}-${String(dt.getMonth() + 1).padStart(2, '0')}-${String(dt.getDate()).padStart(2, '0')}`
}

async function loadDetail() {
  loading.value = true
  notFound.value = false
  err.value = false
  try {
    const res = await request({ url: '/api/v1/enterprises/public/detail?id=' + encodeURIComponent(detailId) })
    ent.value = res || {}
    heroImgError.value = false
    isFav.value = favs.has(detailId)
  } catch (e) {
    // 404（未审核/不存在）走空态；网络/服务端错误走错误态 + 重试
    const status = (e && e.statusCode) || 0
    if (status === 404) notFound.value = true
    else err.value = true
  } finally {
    loading.value = false
  }
}

onShareAppMessage(() => ({
  title: ent.value ? '入驻企业：' + (ent.value.name || '') : '入驻企业',
  path: '/pkg-eco/pages/enterprise/detail?id=' + encodeURIComponent(detailId),
}))

onLoad(async (query) => {
  // 状态栏高度已在 setup 同步取得（与 u-nav-bar 组件一致，首帧即正确，不再闪跳）
  try {
    statusBarHeight.value = uni.getSystemInfoSync().statusBarHeight || 20
  } catch (e) {
    // 默认 20
  }
  detailId = query && query.id
  if (!detailId) {
    notFound.value = true
    loading.value = false
    return
  }
  loadDetail()
})
</script>

<style scoped>
.ent-detail-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(140rpx + env(safe-area-inset-bottom));
  overflow-x: hidden;
}

.tap-fade { opacity: 0.85; }

/* ═══════ ① Hero（全宽 500rpx + 三层蒙层 + 底部信息区） ═══════ */
.hero {
  position: relative;
  width: 100%;
  height: 500rpx;
  overflow: hidden;
}
.hero-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}
.hero-fallback {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(160deg, #0a5897 0%, #074D92 100%);
}
.hero-fallback-icon {
  width: 112rpx;
  height: 112rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 30rpx;
  background: rgba(255, 255, 255, 0.14);
}
.hero-fallback-char {
  font-size: 56rpx;
  font-weight: 700;
  color: rgba(255, 255, 255, 0.92);
}
.hero-mask {
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, rgba(4, 30, 68, 0.08) 0%, rgba(4, 30, 68, 0.05) 34%, rgba(4, 30, 68, 0.8) 100%);
}
.status-badge {
  position: absolute;
  top: 18rpx;
  left: 18rpx;
  padding: 7rpx 14rpx;
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.92);
  z-index: 4;
}
.status-text { color: #0A66C2; font-size: 20rpx; font-weight: 650; }
.hero-bottom {
  position: absolute;
  left: 32rpx;
  right: 32rpx;
  bottom: 56rpx;
  z-index: 3;
}
.hero-title {
  display: block;
  font-size: 40rpx;
  font-weight: 700;
  color: #ffffff;
  line-height: 1.35;
  text-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.32);
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}
.hero-org {
  display: block;
  margin-top: 6rpx;
  font-size: 24rpx;
  color: rgba(255, 255, 255, 0.78);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.hero-meta-row { display: flex; align-items: center; gap: 8rpx; margin-top: 10rpx; }
.hero-meta-text { font-size: 24rpx; color: rgba(255, 255, 255, 0.85); }

/* ═══════ ② 白色内容区（浮起：28rpx 上圆角 + 上投影） ═══════ */
.content {
  position: relative;
  background: #F4F6F8;
  border-radius: 28rpx 28rpx 0 0;
  margin-top: -28rpx;
  padding: 28rpx 24rpx 0;
  box-shadow: 0 -16rpx 48rpx rgba(7, 77, 146, 0.12);
}

/* 品牌卡 */
.brand-card {
  display: flex;
  align-items: center;
  gap: 20rpx;
  background: #fff;
  border: 1rpx solid #E4EAF2;
  border-radius: 20rpx;
  padding: 24rpx;
  margin-bottom: 24rpx;
  box-shadow: 0 6rpx 18rpx rgba(16, 24, 40, 0.06);
}
.brand-logo {
  width: 96rpx;
  height: 96rpx;
  flex-shrink: 0;
  border-radius: 20rpx;
  overflow: hidden;
  background: #fff;
  border: 2rpx solid #EAF1F8;
}
.brand-logo-img { width: 100%; height: 100%; }
.brand-logo-fallback {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 40rpx;
  font-weight: 700;
  color: #0A66C2;
  background: linear-gradient(150deg, #EAF3FB, #DCEBFA);
}
.brand-info { flex: 1; min-width: 0; }
.brand-name-row { display: flex; align-items: center; gap: 12rpx; }
.brand-name {
  max-width: 340rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 30rpx;
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
  margin-top: 10rpx;
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
  margin: 0 0 24rpx;
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
  font-size: 30rpx;
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

/* ═══════ 底部操作栏（训练详情同款：收藏 + 分享 + 橙色主 CTA） ═══════ */
.bb-space { height: 140rpx; }
.bb {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  gap: 16rpx;
  padding: 16rpx 24rpx;
  padding-bottom: calc(16rpx + env(safe-area-inset-bottom));
  background: #fff;
  border-top: 1rpx solid #EEF1F4;
  box-shadow: 0 -8rpx 24rpx rgba(16, 24, 40, 0.05);
  z-index: 60;
}
.bi {
  width: 88rpx;
  height: 76rpx;
  border-radius: 999rpx;
  background: #fff;
  border: 1rpx solid #EEF1F4;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.heart {
  position: relative;
  width: 36rpx;
  height: 32rpx;
}
.heart::before,
.heart::after {
  content: '';
  position: absolute;
  top: 0;
  width: 18rpx;
  height: 28rpx;
  border-radius: 9rpx 9rpx 0 0;
  background: #98A2B3;
}
.heart::before {
  left: 18rpx;
  transform: rotate(-45deg);
  transform-origin: 0 100%;
}
.heart::after {
  left: 0;
  transform: rotate(45deg);
  transform-origin: 100% 100%;
}
.bi.fv .heart::before,
.bi.fv .heart::after { background: #ff3b30; }
.bo {
  height: 76rpx;
  border-radius: 999rpx;
  border: 2rpx solid #0A66C2;
  background: #fff;
  color: #0A66C2;
  font-size: 26rpx;
  font-weight: 600;
  padding: 0 36rpx;
  display: flex;
  align-items: center;
  flex-shrink: 0;
  margin: 0;
}
.bo::after { border: none; }
.bo-hover { opacity: 0.8; }
.bp {
  flex: 1;
  height: 76rpx;
  border-radius: 999rpx;
  background: #F97316;
  color: #fff;
  font-size: 28rpx;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 6rpx 16rpx rgba(249, 115, 22, 0.3);
}
.bp.disabled { background: #98A2B3; box-shadow: none; }

/* ═══════ 骨架屏 ═══════ */
.skeleton-wrap {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
  padding: 24rpx;
  margin-top: -28rpx;
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
