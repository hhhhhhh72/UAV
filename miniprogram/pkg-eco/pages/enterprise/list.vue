<template>
  <view class="ent-list-page">
    <!-- ═══════ 渐变 Hero：返回 + 标题 + 实时统计 ═══════ -->
    <view class="hero" :style="{ paddingTop: (statusBarH + 4) + 'px' }">
      <view class="hero-glow hero-glow-a" />
      <view class="hero-glow hero-glow-b" />

      <view class="topbar-row">
        <view class="back-btn" hover-class="tap-fade" hover-stay-time="120" @tap="goBack">
          <view class="back-arrow"></view>
        </view>
        <view class="topbar-center">
          <text class="top-title">入驻企业</text>
          <text class="top-sub">协会认证 · 优质产业主体</text>
        </view>
        <view class="topbar-spacer"></view>
      </view>

      <!-- 统计条：企业总数 / 协会会员 / 行业覆盖（实时聚合） -->
      <view class="hero-stats">
        <view class="h-stat">
          <text class="h-stat-num">{{ errorMsg ? '--' : list.length }}</text>
          <text class="h-stat-label">入驻企业</text>
        </view>
        <view class="h-stat-divider" />
        <view class="h-stat">
          <text class="h-stat-num">{{ errorMsg ? '--' : memberCount }}</text>
          <text class="h-stat-label">协会会员</text>
        </view>
        <view class="h-stat-divider" />
        <view class="h-stat">
          <text class="h-stat-num">{{ errorMsg ? '--' : industryCount }}</text>
          <text class="h-stat-label">行业覆盖</text>
        </view>
      </view>

      <!-- 搜索框：按企业名称 / 核心能力 / 行业模糊过滤（对齐输入框规范：radius 24rpx） -->
      <view class="hero-search">
        <view class="search-box">
          <view class="search-icon" />
          <input
            class="search-input"
            v-model="keyword"
            placeholder="搜索企业名称、能力、行业"
            placeholder-class="search-ph"
            confirm-type="search"
          />
          <view v-if="keyword" class="search-clear" hover-class="tap-fade" hover-stay-time="120" @tap="keyword = ''" />
        </view>
      </view>
    </view>

    <!-- 加载中：骨架屏 -->
    <view v-if="loading" class="skeleton-list">
      <view class="skeleton-card"></view>
      <view class="skeleton-card"></view>
      <view class="skeleton-card"></view>
    </view>

    <!-- 加载失败（无旧数据）：错误态 + 重试 -->
    <view v-else-if="errorMsg && list.length === 0" class="state-panel">
      <view class="state-mark state-mark-muted">
        <view class="state-mark-inner">
          <view class="state-search-icon" />
        </view>
      </view>
      <text class="state-title">加载失败</text>
      <text class="state-desc">{{ errorMsg }}</text>
      <view class="state-btn" hover-class="tap-fade" hover-stay-time="120" @tap="load">
        <text>重新加载</text>
      </view>
    </view>

    <!-- 空数据（全量无企业） -->
    <view v-else-if="list.length === 0" class="state-panel">
      <view class="state-mark">
        <view class="state-mark-inner">
          <view class="state-building">
            <view class="state-win state-win-1" />
            <view class="state-win state-win-2" />
          </view>
        </view>
      </view>
      <text class="state-title">暂无入驻企业</text>
      <text class="state-desc">企业完成入驻审核后将在此公示</text>
      <view class="state-btn" hover-class="tap-fade" hover-stay-time="120" @tap="goRegister">
        <text>申请入驻</text>
      </view>
    </view>

    <!-- 搜索/筛选无结果 -->
    <view v-else-if="filteredList.length === 0" class="state-panel">
      <view class="state-mark state-mark-muted">
        <view class="state-mark-inner">
          <view class="state-search-icon" />
        </view>
      </view>
      <text class="state-title">未找到相关企业</text>
      <text class="state-desc">换个关键词试试，或浏览全部入驻企业</text>
      <view class="state-btn" hover-class="tap-fade" hover-stay-time="120" @tap="resetFilter">
        <text>清除筛选</text>
      </view>
    </view>

    <template v-else>
      <!-- 有旧数据时刷新失败：错误横幅 + 重试 -->
      <view v-if="errorMsg" class="list-error-banner">
        <text>{{ errorMsg }}</text>
        <text class="list-error-retry" @tap="load">重试</text>
      </view>

      <!-- 行业筛选 chips（从数据聚合） -->
      <scroll-view v-if="cats.length > 1" class="chip-scroll" scroll-x :show-scrollbar="false">
        <view class="chip-row">
          <view
            v-for="c in cats"
            :key="c"
            class="chip"
            :class="{ 'chip-active': activeCat === c }"
            @tap="activeCat = c"
          >
            <text>{{ c === '' ? '全部' : c }}</text>
          </view>
        </view>
      </scroll-view>

      <!-- 企业卡片列表（两列网格，PRD FR-2.3：logo/名称/分类标签/核心能力/认证状态） -->
      <view class="card-list">
        <view
          v-for="(e, i) in filteredList"
          :key="e.id"
          class="ent-card"
          :style="{ animationDelay: (i * 50) + 'ms' }"
          hover-class="card-hover"
          :hover-stay-time="120"
          @tap="openDetail(e)"
        >
          <!-- 顶部品牌渐变条 -->
          <view class="card-strip" />

          <!-- 会员角标 -->
          <text v-if="e.is_member" class="member-badge">会员</text>

          <view class="ent-logo">
            <image v-if="e.logo" :src="resolveUrl(e.logo)" mode="aspectFill" class="ent-logo-img" @error="e.logo = ''" />
            <view v-else class="ent-logo-fallback">{{ e.name ? e.name.charAt(0) : '企' }}</view>
            <view class="logo-ring" />
          </view>

          <text class="ent-name">{{ e.name }}</text>

          <view class="ent-verified">
            <view class="verified-dot" />
            <text class="verified-text">协会已认证</text>
          </view>

          <view v-if="displayTags(e).length" class="tag-row">
            <text v-for="t in displayTags(e)" :key="t.label" class="tag" :class="t.blue ? 'blue' : 'gray'">{{ t.label }}</text>
            <text v-if="tagMore(e) > 0" class="tag-more">+{{ tagMore(e) }}</text>
          </view>

          <text v-if="e.description" class="ent-desc">{{ e.description }}</text>

          <text class="ent-date">入驻 {{ formatDate(e.created_at) }}</text>
        </view>
      </view>

      <!-- 列表底部提示 + 申请入驻 CTA -->
      <view class="list-foot">
        <text class="list-foot-text">共 {{ filteredList.length }} 家入驻企业</text>
        <view class="foot-join" hover-class="tap-fade" hover-stay-time="120" @tap="goRegister">
          <text class="foot-join-text">您的企业也想入驻？立即申请</text>
          <view class="foot-join-arrow" />
        </view>
      </view>
    </template>

    <!-- 底部固定申请入驻条（按钮规范：radius 50rpx + box-shadow） -->
    <view class="join-bar">
      <view class="join-btn" hover-class="join-btn-hover" hover-stay-time="120" @tap="goRegister">
        <text class="join-btn-text">申请企业入驻</text>
        <view class="join-arrow" />
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { request, BASE_URL } from '../../../utils/request'

const loading = ref(false)
const list = ref([])
const errorMsg = ref('')
const statusBarH = ref(20)
const activeCat = ref('')
const keyword = ref('')

const goBack = () => uni.navigateBack()

const goRegister = () => uni.navigateTo({ url: '/pkg-eco/pages/enterprise/register' })

// 清除搜索与行业筛选
const resetFilter = () => {
  keyword.value = ''
  activeCat.value = ''
}

const openDetail = (e) => {
  uni.navigateTo({ url: '/pkg-eco/pages/enterprise/detail?id=' + encodeURIComponent(e.id) })
}

const load = async () => {
  loading.value = true
  try {
    const res = await request({ url: '/api/v1/enterprises/public' })
    list.value = Array.isArray(res) ? res : []
    errorMsg.value = ''
  } catch (e) {
    // P1 修复：失败不再静默清空——空列表展示错误态+重试；有旧数据时保留并降级统计
    errorMsg.value = '网络异常，请稍后重试'
  } finally {
    loading.value = false
  }
}

onPullDownRefresh(async () => {
  await load()
  uni.stopPullDownRefresh()
})

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

// 卡片标签：分类（蓝）优先，不足 2 个补能力标签（灰），超出计数
const displayTags = (e) => {
  const cats = categoryList(e)
  const tags = tagList(e)
  const shown = []
  for (let i = 0; i < Math.min(2, cats.length); i++) shown.push({ label: cats[i], blue: true })
  for (let i = 0; shown.length < 2 && i < tags.length; i++) shown.push({ label: tags[i], blue: false })
  return shown
}
const tagMore = (e) => Math.max(0, categoryList(e).length + tagList(e).length - 2)

// Hero 统计：会员数 + 行业覆盖（industry_category 首个分类去重）
const memberCount = computed(() => list.value.filter((e) => e.is_member).length)
const industryCount = computed(() => {
  const set = new Set()
  list.value.forEach((e) => {
    const first = categoryList(e)[0]
    if (first) set.add(first)
  })
  return set.size
})

// 筛选 chips：全部 + 各企业行业分类去重（保留原始顺序）
const cats = computed(() => {
  const set = new Set()
  list.value.forEach((e) => categoryList(e).forEach((c) => set.add(c)))
  return ['', ...set]
})

// 关键词匹配：企业名称 / 简介 / 行业 / 能力 任一命中
const matchKeyword = (e) => {
  const kw = keyword.value.trim().toLowerCase()
  if (!kw) return true
  const hay = [e.name, e.description, e.industry_category, e.capability_tags]
    .filter(Boolean).join(' ').toLowerCase()
  return hay.includes(kw)
}

const filteredList = computed(() => {
  const base = activeCat.value
    ? list.value.filter((e) => categoryList(e).includes(activeCat.value))
    : list.value
  return base.filter(matchKeyword)
})

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
  await load()
})
</script>

<style scoped>
.ent-list-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(200rpx + env(safe-area-inset-bottom));
}

.tap-fade { opacity: 0.85; }

/* ═══════ 渐变 Hero ═══════ */
.hero {
  position: relative;
  overflow: hidden;
  padding: 16rpx 24rpx 32rpx;
  background: linear-gradient(160deg, #074D92 0%, #0A66C2 62%, #0D7AE0 100%);
  color: #fff;
}
/* 右上角同心圆装饰 */
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
  gap: 6rpx;
}
.top-title {
  font-size: 38rpx;
  font-weight: 700;
  letter-spacing: 2rpx;
}
.top-sub {
  font-size: 22rpx;
  color: rgba(255, 255, 255, 0.78);
}
.topbar-spacer { width: 60rpx; }

/* 统计条 */
.hero-stats {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: center;
  margin-top: 26rpx;
  padding: 20rpx 8rpx 4rpx;
  border-top: 1rpx solid rgba(255, 255, 255, 0.16);
}
.h-stat {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4rpx;
}
.h-stat-num {
  font-size: 36rpx;
  font-weight: 800;
  line-height: 1.1;
}
.h-stat-label {
  font-size: 20rpx;
  color: rgba(255, 255, 255, 0.72);
}
.h-stat-divider {
  width: 1rpx;
  height: 44rpx;
  background: rgba(255, 255, 255, 0.18);
}

/* ═══════ Hero 搜索框（输入框规范：radius 24rpx；深色 hero 上取白色半透明底） ═══════ */
.hero-search {
  position: relative;
  z-index: 2;
  margin-top: 20rpx;
}
.search-box {
  display: flex;
  align-items: center;
  gap: 14rpx;
  height: 72rpx;
  padding: 0 24rpx;
  border-radius: 24rpx;
  background: rgba(255, 255, 255, 0.16);
  border: 1rpx solid rgba(255, 255, 255, 0.22);
  box-shadow: 0 6rpx 20rpx rgba(7, 77, 146, 0.22);
}
/* CSS 放大镜（非 emoji） */
.search-icon {
  width: 24rpx;
  height: 24rpx;
  border: 3rpx solid rgba(255, 255, 255, 0.85);
  border-radius: 50%;
  position: relative;
  flex-shrink: 0;
}
.search-icon::after {
  content: '';
  position: absolute;
  right: -9rpx;
  bottom: -6rpx;
  width: 12rpx;
  height: 3rpx;
  background: rgba(255, 255, 255, 0.85);
  border-radius: 2rpx;
  transform: rotate(45deg);
}
.search-input {
  flex: 1;
  height: 72rpx;
  font-size: 26rpx;
  color: #fff;
}
.search-ph {
  color: rgba(255, 255, 255, 0.55);
}
/* CSS 清除按钮（圆形 ×） */
.search-clear {
  width: 32rpx;
  height: 32rpx;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.28);
  position: relative;
  flex-shrink: 0;
}
.search-clear::before,
.search-clear::after {
  content: '';
  position: absolute;
  left: 50%;
  top: 50%;
  width: 14rpx;
  height: 3rpx;
  border-radius: 2rpx;
  background: #fff;
}
.search-clear::before { transform: translate(-50%, -50%) rotate(45deg); }
.search-clear::after { transform: translate(-50%, -50%) rotate(-45deg); }

/* 有旧数据时刷新失败横幅 */
.list-error-banner {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 14rpx;
  margin: 24rpx 32rpx 0;
  padding: 16rpx 28rpx;
  border-radius: 12rpx;
  font-size: 24rpx;
  color: #B42318;
  background: #FEF0EF;
}
.list-error-retry {
  color: #0A66C2;
  font-weight: 600;
}

/* ═══════ 行业筛选 chips ═══════ */
.chip-scroll {
  width: 100%;
  white-space: nowrap;
  padding-top: 20rpx;
}
.chip-row {
  display: inline-flex;
  gap: 16rpx;
  padding: 0 32rpx 8rpx;
}
.chip {
  display: inline-flex;
  align-items: center;
  height: 56rpx;
  padding: 0 28rpx;
  border-radius: 999rpx;
  background: #fff;
  border: 1rpx solid #E4E7EC;
  font-size: 24rpx;
  color: #475467;
  transition: all 0.18s ease;
}
.chip-active {
  background: linear-gradient(135deg, #0A66C2, #0D7AE0);
  border-color: transparent;
  color: #fff;
  font-weight: 600;
  box-shadow: 0 6rpx 16rpx rgba(10, 102, 194, 0.28);
}

/* ═══════ 骨架屏 ═══════ */
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

/* ═══════ 空状态 ═══════ */
.state-panel {
  min-height: 620rpx;
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
  animation: floaty 3s ease-in-out infinite;
}
@keyframes floaty {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-8rpx); }
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
/* 搜索空态：灰调图标变体 */
.state-mark-muted {
  background: linear-gradient(160deg, #F1F3F5, #F7F9FB);
}
.state-search-icon {
  width: 44rpx;
  height: 44rpx;
  border: 4rpx solid #C0C8D2;
  border-radius: 50%;
  position: relative;
}
.state-search-icon::after {
  content: '';
  position: absolute;
  right: -14rpx;
  bottom: -9rpx;
  width: 22rpx;
  height: 4rpx;
  border-radius: 2rpx;
  background: #C0C8D2;
  transform: rotate(45deg);
}

/* ═══════ 企业卡片（两列网格，竖向布局） ═══════ */
.card-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20rpx;
  padding: 24rpx 32rpx 0;
}
.ent-card {
  position: relative;
  background: #fff;
  border-radius: 16rpx;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.045);
  border: 1px solid rgba(228, 231, 236, 0.7);
  padding: 28rpx 20rpx 22rpx;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  animation: cardIn 0.42s cubic-bezier(0.25, 0.46, 0.45, 0.94) both;
}
@keyframes cardIn {
  from { opacity: 0; transform: translateY(22rpx); }
  to { opacity: 1; transform: translateY(0); }
}
.card-hover {
  transform: scale(0.97);
  box-shadow: 0 8px 20px rgba(16, 24, 40, 0.1);
}
/* 顶部品牌渐变条 */
.card-strip {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4rpx;
  background: linear-gradient(90deg, #0A66C2, #1DD4A8);
  opacity: 0.85;
}

/* 会员角标（右上角） */
.member-badge {
  position: absolute;
  top: 18rpx;
  right: 18rpx;
  font-size: 18rpx;
  font-weight: 600;
  color: #fff;
  background: linear-gradient(135deg, #0FC293, #1DD4A8);
  border-radius: 999rpx;
  padding: 4rpx 12rpx;
  line-height: 1.2;
  box-shadow: 0 4rpx 10rpx rgba(29, 212, 168, 0.32);
  z-index: 2;
}

.ent-logo {
  position: relative;
  width: 104rpx;
  height: 104rpx;
  border-radius: 20rpx;
  overflow: hidden;
  flex-shrink: 0;
  background: #E8F2FC;
  margin-top: 6rpx;
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
  font-size: 40rpx;
  font-weight: 700;
  color: #0A66C2;
  background: linear-gradient(150deg, #EAF3FB, #DCEBFA);
}
.logo-ring {
  position: absolute;
  right: -16rpx;
  bottom: -16rpx;
  width: 50rpx;
  height: 50rpx;
  border-radius: 50%;
  background: rgba(29, 212, 168, 0.22);
}

.ent-name {
  width: 100%;
  margin-top: 16rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 27rpx;
  font-weight: 700;
  color: #17212B;
}

.ent-verified {
  display: flex;
  align-items: center;
  gap: 8rpx;
  margin-top: 8rpx;
}
.verified-dot {
  width: 10rpx;
  height: 10rpx;
  border-radius: 50%;
  background: #0FC293;
}
.verified-text {
  font-size: 20rpx;
  color: #0B8A63;
  font-weight: 500;
}

/* 标签行：最多 2 个 + 溢出计数（分类蓝 / 能力灰） */
.tag-row {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 8rpx;
  margin-top: 14rpx;
}
.tag {
  max-width: 180rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  border-radius: 8rpx;
  padding: 5rpx 12rpx;
  font-size: 19rpx;
  line-height: 1.4;
}
.tag.blue { color: #0A66C2; background: #EAF3FB; border: 1rpx solid rgba(10, 102, 194, 0.12); }
.tag.gray { color: #667085; background: #F1F3F5; border: 1rpx solid rgba(102, 112, 133, 0.1); }
.tag-more {
  border-radius: 8rpx;
  padding: 5rpx 10rpx;
  font-size: 19rpx;
  line-height: 1.4;
  color: #98A2B3;
  background: #F9FAFB;
}

/* 描述：两行截断，同行两卡高度对齐 */
.ent-desc {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  width: 100%;
  margin-top: 14rpx;
  padding-top: 14rpx;
  border-top: 1rpx dashed #E4E7EC;
  font-size: 21rpx;
  color: #667085;
  line-height: 1.55;
  text-align: left;
}

/* 底部入驻时间（贴底，两卡视觉对齐） */
.ent-date {
  width: 100%;
  margin-top: auto;
  padding-top: 12rpx;
  font-size: 20rpx;
  color: #98A2B3;
}

/* ═══════ 列表底部提示 + 入驻引导 ═══════ */
.list-foot {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 18rpx;
  padding: 36rpx 32rpx 24rpx;
}
.list-foot-text {
  font-size: 22rpx;
  color: #98A2B3;
}
.foot-join {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10rpx;
  padding: 14rpx 32rpx;
  border-radius: 50rpx;
  background: #F4F8FC;
  border: 1rpx solid rgba(10, 102, 194, 0.16);
}
.foot-join-text {
  font-size: 24rpx;
  color: #0A66C2;
  font-weight: 600;
}
.foot-join-arrow {
  width: 12rpx;
  height: 12rpx;
  border-top: 3rpx solid #0A66C2;
  border-right: 3rpx solid #0A66C2;
  transform: rotate(45deg);
  margin-left: 2rpx;
}

/* ═══════ 底部固定申请入驻条（按钮规范：radius 50rpx + box-shadow） ═══════ */
.join-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 20;
  padding: 16rpx 32rpx calc(16rpx + env(safe-area-inset-bottom));
  background: linear-gradient(180deg, rgba(244, 246, 248, 0), #F4F6F8 30%);
  pointer-events: none;
}
.join-btn {
  pointer-events: auto;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
  height: 88rpx;
  border-radius: 50rpx;
  background: linear-gradient(135deg, #0A66C2 0%, #0D7AE0 100%);
  box-shadow: 0 10rpx 26rpx rgba(10, 102, 194, 0.34);
}
.join-btn-hover {
  opacity: 0.88;
  transform: scale(0.985);
}
.join-btn-text {
  font-size: 30rpx;
  font-weight: 700;
  color: #fff;
  letter-spacing: 2rpx;
}
.join-arrow {
  width: 14rpx;
  height: 14rpx;
  border-top: 4rpx solid #fff;
  border-right: 4rpx solid #fff;
  transform: rotate(45deg);
}
</style>
