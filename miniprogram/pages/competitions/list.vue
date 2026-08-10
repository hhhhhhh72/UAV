<template>
  <view class="page">
    <!-- ① Tab + 搜索（原生导航栏，顶部与其他页面统一） -->
    <view class="main-card">
      <view class="tabs-container">
        <view
          class="tab-item"
          :class="{ active: currentTab === 'all' }"
          @click="switchTab('all')"
        >全部赛事</view>
        <view
          class="tab-item"
          :class="{ active: currentTab === 'enrolling' }"
          @click="switchTab('enrolling')"
        >报名中</view>
      </view>

      <view class="search-bar">
        <u-icon name="search" size="28rpx" color="#0A66C2" />
        <input
          class="search-input"
          v-model="keyword"
          placeholder="搜索赛事名称"
          @input="onSearch"
        />
      </view>

      <!-- ③ 赛事卡片列表 -->
      <StateView
        class="state-fill"
        :loading="loading"
        :error="!!errorMsg"
        :empty="!loading && !errorMsg && list.length === 0"
        empty-text="暂无赛事"
        @retry="loadData"
      >
        <scroll-view class="list-scroll" scroll-y @scrolltolower="loadMore">
          <view
            v-for="item in list"
            :key="item.id"
            class="card"
            :class="{ 'card--closed': isClosed(item) }"
            hover-class="press-feedback"
            :hover-stay-time="120"
            @click="goDetail(item)"
          >
            <!-- 封面图区 -->
            <view class="card-cover" :class="'cover--' + thumbType(item)">
              <!-- 真实赛事海报图（有则显示） -->
              <image
                v-if="coverOf(item)"
                :src="coverOf(item)"
                class="cover-img"
                mode="aspectFill"
                lazy-load
                @load="onPosterLoad(item.id)"
                :style="{ opacity: imgLoaded[item.id] ? 1 : 0 }"
              />
              <!-- 无图兜底：类型渐变色块 + 简称 + 状态 -->
              <view v-else class="cover-fallback">
                <view class="cover-glow" />
                <text class="cover-char">{{ thumbChar(item) }}</text>
                <text class="cover-caption">{{ thumbCaption(item) }}</text>
              </view>

              <!-- 底部渐变蒙层 + 类型简称 -->
              <view class="cover-mask">
                <text class="cover-mask-char">{{ thumbChar(item) }}</text>
              </view>

              <!-- 左上角类型胶囊 -->
              <view class="type-pill">{{ thumbCaption(item) }}</view>
            </view>

            <!-- 信息区 -->
            <view class="card-info">
              <view class="info-top">
                <text class="card-title">{{ item.title || item.name || '未知赛事' }}</text>
                <view class="status-badge" :class="statusClass(item.status)">
                  <text class="status-text">{{ statusText(item.status) }}</text>
                </view>
              </view>

              <view class="card-meta">
                <text class="meta-line meta-line--date">📅 {{ fmtDate(item.start_date) }} - {{ fmtDate(item.end_date) }}</text>
                <text class="meta-line meta-line--loc">📍 {{ item.location || '待定' }}</text>
                <text class="meta-line meta-line--org">{{ item.organizer || item.sponsor || '待定' }}</text>
              </view>

              <view class="card-tags">
                <text v-for="t in compTags(item)" :key="t" class="pill" :class="tagTypeClass(t)">{{ t }}</text>
              </view>

              <view class="card-bottom">
                <view v-if="isFree(item)" class="free-badge">免费</view>
                <view v-else class="price-cap">
                  <text class="price-symbol">¥</text>
                  <text class="price-num">{{ compFee(item).toLocaleString() }}</text>
                  <text class="price-suffix">/人</text>
                </view>
                <view v-if="!isClosed(item)" class="btn-enroll" hover-class="press-feedback" :hover-stay-time="120" @click.stop="goRegister(item)">
                  立即报名
                </view>
                <text v-else class="closed-label">已截止</text>
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

    <!-- 底部无人机剪影装饰 -->
    <view class="drone-decor" />
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

/* ===== 状态 ===== */
function isClosed(item) {
  return item.status === 'closed' || item.status === 'full'
}

function statusText(item) {
  var map = { enrolling: '报名中', open: '报名中', ongoing: '进行中', closed: '已截止', full: '已满员' }
  return map[item.status] || (item.registration_status === 'open' ? '报名中' : '报名中')
}

function statusClass(item) {
  if (item.status === 'ongoing') return 'badge--ongoing'
  if (isClosed(item)) return 'badge--closed'
  return 'badge--enrolling'
}

/* ===== 数据映射 ===== */

function fmtDate(d) {
  if (!d) return '待定'
  if (String(d).indexOf('.') >= 0 || String(d).indexOf('年') >= 0) return String(d)
  return String(d).slice(0, 10)
}

function compTags(item) {
  if (Array.isArray(item.tags) && item.tags.length > 0) return item.tags.slice(0, 3)
  var tags = []
  if (item.category) tags.push(item.category)
  if (tags.length === 0) tags = ['多旋翼', '国家级']
  return tags
}

function compFee(item) {
  if (item.fee != null) return item.fee
  if (item.price_fen != null) return item.price_fen / 100
  if (item.price != null) return item.price
  return 380
}

function isFree(item) {
  return compFee(item) <= 0
}

/* 分类类型：竞技=orange, FPV=purple, 创新=teal */
function thumbType(item) {
  var t = item.title || ''
  if (t.indexOf('FPV') >= 0 || t.indexOf('竞速') >= 0) return 'purple'
  if (t.indexOf('创新') >= 0 || t.indexOf('应用') >= 0) return 'teal'
  return 'orange'
}

function thumbChar(item) {
  var t = item.title || ''
  if (t.indexOf('全国') >= 0) return '国'
  if (t.indexOf('西南') >= 0) return '西'
  if (t.indexOf('国际') >= 0) return '国'
  if (t.indexOf('青少年') >= 0) return '青'
  if (t.indexOf('贵州') >= 0) return '贵'
  if (t.indexOf('创新') >= 0) return '创'
  return '赛'
}

function thumbCaption(item) {
  var map = { enrolling: '报名中', open: '报名中', ongoing: '进行中', closed: '已结束', full: '已满' }
  return map[item.status] || ''
}

/** 封面图 URL：兼容 poster / cover / image / banner */
function coverOf(item) {
  const u = item.poster || item.cover || item.image || item.banner
  return u ? u : ''
}

/** 海报图加载完成淡入 */
function onPosterLoad(id) {
  imgLoaded.value[id] = true
}

function tagTypeClass(tag) {
  if (['国家级', '国际赛', '国际', '航空级'].indexOf(tag) >= 0) return 'pill--level'
  return 'pill--model'
}

/* ===== 数据获取 ===== */

async function loadData(reset) {
  if (reset === undefined) reset = true
  if (reset) { page.value = 1; hasMore.value = true; loading.value = true }
  else { loadingMore.value = true }
  errorMsg.value = ''

  try {
    var params = { page: page.value, page_size: pageSize }
    if (currentTab.value === 'enrolling') params.status = 'enrolling'
    if (keyword.value) params.keyword = keyword.value

    var res = await request({ url: '/api/v1/competitions', data: params })
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

/* ===== 筛选 & 搜索 ===== */

function switchTab(tab) {
  if (currentTab.value === tab) return
  currentTab.value = tab
  loadData(true)
}

var searchTimer = null
function onSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(function () { loadData(true) }, 300)
}

/* ===== 交互 ===== */

function goDetail(item) {
  uni.navigateTo({ url: '/pages/competitions/detail?id=' + encodeURIComponent(item.id) })
}

function goRegister(item) {
  uni.navigateTo({ url: '/pages/competitions/register?id=' + encodeURIComponent(item.id) })
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
  position: relative;
  overflow: hidden;
}

/* ================================================================= */
/* ① Tab + 搜索（原生导航栏，无自定义头部）                            */
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
  width: 320rpx;
  height: 72rpx;
  line-height: 72rpx;
  text-align: center;
  border-radius: 999rpx;
  font-size: 28rpx;
  font-weight: 400;
  color: #969799;
  background: #ffffff;
  box-shadow: 0 2rpx 8rpx rgba(10, 31, 68, 0.06);
  transition: background-color var(--anim-fast) ease, color var(--anim-fast) ease;
}

.tab-item.active {
  background: #0A66C2;
  color: #ffffff;
  font-weight: 600;
  box-shadow: 0 4rpx 16rpx rgba(10, 31, 68, 0.35);
}

/* 搜索框 */
.search-bar {
  margin: 24rpx 24rpx 0;
  background: rgba(245, 248, 252, 0.8);
  border: 1rpx solid rgba(10, 102, 194, 0.12);
  border-radius: 999rpx;
  padding: 16rpx 24rpx;
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.search-input { flex: 1; font-size: 28rpx; color: #17212B; }

/* ================================================================= */
/* ③ 卡片                                                             */
/* ================================================================= */
.list-scroll { padding: 24rpx 24rpx 0; height: auto; flex: 1; min-height: 0; box-sizing: border-box; }

.card {
  background: #ffffff;
  border: 1rpx solid rgba(10, 31, 68, 0.06);
  border-radius: 16rpx;
  overflow: hidden;
  margin-bottom: 16rpx;
  box-shadow: 0 8rpx 24rpx rgba(10, 31, 68, 0.08);
  box-sizing: border-box;
  position: relative;
  animation: cardIn var(--anim-base) var(--ease-out) both;
  transition: transform var(--anim-fast) ease, box-shadow var(--anim-fast) ease;
}

.card:nth-child(1) { animation-delay: 60ms; }
.card:nth-child(2) { animation-delay: 120ms; }
.card:nth-child(3) { animation-delay: 180ms; }

/* 已截止卡片降透明度 */
.card--closed { opacity: 0.9; }

/* ===== 封面图区 ===== */
.card-cover {
  position: relative;
  width: 100%;
  height: 240rpx;
  overflow: hidden;
}

/* 分类兜底色：竞技=橙、FPV=紫、创新=青 */
.cover--orange { background: linear-gradient(135deg, #074D92, #F97316); }
.cover--purple { background: linear-gradient(135deg, #4C1D95, #DB2777); }
.cover--teal { background: linear-gradient(135deg, #065F46, #06B6D4); }

.cover-img {
  width: 100%;
  height: 100%;
  display: block;
  transition: opacity var(--anim-base) ease-out;
}

/* 无图兜底：渐变色块 + 简称 + 状态 */
.cover-fallback {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8rpx;
}

.cover-glow {
  position: absolute;
  top: -40rpx;
  right: -40rpx;
  width: 160rpx;
  height: 160rpx;
  background: radial-gradient(circle, rgba(0, 229, 255, 0.4), transparent);
}

.cover-char {
  font-size: 56rpx;
  font-weight: 700;
  color: #ffffff;
  position: relative;
  z-index: 1;
}

.cover-caption {
  font-size: 22rpx;
  color: rgba(255, 255, 255, 0.85);
  position: relative;
  z-index: 1;
}

/* 底部渐变蒙层 + 类型简称 */
.cover-mask {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 72rpx;
  padding: 0 16rpx;
  display: flex;
  align-items: flex-end;
  justify-content: flex-start;
  background: linear-gradient(180deg, rgba(10, 31, 68, 0) 0%, rgba(10, 31, 68, 0.55) 100%);
  pointer-events: none;
}

.cover-mask-char {
  font-size: 13px;
  font-weight: 600;
  color: #ffffff;
  padding-bottom: 8rpx;
}

/* 左上角类型胶囊 */
.type-pill {
  position: absolute;
  top: 10rpx;
  left: 10rpx;
  padding: 2px 10px;
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.92);
  font-size: 11px;
  font-weight: 600;
  color: #0A66C2;
}

/* 信息区 */
.card-info {
  padding: 14rpx 16rpx 16rpx;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4rpx;
}

.info-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 8rpx;
}

.card-title {
  font-size: 17px;
  font-weight: 700;
  color: #17212B;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  line-height: 1.4;
}

/* 状态徽章（3D 凸起胶囊） */
.status-badge {
  padding: 4rpx 12rpx;
  border-radius: 999rpx;
  font-size: 10px;
  font-weight: 600;
  flex-shrink: 0;
  box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.15);
}

.badge--enrolling {
  background: linear-gradient(135deg, #F97316, #E96012);
  color: #ffffff;
  animation: badgePulse 2s ease-in-out infinite;
}

.badge--ongoing {
  background: linear-gradient(135deg, #00E5FF, #0A66C2);
  color: #ffffff;
  animation: badgePulse 2s ease-in-out infinite;
}

.badge--closed {
  background: #CBD5E1;
  color: #64748B;
}

.card-meta {
  display: flex;
  flex-direction: column;
  gap: 2rpx;
  margin-top: 2rpx;
}

.meta-line {
  font-size: 12px;
  color: #969799;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.6;
}

.meta-line--date::before { content: '· '; color: #0A66C2; }
.meta-line--loc::before { content: '· '; color: #00E5FF; }
.meta-line--org { color: #94A3B8; }

/* 标签 pills */
.card-tags {
  display: flex;
  gap: 8rpx;
  margin-top: 4rpx;
  overflow: hidden;
}

.pill {
  padding: 2rpx 12rpx;
  border-radius: 999rpx;
  font-size: 10px;
  font-weight: 500;
  white-space: nowrap;
}

.pill--model {
  background: rgba(10, 102, 194, 0.08);
  color: #0A66C2;
  border: 1rpx solid rgba(10, 102, 194, 0.2);
}

.pill--level {
  background: rgba(255, 142, 60, 0.08);
  color: #E96012;
  border: 1rpx solid rgba(255, 142, 60, 0.2);
}

/* 底部：价格胶囊 + 按钮 */
.card-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 6rpx;
  padding-top: 10rpx;
  border-top: 1rpx solid rgba(10, 31, 68, 0.06);
}

.price-cap {
  display: flex;
  align-items: baseline;
}

/* 免费徽章 */
.free-badge {
  padding: 4rpx 14rpx;
  background: rgba(52, 199, 89, 0.1);
  border: 1rpx solid rgba(52, 199, 89, 0.3);
  color: #34c759;
  font-size: 12px;
  font-weight: 700;
  border-radius: 999rpx;
}

.price-symbol {
  font-size: 14px;
  color: #E96012;
  font-weight: 700;
}

.price-num {
  font-size: 22px;
  font-weight: 800;
  background: linear-gradient(135deg, #F97316, #E96012);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  line-height: 1;
}

.price-suffix {
  font-size: 10px;
  color: #98A2B3;
  margin-left: 4rpx;
}

.btn-enroll {
  padding: 8rpx 20rpx;
  background: linear-gradient(135deg, #074D92, #0A66C2);
  color: #ffffff;
  font-size: 12px;
  font-weight: 600;
  border-radius: 50rpx;
  box-shadow: 0 4rpx 12rpx rgba(10, 102, 194, 0.3);
  transition: transform var(--anim-fast) ease, opacity var(--anim-fast) ease;
}

.closed-label {
  font-size: 12px;
  color: #94A3B8;
  font-weight: 500;
}

/* 加载更多 */
.load-more-wrap { text-align: center; padding: 20rpx 0; }
.loading-inline { display: flex; align-items: center; justify-content: center; gap: 8rpx; font-size: 24rpx; color: #969799; }
.no-more { font-size: 24rpx; color: #969799; }

/* 底部无人机剪影装饰 */
.drone-decor {
  position: fixed;
  right: -30rpx;
  bottom: 30rpx;
  width: 200rpx;
  height: 160rpx;
  opacity: 0.03;
  pointer-events: none;
  background: radial-gradient(ellipse at center, #074D92 0%, transparent 70%);
  z-index: 1;
}

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

@keyframes badgePulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(255, 142, 60, 0.4); }
  50% { box-shadow: 0 0 0 8rpx rgba(255, 142, 60, 0); }
}

.press-feedback {
  transform: scale(0.98);
  opacity: 0.92;
}

@media (prefers-reduced-motion: reduce) {
  .main-card, .card, .btn-enroll {
    animation: none !important;
    transition: none !important;
  }
}
</style>
