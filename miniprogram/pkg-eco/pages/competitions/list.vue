<template>
  <view class="page" :class="{ 'no-motion': noMotion }">
    <!-- ① 白底头部：搜索 + Tab 筛选 + 发布入口 -->
    <view class="head-zone">
      <!-- 搜索框（白上白：双层投影浮起） -->
      <view class="sbar">
        <view class="b-search">
          <view class="b-search-ic"><view class="ic-ring" /><view class="ic-bar" /></view>
          <input
            class="b-sinp"
            v-model="keyword"
            placeholder="搜索赛事名称"
            placeholder-class="b-ph"
            confirm-type="search"
            @input="onSearch"
          />
          <text v-if="keyword" class="b-sclr" @tap="clearSearch">×</text>
          <view class="b-sep" />
          <text class="b-sbtn" @tap="searchNow">搜索</text>
        </view>
      </view>

      <!-- 筛选胶囊：全部赛事 / 报名中 + 发布入口（仅企业可见） -->
      <view class="fbar">
        <view
          class="fpill"
          :class="{ on: currentTab === 'all' }"
          hover-class="fpill-press"
          :hover-stay-time="100"
          @tap="switchTab('all')"
        >
          <text class="fpv">全部赛事</text>
        </view>
        <view
          class="fpill"
          :class="{ on: currentTab === 'enrolling' }"
          hover-class="fpill-press"
          :hover-stay-time="100"
          @tap="switchTab('enrolling')"
        >
          <text class="fpv">报名中</text>
        </view>
        <view
          v-if="canPublish"
          class="publish-entry"
          hover-class="publish-press"
          :hover-stay-time="100"
          @tap="goPublish"
        >＋ 发布赛事</view>
      </view>
    </view>

    <!-- ② 信息行 -->
    <view class="ir">
      <text>共 <text class="irn">{{ list.length }}</text> 个赛事</text>
      <text class="ir-hint">{{ currentTab === 'enrolling' ? '报名中' : '全部赛事' }}</text>
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
          :hover-stay-time="100"
          @tap="goDetail(item)"
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
              <text class="meta-line">{{ fmtDate(item.start_date) }} - {{ fmtDate(item.end_date) }}</text>
              <text class="meta-line">{{ item.location || '地点待定' }}</text>
              <text class="meta-line meta-line--org">{{ item.organizer || item.sponsor || '主办方待定' }}</text>
            </view>

            <view class="card-tags">
              <text v-for="t in compTags(item)" :key="t" class="pill" :class="tagTypeClass(t)">{{ t }}</text>
            </view>

            <view class="card-bottom">
              <view v-if="isFree(item)" class="free-badge">免费</view>
              <view v-else class="price-cap">
                <text class="price-symbol">¥</text>
                <text class="price-num">{{ compFee(item).toLocaleString() }}</text>
                <text v-if="origFee(item)" class="price-orig">¥{{ origFee(item).toLocaleString() }}</text>
                <text class="price-suffix">/人</text>
              </view>
              <view v-if="!isClosed(item)" class="btn-enroll" hover-class="press-feedback" :hover-stay-time="100" @tap.stop="goRegister(item)">
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
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { useReduceMotion } from '../../../utils/motion'
import { request, getStoredUser } from '../../../utils/request'
import StateView from '../../../components/StateView.vue'

const { noMotion, checkMotion } = useReduceMotion()
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

/* ===== 发布入口（仅企业账号） ===== */
const canPublish = computed(() => {
  const u = getStoredUser()
  return !!(u && (u.role === 'enterprise' || u.user_type === 'enterprise'))
})

function goPublish() {
  uni.navigateTo({ url: '/pkg-eco/pages/competitions/publish' })
}

/* ===== 状态 ===== */
/* 复用详情页单一状态源：status 与报名截止时间统一判定（deadline 已过同样视为已截止） */
function deadlineDate(item) {
  const d = item.deadline || item.enroll_deadline
  if (!d) return null
  const t = Date.parse(String(d).replace(/-/g, '/'))
  return isNaN(t) ? null : new Date(t)
}

function isClosed(item) {
  if (item.status === 'closed' || item.status === 'full') return true
  const dl = deadlineDate(item)
  return !!(dl && dl.getTime() < Date.now())
}

function statusText(item) {
  var map = { enrolling: '报名中', open: '报名中', ongoing: '进行中', closed: '已截止', full: '已满员' }
  return map[item.status] || '报名中'
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

/* 划线原价：original_fee > 现价才显示 */
function origFee(item) {
  var o = item.original_fee
  if (o == null && item.original_price != null) o = item.original_price
  return o && o > compFee(item) ? o : null
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
function searchNow() {
  clearTimeout(searchTimer)
  loadData(true)
}
function clearSearch() {
  keyword.value = ''
  loadData(true)
}

/* ===== 交互 ===== */

function goDetail(item) {
  uni.navigateTo({ url: '/pkg-eco/pages/competitions/detail?id=' + encodeURIComponent(item.id) })
}

function goRegister(item) {
  uni.navigateTo({ url: '/pkg-eco/pages/competitions/register?id=' + encodeURIComponent(item.id) })
}

onLoad(function () {
  checkMotion()
  loadData(true)
})

onPullDownRefresh(function () {
  loadData(true).then(function () { uni.stopPullDownRefresh() })
})
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: #fff;
  padding-bottom: env(safe-area-inset-bottom);
  display: flex;
  flex-direction: column;
}

/* ===== 白底头部 ===== */
.head-zone { background: #fff; }

/* ===== 搜索框：白上白——纯白填充 + 灰描边 + 双层投影 ===== */
.sbar { padding: 12px 12px 8px; background: #fff; }
.b-search {
  height: 44px;
  padding: 0 11px;
  border: 1px solid #E4E7EC;
  border-radius: 7px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(16, 24, 40, 0.06), 0 4px 12px rgba(16, 24, 40, 0.05);
  display: flex;
  align-items: center;
  gap: 7px;
  box-sizing: border-box;
}
.b-search-ic { position: relative; width: 15px; height: 15px; flex: none; }
.ic-ring {
  width: 9px; height: 9px;
  border: 1.5px solid #98A2B3;
  border-radius: 50%;
  position: absolute; top: 0; left: 0;
}
.ic-bar {
  position: absolute; right: 0; bottom: 1px;
  width: 5px; height: 1.5px;
  background: #98A2B3;
  transform: rotate(45deg);
}
.b-sinp { flex: 1; min-width: 0; background: transparent; font-size: 13px; color: #17212B; }
.b-ph { color: #667085; }
.b-sclr { color: #667085; font-size: 15px; padding: 10px; margin: -10px; }
.b-sep { width: 1px; height: 15px; background: #DDE1E6; margin: 0 9px 0 6px; flex: none; }
.b-sbtn { flex: none; color: #344054; font-size: 13px; line-height: 1; padding: 6px 2px 6px 0; }

/* ===== 筛选胶囊 + 发布入口 ===== */
.fbar { display: flex; gap: 8px; padding: 10px 12px 4px; background: #fff; }
.fpill {
  flex: 1;
  min-width: 0;
  min-height: 40px;
  border: 1px solid #E4E7EC;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(16, 24, 40, 0.04), 0 3px 10px rgba(16, 24, 40, 0.04);
  color: #344054;
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  overflow: hidden;
  transition: transform .2s ease, border-color .2s ease, background .2s ease, color .2s ease;
}
.fpill.on { border-color: #0A66C2; color: #0A66C2; font-weight: 600; background: #F4F8FC; }
.fpill-press { transform: scale(0.95); opacity: 0.85; }
.fpv { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.publish-entry {
  flex: none;
  min-height: 40px;
  padding: 0 14px;
  border-radius: 8px;
  background: #0A66C2;
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  display: flex;
  align-items: center;
  box-shadow: 0 2px 8px rgba(10, 102, 194, 0.24);
}
.publish-press { transform: scale(0.95); opacity: 0.9; }

/* ===== 信息行 ===== */
.ir {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 14px 8px;
  font-size: 12px;
  color: #667085;
  animation: fadeUp .25s ease-out backwards;
  animation-delay: 60ms;
}
.irn { color: #0A66C2; font-weight: 600; }
.ir-hint { font-size: 12px; color: #98A2B3; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: translateY(0); } }

/* ===== StateView 撑满剩余空间 ===== */
.state-fill {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

/* ===== 卡片 ===== */
.list-scroll { padding: 0 12px; height: auto; flex: 1; min-height: 0; box-sizing: border-box; }

.card {
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
  overflow: hidden;
  margin-bottom: 10px;
  box-shadow: 0 4px 20px rgba(16, 24, 40, 0.06);
  box-sizing: border-box;
  position: relative;
  transition: transform .35s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease;
}
.card:nth-child(-n+6) { animation: cardIn .22s ease-out backwards; }
.card:nth-child(1) { animation-delay: 80ms; }
.card:nth-child(2) { animation-delay: 100ms; }
.card:nth-child(3) { animation-delay: 120ms; }
.card:nth-child(4) { animation-delay: 140ms; }
.card:nth-child(5) { animation-delay: 160ms; }
.card:nth-child(6) { animation-delay: 180ms; }
@keyframes cardIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
.press-feedback { transform: scale(0.98); opacity: 0.92; }

/* 已截止卡片降透明度 */
.card--closed { opacity: 0.9; }

/* ===== 封面图区 ===== */
.card-cover {
  position: relative;
  width: 100%;
  height: 120px;
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
  transition: opacity .25s ease-out;
}

/* 无图兜底：渐变色块 + 简称 + 状态 */
.cover-fallback {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
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
  font-size: 28px;
  font-weight: 700;
  color: #ffffff;
  position: relative;
  z-index: 1;
}

.cover-caption {
  font-size: 11px;
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
  height: 36px;
  padding: 0 8px;
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
  padding-bottom: 4px;
}

/* 左上角类型胶囊 */
.type-pill {
  position: absolute;
  top: 8px;
  left: 8px;
  padding: 2px 10px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.92);
  font-size: 11px;
  font-weight: 600;
  color: #0A66C2;
}

/* ===== 信息区 ===== */
.card-info {
  padding: 10px 12px 12px;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.info-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 8px;
}

.card-title {
  font-size: 15px;
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

/* 状态徽章（静色胶囊，无脉冲） */
.status-badge {
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 10px;
  font-weight: 600;
  flex-shrink: 0;
}

.badge--enrolling { background: #E9F7F0; color: #0B6B41; }
.badge--ongoing { background: #EAF3FB; color: #0A66C2; }
.badge--closed { background: #EEF1F4; color: #5D6B82; }

.card-meta {
  display: flex;
  flex-direction: column;
  gap: 1px;
  margin-top: 2px;
}

.meta-line {
  font-size: 12px;
  color: #667085;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.6;
}

.meta-line--org { color: #98A2B3; }

/* 标签 pills */
.card-tags {
  display: flex;
  gap: 6px;
  margin-top: 4px;
  overflow: hidden;
}

.pill {
  padding: 1px 8px;
  border-radius: 4px;
  font-size: 10px;
  font-weight: 500;
  white-space: nowrap;
}

.pill--model { background: #EAF3FB; color: #0A66C2; }
.pill--level { background: #FFF4EC; color: #E96012; }

/* 底部：价格 + 按钮 */
.card-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 4px;
  padding-top: 8px;
  border-top: 1px solid #F0F1F3;
}

.price-cap {
  display: flex;
  align-items: baseline;
}

/* 免费徽章 */
.free-badge {
  padding: 2px 10px;
  background: #E9F7F0;
  border: 1px solid #C9EEDC;
  color: #0B6B41;
  font-size: 11px;
  font-weight: 600;
  border-radius: 999px;
}

.price-symbol { font-size: 13px; color: #C2410C; font-weight: 700; }
.price-num { font-size: 20px; font-weight: 800; color: #C2410C; line-height: 1; }
.price-orig {
  font-size: 11px;
  color: #98A2B3;
  text-decoration: line-through;
  margin-left: 4px;
}
.price-suffix { font-size: 10px; color: #98A2B3; margin-left: 2px; }

.btn-enroll {
  padding: 6px 16px;
  background: #0A66C2;
  color: #ffffff;
  font-size: 12px;
  font-weight: 600;
  border-radius: 999px;
  box-shadow: 0 2px 8px rgba(10, 102, 194, 0.24);
  transition: transform .35s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease;
}

.closed-label { font-size: 12px; color: #98A2B3; font-weight: 500; }

/* 加载更多 */
.load-more-wrap { text-align: center; padding: 10px 0; }
.loading-inline { display: flex; align-items: center; justify-content: center; gap: 8px; font-size: 12px; color: #667085; }
.no-more { font-size: 12px; color: #98A2B3; }

/* ===== 减弱动效（无障碍） ===== */
.page.no-motion .card,
.page.no-motion .ir { animation: none; }
</style>
