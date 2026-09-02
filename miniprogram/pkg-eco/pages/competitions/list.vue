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

      <!-- 筛选分段（对齐成果库：下划线 tab + ▾ 状态浮层面板）+ 发布入口同排（tab 左 / 按钮右） -->
      <view class="stage-wrap">
        <view class="stage-line">
          <view class="stages">
            <view
              v-for="t in STATUS_TABS"
              :key="t.value"
              class="stg"
              :class="{ on: currentTab === t.value }"
              @tap="pickStageTab(t.value)"
            >
              <text>{{ t.label }}</text>
              <!-- ▾ 独立面板开关（方案 A）：未停在「全部」时点「全部」先清筛；已停时再点开面板 -->
              <text v-if="t.value === 'all'" class="stg-arr" :class="{ up: panel === 'all' }" @tap.stop="togglePanel">▾</text>
            </view>
          </view>
          <!-- 发布入口（仅企业可见）：与筛选分段同一行，右端 -->
          <view v-if="canPublish" class="publish-entry" hover-class="publish-press" :hover-stay-time="100" @tap="goPublish">＋ 发布赛事</view>
        </view>
        <!-- 状态面板：absolute 浮层（同成果库），展开时不挤动下方内容 -->
        <view v-if="panel === 'all'" class="field-panel" :class="{ closing }">
          <view class="p-group">状态</view>
          <view class="p-chips">
            <text
              v-for="t in STATUS_OPTS"
              :key="t.value"
              class="p-chip"
              :class="{ act: currentTab === t.value }"
              @tap="pickStatus(t.value)"
            >{{ t.label }}</text>
          </view>
        </view>
      </view>
    </view>
    <!-- 蒙层：从 tab 分段底部开始置灰（top 由 maskTop 实测），点击外部退场收起 -->
    <view v-if="panel" class="panel-mask" :style="{ top: maskTop + 'px' }" @tap="startClosePanel" />

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
import { request, getStoredUser, requireLogin } from '../../../utils/request'
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

/* ===== 筛选状态（对齐成果库：tab 分段 + ▾ 状态浮层面板 + 蒙层） ===== */
const panel = ref('')       // '' = 收起；'all' = 状态面板展开
const closing = ref(false)  // 面板退场中（先播退场动画再 v-if 移除）
const maskTop = ref(0)      // 蒙层起点（面板打开时实测：tab 分段底部）
let panelCloseT = null
const PANEL_CLOSE_MS = 210 // 退场动画 .21s ease-in（= 进场 ×0.7）
const STATUS_TABS = [
  { label: '全部赛事', value: 'all' },
  { label: '报名中', value: 'enrolling' },
]
const STATUS_OPTS = STATUS_TABS // 面板 chips 与一级 tab 同维（状态维度，两组即全部）

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

// ---- 筛选交互（tab 分段 + 「全部」状态面板，方案 A，同成果库） ----
function measureMaskTop() {
  uni.createSelectorQuery().select('.stage-wrap').boundingClientRect(function (rect) {
    if (rect && rect.bottom) maskTop.value = Math.round(rect.bottom)
  }).exec()
}
function startClosePanel() {
  if (closing.value) return // 已在退场中，防重复触发叠加定时器
  closing.value = true
  clearTimeout(panelCloseT)
  panelCloseT = setTimeout(function () { panel.value = ''; closing.value = false; panelCloseT = null }, PANEL_CLOSE_MS)
}
function togglePanel() {
  if (panel.value === 'all') { startClosePanel(); return } // 再点「全部」→ 退场收起
  clearTimeout(panelCloseT); panelCloseT = null; closing.value = false
  panel.value = 'all'
  measureMaskTop()
}
// 方案 A（同成果库）：非全部 tab 再点取消；「全部」未停时先清筛、已停时开面板；▾ 独立开关
function pickStageTab(k) {
  if (k !== 'all') {
    startClosePanel()
    currentTab.value = currentTab.value === k ? 'all' : k
    loadData(true)
    return
  }
  if (currentTab.value !== 'all') {
    startClosePanel()
    currentTab.value = 'all'
    loadData(true)
    return
  }
  togglePanel()
}
// 面板 chip 点选即筛（选中即时高亮），再点一次取消；「全部」chip 恒为全部
function pickStatus(v) {
  currentTab.value = (v === 'all' || currentTab.value !== v) ? v : 'all'
  startClosePanel()
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
  // 登录拦截：未登录点「立即报名」→ 跳登录页（列表与详情行为一致）
  if (!requireLogin('请先登录后再报名')) return
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

/* ===== 筛选分段（对齐成果库：下划线 tab + ▾ 状态面板 + 蒙层）+ 发布入口同排 ===== */
.stage-wrap {
  position: relative;
  z-index: 42;
  background: #fff;
  padding: 2px 12px 0;
}
/* 同排行：tab 左（仅 2 个 tab 单行放得下），发布按钮右 */
.stage-line {
  display: flex;
  align-items: center;
  gap: 12px;
}
.stages {
  display: flex;
  gap: 40rpx;
  padding: 4rpx 16rpx 16rpx 4rpx;
  white-space: nowrap;
  flex: 1;
  min-width: 0;
}
.stg {
  position: relative;
  flex-shrink: 0;
  min-height: 88rpx;
  display: flex;
  align-items: center;
  gap: 4rpx;
  padding: 0 8rpx;
  font-size: 24rpx;
  color: #667085;
}
.stg.on { color: #074D92; font-weight: 600; }
.stg.on::after {
  content: '';
  position: absolute;
  left: 8rpx;
  right: 8rpx;
  bottom: 16rpx;
  height: 3rpx;
  border-radius: 2rpx;
  background: #074D92;
  animation: toc-in .22s ease-out;
}
@keyframes toc-in { from { transform: scaleX(0); } to { transform: scaleX(1); } }
.stg-arr {
  font-size: 24rpx;
  color: #667085;
  transition: transform .2s ease, color .2s ease;
  padding: 20rpx 16rpx;
  margin: -20rpx -16rpx;
}
.stg-arr.up { transform: rotate(180deg); color: #074D92; }

/* 状态面板：absolute 浮层（同成果库），展开时不挤动下方内容 */
.field-panel {
  position: absolute;
  left: 0;
  right: 0;
  top: 100%;
  z-index: 43;
  background: #fff;
  border-radius: 0 0 12px 12px;
  box-shadow: 0 12px 24px rgba(16, 24, 40, 0.08);
  padding: 12px 14px 14px;
  max-height: 62vh;
  overflow-y: auto;
  animation: panelIn .3s cubic-bezier(.32, .72, 0, 1);
}
.field-panel.closing { animation: panelOut .21s ease-in forwards; }
@keyframes panelOut {
  from { opacity: 1; transform: translateY(0); }
  to { opacity: 0; transform: translateY(-10px); }
}
@keyframes panelIn {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}
.field-panel .p-group { font-size: 13px; font-weight: 700; color: #344054; margin: 12px 0 6px; }
.field-panel .p-group:first-child { margin-top: 0; }
.p-chips { display: flex; flex-wrap: wrap; gap: 8px; }
.p-chip {
  min-height: 40px;
  padding: 0 13px;
  border: 1px solid #E4E7EC;
  border-radius: 6px;
  background: #fff;
  color: #667085;
  font-size: 13px;
  display: inline-flex;
  align-items: center;
}
.p-chip.act { color: #fff; border-color: #074D92; background: #074D92; font-weight: 600; }
.p-chip { transition: background .2s ease, border-color .2s ease, color .2s ease, transform .3s cubic-bezier(.34, 1.8, .64, 1); }
.p-chip:active { transform: scale(.94); transition: transform .08s linear; }
.p-chip.act { animation: chipPop .3s cubic-bezier(.34, 1.8, .64, 1); }
@keyframes chipPop { 0% { transform: scale(1); } 40% { transform: scale(.94); } 100% { transform: scale(1); } }

/* 蒙层：从 tab 分段底部开始置灰（top 由 maskTop 实测）；低于分段(41<42)，面板(43)在上不被遮（同成果库） */
.panel-mask {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 41;
  background: rgba(16, 24, 40, 0.2);
  animation: maskIn .22s ease-out;
}
@keyframes maskIn { from { opacity: 0; } to { opacity: 1; } }

/* 发布入口（仅企业可见）：与筛选分段同排右端，紧凑高度 */
.publish-entry {
  flex: none;
  min-height: 36px;
  padding: 0 12px;
  border-radius: 8px;
  background: #0A66C2;
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  display: flex;
  align-items: center;
  margin-bottom: 8px; /* 与 tab 底部注线留白(16rpx)对齐，视觉同水平线 */
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
/* 筛选分段/面板（对齐成果库 no-motion 适配） */
.page.no-motion .stg-arr { transition: none; }
.page.no-motion .p-chip { transition: none; }
.page.no-motion .p-chip.act { animation: none; }
.page.no-motion .stg.on::after { animation: none; }
.page.no-motion .field-panel { animation: panelIn .3s ease-out; }
.page.no-motion .field-panel.closing { animation: panelOut .16s ease-in forwards; }
.page.no-motion .panel-mask { animation: maskIn .22s ease-out; }
.page.no-motion .p-chip:active,
.page.no-motion .stg:active { transform: none; }

/* ===== prefers-reduced-motion（系统级减弱动效）：筛选分段/面板动画与过渡全关 ===== */
@media (prefers-reduced-motion: reduce) {
  .stg, .stg-arr, .p-chip, .field-panel, .panel-mask {
    animation: none !important;
    transition: none !important;
  }
}
</style>
