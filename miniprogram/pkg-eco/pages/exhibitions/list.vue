<template>
  <view class="page" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar title="展会排期" show-back :fixed="true" @back="goBack" />

    <!-- 搜索 -->
    <view class="sbar">
      <view class="sbox">
        <u-icon name="search" size="28rpx" color="#98A2B3" />
        <input class="sinp" v-model="q" placeholder="搜索展会名称、地点" placeholder-class="ph" confirm-type="search" @input="onSearch" />
        <view v-if="q" class="sclr" @tap="clearSearch">×</view>
      </view>
    </view>

    <!-- 筛选器 -->
    <view class="fbar">
      <view class="fpill" :class="{ on: showPanel === 'cat' }" @tap="togglePanel('cat')">
        <text class="fval">{{ filters.cat.label }}</text><text class="farr">▾</text>
      </view>
      <view class="fpill" :class="{ on: showPanel === 'status' }" @tap="togglePanel('status')">
        <text class="fval">{{ filters.status.label }}</text><text class="farr">▾</text>
      </view>
      <view class="fpill" :class="{ on: showPanel === 'time' }" @tap="togglePanel('time')">
        <text class="fval">{{ filters.time.label }}</text><text class="farr">▾</text>
      </view>
    </view>

    <!-- 筛选面板 -->
    <view v-if="showPanel" class="panel-wrap" @tap.stop>
      <view class="panel">
        <view
          v-for="o in panelOptions"
          :key="o.value"
          class="p-opt"
          :class="{ act: filters[showPanel].value === o.value }"
          @tap="pickFilter(o.value)"
        >{{ o.label }}</view>
      </view>
    </view>

    <!-- 信息行：共 N 个展会 + 排序 -->
    <view class="ir">
      <text>共 <text class="irn">{{ list.length }}</text> 个展会</text>
      <view class="irs-wrap">
        <text class="irs" @tap="toggleSort">{{ sortLabel }} ▾</text>
        <view v-if="showSort" class="spop" @tap.stop>
          <view v-for="s in SORTS" :key="s.v" class="sp-opt" :class="{ act: sort === s.v }" @tap="pickSort(s.v)">
            <text>{{ s.l }}</text><text v-if="sort === s.v" class="chk">✓</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 列表 -->
    <view v-if="loading && !list.length" class="sk-list">
      <view v-for="i in 3" :key="'sk' + i" class="sk-card">
        <view class="sk-cv"></view>
        <view class="sk-bd"><view class="sk-l w80"></view><view class="sk-l w60"></view><view class="sk-l w40"></view></view>
      </view>
    </view>

    <view v-else-if="!list.length && !err" class="st">
      <u-empty description="暂无匹配展会">
        <text class="st-t">试试调整筛选或搜索关键词</text>
        <view class="stb" @tap="resetAll">清除筛选</view>
      </u-empty>
    </view>

    <view v-else-if="err && !list.length" class="st">
      <u-empty description="加载失败，请检查网络">
        <view class="stb" @tap="fetchList(true)">重新加载</view>
      </u-empty>
    </view>

    <view v-else class="ex-list">
      <view v-for="x in list" :key="x.id" class="ex-card" hover-class="tap-fade" @tap="goDetail(x)">
        <image v-if="x.cover" class="ex-cv" :src="x.cover" mode="aspectFill" />
        <view v-else class="ex-cv" :class="x.grad">
          <text class="cv-tag">{{ x.catLabel }}</text>
          <text class="cv-char">{{ x.char }}</text>
        </view>
        <view class="ex-bd">
          <text class="ex-t">{{ x.title }}</text>
          <text class="ex-line">{{ x.dateText }} · {{ x.location }}</text>
          <view class="ex-tags">
            <text class="ex-tag cat">{{ x.catLabel }}</text>
            <text class="ex-tag" :class="'st--' + x.status">{{ x.statusLabel }}</text>
          </view>
          <view class="ex-foot">
            <text class="ex-booth">{{ x.boothCount }} 个展位</text>
            <text class="ex-price">{{ x.price }}/展位</text>
          </view>
        </view>
      </view>
      <view class="lm">
        <text v-if="loadingMore">加载中...</text>
        <text v-else-if="!hasMore">— 没有更多了 —</text>
        <text v-else @tap="loadMore">上拉加载更多</text>
      </view>
    </view>

    <view class="mock-note" v-if="mockMode">当前为演示数据 · 接口就绪后自动切换</view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { MOCK_EXHIBITIONS, EXPO_CATEGORY_LABEL, EXPO_STATUS_LABEL, fmtRange, fmtFen, gradOfCategory } from '@/utils/mockExhibitions'

// ===== 静态配置 =====
const CAT_OPTS = [
  { value: 'all', label: '全部展会' },
  { value: 'drone_show', label: '无人机展' },
  { value: 'equipment_expo', label: '装备展' },
  { value: 'innovation_week', label: '创新周' },
]
const STATUS_OPTS = [
  { value: 'all', label: '全部状态' },
  { value: 'recruiting', label: '报名中' },
  { value: 'underway', label: '进行中' },
  { value: 'ended', label: '已结束' },
]
const TIME_OPTS = [
  { value: 'all', label: '全部时间' },
  { value: 'upcoming', label: '即将举办' },
  { value: 'ended', label: '已结束' },
]
const PANEL_OPTS = { cat: CAT_OPTS, status: STATUS_OPTS, time: TIME_OPTS }
const SORTS = [
  { v: 'default', l: '默认排序' },
  { v: 'upcoming', l: '即将开始' },
  { v: 'booths', l: '展位最多' },
]
const SORT_LABEL = { default: '默认排序', upcoming: '即将开始', booths: '展位最多' }
const PAGE_SIZE = 50

// ===== 状态 =====
const statusBarHeight = ref(20)
const q = ref('')
const showPanel = ref('')
const showSort = ref(false)
const sort = ref('default')
const filters = ref({
  cat: { value: 'all', label: '全部展会' },
  status: { value: 'all', label: '全部状态' },
  time: { value: 'all', label: '全部时间' },
})
const loading = ref(true)
const loadingMore = ref(false)
const err = ref(false)
const mockMode = ref(false)
const list = ref([])
const fullList = ref([])
const page = ref(1)
const hasMore = ref(true)
const rawById = new Map()

const panelOptions = computed(() => (showPanel.value ? PANEL_OPTS[showPanel.value] : []))
const sortLabel = computed(() => SORT_LABEL[sort.value] || '默认排序')

// ===== 数据映射 =====
const mapItem = (it) => ({
  id: it.id,
  title: it.title || '',
  catLabel: EXPO_CATEGORY_LABEL[it.category] || it.category || '展会',
  status: it.status || '',
  statusLabel: EXPO_STATUS_LABEL[it.status] || (it.status === 'draft' ? '未发布' : '未知'),
  location: it.location || '',
  dateText: fmtRange(it.start_date, it.end_date),
  boothCount: it.booth_count || 0,
  price: fmtFen(it.booth_price_fen),
  cover: it.cover_url || '',
  char: '展',
  grad: gradOfCategory(it.category),
  start: String(it.start_date || ''),
  raw: it,
})

// ===== 筛选 / 排序（前端实现；后端列表接口暂只支持分页） =====
const applyFilter = () => {
  const kw = (q.value || '').trim().toLowerCase()
  let items = fullList.value.slice()
  if (kw) items = items.filter((x) => (x.title + ' ' + x.location).toLowerCase().includes(kw))
  if (filters.value.cat.value !== 'all') items = items.filter((x) => x.raw.category === filters.value.cat.value)
  if (filters.value.status.value !== 'all') items = items.filter((x) => x.status === filters.value.status.value)
  if (filters.value.time.value !== 'all') {
    items = items.filter((x) => (filters.value.time.value === 'ended' ? x.status === 'ended' : x.status !== 'ended'))
  }
  if (sort.value === 'upcoming') items.sort((a, b) => (a.start || '').localeCompare(b.start || ''))
  else if (sort.value === 'booths') items.sort((a, b) => b.boothCount - a.boothCount)
  list.value = items
}

// ===== 数据获取 =====
// 接口替换点：GET /api/v1/exhibitions (page / page_size)；失败/为空时回退 mock
const fetchList = async (reset) => {
  if (reset) {
    page.value = 1
    hasMore.value = true
    loading.value = true
  } else {
    loadingMore.value = true
  }
  err.value = false
  try {
    const res = await request({ url: '/api/v1/exhibitions', data: { page: page.value, page_size: PAGE_SIZE } })
    const data = Array.isArray(res) ? res : (res && res.data) || res || {}
    const items = Array.isArray(data) ? data : (data && data.items) || []
    const total = (data && data.total) != null ? data.total : items.length
    if (reset) {
      rawById.clear()
      fullList.value = items.map(mapItem)
    } else {
      fullList.value = fullList.value.concat(items.map(mapItem))
    }
    items.forEach((it) => rawById.set(it.id, it))
    hasMore.value = fullList.value.length < total
    mockMode.value = false
  } catch {
    if (reset) {
      if (import.meta.env.DEV) { useMock() }
    } else {
      hasMore.value = false
    }
  } finally {
    loading.value = false
    loadingMore.value = false
  }
  applyFilter()
}

const useMock = () => {
  rawById.clear()
  MOCK_EXHIBITIONS.forEach((m) => rawById.set(m.id, m))
  fullList.value = MOCK_EXHIBITIONS.map(mapItem)
  hasMore.value = false
  mockMode.value = true
}

const loadMore = () => {
  if (loadingMore.value || !hasMore.value) return
  page.value += 1
  fetchList(false)
}

// ===== 交互 =====
const onSearch = () => applyFilter()
const clearSearch = () => { q.value = ''; applyFilter() }
const togglePanel = (key) => {
  showPanel.value = showPanel.value === key ? '' : key
  showSort.value = false
}
const pickFilter = (value) => {
  const key = showPanel.value
  const opt = PANEL_OPTS[key].find((o) => o.value === value)
  if (opt) filters.value[key] = { value: opt.value, label: opt.label }
  showPanel.value = ''
  applyFilter()
}
const toggleSort = () => {
  showSort.value = !showSort.value
  showPanel.value = ''
}
const pickSort = (v) => { sort.value = v; showSort.value = false; applyFilter() }
const resetAll = () => {
  q.value = ''
  filters.value = { cat: { value: 'all', label: '全部展会' }, status: { value: 'all', label: '全部状态' }, time: { value: 'all', label: '全部时间' } }
  sort.value = 'default'
  applyFilter()
}

const cacheRaw = (id) => {
  const raw = rawById.get(id)
  if (raw) uni.setStorageSync('exhibition_cache_' + id, raw)
}
const goDetail = (x) => {
  cacheRaw(x.id)
  uni.navigateTo({ url: '/pkg-eco/pages/exhibitions/detail?id=' + encodeURIComponent(x.id) })
}
const goBack = () => uni.navigateBack()

onLoad(() => {
  try {
    const sys = uni.getSystemInfoSync()
    statusBarHeight.value = sys.statusBarHeight || 20
  } catch (e) { /* 保持默认 */ }
  fetchList(true)
})
onPullDownRefresh(async () => {
  await fetchList(true)
  uni.stopPullDownRefresh()
})
onReachBottom(() => {
  if (!loading.value) loadMore()
})
</script>

<style scoped>
.page { min-height: 100vh; background: #F4F6F8; padding-bottom: 40rpx; }

/* ===== 搜索 ===== */
.sbar { display: flex; padding: 12rpx 24rpx 8rpx; background: #fff; }
.sbox { flex: 1; display: flex; align-items: center; gap: 8rpx; height: 76rpx; background: #F4F6F8; border: 1px solid #EEF1F4; border-radius: 16rpx; padding: 0 20rpx; }
.sinp { flex: 1; min-width: 0; font-size: 26rpx; color: #17212B; }
.ph { color: #98A2B3; }
.sclr { width: 40rpx; height: 40rpx; display: flex; align-items: center; justify-content: center; color: #98A2B3; font-size: 32rpx; }

/* ===== 筛选器 ===== */
.fbar { display: flex; gap: 16rpx; padding: 8rpx 24rpx 12rpx; background: #fff; }
.fpill { flex: 1; height: 64rpx; display: flex; align-items: center; justify-content: center; gap: 6rpx; background: #F4F6F8; border: 1px solid #EEF1F4; border-radius: 12rpx; font-size: 24rpx; color: #344054; }
.fpill.on { border-color: #0A66C2; color: #0A66C2; background: #EAF3FB; }
.farr { font-size: 18rpx; color: #98A2B3; }
.panel-wrap { position: relative; z-index: 40; background: #fff; }
.panel { position: absolute; left: 24rpx; right: 24rpx; top: 4rpx; background: #fff; border-radius: 20rpx; box-shadow: 0 12rpx 36rpx rgba(0,0,0,.16); padding: 16rpx; animation: panelIn .2s cubic-bezier(.2,.9,.3,1); }
@keyframes panelIn { from { opacity: 0; transform: translateY(-8rpx) scale(.98); } to { opacity: 1; transform: translateY(0) scale(1); } }
.p-opt { padding: 22rpx 28rpx; border-radius: 12rpx; font-size: 26rpx; color: #17212B; }
.p-opt:active { background: #F4F6F8; }
.p-opt.act { color: #0A66C2; font-weight: 600; background: #EAF3FB; }

/* ===== 信息行 + 排序 ===== */
.ir { display: flex; justify-content: space-between; align-items: center; padding: 16rpx 24rpx 8rpx; font-size: 24rpx; color: #667085; position: relative; z-index: 30; }
.irn { color: #0A66C2; font-weight: 600; }
.irs-wrap { position: relative; }
.irs { color: #0A66C2; font-weight: 500; padding: 8rpx 16rpx; border-radius: 12rpx; }
.irs:active { background: #EAF3FB; }
.spop { position: absolute; top: 56rpx; right: 0; z-index: 90; background: #fff; border-radius: 20rpx; box-shadow: 0 8rpx 28rpx rgba(0,0,0,.14); padding: 12rpx; min-width: 260rpx; animation: panelIn .16s ease; }
.sp-opt { padding: 22rpx 24rpx; border-radius: 14rpx; font-size: 26rpx; color: #17212B; display: flex; align-items: center; justify-content: space-between; }
.sp-opt:active { background: #F4F6F8; }
.sp-opt.act { color: #0A66C2; font-weight: 600; background: #EAF3FB; }
.chk { color: #0A66C2; font-size: 24rpx; }

/* ===== 展会卡片 ===== */
.ex-list { padding: 8rpx 24rpx 24rpx; }
.ex-card { display: flex; gap: 24rpx; background: #fff; border: 1px solid #EEF1F4; border-radius: 20rpx; padding: 24rpx; margin-bottom: 20rpx; }
.ex-cv { width: 208rpx; height: 208rpx; border-radius: 16rpx; overflow: hidden; position: relative; flex: none; display: flex; align-items: center; justify-content: center; }
.cv-char { font-size: 68rpx; font-weight: 800; color: rgba(255,255,255,.95); text-shadow: 0 2px 10px rgba(0,0,0,.25); }
.cv-tag { position: absolute; left: 12rpx; top: 12rpx; font-size: 18rpx; color: #fff; background: rgba(0,0,0,.4); padding: 4rpx 12rpx; border-radius: 8rpx; }
.ex-bd { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 8rpx; }
.ex-t { font-size: 29rpx; font-weight: 600; color: #17212B; line-height: 1.35; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.ex-line { font-size: 23rpx; color: #667085; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.ex-tags { display: flex; gap: 12rpx; margin-top: 2rpx; }
.ex-tag { font-size: 19rpx; padding: 4rpx 14rpx; border-radius: 8rpx; font-weight: 500; }
.ex-tag.cat { color: #0A66C2; background: #EAF3FB; }
.ex-tag.st--recruiting { color: #168A55; background: #E9F7F0; }
.ex-tag.st--underway { color: #0A66C2; background: #EAF3FB; }
.ex-tag.st--ended { color: #667085; background: #F2F4F7; }
.ex-foot { display: flex; justify-content: space-between; align-items: baseline; margin-top: auto; }
.ex-booth { font-size: 22rpx; color: #0A66C2; font-weight: 600; }
.ex-price { font-size: 22rpx; color: #98A2B3; }

/* ===== 骨架 ===== */
.sk-list { padding: 8rpx 24rpx; }
.sk-card { display: flex; gap: 24rpx; background: #fff; border: 1px solid #EEF1F4; border-radius: 20rpx; padding: 24rpx; margin-bottom: 20rpx; }
.sk-cv { width: 208rpx; height: 208rpx; border-radius: 16rpx; background: #f0f1f3; animation: shimmer 1.5s infinite; flex: none; }
.sk-bd { flex: 1; display: flex; flex-direction: column; gap: 20rpx; padding-top: 8rpx; }
.sk-l { height: 24rpx; background: #f0f1f3; border-radius: 12rpx; animation: shimmer 1.5s infinite; }
.sk-l.w80 { width: 80%; }
.sk-l.w60 { width: 60%; }
.sk-l.w40 { width: 40%; }
@keyframes shimmer { 0%, 100% { opacity: 1; } 50% { opacity: .45; } }

/* ===== 空态 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 160rpx 48rpx; }
.st-t { font-size: 24rpx; color: #98A2B3; margin-top: 24rpx; }
.stb { margin-top: 28rpx; padding: 16rpx 48rpx; border-radius: 16rpx; background: #0A66C2; color: #fff; font-size: 26rpx; font-weight: 500; }

/* ===== 加载更多 ===== */
.lm { text-align: center; padding: 20rpx 0; font-size: 22rpx; color: #98A2B3; }

/* ===== 演示提示 ===== */
.mock-note { text-align: center; padding: 0 0 24rpx; font-size: 20rpx; color: #98A2B3; }

/* ===== 封面渐变 ===== */
.gd-1 { background: linear-gradient(135deg,#0d47a1,#1565c0 60%,#42a5f5); }
.gd-2 { background: linear-gradient(135deg,#004d40,#00695c 60%,#26a69a); }
.gd-4 { background: linear-gradient(135deg,#4a148c,#6a1b9a 60%,#ab47bc); }
.gd-5 { background: linear-gradient(135deg,#263238,#37474f 60%,#607d8b); }
</style>
