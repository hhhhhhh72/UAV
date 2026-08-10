<template>
  <view class="page">
    <!-- ① Tab + 搜索（原生导航栏，顶部与培训课程列表一致） -->
    <view class="main-card">
      <view class="tabs-container">
        <view class="tab-item" :class="{ active: currentTab === 'all' }" @click="switchTab('all')">全部院校</view>
        <view class="tab-item" :class="{ active: currentTab === 'undergraduate' }" @click="switchTab('undergraduate')">本科院校</view>
        <view class="tab-item" :class="{ active: currentTab === 'vocational' }" @click="switchTab('vocational')">专科院校</view>
      </view>

      <view class="search-bar">
        <u-icon name="search" size="28rpx" color="#98A2B3" />
        <input class="search-input" v-model="keyword" placeholder="搜索院校名称" @input="onSearch" />
      </view>

      <!-- ③ 院校卡片 -->
      <StateView
        class="state-fill"
        :loading="loading"
        :error="!!errorMsg"
        :empty="!loading && !errorMsg && list.length === 0"
        empty-text="暂无院校"
        @retry="loadData"
      >
        <scroll-view class="list-scroll" scroll-y @scrolltolower="loadMore">
          <view
            v-for="item in list"
            :key="item.id"
            class="college-card"
            hover-class="press-feedback"
            :hover-stay-time="120"
            @click="goDetail(item)"
          >
            <!-- 封面（真实校园图 + 类型兜底） -->
            <view class="card-cover" :class="'cover--' + coverType(item)">
              <!-- 真实校园图（有则显示） -->
              <image
                v-if="coverOf(item)"
                :src="coverOf(item)"
                class="cover-img"
                mode="aspectFill"
                lazy-load
                :style="{ opacity: imgLoaded[item.id] ? 1 : 0 }"
                @load="onCoverLoad(item.id)"
              />
              <!-- 无图兜底：类型渐变色块 + 装饰 -->
              <view v-else class="cover-deco-wrap">
                <view class="cover-deco" />
              </view>

              <!-- 左上角类型胶囊 -->
              <view class="cover-type-pill" :class="'cover-type-pill--' + coverType(item)">{{ levelLabel(item) }}</view>
            </view>

            <!-- 圆形校徽（半嵌封面底部，作为独立层可溢出） -->
            <view class="cover-badge">{{ initShort(item) }}</view>

            <view class="card-body">
              <!-- 名称 + 箭头 -->
              <view class="card-header">
                <text class="college-name">{{ item.name || item.title || '未知院校' }}</text>
                <text class="card-arrow">›</text>
              </view>
              <text class="college-location">{{ item.city || '未知城市' }} · {{ (item.tags || ['无人机专业']).join(' · ') }}</text>

              <!-- 数据栏（三行图标式） -->
              <view class="stats-list">
                <view class="stat-line">
                  <view class="stat-icon stat-icon--major"><text class="stat-icon-text">专</text></view>
                  <text class="stat-label">无人机专业</text>
                  <text class="stat-value">{{ item.majorCount || item.major_count || '6' }}</text>
                </view>
                <view class="stat-line">
                  <view class="stat-icon stat-icon--partner"><text class="stat-icon-text">企</text></view>
                  <text class="stat-label">合作企业</text>
                  <text class="stat-value">{{ item.partnerCount || item.partner_count || '28' }}</text>
                </view>
                <view class="stat-line stat-line-last">
                  <view class="stat-icon stat-icon--student"><text class="stat-icon-text">学</text></view>
                  <text class="stat-label">在读学生</text>
                  <text class="stat-value">{{ item.studentCount || item.student_count || '320' }}+</text>
                </view>
              </view>

              <!-- 简介（最多 2 行） -->
              <text class="college-intro">{{ item.intro || item.description || '暂无简介' }}</text>

              <!-- 标签（配色统一） -->
              <view v-if="specTags(item).length > 0" class="tag-row">
                <text v-for="tag in specTags(item)" :key="tag" class="spec-tag" :class="levelTagClass(tag)">{{ tag }}</text>
              </view>
            </view>
          </view>

          <view v-if="list.length > 0" class="load-more-wrap">
            <view v-if="loadingMore" class="loading-inline">
              <u-loading size="24rpx" />
              <text>加载更多...</text>
            </view>
            <text v-else-if="!hasMore" class="no-more">没有更多了</text>
          </view>
          <view style="height:40rpx" />
        </scroll-view>
      </StateView>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { request } from '../../utils/request'
import StateView from '../../components/StateView.vue'

const currentTab = ref('all')
const keyword = ref('')
const loading = ref(false)
const loadingMore = ref(false)
const errorMsg = ref('')
const list = ref([])
const imgLoaded = ref({})
const page = ref(1)
const pageSize = 20
const hasMore = ref(true)

/* 院校层级：985/211=顶尖、本科、专科 */
function collegeLevel(item) {
  var tags = item.tags || []
  if (tags.indexOf('985') >= 0 || tags.indexOf('211') >= 0) return 'top'
  if (tags.indexOf('专科') >= 0 || tags.indexOf('高职') >= 0) return 'vocational'
  return 'undergraduate'
}

/** 封面图 URL：兼容 cover / image / campus_image */
function coverOf(item) {
  const u = item.cover || item.image || item.campus_image || item.cover_image
  return u ? u : ''
}

/** 封面图加载完成淡入 */
function onCoverLoad(id) {
  imgLoaded.value[id] = true
}

function levelLabel(item) {
  var map = { top: '985/211', undergraduate: '本科', vocational: '专科' }
  return map[collegeLevel(item)] || '本科'
}

/* 封面类型（无图时按层级分色） */
function coverType(item) {
  return collegeLevel(item)
}

/* 标签配色：学历层次按级配色，特色专业统一蓝 */
function levelTagClass(tag) {
  if (['博士点', '博士'].indexOf(tag) >= 0) return 'tag--phd'
  if (['硕士点', '硕士'].indexOf(tag) >= 0) return 'tag--master'
  if (['本科'].indexOf(tag) >= 0) return 'tag--undergrad'
  return 'tag--feature'
}

function initShort(item) {
  if (item.short_name || item.shortName) return item.short_name || item.shortName
  var name = item.name || ''
  return name.charAt(0) || '院'
}

function specTags(item) {
  if (Array.isArray(item.specialties) && item.specialties.length > 0) return item.specialties
  if (Array.isArray(item.majors)) return item.majors
  if (item.tags) return item.tags
  return ['飞行器设计', '无人机工程']
}

function switchTab(tab) {
  if (currentTab.value === tab) return
  currentTab.value = tab
  page.value = 1
  loadData(true)
}

var searchTimer = null
function onSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(function () { page.value = 1; loadData(true) }, 300)
}

async function loadData(reset) {
  if (reset === undefined) reset = true
  if (reset) { page.value = 1; hasMore.value = true; loading.value = true }
  else { loadingMore.value = true }
  errorMsg.value = ''

  try {
    var params = { page: page.value, page_size: pageSize }
    if (currentTab.value !== 'all') params.type = currentTab.value
    if (keyword.value) params.keyword = keyword.value

    var res = await request({ url: '/api/v1/colleges', data: params })
    var data = Array.isArray(res) ? res : (res && res.data) || res || {}
    var items = Array.isArray(data) ? data : (data && data.items) || []
    var total = (data && data.total) != null ? data.total : items.length

    if (reset) { list.value = items } else { list.value = list.value.concat(items) }
    hasMore.value = list.value.length < total
  } catch (e) {
    errorMsg.value = '加载失败，请稍后重试'
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

function loadMore() {
  if (!loadingMore.value && hasMore.value) { page.value++; loadData(false) }
}

function goDetail(item) {
  uni.navigateTo({ url: '/pages/colleges/detail?id=' + encodeURIComponent(item.id) })
}
function goBack() { uni.navigateBack({ delta: 1 }) }

onLoad(function () { loadData(true) })

onPullDownRefresh(function () {
  loadData(true).then(function () { uni.stopPullDownRefresh() })
})
</script>

<style scoped>
.page {
  --anim-fast: 160ms;
  --anim-base: 240ms;
  --anim-slow: 320ms;
  --ease-out: cubic-bezier(0.25, 0.46, 0.45, 0.94);
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: linear-gradient(180deg, #f5f6f8 0%, #E8F2FC 100%);
  padding-bottom: env(safe-area-inset-bottom);
}

/* ================================================================= */
/* ① Tab + 搜索（原生导航栏，与培训课程列表一致）                       */
/* ================================================================= */
.main-card {
  background: #ffffff;
  border-radius: 0;
  position: relative;
  z-index: 2;
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  animation: pageIn var(--anim-slow) var(--ease-out) both;
}

/* StateView 撑满剩余空间，scroll-view 才能自适应高度 */
.state-fill {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.tabs-container {
  display: flex;
  justify-content: center;
  gap: 16rpx;
  padding: 24rpx 24rpx 0;
}

.tab-item {
  width: 200rpx;
  height: 72rpx;
  line-height: 72rpx;
  text-align: center;
  border-radius: 40rpx;
  font-size: 26rpx;
  font-weight: 400;
  color: #969799;
  background: #ffffff;
  box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.06);
  transition: background-color var(--anim-fast) ease, color var(--anim-fast) ease;
}

.tab-item.active {
  background: #0A66C2;
  color: #ffffff;
  font-weight: 600;
  box-shadow: 0 4rpx 16rpx rgba(10, 31, 68, 0.35);
}

.search-bar {
  margin: 24rpx 24rpx 0;
  background: #fafafa;
  border-radius: 40rpx;
  padding: 16rpx 24rpx;
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.search-input { flex: 1; font-size: 28rpx; color: #17212B; }

/* ================================================================= */
/* ③ 卡片                                                             */
/* ================================================================= */
.list-scroll { padding: 24rpx 24rpx 0; height: auto; flex: 1; min-height: 0; }

.college-card {
  position: relative;
  width: 100%;
  background: #ffffff;
  border-radius: 16rpx;
  border: 1rpx solid #f0f1f3;
  overflow: hidden;
  margin-bottom: 16rpx;
  box-shadow: 0 4rpx 16rpx rgba(10, 31, 68, 0.06);
  box-sizing: border-box;
  animation: cardIn var(--anim-base) var(--ease-out) both;
}

.college-card:nth-child(1) { animation-delay: 60ms; }
.college-card:nth-child(2) { animation-delay: 120ms; }
.college-card:nth-child(3) { animation-delay: 180ms; }

/* 封面（按类型分色） */
.card-cover {
  width: 100%;
  height: 240rpx;
  position: relative;
  overflow: hidden;
}

/* 类型兜底色 */
.cover--top { background: linear-gradient(135deg, #074D92, #0A66C2); }
.cover--undergraduate { background: linear-gradient(135deg, #6D28D9, #8B5CF6); }
.cover--vocational { background: linear-gradient(135deg, #059669, #34c759); }

/* 真实校园图（普通流元素，与赛事列表一致，避免小程序 image 绝对定位裁剪 bug） */
.cover-img {
  width: 100%;
  height: 100%;
  display: block;
  transition: opacity var(--anim-base) ease-out;
}

/* 兜底装饰层 */
.cover-deco-wrap { position: absolute; inset: 0; }

.cover-deco {
  position: absolute;
  right: -40rpx;
  top: -40rpx;
  width: 200rpx;
  height: 200rpx;
  border: 2rpx solid rgba(255, 255, 255, 0.12);
  border-radius: 50%;
}

.cover-deco::after {
  content: '';
  position: absolute;
  inset: 40rpx;
  border: 2rpx solid rgba(255, 255, 255, 0.1);
  border-radius: 50%;
}

/* 左上角类型胶囊（白底，按类型分字色） */
.cover-type-pill {
  position: absolute;
  top: 16rpx;
  left: 16rpx;
  padding: 4rpx 16rpx;
  background: rgba(255, 255, 255, 0.92);
  color: #0A66C2;
  font-size: 22rpx;
  font-weight: 600;
  border-radius: 999rpx;
  z-index: 2;
  box-shadow: 0 2rpx 8rpx rgba(10, 31, 68, 0.08);
}

.cover-type-pill--undergraduate { color: #8B5CF6; }
.cover-type-pill--vocational { color: #059669; }

/* 圆形校徽：半嵌在封面底部边缘（独立层，可溢出 cover） */
.cover-badge {
  position: absolute;
  top: 216rpx; /* 240rpx cover 高 - 24rpx 半嵌 */
  left: 24rpx;
  width: 80rpx;
  height: 80rpx;
  background: #ffffff;
  border: 4rpx solid rgba(10, 102, 194, 0.2);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #0A66C2;
  font-size: 32rpx;
  font-weight: 700;
  box-shadow: 0 4rpx 12rpx rgba(10, 31, 68, 0.12);
  z-index: 2;
}

/* 卡片体 */
.card-body { padding: 56rpx 32rpx 24rpx; }

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8rpx;
  margin-bottom: 4rpx;
}

.college-name { font-size: 34rpx; font-weight: 700; color: #17212B; flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.card-arrow { font-size: 36rpx; color: #c8c9cc; flex-shrink: 0; margin-right: 16rpx; }

.college-location { font-size: 24rpx; color: #969799; display: block; margin-bottom: 16rpx; }

/* 数据栏（三行图标式） */
.stats-list {
  background: #fafafa;
  border-radius: 12rpx;
  padding: 4rpx 28rpx;
  margin-bottom: 16rpx;
}

.stat-line {
  display: flex;
  align-items: center;
  gap: 12rpx;
  padding: 16rpx 0;
  border-bottom: 1rpx solid #ebedf0;
}

.stat-line-last { border-bottom: none; }

.stat-icon {
  width: 40rpx;
  height: 40rpx;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-icon--major { background: rgba(10, 102, 194, 0.1); }
.stat-icon--partner { background: rgba(139, 92, 246, 0.12); }
.stat-icon--student { background: rgba(52, 199, 89, 0.12); }

.stat-icon-text { font-size: 24rpx; font-weight: 600; color: #0A66C2; }

.stat-icon--partner .stat-icon-text { color: #8B5CF6; }
.stat-icon--student .stat-icon-text { color: #34c759; }

.stat-label { flex: 1; font-size: 26rpx; color: #969799; }
.stat-value { font-size: 30rpx; font-weight: 700; color: #17212B; margin-right: 8rpx; }

/* 简介（最多 2 行） */
.college-intro {
  font-size: 26rpx;
  color: #969799;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin-bottom: 16rpx;
}

/* 标签（配色统一） */
.tag-row { display: flex; flex-wrap: wrap; gap: 10rpx; }
.spec-tag { padding: 4rpx 16rpx; border-radius: 999rpx; font-size: 22rpx; font-weight: 500; }

.tag--feature { background: #E8F2FC; color: #0A66C2; }
.tag--phd { background: #FFF4E6; color: #E96012; }
.tag--master { background: #F5F3FF; color: #8B5CF6; }
.tag--undergrad { background: #E8F5E9; color: #34c759; }

/* 加载更多 */
.load-more-wrap { text-align: center; padding: 20rpx 0; }
.loading-inline { display: flex; align-items: center; justify-content: center; gap: 8rpx; font-size: 24rpx; color: #969799; }
.no-more { font-size: 24rpx; color: #969799; }

/* ================================================================= */
/* 动效                                                              */
/* ================================================================= */
@keyframes pageIn {
  from { opacity: 0; transform: translateY(12px); }
  to   { opacity: 1; transform: translateY(0); }
}

@keyframes cardIn {
  from { opacity: 0; transform: translateY(16px); }
  to   { opacity: 1; transform: translateY(0); }
}

@keyframes twinkle {
  0%, 100% { opacity: 0.2; }
  50%      { opacity: 0.8; }
}

.press-feedback {
  transform: scale(0.98);
  opacity: 0.92;
}

@media (prefers-reduced-motion: reduce) {
  .main-card, .college-card {
    animation: none !important;
    transition: none !important;
  }
}
</style>
